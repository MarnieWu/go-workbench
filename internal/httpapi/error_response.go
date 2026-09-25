package httpapi

import (
	"net/http"

	"github.com/gin-gonic/gin"
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
