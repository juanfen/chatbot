package httpserver

import (
	"fmt"
	"net"
	"sort"
	"strings"

	"chatbot-go-admin/server/internal/config"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// LogStartup logs runtime configuration and all registered HTTP routes.
func LogStartup(logger *zap.Logger, engine *gin.Engine, cfg config.Config) {
	audioDir, err := cfg.ResolveAudioDir()
	if err != nil {
		audioDir = cfg.AudioTargetPath
	}

	corsOrigins := "none"
	if len(cfg.CORSOrigins) > 0 {
		corsOrigins = strings.Join(cfg.CORSOrigins, ", ")
	}

	logger.Info("application configuration",
		zap.String("http_addr", cfg.HTTPAddr),
		zap.String("log_level", cfg.LogLevel),
		zap.String("schema_dir", cfg.SchemaDir),
		zap.String("db_host", cfg.DB.Host),
		zap.Int("db_port", cfg.DB.Port),
		zap.String("db_name", cfg.DB.DBName),
		zap.String("db_sslmode", cfg.DB.SSLMode),
		zap.String("xiaozhi_core_base_url", cfg.XiaoZhiCoreBaseURL),
		zap.String("audio_dir", audioDir),
		zap.String("ffmpeg_bin", cfg.FFmpegBin),
		zap.String("ffprobe_bin", cfg.FFprobeBin),
		zap.Bool("audio_denoise", cfg.AudioDenoise),
		zap.String("cors_origins", corsOrigins),
		zap.String("gin_mode", gin.Mode()),
	)

	routes := sortedRoutes(engine.Routes())
	logger.Info("registered http routes", zap.Int("total", len(routes)))
	for _, route := range routes {
		logger.Info("http route",
			zap.String("method", route.Method),
			zap.String("path", route.Path),
		)
	}

	baseURL := listenBaseURL(cfg.HTTPAddr)
	logger.Info("http server ready",
		zap.String("addr", cfg.HTTPAddr),
		zap.String("health", baseURL+"/health"),
		zap.String("health_v1", baseURL+"/api/v1/health"),
	)
}

func sortedRoutes(routes gin.RoutesInfo) gin.RoutesInfo {
	out := append(gin.RoutesInfo(nil), routes...)
	sort.Slice(out, func(i, j int) bool {
		if out[i].Path == out[j].Path {
			return out[i].Method < out[j].Method
		}
		return out[i].Path < out[j].Path
	})
	return out
}

func listenBaseURL(addr string) string {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		if strings.HasPrefix(addr, ":") {
			return fmt.Sprintf("http://127.0.0.1%s", addr)
		}
		return "http://" + addr
	}
	if host == "" || host == "0.0.0.0" || host == "[::]" {
		host = "127.0.0.1"
	}
	return fmt.Sprintf("http://%s:%s", host, port)
}
