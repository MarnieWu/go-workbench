package postgres

import (
	"fmt"
	"strings"
	"testing"
)

func TestSchemaTasks(t *testing.T) {
	ctx, conn, schema := newTestDatabase(t)
	applyBusinessMigration(t, ctx, conn, schema)

	t.Run("InvalidStatus", func(t *testing.T) {
		var ownerID string
		err := conn.QueryRow(ctx, `
			INSERT INTO owners (oidc_issuer, oidc_subject)
			VALUES ($1, $2)
			RETURNING id::text
		`, "https://issuer.example.com", "test-subject").Scan(&ownerID)
		if err != nil {
			t.Fatal(err)
		}

		_, err = conn.Exec(ctx, "INSERT INTO tasks (owner_id, title, status) VALUES ($1::uuid, $2, $3)", ownerID, "invalid status task", "unknown")
		assertPgError(t, err, "23514", "chk_task_status")
	})

	t.Run("MissingOwner", func(t *testing.T) {
		// TODO: otherwise valid Task with nonexistent owner; assert 23503 and no row.
		_, err := conn.Exec(ctx, "INSERT INTO tasks (owner_id, title, status) VALUES ($1::uuid, $2, $3)", "00000000-0000-0000-0000-000000000000", "task with missing owner", "pending")
		assertPgError(t, err, "23503", "fk_task_owner_id_project")
	})

	t.Run("CrossOwnerProject", func(t *testing.T) {
		// TODO: Owner A Task referencing Owner B Project; assert 23503 and no Task.
		var ownerIdA, ownerIdB string
		err := conn.QueryRow(ctx, `
			INSERT INTO owners (oidc_issuer, oidc_subject)
			VALUES ($1, $2)
			RETURNING id::text
		`, "https://issuer.example.com", "test-subject-a").Scan(&ownerIdA)
		if err != nil {
			t.Fatal(err)
		}
		err = conn.QueryRow(ctx, `
			INSERT INTO owners (oidc_issuer, oidc_subject)
			VALUES ($1, $2)
			RETURNING id::text
		`, "https://issuer.example.com", "test-subject-b").Scan(&ownerIdB)
		if err != nil {
			t.Fatal(err)
		}
		var projectIdB string
		err = conn.QueryRow(ctx, `
			INSERT INTO projects (owner_id, name)
			VALUES ($1::uuid, $2)
			RETURNING id::text
		`, ownerIdB, "project-b").Scan(&projectIdB)
		if err != nil {
			t.Fatal(err)
		}
		_, err = conn.Exec(ctx, "INSERT INTO tasks (owner_id, project_id, title, status) VALUES ($1::uuid, $2::uuid, $3, $4)", ownerIdA, projectIdB, "task with cross-owner project", "pending")
		assertPgError(t, err, "23503", "fk_task_owner_id_project")
	})
}

func TestSchemaCaptures(t *testing.T) {
	ctx, conn, schema := newTestDatabase(t)
	applyBusinessMigration(t, ctx, conn, schema)

	ownerIDs := make([]string, 2)
	for idx := range ownerIDs {
		err := conn.QueryRow(ctx, `
		INSERT INTO owners (oidc_issuer, oidc_subject)
		VALUES ($1, $2)
		RETURNING id::text
	`, "https://issuer.example.com", fmt.Sprintf("test-subject-%d", idx)).Scan(&ownerIDs[idx])
		if err != nil {
			t.Fatal(err)
		}
	}

	type capture struct {
		IdempotencyKey string
		InputHash      string
		InputText      string
		SourceType     string
	}

	t.Run("DuplicateKey", func(t *testing.T) {
		capture := capture{
			IdempotencyKey: "duplicate-key",
			InputHash:      "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9",
			InputText:      "input text",
			SourceType:     "test-source",
		}
		sql := "INSERT INTO captures (owner_id, idempotency_key, input_hash, input_text, source_type) VALUES ($1::uuid, $2, $3, $4, $5)"

		_, err := conn.Exec(ctx, sql, ownerIDs[0], capture.IdempotencyKey, capture.InputHash, capture.InputText, capture.SourceType)
		if err != nil {
			t.Fatal(err)
		}
		_, err = conn.Exec(ctx, sql, ownerIDs[0], capture.IdempotencyKey, capture.InputHash, capture.InputText, capture.SourceType)
		assertPgError(t, err, "23505", "uk_capture_owner_id_idempotency_key")

		var count int
		err = conn.QueryRow(ctx, "SELECT COUNT(*) FROM captures WHERE owner_id = $1 AND idempotency_key = $2", ownerIDs[0], capture.IdempotencyKey).Scan(&count)
		if err != nil {
			t.Fatal(err)
		}
		if count != 1 {
			t.Errorf("Expected 1 capture, got %d", count)
		}
	})

	t.Run("InputHash", func(t *testing.T) {
		insert := `INSERT INTO captures (owner_id, idempotency_key, input_hash, input_text, source_type)
		VALUES ($1::uuid, $2, $3, $4, $5)`
		_, err := conn.Exec(ctx, insert, ownerIDs[0], "valid", strings.Repeat("a", 64), "input text", "test-source")
		if err != nil {
			t.Fatal(err)
		}

		for _, tc := range []struct {
			name string
			hash string
		}{
			{"too short", strings.Repeat("a", 63)},
			{"too long", strings.Repeat("a", 65)},
			{"non-hex character", strings.Repeat("a", 63) + "g"},
			{"uppercase character", strings.Repeat("a", 63) + "A"},
		} {
			t.Run(tc.name, func(t *testing.T) {
				_, err := conn.Exec(ctx, insert, ownerIDs[0], tc.name, tc.hash, "input text", "test-source")
				assertPgError(t, err, "23514", "chk_capture_input_hash")
			})
		}
	})

	t.Run("KeyAcrossOwners", func(t *testing.T) {
		idempotencyKey := "same-key"
		for _, ownerID := range ownerIDs {
			_, err := conn.Exec(ctx, "INSERT INTO captures (owner_id, idempotency_key, input_hash, input_text, source_type) VALUES ($1::uuid, $2, $3, $4, $5)", ownerID, idempotencyKey, "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde1", "input text", "test-source")
			if err != nil {
				t.Fatal(err)
			}
		}

		var count int
		err := conn.QueryRow(ctx, `
			SELECT COUNT(*) FROM captures WHERE idempotency_key = $1
			AND owner_id IN ($2::uuid, $3::uuid)
		`, idempotencyKey, ownerIDs[0], ownerIDs[1]).Scan(&count)
		if err != nil {
			t.Fatal(err)
		}
		if count != 2 {
			t.Fatalf("captures with shared key = %d, want 2", count)
		}
	})
}

func TestSchemaTaskEvidenceLinks(t *testing.T) {
	ctx, conn, schema := newTestDatabase(t)
	applyBusinessMigration(t, ctx, conn, schema)

	t.Run("DuplicateEvidenceLink", func(t *testing.T) {
		// TODO: valid Task and Evidence; link twice; assert 23505 and one link.
		t.Fatal("RED: fill duplicate evidence link assertions")
	})
}
