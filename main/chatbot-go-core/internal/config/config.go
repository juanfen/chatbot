package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type ServerConfig struct {
	RequestTimeoutMs    int    `json:"request_timeout_ms"`
	StreamIdleTimeoutMs int    `json:"stream_idle_timeout_ms"`
	MaxConcurrency      int    `json:"max_concurrency"`
}

type GatewayAuthConfig struct {
	Enabled        bool     `json:"enabled"`
	AllowedDevices []string `json:"allowed_devices"`
	ExpireSeconds  int64    `json:"expire_seconds"`
}

type GatewayConfig struct {
	Enabled                        bool              `json:"enabled"`
	WSListenAddr                   string            `json:"ws_listen_addr"`
	HTTPListenAddr                 string            `json:"http_listen_addr"`
	WSPingIntervalSeconds          int               `json:"ws_ping_interval_seconds"`
	WSPongTimeoutSeconds           int               `json:"ws_pong_timeout_seconds"`
	WebSocketURL                   string            `json:"websocket_url"`
	AuthKey                        string            `json:"auth_key"`
	Auth                           GatewayAuthConfig `json:"auth"`
	MQTTGateway                    string            `json:"mqtt_gateway"`
	WakeupWords                    []string          `json:"wakeup_words"`
	EnableGreeting                 bool              `json:"enable_greeting"`
	CloseConnectionNoVoiceTime     int               `json:"close_connection_no_voice_time"`
	EndPrompt                      string            `json:"end_prompt"`
	EndPromptEnabled               bool              `json:"end_prompt_enabled"`
	SystemPrompt                   string            `json:"system_prompt"`
	EnableStopTTSNotify            bool              `json:"enable_stop_tts_notify"`
	StopTTSNotifyVoice             string            `json:"stop_tts_notify_voice"`
	TTSAudioSendDelayMs            int               `json:"tts_audio_send_delay_ms"`
	EnableWakeupWordsResponseCache bool              `json:"enable_wakeup_words_response_cache"`
	WakeupWordsConfigFile          string            `json:"wakeup_words_config_file"`
	WakeupWordsFallbackFile        string            `json:"wakeup_words_fallback_file"`
}

type OTAConfig struct {
	Enabled          bool   `json:"enabled"`
	BinDir           string `json:"bin_dir"`
	FirmwareCacheTTL int    `json:"firmware_cache_ttl"`
}

type AudioParamsConfig struct {
	Format        string `json:"format"`
	SampleRate    int    `json:"sample_rate"`
	Channels      int    `json:"channels"`
	FrameDuration int    `json:"frame_duration"`
}

type XiaozhiConfig struct {
	Type        string            `json:"type"`
	Version     int               `json:"version"`
	Transport   string            `json:"transport"`
	AudioParams AudioParamsConfig `json:"audio_params"`
}

type LLMConfig struct {
	BaseURL     string            `json:"base_url"`
	APIKey      string            `json:"api_key"`
	ModelName   string            `json:"model_name"`
	TimeoutMs   int               `json:"timeout_ms"`
	UseExternal bool              `json:"use_external"`
	External    LLMExternalConfig `json:"external"`
	// DisableThinkTagsStrip 为 true 时，SSE 正文不再经过 stripThinkTagsStream（避免误伤或与模型实际标签不一致）。
	DisableThinkTagsStrip bool `json:"disable_think_tags_strip"`
}

type LLMExternalConfig struct {
	BaseURL   string `json:"base_url"`
	APIKey    string `json:"api_key"`
	ModelName string `json:"model_name"`
	TimeoutMs int    `json:"timeout_ms"`
}

// TTSServiceConfig 与 xiaozhi-server 中 IndexStreamTTS（index_stream.py）一致：
// POST JSON {"text","character"}，流式 16-bit LE PCM；未写 sample_rate_hz 时默认 24000。
type TTSServiceConfig struct {
	Name            string            `json:"name"`
	Enabled         bool              `json:"enabled"`
	URL             string            `json:"url,omitempty"`
	APIURL          string            `json:"api_url,omitempty"`
	Voice           string            `json:"voice"`
	APIKey          string            `json:"api_key,omitempty"`
	TimeoutMs       int               `json:"timeout_ms"`
	SampleRateHz    int               `json:"sample_rate_hz"`
	Channels        int               `json:"channels"`
	FrameDurationMs int               `json:"frame_duration_ms"`
	Headers         map[string]string `json:"headers"`
}

// ResolveURL 返回 TTS 接口地址（兼容 url 与 api_url 两种配置键）。
func (s *TTSServiceConfig) ResolveURL() string {
	if s.URL != "" {
		return s.URL
	}
	return s.APIURL
}

// EffectiveSampleRateHz 未配置时与 index_stream.py 中 OpusEncoderUtils(sample_rate=24000) 一致。
func (s *TTSServiceConfig) EffectiveSampleRateHz() int {
	if s.SampleRateHz > 0 {
		return s.SampleRateHz
	}
	return 24000
}

// Pickable 表示该条 service 是否参与调度（需 api_url 或 url，且 voice 非空）。
func (s *TTSServiceConfig) Pickable() bool {
	return s.Enabled && s.ResolveURL() != "" && strings.TrimSpace(s.Voice) != ""
}

type TTSConfig struct {
	Enabled         bool               `json:"enabled"`
	MaxParallel     int                `json:"max_parallel"`
	SegmentMaxChars int                `json:"segment_max_chars"`
	FlushMinChars   int                `json:"flush_min_chars"`
	FlushIntervalMs int                `json:"flush_interval_ms"`
	// FlushStrongSuffixes 为指针：nil 表示使用内置默认；&[]string{} 表示显式关闭强标点刷新。
	FlushStrongSuffixes *[]string `json:"flush_strong_suffixes,omitempty"`
	// FlushWeakSuffixes 为指针：nil 表示使用内置默认；弱标点不会单独触发刷新，仅用于识别“不要当句末处理”。
	FlushWeakSuffixes *[]string `json:"flush_weak_suffixes,omitempty"`
	Services          []TTSServiceConfig `json:"services"`
}

// DefaultFlushStrongSuffixes 尾部落在此类标点上时可触发提前送 TTS（与 flush_min_chars 等组合使用）。
func DefaultFlushStrongSuffixes() []string {
	return []string{"。", "！", "？", ".", "!", "?", "\n"}
}

// DefaultFlushWeakSuffixes 此类标点单独出现时不在此处触发刷新（继续攒字，除非已达字数/时间阈值）。
func DefaultFlushWeakSuffixes() []string {
	return []string{"，", ",", "；", ";", "：", ":"}
}

// EffectiveFlushStrongSuffixes 返回当前生效的强标点列表。
func EffectiveFlushStrongSuffixes(tts *TTSConfig) []string {
	if tts == nil || tts.FlushStrongSuffixes == nil {
		return DefaultFlushStrongSuffixes()
	}
	return *tts.FlushStrongSuffixes
}

// EffectiveFlushWeakSuffixes 返回当前生效的弱标点列表。
func EffectiveFlushWeakSuffixes(tts *TTSConfig) []string {
	if tts == nil || tts.FlushWeakSuffixes == nil {
		return DefaultFlushWeakSuffixes()
	}
	return *tts.FlushWeakSuffixes
}

type VADConfig struct {
	BaseURL              string  `json:"base_url"`
	TimeoutMs            int     `json:"timeout_ms"`
	Threshold            float64 `json:"threshold"`
	ThresholdLow         float64 `json:"threshold_low"`
	MinSilenceDurationMs int     `json:"min_silence_duration_ms"`
	FrameWindowThreshold int     `json:"frame_window_threshold"`
	APIKey               string  `json:"api_key"`
}

type ASRConfig struct {
	Host      string `json:"host"`
	Port      int    `json:"port"`
	IsSSL     bool   `json:"is_ssl"`
	APIKey    string `json:"api_key"`
	TimeoutMs int    `json:"timeout_ms"`
}

type VoiceprintConfig struct {
	Enabled             bool     `json:"enabled"`
	URL                 string   `json:"url"`
	APIKey              string   `json:"api_key"`
	Speakers            []string `json:"speakers"`
	SimilarityThreshold float64  `json:"similarity_threshold"`
	TimeoutMs           int      `json:"timeout_ms"`
}

// ChatbotAdminConfig 长音频上报至 chatbot-go-admin（/api/v1/device/audio/upload）。
type ChatbotAdminConfig struct {
	Enabled           bool   `json:"enabled"`
	IP                string `json:"ip"`
	Port              int    `json:"port"`
	RecordAudioLenSec int    `json:"record_audio_len_sec"`
	TimeoutMs         int    `json:"timeout_ms"`
}

// EffectiveRecordAudioLenSec 未配置时默认 5 秒。
func (c ChatbotAdminConfig) EffectiveRecordAudioLenSec() int {
	if c.RecordAudioLenSec > 0 {
		return c.RecordAudioLenSec
	}
	return 5
}

// APIBaseURL 返回 chatbot-go-admin HTTP 根地址（不含路径）。
func (c ChatbotAdminConfig) APIBaseURL() string {
	ip := strings.TrimSpace(c.IP)
	if ip == "" || c.Port <= 0 {
		return ""
	}
	return fmt.Sprintf("http://%s:%d", ip, c.Port)
}

// UploadURL 返回设备音频上传地址。
func (c ChatbotAdminConfig) UploadURL() string {
	base := c.APIBaseURL()
	if base == "" {
		return ""
	}
	return base + "/api/v1/device/audio/upload"
}

// SpeakerIDsURL 返回按 mac 查询已注册 speaker_id 列表的地址。
func (c ChatbotAdminConfig) SpeakerIDsURL() string {
	base := c.APIBaseURL()
	if base == "" {
		return ""
	}
	return base + "/api/v1/device/audio/speaker-ids"
}

// SpeakerNameURL 返回按 speaker_id 查询音色名称的地址。
func (c ChatbotAdminConfig) SpeakerNameURL() string {
	base := c.APIBaseURL()
	if base == "" {
		return ""
	}
	return base + "/api/v1/device/audio/speaker-name"
}

type AudioPipelineConfig struct {
	Enabled              bool `json:"enabled"`
	PrerollFrames        int  `json:"preroll_frames"`
	MinUtteranceFrames   int  `json:"min_utterance_frames"`
	AudioFrameBufferSize int  `json:"audio_frame_buffer_size"`
}

// ShortTermMemoryConfig 短期记忆 / RAG 服务（与 Python 侧 /short-term/build_prompt、/short-term/commit_turn 对齐）。
type ShortTermMemoryConfig struct {
	Enabled       bool   `json:"enabled"`
	BaseURL       string `json:"base_url"`
	TimeoutMs     int    `json:"timeout_ms"`
	DefaultUserID string `json:"default_user_id"`
}

type AppConfig struct {
	Server          ServerConfig          `json:"server"`
	Gateway         GatewayConfig         `json:"gateway"`
	OTA             OTAConfig             `json:"ota"`
	Xiaozhi         XiaozhiConfig         `json:"xiaozhi"`
	LLM             LLMConfig             `json:"llm"`
	TTS             TTSConfig             `json:"tts"`
	VAD             VADConfig             `json:"vad"`
	ASR             ASRConfig             `json:"asr"`
	Voiceprint      VoiceprintConfig      `json:"voiceprint"`
	ChatbotAdmin    ChatbotAdminConfig    `json:"chatbot_admin"`
	AudioPipeline   AudioPipelineConfig   `json:"audio_pipeline"`
	ShortTermMemory ShortTermMemoryConfig `json:"short_term_memory"`
}

func DefaultConfig() *AppConfig {
	return &AppConfig{
		Server: ServerConfig{
			RequestTimeoutMs:    8000,
			StreamIdleTimeoutMs: 15000,
			MaxConcurrency:      100,
		},
		Gateway: GatewayConfig{
			Enabled:               true,
			WSListenAddr:          ":8000",
			HTTPListenAddr:        ":8003",
			WSPingIntervalSeconds: 15,
			WSPongTimeoutSeconds:  30,
			WebSocketURL:          "你的websocket地址",
			AuthKey:               "",
			WakeupWords: []string{
				"你好小智",
				"嘿你好呀",
				"你好小志",
			},
			EnableGreeting:                 true,
			CloseConnectionNoVoiceTime:     120,
			EndPromptEnabled:               true,
			EndPrompt:                      "请你以```时间过得真快```未来头，用富有感情、依依不舍的话来结束这场对话吧。！",
			SystemPrompt:                   "你是玛奇玛，电次是你的狗",
			EnableStopTTSNotify:            false,
			StopTTSNotifyVoice:             "config/assets/tts_notify.mp3",
			TTSAudioSendDelayMs:            -1,
			EnableWakeupWordsResponseCache: false,
			WakeupWordsConfigFile:          "data/.wakeup_words.yaml",
			WakeupWordsFallbackFile:        "config/assets/wakeup_words_short.wav",
			Auth: GatewayAuthConfig{
				Enabled:        false,
				AllowedDevices: []string{},
				ExpireSeconds:  60 * 60 * 24 * 30,
			},
		},
		OTA: OTAConfig{
			Enabled:          true,
			BinDir:           "data/bin",
			FirmwareCacheTTL: 30,
		},
		Xiaozhi: XiaozhiConfig{
			Type:      "hello",
			Version:   1,
			Transport: "websocket",
			AudioParams: AudioParamsConfig{
				Format:        "opus",
				SampleRate:    16000,
				Channels:      1,
				FrameDuration: 60,
			},
		},
		LLM: LLMConfig{
			BaseURL:     "https://open.bigmodel.cn/api/paas/v4/",
			APIKey:      "",
			ModelName:   "glm-4-flash",
			TimeoutMs:   6000,
			UseExternal: false,
			External: LLMExternalConfig{
				BaseURL:   "https://api.openai.com/v1",
				APIKey:    "",
				ModelName: "gpt-4o-mini",
				TimeoutMs: 8000,
			},
		},
		TTS: TTSConfig{
			Enabled:         false,
			MaxParallel:     4,
			SegmentMaxChars: 80,
			FlushMinChars:   40,
			FlushIntervalMs: 600,
			Services:        []TTSServiceConfig{},
		},
		VAD: VADConfig{
			BaseURL:              "http://127.0.0.1:8766",
			TimeoutMs:            1500,
			Threshold:            0.5,
			ThresholdLow:         0.3,
			MinSilenceDurationMs: 200,
			FrameWindowThreshold: 3,
		},
		ASR: ASRConfig{
			Host:      "127.0.0.1",
			Port:      10095,
			IsSSL:     false,
			APIKey:    "none",
			TimeoutMs: 15000,
		},
		Voiceprint: VoiceprintConfig{
			Enabled:             false,
			Speakers:            []string{},
			SimilarityThreshold: 0.4,
			TimeoutMs:           10000,
		},
		ChatbotAdmin: ChatbotAdminConfig{
			Enabled:           false,
			IP:                "127.0.0.1",
			Port:              8088,
			RecordAudioLenSec: 5,
			TimeoutMs:         15000,
		},
		AudioPipeline: AudioPipelineConfig{
			Enabled:              false,
			PrerollFrames:        10,
			MinUtteranceFrames:   15,
			AudioFrameBufferSize: 128,
		},
		ShortTermMemory: ShortTermMemoryConfig{
			Enabled:   false,
			BaseURL:   "",
			TimeoutMs: 30000,
		},
	}
}

func LoadFromFile(path string) (*AppConfig, error) {
	cfg := DefaultConfig()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, err
	}
	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, err
	}
	applyEnvOverrides(cfg)
	return cfg, nil
}

func applyEnvOverrides(cfg *AppConfig) {
	if cfg == nil {
		return
	}
	if v := os.Getenv("XIAOZHI_GO_LLM_API_KEY"); v != "" {
		cfg.LLM.APIKey = v
	}
	if v := os.Getenv("XIAOZHI_GO_LLM_BASE_URL"); v != "" {
		cfg.LLM.BaseURL = v
	}
	if v := os.Getenv("XIAOZHI_GO_LLM_MODEL"); v != "" {
		cfg.LLM.ModelName = v
	}
	if v := os.Getenv("XIAOZHI_GO_LLM_USE_EXTERNAL"); v != "" {
		cfg.LLM.UseExternal = parseEnvBool(v, cfg.LLM.UseExternal)
	}
	if v := os.Getenv("XIAOZHI_GO_LLM_EXTERNAL_BASE_URL"); v != "" {
		cfg.LLM.External.BaseURL = v
	}
	if v := os.Getenv("XIAOZHI_GO_LLM_EXTERNAL_API_KEY"); v != "" {
		cfg.LLM.External.APIKey = v
	}
	if v := os.Getenv("XIAOZHI_GO_LLM_EXTERNAL_MODEL"); v != "" {
		cfg.LLM.External.ModelName = v
	}
	if v := os.Getenv("XIAOZHI_GO_WS_LISTEN_ADDR"); v != "" {
		cfg.Gateway.WSListenAddr = v
	}
	if v := os.Getenv("XIAOZHI_GO_HTTP_LISTEN_ADDR"); v != "" {
		cfg.Gateway.HTTPListenAddr = v
	}
	if v := os.Getenv("XIAOZHI_GO_VAD_BASE_URL"); v != "" {
		cfg.VAD.BaseURL = v
	}
	if v := os.Getenv("XIAOZHI_GO_ASR_HOST"); v != "" {
		cfg.ASR.Host = v
	}
}

// RedactAPIKey masks an API key for safe logging.
func RedactAPIKey(key string) string {
	if key == "" {
		return "(empty)"
	}
	if len(key) <= 12 {
		return "***"
	}
	return key[:6] + "..." + key[len(key)-4:]
}

// ToLogJSON returns indented JSON of the effective config with api_key redacted.
func ToLogJSON(cfg *AppConfig) ([]byte, error) {
	if cfg == nil {
		return []byte("{}"), nil
	}
	payload := map[string]any{
		"server": map[string]any{
			"request_timeout_ms":     cfg.Server.RequestTimeoutMs,
			"stream_idle_timeout_ms": cfg.Server.StreamIdleTimeoutMs,
			"max_concurrency":        cfg.Server.MaxConcurrency,
		},
		"gateway": map[string]any{
			"enabled":                            cfg.Gateway.Enabled,
			"ws_listen_addr":                     cfg.Gateway.WSListenAddr,
			"http_listen_addr":                   cfg.Gateway.HTTPListenAddr,
			"ws_ping_interval_seconds":           cfg.Gateway.WSPingIntervalSeconds,
			"ws_pong_timeout_seconds":            cfg.Gateway.WSPongTimeoutSeconds,
			"websocket_url":                      cfg.Gateway.WebSocketURL,
			"auth_enabled":                       cfg.Gateway.Auth.Enabled,
			"auth_key":                           RedactAPIKey(cfg.Gateway.AuthKey),
			"allowed_devices":                    len(cfg.Gateway.Auth.AllowedDevices),
			"mqtt_gateway":                       cfg.Gateway.MQTTGateway,
			"wakeup_words":                       len(cfg.Gateway.WakeupWords),
			"enable_greeting":                    cfg.Gateway.EnableGreeting,
			"close_connection_no_voice_time":     cfg.Gateway.CloseConnectionNoVoiceTime,
			"end_prompt_enabled":                 cfg.Gateway.EndPromptEnabled,
			"system_prompt_set":                  strings.TrimSpace(cfg.Gateway.SystemPrompt) != "",
			"enable_stop_tts_notify":             cfg.Gateway.EnableStopTTSNotify,
			"tts_audio_send_delay_ms":            cfg.Gateway.TTSAudioSendDelayMs,
			"enable_wakeup_words_response_cache": cfg.Gateway.EnableWakeupWordsResponseCache,
		},
		"ota": map[string]any{
			"enabled":            cfg.OTA.Enabled,
			"bin_dir":            cfg.OTA.BinDir,
			"firmware_cache_ttl": cfg.OTA.FirmwareCacheTTL,
		},
		"xiaozhi": map[string]any{
			"type":      cfg.Xiaozhi.Type,
			"version":   cfg.Xiaozhi.Version,
			"transport": cfg.Xiaozhi.Transport,
			"audio_params": map[string]any{
				"format":         cfg.Xiaozhi.AudioParams.Format,
				"sample_rate":    cfg.Xiaozhi.AudioParams.SampleRate,
				"channels":       cfg.Xiaozhi.AudioParams.Channels,
				"frame_duration": cfg.Xiaozhi.AudioParams.FrameDuration,
			},
		},
		"llm": map[string]any{
			"base_url":                   cfg.LLM.BaseURL,
			"api_key":                    RedactAPIKey(cfg.LLM.APIKey),
			"model_name":                 cfg.LLM.ModelName,
			"timeout_ms":                 cfg.LLM.TimeoutMs,
			"use_external":               cfg.LLM.UseExternal,
			"disable_think_tags_strip":   cfg.LLM.DisableThinkTagsStrip,
			"active_source": map[bool]string{true: "external", false: "default"}[cfg.LLM.UseExternal],
			"external": map[string]any{
				"base_url":   cfg.LLM.External.BaseURL,
				"api_key":    RedactAPIKey(cfg.LLM.External.APIKey),
				"model_name": cfg.LLM.External.ModelName,
				"timeout_ms": cfg.LLM.External.TimeoutMs,
			},
		},
		"tts": func() map[string]any {
			es := EffectiveFlushStrongSuffixes(&cfg.TTS)
			ew := EffectiveFlushWeakSuffixes(&cfg.TTS)
			return map[string]any{
				"enabled":                      cfg.TTS.Enabled,
				"max_parallel":                 cfg.TTS.MaxParallel,
				"segment_max_chars":            cfg.TTS.SegmentMaxChars,
				"flush_min_chars":              cfg.TTS.FlushMinChars,
				"flush_interval_ms":            cfg.TTS.FlushIntervalMs,
				"flush_strong_suffixes_len":    len(es),
				"flush_weak_suffixes_len":      len(ew),
				"flush_strong_suffixes_custom": cfg.TTS.FlushStrongSuffixes != nil,
				"flush_weak_suffixes_custom":   cfg.TTS.FlushWeakSuffixes != nil,
				"services_count":               len(cfg.TTS.Services),
			}
		}(),
		"vad": map[string]any{
			"base_url":                cfg.VAD.BaseURL,
			"timeout_ms":              cfg.VAD.TimeoutMs,
			"threshold":               cfg.VAD.Threshold,
			"threshold_low":           cfg.VAD.ThresholdLow,
			"min_silence_duration_ms": cfg.VAD.MinSilenceDurationMs,
			"frame_window_threshold":  cfg.VAD.FrameWindowThreshold,
			"api_key":                 RedactAPIKey(cfg.VAD.APIKey),
		},
		"asr": map[string]any{
			"host":       cfg.ASR.Host,
			"port":       cfg.ASR.Port,
			"is_ssl":     cfg.ASR.IsSSL,
			"api_key":    RedactAPIKey(cfg.ASR.APIKey),
			"timeout_ms": cfg.ASR.TimeoutMs,
		},
		"voiceprint": map[string]any{
			"enabled":              cfg.Voiceprint.Enabled,
			"url_set":              cfg.Voiceprint.URL != "",
			"speakers_count":       len(cfg.Voiceprint.Speakers),
			"similarity_threshold": cfg.Voiceprint.SimilarityThreshold,
			"timeout_ms":           cfg.Voiceprint.TimeoutMs,
		},
		"chatbot_admin": map[string]any{
			"enabled":               cfg.ChatbotAdmin.Enabled,
			"ip":                    cfg.ChatbotAdmin.IP,
			"port":                  cfg.ChatbotAdmin.Port,
			"record_audio_len_sec":  cfg.ChatbotAdmin.RecordAudioLenSec,
			"timeout_ms":            cfg.ChatbotAdmin.TimeoutMs,
			"upload_url_set":        cfg.ChatbotAdmin.UploadURL() != "",
		},
		"short_term_memory": map[string]any{
			"enabled":              cfg.ShortTermMemory.Enabled,
			"base_url_set":         strings.TrimSpace(cfg.ShortTermMemory.BaseURL) != "",
			"timeout_ms":           cfg.ShortTermMemory.TimeoutMs,
			"default_user_id_set":  strings.TrimSpace(cfg.ShortTermMemory.DefaultUserID) != "",
		},
		"audio_pipeline": map[string]any{
			"enabled":                 cfg.AudioPipeline.Enabled,
			"preroll_frames":          cfg.AudioPipeline.PrerollFrames,
			"min_utterance_frames":    cfg.AudioPipeline.MinUtteranceFrames,
			"audio_frame_buffer_size": cfg.AudioPipeline.AudioFrameBufferSize,
		},
		"env_overrides": map[string]any{
			"XIAOZHI_GO_LLM_API_KEY_set":      os.Getenv("XIAOZHI_GO_LLM_API_KEY") != "",
			"XIAOZHI_GO_LLM_BASE_URL_set":     os.Getenv("XIAOZHI_GO_LLM_BASE_URL") != "",
			"XIAOZHI_GO_LLM_MODEL_set":        os.Getenv("XIAOZHI_GO_LLM_MODEL") != "",
			"XIAOZHI_GO_LLM_USE_EXTERNAL_set": os.Getenv("XIAOZHI_GO_LLM_USE_EXTERNAL") != "",
			"XIAOZHI_GO_LLM_EXTERNAL_BASE_URL_set": os.Getenv("XIAOZHI_GO_LLM_EXTERNAL_BASE_URL") != "",
			"XIAOZHI_GO_LLM_EXTERNAL_API_KEY_set":  os.Getenv("XIAOZHI_GO_LLM_EXTERNAL_API_KEY") != "",
			"XIAOZHI_GO_LLM_EXTERNAL_MODEL_set":    os.Getenv("XIAOZHI_GO_LLM_EXTERNAL_MODEL") != "",
			"XIAOZHI_GO_WS_LISTEN_ADDR_set":   os.Getenv("XIAOZHI_GO_WS_LISTEN_ADDR") != "",
			"XIAOZHI_GO_HTTP_LISTEN_ADDR_set": os.Getenv("XIAOZHI_GO_HTTP_LISTEN_ADDR") != "",
			"XIAOZHI_GO_VAD_BASE_URL_set":     os.Getenv("XIAOZHI_GO_VAD_BASE_URL") != "",
			"XIAOZHI_GO_ASR_HOST_set":         os.Getenv("XIAOZHI_GO_ASR_HOST") != "",
		},
	}
	return json.MarshalIndent(payload, "", "  ")
}

func parseEnvBool(value string, fallback bool) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	default:
		return fallback
	}
}
