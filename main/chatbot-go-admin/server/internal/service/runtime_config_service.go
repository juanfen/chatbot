package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// RuntimeConfigService 反向代理到 xiaozhi-go-core 的运行时配置接口。
//
// 设计说明：
// - admin 后端**不持久化**该配置（与 RUNTIME_CONFIG_API.md 对齐：仅修改进程内存）。
// - 通过环境变量 XIAOZHI_CORE_BASE_URL 指定上游基础地址。
// - GET 透传上游 JSON；POST 把补丁体透传给上游 `/xiaozhi/v1/runtime/config`。
type RuntimeConfigService struct {
	baseURL string
	client  *http.Client
}

func NewRuntimeConfigService(baseURL string) *RuntimeConfigService {
	if baseURL == "" {
		baseURL = "http://127.0.0.1:8003"
	}
	return &RuntimeConfigService{
		baseURL: strings.TrimRight(baseURL, "/"),
		client:  &http.Client{Timeout: 10 * time.Second},
	}
}

// AllowedKeys 与 RUNTIME_CONFIG_API.md 中允许补丁的顶层键保持一致。
var AllowedKeys = []string{"gateway", "ota", "llm", "tts", "vad", "asr", "voiceprint", "chatbot_admin", "audio_pipeline"}

func (s *RuntimeConfigService) endpoint() string {
	return s.baseURL + "/xiaozhi/v1/runtime/config"
}

// Get 获取上游当前生效配置（仅 8 个允许键）。
func (s *RuntimeConfigService) Get(ctx context.Context) (map[string]any, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.endpoint(), nil)
	if err != nil {
		return nil, err
	}
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("call upstream failed: %w", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode/100 != 2 {
		return nil, fmt.Errorf("upstream %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	var out map[string]any
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, fmt.Errorf("parse upstream response failed: %w", err)
	}
	return out, nil
}

// Patch 把补丁透传给上游进行深度合并。
//
// 仅校验顶层 key 是否在允许列表，避免将无关键值发出去；
// 详细的字段语义校验由上游负责。
func (s *RuntimeConfigService) Patch(ctx context.Context, patch map[string]any) error {
	if len(patch) == 0 {
		return errors.New("empty patch")
	}
	allowed := map[string]struct{}{}
	for _, k := range AllowedKeys {
		allowed[k] = struct{}{}
	}
	for k := range patch {
		if _, ok := allowed[k]; !ok {
			return fmt.Errorf("unknown or disallowed key %q", k)
		}
	}
	body, err := json.Marshal(patch)
	if err != nil {
		return fmt.Errorf("marshal patch failed: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.endpoint(), bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("call upstream failed: %w", err)
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode/100 != 2 {
		return fmt.Errorf("upstream %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}
	return nil
}
