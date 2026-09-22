package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"go-workbench/internal/task"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

type stubTaskRepository struct {
	tasks       []task.Task
	err         error
	panicOnList bool
	callCount   int
	gotOwnerID  string
}

func (r *stubTaskRepository) List(_ context.Context, ownerID string) ([]task.Task, error) {
	r.callCount++
	r.gotOwnerID = ownerID

	if r.panicOnList {
		panic("repository panic")
	}

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
		t.Fatalf("repository calls = %d, want 1", repository.callCount)
	}
	if repository.gotOwnerID != "owner-1" {
		t.Fatalf("repository got ownerID = %q, want %q", repository.gotOwnerID, "owner-1")
	}
}

func TestRouterListTasksErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name            string
		requestURL      string
		ownerID         string
		repositoryError error
		repositoryPanic bool
		wantStatus      int
		wantCode        string
		wantMessage     string
		wantCalls       int
	}{
		{
			name:        "invalid status",
			requestURL:  "/v1/tasks?status=unknown",
			ownerID:     "owner-1",
			wantStatus:  http.StatusBadRequest,
			wantCode:    codeInvalidStatus,
			wantMessage: messageInvalidStatus,
			wantCalls:   0,
		},
		{
			name:        "missing owner",
			requestURL:  "/v1/tasks",
			wantStatus:  http.StatusUnauthorized,
			wantCode:    codeUnauthorized,
			wantMessage: messageUnauthorized,
			wantCalls:   0,
		},
		{
			name:            "repository error",
			requestURL:      "/v1/tasks",
			ownerID:         "owner-1",
			repositoryError: errors.New("database error"),
			wantStatus:      http.StatusInternalServerError,
			wantCode:        codeInternalError,
			wantMessage:     messageInternalError,
			wantCalls:       1,
		},
		{
			name:            "repository panic",
			requestURL:      "/v1/tasks",
			ownerID:         "owner-1",
			repositoryPanic: true,
			wantStatus:      http.StatusInternalServerError,
			wantCode:        codeInternalError,
			wantMessage:     messageInternalError,
			wantCalls:       1,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repository := &stubTaskRepository{
				err:         test.repositoryError,
				panicOnList: test.repositoryPanic,
			}
			service := task.NewService(repository)

			var middlewares []gin.HandlerFunc
			if test.ownerID != "" {
				middlewares = append(middlewares, testOwner(test.ownerID))
			}

			router := NewRouter(service, middlewares...)
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, test.requestURL, nil)

			router.ServeHTTP(recorder, request)

			if recorder.Code != test.wantStatus {
				t.Fatalf(
					"status = %d, want %d; body = %s",
					recorder.Code,
					test.wantStatus,
					recorder.Body.String(),
				)
			}

			var response struct {
				Code    string `json:"code"`
				Message string `json:"message"`
			}

			if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
				t.Fatalf(
					"failed to decode response: %v; body = %s",
					err,
					recorder.Body.String(),
				)
			}

			if response.Code != test.wantCode {
				t.Errorf(
					"code = %q, want %q",
					response.Code,
					test.wantCode,
				)
			}

			if response.Message != test.wantMessage {
				t.Errorf(
					"message = %q, want %q",
					response.Message,
					test.wantMessage,
				)
			}

			if repository.callCount != test.wantCalls {
				t.Errorf(
					"repository calls = %d, want %d",
					repository.callCount,
					test.wantCalls,
				)
			}
		})
	}
}
