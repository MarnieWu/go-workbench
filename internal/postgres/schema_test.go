package postgres

import "testing"

func TestSchemaInvalidTaskStatus(t *testing.T) {
	ctx, conn, schema := newTestDatabase(t)
	applyBusinessMigration(t, ctx, conn, schema)
	// TODO: insert a valid Owner; insert a Task with status 'unknown'.
	// TODO: assertPgError(t, err, "23514", yourConstraintName); assert no invalid Task.
	t.Fatal("RED: fill invalid Task status assertions")
}

func TestSchemaDuplicateCaptureKey(t *testing.T) {
	ctx, conn, schema := newTestDatabase(t)
	applyBusinessMigration(t, ctx, conn, schema)
	// TODO: same owner and same key twice; assert 23505 and the named constraint.
	// TODO: assert only the first Capture remains.
	t.Fatal("RED: fill duplicate Capture key assertions")
}

func TestSchemaCaptureKeyAcrossOwners(t *testing.T) {
	ctx, conn, schema := newTestDatabase(t)
	applyBusinessMigration(t, ctx, conn, schema)
	// TODO: two Owners, one Capture each using the same key; assert both exist.
	t.Fatal("RED: fill cross-owner Capture key assertions")
}

func TestSchemaMissingOwner(t *testing.T) {
	ctx, conn, schema := newTestDatabase(t)
	applyBusinessMigration(t, ctx, conn, schema)
	// TODO: otherwise valid Task with nonexistent owner; assert 23503 and no row.
	t.Fatal("RED: fill missing Owner assertions")
}

func TestSchemaCrossOwnerProject(t *testing.T) {
	ctx, conn, schema := newTestDatabase(t)
	applyBusinessMigration(t, ctx, conn, schema)
	// TODO: Owner A Task referencing Owner B Project; assert 23503 and no Task.
	t.Fatal("RED: fill cross-owner Project assertions")
}

func TestSchemaDuplicateEvidenceLink(t *testing.T) {
	ctx, conn, schema := newTestDatabase(t)
	applyBusinessMigration(t, ctx, conn, schema)
	// TODO: valid Task and Evidence; link twice; assert 23505 and one link.
	t.Fatal("RED: fill duplicate evidence link assertions")
}
