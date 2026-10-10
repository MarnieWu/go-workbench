package candidate

import (
	"context"
	"reflect"
	"testing"
)

type stubRepository struct {
	items   []Candidate
	err     error
	ownerID string
}

func (r *stubRepository) ListPending(_ context.Context, ownerID string) ([]Candidate, error) {
	r.ownerID = ownerID
	return r.items, r.err
}

func TestServiceListPendingReturnsRepositoryItems(t *testing.T) {
	want := []Candidate{{ID: "candidate-1", OwnerID: "owner-1", Status: StatusPendingReview}}
	repository := &stubRepository{items: want}
	service := NewService(repository)

	got, err := service.ListPending(context.Background(), "owner-1")
	if err != nil {
		t.Fatalf("ListPending() error = %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ListPending() = %#v, want %#v", got, want)
	}
	if repository.ownerID != "owner-1" {
		t.Fatalf("ownerID = %q, want owner-1", repository.ownerID)
	}
}
