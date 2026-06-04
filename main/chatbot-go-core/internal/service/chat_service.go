package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"sync/atomic"
	"time"
	"unicode"
	"unicode/utf8"

	"xiaozhi-go-core/internal/adapter"
	"xiaozhi-go-core/internal/config"
	"xiaozhi-go-core/internal/domain"
	"xiaozhi-go-core/internal/observability"
	"xiaozhi-go-core/internal/pb/chatv1"
)

type ChatService struct {
	llm                *adapter.LLMClient
	tts                *adapter.TTSClient
	memory             *adapter.ShortTermMemoryClient
	store              *config.Store
	logger             *slog.Logger
	maxRetries int
	inflight   atomic.Int64
}

type metadataPayload struct {
	Functions []map[string]interface{} `json:"functions"`
	TTSConfig *ttsRuntimeConfig        `json:"tts_config,omitempty"`
}

type ttsRuntimeConfig struct {
	Enabled         bool   `json:"enabled"`
	Streams         int    `json:"streams"`
	Codec           string `json:"codec"`
	SampleRateHz    uint32 `json:"sample_rate_hz"`
	Channels        uint32 `json:"channels"`
	FrameDurationMs uint32 `json:"frame_duration_ms"`
}

func NewChatService(store *config.Store, logger *slog.Logger) *ChatService {
	return &ChatService{
		llm:        adapter.NewLLMClient(store, logger),
		tts:        adapter.NewTTSClient(store, logger),
		memory:     adapter.NewShortTermMemoryClient(store, logger),
		store:      store,
		logger:     logger,
		maxRetries: 1,
	}
}

func (s *ChatService) Inflight() int64      { return s.inflight.Load() }
func (s *ChatService) Logger() *slog.Logger { return s.logger }

// StreamChat processes one request. It handles the complete LLM conversation
// including tool_call detection. When tool_calls are detected, the service
// sends TOOL_CALL events and a FINISH with reason "tool_calls", expecting
// the Python gateway to execute tools and re-invoke StreamChat with results.
func (s *ChatService) StreamChat(req *chatv1.StreamChatRequest, send func(*chatv1.StreamChatResponse) error) error {
	s.inflight.Add(1)
	defer s.inflight.Add(-1)
	reqLogger := s.logger.With(
		"request_id", req.RequestId,
		"session_id", req.SessionId,
		"device_id", req.DeviceId,
	)
	reqTrace, traceErr := observability.NewRequestTrace(req.RequestId)
	if traceErr != nil {
		reqLogger.Warn("request_trace_init_failed", "error", traceErr)
	}
	defer reqTrace.Close()
	logMilestone := func(stage string, attrs ...any) {
		reqLogger.Info(stage, attrs...)
		reqTrace.Log(stage, attrs...)
	}

	srv := s.store.Get().Server
	if srv.MaxConcurrency > 0 && s.inflight.Load() > int64(srv.MaxConcurrency) {
		return s.sendError(send, "CHAT_OVERLOADED", fmt.Sprintf("concurrency limit %d exceeded", srv.MaxConcurrency))
	}

	logMilestone("chat_request_in",
		"query_len", len(req.Query),
		"msg_count", len(req.Messages),
		"intent_type", req.IntentType,
		"inflight", s.inflight.Load(),
	)

	if req.Query == "" && len(req.Messages) == 0 {
		return s.sendError(send, "CHAT_INVALID_ARGUMENT", "query and messages are both empty")
	}

	ctx, touchActivity, stopStream := streamContext(srv.RequestTimeoutMs, srv.StreamIdleTimeoutMs)
	defer stopStream()

	ttsFlush := s.store.Get().TTS
	flushMinChars := ttsFlush.FlushMinChars
	if flushMinChars <= 0 {
		flushMinChars = 40
	}
	flushIntervalMs := ttsFlush.FlushIntervalMs
	if flushIntervalMs <= 0 {
		flushIntervalMs = 600
	}
	flushStrongSuffixes := config.EffectiveFlushStrongSuffixes(&ttsFlush)
	flushWeakSuffixes := config.EffectiveFlushWeakSuffixes(&ttsFlush)

	functions, ttsConfig := s.parseMetadata(req.MetadataJson)
	msgs := s.toDomainMessages(req.Messages)
	stmCfg := s.store.Get().ShortTermMemory
	stmQuery := effectiveChatQuery(req)
	stmUserID := resolveShortTermUserID(req, &stmCfg)
	stmUsed := false
	if stmCfg.Enabled && strings.TrimSpace(stmCfg.BaseURL) != "" && stmQuery != "" &&
		strings.TrimSpace(req.SessionId) != "" {
		built, err := s.memory.BuildPrompt(ctx, stmUserID, req.DeviceId, req.SessionId, stmQuery)
		if err != nil {
			reqLogger.Warn("short_term_build_prompt_failed", "error", err)
		} else if len(built) > 0 {
			msgs = mergeGatewaySystemPromptAtHead(built, s.store.Get().Gateway.SystemPrompt)
			stmUsed = true
			logMilestone("short_term_build_prompt_ok", "messages_count", len(msgs))
		}
	}

	opts := adapter.StreamChatOpts{
		Messages:  msgs,
		Functions: functions,
	}
	logMilestone("llm_call_start",
		"messages_count", len(opts.Messages),
		"functions_count", len(functions),
		"tts_enabled", ttsConfig.Enabled,
		"tts_streams", ttsConfig.Streams,
		"tts_codec", ttsConfig.Codec,
		"tts_sample_rate_hz", ttsConfig.SampleRateHz,
		"tts_channels", ttsConfig.Channels,
		"tts_frame_duration_ms", ttsConfig.FrameDurationMs,
		"metadata_json_len", len(req.MetadataJson),
		"query_preview", previewText(req.Query, 120),
	)

	chunkCh, err := s.llm.StreamChatWithRetry(ctx, opts, s.maxRetries)
	if err != nil {
		return s.sendError(send, "CHAT_UPSTREAM_UNAVAILABLE", err.Error())
	}

	startTime := time.Now()
	firstChunk := true
	llmChunkCount := 0
	outputChars := 0
	maxOutput := int(req.MaxOutputSize)
	audioStarted := false
	streamCounter := 0
	seqByStream := map[string]uint64{}
	audioChunkCount := 0
	ttsBuffer := strings.Builder{}
	ttsBufferRunes := 0
	lastTTSFlushAt := time.Now()
	normalizedOutput := strings.Builder{}

	flushTTSBuffer := func(reason string, force bool) error {
		if !ttsConfig.Enabled {
			return nil
		}
		if ttsBufferRunes == 0 {
			return nil
		}
		bufferText := ttsBuffer.String()
		if !force {
			if !shouldFlushTTSBuffer(
				bufferText,
				ttsBufferRunes,
				lastTTSFlushAt,
				s.tts.SegmentMaxChars(),
				flushMinChars,
				time.Duration(flushIntervalMs)*time.Millisecond,
				flushStrongSuffixes,
				flushWeakSuffixes,
			) {
				return nil
			}
		}
		if !audioStarted {
			audioStarted = true
			if err := send(&chatv1.StreamChatResponse{
				Event:           chatv1.StreamEvent_STREAM_EVENT_TTS_START,
				Codec:           ttsConfig.Codec,
				SampleRateHz:    ttsConfig.SampleRateHz,
				Channels:        ttsConfig.Channels,
				FrameDurationMs: ttsConfig.FrameDurationMs,
			}); err != nil {
				return err
			}
		}
		touchActivity()
		logMilestone("tts_call_start",
			"reason", reason,
			"runes", ttsBufferRunes,
			"preview", previewText(bufferText, 120),
		)
		emitted, err := s.emitAudioChunks(ctx, reqLogger, reqTrace, req.RequestId, send, bufferText, ttsConfig, &streamCounter, seqByStream, touchActivity)
		if err != nil {
			return err
		}
		audioChunkCount += emitted
		logMilestone("tts_call_end",
			"reason", reason,
			"audio_chunks_emitted", emitted,
		)
		ttsBuffer.Reset()
		ttsBufferRunes = 0
		lastTTSFlushAt = time.Now()
		return nil
	}

	for chunk := range chunkCh {
		llmChunkCount++
		if firstChunk {
			logMilestone("llm_first_chunk",
				"first_chunk_ms", time.Since(startTime).Milliseconds(),
			)
			firstChunk = false
		}

		switch chunk.Event {
		case "CONTENT_CHUNK":
			content := normalizeStreamingDelta(normalizedOutput.String(), chunk.ContentChunk)
			if content == "" {
				continue
			}
			normalizedOutput.WriteString(content)
			if llmChunkCount%20 == 0 {
				logMilestone("llm_stream_progress",
					"chunks", llmChunkCount,
					"output_chars", outputChars,
				)
			}
			if maxOutput > 0 {
				remaining := maxOutput - outputChars
				if remaining <= 0 {
					continue
				}
				charCount := utf8.RuneCountInString(content)
				if charCount > remaining {
					runes := []rune(content)
					content = string(runes[:remaining])
				}
				outputChars += utf8.RuneCountInString(content)
			}
			// 先发送 TTS_START，再发送首个 CONTENT_CHUNK，
			// 以兼容 Python 侧 disable_python_fallback=true 的严格校验。
			if ttsConfig.Enabled && content != "" && !audioStarted {
				audioStarted = true
				if err := send(&chatv1.StreamChatResponse{
					Event:           chatv1.StreamEvent_STREAM_EVENT_TTS_START,
					Codec:           ttsConfig.Codec,
					SampleRateHz:    ttsConfig.SampleRateHz,
					Channels:        ttsConfig.Channels,
					FrameDurationMs: ttsConfig.FrameDurationMs,
				}); err != nil {
					return err
				}
			}
			if err := send(&chatv1.StreamChatResponse{
				Event:        chatv1.StreamEvent_STREAM_EVENT_CONTENT_CHUNK,
				ContentChunk: content,
			}); err != nil {
				return err
			}
			if ttsConfig.Enabled && content != "" {
				ttsBuffer.WriteString(content)
				ttsBufferRunes += utf8.RuneCountInString(content)
				if err := flushTTSBuffer("auto", false); err != nil {
					return err
				}
			} else if content != "" {
				touchActivity()
			}

		case "TOOL_CALLS":
			if err := flushTTSBuffer("tool_calls", true); err != nil {
				return err
			}
			logMilestone("llm_tool_calls",
				"tool_calls_count", len(chunk.ToolCalls),
			)
			for _, tc := range chunk.ToolCalls {
				if err := send(&chatv1.StreamChatResponse{
					Event: chatv1.StreamEvent_STREAM_EVENT_TOOL_CALL,
					ToolCall: &chatv1.ToolCall{
						Id:            tc.ID,
						Name:          tc.Function.Name,
						ArgumentsJson: tc.Function.Arguments,
					},
				}); err != nil {
					return err
				}
			}
			logMilestone("llm_call_end",
				"finish_reason", "tool_calls",
				"llm_chunks", llmChunkCount,
				"tool_count", len(chunk.ToolCalls),
				"total_ms", time.Since(startTime).Milliseconds(),
			)
			if audioStarted {
				if err := send(&chatv1.StreamChatResponse{
					Event: chatv1.StreamEvent_STREAM_EVENT_TTS_END,
				}); err != nil {
					return err
				}
				logMilestone("tts_stream_end",
					"audio_chunk_count", audioChunkCount,
				)
			}
			logMilestone("chat_response_out",
				"finish_reason", "tool_calls",
				"total_ms", time.Since(startTime).Milliseconds(),
			)
			return send(&chatv1.StreamChatResponse{
				Event:        chatv1.StreamEvent_STREAM_EVENT_FINISH,
				FinishReason: "tool_calls",
			})

		case "ERROR":
			return s.sendError(send, chunk.ErrorCode, chunk.ErrorMessage)
		}
	}

	if err := flushTTSBuffer("finish", true); err != nil {
		return err
	}
	fullAnswer := normalizedOutput.String()
	if stmUsed && strings.TrimSpace(fullAnswer) != "" {
		commitMs := stmCfg.TimeoutMs
		if commitMs <= 0 {
			commitMs = 30000
		}
		cctx, cancel := context.WithTimeout(context.Background(), time.Duration(commitMs)*time.Millisecond)
		err := s.memory.CommitTurn(cctx, stmUserID, req.DeviceId, req.SessionId, stmQuery, fullAnswer)
		cancel()
		if err != nil {
			reqLogger.Warn("short_term_commit_turn_failed", "error", err)
		} else {
			logMilestone("short_term_commit_turn_ok",
				"answer_runes", utf8.RuneCountInString(fullAnswer),
			)
		}
	}
	logMilestone("llm_call_end",
		"finish_reason", "stop",
		"llm_chunks", llmChunkCount,
		"output_chars", outputChars,
		"total_ms", time.Since(startTime).Milliseconds(),
	)
	if audioStarted {
		if err := send(&chatv1.StreamChatResponse{
			Event: chatv1.StreamEvent_STREAM_EVENT_TTS_END,
		}); err != nil {
			return err
		}
		logMilestone("tts_stream_end",
			"audio_chunk_count", audioChunkCount,
		)
	}
	logMilestone("chat_response_out",
		"finish_reason", "stop",
		"total_ms", time.Since(startTime).Milliseconds(),
	)
	return send(&chatv1.StreamChatResponse{
		Event:        chatv1.StreamEvent_STREAM_EVENT_FINISH,
		FinishReason: "stop",
	})
}

func (s *ChatService) sendError(send func(*chatv1.StreamChatResponse) error, code, message string) error {
	return send(&chatv1.StreamChatResponse{
		Event:        chatv1.StreamEvent_STREAM_EVENT_ERROR,
		ErrorCode:    code,
		ErrorMessage: message,
		FinishReason: "error",
	})
}

func (s *ChatService) parseMetadata(metadata string) ([]map[string]interface{}, ttsRuntimeConfig) {
	defaultTTS := ttsRuntimeConfig{
		Enabled:         false,
		Streams:         2,
		Codec:           "opus",
		SampleRateHz:    24000,
		Channels:        1,
		FrameDurationMs: 60,
	}
	if metadata == "" {
		s.logger.Info("parse_metadata_empty")
		return nil, defaultTTS
	}
	var payload metadataPayload
	if err := json.Unmarshal([]byte(metadata), &payload); err != nil {
		s.logger.Warn("parse_metadata_failed", "error", err.Error(), "metadata_preview", previewText(metadata, 300))
		return nil, defaultTTS
	}
	if payload.TTSConfig == nil {
		s.logger.Info("parse_metadata_no_tts_config", "functions_count", len(payload.Functions))
		return payload.Functions, defaultTTS
	}
	cfg := *payload.TTSConfig
	if cfg.Streams <= 0 {
		cfg.Streams = 2
	}
	if cfg.Codec == "" {
		cfg.Codec = "opus"
	}
	if cfg.SampleRateHz == 0 {
		cfg.SampleRateHz = 24000
	}
	if cfg.Channels == 0 {
		cfg.Channels = 1
	}
	if cfg.FrameDurationMs == 0 {
		cfg.FrameDurationMs = 60
	}
	s.logger.Info("parse_metadata_ok",
		"functions_count", len(payload.Functions),
		"tts_enabled", cfg.Enabled,
		"tts_streams", cfg.Streams,
		"tts_codec", cfg.Codec,
		"tts_sample_rate_hz", cfg.SampleRateHz,
		"tts_channels", cfg.Channels,
		"tts_frame_duration_ms", cfg.FrameDurationMs,
	)
	return payload.Functions, cfg
}

func (s *ChatService) emitAudioChunks(
	ctx context.Context,
	reqLogger *slog.Logger,
	reqTrace *observability.RequestTrace,
	requestID string,
	send func(*chatv1.StreamChatResponse) error,
	content string,
	cfg ttsRuntimeConfig,
	streamCounter *int,
	seqByStream map[string]uint64,
	touchActivity func(),
) (int, error) {
	if !s.tts.Enabled() || !cfg.Enabled {
		return 0, nil
	}
	if touchActivity != nil {
		touchActivity()
	}
	segments := splitTextForTTS(content, s.tts.SegmentMaxChars())
	if len(segments) == 0 {
		return 0, nil
	}
	reqLogger.Info("tts_emit_start",
		"segments_count", len(segments),
		"content_preview", previewText(content, 120),
	)
	reqTrace.Log("tts_emit_start",
		"segments_count", len(segments),
		"content_preview", previewText(content, 120),
	)
	maxParallel := s.tts.MaxParallel()
	if cfg.Streams > 0 && cfg.Streams < maxParallel {
		maxParallel = cfg.Streams
	}
	if maxParallel <= 0 {
		maxParallel = 1
	}
	type segFrame struct {
		frame []byte
		err   error
		route string
	}

	// 计算最优缓冲大小：基于音频帧持续时间和采样率
	// 目标：允许 1-2 帧的缓冲空间，避免背压同时防止堆积
	frameDurationMs := int(cfg.FrameDurationMs)
	if frameDurationMs <= 0 {
		frameDurationMs = 60
	}
	sampleRateHz := int(cfg.SampleRateHz)
	if sampleRateHz == 0 {
		sampleRateHz = 24000
	}
	channels := int(cfg.Channels)
	if channels <= 0 {
		channels = 1
	}

	// 每帧字节数 = sampleRate * channels * 2 (16-bit) / 1000 * frameDuration
	bytesPerFrame := sampleRateHz * channels * 2 / 1000 * frameDurationMs
	// 缓冲大小（按帧数）：
	// 之前的 2~6 帧在网络/浏览器稍有抖动时会频繁触发 buffer 满并丢帧，导致“只听到零碎字”。
	// 这里改为更大的缓冲，配合阻塞写入形成自然 backpressure，避免丢帧。
	bufferFrames := 60 // 60ms/frame => ~3.6s 缓冲
	if bytesPerFrame >= 7680 {
		bufferFrames = 40 // 大帧每帧更大，适当降低帧数避免占用过多内存
	}
	if bufferFrames < 20 {
		bufferFrames = 20
	}
	if bufferFrames > 120 {
		bufferFrames = 120
	}
	bufferSize := bufferFrames
	
	reqLogger.Info("tts_emit_config",
		"segments_count", len(segments),
		"max_parallel", maxParallel,
		"buffer_frames", bufferSize,
		"bytes_per_frame", bytesPerFrame,
		"estimated_buffer_mb", float64(bufferSize)*float64(len(segments))*float64(bytesPerFrame)/1024/1024,
	)

	sem := make(chan struct{}, maxParallel)
	streamIDBySeg := make([]string, len(segments))
	frameChBySeg := make([]chan segFrame, len(segments))

	for segIdx := range segments {
		streamIDBySeg[segIdx] = fmt.Sprintf("stream-%d", segIdx%maxParallel)
		frameChBySeg[segIdx] = make(chan segFrame, bufferSize)
	}

		for segIdx, segText := range segments {
		segIdx, segText := segIdx, segText
		streamID := streamIDBySeg[segIdx]
		ch := frameChBySeg[segIdx]
		go func() {
			sem <- struct{}{}
			defer func() { <-sem }()
			defer close(ch)

			reqLogger.Info("tts_segment_begin", "segment_index", segIdx, "segment_preview", previewText(segText, 80))
			reqTrace.Log("tts_segment_begin", "segment_index", segIdx, "segment_preview", previewText(segText, 80))

			// 使用带超时重试的 TTS 合成，避免永久阻塞
			route, err := s.tts.SynthesizeStreamWithBackpressure(ctx, segText, -1, func(frame []byte) error {
				if touchActivity != nil && len(frame) > 0 {
					touchActivity()
				}
				// 阻塞写入：避免缓冲满时丢帧造成音频“跳跃/缺字”。
				// 当连接关闭/请求取消时 ctx 会结束，此处能及时退出。
				select {
				case ch <- segFrame{frame: frame}:
					return nil
				case <-ctx.Done():
					return ctx.Err()
				}
			})
			if err != nil {
				ch <- segFrame{err: err}
				return
			}
			if route != "" {
				// route 仅用于观测；流式帧优先下发，route 可在结束时再回传
				ch <- segFrame{route: route}
				reqLogger.Info("tts_route_hit",
					"request_id", requestID,
					"segment_index", segIdx,
					"prefer_index", -1,
					"stream_id", streamID,
					"route", route,
					"segment_preview", previewText(segText, 120),
				)
				reqTrace.Log("tts_route_hit",
					"request_id", requestID,
					"segment_index", segIdx,
					"stream_id", streamID,
					"route", route,
				)
			}
		}()
	}

	emitted := 0
	var firstErr error

	// 按 segment 顺序输出，保证最终音频播放不乱序（即使 TTS 并发生成）。
	for segIdx := 0; segIdx < len(segments); segIdx++ {
		streamID := streamIDBySeg[segIdx]
		ch := frameChBySeg[segIdx]

		var pendingRoute string
		var prev []byte
		havePrev := false

		for item := range ch {
			if item.route != "" && pendingRoute == "" {
				pendingRoute = item.route
				continue
			}
			if item.err != nil {
				reqLogger.Error("tts_synthesize_failed", "segment", segIdx, "error", item.err)
				reqTrace.Log("tts_synthesize_failed", "segment", segIdx, "error", item.err)
				if sendErr := send(&chatv1.StreamChatResponse{
					Event:        chatv1.StreamEvent_STREAM_EVENT_TTS_ERROR,
					TtsStreamId:  streamID,
					ErrorCode:    "TTS_SYNTH_FAILED",
					ErrorMessage: item.err.Error(),
				}); sendErr != nil && firstErr == nil {
					firstErr = sendErr
				}
				// 出错后仍把当前 segment 的 channel 消耗完（goroutine 会 close），继续后续 segment
				continue
			}
			if len(item.frame) == 0 {
				continue
			}
			if !havePrev {
				prev = item.frame
				havePrev = true
				continue
			}

			seq := seqByStream[streamID]
			seqByStream[streamID] = seq + 1
			if err := send(&chatv1.StreamChatResponse{
				Event:           chatv1.StreamEvent_STREAM_EVENT_AUDIO_CHUNK,
				TtsStreamId:     streamID,
				ChunkSeq:        seq,
				IsLastChunk:     false,
				Codec:           "pcm_s16le",
				SampleRateHz:    cfg.SampleRateHz,
				Channels:        cfg.Channels,
				FrameDurationMs: cfg.FrameDurationMs,
				AudioChunk:      prev,
			}); err != nil {
				return emitted, err
			}
			if touchActivity != nil {
				touchActivity()
			}
			emitted++
			prev = item.frame
		}

		if havePrev {
			seq := seqByStream[streamID]
			seqByStream[streamID] = seq + 1
			if err := send(&chatv1.StreamChatResponse{
				Event:           chatv1.StreamEvent_STREAM_EVENT_AUDIO_CHUNK,
				TtsStreamId:     streamID,
				ChunkSeq:        seq,
				IsLastChunk:     true,
				Codec:           "pcm_s16le",
				SampleRateHz:    cfg.SampleRateHz,
				Channels:        cfg.Channels,
				FrameDurationMs: cfg.FrameDurationMs,
				AudioChunk:      prev,
			}); err != nil {
				return emitted, err
			}
			if touchActivity != nil {
				touchActivity()
			}
			reqLogger.Info("tts_audio_chunk_sent",
				"stream_id", streamID,
				"segment_index", segIdx,
				"chunk_seq", seq,
				"is_last_chunk", true,
				"bytes", len(prev),
				"route", pendingRoute,
			)
			reqTrace.Log("tts_audio_chunk_sent",
				"stream_id", streamID,
				"segment_index", segIdx,
				"chunk_seq", seq,
				"is_last_chunk", true,
				"bytes", len(prev),
				"route", pendingRoute,
			)
			emitted++
		}
	}

	return emitted, firstErr
}

func previewText(s string, max int) string {
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

// splitTextForTTS splits text into segments optimized for TTS naturalness.
// It prioritizes splitting at sentence boundaries (punctuation) to avoid
// cutting words mid-sentence, which would cause unnatural TTS pauses.
// 不对分段结果做 TrimSpace，避免误删首尾标点或与标点相邻的空白展示。
func splitTextForTTS(text string, maxChars int) []string {
	if strings.TrimSpace(text) == "" {
		return nil
	}
	if maxChars <= 0 {
		maxChars = 80
	}

	runes := []rune(text)
	if len(runes) <= maxChars {
		return []string{text}
	}

	var segments []string
	start := 0

	for start < len(runes) {
		remaining := len(runes) - start
		if remaining <= maxChars {
			segment := string(runes[start:])
			if segment != "" {
				segments = append(segments, segment)
			}
			break
		}

		end := start + maxChars
		
		// Find the best split point within [start, end)
		splitPoint := findBestSplitPoint(runes, start, min(end, len(runes)))

		if splitPoint == -1 {
			// Fallback: hard cut at maxChars
			segment := string(runes[start:end])
			if segment != "" {
				segments = append(segments, segment)
			}
			start = end
			continue
		}

		segment := string(runes[start:splitPoint])
		if segment != "" {
			segments = append(segments, segment)
		}
		start = splitPoint
	}

	return segments
}

// findBestSplitPoint finds the best place to split text within a range.
// Priority order:
// 1. Strong sentence endings (。！？) - always split here
// 2. Medium separators (; : , ) - split if segment length reasonable
// 3. Soft boundaries (\n, space) - fallback
// 4. Hard cut at nearest character if no other option
func findBestSplitPoint(runes []rune, start, end int) int {
	strongBoundary := []rune("。！？.!?")
	mediumBoundary := []rune("；;：:,.")
	softBoundary := []rune("\n")

	// Strategy 1: Find first strong boundary in [start, end)
	for i := start; i < end && i < len(runes); i++ {
		if containsRune(strongBoundary, runes[i]) {
			return i + 1 // Include punctuation
		}
	}

	// Strategy 2: Find medium boundary where previous segment is reasonable length
	minSegmentLen := max(2, end-start-max(2, (end-start)/2)) // At least 2 chars or half of remaining
	for i := start + minSegmentLen; i < end && i < len(runes); i++ {
		if containsRune(mediumBoundary, runes[i]) {
			return i + 1
		}
	}

	// Strategy 3: Find soft boundary (newline)
	for i := start; i < end && i < len(runes); i++ {
		if containsRune(softBoundary, runes[i]) {
			return i + 1
		}
	}

	// Strategy 4: No boundary found, return -1 to trigger fallback
	return -1
}

// containsRune checks if a rune exists in the rune slice
func containsRune(runes []rune, target rune) bool {
	for _, r := range runes {
		if r == target {
			return true
		}
	}
	return false
}

func shouldFlushTTSBuffer(
	buffer string,
	runes int,
	lastFlushAt time.Time,
	segmentMaxChars int,
	flushMinChars int,
	flushInterval time.Duration,
	strongSuffixes []string,
	weakSuffixes []string,
) bool {
	if runes <= 0 {
		return false
	}
	flushChars := flushMinChars
	if flushChars <= 0 {
		flushChars = 40
	}
	if segmentMaxChars > 0 && segmentMaxChars < flushChars {
		flushChars = segmentMaxChars
	}
	if runes >= flushChars {
		return true
	}
	// 强/弱后缀列表来自配置（config.tts.flush_strong_suffixes / flush_weak_suffixes）。
	switch ttsTrailingPunctClass(buffer, strongSuffixes, weakSuffixes) {
	case ttsPunctStrong:
		return true
	case ttsPunctWeak:
		// 继续攒字，依赖字数阈值与时间间隔
	}
	if flushInterval <= 0 {
		flushInterval = 600 * time.Millisecond
	}
	if time.Since(lastFlushAt) >= flushInterval {
		return true
	}
	return false
}

type ttsPunctClass int

const (
	ttsPunctNone ttsPunctClass = iota
	ttsPunctStrong
	ttsPunctWeak
)

// trimTrailingUnicodeSpace 仅去掉尾部 Unicode 空白，便于 HasSuffix 识别句末标点；
// 不 Trim 首部，也不改动送入 TTS 的 buffer 正文（本函数仅用于刷新判断的副本）。
func trimTrailingUnicodeSpace(s string) string {
	r := []rune(s)
	i := len(r)
	for i > 0 && unicode.IsSpace(r[i-1]) {
		i--
	}
	return string(r[:i])
}

// ttsTrailingPunctClass 根据配置的强/弱后缀列表分类尾部标点。
func ttsTrailingPunctClass(s string, strongSuffixes, weakSuffixes []string) ttsPunctClass {
	t := trimTrailingUnicodeSpace(s)
	if t == "" {
		return ttsPunctNone
	}
	for _, p := range strongSuffixes {
		if p != "" && strings.HasSuffix(t, p) {
			return ttsPunctStrong
		}
	}
	for _, p := range weakSuffixes {
		if p != "" && strings.HasSuffix(t, p) {
			return ttsPunctWeak
		}
	}
	return ttsPunctNone
}

// toDomainMessages converts proto messages to domain messages.
func (s *ChatService) toDomainMessages(in []*chatv1.Message) []domain.Message {
	if len(in) == 0 {
		return nil
	}
	out := make([]domain.Message, 0, len(in))
	for _, m := range in {
		content := m.Content
		if strings.EqualFold(strings.TrimSpace(m.Role), "user") {
			content = appendNoThinkSuffix(content)
		}
		msg := domain.Message{
			Role:        m.Role,
			Content:     content,
			ToolCallID:  m.ToolCallId,
			IsTemporary: m.IsTemporary,
		}
		if len(m.ToolCalls) > 0 {
			msg.ToolCalls = make([]domain.ToolCall, 0, len(m.ToolCalls))
			for idx, tc := range m.ToolCalls {
				msg.ToolCalls = append(msg.ToolCalls, domain.ToolCall{
					ID:    tc.Id,
					Type:  "function",
					Index: idx,
					Function: domain.FunctionCall{
						Name:      tc.Name,
						Arguments: tc.ArgumentsJson,
					},
				})
			}
		}
		out = append(out, msg)
	}
	return out
}

func appendNoThinkSuffix(content string) string {
	if strings.Contains(content, "/no_think") {
		return content
	}
	trimmed := strings.TrimRight(content, " \t\r\n")
	if trimmed == "" {
		return "/no_think"
	}
	return trimmed + "\n/no_think"
}

// normalizeStreamingDelta 将上游 SSE 的「可能含重叠前缀」的片段转为纯增量。
// 仅按「已有正文后缀」与「本片段前缀」取最长重叠后剥掉重复部分。
// 注意：不得使用 strings.Contains(existing, incoming) —— 短片段（如单独一个「。」）若在前文任意位置
// 已出现过，会被误当成整段重复而丢弃，导致句末标点、叠字等丢失。
func normalizeStreamingDelta(existing string, incoming string) string {
	if incoming == "" {
		return ""
	}
	if existing == "" {
		return incoming
	}
	if existing == incoming {
		return ""
	}
	existingRunes := []rune(existing)
	incomingRunes := []rune(incoming)
	maxOverlap := len(existingRunes)
	if len(incomingRunes) < maxOverlap {
		maxOverlap = len(incomingRunes)
	}
	for overlap := maxOverlap; overlap > 0; overlap-- {
		if string(existingRunes[len(existingRunes)-overlap:]) == string(incomingRunes[:overlap]) {
			return string(incomingRunes[overlap:])
		}
	}
	return incoming
}

func effectiveChatQuery(req *chatv1.StreamChatRequest) string {
	if req == nil {
		return ""
	}
	q := strings.TrimSpace(req.Query)
	if q != "" {
		return q
	}
	for i := len(req.Messages) - 1; i >= 0; i-- {
		m := req.Messages[i]
		if m == nil {
			continue
		}
		if strings.EqualFold(strings.TrimSpace(m.Role), "user") {
			return strings.TrimSpace(m.Content)
		}
	}
	return ""
}

func resolveShortTermUserID(req *chatv1.StreamChatRequest, stm *config.ShortTermMemoryConfig) string {
	if req == nil {
		return ""
	}
	if strings.TrimSpace(req.ClientId) != "" {
		return strings.TrimSpace(req.ClientId)
	}
	if stm != nil && strings.TrimSpace(stm.DefaultUserID) != "" {
		return strings.TrimSpace(stm.DefaultUserID)
	}
	return strings.TrimSpace(req.DeviceId)
}

// mergeGatewaySystemPromptAtHead 将 gateway.system_prompt 置于短期记忆 build_prompt 返回的 messages 最前。
// 若首条已是 system，则把网关提示拼在该条内容之前（两段之间换行），避免重复多条 system 分散在索引 0。
func mergeGatewaySystemPromptAtHead(built []domain.Message, gatewaySystem string) []domain.Message {
	gw := strings.TrimSpace(gatewaySystem)
	if gw == "" || len(built) == 0 {
		return built
	}
	first := built[0]
	if strings.EqualFold(strings.TrimSpace(first.Role), "system") {
		rest := strings.TrimSpace(first.Content)
		merged := gw
		if rest != "" {
			merged = gw + "\n\n" + rest
		}
		out := make([]domain.Message, 0, len(built))
		out = append(out, domain.Message{Role: "system", Content: merged})
		out = append(out, built[1:]...)
		return out
	}
	out := make([]domain.Message, 0, len(built)+1)
	out = append(out, domain.Message{Role: "system", Content: gw})
	out = append(out, built...)
	return out
}
