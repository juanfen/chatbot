package adapter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"
)

// ProfileImportConfig configures the long-term memory profile import endpoint.
type ProfileImportConfig struct {
	BaseURL string
	Timeout time.Duration
}

// ProfileImportClient calls POST /profile/import_basic_info on the memory service.
type ProfileImportClient struct {
	cfg    ProfileImportConfig
	logger *zap.Logger
}

func NewProfileImportClient(cfg ProfileImportConfig, logger *zap.Logger) *ProfileImportClient {
	if cfg.Timeout <= 0 {
		cfg.Timeout = 30 * time.Second
	}
	if logger == nil {
		logger = zap.NewNop()
	}
	return &ProfileImportClient{cfg: cfg, logger: logger}
}

// Enabled reports whether profile import is configured.
func (c *ProfileImportClient) Enabled() bool {
	return strings.TrimSpace(c.cfg.BaseURL) != ""
}

type profileImportBasicInfoReq struct {
	MacID             string         `json:"mac_id"`
	UserID            string         `json:"user_id"`
	UserBasicInfoList []string       `json:"user_basic_info_list"`
	Source            string         `json:"source,omitempty"`
	UpdateProfile     *bool          `json:"update_profile,omitempty"`
	InferMemories     *bool          `json:"infer_memories,omitempty"`
	Metadata          map[string]any `json:"metadata,omitempty"`
}

// ProfileImportResult is a subset of the memory service response.
type ProfileImportResult struct {
	ProfileFactsUpserted int `json:"profile_facts_upserted"`
	MemoriesWritten      int `json:"memories_written"`
}

type profileImportBasicInfoResp struct {
	Result struct {
		ProfileFactsUpserted int `json:"profile_facts_upserted"`
		MemoriesWritten      int `json:"memories_written"`
	} `json:"result"`
}

// ImportBasicInfo writes user basic info lines into the memory service for one device scope.
func (c *ProfileImportClient) ImportBasicInfo(
	ctx context.Context,
	macID, userID string,
	lines []string,
	metadata map[string]any,
) (*ProfileImportResult, error) {
	if !c.Enabled() {
		return nil, fmt.Errorf("profile import not configured")
	}
	macID = strings.TrimSpace(macID)
	userID = strings.TrimSpace(userID)
	if macID == "" {
		return nil, fmt.Errorf("mac_id is empty")
	}
	if userID == "" {
		return nil, fmt.Errorf("user_id is empty")
	}
	if len(lines) == 0 {
		return nil, fmt.Errorf("user_basic_info_list is empty")
	}

	base := strings.TrimRight(strings.TrimSpace(c.cfg.BaseURL), "/")
	url := base + "/profile/import_basic_info"

	updateProfile := true
	inferMemories := true
	body := profileImportBasicInfoReq{
		MacID:             macID,
		UserID:            userID,
		UserBasicInfoList: lines,
		Source:            "chatbot_speaker_personal_lines",
		UpdateProfile:     &updateProfile,
		InferMemories:     &inferMemories,
		Metadata:          metadata,
	}
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	requestBodyLog := formatJSONForLog(bodyBytes)

	start := time.Now()
	c.logger.Info("【画像导入】准备调用 import_basic_info 接口",
		zap.String("请求方法", http.MethodPost),
		zap.String("请求地址", url),
		zap.Duration("超时时间", c.cfg.Timeout),
		zap.String("请求体完整内容", requestBodyLog),
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	c.logger.Info("【画像导入】正在发送 HTTP 请求",
		zap.String("请求地址", url),
		zap.String("设备MAC", macID),
		zap.String("用户ID", userID),
		zap.Int("个性化条目数", len(lines)),
	)

	client := &http.Client{Timeout: c.cfg.Timeout}
	resp, err := client.Do(req)
	elapsed := time.Since(start)
	if err != nil {
		c.logger.Warn("【画像导入】HTTP 请求失败",
			zap.String("请求地址", url),
			zap.String("设备MAC", macID),
			zap.String("用户ID", userID),
			zap.String("请求体完整内容", requestBodyLog),
			zap.Duration("耗时", elapsed),
			zap.Error(err),
		)
		return nil, fmt.Errorf("profile import request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	responseBodyLog := formatJSONForLog(respBody)
	c.logger.Info("【画像导入】收到 HTTP 响应",
		zap.String("请求地址", url),
		zap.String("设备MAC", macID),
		zap.String("用户ID", userID),
		zap.Int("HTTP状态码", resp.StatusCode),
		zap.Duration("耗时", elapsed),
		zap.Int("响应体字节数", len(respBody)),
		zap.String("响应体完整内容", responseBodyLog),
	)

	if resp.StatusCode != http.StatusOK {
		c.logger.Warn("【画像导入】接口返回非 200",
			zap.String("请求地址", url),
			zap.String("设备MAC", macID),
			zap.String("用户ID", userID),
			zap.Int("HTTP状态码", resp.StatusCode),
			zap.String("请求体完整内容", requestBodyLog),
			zap.String("响应体完整内容", responseBodyLog),
		)
		return nil, fmt.Errorf("profile import status %d: %s", resp.StatusCode, string(respBody))
	}

	var out profileImportBasicInfoResp
	if len(respBody) > 0 {
		if err := json.Unmarshal(respBody, &out); err != nil {
			c.logger.Warn("【画像导入】响应 JSON 解析失败",
				zap.String("设备MAC", macID),
				zap.String("用户ID", userID),
				zap.String("响应体完整内容", responseBodyLog),
				zap.Error(err),
			)
			return nil, fmt.Errorf("decode profile import response: %w", err)
		}
	}
	result := &ProfileImportResult{
		ProfileFactsUpserted: out.Result.ProfileFactsUpserted,
		MemoriesWritten:      out.Result.MemoriesWritten,
	}
	c.logger.Info("【画像导入】调用成功",
		zap.String("设备MAC", macID),
		zap.String("用户ID", userID),
		zap.Int("写入画像事实数", result.ProfileFactsUpserted),
		zap.Int("写入长期记忆数", result.MemoriesWritten),
		zap.Duration("耗时", elapsed),
		zap.String("请求体完整内容", requestBodyLog),
		zap.String("响应体完整内容", responseBodyLog),
	)
	return result, nil
}

func formatJSONForLog(raw []byte) string {
	if len(raw) == 0 {
		return "(空)"
	}
	var pretty bytes.Buffer
	if err := json.Indent(&pretty, raw, "", "  "); err != nil {
		return string(raw)
	}
	return pretty.String()
}
