package httpserver

import (
	"chatbot-go-admin/server/internal/config"
	"chatbot-go-admin/server/internal/handler"
	"chatbot-go-admin/server/internal/middleware"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// NewEngine creates the gin engine with standard middlewares and routes.
func NewEngine(logger *zap.Logger, gormDB *gorm.DB, cfg config.Config) *gin.Engine {
	if cfg.LogLevel == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(middleware.RequestID())
	r.Use(middleware.MethodLimit())
	r.Use(middleware.Recovery(logger))
	r.Use(middleware.DevCORS(cfg.CORSOrigins))

	h := handler.NewHandlers(logger, gormDB, cfg)
	h.Register(r)

	return r
}

