package task

import (
	"context"
	"errors"
	"time"
)

type Status string

const (
	StatusBacklog    Status = "backlog"
	StatusInProgress Status = "in_progress"
	StatusBlocked    Status = "blocked"
	StatusDone       Status = "done"
)

var (
	ErrNotFound        = errors.New("task not found")
	ErrVersionConflict = errors.New("task version conflict")
)

type Priority string

const (
	PriorityNone   Priority = "none"
	PriorityLow    Priority = "low"
	PriorityMedium Priority = "medium"
	PriorityHigh   Priority = "high"
)

type Task struct {
	ID          string
	OwnerID     string
	ProjectID   *string
	Title       string
	Description string
	Status      Status
	Priority    Priority
	Labels      []string
	DueAt       *time.Time
	ArchivedAt  *time.Time
	Version     int64
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type UpdateInput struct {
	OwnerID     string
	TaskID      string
	Version     int64
	ProjectID   *string
	Title       string
	Description string
	Status      Status
	Priority    Priority
	Labels      []string
	DueAt       *time.Time
	RequestID   string
}

type ListTasksFilter struct {
	Status *Status
}

type Repository interface {
	List(ctx context.Context, ownerID string, filter ListTasksFilter) ([]Task, error)
}

type Service struct {
	repository Repository
}

func NewService(repository Repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) List(ctx context.Context, ownerId string, filter ListTasksFilter) ([]Task, error) {
	return s.repository.List(ctx, ownerId, filter)
}
