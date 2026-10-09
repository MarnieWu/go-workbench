package httpapi

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	requestIDLoggerKey      = "request_id"
	requestIDLoggerMessage  = "request completed"
	codeInternalError       = "INTERNAL_ERROR"
	messageInternalError    = "internal server error"
	codeUnauthorized        = "UNAUTHORIZED"
	messageUnauthorized     = "authentication is required"
	codeInvalidRequestID    = "INVALID_REQUEST_ID"
	messageInvalidRequestID = "invalid request ID"
)

type errorResponse struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"requestId"`
}

func writeError(
	c *gin.Context,
	status int,
	code string,
	message string,
) {
	c.AbortWithStatusJSON(status, errorResponse{
		Code:      code,
		Message:   message,
		RequestID: c.GetString(requestIDKey),
	})
}

func writeInternalError(c *gin.Context) {
	writeError(c, http.StatusInternalServerError, codeInternalError, messageInternalError)
}
