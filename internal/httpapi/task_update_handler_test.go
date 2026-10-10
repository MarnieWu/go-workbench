package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"go-workbench/internal/task"
	"net/http"
	"net/http/httptest"
	"testing"
)

type stubTaskUpdater struct {
	input  task.UpdateInput
	result task.Task
	err    error
}

func (s *stubTaskUpdater) Update(_ context.Context, input task.UpdateInput) (task.Task, error) {
	s.input = input
	return s.result, s.err
}

func TestRouterUpdateTaskConnectsVersionedInput(t *testing.T) {
	updater := &stubTaskUpdater{result: task.Task{
		ID: "task-1", Title: "Updated", Status: task.StatusInProgress,
		Priority: task.PriorityHigh, Labels: []string{"go"}, Version: 3,
	}}
	router := NewRouter(RouterConfig{TaskUpdater: updater}, testOwner("owner-1"))
	body := bytes.NewBufferString(`{
		"version":2,"title":"Updated","description":"Keep input",
		"status":"in_progress","priority":"high","labels":["go"]
	}`)
	request := httptest.NewRequest(http.MethodPatch, "/v1/tasks/task-1", body)
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body = %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	if updater.input.OwnerID != "owner-1" || updater.input.TaskID != "task-1" || updater.input.Version != 2 {
		t.Fatalf("input = %#v", updater.input)
	}
	if updater.input.RequestID == "" {
		t.Fatal("request ID is empty")
	}
}

func TestRouterUpdateTaskMapsVersionConflict(t *testing.T) {
	updater := &stubTaskUpdater{err: task.ErrVersionConflict}
	router := NewRouter(RouterConfig{TaskUpdater: updater}, testOwner("owner-1"))
	body := bytes.NewBufferString(`{"version":1,"title":"Local edit","description":"","status":"backlog","priority":"none","labels":[]}`)
	request := httptest.NewRequest(http.MethodPatch, "/v1/tasks/task-1", body)
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusConflict)
	}
	var response errorResponse
	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.Code != codeTaskVersionConflict {
		t.Fatalf("code = %q, want %q", response.Code, codeTaskVersionConflict)
	}
}
