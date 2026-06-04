package adapter

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"math"
	"net/http"
	"strings"
	"time"

	"xiaozhi-go-core/internal/config"
	"xiaozhi-go-core/internal/domain"
)

type LLMClient struct {
	store  *config.Store
	logger *slog.Logger
}

func NewLLMClient(store *config.Store, logger *slog.Logger) *LLMClient {
	return &LLMClient{store: store, logger: logger}
}

type StreamChatOpts struct {
	Messages  []domain.Message
	Functions []map[string]interface{}
}

// StreamChat sends messages to the LLM and yields ChatChunk via a channel.
// Supports both plain text streaming and tool_calls detection.
func (c *LLMClient) StreamChat(ctx context.Context, opts StreamChatOpts) (<-chan domain.ChatChunk, error) {
	eff := config.EffectiveLLMHTTP(c.store.Get().LLM)
	client := &http.Client{
		Timeout: time.Duration(eff.TimeoutMs) * time.Millisecond,
	}

	apiMessages := buildAPIMessages(opts.Messages)

	body := map[string]interface{}{
		"model":    eff.ModelName,
		"messages": apiMessages,
		"stream":   true,
	}
	if len(opts.Functions) > 0 {
		body["tools"] = opts.Functions
		body["tool_choice"] = "auto"
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	baseURL := strings.TrimRight(eff.BaseURL, "/")
	url := baseURL + "/chat/completions"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+eff.APIKey)
	c.logger.Info("llm_request_start",
		"url", url,
		"model", eff.ModelName,
		"messages_count", len(apiMessages),
		"tools_count", len(opts.Functions),
		"body_preview", previewJSONString(bodyBytes, 1200),
	)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request: %w", err)
	}
	c.logger.Info("llm_response_headers",
		"status_code", resp.StatusCode,
		"content_type", resp.Header.Get("Content-Type"),
	)
	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("upstream status %d: %s", resp.StatusCode, string(respBody))
	}

	ch := make(chan domain.ChatChunk, 32)
	go c.consumeSSE(ctx, resp.Body, ch)
	return ch, nil
}

// StreamChatWithRetry wraps StreamChat with exponential backoff retry.
func (c *LLMClient) StreamChatWithRetry(ctx context.Context, opts StreamChatOpts, maxRetries int) (<-chan domain.ChatChunk, error) {
	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			backoff := time.Duration(math.Pow(2, float64(attempt-1))*200) * time.Millisecond
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff):
			}
			c.logger.Warn("LLM retry", "attempt", attempt, "last_error", lastErr)
		}
		ch, err := c.StreamChat(ctx, opts)
		if err == nil {
			return ch, nil
		}
		lastErr = err
	}
	return nil, fmt.Errorf("all %d retries exhausted: %w", maxRetries, lastErr)
}

func buildAPIMessages(messages []domain.Message) []map[string]interface{} {
	// Upstream (OpenAI-compatible) strictness:
	// - system messages must be at the beginning (some gateways reject any system role after index 0)
	// Normalize by keeping at most one system message and forcing it to index 0.
	system := ""
	normalized := make([]domain.Message, 0, len(messages))
	for _, m := range messages {
		if strings.EqualFold(strings.TrimSpace(m.Role), "system") {
			if system == "" {
				system = m.Content
			}
			continue
		}
		normalized = append(normalized, m)
	}
	if system != "" {
		normalized = append([]domain.Message{{Role: "system", Content: system}}, normalized...)
	}

	result := make([]map[string]interface{}, 0, len(normalized))
	for _, m := range normalized {
		msg := map[string]interface{}{
			"role":    m.Role,
			"content": m.Content,
		}
		if m.ToolCallID != "" {
			msg["tool_call_id"] = m.ToolCallID
		}
		if len(m.ToolCalls) > 0 {
			tc := make([]map[string]interface{}, 0, len(m.ToolCalls))
			for _, t := range m.ToolCalls {
				tc = append(tc, map[string]interface{}{
					"id":    t.ID,
					"type":  "function",
					"index": t.Index,
					"function": map[string]string{
						"name":      t.Function.Name,
						"arguments": t.Function.Arguments,
					},
				})
			}
			msg["tool_calls"] = tc
		}
		result = append(result, msg)
	}
	return result
}

// sseToolCallDelta tracks incremental tool_call construction from SSE chunks.
type sseToolCallDelta struct {
	ID        string
	Index     int
	Name      string
	Arguments strings.Builder
}

func (c *LLMClient) consumeSSE(ctx context.Context, body io.ReadCloser, ch chan<- domain.ChatChunk) {
	defer close(ch)
	defer body.Close()

	noThinkStrip := c.store.Get().LLM.DisableThinkTagsStrip

	var pendingToolCalls []*sseToolCallDelta
	inThink := false
	thinkCarry := ""

	scanner := bufio.NewScanner(body)
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			ch <- domain.ChatChunk{Event: "ERROR", ErrorCode: "CHAT_TIMEOUT", ErrorMessage: "context cancelled"}
			return
		default:
		}

		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			break
		}

		var sse sseChunk
		if err := json.Unmarshal([]byte(data), &sse); err != nil {
			c.logger.Warn("parse SSE chunk failed", "error", err)
			continue
		}
		if len(sse.Choices) == 0 {
			continue
		}
		choice := sse.Choices[0]

		// Handle tool_calls delta
		if len(choice.Delta.ToolCalls) > 0 {
			for _, tc := range choice.Delta.ToolCalls {
				idx := tc.Index
				for len(pendingToolCalls) <= idx {
					pendingToolCalls = append(pendingToolCalls, &sseToolCallDelta{Index: len(pendingToolCalls)})
				}
				if tc.ID != "" {
					pendingToolCalls[idx].ID = tc.ID
				}
				if tc.Function.Name != "" {
					pendingToolCalls[idx].Name = tc.Function.Name
				}
				if tc.Function.Arguments != "" {
					pendingToolCalls[idx].Arguments.WriteString(tc.Function.Arguments)
				}
			}
		}

		// Handle content delta
		content := choice.Delta.Content
		if content != "" && len(pendingToolCalls) == 0 {
			var filtered string
			if noThinkStrip {
				filtered = content
			} else {
				filtered = stripThinkTagsStream(content, &inThink, &thinkCarry)
			}
			if filtered != "" {
				// 内容 delta 频率非常高，并发场景下会被日志 I/O 放大成瓶颈：
				// - 降级为 Debug，保留必要排查能力，同时避免压测/生产下刷屏拖慢主链路
				c.logger.Debug("llm_sse_content_delta", "delta_preview", previewText(filtered, 180), "delta_len", len(filtered))
				ch <- domain.ChatChunk{Event: "CONTENT_CHUNK", ContentChunk: filtered}
			}
		}

		if choice.FinishReason != nil && *choice.FinishReason != "" {
			reason := *choice.FinishReason
			c.logger.Info("llm_sse_finish_reason", "finish_reason", reason, "pending_tool_calls", len(pendingToolCalls))
			if (reason == "tool_calls" || reason == "function_call") && len(pendingToolCalls) > 0 {
				toolCalls := make([]domain.ToolCall, 0, len(pendingToolCalls))
				for _, ptc := range pendingToolCalls {
					toolCalls = append(toolCalls, domain.ToolCall{
						ID:    ptc.ID,
						Type:  "function",
						Index: ptc.Index,
						Function: domain.FunctionCall{
							Name:      ptc.Name,
							Arguments: ptc.Arguments.String(),
						},
					})
				}
				ch <- domain.ChatChunk{Event: "TOOL_CALLS", ToolCalls: toolCalls, FinishReason: reason}
			}
			return
		}
	}

	// If we exit the loop with pending tool calls (some models don't send finish_reason for tools)
	if len(pendingToolCalls) > 0 {
		toolCalls := make([]domain.ToolCall, 0, len(pendingToolCalls))
		for _, ptc := range pendingToolCalls {
			if ptc.Name != "" {
				toolCalls = append(toolCalls, domain.ToolCall{
					ID:    ptc.ID,
					Type:  "function",
					Index: ptc.Index,
					Function: domain.FunctionCall{
						Name:      ptc.Name,
						Arguments: ptc.Arguments.String(),
					},
				})
			}
		}
		if len(toolCalls) > 0 {
			c.logger.Info("llm_sse_tool_calls_fallback_finish", "tool_calls_count", len(toolCalls))
			ch <- domain.ChatChunk{Event: "TOOL_CALLS", ToolCalls: toolCalls, FinishReason: "tool_calls"}
		}
	}
}

func stripThinkTagsStream(content string, inThink *bool, carry *string) string {
	const (
		startTag = "<think>"
		endTag   = "</think>"
	)

	input := *carry + content
	*carry = ""
	if input == "" {
		return ""
	}

	var out strings.Builder
	i := 0
	for i < len(input) {
		if *inThink {
			endIdx := strings.Index(input[i:], endTag)
			if endIdx < 0 {
				rest := input[i:]
				keep := maxPartialTagPrefixSuffix(rest, []string{endTag})
				if keep > 0 {
					*carry = rest[len(rest)-keep:]
				}
				return out.String()
			}
			i += endIdx + len(endTag)
			*inThink = false
			continue
		}

		startIdx := strings.Index(input[i:], startTag)
		endIdx := strings.Index(input[i:], endTag)
		switch {
		case startIdx >= 0 && (endIdx < 0 || startIdx < endIdx):
			out.WriteString(input[i : i+startIdx])
			i += startIdx + len(startTag)
			*inThink = true
		case endIdx >= 0:
			out.WriteString(input[i : i+endIdx])
			i += endIdx + len(endTag)
		default:
			rest := input[i:]
			keep := maxPartialTagPrefixSuffix(rest, []string{startTag, endTag})
			if keep > 0 {
				out.WriteString(rest[:len(rest)-keep])
				*carry = rest[len(rest)-keep:]
			} else {
				out.WriteString(rest)
			}
			i = len(input)
		}
	}
	return out.String()
}

func maxPartialTagPrefixSuffix(s string, tags []string) int {
	maxKeep := 0
	for _, tag := range tags {
		limit := len(tag) - 1
		if limit <= 0 {
			continue
		}
		if limit > len(s) {
			limit = len(s)
		}
		for n := limit; n > 0; n-- {
			if strings.HasSuffix(s, tag[:n]) {
				if n > maxKeep {
					maxKeep = n
				}
				break
			}
		}
	}
	return maxKeep
}

func previewJSONString(b []byte, max int) string {
	if max <= 0 {
		max = 1000
	}
	s := strings.TrimSpace(string(b))
	if len(s) <= max {
		return s
	}
	return s[:max] + "...(truncated)"
}

func previewText(s string, max int) string {
	if max <= 0 {
		max = 200
	}
	s = strings.TrimSpace(s)
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	return string(r[:max]) + "...(truncated)"
}

type sseChunk struct {
	Choices []struct {
		Delta struct {
			Content   string `json:"content"`
			ToolCalls []struct {
				Index    int    `json:"index"`
				ID       string `json:"id"`
				Function struct {
					Name      string `json:"name"`
					Arguments string `json:"arguments"`
				} `json:"function"`
			} `json:"tool_calls"`
		} `json:"delta"`
		FinishReason *string `json:"finish_reason"`
	} `json:"choices"`
}
