package capture

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"testing"
)

type stubRepository struct {
	record CreateRecord
	result Capture
	err    error
	calls  int
}

func (r *stubRepository) Create(_ context.Context, record CreateRecord) (Capture, error) {
	r.calls++
	r.record = record
	return r.result, r.err
}

func TestServiceCreateHashesInputAndCreatesQueuedCapture(t *testing.T) {
	repository := &stubRepository{
		result: Capture{
			ID:             "capture-1",
			OwnerID:        "owner-1",
			IdempotencyKey: "idem-1",
			InputText:      "Add this to the workbench",
			SourceType:     "manual",
			Status:         StatusQueued,
		},
	}
	service := NewService(repository)

	got, err := service.Create(context.Background(), CreateInput{
		OwnerID:        "owner-1",
		IdempotencyKey: "idem-1",
		InputText:      "Add this to the workbench",
		SourceType:     "manual",
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if got.Status != StatusQueued {
		t.Fatalf("status = %q, want %q", got.Status, StatusQueued)
	}
	if repository.calls != 1 {
		t.Fatalf("repository calls = %d, want 1", repository.calls)
	}
	if repository.record.OwnerID != "owner-1" {
		t.Fatalf("ownerID = %q, want owner-1", repository.record.OwnerID)
	}
	if repository.record.IdempotencyKey != "idem-1" {
		t.Fatalf("idempotency key = %q, want idem-1", repository.record.IdempotencyKey)
	}
	if repository.record.InputText != "Add this to the workbench" {
		t.Fatalf("input text = %q", repository.record.InputText)
	}
	if repository.record.SourceType != "manual" {
		t.Fatalf("source type = %q, want manual", repository.record.SourceType)
	}

	sum := sha256.Sum256([]byte("Add this to the workbench"))
	wantHash := hex.EncodeToString(sum[:])
	if repository.record.InputHash != wantHash {
		t.Fatalf("input hash = %q, want %q", repository.record.InputHash, wantHash)
	}
}
