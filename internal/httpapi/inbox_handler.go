package httpapi

import (
	"go-workbench/internal/candidate"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type candidateResponse struct {
	ID                  string    `json:"id"`
	CaptureID           string    `json:"captureId"`
	ProposedTitle       string    `json:"proposedTitle"`
	ProposedDescription *string   `json:"proposedDescription"`
	ProposedProjectID   *string   `json:"proposedProjectId"`
	Labels              []string  `json:"labels"`
	Status              string    `json:"status"`
	CreatedAt           time.Time `json:"createdAt"`
	UpdatedAt           time.Time `json:"updatedAt"`
}

type listInboxResponse struct {
	Items []candidateResponse `json:"items"`
}

func listInbox(service *candidate.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		ownerID := c.GetString(ownerIDKey)
		if ownerID == "" {
			writeError(c, http.StatusUnauthorized, codeUnauthorized, messageUnauthorized)
			return
		}
		if service == nil {
			writeInternalError(c)
			return
		}

		candidates, err := service.ListPending(c.Request.Context(), ownerID)
		if err != nil {
			writeInternalError(c)
			return
		}

		items := make([]candidateResponse, 0, len(candidates))
		for _, item := range candidates {
			items = append(items, candidateResponse{
				ID:                  item.ID,
				CaptureID:           item.CaptureID,
				ProposedTitle:       item.ProposedTitle,
				ProposedDescription: item.ProposedDescription,
				ProposedProjectID:   item.ProposedProjectID,
				Labels:              append(make([]string, 0, len(item.Labels)), item.Labels...),
				Status:              string(item.Status),
				CreatedAt:           item.CreatedAt,
				UpdatedAt:           item.UpdatedAt,
			})
		}

		c.JSON(http.StatusOK, listInboxResponse{Items: items})
	}
}
