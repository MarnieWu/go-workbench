package postgres

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestMigrationEmptySchema(t *testing.T) {
	ctx, conn, schema := newTestDatabase(t)
	applyBusinessMigration(t, ctx, conn, schema)

	tableNames := []string{
		"owners",
		"projects",
		"tasks",
		"captures",
		"check_items",
		"candidates",
		"source_evidence",
		"task_evidence_links",
		"pending_actions",
		"audit_events",
	}
	var exists bool
	for _, tableName := range tableNames {
		err := conn.QueryRow(ctx, `SELECT EXISTS (
		SELECT 1 FROM information_schema.tables
		WHERE table_schema = $1 AND table_name = $2
	)`, schema, tableName).Scan(&exists)
		if err != nil {
			t.Fatal(err)
		}
		if !exists {
			t.Fatalf("%s table is missing", tableName)
		}
	}
	columns := []struct {
		table    string
		name     string
		dataType string
		nullable string
	}{
		{"owners", "id", "uuid", "NO"},
		{"owners", "oidc_subject", "text", "NO"},

		{"projects", "owner_id", "uuid", "NO"},
		{"projects", "version", "bigint", "NO"},

		{"tasks", "project_id", "uuid", "YES"},
		{"tasks", "labels", "ARRAY", "NO"},
		{"tasks", "due_at", "timestamp with time zone", "YES"},

		{"captures", "idempotency_key", "text", "NO"},
		{"captures", "input_hash", "text", "NO"},

		{"check_items", "task_id", "uuid", "NO"},
		{"check_items", "completed", "boolean", "NO"},

		{"candidates", "capture_id", "uuid", "NO"},
		{"candidates", "accepted_task_id", "uuid", "YES"},

		{"source_evidence", "capture_id", "uuid", "NO"},
		{"source_evidence", "source_type", "text", "NO"},

		{"task_evidence_links", "task_id", "uuid", "NO"},
		{"task_evidence_links", "source_evidence_id", "uuid", "NO"},

		{"pending_actions", "target_version", "bigint", "NO"},
		{"pending_actions", "parameters", "jsonb", "NO"},
		{"pending_actions", "expires_at", "timestamp with time zone", "NO"},

		{"audit_events", "actor_id", "text", "NO"},
		{"audit_events", "metadata", "jsonb", "NO"},
		{"audit_events", "request_id", "text", "YES"},
	}
	for _, c := range columns {
		var matches bool
		err := conn.QueryRow(ctx, `SELECT EXISTS (
			SELECT 1 FROM information_schema.columns
			WHERE table_schema = $1
			  AND table_name = $2
				AND column_name = $3
				AND data_type = $4
				AND is_nullable = $5
		)`, schema, c.table, c.name, c.dataType, c.nullable).Scan(&matches)
		if err != nil {
			t.Fatal(err)
		}
		if !matches {
			t.Errorf("%s.%s is missing or has the wrong type/nullability", c.table, c.name)
		}
	}

	constraints := []struct {
		table string
		name  string
	}{
		{"owners", "uk_owner_oidc_issuer_oidc_subject_key"},
		{"projects", "uk_project_owner_id_id"},
		{"tasks", "fk_task_project"},
		{"captures", "uk_capture_owner_id_idempotency_key"},
		{"captures", "chk_capture_input_hash"},
		{"check_items", "uk_check_item_task_id_position"},
		{"candidates", "fk_candidate_capture"},
		{"candidates", "chk_candidate_acceptance"},
		{"source_evidence", "fk_source_evidence_owner_id_capture_id"},
		{"task_evidence_links", "fk_task_evidence_link_owner_id_task_id"},
		{"task_evidence_links", "fk_task_evidence_link_owner_id_source_evidence_id"},
		{"task_evidence_links", "uk_task_evidence_link_task_id_source_evidence_id"},
		{"pending_actions", "chk_pending_parameters_object"},
		{"audit_events", "chk_audit_metadata_object"},
	}
	for _, constraint := range constraints {
		var exists bool
		err := conn.QueryRow(ctx, `SELECT EXISTS (
			SELECT 1 FROM pg_catalog.pg_constraint AS c
			JOIN pg_catalog.pg_class AS t ON t.oid = c.conrelid
			JOIN pg_catalog.pg_namespace AS n ON n.oid = t.relnamespace
			WHERE n.nspname = $1
				AND t.relname = $2
				AND c.conname = $3
		)`, schema, constraint.table, constraint.name).Scan(&exists)
		if err != nil {
			t.Fatal(err)
		}
		if !exists {
			t.Errorf("%s.%s is missing constraint %s", schema, constraint.table, constraint.name)
		}
	}
}

func TestMigrationRepeat(t *testing.T) {
	ctx, conn, schema := newTestDatabase(t)
	applyBusinessMigration(t, ctx, conn, schema)

	var ownerID, taskID string
	oidcIssuer := "https://issuer.example.com"
	oidcSubject := "test-subject"
	taskTitle := "test task"
	err := conn.QueryRow(ctx, `
	  INSERT INTO owners (oidc_issuer, oidc_subject)
		VALUES ($1, $2)
		RETURNING id::text
	`, oidcIssuer, oidcSubject).Scan(&ownerID)
	if err != nil {
		t.Fatal(err)
	}

	err = conn.QueryRow(ctx, `
	  INSERT INTO tasks (owner_id, title)
		VALUES ($1::uuid, $2)
		RETURNING id::text
	`, ownerID, taskTitle).Scan(&taskID)
	if err != nil {
		t.Fatal(err)
	}

	applied, err := Migrate(ctx, conn, "../../migrations", schema)
	if err != nil {
		t.Fatal(err)
	}
	if applied != 0 {
		t.Errorf("expected 0 migrations applied, got %d", applied)
	}

	var currOwnerID, currTaskID string
	err = conn.QueryRow(ctx, `
		SELECT id::text FROM owners
		WHERE oidc_issuer = $1
			AND oidc_subject = $2
	`, oidcIssuer, oidcSubject).Scan(&currOwnerID)
	if err != nil {
		t.Fatal(err)
	}
	if currOwnerID != ownerID {
		t.Errorf("expected owner ID %s, got %s", ownerID, currOwnerID)
	}

	err = conn.QueryRow(ctx, `
		SELECT id::text FROM tasks
		WHERE owner_id = $1
			AND title = $2
	`, currOwnerID, taskTitle).Scan(&currTaskID)
	if err != nil {
		t.Fatal(err)
	}
	if currTaskID != taskID {
		t.Errorf("expected task ID %s, got %s", taskID, currTaskID)
	}
}

func TestMigrationAtomicFailure(t *testing.T) {
	ctx, conn, schema := newTestDatabase(t)
	tableNames := []string{
		"rollback_probe",
		"schema_migrations",
	}
	dir := t.TempDir()
	sql := fmt.Sprintf("CREATE TABLE %s (id integer); SELECT 1 / 0;", tableNames[0])
	err := os.WriteFile(
		filepath.Join(dir, "000001_broken.up.sql"),
		[]byte(sql),
		0600,
	)
	if err != nil {
		t.Fatal(err)
	}
	_, err = Migrate(ctx, conn, dir, schema)
	if err == nil {
		t.Fatal("expected migration to fail after applying invalid SQL, but it succeeded")
	}

	for _, tableName := range tableNames {
		var exists bool
		err = conn.QueryRow(ctx, `SELECT EXISTS (
			SELECT 1 FROM information_schema.tables
			WHERE table_schema = $1
				AND table_name = $2
				AND table_type = 'BASE TABLE'
		)`, schema, tableName).Scan(&exists)
		if err != nil {
			t.Fatal(err)
		}
		if exists {
			t.Errorf("expected %s.%s table to not exist after rollback", schema, tableName)
		}
	}
}
