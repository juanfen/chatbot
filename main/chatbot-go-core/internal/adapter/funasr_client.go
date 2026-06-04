package adapter

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/gorilla/websocket"

	"xiaozhi-go-core/internal/config"
)

type FunASRClient struct {
	store  *config.Store
	logger *slog.Logger
}

type funASRResponse struct {
	Text    string `json:"text"`
	IsFinal bool   `json:"is_final"`
}

var funASRPrefixPattern = regexp.MustCompile(`^<\|.*?\|><\|.*?\|><\|.*?\|>`)

func NewFunASRClient(store *config.Store, logger *slog.Logger) *FunASRClient {
	return &FunASRClient{store: store, logger: logger}
}

func (c *FunASRClient) Recognize(ctx context.Context, pcmData []byte, sessionID string) (string, error) {
	if len(pcmData) == 0 {
		return "", nil
	}
	cfg := c.store.Get().ASR
	start := time.Now()
	timeoutMs := cfg.TimeoutMs
	if timeoutMs <= 0 {
		timeoutMs = 15000
	}
	reqCtx, cancel := context.WithTimeout(ctx, time.Duration(timeoutMs)*time.Millisecond)
	defer cancel()

	uri := c.buildURI(cfg)
	c.logger.Info("asr_connect_start",
		"session_id", sessionID,
		"uri", uri,
		"pcm_bytes", len(pcmData),
		"timeout_ms", timeoutMs,
	)
	dialer := websocket.Dialer{
		Subprotocols: []string{"binary"},
	}
	if cfg.IsSSL {
		dialer.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}
	headers := http.Header{}
	apiKey := cfg.APIKey
	if apiKey == "" {
		apiKey = "none"
	}
	headers.Set("Authorization", "Bearer; "+apiKey)

	conn, _, err := dialer.DialContext(reqCtx, uri, headers)
	if err != nil {
		c.logger.Error("asr_connect_failed",
			"session_id", sessionID,
			"uri", uri,
			"elapsed_ms", time.Since(start).Milliseconds(),
			"error", err,
		)
		return "", fmt.Errorf("connect funasr %s: %w", uri, err)
	}
	defer conn.Close()
	c.logger.Info("asr_connected",
		"session_id", sessionID,
		"uri", uri,
		"elapsed_ms", time.Since(start).Milliseconds(),
	)

	configMessage := map[string]any{
		"mode":           "offline",
		"chunk_size":     []int{5, 10, 5},
		"chunk_interval": 10,
		"wav_name":       sessionID,
		"is_speaking":    true,
		"itn":            false,
	}
	if err := conn.WriteJSON(configMessage); err != nil {
		c.logger.Error("asr_send_config_failed", "session_id", sessionID, "error", err)
		return "", fmt.Errorf("send funasr config: %w", err)
	}
	if err := conn.WriteMessage(websocket.BinaryMessage, pcmData); err != nil {
		c.logger.Error("asr_send_pcm_failed", "session_id", sessionID, "pcm_bytes", len(pcmData), "error", err)
		return "", fmt.Errorf("send funasr pcm: %w", err)
	}
	if err := conn.WriteJSON(map[string]bool{"is_speaking": false}); err != nil {
		c.logger.Error("asr_send_end_failed", "session_id", sessionID, "error", err)
		return "", fmt.Errorf("send funasr end: %w", err)
	}

	var text strings.Builder
	for {
		if deadline, ok := reqCtx.Deadline(); ok {
			_ = conn.SetReadDeadline(deadline)
		}
		_, data, err := conn.ReadMessage()
		if err != nil {
			if reqCtx.Err() != nil {
				c.logger.Error("asr_read_timeout",
					"session_id", sessionID,
					"elapsed_ms", time.Since(start).Milliseconds(),
					"error", reqCtx.Err(),
				)
				return "", reqCtx.Err()
			}
			c.logger.Error("asr_read_failed",
				"session_id", sessionID,
				"elapsed_ms", time.Since(start).Milliseconds(),
				"error", err,
			)
			return "", fmt.Errorf("read funasr response: %w", err)
		}
		var resp funASRResponse
		if err := json.Unmarshal(data, &resp); err != nil {
			c.logger.Warn("funasr_response_decode_failed", "body", string(data), "error", err)
			continue
		}
		text.WriteString(resp.Text)
		if resp.IsFinal {
			break
		}
	}
	result := strings.TrimSpace(funASRPrefixPattern.ReplaceAllString(text.String(), ""))
	c.logger.Info("asr_recognize_done",
		"session_id", sessionID,
		"elapsed_ms", time.Since(start).Milliseconds(),
		"text_len", len([]rune(result)),
		"text_preview", previewForLog(result, 80),
	)
	return result, nil
}

func (c *FunASRClient) buildURI(cfg config.ASRConfig) string {
	scheme := "ws"
	if cfg.IsSSL {
		scheme = "wss"
	}
	host := strings.TrimSpace(cfg.Host)
	if host == "" {
		host = "127.0.0.1"
	}
	if parsed, err := url.Parse(host); err == nil && parsed.Scheme != "" {
		switch parsed.Scheme {
		case "https", "wss":
			scheme = "wss"
		case "http", "ws":
			scheme = "ws"
		}
		if parsed.Host != "" {
			host = parsed.Host
		} else {
			host = strings.Trim(parsed.Path, "/")
		}
	}
	host = strings.TrimRight(host, "/")
	if h, _, err := net.SplitHostPort(host); err == nil && h != "" {
		return fmt.Sprintf("%s://%s", scheme, host)
	}
	port := cfg.Port
	if port <= 0 {
		port = 10095
	}
	return fmt.Sprintf("%s://%s:%d", scheme, host, port)
}

func previewForLog(text string, maxRunes int) string {
	runes := []rune(strings.TrimSpace(text))
	if len(runes) <= maxRunes {
		return string(runes)
	}
	return string(runes[:maxRunes]) + "...(truncated)"
}
