package common

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	CodeOK           = 0
	CodeBadRequest   = 40000
	CodeUnauthorized = 40100
	CodeForbidden    = 40300
	CodeNotFound     = 40400
	CodeInternal     = 50000
)

type Response[T any] struct {
	// Code is the business status code.
	//
	// - 0: success
	// - 4xxxx: client-side / request / auth errors (see constants below)
	// - 5xxxx: server internal errors
	Code    int    `json:"code"`
	// Message is the human-readable message for UI display.
	// IMPORTANT: the JSON field name is "message" (NOT "msg").
	Message string `json:"message"`
	// Data is the business payload. It can be null on both success and failure.
	Data    *T     `json:"data"`
}

func OK[T any](c *gin.Context, data *T) {
	// NOTE: business-level success is always expressed by CodeOK=0,
	// while HTTP status remains 200.
	c.JSON(http.StatusOK, Response[T]{Code: CodeOK, Message: "ok", Data: data})
}

func Fail(c *gin.Context, httpStatus int, code int, message string) {
	// NOTE: Fail MUST still follow the same envelope shape.
	// The httpStatus should reflect protocol/auth/resource semantics (401/403/404/405/500...),
	// while the business error is carried by "code" and "message".
	c.JSON(httpStatus, Response[any]{Code: code, Message: message, Data: nil})
}

