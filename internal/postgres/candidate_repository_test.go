package postgres

import (
	"context"
	"go-workbench/internal/candidate"
	"testing"

	"github.com/jackc/pgx/v5"
)

func createCandidateFixture(
	t *testing.T,
	ctx context.Context,
	conn *pgx.Conn,
	ownerID string,
	status candidate.Status,
	title string,
) string {
	t.Helper()
	var captureID string
	if err := conn.QueryRow(ctx, `
		INSERT INTO captures (owner_id, idempotency_key, input_hash, input_text, source_type)
		VALUES ($1::uuid, gen_random_uuid()::text, repeat('a', 64), 'fixture', 'manual')
		RETURNING id::text
	`, ownerID).Scan(&captureID); err != nil {
		t.Fatalf("create capture fixture: %v", err)
	}
	var candidateID string
	if err := conn.QueryRow(ctx, `
		INSERT INTO candidates (owner_id, capture_id, proposed_title, status)
		VALUES ($1::uuid, $2::uuid, $3, $4)
		RETURNING id::text
	`, ownerID, captureID, title, status).Scan(&candidateID); err != nil {
		t.Fatalf("create candidate fixture: %v", err)
	}
	return candidateID
}

func TestCandidateRepositoryListPendingScopesOwnerAndStatus(t *testing.T) {
	ctx, conn, schema := newTestDatabase(t)
	applyBusinessMigration(t, ctx, conn, schema)
	ownerAID := createTestOwner(t, ctx, conn, "inbox-owner-a")
	ownerBID := createTestOwner(t, ctx, conn, "inbox-owner-b")
	pendingID := createCandidateFixture(t, ctx, conn, ownerAID, candidate.StatusPendingReview, "pending")
	createCandidateFixture(t, ctx, conn, ownerAID, candidate.StatusRejected, "rejected")
	createCandidateFixture(t, ctx, conn, ownerBID, candidate.StatusPendingReview, "other owner")

	items, err := NewCandidateRepository(conn).ListPending(ctx, ownerAID)
	if err != nil {
		t.Fatalf("ListPending() error = %v", err)
	}
	if len(items) != 1 || items[0].ID != pendingID {
		t.Fatalf("items = %#v, want only pending candidate %q", items, pendingID)
	}
	if items[0].OwnerID != ownerAID || items[0].Status != candidate.StatusPendingReview {
		t.Fatalf("candidate = %#v, want owner-scoped pending item", items[0])
	}
}

func TestCandidateRepositoryListPendingReturnsEmptySlice(t *testing.T) {
	ctx, conn, schema := newTestDatabase(t)
	applyBusinessMigration(t, ctx, conn, schema)
	ownerID := createTestOwner(t, ctx, conn, "empty-inbox")

	items, err := NewCandidateRepository(conn).ListPending(ctx, ownerID)
	if err != nil {
		t.Fatalf("ListPending() error = %v", err)
	}
	if items == nil || len(items) != 0 {
		t.Fatalf("items = %#v, want []", items)
	}
}
