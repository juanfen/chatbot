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
	"sync/atomic"
	"time"

	"xiaozhi-go-core/internal/config"
)

// TTSClient 仅对接 xiaozhi-server 中 IndexStreamTTS（index_stream.py）：
// POST JSON {"text","character"}，流式 16-bit LE PCM 分帧；Opus 在 xiaozhi-server 侧编码。
type TTSClient struct {
	store  *config.Store
	logger *slog.Logger
	nextIdx atomic.Uint64
}

func NewTTSClient(store *config.Store, logger *slog.Logger) *TTSClient {
	return &TTSClient{store: store, logger: logger}
}

func (c *TTSClient) ttsCfg() config.TTSConfig {
	return c.store.Get().TTS
}

func (c *TTSClient) Enabled() bool {
	cfg := c.ttsCfg()
	if !cfg.Enabled {
		return false
	}
	for _, svc := range cfg.Services {
		if svc.Pickable() {
			return true
		}
	}
	return false
}

func (c *TTSClient) MaxParallel() int {
	cfg := c.ttsCfg()
	if cfg.MaxParallel > 0 {
		return cfg.MaxParallel
	}
	return 4
}

func (c *TTSClient) SegmentMaxChars() int {
	cfg := c.ttsCfg()
	if cfg.SegmentMaxChars > 0 {
		return cfg.SegmentMaxChars
	}
	return 80
}

func (c *TTSClient) pickService(preferIndex int) (config.TTSServiceConfig, int, error) {
	cfg := c.ttsCfg()
	enabled := make([]int, 0, len(cfg.Services))
	for i := range cfg.Services {
		if cfg.Services[i].Pickable() {
			enabled = append(enabled, i)
		}
	}
	if len(enabled) == 0 {
		return config.TTSServiceConfig{}, -1, fmt.Errorf("no enabled IndexStream tts service (need api_url or url, and non-empty voice)")
	}
	if preferIndex >= 0 && preferIndex < len(enabled) {
		idx := enabled[preferIndex]
		svc := cfg.Services[idx]
		c.logger.Warn("TTS_ROUTE_SELECTED",
			"reason", "prefer_index",
			"prefer_index", preferIndex,
			"enabled_pool_size", len(enabled),
			"service_index", idx,
			"service", svc.Name,
			"url", svc.ResolveURL(),
			"voice", svc.Voice,
		)
		return cfg.Services[idx], idx, nil
	}
	rr := int(c.nextIdx.Add(1)-1) % len(enabled)
	idx := enabled[rr]
	svc := cfg.Services[idx]
	c.logger.Warn("TTS_ROUTE_SELECTED",
		"reason", "round_robin",
		"round_robin_slot", rr,
		"enabled_pool_size", len(enabled),
		"service_index", idx,
		"service", svc.Name,
		"url", svc.ResolveURL(),
		"voice", svc.Voice,
	)
	return cfg.Services[idx], idx, nil
}

// SynthesizeStream 会边读取上游 PCM 流边回调 onFrame（每次一帧 16-bit LE PCM）。
// 这样可以避免把整段音频读入内存，并显著降低并发场景的尾延迟与 GC 压力。
func (c *TTSClient) SynthesizeStream(
	ctx context.Context,
	text string,
	preferIndex int,
	onFrame func([]byte) error,
) (string, error) {
	return c.synthesizeStreamImpl(ctx, text, preferIndex, onFrame)
}

// SynthesizeStreamWithBackpressure 与 SynthesizeStream 行为一致；保留 API 兼容。
func (c *TTSClient) SynthesizeStreamWithBackpressure(
	ctx context.Context,
	text string,
	preferIndex int,
	onFrame func([]byte) error,
) (string, error) {
	return c.synthesizeStreamImpl(ctx, text, preferIndex, onFrame)
}

func (c *TTSClient) synthesizeStreamImpl(
	ctx context.Context,
	text string,
	preferIndex int,
	onFrame func([]byte) error,
) (string, error) {
	if onFrame == nil {
		return "", fmt.Errorf("onFrame is nil")
	}
	svc, idx, err := c.pickService(preferIndex)
	if err != nil {
		return "", err
	}
	timeoutMs := svc.TimeoutMs
	if timeoutMs <= 0 {
		timeoutMs = 15000
	}
	reqCtx, cancel := context.WithTimeout(ctx, time.Duration(timeoutMs)*time.Millisecond)
	defer cancel()

	body := map[string]string{
		"text":      text,
		"character": svc.Voice,
	}
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("marshal index_stream tts body: %w", err)
	}
	c.logger.Info("tts_request_start",
		"service", svc.Name,
		"service_index", idx,
		"url", svc.ResolveURL(),
		"voice", svc.Voice,
		"text_preview", previewTextForTTS(text, 200),
		"text_runes", len([]rune(text)),
		"timeout_ms", timeoutMs,
	)

	req, err := http.NewRequestWithContext(reqCtx, http.MethodPost, svc.ResolveURL(), bytes.NewReader(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("create tts request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if svc.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+svc.APIKey)
	}
	for k, v := range svc.Headers {
		req.Header.Set(k, v)
	}

	httpCl := &http.Client{Timeout: time.Duration(timeoutMs) * time.Millisecond}
	resp, err := httpCl.Do(req)
	if err != nil {
		return "", fmt.Errorf("index_stream tts request failed (service=%s idx=%d): %w", svc.Name, idx, err)
	}
	defer resp.Body.Close()
	c.logger.Info("tts_response_headers",
		"service", svc.Name,
		"service_index", idx,
		"status_code", resp.StatusCode,
		"content_type", resp.Header.Get("Content-Type"),
	)
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("index_stream tts status %d (service=%s idx=%d): %s", resp.StatusCode, svc.Name, idx, string(b))
	}

	sampleRate := svc.EffectiveSampleRateHz()
	channels := svc.Channels
	if channels <= 0 {
		channels = 1
	}
	frameMs := svc.FrameDurationMs
	if frameMs <= 0 {
		frameMs = 60
	}

	route := strings.TrimSpace(svc.Name)
	if route == "" {
		route = fmt.Sprintf("route-%d", idx)
	}

	pcmReader, strippedWAV, err := maybePCMReaderFromHTTPBody(resp.Header.Get("Content-Type"), resp.Body)
	if err != nil {
		return "", fmt.Errorf("tts body open (service=%s idx=%d): %w", svc.Name, idx, err)
	}
	if strippedWAV {
		c.logger.Info("tts_wav_container_stripped",
			"service", svc.Name,
			"service_index", idx,
			"content_type", resp.Header.Get("Content-Type"),
		)
	}

	totalBytes, encErr := pcmStreamToS16FramesStream(pcmReader, sampleRate, channels, frameMs, onFrame)
	if encErr != nil {
		return "", fmt.Errorf("pcm frame split (service=%s idx=%d): %w", svc.Name, idx, encErr)
	}

	c.logger.Info("tts_request_done",
		"service", svc.Name,
		"service_index", idx,
		"route", route,
		"total_pcm_bytes", totalBytes,
		"sample_rate_hz", sampleRate,
		"channels", channels,
		"frame_duration_ms", frameMs,
	)
	return route, nil
}

// Synthesize 返回多帧 16-bit LE PCM；上层将 AUDIO_CHUNK 标为 pcm_s16le。
func (c *TTSClient) Synthesize(ctx context.Context, text string, preferIndex int) ([][]byte, string, error) {
	var packets [][]byte
	route, err := c.SynthesizeStream(ctx, text, preferIndex, func(frame []byte) error {
		packets = append(packets, append([]byte(nil), frame...))
		return nil
	})
	if err != nil {
		return nil, "", err
	}
	return packets, route, nil
}

func pcmStreamToS16Frames(r io.Reader, sampleRate int, channels int, frameMs int) ([][]byte, error) {
	samplesPerChannelPerFrame := sampleRate * frameMs / 1000
	totalSamples := samplesPerChannelPerFrame * channels
	frameBytes := totalSamples * 2
	if frameBytes <= 0 {
		return nil, fmt.Errorf("invalid frame: sampleRate=%d channels=%d frameMs=%d", sampleRate, channels, frameMs)
	}

	var out [][]byte
	buf := make([]byte, 0, frameBytes*4)
	readBuf := make([]byte, 4096)
	tail := make([]byte, frameBytes)

	for {
		n, readErr := r.Read(readBuf)
		if n > 0 {
			buf = append(buf, readBuf[:n]...)
		}
		for len(buf) >= frameBytes {
			out = append(out, append([]byte(nil), buf[:frameBytes]...))
			buf = buf[frameBytes:]
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return out, readErr
		}
	}

	if len(buf) == 0 {
		return out, nil
	}

	if len(buf)%2 != 0 {
		buf = append(buf, 0)
	}
	copy(tail, buf)
	for i := len(buf); i < frameBytes; i++ {
		tail[i] = 0
	}
	out = append(out, append([]byte(nil), tail...))
	return out, nil
}

func pcmStreamToS16FramesStream(
	r io.Reader,
	sampleRate int,
	channels int,
	frameMs int,
	onFrame func([]byte) error,
) (int, error) {
	samplesPerChannelPerFrame := sampleRate * frameMs / 1000
	totalSamples := samplesPerChannelPerFrame * channels
	frameBytes := totalSamples * 2
	if frameBytes <= 0 {
		return 0, fmt.Errorf("invalid frame: sampleRate=%d channels=%d frameMs=%d", sampleRate, channels, frameMs)
	}
	totalBytes := 0

	buf := make([]byte, 0, frameBytes*4)
	readBuf := make([]byte, 4096)
	tail := make([]byte, frameBytes)

	for {
		n, readErr := r.Read(readBuf)
		if n > 0 {
			buf = append(buf, readBuf[:n]...)
		}
		for len(buf) >= frameBytes {
			frame := append([]byte(nil), buf[:frameBytes]...)
			buf = buf[frameBytes:]
			totalBytes += len(frame)
			if err := onFrame(frame); err != nil {
				return totalBytes, err
			}
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return totalBytes, readErr
		}
	}

	if len(buf) == 0 {
		return totalBytes, nil
	}
	if len(buf)%2 != 0 {
		buf = append(buf, 0)
	}
	copy(tail, buf)
	for i := len(buf); i < frameBytes; i++ {
		tail[i] = 0
	}
	frame := append([]byte(nil), tail...)
	totalBytes += len(frame)
	if err := onFrame(frame); err != nil {
		return totalBytes, err
	}
	return totalBytes, nil
}

func previewTextForTTS(s string, max int) string {
	s = strings.TrimSpace(s)
	if max <= 0 {
		max = 200
	}
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	return string(r[:max]) + "...(truncated)"
}
