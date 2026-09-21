package httpapi

import (
	"context"
	"encoding/json"
	"go-workbench/internal/task"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

type stubTaskRepository struct {
	tasks      []task.Task
	err        error
	callCount  int
	gotOwnerID string
}

func (r *stubTaskRepository) List(_ context.Context, ownerID string) ([]task.Task, error) {
	r.callCount++
	r.gotOwnerID = ownerID
	return r.tasks, r.err
}

func testOwner(ownerID string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(ownerIDKey, ownerID)
		c.Next()
	}
}

func TestRouterListTasks(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repository := &stubTaskRepository{
		tasks: []task.Task{{
			ID:      "task-1",
			OwnerID: "owner-1",
			Title:   "Learn Go service boundaries",
			Status:  task.StatusBacklog,
		}},
	}
	service := task.NewService(repository)
	router := NewRouter(service, testOwner("owner-1"))
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/v1/tasks", nil)

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
}

func TestRouterListTasksReturnsEmptyItems(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repository := &stubTaskRepository{tasks: nil}
	service := task.NewService(repository)
	router := NewRouter(service, testOwner("owner-1"))
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/v1/tasks", nil)

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}

	var response struct {
		Items []json.RawMessage `json:"items"`
	}
	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if response.Items == nil {
		t.Fatal("items = null or missing, want []")
	}
	if len(response.Items) != 0 {
		t.Fatalf("len(items) = %d, want 0", len(response.Items))
	}
	if repository.callCount != 1 {
		t.Fatalf("repository.List() call count = %d, want 1", repository.callCount)
	}
	if repository.gotOwnerID != "owner-1" {
		t.Fatalf("repository.List() got ownerID = %q, want %q", repository.gotOwnerID, "owner-1")
	}
}
