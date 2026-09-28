package postgres

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func TestRunnerUnitConfig(t *testing.T) {
	config, err := testConfig("postgres://workbench_test:placeholder@127.0.0.1:5433/go_workbench_test?sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}
	if config.Host != "127.0.0.1" || config.Port != 5433 || config.Database != testDatabase || len(config.Fallbacks) != 0 || config.TLSConfig != nil {
		t.Fatal("unexpected test database target or fallbacks")
	}
}

func TestMigrationRejectUnsafeTarget(t *testing.T) {
	for _, target := range []string{
		"",
		"not a URL",
		"postgres://workbench_test:placeholder@127.0.0.1:5433/production?sslmode=disable",
		"postgres://workbench_test:placeholder@example.invalid:5433/go_workbench_test?sslmode=disable",
		"postgres://workbench_test:placeholder@127.0.0.1:5432/go_workbench_test?sslmode=disable",
		"postgres://workbench_test:placeholder@127.0.0.1:5433/go_workbench_test?sslmode=disable&host=example.invalid",
		"postgres://workbench_test:placeholder@127.0.0.1:5433/go_workbench_test?sslmode=disable&service=private",
		"postgres://workbench_test:placeholder@127.0.0.1:5433/go_workbench_test?sslmode=disable&sslmode=require",
		"postgres://workbench_test@127.0.0.1:5433/go_workbench_test?sslmode=disable",
	} {
		conn, err := ConnectTestDatabase(context.Background(), target)
		if err == nil || conn != nil {
			t.Fatal("unsafe target was accepted")
		}
		if strings.Contains(err.Error(), "placeholder") || strings.Contains(err.Error(), "postgres://") {
			t.Fatal("error leaked connection input")
		}
	}
}

func TestRunnerUnitFiles(t *testing.T) {
	dir := t.TempDir()
	if _, err := readMigrations(dir); err == nil {
		t.Fatal("empty migration directory was accepted")
	}
	writeRunnerSQL(t, dir, "000002_second.up.sql", "SELECT 2;")
	writeRunnerSQL(t, dir, "000001_first.up.sql", "SELECT 1;")
	files, err := readMigrations(dir)
	if err != nil || len(files) != 2 || files[0].version != 1 || len(files[0].checksum) != 64 {
		t.Fatal("migration ordering or checksum is incorrect")
	}
	writeRunnerSQL(t, dir, "000001_duplicate.up.sql", "SELECT 3;")
	if _, err := readMigrations(dir); err == nil {
		t.Fatal("duplicate version was accepted")
	}
	for _, name := range []string{"000000_zero.up.sql", "wrong.sql", "000001_first.up.sql"} {
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			writeRunnerSQL(t, dir, name, "")
			if _, err := readMigrations(dir); err == nil {
				t.Fatal("invalid or empty migration was accepted")
			}
		})
	}
}

func TestRunnerUnitSafeErrors(t *testing.T) {
	for _, err := range []error{
		errors.New("private connection details"),
		&pgconn.PgError{Code: "23505", Message: "private values", Detail: "private values"},
	} {
		if strings.Contains(safeError("operation", err).Error(), "private") {
			t.Fatal("error leaked internal details")
		}
	}
}

func writeRunnerSQL(t *testing.T, dir, name, sql string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(sql), 0600); err != nil {
		t.Fatal("cannot write temporary runner test SQL")
	}
}

// This tests runner mechanics only; it does not implement the business schema.
func TestRunnerIntegration(t *testing.T) {
	ctx, conn, schema := newTestDatabase(t)
	dir := t.TempDir()
	writeRunnerSQL(t, dir, "000001_probe.up.sql", "CREATE TABLE runner_probe (id integer PRIMARY KEY); INSERT INTO runner_probe VALUES (1);")
	if n, err := Migrate(ctx, conn, dir, schema); err != nil || n != 1 {
		t.Fatalf("first migration applied = %d, error = %v", n, err)
	}
	if n, err := Migrate(ctx, conn, dir, schema); err != nil || n != 0 {
		t.Fatalf("repeat migration applied = %d, error = %v", n, err)
	}
	var count int
	if err := conn.QueryRow(ctx, "SELECT count(*) FROM runner_probe").Scan(&count); err != nil || count != 1 {
		t.Fatal("repeat migration did not preserve probe data")
	}
	writeRunnerSQL(t, dir, "000002_broken.up.sql", "CREATE TABLE rollback_probe (id integer); SELECT * FROM nonexistent_runner_table;")
	if _, err := Migrate(ctx, conn, dir, schema); err == nil {
		t.Fatal("invalid SQL was accepted")
	}
	var exists bool
	if err := conn.QueryRow(ctx, "SELECT to_regclass('rollback_probe') IS NOT NULL").Scan(&exists); err != nil || exists {
		t.Fatal("failed migration left a partial table")
	}
	if err := conn.QueryRow(ctx, "SELECT count(*) FROM schema_migrations").Scan(&count); err != nil || count != 1 {
		t.Fatal("failed migration was recorded as successful")
	}
	writeRunnerSQL(t, dir, "000001_probe.up.sql", "SELECT 1;")
	if _, err := Migrate(ctx, conn, dir, schema); err == nil || !strings.Contains(err.Error(), "modified") {
		t.Fatal("modified applied migration was accepted")
	}
}

func TestRunnerIntegrationConcurrent(t *testing.T) {
	ctx, first, schema := newTestDatabase(t)
	second, err := ConnectTestDatabase(ctx, os.Getenv("TEST_DATABASE_URL"))
	if err != nil {
		t.Fatal(err)
	}
	defer second.Close(context.Background())
	dir := t.TempDir()
	writeRunnerSQL(t, dir, "000001_parallel.up.sql", "CREATE TABLE parallel_probe (id integer); SELECT pg_sleep(0.05);")
	type result struct {
		applied int
		err     error
	}
	results := make(chan result, 2)
	start := make(chan struct{})
	for _, conn := range []*pgx.Conn{first, second} {
		go func() {
			<-start
			n, err := Migrate(ctx, conn, dir, schema)
			results <- result{n, err}
		}()
	}
	close(start)
	total := 0
	for range 2 {
		r := <-results
		if r.err != nil {
			t.Error(r.err)
		}
		total += r.applied
	}
	if total != 1 {
		t.Fatalf("concurrent applied total = %d, want 1", total)
	}
	var count int
	if err := first.QueryRow(ctx, "SELECT count(*) FROM schema_migrations").Scan(&count); err != nil || count != 1 {
		t.Fatal("concurrent migration history is not unique")
	}
}
