package config

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync/atomic"
)

// Store 保存当前进程内生效的配置快照（原子替换指针，读路径无锁）。
// 不写入磁盘；与启动时加载的 config.json 解耦，仅作为运行时真源。
type Store struct {
	v atomic.Pointer[AppConfig]
}

// NewStore 使用初始配置创建 Store（通常来自 LoadFromFile）。
func NewStore(initial *AppConfig) *Store {
	if initial == nil {
		initial = DefaultConfig()
	}
	s := &Store{}
	s.v.Store(initial)
	return s
}

// Get 返回当前生效配置的指针，只读；请勿修改返回值指向的结构体。
func (s *Store) Get() *AppConfig {
	return s.v.Load()
}

// CloneAppConfig 通过 JSON 深拷贝配置。
func CloneAppConfig(src *AppConfig) (*AppConfig, error) {
	if src == nil {
		return nil, fmt.Errorf("nil app config")
	}
	b, err := json.Marshal(src)
	if err != nil {
		return nil, err
	}
	var dst AppConfig
	if err := json.Unmarshal(b, &dst); err != nil {
		return nil, err
	}
	return &dst, nil
}

var runtimeSectionKeys = map[string]struct{}{
	"gateway":           {},
	"ota":               {},
	"llm":               {},
	"tts":               {},
	"vad":               {},
	"asr":               {},
	"voiceprint":        {},
	"chatbot_admin":     {},
	"audio_pipeline":    {},
	"short_term_memory": {},
}

// ApplyRuntimePatchJSON 合并 JSON 补丁到当前配置并原子替换生效快照。
// 仅允许 gateway、ota、llm、tts、vad、asr、voiceprint、chatbot_admin、audio_pipeline、short_term_memory 顶层键；
// 对每个对象做深度合并（嵌套 map 递归合并；数组整体替换）。
func (s *Store) ApplyRuntimePatchJSON(data []byte) error {
	var patch map[string]json.RawMessage
	if err := json.Unmarshal(data, &patch); err != nil {
		return fmt.Errorf("parse patch json: %w", err)
	}
	if len(patch) == 0 {
		return fmt.Errorf("empty patch")
	}
	for k := range patch {
		if _, ok := runtimeSectionKeys[k]; !ok {
			return fmt.Errorf("unknown or disallowed key %q (allowed: gateway, ota, llm, tts, vad, asr, voiceprint, chatbot_admin, audio_pipeline, short_term_memory)", k)
		}
	}

	cur := s.Get()
	next, err := CloneAppConfig(cur)
	if err != nil {
		return err
	}

	if raw, ok := patch["gateway"]; ok {
		if err := mergeJSONInto(&next.Gateway, raw); err != nil {
			return fmt.Errorf("gateway: %w", err)
		}
	}
	if raw, ok := patch["ota"]; ok {
		if err := mergeJSONInto(&next.OTA, raw); err != nil {
			return fmt.Errorf("ota: %w", err)
		}
	}
	if raw, ok := patch["llm"]; ok {
		if err := mergeJSONInto(&next.LLM, raw); err != nil {
			return fmt.Errorf("llm: %w", err)
		}
	}
	if raw, ok := patch["tts"]; ok {
		if err := mergeJSONInto(&next.TTS, raw); err != nil {
			return fmt.Errorf("tts: %w", err)
		}
	}
	if raw, ok := patch["vad"]; ok {
		if err := mergeJSONInto(&next.VAD, raw); err != nil {
			return fmt.Errorf("vad: %w", err)
		}
	}
	if raw, ok := patch["asr"]; ok {
		if err := mergeJSONInto(&next.ASR, raw); err != nil {
			return fmt.Errorf("asr: %w", err)
		}
	}
	if raw, ok := patch["voiceprint"]; ok {
		if err := mergeJSONInto(&next.Voiceprint, raw); err != nil {
			return fmt.Errorf("voiceprint: %w", err)
		}
	}
	if raw, ok := patch["chatbot_admin"]; ok {
		if err := mergeJSONInto(&next.ChatbotAdmin, raw); err != nil {
			return fmt.Errorf("chatbot_admin: %w", err)
		}
	}
	if raw, ok := patch["audio_pipeline"]; ok {
		if err := mergeJSONInto(&next.AudioPipeline, raw); err != nil {
			return fmt.Errorf("audio_pipeline: %w", err)
		}
	}
	if raw, ok := patch["short_term_memory"]; ok {
		if err := mergeJSONInto(&next.ShortTermMemory, raw); err != nil {
			return fmt.Errorf("short_term_memory: %w", err)
		}
	}

	s.v.Store(next)
	return nil
}

func mergeJSONInto(dst any, patch json.RawMessage) error {
	baseBytes, err := json.Marshal(dst)
	if err != nil {
		return err
	}
	var base map[string]interface{}
	if err := json.Unmarshal(baseBytes, &base); err != nil {
		return err
	}
	var patchObj map[string]interface{}
	if err := json.Unmarshal(patch, &patchObj); err != nil {
		return err
	}
	mergeMapsShallowArrays(base, patchObj)
	merged, err := json.Marshal(base)
	if err != nil {
		return err
	}
	return json.Unmarshal(merged, dst)
}

func mergeMapsShallowArrays(base, patch map[string]interface{}) {
	for k, v := range patch {
		bm, bOk := base[k].(map[string]interface{})
		vm, vOk := v.(map[string]interface{})
		if bOk && vOk {
			mergeMapsShallowArrays(bm, vm)
			continue
		}
		base[k] = v
	}
}

// EffectiveLLMHTTP 应用 use_external 覆盖规则与默认超时，供单次 HTTP 调用使用。
func EffectiveLLMHTTP(cfg LLMConfig) LLMConfig {
	active := cfg
	if cfg.UseExternal {
		if strings.TrimSpace(cfg.External.BaseURL) != "" {
			active.BaseURL = cfg.External.BaseURL
		}
		if strings.TrimSpace(cfg.External.APIKey) != "" {
			active.APIKey = cfg.External.APIKey
		}
		if strings.TrimSpace(cfg.External.ModelName) != "" {
			active.ModelName = cfg.External.ModelName
		}
		if cfg.External.TimeoutMs > 0 {
			active.TimeoutMs = cfg.External.TimeoutMs
		}
	}
	if active.TimeoutMs <= 0 {
		active.TimeoutMs = 6000
	}
	return active
}

// RuntimeSectionsView 供 HTTP 返回「可被远程补丁修改」的配置分区。
type RuntimeSectionsView struct {
	Gateway           GatewayConfig           `json:"gateway"`
	OTA               OTAConfig               `json:"ota"`
	LLM               LLMConfig               `json:"llm"`
	TTS               TTSConfig               `json:"tts"`
	VAD               VADConfig               `json:"vad"`
	ASR               ASRConfig               `json:"asr"`
	Voiceprint        VoiceprintConfig        `json:"voiceprint"`
	ChatbotAdmin      ChatbotAdminConfig      `json:"chatbot_admin"`
	AudioPipeline     AudioPipelineConfig     `json:"audio_pipeline"`
	ShortTermMemory   ShortTermMemoryConfig   `json:"short_term_memory"`
}

// RuntimeSectionsFrom 从完整配置截取运行时区块。
func RuntimeSectionsFrom(cfg *AppConfig) RuntimeSectionsView {
	if cfg == nil {
		return RuntimeSectionsView{}
	}
	return RuntimeSectionsView{
		Gateway:         cfg.Gateway,
		OTA:             cfg.OTA,
		LLM:             cfg.LLM,
		TTS:             cfg.TTS,
		VAD:             cfg.VAD,
		ASR:             cfg.ASR,
		Voiceprint:      cfg.Voiceprint,
		ChatbotAdmin:    cfg.ChatbotAdmin,
		AudioPipeline:   cfg.AudioPipeline,
		ShortTermMemory: cfg.ShortTermMemory,
	}
}
