package postgres

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

const testDatabase = "go_workbench_test"

var migrationName = regexp.MustCompile(`^([0-9]{6})_[a-z][a-z0-9_]*\.up\.sql$`)
var testSchemaName = regexp.MustCompile(`^test_migration_[a-f0-9]{32}$`)

type migration struct {
	version  int64
	name     string
	sql      string
	checksum string
}

func testConfig(rawURL string) (*pgx.ConnConfig, error) {
	if rawURL == "" {
		return nil, errors.New("TEST_DATABASE_URL is required; database tests cannot be skipped")
	}
	u, err := url.Parse(rawURL)
	if err != nil || (u.Scheme != "postgres" && u.Scheme != "postgresql") || u.Opaque != "" || u.Fragment != "" {
		return nil, errors.New("invalid test database URL")
	}
	if u.Host != "127.0.0.1:5433" || u.Path != "/"+testDatabase {
		return nil, errors.New("test database target must be 127.0.0.1:5433/go_workbench_test")
	}
	if u.User == nil || u.User.Username() != "workbench_test" {
		return nil, errors.New("test database user must be workbench_test")
	}
	password, hasPassword := u.User.Password()
	if !hasPassword || password == "" {
		return nil, errors.New("test database URL requires a local password")
	}
	query, err := url.ParseQuery(u.RawQuery)
	if err != nil || len(query) != 1 || len(query["sslmode"]) != 1 || query.Get("sslmode") != "disable" {
		return nil, errors.New("test database URL only permits sslmode=disable; target overrides are forbidden")
	}
	// Prevent pgx from opening private passfiles or service files implicitly.
	query.Set("passfile", "/dev/null")
	query.Set("servicefile", "/dev/null")
	u.RawQuery = query.Encode()
	config, err := pgx.ParseConfig(u.String())
	if err != nil {
		return nil, errors.New("cannot parse test database configuration")
	}
	config.Fallbacks = nil
	config.RuntimeParams = map[string]string{"application_name": "go-workbench-migrate", "client_encoding": "UTF8"}
	return config, nil
}

func ConnectTestDatabase(ctx context.Context, rawURL string) (*pgx.Conn, error) {
	config, err := testConfig(rawURL)
	if err != nil {
		return nil, err
	}
	conn, err := pgx.ConnectConfig(ctx, config)
	if err != nil {
		return nil, safeError("connect to test database", err)
	}
	if err := checkDatabase(ctx, conn); err != nil {
		_ = conn.Close(context.Background())
		return nil, err
	}
	return conn, nil
}

func checkDatabase(ctx context.Context, conn *pgx.Conn) error {
	if conn == nil || conn.IsClosed() {
		return errors.New("test database connection is unavailable")
	}
	cfg := conn.Config()
	if cfg.Host != "127.0.0.1" || cfg.Port != 5433 || cfg.Database != testDatabase || cfg.User != "workbench_test" {
		return errors.New("unsafe database connection rejected")
	}
	var actual string
	if err := conn.QueryRow(ctx, "SELECT current_database()").Scan(&actual); err != nil {
		return safeError("verify test database", err)
	}
	if actual != testDatabase {
		return errors.New("connected database is not go_workbench_test")
	}
	return nil
}

func readMigrations(dir string) ([]migration, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, errors.New("cannot read migration directory")
	}
	var files []migration
	versions := make(map[int64]bool)
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}
		match := migrationName.FindStringSubmatch(entry.Name())
		if match == nil || entry.Type()&os.ModeSymlink != 0 {
			return nil, errors.New("migration filenames must use NNNNNN_name.up.sql and cannot be symlinks")
		}
		version, _ := strconv.ParseInt(match[1], 10, 64)
		if version == 0 || versions[version] {
			return nil, errors.New("migration versions must be positive and unique")
		}
		versions[version] = true
		data, err := os.ReadFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			return nil, errors.New("cannot read migration SQL")
		}
		if strings.TrimSpace(string(data)) == "" {
			return nil, errors.New("migration SQL cannot be empty")
		}
		files = append(files, migration{version, entry.Name(), string(data), fmt.Sprintf("%x", sha256.Sum256(data))})
	}
	if len(files) == 0 {
		return nil, errors.New("no .up.sql migrations found; write migrations/000001_initial.up.sql first")
	}
	sort.Slice(files, func(i, j int) bool { return files[i].version < files[j].version })
	return files, nil
}

// Migrate applies all pending versions in one transaction, including history.
// SQL files are trusted repository code and must not manage transactions themselves.
func Migrate(ctx context.Context, conn *pgx.Conn, dir, schema string) (int, error) {
	if schema != "public" && !testSchemaName.MatchString(schema) {
		return 0, errors.New("migration schema must be public or an isolated test_migration schema")
	}
	files, err := readMigrations(dir)
	if err != nil {
		return 0, err
	}
	if err := checkDatabase(ctx, conn); err != nil {
		return 0, err
	}
	tx, err := conn.Begin(ctx)
	if err != nil {
		return 0, safeError("begin migration", err)
	}
	defer tx.Rollback(context.Background())
	if _, err := tx.Exec(ctx, "SELECT pg_advisory_xact_lock(hashtext(current_database()), hashtext($1))", schema); err != nil {
		return 0, safeError("lock migrations", err)
	}
	if _, err := tx.Exec(ctx, "SELECT set_config('search_path', $1, true)", pgx.Identifier{schema}.Sanitize()); err != nil {
		return 0, safeError("set migration schema", err)
	}
	if _, err := tx.Exec(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations (
		version bigint PRIMARY KEY,
		name text NOT NULL,
		checksum text NOT NULL CHECK (length(checksum) = 64),
		applied_at timestamptz NOT NULL DEFAULT now()
	)`); err != nil {
		return 0, safeError("create migration history", err)
	}
	rows, err := tx.Query(ctx, "SELECT version, name, checksum FROM schema_migrations ORDER BY version")
	if err != nil {
		return 0, safeError("read migration history", err)
	}
	previous := make(map[int64]migration)
	var latest int64
	for rows.Next() {
		var m migration
		if err := rows.Scan(&m.version, &m.name, &m.checksum); err != nil {
			rows.Close()
			return 0, safeError("decode migration history", err)
		}
		previous[m.version] = m
		latest = m.version
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return 0, safeError("read migration history", err)
	}
	for _, file := range files {
		old, applied := previous[file.version]
		if applied {
			if old.name != file.name || old.checksum != file.checksum {
				return 0, fmt.Errorf("applied migration %06d was modified", file.version)
			}
			delete(previous, file.version)
		} else if file.version <= latest {
			return 0, errors.New("new migrations must follow the latest applied version")
		}
	}
	if len(previous) != 0 {
		return 0, errors.New("an applied migration file is missing")
	}
	applied := 0
	for _, file := range files {
		if file.version <= latest {
			continue
		}
		if _, err := tx.Exec(ctx, file.sql); err != nil {
			return 0, safeError(fmt.Sprintf("apply migration %06d", file.version), err)
		}
		if _, err := tx.Exec(ctx, "INSERT INTO schema_migrations (version, name, checksum) VALUES ($1, $2, $3)", file.version, file.name, file.checksum); err != nil {
			return 0, safeError("record migration version", err)
		}
		applied++
	}
	if err := tx.Commit(ctx); err != nil {
		return 0, safeError("commit migrations", err)
	}
	return applied, nil
}

func safeError(operation string, err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return fmt.Errorf("%s failed (SQLSTATE %s)", operation, pgErr.Code)
	}
	return fmt.Errorf("%s failed; check local configuration and database availability", operation)
}
