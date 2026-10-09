package httpapi

import (
	"go-workbench/internal/task"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type taskResponse struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Status    string    `json:"status"`
	Priority  string    `json:"priority"`
	Labels    []string  `json:"labels"`
	Version   int64     `json:"version"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

const (
	codeInvalidStatus    = "INVALID_STATUS"
	messageInvalidStatus = "invalid status"
)

type listTasksResponse struct {
	Items []taskResponse `json:"items"`
}

const ownerIDKey = "ownerID"
const statusQueryKey = "status"

func listTasks(service *task.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		ownerID := c.GetString(ownerIDKey)

		if ownerID == "" {
			writeError(
				c,
				http.StatusUnauthorized,
				codeUnauthorized,
				messageUnauthorized,
			)
			return
		}

		status := c.Query(statusQueryKey)
		var filter task.ListTasksFilter

		switch status {
		case "":
			// no filter
		case "backlog", "in_progress", "blocked", "done":
			taskStatus := task.Status(status)
			filter.Status = &taskStatus
		default:
			writeError(
				c,
				http.StatusBadRequest,
				codeInvalidStatus,
				messageInvalidStatus,
			)
			return
		}

		tasks, err := service.List(c.Request.Context(), ownerID, filter)

		if err != nil {
			writeInternalError(c)
			return
		}

		items := make([]taskResponse, 0, len(tasks))
		for _, t := range tasks {
			items = append(items, taskResponse{
				ID:        t.ID,
				Title:     t.Title,
				Status:    string(t.Status),
				Priority:  string(t.Priority),
				Labels:    append(make([]string, 0, len(t.Labels)), t.Labels...),
				Version:   t.Version,
				CreatedAt: t.CreatedAt,
				UpdatedAt: t.UpdatedAt,
			})
		}

		c.JSON(http.StatusOK, listTasksResponse{Items: items})
	}
}
