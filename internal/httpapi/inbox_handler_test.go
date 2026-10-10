package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"go-workbench/internal/candidate"
	"net/http"
	"net/http/httptest"
	"testing"
)

type stubCandidateRepository struct {
	items   []candidate.Candidate
	err     error
	ownerID string
	calls   int
}

func (r *stubCandidateRepository) ListPending(_ context.Context, ownerID string) ([]candidate.Candidate, error) {
	r.calls++
	r.ownerID = ownerID
	return r.items, r.err
}

func TestRouterListInboxReturnsOwnerPendingCandidates(t *testing.T) {
	repository := &stubCandidateRepository{items: []candidate.Candidate{{
		ID:            "candidate-1",
		OwnerID:       "owner-1",
		CaptureID:     "capture-1",
		ProposedTitle: "Review generated task",
		Labels:        []string{"review"},
		Status:        candidate.StatusPendingReview,
	}}}
	router := NewRouter(RouterConfig{CandidateService: candidate.NewService(repository)}, testOwner("owner-1"))
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/v1/inbox", nil)

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body = %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	var response listInboxResponse
	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(response.Items) != 1 || response.Items[0].ID != "candidate-1" {
		t.Fatalf("items = %#v, want candidate-1", response.Items)
	}
	if repository.ownerID != "owner-1" || repository.calls != 1 {
		t.Fatalf("repository owner/calls = %q/%d, want owner-1/1", repository.ownerID, repository.calls)
	}
}

func TestRouterListInboxReturnsEmptyItems(t *testing.T) {
	repository := &stubCandidateRepository{}
	router := NewRouter(RouterConfig{CandidateService: candidate.NewService(repository)}, testOwner("owner-1"))
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/v1/inbox", nil))

	var response struct {
		Items []json.RawMessage `json:"items"`
	}
	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.Items == nil || len(response.Items) != 0 {
		t.Fatalf("items = %#v, want []", response.Items)
	}
}

func TestRouterListInboxErrorsAreSafe(t *testing.T) {
	tests := []struct {
		name       string
		ownerID    string
		repository *stubCandidateRepository
		wantStatus int
		wantCode   string
	}{
		{name: "missing owner", repository: &stubCandidateRepository{}, wantStatus: http.StatusUnauthorized, wantCode: codeUnauthorized},
		{name: "repository failure", ownerID: "owner-1", repository: &stubCandidateRepository{err: errors.New("sql detail")}, wantStatus: http.StatusInternalServerError, wantCode: codeInternalError},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			config := RouterConfig{CandidateService: candidate.NewService(test.repository)}
			var router http.Handler
			if test.ownerID == "" {
				router = NewRouter(config)
			} else {
				router = NewRouter(config, testOwner(test.ownerID))
			}
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/v1/inbox", nil))
			if recorder.Code != test.wantStatus {
				t.Fatalf("status = %d, want %d", recorder.Code, test.wantStatus)
			}
			var response errorResponse
			if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
				t.Fatalf("decode response: %v", err)
			}
			if response.Code != test.wantCode {
				t.Fatalf("code = %q, want %q", response.Code, test.wantCode)
			}
		})
	}
}
