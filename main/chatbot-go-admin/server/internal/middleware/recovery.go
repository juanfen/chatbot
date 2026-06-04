package middleware

import (
	"net/http"

	"chatbot-go-admin/server/internal/common"
	"chatbot-go-admin/server/internal/logging"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Recovery recovers panics and returns unified error responses.
func Recovery(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				logger.Error("panic recovered",
					zap.Any("recover", r),
					logging.String("path", c.FullPath()),
				)
				common.Fail(c, http.StatusInternalServerError, common.CodeInternal, "internal error")
				c.Abort()
			}
		}()
		c.Next()
	}
}

