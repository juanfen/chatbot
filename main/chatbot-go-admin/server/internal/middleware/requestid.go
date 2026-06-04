package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const CtxKeyRequestID = "request_id"

// RequestID sets request id to context and response headers.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		rid := c.GetHeader("X-Request-Id")
		if rid == "" {
			rid = uuid.NewString()
		}
		c.Set(CtxKeyRequestID, rid)
		c.Header("X-Request-Id", rid)
		c.Next()
	}
}

