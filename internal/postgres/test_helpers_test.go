package postgres

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func newTestDatabase(t *testing.T) (context.Context, *pgx.Conn, string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	t.Cleanup(cancel)
	rawURL := os.Getenv("TEST_DATABASE_URL")
	conn, err := ConnectTestDatabase(ctx, rawURL)
	if err != nil {
		t.Fatal(err)
	}
	schema := fmt.Sprintf("test_migration_%x", randomSchemaID(t))
	identifier := pgx.Identifier{schema}.Sanitize()
	if _, err := conn.Exec(ctx, "CREATE SCHEMA "+identifier); err != nil {
		_ = conn.Close(context.Background())
		t.Fatal(safeError("create isolated schema", err))
	}
	t.Cleanup(func() {
		cleanupCtx, stop := context.WithTimeout(context.Background(), 5*time.Second)
		defer stop()
		_ = conn.Close(cleanupCtx)
		cleanup, err := ConnectTestDatabase(cleanupCtx, rawURL)
		if err != nil {
			t.Error(err)
			return
		}
		defer cleanup.Close(cleanupCtx)
		if !testSchemaName.MatchString(schema) {
			t.Error("refusing to clean an unrecognized test schema")
			return
		}
		if _, err := cleanup.Exec(cleanupCtx, "DROP SCHEMA "+identifier+" CASCADE"); err != nil {
			t.Error(safeError("clean isolated schema", err))
		}
	})
	if _, err := conn.Exec(ctx, "SELECT set_config('search_path', $1, false)", identifier); err != nil {
		t.Fatal(safeError("select isolated schema", err))
	}
	return ctx, conn, schema
}

func randomSchemaID(t *testing.T) [16]byte {
	t.Helper()
	var id [16]byte
	if _, err := rand.Read(id[:]); err != nil {
		t.Fatal("cannot generate isolated schema name")
	}
	return id
}

func applyBusinessMigration(t *testing.T, ctx context.Context, conn *pgx.Conn, schema string) {
	t.Helper()
	if _, err := Migrate(ctx, conn, "../../migrations", schema); err != nil {
		t.Fatal(err)
	}
}

func assertPgError(t *testing.T, err error, code, constraint string) {
	t.Helper()
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		t.Fatal("expected a PostgreSQL constraint error")
	}
	if pgErr.Code != code || pgErr.ConstraintName != constraint {
		t.Fatalf("SQLSTATE/constraint = %s/%s, want %s/%s", pgErr.Code, pgErr.ConstraintName, code, constraint)
	}
}
