package task

import (
	"context"
	"reflect"
	"testing"
)

type stubRepository struct {
	tasks   []Task
	err     error
	ownerID string
	filter  ListTasksFilter
	calls   int
}

func (r *stubRepository) List(
	_ context.Context,
	ownerID string,
	filter ListTasksFilter,
) ([]Task, error) {
	r.calls++
	r.ownerID = ownerID
	r.filter = filter

	return r.tasks, r.err
}

func TestServiceListReturnsOwnerTasks(t *testing.T) {
	t.Parallel()

	status := StatusBacklog
	want := []Task{{
		ID:      "task-1",
		OwnerID: "owner-1",
		Title:   "Learn Go service boundaries",
		Status:  status,
	}}
	repository := &stubRepository{tasks: want}
	service := NewService(repository)

	got, err := service.List(context.Background(), "owner-1", ListTasksFilter{
		Status: &status,
	})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("List() = %#v, want %#v", got, want)
	}
	if repository.calls != 1 {
		t.Fatalf("repository calls = %d, want 1", repository.calls)
	}
	if repository.ownerID != "owner-1" {
		t.Fatalf("repository ownerID = %q, want %q", repository.ownerID, "owner-1")
	}
	if repository.filter.Status == nil || *repository.filter.Status != status {
		t.Fatalf("repository status filter = %#v, want %q", repository.filter.Status, status)
	}
}
