package adapter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
	"time"

	"xiaozhi-go-core/internal/config"
)

type VoiceprintClient struct {
	store  *config.Store
	logger *slog.Logger
}

type voiceprintRuntime struct {
	apiURL    string
	apiKey    string
	timeout   time.Duration
	threshold float64
	enabled   bool
}

type VoiceprintIdentifyResult struct {
	SpeakerID string
	Score     float64
}

type voiceprintResponse struct {
	SpeakerID string  `json:"speaker_id"`
	Score     float64 `json:"score"`
}

func NewVoiceprintClient(store *config.Store, logger *slog.Logger) *VoiceprintClient {
	return &VoiceprintClient{store: store, logger: logger}
}

func (c *VoiceprintClient) runtime() voiceprintRuntime {
	cfg := c.store.Get().Voiceprint
	timeoutMs := cfg.TimeoutMs
	if timeoutMs <= 0 {
		timeoutMs = 10000
	}
	rt := voiceprintRuntime{
		timeout:   time.Duration(timeoutMs) * time.Millisecond,
		threshold: cfg.SimilarityThreshold,
		enabled:   cfg.Enabled,
	}
	if strings.TrimSpace(cfg.URL) == "" {
		return rt
	}
	apiURL, apiKey, err := resolveVoiceprintIdentifyURL(cfg.URL, cfg.APIKey)
	if err != nil {
		if c.logger != nil {
			c.logger.Warn("voiceprint_url_parse_failed", "url", cfg.URL, "error", err)
		}
		return rt
	}
	if apiKey == "" {
		if c.logger != nil {
			c.logger.Warn("voiceprint_api_key_missing")
		}
		return rt
	}
	rt.apiURL = apiURL
	rt.apiKey = apiKey
	return rt
}

// resolveVoiceprintIdentifyURL 支持 health 地址、identify 全路径或仅 host:port。
func resolveVoiceprintIdentifyURL(rawURL, cfgAPIKey string) (identifyURL, apiKey string, err error) {
	rawURL = strings.TrimSpace(rawURL)
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "", "", fmt.Errorf("parse voiceprint url: %w", err)
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return "", "", fmt.Errorf("voiceprint url missing scheme or host")
	}
	apiKey = strings.TrimSpace(cfgAPIKey)
	if apiKey == "" {
		apiKey = parsed.Query().Get("key")
	}
	path := strings.TrimRight(parsed.Path, "/")
	switch {
	case strings.HasSuffix(path, "/voiceprint/identify"):
		identifyURL = parsed.Scheme + "://" + parsed.Host + path
	case strings.HasSuffix(path, "/voiceprint/health"):
		identifyURL = parsed.Scheme + "://" + parsed.Host + "/voiceprint/identify"
	default:
		identifyURL = parsed.Scheme + "://" + parsed.Host + "/voiceprint/identify"
	}
	return identifyURL, apiKey, nil
}

func (c *VoiceprintClient) Enabled() bool {
	rt := c.runtime()
	return rt.enabled && rt.apiURL != "" && rt.apiKey != ""
}

func (c *VoiceprintClient) Identify(
	ctx context.Context,
	wavData []byte,
	speakerIDs []string,
) (VoiceprintIdentifyResult, error) {
	rt := c.runtime()
	if !rt.enabled || rt.apiURL == "" || rt.apiKey == "" || len(wavData) == 0 || len(speakerIDs) == 0 {
		return VoiceprintIdentifyResult{}, nil
	}

	httpClient := &http.Client{Timeout: rt.timeout}
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	if err := writer.WriteField("speaker_ids", strings.Join(speakerIDs, ",")); err != nil {
		return VoiceprintIdentifyResult{}, err
	}
	part, err := writer.CreateFormFile("audio", "audio.wav")
	if err != nil {
		return VoiceprintIdentifyResult{}, err
	}
	if _, err := part.Write(wavData); err != nil {
		return VoiceprintIdentifyResult{}, err
	}
	if err := writer.Close(); err != nil {
		return VoiceprintIdentifyResult{}, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, rt.apiURL, &body)
	if err != nil {
		return VoiceprintIdentifyResult{}, fmt.Errorf("create voiceprint request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+rt.apiKey)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := httpClient.Do(req)
	if err != nil {
		return VoiceprintIdentifyResult{}, fmt.Errorf("voiceprint request failed: %w", err)
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return VoiceprintIdentifyResult{}, fmt.Errorf("voiceprint status %d: %s", resp.StatusCode, string(respBody))
	}
	var payload voiceprintResponse
	if err := json.Unmarshal(respBody, &payload); err != nil {
		return VoiceprintIdentifyResult{}, fmt.Errorf("decode voiceprint response: %w", err)
	}
	threshold := rt.threshold
	if threshold <= 0 {
		threshold = 0.4
	}
	if payload.Score < threshold || strings.TrimSpace(payload.SpeakerID) == "" {
		return VoiceprintIdentifyResult{Score: payload.Score}, nil
	}
	return VoiceprintIdentifyResult{
		SpeakerID: strings.TrimSpace(payload.SpeakerID),
		Score:     payload.Score,
	}, nil
}
