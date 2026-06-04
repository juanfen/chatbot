package middleware

import (
	"net/http"
	"strings"

	"chatbot-go-admin/server/internal/common"
	"chatbot-go-admin/server/internal/util"
	"github.com/gin-gonic/gin"
)

const (
	CtxKeyUserID  = "user_id"
	CtxKeyIsAdmin = "is_admin"
)

// JWTAuth validates Authorization: Bearer <token>.
func JWTAuth(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if auth == "" || !strings.HasPrefix(auth, "Bearer ") {
			common.Fail(c, http.StatusUnauthorized, common.CodeUnauthorized, "unauthorized")
			c.Abort()
			return
		}
		token := strings.TrimSpace(strings.TrimPrefix(auth, "Bearer "))
		claims, err := util.ParseToken(secret, token)
		if err != nil {
			common.Fail(c, http.StatusUnauthorized, common.CodeUnauthorized, "unauthorized")
			c.Abort()
			return
		}
		c.Set(CtxKeyUserID, claims.UserID)
		c.Set(CtxKeyIsAdmin, claims.IsAdmin)
		c.Next()
	}
}

