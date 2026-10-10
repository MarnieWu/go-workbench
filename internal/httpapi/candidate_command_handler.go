package httpapi

import (
	"errors"
	"go-workbench/internal/candidate"
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	codeCandidateNotFound      = "CANDIDATE_NOT_FOUND"
	messageCandidateNotFound   = "candidate not found"
	codeCandidateStateConflict = "CANDIDATE_STATE_CONFLICT"
	messageCandidateConflict   = "candidate state changed; refresh and try again"
)

func rejectCandidate(service CandidateRejecter) gin.HandlerFunc {
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

		item, err := service.Reject(c.Request.Context(), candidate.RejectInput{
			OwnerID:     ownerID,
			CandidateID: c.Param("id"),
			RequestID:   c.GetString(requestIDKey),
		})
		if err != nil {
			writeCandidateCommandError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"id": item.ID, "status": item.Status})
	}
}

type acceptCandidateRequest struct {
	Title       string   `json:"title"`
	Description *string  `json:"description"`
	ProjectID   *string  `json:"projectId"`
	Labels      []string `json:"labels"`
}

func acceptCandidate(service CandidateAccepter) gin.HandlerFunc {
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

		var request acceptCandidateRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			writeError(c, http.StatusBadRequest, codeInvalidCaptureRequest, "invalid candidate request")
			return
		}

		result, err := service.Accept(c.Request.Context(), candidate.AcceptInput{
			OwnerID:     ownerID,
			CandidateID: c.Param("id"),
			Title:       request.Title,
			Description: request.Description,
			ProjectID:   request.ProjectID,
			Labels:      append([]string(nil), request.Labels...),
			RequestID:   c.GetString(requestIDKey),
		})
		if err != nil {
			writeCandidateCommandError(c, err)
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"candidateId": result.CandidateID,
			"taskId":      result.TaskID,
			"status":      candidate.StatusAccepted,
		})
	}
}

func writeCandidateCommandError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, candidate.ErrNotFound):
		writeError(c, http.StatusNotFound, codeCandidateNotFound, messageCandidateNotFound)
	case errors.Is(err, candidate.ErrStateConflict):
		writeError(c, http.StatusConflict, codeCandidateStateConflict, messageCandidateConflict)
	default:
		writeInternalError(c)
	}
}
