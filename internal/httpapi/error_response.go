package httpapi

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	codeInternalError       = "INTERNAL_ERROR"
	messageInternalError    = "internal server error"
	codeInvalidStatus       = "INVALID_STATUS"
	messageInvalidStatus    = "invalid status"
	codeUnauthorized        = "UNAUTHORIZED"
	messageUnauthorized     = "authentication is required"
	codeInvalidRequestID    = "INVALID_REQUEST_ID"
	messageInvalidRequestID = "invalid request ID"
)

type errorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func writeError(
	c *gin.Context,
	status int,
	code string,
	message string,
) {
	c.AbortWithStatusJSON(status, errorResponse{
		Code:    code,
		Message: message,
	})
}

func writeInternalError(c *gin.Context) {
	writeError(c, http.StatusInternalServerError, codeInternalError, messageInternalError)
}
