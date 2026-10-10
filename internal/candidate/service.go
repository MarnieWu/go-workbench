package candidate

import (
	"context"
	"errors"
	"time"
)

type Status string

const (
	StatusPendingReview Status = "pending_review"
	StatusAccepted      Status = "accepted"
	StatusRejected      Status = "rejected"
)

var (
	ErrNotFound      = errors.New("candidate not found")
	ErrStateConflict = errors.New("candidate state conflict")
)

type RejectInput struct {
	OwnerID     string
	CandidateID string
	RequestID   string
}

type AcceptInput struct {
	OwnerID     string
	CandidateID string
	Title       string
	Description *string
	ProjectID   *string
	Labels      []string
	RequestID   string
}

type AcceptResult struct {
	CandidateID string
	TaskID      string
}

type Candidate struct {
	ID                  string
	OwnerID             string
	CaptureID           string
	ProposedTitle       string
	ProposedDescription *string
	ProposedProjectID   *string
	Labels              []string
	Status              Status
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

type Repository interface {
	ListPending(ctx context.Context, ownerID string) ([]Candidate, error)
}

type Service struct {
	repository Repository
}

func NewService(repository Repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) ListPending(ctx context.Context, ownerID string) ([]Candidate, error) {
	return s.repository.ListPending(ctx, ownerID)
}
