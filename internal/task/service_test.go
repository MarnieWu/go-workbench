package task

import (
	"context"
	"reflect"
	"testing"
)

type stubRepository struct {
	tasks []Task
	err   error
}

func (r stubRepository) List(_ context.Context, _ string) ([]Task, error) {
	return r.tasks, r.err
}

func TestServiceListReturnsOwnerTasks(t *testing.T) {
	t.Parallel()

	want := []Task{{
		ID:      "task-1",
		OwnerID: "owner-1",
		Title:   "Learn Go service boundaries",
		Status:  StatusBacklog,
	}}
	service := NewService(stubRepository{tasks: want})

	got, err := service.List(context.Background(), "owner-1")
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("List() = %#v, want %#v", got, want)
	}
}
