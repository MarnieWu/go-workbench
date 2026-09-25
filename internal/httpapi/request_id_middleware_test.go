package httpapi

import (
	"bytes"
	"encoding/json"
	"go-workbench/internal/task"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequestIDMiddleware(t *testing.T) {
	service := task.NewService(&stubTaskRepository{})
	router := NewRouter(service, testOwner("owner-1"))

	requestID := func() string {
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodGet, "/v1/tasks", nil)

		router.ServeHTTP(recorder, request)

		id := recorder.Header().Get(requestIDHeader)

		if id == "" {
			t.Fatal("response X-Request-ID is empty")
		}

		return id
	}

	firstID := requestID()
	secondID := requestID()

	if firstID == secondID {
		t.Errorf("two requests received the same request ID %q", firstID)
	}

	tests := []struct {
		name          string
		requestID     string
		wantPreserved bool
	}{
		{
			name:      "client with request ID",
			requestID: "learn-go-001",
		},
		{
			name:      "client no request ID",
			requestID: "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			testID := requestID()

			if testID == test.requestID {
				t.Errorf("response X-Request-ID is %q, want %q", test.requestID, testID)
			}
		})
	}
}

func TestRequestIDIsConsistentAcrossErrorResponseAndLog(t *testing.T) {
	// In the test environment, data is written to memory using `logs`.
	var logs bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logs, nil))

	service := task.NewService(&stubTaskRepository{})
	router := NewRouter(service, requestLoggingMiddleware(logger))

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/v1/tasks", nil)

	router.ServeHTTP(recorder, request)

	var response struct {
		RequestID string `json:"requestId"`
	}
	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatal(err)
	}

	var logEntry struct {
		RequestID string `json:"request_id"`
	}
	if err := json.NewDecoder(&logs).Decode(&logEntry); err != nil {
		t.Fatal(err)
	}

	headerID := recorder.Header().Get(requestIDHeader)

	if headerID == "" {
		t.Fatal("response request ID is empty")
	}

	if headerID != response.RequestID || headerID != logEntry.RequestID {
		t.Fatalf(
			"request IDs differ: header=%q body=%q log=%q",
			headerID,
			response.RequestID,
			logEntry.RequestID,
		)
	}

}
