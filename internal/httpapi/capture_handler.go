package httpapi

import (
	"errors"
	"go-workbench/internal/capture"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	codeInvalidCaptureRequest    = "INVALID_CAPTURE_REQUEST"
	messageInvalidCaptureRequest = "invalid capture request"
	codeIdempotencyConflict      = "IDEMPOTENCY_CONFLICT"
	messageIdempotencyConflict   = "idempotency key was already used with different input"
)

type createCaptureRequest struct {
	IdempotencyKey string  `json:"idempotencyKey"`
	InputText      string  `json:"inputText"`
	SourceType     string  `json:"sourceType"`
	ExternalRef    *string `json:"externalRef"`
	SourceURL      *string `json:"sourceUrl"`
	Excerpt        *string `json:"excerpt"`
	ConsentScope   *string `json:"consentScope"`
}

type captureResponse struct {
	ID        string    `json:"id"`
	Status    string    `json:"status"`
	RequestID string    `json:"requestId"`
	CreatedAt time.Time `json:"createdAt"`
}

func createCapture(service *capture.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		if service == nil {
			writeInternalError(c)
			return
		}

		ownerID := c.GetString(ownerIDKey)
		if ownerID == "" {
			writeError(c, http.StatusUnauthorized, codeUnauthorized, messageUnauthorized)
			return
		}

		var request createCaptureRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			writeError(c, http.StatusBadRequest, codeInvalidCaptureRequest, messageInvalidCaptureRequest)
			return
		}

		created, err := service.Create(c.Request.Context(), capture.CreateInput{
			OwnerID:        ownerID,
			IdempotencyKey: request.IdempotencyKey,
			InputText:      request.InputText,
			SourceType:     request.SourceType,
			ExternalRef:    request.ExternalRef,
			SourceURL:      request.SourceURL,
			Excerpt:        request.Excerpt,
			ConsentScope:   request.ConsentScope,
		})
		if err != nil {
			if errors.Is(err, capture.ErrIdempotencyConflict) {
				writeError(c, http.StatusConflict, codeIdempotencyConflict, messageIdempotencyConflict)
				return
			}
			writeInternalError(c)
			return
		}

		c.JSON(http.StatusCreated, captureResponse{
			ID:        created.ID,
			Status:    string(created.Status),
			RequestID: c.GetString(requestIDKey),
			CreatedAt: created.CreatedAt,
		})
	}
}
