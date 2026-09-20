package task

import (
	"context"
)

type Status string

const StatusBacklog Status = "backlog"

type Task struct {
	ID      string
	OwnerID string
	Title   string
	Status  Status
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
