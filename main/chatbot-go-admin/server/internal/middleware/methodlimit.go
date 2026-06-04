package middleware

import (
	"net/http"

	"chatbot-go-admin/server/internal/common"
	"github.com/gin-gonic/gin"
)

// MethodLimit blocks non-GET/POST methods.
func MethodLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		switch c.Request.Method {
		case http.MethodGet, http.MethodPost, http.MethodOptions:
			c.Next()
			return
		default:
			common.Fail(c, http.StatusMethodNotAllowed, common.CodeBadRequest, "method not allowed")
			c.Abort()
			return
		}
	}
}

