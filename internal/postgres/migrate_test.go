package postgres

import "testing"

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
	// TODO: insert a valid Owner and Task, run Migrate again and assert applied == 0.
	// TODO: assert history has one row and the original Task still exists.
	t.Fatal("RED: fill repeat migration and data preservation assertions")
}

func TestMigrationAtomicFailure(t *testing.T) {
	_, _, _ = newTestDatabase(t)
	// TODO: write a temporary .up.sql using t.TempDir and os.WriteFile.
	// TODO: CREATE TABLE followed by invalid SQL; assert Migrate fails.
	// TODO: assert the table and successful history record do not exist.
	t.Fatal("RED: fill failed migration rollback assertions")
}
