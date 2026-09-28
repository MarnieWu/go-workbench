package postgres

import "testing"

func TestMigrationEmptySchema(t *testing.T) {
	ctx, conn, schema := newTestDatabase(t)
	applyBusinessMigration(t, ctx, conn, schema)
	// TODO: query information_schema for all ten tables, columns and named constraints.
	// TODO: assert schema_migrations contains exactly version 1.
	t.Fatal("RED: fill the initial business schema assertions")
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
