package httpapi

import (
	"context"
	"crypto/rand"
	"go-workbench/internal/candidate"
	"go-workbench/internal/capture"
	"go-workbench/internal/task"
	"log/slog"
	"os"

	"github.com/gin-gonic/gin"
)

type CandidateRejecter interface {
	Reject(ctx context.Context, input candidate.RejectInput) (candidate.Candidate, error)
}

type CandidateAccepter interface {
	Accept(ctx context.Context, input candidate.AcceptInput) (candidate.AcceptResult, error)
}

type TaskUpdater interface {
	Update(ctx context.Context, input task.UpdateInput) (task.Task, error)
}

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
	requestIDHeader = "X-Request-ID"
	requestIDKey    = "requestID"
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

func LocalOwnerMiddleware(ownerID string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(ownerIDKey, ownerID)
		c.Next()
	}
}

type RouterConfig struct {
	TaskService       *task.Service
	CaptureService    *capture.Service
	CandidateService  *candidate.Service
	CandidateRejecter CandidateRejecter
	CandidateAccepter CandidateAccepter
	TaskUpdater       TaskUpdater
	Logger            *slog.Logger
}

func NewRouter(
	config RouterConfig,
	middlewares ...gin.HandlerFunc,
) *gin.Engine {
	router := gin.New()
	logger := config.Logger
	if logger == nil {
		logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))
	}

	router.Use(requestIDMiddleware())
	router.Use(requestLoggingMiddleware(logger))
	router.Use(recoveryMiddleware())
	router.Use(middlewares...)
	router.GET("/v1/tasks", listTasks(config.TaskService))
	router.POST("/v1/captures", createCapture(config.CaptureService))
	router.GET("/v1/inbox", listInbox(config.CandidateService))
	router.POST("/v1/candidates/:id/reject", rejectCandidate(config.CandidateRejecter))
	router.POST("/v1/candidates/:id/accept", acceptCandidate(config.CandidateAccepter))
	router.PATCH("/v1/tasks/:id", updateTask(config.TaskUpdater))

	return router
}
