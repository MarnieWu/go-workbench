package httpapi

import (
	"errors"
	"go-workbench/internal/task"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type taskResponse struct {
	ID          string     `json:"id"`
	ProjectID   *string    `json:"projectId"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Status      string     `json:"status"`
	Priority    string     `json:"priority"`
	Labels      []string   `json:"labels"`
	DueAt       *time.Time `json:"dueAt"`
	ArchivedAt  *time.Time `json:"archivedAt"`
	Version     int64      `json:"version"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
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
				ID:          t.ID,
				ProjectID:   t.ProjectID,
				Title:       t.Title,
				Description: t.Description,
				Status:      string(t.Status),
				Priority:    string(t.Priority),
				Labels:      append(make([]string, 0, len(t.Labels)), t.Labels...),
				DueAt:       t.DueAt,
				ArchivedAt:  t.ArchivedAt,
				Version:     t.Version,
				CreatedAt:   t.CreatedAt,
				UpdatedAt:   t.UpdatedAt,
			})
		}

		c.JSON(http.StatusOK, listTasksResponse{Items: items})
	}
}

const (
	codeTaskNotFound        = "TASK_NOT_FOUND"
	messageTaskNotFound     = "task not found"
	codeTaskVersionConflict = "VERSION_CONFLICT"
	messageVersionConflict  = "task version changed; refresh and try again"
)

type updateTaskRequest struct {
	Version     int64         `json:"version"`
	ProjectID   *string       `json:"projectId"`
	Title       string        `json:"title"`
	Description string        `json:"description"`
	Status      task.Status   `json:"status"`
	Priority    task.Priority `json:"priority"`
	Labels      []string      `json:"labels"`
	DueAt       *time.Time    `json:"dueAt"`
}

func updateTask(service TaskUpdater) gin.HandlerFunc {
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
		var request updateTaskRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			writeError(c, http.StatusBadRequest, "INVALID_TASK_REQUEST", "invalid task request")
			return
		}
		updated, err := service.Update(c.Request.Context(), task.UpdateInput{
			OwnerID: ownerID, TaskID: c.Param("id"), Version: request.Version,
			ProjectID: request.ProjectID, Title: request.Title, Description: request.Description,
			Status: request.Status, Priority: request.Priority, Labels: append([]string(nil), request.Labels...),
			DueAt: request.DueAt, RequestID: c.GetString(requestIDKey),
		})
		if err != nil {
			switch {
			case errors.Is(err, task.ErrNotFound):
				writeError(c, http.StatusNotFound, codeTaskNotFound, messageTaskNotFound)
			case errors.Is(err, task.ErrVersionConflict):
				writeError(c, http.StatusConflict, codeTaskVersionConflict, messageVersionConflict)
			default:
				writeInternalError(c)
			}
			return
		}
		c.JSON(http.StatusOK, taskResponse{
			ID: updated.ID, ProjectID: updated.ProjectID, Title: updated.Title,
			Description: updated.Description, Status: string(updated.Status), Priority: string(updated.Priority),
			Labels: append([]string(nil), updated.Labels...), DueAt: updated.DueAt, ArchivedAt: updated.ArchivedAt,
			Version: updated.Version, CreatedAt: updated.CreatedAt, UpdatedAt: updated.UpdatedAt,
		})
	}
}
