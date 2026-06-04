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

type DeviceAudioClient struct {
	store  *config.Store
	logger *slog.Logger
}

type DeviceAudioUploadResult struct {
	ID         string `json:"id"`
	MacID      string `json:"macId"`
	AudioLen   string `json:"audioLen"`
	Context    string `json:"context"`
	CreateTime string `json:"createTime"`
}

type adminEnvelope struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

func NewDeviceAudioClient(store *config.Store, logger *slog.Logger) *DeviceAudioClient {
	return &DeviceAudioClient{store: store, logger: logger}
}

func (c *DeviceAudioClient) Enabled() bool {
	cfg := c.store.Get().ChatbotAdmin
	return cfg.Enabled && cfg.UploadURL() != ""
}

func (c *DeviceAudioClient) AdminAPIReady() bool {
	return c.store.Get().ChatbotAdmin.APIBaseURL() != ""
}

type deviceSpeakerIDsResult struct {
	MacID      string   `json:"macId"`
	SpeakerIDs []string `json:"speakerIds"`
}

type deviceSpeakerNameResult struct {
	SpeakerID     string   `json:"speakerId"`
	SpeakerName   string   `json:"speakerName"`
	PersonalLines []string `json:"personalLines"`
}

// SpeakerProfile 音色名称与个性化条目（来自 chatbot-go-admin）。
type SpeakerProfile struct {
	Name          string
	PersonalLines []string
}

func (c *DeviceAudioClient) adminTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	cfg := c.store.Get().ChatbotAdmin
	timeoutMs := cfg.TimeoutMs
	if timeoutMs <= 0 {
		timeoutMs = 15000
	}
	return context.WithTimeout(ctx, time.Duration(timeoutMs)*time.Millisecond)
}

func (c *DeviceAudioClient) ListSpeakerIDs(ctx context.Context, macID string) ([]string, error) {
	cfg := c.store.Get().ChatbotAdmin
	baseURL := cfg.SpeakerIDsURL()
	if baseURL == "" {
		return nil, fmt.Errorf("chatbot_admin speaker-ids url is not configured")
	}
	macID = strings.TrimSpace(macID)
	if macID == "" {
		return nil, fmt.Errorf("mac_id is required")
	}

	reqCtx, cancel := c.adminTimeout(ctx)
	defer cancel()

	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("parse speaker-ids url: %w", err)
	}
	q := u.Query()
	q.Set("mac_id", macID)
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("create speaker-ids request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("speaker-ids request: %w", err)
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read speaker-ids response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("speaker-ids status %d: %s", resp.StatusCode, string(respBody))
	}

	var env adminEnvelope
	if err := json.Unmarshal(respBody, &env); err != nil {
		return nil, fmt.Errorf("decode speaker-ids envelope: %w", err)
	}
	if env.Code != 0 {
		return nil, fmt.Errorf("speaker-ids code %d: %s", env.Code, env.Message)
	}
	var result deviceSpeakerIDsResult
	if len(env.Data) > 0 && string(env.Data) != "null" {
		if err := json.Unmarshal(env.Data, &result); err != nil {
			return nil, fmt.Errorf("decode speaker-ids data: %w", err)
		}
	}
	return result.SpeakerIDs, nil
}

func (c *DeviceAudioClient) GetSpeakerProfile(ctx context.Context, speakerID string) (SpeakerProfile, error) {
	cfg := c.store.Get().ChatbotAdmin
	baseURL := cfg.SpeakerNameURL()
	if baseURL == "" {
		return SpeakerProfile{}, fmt.Errorf("chatbot_admin speaker-name url is not configured")
	}
	speakerID = strings.TrimSpace(speakerID)
	if speakerID == "" {
		return SpeakerProfile{}, fmt.Errorf("speaker_id is required")
	}

	reqCtx, cancel := c.adminTimeout(ctx)
	defer cancel()

	u, err := url.Parse(baseURL)
	if err != nil {
		return SpeakerProfile{}, fmt.Errorf("parse speaker-name url: %w", err)
	}
	q := u.Query()
	q.Set("speaker_id", speakerID)
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, u.String(), nil)
	if err != nil {
		return SpeakerProfile{}, fmt.Errorf("create speaker-name request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return SpeakerProfile{}, fmt.Errorf("speaker-name request: %w", err)
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return SpeakerProfile{}, fmt.Errorf("read speaker-name response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return SpeakerProfile{}, fmt.Errorf("speaker-name status %d: %s", resp.StatusCode, string(respBody))
	}

	var env adminEnvelope
	if err := json.Unmarshal(respBody, &env); err != nil {
		return SpeakerProfile{}, fmt.Errorf("decode speaker-name envelope: %w", err)
	}
	if env.Code != 0 {
		return SpeakerProfile{}, fmt.Errorf("speaker-name code %d: %s", env.Code, env.Message)
	}
	var result deviceSpeakerNameResult
	if len(env.Data) > 0 && string(env.Data) != "null" {
		if err := json.Unmarshal(env.Data, &result); err != nil {
			return SpeakerProfile{}, fmt.Errorf("decode speaker-name data: %w", err)
		}
	}
	lines := result.PersonalLines
	if lines == nil {
		lines = []string{}
	}
	return SpeakerProfile{
		Name:          strings.TrimSpace(result.SpeakerName),
		PersonalLines: lines,
	}, nil
}

func (c *DeviceAudioClient) GetSpeakerName(ctx context.Context, speakerID string) (string, error) {
	profile, err := c.GetSpeakerProfile(ctx, speakerID)
	if err != nil {
		return "", err
	}
	return profile.Name, nil
}

func (c *DeviceAudioClient) Upload(
	ctx context.Context,
	macID string,
	audioLenLabel string,
	contextText string,
	wav []byte,
) (*DeviceAudioUploadResult, error) {
	cfg := c.store.Get().ChatbotAdmin
	url := cfg.UploadURL()
	if url == "" {
		return nil, fmt.Errorf("chatbot_admin upload url is not configured")
	}
	macID = strings.TrimSpace(macID)
	if macID == "" {
		return nil, fmt.Errorf("mac_id is required")
	}
	audioLenLabel = strings.TrimSpace(audioLenLabel)
	if audioLenLabel == "" {
		return nil, fmt.Errorf("audio_len is required")
	}
	if len(wav) == 0 {
		return nil, fmt.Errorf("wav is empty")
	}

	timeoutMs := cfg.TimeoutMs
	if timeoutMs <= 0 {
		timeoutMs = 15000
	}
	reqCtx, cancel := context.WithTimeout(ctx, time.Duration(timeoutMs)*time.Millisecond)
	defer cancel()

	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	_ = w.WriteField("mac_id", macID)
	_ = w.WriteField("audio_len", audioLenLabel)
	_ = w.WriteField("context", contextText)
	part, err := w.CreateFormFile("audio_blob", "utterance.wav")
	if err != nil {
		return nil, fmt.Errorf("create form file: %w", err)
	}
	if _, err := part.Write(wav); err != nil {
		return nil, fmt.Errorf("write wav: %w", err)
	}
	if err := w.Close(); err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(reqCtx, http.MethodPost, url, &body)
	if err != nil {
		return nil, fmt.Errorf("create upload request: %w", err)
	}
	req.Header.Set("Content-Type", w.FormDataContentType())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("device audio upload: %w", err)
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read upload response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("device audio upload status %d: %s", resp.StatusCode, string(respBody))
	}

	var env adminEnvelope
	if err := json.Unmarshal(respBody, &env); err != nil {
		return nil, fmt.Errorf("decode upload envelope: %w", err)
	}
	if env.Code != 0 {
		return nil, fmt.Errorf("device audio upload code %d: %s", env.Code, env.Message)
	}
	var result DeviceAudioUploadResult
	if len(env.Data) > 0 && string(env.Data) != "null" {
		if err := json.Unmarshal(env.Data, &result); err != nil {
			return nil, fmt.Errorf("decode upload data: %w", err)
		}
	}
	if result.ID == "" {
		return nil, fmt.Errorf("device audio upload: empty record id")
	}
	return &result, nil
}
