package httpapi

import (
	"crypto/rand"
	"go-workbench/internal/task"
	"log/slog"
	"os"

	"github.com/gin-gonic/gin"
)

func recoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if recovered := recover(); recovered != nil {
				writeInternalError(c)
			}
		}()

		c.Next()
	}
}

const (
	requestIDHeader         = "X-Request-ID"
	requestIDKey            = "requestID"
	requestIDLoggerKey      = "request_id"
	requestIDLoggerMessage  = "request completed"
	codeInternalError       = "INTERNAL_ERROR"
	messageInternalError    = "internal server error"
	codeInvalidStatus       = "INVALID_STATUS"
	messageInvalidStatus    = "invalid status"
	codeUnauthorized        = "UNAUTHORIZED"
	messageUnauthorized     = "authentication is required"
	codeInvalidRequestID    = "INVALID_REQUEST_ID"
	messageInvalidRequestID = "invalid request ID"
)

func validRequestID(requestID string) bool {
	if len(requestID) < 1 || len(requestID) > 64 {
		return false
	}

	for _, r := range requestID {
		if !((r >= 'a' && r <= 'z') ||
			(r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') ||
			r == '-' || r == '_' || r == '.') {
			return false
		}
	}

	return true
}

func requestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := rand.Text()

		c.Set(requestIDKey, requestID)
		c.Header(requestIDHeader, requestID)
		c.Next()
	}
}

func requestLoggingMiddleware(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		logger.InfoContext(
			c.Request.Context(),
			requestIDLoggerMessage,
			slog.String(requestIDLoggerKey, c.GetString(requestIDKey)),
		)
	}
}

func NewRouter(service *task.Service, middlewares ...gin.HandlerFunc) *gin.Engine {
	router := gin.New()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	router.Use(requestIDMiddleware())
	router.Use(requestLoggingMiddleware(logger))
	router.Use(recoveryMiddleware())
	router.Use(middlewares...)
	router.GET("/v1/tasks", listTasks(service))
	return router
}
