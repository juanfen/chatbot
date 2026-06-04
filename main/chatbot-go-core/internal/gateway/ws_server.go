package gateway

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/gorilla/websocket"

	"xiaozhi-go-core/internal/adapter"
	"xiaozhi-go-core/internal/config"
	"xiaozhi-go-core/internal/pb/chatv1"
	"xiaozhi-go-core/internal/service"
)

type WSServer struct {
	store  *config.Store
	voice  *service.VoicePipelineService
	logger *slog.Logger
	server *http.Server
}

type gatewaySession struct {
	store                  *config.Store
	voice                  *service.VoicePipelineService
	logger                 *slog.Logger
	conn                   *websocket.Conn
	sendMu                 sync.Mutex
	ctx                    context.Context
	cancel                 context.CancelFunc
	sessionID              string
	requestID              string
	chatCancel             context.CancelFunc
	chatMu                 sync.Mutex
	deviceID               string
	clientID               string
	listenMode             string
	frameSeq               uint64
	outSeq                 uint32
	fromMQTT               bool
	lastActivity           time.Time
	closeAfterChat         bool
	sentenceStarted        bool
	ttsActive              bool
	audioTimestampBuffer   map[uint32][]byte
	lastProcessedTimestamp uint32
	voiceSession           *service.VoicePipelineSession
	opusEncoder            *adapter.OpusEncoder
	rateController         *AudioRateController
	responseText           strings.Builder
	dialogue               []*chatv1.Message
}

func (s *gatewaySession) ac() *config.AppConfig { return s.store.Get() }

type wsEventStream struct {
	ctx     context.Context
	session *gatewaySession
}

func (s wsEventStream) Context() context.Context { return s.ctx }
func (s wsEventStream) Send(resp *chatv1.VoiceStreamResponse) error {
	return s.session.handleVoiceEvent(resp)
}
func (s wsEventStream) ShouldContinueAfterTranscript(transcript string, speaker string) bool {
	return s.session.shouldContinueAfterTranscript(transcript, speaker)
}

func NewWSServer(store *config.Store, voice *service.VoicePipelineService, logger *slog.Logger) *WSServer {
	return &WSServer{
		store:  store,
		voice:  voice,
		logger: logger,
	}
}

func (s *WSServer) Start() {
	gw := s.store.Get().Gateway
	if !gw.Enabled || gw.WSListenAddr == "" {
		return
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/xiaozhi/v1/", s.handleWS)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("Server is running"))
	})
	s.server = &http.Server{
		Addr:              gw.WSListenAddr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}
	go func() {
		s.logger.Info("gateway_ws_listening", "addr", gw.WSListenAddr)
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Error("gateway_ws_error", "error", err)
		}
	}()
}

func (s *WSServer) Stop(ctx context.Context) {
	if s.server != nil {
		_ = s.server.Shutdown(ctx)
	}
}

func (s *WSServer) handleWS(w http.ResponseWriter, r *http.Request) {
	deviceID := firstNonEmpty(r.Header.Get("device-id"), r.URL.Query().Get("device-id"))
	clientID := firstNonEmpty(r.Header.Get("client-id"), r.URL.Query().Get("client-id"))
	authHeader := firstNonEmpty(r.Header.Get("authorization"), r.URL.Query().Get("authorization"))
	if deviceID == "" {
		http.Error(w, "missing device-id", http.StatusBadRequest)
		return
	}
	if clientID == "" {
		clientID = deviceID
	}
	if !s.authorized(deviceID, clientID, authHeader) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	upgrader := websocket.Upgrader{CheckOrigin: func(*http.Request) bool { return true }}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		s.logger.Warn("ws_upgrade_failed", "device_id", deviceID, "error", err)
		return
	}
	ctx, cancel := context.WithCancel(r.Context())
	session := &gatewaySession{
		store:                s.store,
		voice:                s.voice,
		logger:               s.logger.With("session_id", randomID(), "device_id", deviceID),
		conn:                 conn,
		ctx:                  ctx,
		cancel:               cancel,
		sessionID:            randomID(),
		requestID:            randomID(),
		deviceID:             deviceID,
		clientID:             clientID,
		listenMode:           "auto",
		fromMQTT:             r.URL.Query().Get("from") == "mqtt_gateway",
		lastActivity:         time.Now(),
		audioTimestampBuffer: map[uint32][]byte{},
	}
	session.logger.Info("ws_connected", "client_id", clientID, "remote", r.RemoteAddr)
	session.run()
}

func (s *WSServer) authorized(deviceID string, clientID string, authHeader string) bool {
	g := s.store.Get().Gateway
	if !g.Auth.Enabled {
		return true
	}
	for _, allowed := range g.Auth.AllowedDevices {
		if allowed == deviceID {
			return true
		}
	}
	auth := NewAuthManager(g.AuthKey, g.Auth.ExpireSeconds)
	return auth.VerifyToken(BearerToken(authHeader), clientID, deviceID)
}

func (s *gatewaySession) run() {
	go s.monitorIdle()
	go s.keepAlive()
	s.installPongHandler()
	defer func() {
		if s.voiceSession != nil {
			s.voiceSession.Close()
		}
		if s.rateController != nil {
			s.rateController.Stop()
		}
		s.cancel()
		_ = s.conn.Close()
		s.logger.Info("ws_closed")
	}()
	for {
		messageType, data, err := s.conn.ReadMessage()
		if err != nil {
			s.logger.Warn("ws_read_closed", "error", err)
			return
		}
		switch messageType {
		case websocket.TextMessage:
			s.lastActivity = time.Now()
			s.handleText(data)
		case websocket.BinaryMessage:
			s.handleAudio(data)
		}
	}
}

func (s *gatewaySession) installPongHandler() {
	s.conn.SetPongHandler(func(string) error {
		s.logger.Debug("ws_pong_received")
		return nil
	})
}

func (s *gatewaySession) keepAlive() {
	interval := time.Duration(effectiveInt(s.ac().Gateway.WSPingIntervalSeconds, 15)) * time.Second
	pongTimeout := time.Duration(effectiveInt(s.ac().Gateway.WSPongTimeoutSeconds, 30)) * time.Second
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			deadline := time.Now().Add(pongTimeout)
			s.sendMu.Lock()
			err := s.conn.WriteControl(websocket.PingMessage, []byte("ping"), deadline)
			s.sendMu.Unlock()
			if err != nil {
				s.logger.Warn("ws_ping_failed", "error", err)
				if !s.ttsActive {
					_ = s.conn.Close()
				}
				return
			}
		}
	}
}

func (s *gatewaySession) monitorIdle() {
	timeout := time.Duration(s.ac().Gateway.CloseConnectionNoVoiceTime) * time.Second
	if timeout <= 0 {
		timeout = 120 * time.Second
	}
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			if s.closeAfterChat || time.Since(s.lastActivity) <= timeout {
				continue
			}
			s.closeAfterChat = true
			if !s.ac().Gateway.EndPromptEnabled {
				s.cancel()
				_ = s.conn.Close()
				return
			}
			prompt := s.ac().Gateway.EndPrompt
			if strings.TrimSpace(prompt) == "" {
				prompt = "请你以```时间过得真快```未来头，用富有感情、依依不舍的话来结束这场对话吧。！"
			}
			s.cancelInflightChat("idle_end_prompt")
			s.requestID = randomID()
			s.sendTTSStart()
			chatCtx := s.startInflightChatContext()
			if err := s.voice.RunChat(wsEventStream{ctx: chatCtx, session: s}, s.chatRequest(prompt)); err != nil {
				s.logger.Error("idle_end_prompt_failed", "error", err)
			}
		}
	}
}

func (s *gatewaySession) handleText(data []byte) {
	var msg map[string]any
	if err := json.Unmarshal(data, &msg); err != nil {
		s.logger.Warn("ws_text_decode_failed", "error", err)
		return
	}
	switch stringValue(msg["type"]) {
	case "hello":
		s.handleHello(msg)
	case "listen":
		s.handleListen(msg)
	case "abort":
		s.handleAbort()
	case "ping":
		s.sendJSON(map[string]any{"type": "pong", "session_id": s.sessionID, "timestamp": time.Now().UnixMilli()})
	case "iot", "mcp", "server":
		s.logger.Info("ws_message_ignored_placeholder", "type", stringValue(msg["type"]))
	default:
		s.logger.Warn("ws_message_unknown", "type", stringValue(msg["type"]))
	}
}

func (s *gatewaySession) handleHello(msg map[string]any) {
	if audioParams, ok := msg["audio_params"].(map[string]any); ok {
		s.logger.Info("client_audio_params", "audio_params", audioParams)
	}
	welcome := map[string]any{
		"type":       firstNonEmpty(s.ac().Xiaozhi.Type, "hello"),
		"version":    s.ac().Xiaozhi.Version,
		"transport":  firstNonEmpty(s.ac().Xiaozhi.Transport, "websocket"),
		"session_id": s.sessionID,
		"audio_params": map[string]any{
			"format":         firstNonEmpty(s.ac().Xiaozhi.AudioParams.Format, "opus"),
			"sample_rate":    effectiveInt(s.ac().Xiaozhi.AudioParams.SampleRate, 16000),
			"channels":       effectiveInt(s.ac().Xiaozhi.AudioParams.Channels, 1),
			"frame_duration": effectiveInt(s.ac().Xiaozhi.AudioParams.FrameDuration, 60),
		},
	}
	s.sendJSON(welcome)
}

func (s *gatewaySession) handleListen(msg map[string]any) {
	if mode := stringValue(msg["mode"]); mode != "" {
		s.listenMode = mode
	}
	switch stringValue(msg["state"]) {
	case "start":
		if s.voiceSession != nil {
			s.voiceSession.Reset()
		}
	case "stop":
		s.cancelInflightChat("listen_stop")
		if s.voiceSession != nil {
			s.voiceSession.SetMessages(s.recentDialogue())
			if err := s.voiceSession.ForceStop(); err != nil {
				s.logger.Error("listen_stop_asr_failed", "error", err)
			}
		}
	case "detect":
		if text := stringValue(msg["text"]); text != "" {
			s.cancelInflightChat("listen_detect")
			s.requestID = randomID()
			s.lastActivity = time.Now()
			chatText := s.detectChatText(text)
			if chatText == "" {
				return
			}
			s.sendSTTMessage(chatText)
			req := s.chatRequest(chatText)
			chatCtx := s.startInflightChatContext()
			if err := s.voice.RunChat(wsEventStream{ctx: chatCtx, session: s}, req); err != nil {
				s.logger.Error("detect_chat_failed", "error", err)
			}
		}
	}
}

func (s *gatewaySession) detectChatText(text string) string {
	filtered := normalizeSpeechText(text)
	if s.isWakeupWord(filtered) && !s.ac().Gateway.EnableGreeting {
		s.sendJSON(map[string]any{"type": "stt", "text": trimPunctuationOrEmoji(text), "session_id": s.sessionID})
		s.sendTTSStop()
		return ""
	}
	if s.isWakeupWord(filtered) {
		return "嘿，你好呀"
	}
	return text
}

func (s *gatewaySession) isWakeupWord(text string) bool {
	for _, word := range s.ac().Gateway.WakeupWords {
		if normalizeSpeechText(word) == text {
			return true
		}
	}
	return false
}

func (s *gatewaySession) handleAbort() {
	s.cancelInflightChat("abort")
	if s.voiceSession != nil {
		s.voiceSession.Reset()
	}
	s.sendTTSStop()
}

func (s *gatewaySession) cancelInflightChat(reason string) {
	s.chatMu.Lock()
	cancel := s.chatCancel
	s.chatCancel = nil
	s.chatMu.Unlock()
	if cancel != nil {
		s.logger.Info("chat_cancel_inflight", "reason", reason)
		cancel()
	}
}

func (s *gatewaySession) startInflightChatContext() context.Context {
	s.chatMu.Lock()
	// cancel existing (defensive)
	if s.chatCancel != nil {
		s.chatCancel()
		s.chatCancel = nil
	}
	ctx, cancel := context.WithCancel(s.ctx)
	s.chatCancel = cancel
	s.chatMu.Unlock()
	return ctx
}

func (s *gatewaySession) handleAudio(data []byte) {
	if s.fromMQTT && len(data) >= 16 {
		timestamp := uint32(data[8])<<24 | uint32(data[9])<<16 | uint32(data[10])<<8 | uint32(data[11])
		audioLength := int(uint32(data[12])<<24 | uint32(data[13])<<16 | uint32(data[14])<<8 | uint32(data[15]))
		if audioLength > 0 && len(data) >= 16+audioLength {
			s.processMQTTAudio(data[16:16+audioLength], timestamp)
			return
		}
		if len(data) > 16 {
			data = data[16:]
		}
	}
	s.processAudioFrame(data)
}

func (s *gatewaySession) processMQTTAudio(audio []byte, timestamp uint32) {
	if timestamp >= s.lastProcessedTimestamp {
		s.processAudioFrame(audio)
		s.lastProcessedTimestamp = timestamp
		for {
			var nextTs uint32
			var nextAudio []byte
			found := false
			for ts, buffered := range s.audioTimestampBuffer {
				if ts > s.lastProcessedTimestamp && (!found || ts < nextTs) {
					nextTs = ts
					nextAudio = buffered
					found = true
				}
			}
			if !found {
				return
			}
			delete(s.audioTimestampBuffer, nextTs)
			s.processAudioFrame(nextAudio)
			s.lastProcessedTimestamp = nextTs
		}
	}
	if len(s.audioTimestampBuffer) < 20 {
		s.audioTimestampBuffer[timestamp] = append([]byte(nil), audio...)
		return
	}
	s.processAudioFrame(audio)
}

func (s *gatewaySession) processAudioFrame(data []byte) {
	if s.voiceSession == nil {
		req := s.voiceStartRequest()
		session, err := s.voice.NewSession(wsEventStream{ctx: s.ctx, session: s}, req)
		if err != nil {
			s.logger.Error("voice_session_start_failed", "error", err)
			return
		}
		s.voiceSession = session
	}
	s.voiceSession.SetMessages(s.recentDialogue())
	s.frameSeq++
	err := s.voiceSession.HandleAudio(&chatv1.VoiceStreamRequest{
		Type:            chatv1.VoiceRequestType_VOICE_REQUEST_TYPE_AUDIO,
		SessionId:       s.sessionID,
		RequestId:       s.requestID,
		DeviceId:        s.deviceID,
		ClientId:        s.clientID,
		FrameSeq:        s.frameSeq,
		Codec:           "opus",
		SampleRateHz:    16000,
		Channels:        1,
		FrameDurationMs: 60,
		AudioFrame:      data,
		ListenMode:      s.listenMode,
	})
	if err != nil {
		s.logger.Error("voice_audio_failed", "error", err)
	}
}

func (s *gatewaySession) handleVoiceEvent(resp *chatv1.VoiceStreamResponse) error {
	switch resp.Event {
	case chatv1.VoiceEvent_VOICE_EVENT_ASR_TRANSCRIPT:
		if resp.Transcript != "" {
			s.appendDialogue("user", resp.Transcript)
			s.sendSTTMessage(resp.Transcript)
		}
	case chatv1.VoiceEvent_VOICE_EVENT_TTS_START:
		s.sentenceStarted = false
		s.ttsActive = true
		s.sendTTSStart()
	case chatv1.VoiceEvent_VOICE_EVENT_CONTENT_CHUNK:
		s.responseText.WriteString(resp.ContentChunk)
		if !s.sentenceStarted && strings.TrimSpace(resp.ContentChunk) != "" {
			s.sentenceStarted = true
			s.sendJSON(map[string]any{"type": "tts", "state": "sentence_start", "text": resp.ContentChunk, "session_id": s.sessionID, "request_id": s.requestID})
		}
		s.sendJSON(map[string]any{"type": "llm", "text": resp.ContentChunk, "session_id": s.sessionID, "request_id": s.requestID})
	case chatv1.VoiceEvent_VOICE_EVENT_AUDIO_CHUNK:
		return s.sendAudioChunk(resp)
	case chatv1.VoiceEvent_VOICE_EVENT_TTS_END:
		if text := strings.TrimSpace(s.responseText.String()); text != "" {
			s.appendDialogue("assistant", text)
		}
		s.sendTTSStop()
		s.ttsActive = false
		s.responseText.Reset()
		if s.closeAfterChat {
			_ = s.conn.Close()
		}
	case chatv1.VoiceEvent_VOICE_EVENT_ERROR:
		s.logger.Error("voice_event_error", "code", resp.ErrorCode, "error", resp.ErrorMessage)
	}
	return nil
}

func (s *gatewaySession) sendAudioChunk(resp *chatv1.VoiceStreamResponse) error {
	codec := strings.ToLower(resp.Codec)
	payload := resp.AudioChunk
	if codec == "" {
		codec = "opus"
	}
	frameMs := int(resp.FrameDurationMs)
	if frameMs <= 0 {
		frameMs = effectiveInt(s.ac().Xiaozhi.AudioParams.FrameDuration, 60)
	}

	if codec == "pcm_s16le" || codec == "pcm" || codec == "s16le" {
		sampleRate := int(resp.SampleRateHz)
		if sampleRate <= 0 {
			sampleRate = 24000
		}
		channels := int(resp.Channels)
		if channels <= 0 {
			channels = 1
		}
		if err := s.ensureOpusEncoder(sampleRate, channels); err != nil {
			return err
		}
		frames, err := s.opusEncoder.EncodeS16LEFrames(payload, frameMs)
		if err != nil {
			return err
		}
		s.logger.Debug("tts_send_audio_chunk",
			"pcm_bytes", len(payload),
			"opus_frames", len(frames),
			"frame_ms", frameMs,
			"sample_rate_hz", sampleRate,
		)
		for _, frame := range frames {
			pkt := frame
			if s.fromMQTT {
				pkt = s.withMQTTHeader(frame)
			}
			if err := s.sendAudio(pkt, frameMs); err != nil {
				return err
			}
		}
		return nil
	}

	if codec != "opus" {
		return fmt.Errorf("unsupported audio codec: %s", resp.Codec)
	}
	if s.fromMQTT {
		payload = s.withMQTTHeader(payload)
	}
	return s.sendAudio(payload, frameMs)
}

func (s *gatewaySession) ensureOpusEncoder(sampleRate, channels int) error {
	if s.opusEncoder != nil &&
		s.opusEncoder.SampleRate() == sampleRate &&
		s.opusEncoder.Channels() == channels {
		return nil
	}
	enc, err := adapter.NewOpusEncoder(sampleRate, channels)
	if err != nil {
		return err
	}
	s.opusEncoder = enc
	return nil
}

func (s *gatewaySession) sendAudio(packet []byte, durationMs int) error {
	if len(packet) == 0 {
		return nil
	}
	if s.rateController == nil {
		s.rateController = NewAudioRateController(
			effectiveInt(s.ac().Xiaozhi.AudioParams.FrameDuration, 60),
			s.ac().Gateway.TTSAudioSendDelayMs,
			s.writeAudioPacket,
		)
	}
	return s.rateController.AddAudio(packet, durationMs)
}

func (s *gatewaySession) writeAudioPacket(packet []byte) error {
	s.lastActivity = time.Now()
	s.sendMu.Lock()
	defer s.sendMu.Unlock()
	if err := s.conn.WriteMessage(websocket.BinaryMessage, packet); err != nil {
		s.logger.Warn("ws_write_audio_failed", "bytes", len(packet), "error", err)
		return err
	}
	return nil
}

func (s *gatewaySession) waitAudioCompletion() {
	if s.rateController != nil {
		s.rateController.WaitEmpty()
	}
}

func (s *gatewaySession) sendTTSStart() {
	if s.rateController != nil {
		s.rateController.Reset()
	}
	s.sendJSON(map[string]any{"type": "tts", "state": "start", "session_id": s.sessionID, "request_id": s.requestID})
}

func (s *gatewaySession) sendTTSStop() {
	if s.ac().Gateway.EnableStopTTSNotify {
		s.playAudioFile(s.ac().Gateway.StopTTSNotifyVoice, "")
	}
	s.waitAudioCompletion()
	s.sendJSON(map[string]any{"type": "tts", "state": "stop", "session_id": s.sessionID, "request_id": s.requestID})
}

func (s *gatewaySession) playAudioFile(path string, text string) bool {
	packets, err := loadAudioFileAsOpus(path, effectiveInt(s.ac().Xiaozhi.AudioParams.SampleRate, 16000), effectiveInt(s.ac().Xiaozhi.AudioParams.Channels, 1), effectiveInt(s.ac().Xiaozhi.AudioParams.FrameDuration, 60))
	if err != nil || len(packets) == 0 {
		s.logger.Warn("play_audio_file_failed", "path", path, "error", err)
		return false
	}
	if strings.TrimSpace(text) != "" {
		s.sendJSON(map[string]any{"type": "tts", "state": "sentence_start", "text": text, "session_id": s.sessionID})
	}
	frameMs := effectiveInt(s.ac().Xiaozhi.AudioParams.FrameDuration, 60)
	for _, pkt := range packets {
		_ = s.sendAudio(pkt, frameMs)
	}
	s.waitAudioCompletion()
	return true
}

func (s *gatewaySession) shouldContinueAfterTranscript(transcript string, speaker string) bool {
	filtered := normalizeSpeechText(transcript)
	if !s.isWakeupWord(filtered) || !s.ac().Gateway.EnableWakeupWordsResponseCache {
		return true
	}
	path, text := readWakeupResponse(s.ac().Gateway.WakeupWordsConfigFile, s.ac().Gateway.WakeupWordsFallbackFile)
	s.sendTTSStart()
	if s.playAudioFile(path, text) {
		s.sendTTSStop()
		s.appendDialogue("assistant", text)
		return false
	}
	return true
}

func (s *gatewaySession) withMQTTHeader(opusPacket []byte) []byte {
	s.outSeq++
	header := make([]byte, 16)
	header[0] = 1
	header[2] = byte(len(opusPacket) >> 8)
	header[3] = byte(len(opusPacket))
	header[4] = byte(s.outSeq >> 24)
	header[5] = byte(s.outSeq >> 16)
	header[6] = byte(s.outSeq >> 8)
	header[7] = byte(s.outSeq)
	ts := uint32(time.Now().UnixMilli())
	header[8] = byte(ts >> 24)
	header[9] = byte(ts >> 16)
	header[10] = byte(ts >> 8)
	header[11] = byte(ts)
	header[12] = byte(len(opusPacket) >> 24)
	header[13] = byte(len(opusPacket) >> 16)
	header[14] = byte(len(opusPacket) >> 8)
	header[15] = byte(len(opusPacket))
	return append(header, opusPacket...)
}

func (s *gatewaySession) voiceStartRequest() *chatv1.VoiceStreamRequest {
	return &chatv1.VoiceStreamRequest{
		Type:              chatv1.VoiceRequestType_VOICE_REQUEST_TYPE_START,
		SessionId:         s.sessionID,
		RequestId:         s.requestID,
		DeviceId:          s.deviceID,
		ClientId:          s.clientID,
		ListenMode:        s.listenMode,
		VoiceprintEnabled: s.ac().Voiceprint.Enabled && strings.TrimSpace(s.ac().Voiceprint.URL) != "",
		MetadataJson:      s.metadataJSON(),
	}
}

func (s *gatewaySession) chatRequest(query string) *chatv1.StreamChatRequest {
	messages := s.recentDialogue()
	messages = append(messages, &chatv1.Message{Role: "user", Content: query})
	s.appendDialogue("user", query)
	return &chatv1.StreamChatRequest{
		SessionId:         s.sessionID,
		RequestId:         s.requestID,
		DeviceId:          s.deviceID,
		ClientId:          s.clientID,
		Query:             query,
		Messages:          messages,
		VoiceprintEnabled: s.ac().Voiceprint.Enabled && strings.TrimSpace(s.ac().Voiceprint.URL) != "",
		MetadataJson:      s.metadataJSON(),
	}
}

func (s *gatewaySession) appendDialogue(role string, content string) {
	content = strings.TrimSpace(content)
	if content == "" {
		return
	}
	s.dialogue = append(s.dialogue, &chatv1.Message{Role: role, Content: content})
	if len(s.dialogue) > 20 {
		s.dialogue = s.dialogue[len(s.dialogue)-20:]
	}
}

func (s *gatewaySession) recentDialogue() []*chatv1.Message {
	start := 0
	if len(s.dialogue) > 10 {
		start = len(s.dialogue) - 10
	}
	out := make([]*chatv1.Message, 0, len(s.dialogue)-start+1)
	if strings.TrimSpace(s.ac().Gateway.SystemPrompt) != "" {
		out = append(out, &chatv1.Message{Role: "system", Content: s.ac().Gateway.SystemPrompt})
	}
	for _, msg := range s.dialogue[start:] {
		cp := *msg
		out = append(out, &cp)
	}
	return out
}

func (s *gatewaySession) metadataJSON() string {
	payload := map[string]any{
		"functions": []any{},
		"tts_config": map[string]any{
			"enabled":           s.ac().TTS.Enabled,
			"streams":           s.ac().TTS.MaxParallel,
			"codec":             "pcm_s16le",
			"sample_rate_hz":    24000,
			"channels":          1,
			"frame_duration_ms": 60,
		},
	}
	data, _ := json.Marshal(payload)
	return string(data)
}

func (s *gatewaySession) sendJSON(payload any) {
	data, _ := json.Marshal(payload)
	s.sendMu.Lock()
	defer s.sendMu.Unlock()
	if err := s.conn.WriteMessage(websocket.TextMessage, data); err != nil {
		s.logger.Warn("ws_write_text_failed", "payload", string(data), "error", err)
	}
}

func (s *gatewaySession) sendSTTMessage(text string) {
	if strings.TrimSpace(s.ac().Gateway.EndPrompt) != "" && s.ac().Gateway.EndPrompt == text {
		s.sendTTSStart()
		return
	}
	displayText := text
	var payload map[string]any
	trimmed := strings.TrimSpace(text)
	if strings.HasPrefix(trimmed, "{") && strings.HasSuffix(trimmed, "}") {
		if err := json.Unmarshal([]byte(trimmed), &payload); err == nil {
			if content, ok := payload["content"].(string); ok {
				displayText = content
			}
		}
	}
	s.sendJSON(map[string]any{"type": "stt", "text": trimPunctuationOrEmoji(displayText), "session_id": s.sessionID, "request_id": s.requestID})
	s.sendTTSStart()
}

func randomID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b[:])
}

func effectiveInt(value int, fallback int) int {
	if value > 0 {
		return value
	}
	return fallback
}

func normalizeSpeechText(text string) string {
	var b strings.Builder
	for _, r := range text {
		if unicode.IsSpace(r) || unicode.IsPunct(r) || unicode.IsSymbol(r) {
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

func trimPunctuationOrEmoji(text string) string {
	runes := []rune(text)
	start := 0
	for start < len(runes) && isPunctuationOrEmoji(runes[start]) {
		start++
	}
	end := len(runes) - 1
	for end >= start && isPunctuationOrEmoji(runes[end]) {
		end--
	}
	if start > end {
		return ""
	}
	return string(runes[start : end+1])
}

func isPunctuationOrEmoji(r rune) bool {
	if unicode.IsSpace(r) {
		return true
	}
	if strings.ContainsRune("，,。.！!“”\"：:-－、[]【】", r) {
		return true
	}
	return isEmoji(r)
}

func isEmoji(r rune) bool {
	return (r >= 0x1F600 && r <= 0x1F64F) ||
		(r >= 0x1F300 && r <= 0x1F5FF) ||
		(r >= 0x1F680 && r <= 0x1F6FF) ||
		(r >= 0x1F900 && r <= 0x1F9FF) ||
		(r >= 0x1FA70 && r <= 0x1FAFF) ||
		(r >= 0x2600 && r <= 0x26FF) ||
		(r >= 0x2700 && r <= 0x27BF)
}
