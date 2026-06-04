package adapter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"
)

const voiceprintLogBodyMax = 2048

// VoiceprintConfig configures the SV voiceprint register endpoint.
type VoiceprintConfig struct {
	RegisterURL string
	Token       string
	Timeout     time.Duration
}

// VoiceprintClient reports speaker enrollment to the SV voiceprint service.
type VoiceprintClient struct {
	cfg    VoiceprintConfig
	logger *zap.Logger
}

func NewVoiceprintClient(cfg VoiceprintConfig, logger *zap.Logger) *VoiceprintClient {
	if cfg.Timeout <= 0 {
		cfg.Timeout = 120 * time.Second
	}
	if logger == nil {
		logger = zap.NewNop()
	}
	return &VoiceprintClient{cfg: cfg, logger: logger}
}

// Enabled reports whether SV reporting is configured.
func (c *VoiceprintClient) Enabled() bool {
	return strings.TrimSpace(c.cfg.RegisterURL) != "" && strings.TrimSpace(c.cfg.Token) != ""
}

type voiceprintRegisterResp struct {
	SpeakerID string  `json:"speaker_id"`
	Score     float64 `json:"score"`
	Message   string  `json:"message"`
	Code      int     `json:"code"`
}

// Register enrolls WAV audio for the given speaker_id (SV /voiceprint/register).
func (c *VoiceprintClient) Register(ctx context.Context, speakerID string, wavData []byte) (*voiceprintRegisterResp, error) {
	if !c.Enabled() {
		return nil, fmt.Errorf("voiceprint not configured")
	}
	speakerID = strings.TrimSpace(speakerID)
	if speakerID == "" {
		return nil, fmt.Errorf("speaker_id is empty")
	}
	if len(wavData) == 0 {
		return nil, fmt.Errorf("audio is empty")
	}

	registerURL := strings.TrimSpace(c.cfg.RegisterURL)
	start := time.Now()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	if err := writer.WriteField("speaker_id", speakerID); err != nil {
		return nil, err
	}
	// SV FastAPI 注册接口要求文件字段名为 audio（非 file）
	part, err := writer.CreateFormFile("audio", "audio.wav")
	if err != nil {
		return nil, err
	}
	if _, err := part.Write(wavData); err != nil {
		return nil, err
	}
	if err := writer.Close(); err != nil {
		return nil, err
	}
	contentType := writer.FormDataContentType()

	c.logger.Info("sv voiceprint request preparing",
		zap.String("method", http.MethodPost),
		zap.String("url", registerURL),
		zap.String("speaker_id", speakerID),
		zap.Int("audio_bytes", len(wavData)),
		zap.Bool("wav_riff_header", hasRIFFHeader(wavData)),
		zap.String("form_field_speaker", "speaker_id"),
		zap.String("form_field_file", "audio"),
		zap.String("form_filename", "audio.wav"),
		zap.String("content_type", contentType),
		zap.Int("body_bytes", body.Len()),
		zap.Duration("timeout", c.cfg.Timeout),
		zap.String("authorization", maskBearerToken(c.cfg.Token)),
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, registerURL, &body)
	if err != nil {
		c.logRegisterFailed(start, speakerID, registerURL, 0, "", err)
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(c.cfg.Token))
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", contentType)

	c.logger.Info("sv voiceprint request sending",
		zap.String("url", registerURL),
		zap.String("speaker_id", speakerID),
	)

	client := &http.Client{Timeout: c.cfg.Timeout}
	resp, err := client.Do(req)
	elapsed := time.Since(start)
	if err != nil {
		c.logRegisterFailed(start, speakerID, registerURL, 0, "", err)
		return nil, fmt.Errorf("voiceprint request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	c.logger.Info("sv voiceprint response received",
		zap.String("url", registerURL),
		zap.String("speaker_id", speakerID),
		zap.Int("status_code", resp.StatusCode),
		zap.Duration("elapsed", elapsed),
		zap.Int("response_bytes", len(respBody)),
		zap.String("response_body", truncateForLog(string(respBody), voiceprintLogBodyMax)),
	)

	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("voiceprint status %d: %s", resp.StatusCode, string(respBody))
		c.logRegisterFailed(start, speakerID, registerURL, resp.StatusCode, string(respBody), err)
		return nil, err
	}

	var out voiceprintRegisterResp
	if len(respBody) > 0 {
		if err := json.Unmarshal(respBody, &out); err != nil {
			c.logRegisterFailed(start, speakerID, registerURL, resp.StatusCode, string(respBody), err)
			return nil, fmt.Errorf("decode voiceprint response: %w", err)
		}
	}
	if out.SpeakerID == "" {
		out.SpeakerID = speakerID
	}

	c.logger.Info("sv voiceprint register success",
		zap.String("url", registerURL),
		zap.String("request_speaker_id", speakerID),
		zap.String("response_speaker_id", out.SpeakerID),
		zap.Float64("score", out.Score),
		zap.Int("response_code", out.Code),
		zap.String("message", out.Message),
		zap.Duration("elapsed", elapsed),
		zap.Int("audio_bytes", len(wavData)),
	)
	return &out, nil
}

func (c *VoiceprintClient) logRegisterFailed(start time.Time, speakerID, url string, status int, respBody string, err error) {
	fields := []zap.Field{
		zap.String("url", url),
		zap.String("speaker_id", speakerID),
		zap.Duration("elapsed", time.Since(start)),
		zap.Error(err),
	}
	if status > 0 {
		fields = append(fields,
			zap.Int("status_code", status),
			zap.String("response_body", truncateForLog(respBody, voiceprintLogBodyMax)),
		)
	}
	c.logger.Warn("sv voiceprint register failed", fields...)
}

func maskBearerToken(token string) string {
	token = strings.TrimSpace(token)
	if token == "" {
		return "(empty)"
	}
	if len(token) <= 8 {
		return "****"
	}
	return token[:4] + "****" + token[len(token)-4:]
}

func truncateForLog(s string, max int) string {
	if max <= 0 || len(s) <= max {
		return s
	}
	return s[:max] + "...(truncated)"
}

func hasRIFFHeader(data []byte) bool {
	return len(data) >= 12 && string(data[0:4]) == "RIFF" && string(data[8:12]) == "WAVE"
}
