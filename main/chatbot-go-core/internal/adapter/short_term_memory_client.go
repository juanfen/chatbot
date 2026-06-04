package adapter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"xiaozhi-go-core/internal/config"
	"xiaozhi-go-core/internal/domain"
)

// ShortTermMemoryClient 调用短期记忆服务：build_prompt（组装 messages）与 commit_turn（落库一轮对话）。
type ShortTermMemoryClient struct {
	store  *config.Store
	logger *slog.Logger
}

func NewShortTermMemoryClient(store *config.Store, logger *slog.Logger) *ShortTermMemoryClient {
	return &ShortTermMemoryClient{store: store, logger: logger}
}

func (c *ShortTermMemoryClient) effectiveCfg() config.ShortTermMemoryConfig {
	if c == nil || c.store == nil {
		return config.ShortTermMemoryConfig{}
	}
	return c.store.Get().ShortTermMemory
}

// BuildPrompt POST /short-term/build_prompt，返回可供 LLM 使用的 messages。
func (c *ShortTermMemoryClient) BuildPrompt(ctx context.Context, userID, deviceID, sessionID, query string) ([]domain.Message, error) {
	cfg := c.effectiveCfg()
	if !cfg.Enabled || strings.TrimSpace(cfg.BaseURL) == "" {
		return nil, fmt.Errorf("short_term_memory disabled or base_url empty")
	}
	timeoutMs := cfg.TimeoutMs
	if timeoutMs <= 0 {
		timeoutMs = 30000
	}
	reqCtx, cancel := context.WithTimeout(ctx, time.Duration(timeoutMs)*time.Millisecond)
	defer cancel()

	base := strings.TrimRight(strings.TrimSpace(cfg.BaseURL), "/")
	url := base + "/short-term/build_prompt"
	body := map[string]string{
		"user_id":    userID,
		"device_id":  deviceID,
		"session_id": sessionID,
		"query":      query,
	}
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal build_prompt body: %w", err)
	}

	req, err := http.NewRequestWithContext(reqCtx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("create build_prompt request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	c.logger.Info("short_term_build_prompt_start",
		"url", url,
		"session_id", sessionID,
		"device_id", deviceID,
		"query_preview", previewText(query, 120),
	)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("build_prompt http: %w", err)
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("build_prompt read body: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("build_prompt status %d: %s", resp.StatusCode, string(respBody))
	}

	msgs, err := parseQwenMessages(respBody)
	if err != nil {
		return nil, err
	}
	c.logger.Info("short_term_build_prompt_ok",
		"messages_count", len(msgs),
	)
	return msgs, nil
}

// CommitTurn POST /short-term/commit_turn，在 LLM 完整回复后写入记忆。
func (c *ShortTermMemoryClient) CommitTurn(ctx context.Context, userID, deviceID, sessionID, query, answer string) error {
	cfg := c.effectiveCfg()
	if !cfg.Enabled || strings.TrimSpace(cfg.BaseURL) == "" {
		return fmt.Errorf("short_term_memory disabled or base_url empty")
	}
	timeoutMs := cfg.TimeoutMs
	if timeoutMs <= 0 {
		timeoutMs = 30000
	}
	reqCtx, cancel := context.WithTimeout(ctx, time.Duration(timeoutMs)*time.Millisecond)
	defer cancel()

	base := strings.TrimRight(strings.TrimSpace(cfg.BaseURL), "/")
	url := base + "/short-term/commit_turn"
	body := map[string]string{
		"user_id":    userID,
		"device_id":  deviceID,
		"session_id": sessionID,
		"query":      query,
		"answer":     answer,
	}
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("marshal commit_turn body: %w", err)
	}

	req, err := http.NewRequestWithContext(reqCtx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return fmt.Errorf("create commit_turn request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	c.logger.Info("short_term_commit_turn_start",
		"url", url,
		"session_id", sessionID,
		"answer_runes", len([]rune(answer)),
	)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("commit_turn http: %w", err)
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("commit_turn read body: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("commit_turn status %d: %s", resp.StatusCode, string(respBody))
	}
	c.logger.Info("short_term_commit_turn_ok", "session_id", sessionID)
	return nil
}

func parseQwenMessages(raw []byte) ([]domain.Message, error) {
	var envelope struct {
		Result struct {
			QwenMessages []struct {
				Role    string          `json:"role"`
				Content json.RawMessage `json:"content"`
			} `json:"qwen_messages"`
		} `json:"result"`
	}
	if err := json.Unmarshal(raw, &envelope); err != nil {
		return nil, fmt.Errorf("parse build_prompt json: %w", err)
	}
	out := make([]domain.Message, 0, len(envelope.Result.QwenMessages))
	for _, m := range envelope.Result.QwenMessages {
		content, ok := rawJSONToTextContent(m.Content)
		if !ok {
			continue
		}
		role := strings.TrimSpace(m.Role)
		if role == "" {
			role = "user"
		}
		out = append(out, domain.Message{Role: role, Content: content})
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("build_prompt returned empty qwen_messages")
	}
	return out, nil
}

func rawJSONToTextContent(raw json.RawMessage) (string, bool) {
	if len(raw) == 0 {
		return "", false
	}
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		return s, strings.TrimSpace(s) != ""
	}
	return "", false
}
