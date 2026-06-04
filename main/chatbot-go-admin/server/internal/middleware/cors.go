package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// DevCORS enables browser cross-origin requests from local dev servers (e.g. Vite on :9527).
func DevCORS(extraOrigins []string) gin.HandlerFunc {
	allowed := make(map[string]struct{}, len(extraOrigins))
	for _, o := range extraOrigins {
		allowed[o] = struct{}{}
	}
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin != "" && isAllowedDevOrigin(origin, allowed) {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Vary", "Origin")
			c.Header("Access-Control-Allow-Credentials", "true")
			// Echo requested headers when browser asks for preflight (covers custom headers like x-request-id / apifoxtoken).
			if reqHeaders := strings.TrimSpace(c.GetHeader("Access-Control-Request-Headers")); reqHeaders != "" {
				c.Header("Access-Control-Allow-Headers", reqHeaders)
			} else {
				c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type, X-Requested-With, X-Request-Id, ApifoxToken")
			}
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			c.Header("Access-Control-Max-Age", "86400")
		}

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

func isAllowedDevOrigin(origin string, extra map[string]struct{}) bool {
	if _, ok := extra[origin]; ok {
		return true
	}
	// Allow typical local dev origins only (avoid opening wide for arbitrary domains).
	if strings.HasPrefix(origin, "http://localhost:") {
		return true
	}
	if strings.HasPrefix(origin, "http://127.0.0.1:") {
		return true
	}
	return false
}
