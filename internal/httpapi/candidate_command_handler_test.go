package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"go-workbench/internal/candidate"
	"net/http"
	"net/http/httptest"
	"testing"
)

type stubCandidateCommands struct {
	rejectInput candidate.RejectInput
	rejectItem  candidate.Candidate
	acceptInput candidate.AcceptInput
	acceptItem  candidate.AcceptResult
	err         error
}

func (s *stubCandidateCommands) Reject(_ context.Context, input candidate.RejectInput) (candidate.Candidate, error) {
	s.rejectInput = input
	return s.rejectItem, s.err
}

func (s *stubCandidateCommands) Accept(_ context.Context, input candidate.AcceptInput) (candidate.AcceptResult, error) {
	s.acceptInput = input
	return s.acceptItem, s.err
}

func TestRouterRejectCandidateConnectsOwnerAndRequestID(t *testing.T) {
	commands := &stubCandidateCommands{rejectItem: candidate.Candidate{ID: "candidate-1", Status: candidate.StatusRejected}}
	router := NewRouter(RouterConfig{CandidateRejecter: commands}, testOwner("owner-1"))
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/v1/candidates/candidate-1/reject", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body = %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	if commands.rejectInput.OwnerID != "owner-1" || commands.rejectInput.CandidateID != "candidate-1" {
		t.Fatalf("input = %#v", commands.rejectInput)
	}
	if commands.rejectInput.RequestID == "" {
		t.Fatal("request ID is empty")
	}
}

func TestRouterAcceptCandidateConnectsEditableFields(t *testing.T) {
	commands := &stubCandidateCommands{acceptItem: candidate.AcceptResult{CandidateID: "candidate-1", TaskID: "task-1"}}
	router := NewRouter(RouterConfig{CandidateAccepter: commands}, testOwner("owner-1"))
	body := bytes.NewBufferString(`{"title":"Edited title","description":"Edited description","labels":["go"]}`)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/v1/candidates/candidate-1/accept", body)
	request.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d; body = %s", recorder.Code, http.StatusCreated, recorder.Body.String())
	}
	if commands.acceptInput.OwnerID != "owner-1" || commands.acceptInput.Title != "Edited title" {
		t.Fatalf("input = %#v", commands.acceptInput)
	}
	if commands.acceptInput.RequestID == "" {
		t.Fatal("request ID is empty")
	}
}

func TestRouterCandidateCommandsMapSafeErrors(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantCode   string
	}{
		{name: "not found", err: candidate.ErrNotFound, wantStatus: http.StatusNotFound, wantCode: codeCandidateNotFound},
		{name: "conflict", err: candidate.ErrStateConflict, wantStatus: http.StatusConflict, wantCode: codeCandidateStateConflict},
		{name: "internal", err: errors.New("sql detail"), wantStatus: http.StatusInternalServerError, wantCode: codeInternalError},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			commands := &stubCandidateCommands{err: test.err}
			router := NewRouter(RouterConfig{CandidateRejecter: commands}, testOwner("owner-1"))
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/v1/candidates/candidate-1/reject", nil))
			if recorder.Code != test.wantStatus {
				t.Fatalf("status = %d, want %d", recorder.Code, test.wantStatus)
			}
			var response errorResponse
			if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
				t.Fatalf("decode response: %v", err)
			}
			if response.Code != test.wantCode || bytes.Contains(recorder.Body.Bytes(), []byte("sql detail")) {
				t.Fatalf("response = %#v; body = %s", response, recorder.Body.String())
			}
		})
	}
}
