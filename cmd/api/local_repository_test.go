package main

import (
	"context"
	"testing"
	"go-workbench/internal/task"
)

func TestLocalRepositoryList(t *testing.T) {
	fixtures := []string{"success", "empty", "error"}
	status := task.StatusBacklog
	filter := task.ListTasksFilter {
		Status: &status,
	}

	for _, fixture := range fixtures {
		repository := LocalRepository{fixture: fixture}

		t.Run(fixture, func(t *testing.T) {
			tasks, err := repository.List(context.Background(), localOwnerId, filter)

			if fixture == "success" {
				if len(tasks) == 0 {
					t.Fatal("tasks = empty, want non-empty")
				}
				if err != nil {
					t.Fatalf("err = %v, want nil", err)
				}
			}

			if fixture == "empty" {
				if len(tasks) != 0 {
					t.Fatalf("len(tasks) = %d, want 0", len(tasks))
				}
				if err != nil {
					t.Fatalf("err = %v, want nil", err)
				}
			}

			if fixture == "error" {
				if len(tasks) != 0 {
					t.Fatalf("len(tasks) = %d, want 0", len(tasks))
				}
				if err == nil {
					t.Fatal("err = nil, want non-nil")
				}
			}
		})
	}
}
