package gateway

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type AuthManager struct {
	secret        string
	expireSeconds int64
}

func NewAuthManager(secret string, expireSeconds int64) *AuthManager {
	if expireSeconds <= 0 {
		expireSeconds = 60 * 60 * 24 * 30
	}
	return &AuthManager{secret: secret, expireSeconds: expireSeconds}
}

func (a *AuthManager) GenerateToken(clientID string, username string) string {
	ts := time.Now().Unix()
	content := fmt.Sprintf("%s|%s|%d", clientID, username, ts)
	return a.sign(content) + "." + strconv.FormatInt(ts, 10)
}

func (a *AuthManager) VerifyToken(token string, clientID string, username string) bool {
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return false
	}
	ts, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return false
	}
	if time.Now().Unix()-ts > a.expireSeconds {
		return false
	}
	expected := a.sign(fmt.Sprintf("%s|%s|%d", clientID, username, ts))
	return hmac.Equal([]byte(parts[0]), []byte(expected))
}

func (a *AuthManager) sign(content string) string {
	mac := hmac.New(sha256.New, []byte(a.secret))
	_, _ = mac.Write([]byte(content))
	return strings.TrimRight(base64.URLEncoding.EncodeToString(mac.Sum(nil)), "=")
}

func BearerToken(raw string) string {
	raw = strings.TrimSpace(raw)
	if strings.HasPrefix(strings.ToLower(raw), "bearer ") {
		return strings.TrimSpace(raw[7:])
	}
	return raw
}
