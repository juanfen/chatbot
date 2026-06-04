package service

import (
	"time"

	"chatbot-go-admin/server/internal/adapter"
	"chatbot-go-admin/server/internal/config"
	"go.uber.org/zap"
)

// NewProfileImportClientFromConfig builds the memory profile import HTTP client.
func NewProfileImportClientFromConfig(cfg config.Config, logger *zap.Logger) *adapter.ProfileImportClient {
	return adapter.NewProfileImportClient(adapter.ProfileImportConfig{
		BaseURL: cfg.MemoryProfileBaseURL,
		Timeout: time.Duration(cfg.MemoryProfileTimeoutSec) * time.Second,
	}, logger)
}
