package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"go-workbench/internal/capture"
	"go-workbench/internal/task"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRouterCreateCaptureReturnsCreatedCapture(t *testing.T) {
	captureService := capture.NewService(&stubCaptureRepository{
		result: capture.Capture{
			ID:             "capture-1",
			OwnerID:        "owner-1",
			IdempotencyKey: "idem-1",
			InputText:      "Add this to the workbench",
			SourceType:     "manual",
			Status:         capture.StatusQueued,
			CreatedAt:      time.Date(2026, 10, 9, 12, 0, 0, 0, time.UTC),
		},
	})
	taskService := task.NewService(&stubTaskRepository{})
	router := NewRouter(RouterConfig{
		TaskService:    taskService,
		CaptureService: captureService,
	}, testOwner("owner-1"))

	body := bytes.NewBufferString(`{
		"idempotencyKey": "idem-1",
		"inputText": "Add this to the workbench",
		"sourceType": "manual"
	}`)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/v1/captures", body)
	request.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d; body = %s", recorder.Code, http.StatusCreated, recorder.Body.String())
	}

	var response struct {
		ID        string `json:"id"`
		Status    string `json:"status"`
		RequestID string `json:"requestId"`
	}
	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.ID == "" {
		t.Fatal("id is empty")
	}
	if response.Status != string(capture.StatusQueued) {
		t.Fatalf("status = %q, want %q", response.Status, capture.StatusQueued)
	}
	if response.RequestID == "" {
		t.Fatal("requestId is empty")
	}
}

func TestRouterCreateCaptureMapsIdempotencyConflict(t *testing.T) {
	captureService := capture.NewService(&stubCaptureRepository{
		err: capture.ErrIdempotencyConflict,
	})
	taskService := task.NewService(&stubTaskRepository{})
	router := NewRouter(RouterConfig{
		TaskService:    taskService,
		CaptureService: captureService,
	}, testOwner("owner-1"))

	body := bytes.NewBufferString(`{
		"idempotencyKey": "idem-1",
		"inputText": "A different visible task",
		"sourceType": "manual"
	}`)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/v1/captures", body)
	request.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d; body = %s", recorder.Code, http.StatusConflict, recorder.Body.String())
	}

	var response struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	}
	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.Code != codeIdempotencyConflict {
		t.Fatalf("code = %q, want %q", response.Code, codeIdempotencyConflict)
	}
	if response.Message != messageIdempotencyConflict {
		t.Fatalf("message = %q, want %q", response.Message, messageIdempotencyConflict)
	}
	if bytes.Contains(recorder.Body.Bytes(), []byte("Add this to the workbench")) {
		t.Fatal("response leaked prior input body")
	}
}

type stubCaptureRepository struct {
	result capture.Capture
	err    error
	record capture.CreateRecord
	calls  int
}

func (r *stubCaptureRepository) Create(
	_ context.Context,
	record capture.CreateRecord,
) (capture.Capture, error) {
	r.calls++
	r.record = record
	if r.err != nil {
		return capture.Capture{}, r.err
	}
	return r.result, nil
}

func TestRouterCreateCapturePreservesWrappedIdempotencyConflict(t *testing.T) {
	captureService := capture.NewService(&stubCaptureRepository{
		err: errors.Join(errors.New("repository conflict"), capture.ErrIdempotencyConflict),
	})
	taskService := task.NewService(&stubTaskRepository{})
	router := NewRouter(RouterConfig{
		TaskService:    taskService,
		CaptureService: captureService,
	}, testOwner("owner-1"))

	body := bytes.NewBufferString(`{
		"idempotencyKey": "idem-1",
		"inputText": "A different visible task",
		"sourceType": "manual"
	}`)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/v1/captures", body)
	request.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d; body = %s", recorder.Code, http.StatusConflict, recorder.Body.String())
	}
}
