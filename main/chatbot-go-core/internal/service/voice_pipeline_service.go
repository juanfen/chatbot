package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"xiaozhi-go-core/internal/adapter"
	"xiaozhi-go-core/internal/config"
	"xiaozhi-go-core/internal/pb/chatv1"
)

type VoicePipelineService struct {
	store       *config.Store
	chat        *ChatService
	vad         *adapter.VADClient
	asr         *adapter.FunASRClient
	voiceprint  *adapter.VoiceprintClient
	deviceAudio *adapter.DeviceAudioClient
	logger      *slog.Logger
}

type voiceSpeakerResult struct {
	speakerID     string
	speakerName   string
	personalLines []string
	score         float64
}

type VoiceEventStream interface {
	Context() context.Context
	Send(*chatv1.VoiceStreamResponse) error
}

type TranscriptDecisionStream interface {
	ShouldContinueAfterTranscript(transcript string, speaker string) bool
}

type VoicePipelineSession struct {
	service *VoicePipelineService
	state   *voiceSessionState
	stream  VoiceEventStream
}

type voiceSessionState struct {
	started             bool
	sessionID           string
	requestID           string
	deviceID            string
	clientID            string
	listenMode          string
	maxOutputSize       int32
	lastDeviceAudioID   string
	metadataJSON        string
	messages            []*chatv1.Message
	preroll             [][]byte
	prerollPCM          [][]byte
	currentOpus         [][]byte
	currentPCM          [][]byte
	speechStarted       bool
	lastClientHaveVoice bool
	decoder             *adapter.OpusDecoder
}

func NewVoicePipelineService(store *config.Store, chat *ChatService, logger *slog.Logger) *VoicePipelineService {
	return &VoicePipelineService{
		store:       store,
		chat:        chat,
		vad:         adapter.NewVADClient(store, logger),
		asr:         adapter.NewFunASRClient(store, logger),
		voiceprint:  adapter.NewVoiceprintClient(store, logger),
		deviceAudio: adapter.NewDeviceAudioClient(store, logger),
		logger:      logger,
	}
}

func (s *VoicePipelineService) Logger() *slog.Logger { return s.logger }

func (s *VoicePipelineService) NewSession(stream VoiceEventStream, req *chatv1.VoiceStreamRequest) (*VoicePipelineSession, error) {
	session := &VoicePipelineSession{
		service: s,
		state:   &voiceSessionState{},
		stream:  stream,
	}
	if err := s.handleStart(session.state, req); err != nil {
		return nil, err
	}
	return session, nil
}

func (s *VoicePipelineService) RunChat(stream VoiceEventStream, req *chatv1.StreamChatRequest) error {
	state := &voiceSessionState{
		sessionID: req.SessionId,
		requestID: req.RequestId,
		deviceID:  req.DeviceId,
		clientID:  req.ClientId,
	}
	return s.chat.StreamChat(req, func(resp *chatv1.StreamChatResponse) error {
		return stream.Send(chatResponseToVoice(state, resp))
	})
}

func (s *VoicePipelineSession) HandleAudio(req *chatv1.VoiceStreamRequest) error {
	return s.service.handleAudio(s.stream, s.state, req)
}

func (s *VoicePipelineSession) ForceStop() error {
	if len(s.state.currentOpus) == 0 && len(s.state.preroll) > 0 {
		s.state.currentOpus = append(s.state.currentOpus, s.state.preroll...)
		s.state.currentPCM = append(s.state.currentPCM, s.state.prerollPCM...)
		s.state.preroll = nil
		s.state.prerollPCM = nil
	}
	return s.service.processUtterance(s.stream, s.state, true)
}

func (s *VoicePipelineSession) SetMessages(messages []*chatv1.Message) {
	s.state.messages = cloneMessages(messages)
}

func (s *VoicePipelineSession) Reset() {
	s.service.resetUtterance(s.state)
	s.service.resetVAD(s.stream.Context(), s.state)
}

func (s *VoicePipelineSession) Close() {
	s.service.resetVAD(s.stream.Context(), s.state)
}

func (s *VoicePipelineService) handleStart(state *voiceSessionState, req *chatv1.VoiceStreamRequest) error {
	decoder, err := adapter.NewOpusDecoder(16000, 1)
	if err != nil {
		return err
	}
	state.started = true
	state.sessionID = req.SessionId
	state.requestID = req.RequestId
	state.deviceID = req.DeviceId
	state.clientID = req.ClientId
	state.listenMode = req.ListenMode
	state.maxOutputSize = req.MaxOutputSize
	state.metadataJSON = req.MetadataJson
	state.messages = cloneMessages(req.Messages)
	state.decoder = decoder
	state.preroll = nil
	state.prerollPCM = nil
	state.currentOpus = nil
	state.currentPCM = nil
	if state.sessionID == "" {
		return fmt.Errorf("session_id is required")
	}
	s.logger.Info("voice_stream_started",
		"session_id", state.sessionID,
		"request_id", state.requestID,
		"device_id", state.deviceID,
		"listen_mode", state.listenMode,
		"voiceprint_enabled", s.voiceprint.Enabled(),
		"chatbot_admin_enabled", s.store.Get().ChatbotAdmin.Enabled,
	)
	return nil
}

func (s *VoicePipelineService) mergeContext(state *voiceSessionState, req *chatv1.VoiceStreamRequest) {
	if req.RequestId != "" {
		state.requestID = req.RequestId
	}
	if req.MetadataJson != "" {
		state.metadataJSON = req.MetadataJson
	}
	if req.MaxOutputSize > 0 {
		state.maxOutputSize = req.MaxOutputSize
	}
	if len(req.Messages) > 0 {
		state.messages = cloneMessages(req.Messages)
	}
}

func (s *VoicePipelineService) handleAudio(stream VoiceEventStream, state *voiceSessionState, req *chatv1.VoiceStreamRequest) error {
	if len(req.AudioFrame) == 0 {
		return nil
	}
	ctx := stream.Context()
	vadResult, err := s.vad.Infer(ctx, state.sessionID, req.AudioFrame)
	if err != nil {
		return err
	}
	pcm, _ := state.decoder.Decode(req.AudioFrame, 960)
	if vadResult.HaveVoice && !state.speechStarted {
		state.speechStarted = true
		s.logger.Info("vad_speech_start",
			"session_id", state.sessionID,
			"request_id", state.requestID,
			"frame_seq", req.FrameSeq,
		)
		if err := stream.Send(&chatv1.VoiceStreamResponse{
			Event:     chatv1.VoiceEvent_VOICE_EVENT_VAD_SPEECH_START,
			SessionId: state.sessionID,
			RequestId: state.requestID,
			FrameSeq:  req.FrameSeq,
		}); err != nil {
			return err
		}
		state.currentOpus = append(state.currentOpus, state.preroll...)
		state.currentPCM = append(state.currentPCM, state.prerollPCM...)
		state.preroll = nil
		state.prerollPCM = nil
	}
	if state.speechStarted || vadResult.ClientHaveVoice {
		state.currentOpus = append(state.currentOpus, append([]byte(nil), req.AudioFrame...))
		if len(pcm) > 0 {
			state.currentPCM = append(state.currentPCM, pcm)
		}
	} else {
		state.preroll = append(state.preroll, append([]byte(nil), req.AudioFrame...))
		if len(pcm) > 0 {
			state.prerollPCM = append(state.prerollPCM, pcm)
		}
		maxPreroll := s.store.Get().AudioPipeline.PrerollFrames
		if maxPreroll <= 0 {
			maxPreroll = 10
		}
		if len(state.preroll) > maxPreroll {
			state.preroll = state.preroll[len(state.preroll)-maxPreroll:]
		}
		if len(state.prerollPCM) > maxPreroll {
			state.prerollPCM = state.prerollPCM[len(state.prerollPCM)-maxPreroll:]
		}
	}
	state.lastClientHaveVoice = vadResult.ClientHaveVoice
	if vadResult.ClientVoiceStop {
		s.logger.Info("vad_speech_end",
			"session_id", state.sessionID,
			"request_id", state.requestID,
			"frame_seq", req.FrameSeq,
			"opus_frames", len(state.currentOpus),
			"pcm_chunks", len(state.currentPCM),
		)
		if err := stream.Send(&chatv1.VoiceStreamResponse{
			Event:     chatv1.VoiceEvent_VOICE_EVENT_VAD_SPEECH_END,
			SessionId: state.sessionID,
			RequestId: state.requestID,
			FrameSeq:  req.FrameSeq,
		}); err != nil {
			return err
		}
		return s.processUtterance(stream, state, false)
	}
	return nil
}

func (s *VoicePipelineService) processUtterance(stream VoiceEventStream, state *voiceSessionState, force bool) error {
	start := time.Now()
	minFrames := s.store.Get().AudioPipeline.MinUtteranceFrames
	if minFrames <= 0 {
		minFrames = 15
	}
	if len(state.currentOpus) == 0 {
		return nil
	}
	if !force && len(state.currentOpus) <= minFrames {
		s.logger.Info("utterance_ignored_too_short",
			"session_id", state.sessionID,
			"request_id", state.requestID,
			"opus_frames", len(state.currentOpus),
			"min_frames", minFrames,
		)
		s.resetUtterance(state)
		s.resetVAD(stream.Context(), state)
		return nil
	}
	pcm := concatBytes(state.currentPCM)
	const uplinkSampleRate = 16000
	const uplinkChannels = 1
	durationSec := float64(len(pcm)) / float64(uplinkSampleRate*uplinkChannels*2)
	adminCfg := s.store.Get().ChatbotAdmin
	needUpload := adminCfg.Enabled && s.deviceAudio.Enabled() &&
		durationSec >= float64(adminCfg.EffectiveRecordAudioLenSec())
	wav := adapter.PCM16ToWAV(pcm, uplinkSampleRate, uplinkChannels)

	s.logger.Info("utterance_process_start",
		"session_id", state.sessionID,
		"request_id", state.requestID,
		"opus_frames", len(state.currentOpus),
		"pcm_bytes", len(pcm),
		"duration_sec", durationSec,
		"device_audio_upload", needUpload,
		"device_audio_min_sec", adminCfg.EffectiveRecordAudioLenSec(),
		"voiceprint_enabled", s.voiceprint.Enabled(),
	)
	ctx := stream.Context()
	macID := strings.TrimSpace(state.deviceID)
	speakerCh := s.startSpeakerIdentify(ctx, macID, state.sessionID, state.requestID, wav, durationSec)

	transcript, asrErr := s.asr.Recognize(ctx, pcm, state.sessionID)
	speakerResult := s.waitSpeakerIdentify(ctx, speakerCh)
	if asrErr != nil {
		s.logger.Error("utterance_asr_failed",
			"session_id", state.sessionID,
			"request_id", state.requestID,
			"elapsed_ms", time.Since(start).Milliseconds(),
			"error", asrErr,
		)
		s.resetUtterance(state)
		s.resetVAD(ctx, state)
		return stream.Send(s.errorResponse(state, "ASR_FAILED", asrErr.Error()))
	}
	transcript = strings.TrimSpace(transcript)
	s.logger.Info("utterance_process_done",
		"session_id", state.sessionID,
		"request_id", state.requestID,
		"elapsed_ms", time.Since(start).Milliseconds(),
		"transcript_len", len([]rune(transcript)),
	)
	if transcript != "" {
		if err := stream.Send(&chatv1.VoiceStreamResponse{
			Event:      chatv1.VoiceEvent_VOICE_EVENT_ASR_TRANSCRIPT,
			SessionId:  state.sessionID,
			RequestId:  state.requestID,
			Transcript: transcript,
			Speaker:    speakerResult.speakerName,
		}); err != nil {
			return err
		}
		if needUpload {
			if macID == "" {
				s.logger.Warn("device_audio_upload_skipped", "reason", "empty_device_id", "session_id", state.sessionID)
				s.resetUtterance(state)
				s.resetVAD(ctx, state)
				return stream.Send(s.errorResponse(state, "DEVICE_AUDIO_UPLOAD_FAILED", "device_id is required for audio upload"))
			}
			audioLenLabel := formatAudioLenLabel(durationSec)
			s.logger.Info("device_audio_upload_start",
				"session_id", state.sessionID,
				"mac_id", macID,
				"duration_sec", durationSec,
				"audio_len", audioLenLabel,
				"pcm_bytes", len(pcm),
				"transcript_preview", previewUtteranceText(transcript, 80),
			)
			uploadStart := time.Now()
			result, uploadErr := s.deviceAudio.Upload(ctx, macID, audioLenLabel, transcript, wav)
			if uploadErr != nil {
				s.logger.Warn("device_audio_upload_failed",
					"session_id", state.sessionID,
					"mac_id", macID,
					"elapsed_ms", time.Since(uploadStart).Milliseconds(),
					"error", uploadErr,
				)
				s.resetUtterance(state)
				s.resetVAD(ctx, state)
				return stream.Send(s.errorResponse(state, "DEVICE_AUDIO_UPLOAD_FAILED", uploadErr.Error()))
			}
			state.lastDeviceAudioID = result.ID
			s.logger.Info("device_audio_upload_ok",
				"session_id", state.sessionID,
				"mac_id", macID,
				"record_id", result.ID,
				"elapsed_ms", time.Since(uploadStart).Milliseconds(),
			)
		}
		if decider, ok := stream.(TranscriptDecisionStream); ok && !decider.ShouldContinueAfterTranscript(transcript, speakerResult.speakerName) {
			s.resetUtterance(state)
			s.resetVAD(ctx, state)
			return nil
		}
		if err := s.runChat(stream, state, transcript, speakerResult.speakerName, speakerResult.personalLines); err != nil {
			return err
		}
	}
	s.resetUtterance(state)
	s.resetVAD(ctx, state)
	return nil
}

func (s *VoicePipelineService) runChat(stream VoiceEventStream, state *voiceSessionState, query string, speakerName string, personalLines []string) error {
	requestID := state.requestID
	if requestID == "" {
		requestID = fmt.Sprintf("%d", time.Now().UnixNano())
	}
	messages := prependSpeakerToSystemPrompt(appendUserMessage(state.messages, query), speakerName, personalLines)
	req := &chatv1.StreamChatRequest{
		SessionId:         state.sessionID,
		RequestId:         requestID,
		DeviceId:          state.deviceID,
		ClientId:          state.clientID,
		Query:             query,
		MaxOutputSize:     state.maxOutputSize,
		Messages:          messages,
		VoiceprintEnabled: s.voiceprint.Enabled(),
		MetadataJson:      mergeMetadataWithDeviceAudioID(state.metadataJSON, state.lastDeviceAudioID),
	}
	if speakerName != "" {
		s.logger.Info("llm_speaker_context",
			"session_id", state.sessionID,
			"request_id", requestID,
			"speaker_name", speakerName,
		)
	}
	return s.chat.StreamChat(req, func(resp *chatv1.StreamChatResponse) error {
		return stream.Send(chatResponseToVoice(state, resp))
	})
}

func (s *VoicePipelineService) startSpeakerIdentify(
	ctx context.Context,
	macID, sessionID, requestID string,
	wav []byte,
	durationSec float64,
) <-chan voiceSpeakerResult {
	ch := make(chan voiceSpeakerResult, 1)
	if len(wav) == 0 {
		s.logger.Info("声纹识别跳过：音频为空",
			"session_id", sessionID,
			"request_id", requestID,
		)
		ch <- voiceSpeakerResult{}
		close(ch)
		return ch
	}
	if !s.voiceprint.Enabled() {
		s.logger.Info("声纹识别跳过：voiceprint 未启用或缺少 url/api_key",
			"session_id", sessionID,
			"request_id", requestID,
			"duration_sec", durationSec,
		)
		ch <- voiceSpeakerResult{}
		close(ch)
		return ch
	}
	if !s.deviceAudio.AdminAPIReady() {
		s.logger.Info("声纹识别跳过：chatbot-admin 地址未配置",
			"session_id", sessionID,
			"request_id", requestID,
		)
		ch <- voiceSpeakerResult{}
		close(ch)
		return ch
	}
	if macID == "" {
		s.logger.Info("声纹识别跳过：设备 mac_id 为空",
			"session_id", sessionID,
			"request_id", requestID,
		)
		ch <- voiceSpeakerResult{}
		close(ch)
		return ch
	}

	go func() {
		result := voiceSpeakerResult{}
		defer func() { ch <- result; close(ch) }()

		s.logger.Info("声纹识别：开始（不限音频时长，与 record_audio_len_sec 无关）",
			"session_id", sessionID,
			"request_id", requestID,
			"mac_id", macID,
			"duration_sec", durationSec,
			"wav_bytes", len(wav),
		)

		s.logger.Info("声纹识别：开始查询设备已注册 speaker_id 列表",
			"session_id", sessionID,
			"request_id", requestID,
			"mac_id", macID,
		)
		speakerIDs, err := s.deviceAudio.ListSpeakerIDs(ctx, macID)
		if err != nil {
			s.logger.Warn("声纹识别：查询 speaker_id 列表失败",
				"session_id", sessionID,
				"request_id", requestID,
				"mac_id", macID,
				"error", err,
			)
			return
		}
		s.logger.Info("声纹识别：查询 speaker_id 列表完成",
			"session_id", sessionID,
			"request_id", requestID,
			"mac_id", macID,
			"speaker_count", len(speakerIDs),
		)
		if len(speakerIDs) == 0 {
			s.logger.Info("声纹识别跳过：该设备暂无已注册声纹",
				"session_id", sessionID,
				"request_id", requestID,
				"mac_id", macID,
			)
			return
		}

		s.logger.Info("声纹识别：开始调用声纹识别接口",
			"session_id", sessionID,
			"request_id", requestID,
			"mac_id", macID,
			"speaker_count", len(speakerIDs),
			"wav_bytes", len(wav),
		)
		identifyResult, err := s.voiceprint.Identify(ctx, wav, speakerIDs)
		if err != nil {
			s.logger.Warn("声纹识别：声纹识别接口调用失败",
				"session_id", sessionID,
				"request_id", requestID,
				"error", err,
			)
			return
		}
		s.logger.Info("声纹识别：声纹识别接口调用完成",
			"session_id", sessionID,
			"request_id", requestID,
			"speaker_id", identifyResult.SpeakerID,
			"score", identifyResult.Score,
		)
		if identifyResult.SpeakerID == "" {
			s.logger.Info("声纹识别：未匹配到已注册说话人",
				"session_id", sessionID,
				"request_id", requestID,
				"score", identifyResult.Score,
			)
			return
		}

		s.logger.Info("声纹识别：开始查询说话人名称",
			"session_id", sessionID,
			"request_id", requestID,
			"speaker_id", identifyResult.SpeakerID,
		)
		profile, err := s.deviceAudio.GetSpeakerProfile(ctx, identifyResult.SpeakerID)
		if err != nil {
			s.logger.Warn("声纹识别：查询说话人名称失败",
				"session_id", sessionID,
				"request_id", requestID,
				"speaker_id", identifyResult.SpeakerID,
				"error", err,
			)
			return
		}
		s.logger.Info("声纹识别：查询说话人名称完成",
			"session_id", sessionID,
			"request_id", requestID,
			"speaker_id", identifyResult.SpeakerID,
			"speaker_name", profile.Name,
			"personal_lines_count", len(profile.PersonalLines),
		)
		result = voiceSpeakerResult{
			speakerID:     identifyResult.SpeakerID,
			speakerName:   profile.Name,
			personalLines: profile.PersonalLines,
			score:         identifyResult.Score,
		}
	}()
	return ch
}

func (s *VoicePipelineService) waitSpeakerIdentify(ctx context.Context, ch <-chan voiceSpeakerResult) voiceSpeakerResult {
	select {
	case result, ok := <-ch:
		if !ok {
			return voiceSpeakerResult{}
		}
		return result
	case <-ctx.Done():
		return voiceSpeakerResult{}
	}
}

func prependSpeakerToSystemPrompt(messages []*chatv1.Message, speakerName string, personalLines []string) []*chatv1.Message {
	speakerName = strings.TrimSpace(speakerName)
	var parts []string
	if speakerName != "" {
		parts = append(parts, fmt.Sprintf("当前说话人是「%s」。", speakerName))
	}
	for _, line := range personalLines {
		line = strings.TrimSpace(line)
		if line != "" {
			parts = append(parts, line)
		}
	}
	if len(parts) == 0 {
		return messages
	}
	hint := strings.Join(parts, "\n")
	out := cloneMessages(messages)
	for i, m := range out {
		if m != nil && strings.EqualFold(strings.TrimSpace(m.Role), "system") {
			content := strings.TrimSpace(m.Content)
			merged := hint
			if content != "" {
				merged = hint + "\n\n" + content
			}
			out[i] = &chatv1.Message{Role: "system", Content: merged}
			return out
		}
	}
	return append([]*chatv1.Message{{Role: "system", Content: hint}}, out...)
}

func (s *VoicePipelineService) resetUtterance(state *voiceSessionState) {
	state.currentOpus = nil
	state.currentPCM = nil
	state.preroll = nil
	state.prerollPCM = nil
	state.speechStarted = false
	state.lastClientHaveVoice = false
	if decoder, err := adapter.NewOpusDecoder(16000, 1); err == nil {
		state.decoder = decoder
	}
}

func (s *VoicePipelineService) resetVAD(ctx context.Context, state *voiceSessionState) {
	if state.sessionID != "" {
		s.vad.Reset(ctx, state.sessionID)
	}
}

func (s *VoicePipelineService) errorResponse(state *voiceSessionState, code string, message string) *chatv1.VoiceStreamResponse {
	return &chatv1.VoiceStreamResponse{
		Event:        chatv1.VoiceEvent_VOICE_EVENT_ERROR,
		SessionId:    state.sessionID,
		RequestId:    state.requestID,
		ErrorCode:    code,
		ErrorMessage: message,
	}
}

func chatResponseToVoice(state *voiceSessionState, resp *chatv1.StreamChatResponse) *chatv1.VoiceStreamResponse {
	out := &chatv1.VoiceStreamResponse{
		SessionId:       state.sessionID,
		RequestId:       state.requestID,
		ContentChunk:    resp.ContentChunk,
		ToolCall:        resp.ToolCall,
		FinishReason:    resp.FinishReason,
		ErrorCode:       resp.ErrorCode,
		ErrorMessage:    resp.ErrorMessage,
		TtsStreamId:     resp.TtsStreamId,
		ChunkSeq:        resp.ChunkSeq,
		IsLastChunk:     resp.IsLastChunk,
		Codec:           resp.Codec,
		SampleRateHz:    resp.SampleRateHz,
		Channels:        resp.Channels,
		FrameDurationMs: resp.FrameDurationMs,
		AudioChunk:      resp.AudioChunk,
	}
	switch resp.Event {
	case chatv1.StreamEvent_STREAM_EVENT_CONTENT_CHUNK:
		out.Event = chatv1.VoiceEvent_VOICE_EVENT_CONTENT_CHUNK
	case chatv1.StreamEvent_STREAM_EVENT_TOOL_CALL:
		out.Event = chatv1.VoiceEvent_VOICE_EVENT_TOOL_CALL
	case chatv1.StreamEvent_STREAM_EVENT_FINISH:
		out.Event = chatv1.VoiceEvent_VOICE_EVENT_FINISH
	case chatv1.StreamEvent_STREAM_EVENT_ERROR:
		out.Event = chatv1.VoiceEvent_VOICE_EVENT_ERROR
	case chatv1.StreamEvent_STREAM_EVENT_TTS_START:
		out.Event = chatv1.VoiceEvent_VOICE_EVENT_TTS_START
	case chatv1.StreamEvent_STREAM_EVENT_AUDIO_CHUNK:
		out.Event = chatv1.VoiceEvent_VOICE_EVENT_AUDIO_CHUNK
	case chatv1.StreamEvent_STREAM_EVENT_TTS_END:
		out.Event = chatv1.VoiceEvent_VOICE_EVENT_TTS_END
	case chatv1.StreamEvent_STREAM_EVENT_TTS_ERROR:
		out.Event = chatv1.VoiceEvent_VOICE_EVENT_TTS_ERROR
	default:
		out.Event = chatv1.VoiceEvent_VOICE_EVENT_UNSPECIFIED
	}
	return out
}

func appendUserMessage(messages []*chatv1.Message, query string) []*chatv1.Message {
	out := cloneMessages(messages)
	if strings.TrimSpace(query) == "" {
		return out
	}
	out = append(out, &chatv1.Message{Role: "user", Content: query})
	return out
}

// cloneMessages clones a slice of message pointers without deep copying string data.
// The returned pointers reference the same underlying string data as the originals.
func cloneMessages(messages []*chatv1.Message) []*chatv1.Message {
	if len(messages) == 0 {
		return nil
	}
	out := make([]*chatv1.Message, 0, len(messages))
	for _, m := range messages {
		if m == nil {
			continue
		}
		// 创建一个新的指针但复用原始 Message 的字段引用
		cp := &chatv1.Message{
			Role:        m.Role,
			Content:     m.Content,
			ToolCallId:  m.ToolCallId,
			IsTemporary: m.IsTemporary,
		}
		if len(m.ToolCalls) > 0 {
			cp.ToolCalls = append([]*chatv1.ToolCall(nil), m.ToolCalls...)
		}
		out = append(out, cp)
	}
	return out
}

func formatAudioLenLabel(durationSec float64) string {
	if durationSec <= 0 {
		return "0s"
	}
	if durationSec == float64(int(durationSec)) {
		return fmt.Sprintf("%ds", int(durationSec))
	}
	return fmt.Sprintf("%.1fs", durationSec)
}

func previewUtteranceText(s string, max int) string {
	s = strings.TrimSpace(s)
	if max <= 0 {
		max = 80
	}
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	return string(r[:max]) + "..."
}

func mergeMetadataWithDeviceAudioID(metadataJSON, deviceAudioID string) string {
	deviceAudioID = strings.TrimSpace(deviceAudioID)
	if deviceAudioID == "" {
		return metadataJSON
	}
	var payload map[string]any
	if strings.TrimSpace(metadataJSON) != "" {
		if err := json.Unmarshal([]byte(metadataJSON), &payload); err != nil {
			payload = map[string]any{}
		}
	} else {
		payload = map[string]any{}
	}
	payload["device_audio_id"] = deviceAudioID
	data, err := json.Marshal(payload)
	if err != nil {
		return metadataJSON
	}
	return string(data)
}

func concatBytes(chunks [][]byte) []byte {
	total := 0
	for _, chunk := range chunks {
		total += len(chunk)
	}
	out := make([]byte, 0, total)
	for _, chunk := range chunks {
		out = append(out, chunk...)
	}
	return out
}
