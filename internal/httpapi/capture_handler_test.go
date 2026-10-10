package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"go-workbench/internal/capture"
	"go-workbench/internal/task"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
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

func TestRouterCreateCaptureRejectsInvalidRequestsBeforeRepository(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{name: "invalid JSON", body: `{"idempotencyKey":`},
		{name: "unknown field", body: `{"idempotencyKey":"idem-1","inputText":"task","sourceType":"manual","unknown":true}`},
		{name: "empty input", body: `{"idempotencyKey":"idem-1","inputText":"   ","sourceType":"manual"}`},
		{name: "oversized input", body: `{"idempotencyKey":"idem-1","inputText":"` + string(bytes.Repeat([]byte("a"), 251)) + `","sourceType":"manual"}`},
		{name: "missing idempotency key", body: `{"inputText":"task","sourceType":"manual"}`},
		{name: "invalid source", body: `{"idempotencyKey":"idem-1","inputText":"task","sourceType":"unknown"}`},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repository := &stubCaptureRepository{}
			router := NewRouter(RouterConfig{
				TaskService:    task.NewService(&stubTaskRepository{}),
				CaptureService: capture.NewService(repository),
			}, testOwner("owner-1"))
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodPost, "/v1/captures", bytes.NewBufferString(test.body))
			request.Header.Set("Content-Type", "application/json")

			router.ServeHTTP(recorder, request)

			if recorder.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want %d; body = %s", recorder.Code, http.StatusBadRequest, recorder.Body.String())
			}
			if repository.calls != 0 {
				t.Fatalf("repository calls = %d, want 0", repository.calls)
			}
			var response errorResponse
			if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
				t.Fatalf("decode response: %v", err)
			}
			if response.Code != codeInvalidCaptureRequest || response.Message != messageInvalidCaptureRequest {
				t.Fatalf("error = %q/%q, want stable invalid-request response", response.Code, response.Message)
			}
		})
	}
}

func TestRouterCreateCaptureLogOmitsSensitiveRequestData(t *testing.T) {
	var logs bytes.Buffer
	repository := &stubCaptureRepository{err: errors.New("database unavailable")}
	router := NewRouter(RouterConfig{
		TaskService:    task.NewService(&stubTaskRepository{}),
		CaptureService: capture.NewService(repository),
		Logger:         slog.New(slog.NewJSONHandler(&logs, nil)),
	}, testOwner("owner-1"))
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/v1/captures", bytes.NewBufferString(`{
		"idempotencyKey":"secret-idempotency-key",
		"inputText":"private capture body",
		"sourceType":"manual"
	}`))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer secret-token")
	request.Header.Set("Cookie", "session=secret-cookie")

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusInternalServerError)
	}
	for _, secret := range []string{"private capture body", "secret-idempotency-key", "secret-token", "secret-cookie", "database unavailable"} {
		if strings.Contains(logs.String(), secret) || strings.Contains(recorder.Body.String(), secret) {
			t.Fatalf("response or log leaked %q", secret)
		}
	}
	if !strings.Contains(logs.String(), `"request_id"`) {
		t.Fatal("log is missing request_id")
	}
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
