package service

import (
	"time"

	"chatbot-go-admin/server/internal/adapter"
	"chatbot-go-admin/server/internal/config"
	"go.uber.org/zap"
)

// NewVoiceprintClientFromConfig builds the SV voiceprint HTTP client from app config.
func NewVoiceprintClientFromConfig(cfg config.Config, logger *zap.Logger) *adapter.VoiceprintClient {
	return adapter.NewVoiceprintClient(adapter.VoiceprintConfig{
		RegisterURL: cfg.ResolveSVVoiceprintRegisterURL(),
		Token:       cfg.SVVoiceprintToken,
		Timeout:     time.Duration(cfg.SVVoiceprintTimeoutSec) * time.Second,
	}, logger)
}
