package adapter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"

	"xiaozhi-go-core/internal/config"
)

type VADClient struct {
	store  *config.Store
	logger *slog.Logger
	mu     sync.Mutex
	conns  map[string]*websocket.Conn // key = wsURL + "|" + sessionID
	logged map[string]struct{}         // 已打印过连接参数的 session|transport
}

type vadRequestParams struct {
	Transport            string
	BaseURL              string
	RequestURL           string
	SessionID            string
	Threshold            string
	ThresholdLow         string
	MinSilenceDurationMs string
	FrameWindowThreshold string
	TimeoutMs            int
	APIKeySet            bool
}

type VADResult struct {
	HaveVoice       bool `json:"have_voice"`
	ClientHaveVoice bool `json:"client_have_voice"`
	ClientVoiceStop bool `json:"client_voice_stop"`
	LastIsVoice     bool `json:"last_is_voice"`
}

func NewVADClient(store *config.Store, logger *slog.Logger) *VADClient {
	return &VADClient{
		store:  store,
		logger: logger,
		conns:  map[string]*websocket.Conn{},
		logged: map[string]struct{}{},
	}
}

func (c *VADClient) Infer(ctx context.Context, sessionID string, opusPacket []byte) (VADResult, error) {
	cfg := c.store.Get().VAD
	timeoutMs := cfg.TimeoutMs
	if timeoutMs <= 0 {
		timeoutMs = 1500
	}

	if strings.TrimSpace(cfg.BaseURL) == "" {
		return VADResult{}, fmt.Errorf("vad base_url is empty")
	}

	baseURL := strings.TrimSpace(cfg.BaseURL)
	lower := strings.ToLower(baseURL)
	switch {
	case strings.HasPrefix(lower, "ws://") || strings.HasPrefix(lower, "wss://"):
		result, err := c.inferWS(ctx, cfg, sessionID, opusPacket, timeoutMs)
		if err != nil {
			c.logInferError("ws", sessionID, len(opusPacket), err)
		}
		return result, err
	default:
		result, err := c.inferHTTP(ctx, cfg, sessionID, opusPacket, timeoutMs)
		if err != nil {
			c.logInferError("http", sessionID, len(opusPacket), err)
		}
		return result, err
	}
}

func (c *VADClient) logInferError(transport, sessionID string, opusBytes int, err error) {
	if c.logger == nil || err == nil {
		return
	}
	c.logger.Warn("vad_infer_failed",
		"transport", transport,
		"session_id", sessionID,
		"opus_bytes", opusBytes,
		"error", err,
	)
}

func (c *VADClient) Reset(ctx context.Context, sessionID string) {
	cfg := c.store.Get().VAD
	timeoutMs := cfg.TimeoutMs
	if timeoutMs <= 0 {
		timeoutMs = 1500
	}
	if strings.TrimSpace(cfg.BaseURL) == "" || sessionID == "" {
		return
	}

	baseURL := strings.TrimSpace(cfg.BaseURL)
	lower := strings.ToLower(baseURL)
	if strings.HasPrefix(lower, "ws://") || strings.HasPrefix(lower, "wss://") {
		// ws 版：关闭并清理该 session 的长连接
		_, wsURL, err := c.buildWSInferURL(cfg, sessionID)
		if err != nil {
			return
		}
		key := wsURL + "|" + sessionID
		c.mu.Lock()
		conn := c.conns[key]
		delete(c.conns, key)
		c.mu.Unlock()
		if conn != nil {
			_ = conn.Close()
		}
		c.mu.Lock()
		delete(c.logged, "ws|"+sessionID)
		c.mu.Unlock()
		return
	}

	httpClient := &http.Client{Timeout: time.Duration(timeoutMs) * time.Millisecond}
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, strings.TrimRight(baseURL, "/")+"/v1/vad/sessions/"+url.PathEscape(sessionID), nil)
	if err != nil {
		c.logger.Warn("vad_reset_request_create_failed", "session_id", sessionID, "error", err)
		return
	}
	if cfg.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+cfg.APIKey)
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		c.logger.Warn("vad_reset_failed", "session_id", sessionID, "error", err)
		return
	}
	resp.Body.Close()
	if resp.StatusCode >= 300 && resp.StatusCode != http.StatusNotFound {
		c.logger.Warn("vad_reset_failed", "session_id", sessionID, "status", resp.StatusCode)
	}
}

func (c *VADClient) inferHTTP(ctx context.Context, cfg config.VADConfig, sessionID string, opusPacket []byte, timeoutMs int) (VADResult, error) {
	httpClient := &http.Client{Timeout: time.Duration(timeoutMs) * time.Millisecond}

	params, requestURL, err := c.buildHTTPInferURL(cfg, sessionID)
	if err != nil {
		return VADResult{}, err
	}
	c.logConnectParamsOnce(params)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, requestURL, bytes.NewReader(opusPacket))
	if err != nil {
		return VADResult{}, fmt.Errorf("create vad request: %w", err)
	}
	req.Header.Set("Content-Type", "application/octet-stream")
	if cfg.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+cfg.APIKey)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return VADResult{}, fmt.Errorf("vad request failed: %w", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return VADResult{}, fmt.Errorf("vad status %d: %s", resp.StatusCode, string(body))
	}
	var result VADResult
	if err := json.Unmarshal(body, &result); err != nil {
		return VADResult{}, fmt.Errorf("decode vad response: %w", err)
	}
	return result, nil
}

func (c *VADClient) inferWS(ctx context.Context, cfg config.VADConfig, sessionID string, opusPacket []byte, timeoutMs int) (VADResult, error) {
	params, wsURL, err := c.buildWSInferURL(cfg, sessionID)
	if err != nil {
		return VADResult{}, err
	}

	conn, err := c.getOrDialWS(ctx, params, wsURL, sessionID, timeoutMs, cfg.APIKey)
	if err != nil {
		return VADResult{}, err
	}

	// 每次请求设置一次读写 deadline（防止 hang）
	deadline := time.Now().Add(time.Duration(timeoutMs) * time.Millisecond)
	if dl, ok := ctx.Deadline(); ok && dl.Before(deadline) {
		deadline = dl
	}
	_ = conn.SetReadDeadline(deadline)
	_ = conn.SetWriteDeadline(deadline)

	if err := conn.WriteMessage(websocket.BinaryMessage, opusPacket); err != nil {
		c.dropWS(wsURL, sessionID, conn)
		return VADResult{}, fmt.Errorf("vad ws write failed: %w", err)
	}
	_, data, err := conn.ReadMessage()
	if err != nil {
		c.dropWS(wsURL, sessionID, conn)
		return VADResult{}, fmt.Errorf("vad ws read failed: %w", err)
	}
	var result VADResult
	if err := json.Unmarshal(data, &result); err != nil {
		return VADResult{}, fmt.Errorf("decode vad ws response: %w", err)
	}
	return result, nil
}

func (c *VADClient) buildVADQueryValues(cfg config.VADConfig, sessionID string) url.Values {
	q := url.Values{}
	q.Set("session_id", sessionID)
	q.Set("threshold", strconv.FormatFloat(effectiveFloat(cfg.Threshold, 0.5), 'f', -1, 64))
	q.Set("threshold_low", strconv.FormatFloat(effectiveFloat(cfg.ThresholdLow, 0.3), 'f', -1, 64))
	q.Set("min_silence_duration_ms", strconv.Itoa(effectiveInt(cfg.MinSilenceDurationMs, 200)))
	q.Set("frame_window_threshold", strconv.Itoa(effectiveInt(cfg.FrameWindowThreshold, 3)))
	return q
}

func (c *VADClient) buildWSInferURL(cfg config.VADConfig, sessionID string) (vadRequestParams, string, error) {
	baseURL := strings.TrimRight(strings.TrimSpace(cfg.BaseURL), "/")
	u, err := url.Parse(baseURL)
	if err != nil {
		return vadRequestParams{}, "", fmt.Errorf("parse vad ws url: %w", err)
	}
	if u.Scheme == "" {
		return vadRequestParams{}, "", fmt.Errorf("vad ws base_url missing scheme (need ws:// or wss://)")
	}
	if u.Path == "" || u.Path == "/" {
		u.Path = "/v1/vad/infer"
	}
	u.RawQuery = c.buildVADQueryValues(cfg, sessionID).Encode()
	timeoutMs := cfg.TimeoutMs
	if timeoutMs <= 0 {
		timeoutMs = 1500
	}
	params := vadRequestParams{
		Transport:            "ws",
		BaseURL:              baseURL,
		RequestURL:           u.String(),
		SessionID:            sessionID,
		Threshold:            u.Query().Get("threshold"),
		ThresholdLow:         u.Query().Get("threshold_low"),
		MinSilenceDurationMs: u.Query().Get("min_silence_duration_ms"),
		FrameWindowThreshold: u.Query().Get("frame_window_threshold"),
		TimeoutMs:            timeoutMs,
		APIKeySet:            strings.TrimSpace(cfg.APIKey) != "",
	}
	return params, u.String(), nil
}

func (c *VADClient) buildHTTPInferURL(cfg config.VADConfig, sessionID string) (vadRequestParams, string, error) {
	baseURL := strings.TrimRight(strings.TrimSpace(cfg.BaseURL), "/")
	u, err := url.Parse(baseURL + "/v1/vad/infer")
	if err != nil {
		return vadRequestParams{}, "", fmt.Errorf("parse vad url: %w", err)
	}
	u.RawQuery = c.buildVADQueryValues(cfg, sessionID).Encode()
	timeoutMs := cfg.TimeoutMs
	if timeoutMs <= 0 {
		timeoutMs = 1500
	}
	params := vadRequestParams{
		Transport:            "http",
		BaseURL:              baseURL,
		RequestURL:           u.String(),
		SessionID:            sessionID,
		Threshold:            u.Query().Get("threshold"),
		ThresholdLow:         u.Query().Get("threshold_low"),
		MinSilenceDurationMs: u.Query().Get("min_silence_duration_ms"),
		FrameWindowThreshold: u.Query().Get("frame_window_threshold"),
		TimeoutMs:            timeoutMs,
		APIKeySet:            strings.TrimSpace(cfg.APIKey) != "",
	}
	return params, u.String(), nil
}

func (c *VADClient) logConnectParamsOnce(params vadRequestParams) {
	if c.logger == nil {
		return
	}
	key := params.Transport + "|" + params.SessionID
	c.mu.Lock()
	if _, ok := c.logged[key]; ok {
		c.mu.Unlock()
		return
	}
	c.logged[key] = struct{}{}
	c.mu.Unlock()

	c.logger.Info("vad_connect_params",
		"transport", params.Transport,
		"base_url", params.BaseURL,
		"request_url", params.RequestURL,
		"session_id", params.SessionID,
		"threshold", params.Threshold,
		"threshold_low", params.ThresholdLow,
		"min_silence_duration_ms", params.MinSilenceDurationMs,
		"frame_window_threshold", params.FrameWindowThreshold,
		"timeout_ms", params.TimeoutMs,
		"api_key_set", params.APIKeySet,
	)
}

func (c *VADClient) getOrDialWS(ctx context.Context, params vadRequestParams, wsURL string, sessionID string, timeoutMs int, apiKey string) (*websocket.Conn, error) {
	key := wsURL + "|" + sessionID
	c.mu.Lock()
	existing := c.conns[key]
	c.mu.Unlock()
	if existing != nil {
		return existing, nil
	}

	c.logConnectParamsOnce(params)

	dialCtx := ctx
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		dialCtx, cancel = context.WithTimeout(ctx, time.Duration(timeoutMs)*time.Millisecond)
		defer cancel()
	}

	dialer := websocket.Dialer{HandshakeTimeout: time.Duration(timeoutMs) * time.Millisecond}
	header := http.Header{}
	if apiKey != "" {
		header.Set("Authorization", "Bearer "+apiKey)
	}
	conn, _, err := dialer.DialContext(dialCtx, wsURL, header)
	if err != nil {
		if c.logger != nil {
			c.logger.Warn("vad_ws_dial_failed",
				"session_id", sessionID,
				"request_url", params.RequestURL,
				"error", err,
			)
		}
		return nil, fmt.Errorf("vad ws dial failed: %w", err)
	}

	c.mu.Lock()
	// double-check：如果并发抢先放入了，关闭我们新建的并复用旧的
	if prev := c.conns[key]; prev != nil {
		c.mu.Unlock()
		_ = conn.Close()
		return prev, nil
	}
	c.conns[key] = conn
	c.mu.Unlock()
	return conn, nil
}

func (c *VADClient) dropWS(wsURL string, sessionID string, conn *websocket.Conn) {
	key := wsURL + "|" + sessionID
	c.mu.Lock()
	cur := c.conns[key]
	if cur == conn {
		delete(c.conns, key)
	}
	c.mu.Unlock()
	if conn != nil {
		_ = conn.Close()
	}
}

func effectiveInt(v int, fallback int) int {
	if v > 0 {
		return v
	}
	return fallback
}

func effectiveFloat(v float64, fallback float64) float64 {
	if v > 0 {
		return v
	}
	return fallback
}
