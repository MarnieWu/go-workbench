package httpapi

import (
	"go-workbench/internal/task"
	"net/http"

	"github.com/gin-gonic/gin"
)

type taskResponse struct {
	ID        string   `json:"id"`
	Title     string   `json:"title"`
	Status    string   `json:"status"`
	Priority  string   `json:"priority"`
	Labels    []string `json:"labels"`
	Version   int64    `json:"version"`
	CreatedAt string   `json:"createdAt"`
	UpdatedAt string   `json:"updatedAt"`
}

type listTasksResponse struct {
	Items []taskResponse `json:"items"`
}

const ownerIDKey = "ownerID"

func listTasks(service *task.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		ownerID := c.GetString(ownerIDKey)
		tasks, err := service.List(c.Request.Context(), ownerID)

		if err != nil {
			c.Status(http.StatusInternalServerError)
			return
		}

		items := make([]taskResponse, 0, len(tasks))

		c.JSON(http.StatusOK, listTasksResponse{Items: items})
	}
}
