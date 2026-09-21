package task

import (
	"context"
	"time"
)

type Status string

const StatusBacklog Status = "backlog"

type Priority string

const (
	PriorityNone   Priority = "none"
	PriorityLow    Priority = "low"
	PriorityMedium Priority = "medium"
	PriorityHigh   Priority = "high"
)

type Task struct {
	ID        string
	OwnerID   string
	Title     string
	Status    Status
	Priority  Priority
	Labels    []string
	Version   int64
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Repository interface {
	List(ctx context.Context, ownerID string) ([]Task, error)
}

type Service struct {
	repository Repository
}

func NewService(repository Repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) List(ctx context.Context, ownerId string) ([]Task, error) {
	return s.repository.List(ctx, ownerId)
}
