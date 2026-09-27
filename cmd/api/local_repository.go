package main

import (
	"context"
	"errors"
	"go-workbench/internal/task"
	"time"
)

const (
	FixtureSuccess string = "success"
	FixtureEmpty   string = "empty"
	FixtureError   string = "error"
)

type LocalRepository struct {
	fixture string
}

const localOwnerId = "owner-local"

func (r LocalRepository) List(
	ctx context.Context,
	ownerID string,
) ([]task.Task, error) {

	if ownerID == localOwnerId {
		switch r.fixture {
		case FixtureSuccess:
			return []task.Task{{
				ID:       "task-1",
				OwnerID:  ownerID,
				Title:    "Local Repository Success Task",
				Status:   task.StatusBacklog,
				Priority: task.PriorityMedium,
				Labels: []string{
					"local",
					"repository",
					"success",
				},
				Version:   1,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}}, nil
		case FixtureEmpty:
			return []task.Task{}, nil
		case FixtureError:
			return nil, errors.New("local fixture query failed")
		default:
			return nil, nil
		}
	}

	return []task.Task{}, nil
}
