package postgres

import (
	"context"
	"errors"
	"go-workbench/internal/capture"
	"testing"

	"github.com/jackc/pgx/v5"
)

func createTestOwner(t *testing.T, ctx context.Context, conn *pgx.Conn, suffix string) string {
	t.Helper()

	var ownerID string
	err := conn.QueryRow(ctx, `
		INSERT INTO owners (oidc_issuer, oidc_subject)
		VALUES ($1, $2)
		RETURNING id::text
	`, "https://issuer.example.com", "capture-test-"+suffix).Scan(&ownerID)
	if err != nil {
		t.Fatalf("create owner: %v", err)
	}

	return ownerID
}

func TestCaptureRepositoryCreateIsIdempotentForSameInput(t *testing.T) {
	ctx, conn, schema := newTestDatabase(t)
	applyBusinessMigration(t, ctx, conn, schema)

	ownerID := createTestOwner(t, ctx, conn, "same-input")
	repository := NewCaptureRepository(conn)
	record := capture.CreateRecord{
		OwnerID:        ownerID,
		IdempotencyKey: "idem-1",
		InputHash:      "9b5f96192bdca67df6dc7c4b9b9c46b4642a4139c2896ad40bd6e0f64d008e47",
		InputText:      "Add this to the workbench",
		SourceType:     "manual",
	}

	first, err := repository.Create(ctx, record)
	if err != nil {
		t.Fatalf("first Create() error = %v", err)
	}

	second, err := repository.Create(ctx, record)
	if err != nil {
		t.Fatalf("second Create() error = %v", err)
	}

	if first.ID == "" {
		t.Fatal("first ID is empty")
	}
	if second.ID != first.ID {
		t.Fatalf("second ID = %q, want %q", second.ID, first.ID)
	}
	if second.Status != capture.StatusQueued {
		t.Fatalf("status = %q, want %q", second.Status, capture.StatusQueued)
	}

	var count int
	err = conn.QueryRow(ctx, `
		SELECT COUNT(*) FROM captures WHERE owner_id = $1::uuid AND idempotency_key = $2
	`, ownerID, "idem-1").Scan(&count)
	if err != nil {
		t.Fatalf("count captures: %v", err)
	}
	if count != 1 {
		t.Fatalf("capture count = %d, want 1", count)
	}
}

func TestCaptureRepositoryCreateRejectsSameKeyWithDifferentInput(t *testing.T) {
	ctx, conn, schema := newTestDatabase(t)
	applyBusinessMigration(t, ctx, conn, schema)

	ownerID := createTestOwner(t, ctx, conn, "conflict")
	repository := NewCaptureRepository(conn)

	_, err := repository.Create(ctx, capture.CreateRecord{
		OwnerID:        ownerID,
		IdempotencyKey: "idem-1",
		InputHash:      "9b5f96192bdca67df6dc7c4b9b9c46b4642a4139c2896ad40bd6e0f64d008e47",
		InputText:      "Add this to the workbench",
		SourceType:     "manual",
	})
	if err != nil {
		t.Fatalf("first Create() error = %v", err)
	}

	_, err = repository.Create(ctx, capture.CreateRecord{
		OwnerID:        ownerID,
		IdempotencyKey: "idem-1",
		InputHash:      "ef70b4c194f9940a937b97b1bf6eb06071e45e3d4a8e49c9b09bd837d1ad502a",
		InputText:      "A different visible task",
		SourceType:     "manual",
	})
	if err == nil {
		t.Fatal("second Create() error = nil, want conflict")
	}
	if !errors.Is(err, capture.ErrIdempotencyConflict) {
		t.Fatalf("second Create() error = %v, want idempotency conflict", err)
	}
}

func TestCaptureRepositoryCreateAllowsSameKeyForDifferentOwners(t *testing.T) {
	ctx, conn, schema := newTestDatabase(t)
	applyBusinessMigration(t, ctx, conn, schema)

	ownerAID := createTestOwner(t, ctx, conn, "owner-a")
	ownerBID := createTestOwner(t, ctx, conn, "owner-b")
	repository := NewCaptureRepository(conn)

	for _, ownerID := range []string{ownerAID, ownerBID} {
		_, err := repository.Create(ctx, capture.CreateRecord{
			OwnerID:        ownerID,
			IdempotencyKey: "shared-key",
			InputHash:      "9b5f96192bdca67df6dc7c4b9b9c46b4642a4139c2896ad40bd6e0f64d008e47",
			InputText:      "Add this to the workbench",
			SourceType:     "manual",
		})
		if err != nil {
			t.Fatalf("Create(%s) error = %v", ownerID, err)
		}
	}
}
