TEST_ENV_FILE := deploy/test.env
GO_CACHE := /tmp/go-workbench-go-cache

.PHONY: help check-test-env-file check-test-database-url check-database-url check-allow-write \
	test-task test-capture test-candidate \
	test-httpapi test-cmd-api test-cmd-seed test-go-core test-backend \
	test-postgres-migration test-postgres-schema test-postgres-task test-postgres-capture test-postgres-candidate test-postgres \
	test-web test-frontend test-all \
	test-db-config test-db-up test-db-down test \
	run-api-test-db run-migrate-test-db run-seed-test-db \
	run-api run-migrate run-seed run-frontend api migrate seed

help:
	@printf '%s\n' 'Business-block test targets:'
	@printf '%s\n' '  make test-task                 Run Task domain tests'
	@printf '%s\n' '  make test-capture              Run Capture domain and HTTP API tests'
	@printf '%s\n' '  make test-candidate            Run Candidate/Inbox domain and HTTP API tests'
	@printf '%s\n' '  make test-httpapi              Run HTTP API tests'
	@printf '%s\n' '  make test-cmd-api              Run API command tests'
	@printf '%s\n' '  make test-go-core              Run non-database Go package tests'
	@printf '%s\n' '  make test-backend              Run all Go backend tests; requires exported TEST_DATABASE_URL'
	@printf '%s\n' '  make test-frontend             Run Web lint and typecheck'
	@printf '%s\n' '  make test-all                  Run backend and frontend tests'
	@printf '%s\n' '  make test                      Alias for test-all'
	@printf '%s\n' ''
	@printf '%s\n' 'PostgreSQL test targets:'
	@printf '%s\n' '  make test-postgres-migration   Run migration tests; requires exported TEST_DATABASE_URL'
	@printf '%s\n' '  make test-postgres-schema      Run schema tests; requires exported TEST_DATABASE_URL'
	@printf '%s\n' '  make test-postgres-task        Run Task repository tests; requires exported TEST_DATABASE_URL'
	@printf '%s\n' '  make test-postgres-capture     Run Capture repository tests; requires exported TEST_DATABASE_URL'
	@printf '%s\n' '  make test-postgres-candidate   Run Candidate repository tests; requires exported TEST_DATABASE_URL'
	@printf '%s\n' '  make test-postgres             Run all PostgreSQL package tests; requires exported TEST_DATABASE_URL'
	@printf '%s\n' ''
	@printf '%s\n' 'Test/local database targets:'
	@printf '%s\n' '  make test-db-config  Validate the local test Docker Compose config'
	@printf '%s\n' '  make test-db-up      Start the local test PostgreSQL container'
	@printf '%s\n' '  make test-db-down    Stop the local test PostgreSQL container'
	@printf '%s\n' '  make run-migrate-test-db  Apply migrations to TEST_DATABASE_URL'
	@printf '%s\n' '  make run-seed-test-db     Insert browser-verification data into TEST_DATABASE_URL'
	@printf '%s\n' '  make run-api-test-db      Start the API with DATABASE_URL=TEST_DATABASE_URL'
	@printf '%s\n' ''
	@printf '%s\n' 'Run targets using DATABASE_URL:'
	@printf '%s\n' '  make run-api         Start the API with DATABASE_URL'
	@printf '%s\n' '  make run-migrate     Apply migrations to DATABASE_URL. Requires ALLOW_WRITE=1'
	@printf '%s\n' '  make run-seed        Insert seed data into DATABASE_URL. Requires ALLOW_WRITE=1'
	@printf '%s\n' '  make run-frontend    Start the frontend dev server'

check-test-env-file:
	@test -f "$(TEST_ENV_FILE)" || { \
		printf '%s\n' 'Missing deploy/test.env. Copy deploy/test.env.example and fill local-only values.'; \
		exit 1; \
	}

check-test-database-url:
	@test -n "$(TEST_DATABASE_URL)" || { \
		printf '%s\n' 'TEST_DATABASE_URL is required.'; \
		printf '%s\n' 'Start the local test database with `make test-db-up`, then export TEST_DATABASE_URL from your local-only test config before rerunning this target.'; \
		printf '%s\n' 'Do not paste or commit the database URL.'; \
		exit 1; \
	}

check-database-url:
	@test -n "$(DATABASE_URL)" || { \
		printf '%s\n' 'DATABASE_URL is required.'; \
		printf '%s\n' 'For local API development, export DATABASE_URL to a local/test database URL, then rerun this target.'; \
		printf '%s\n' 'If you want the managed test database flow, use `make run-api-test-db` with exported TEST_DATABASE_URL instead.'; \
		printf '%s\n' 'Do not paste or commit the database URL.'; \
		exit 1; \
	}

check-allow-write:
	@test "$(ALLOW_WRITE)" = "1" || { \
		printf '%s\n' 'Refusing DATABASE_URL write. Re-run with ALLOW_WRITE=1 after confirming the target.'; \
		exit 1; \
	}

test-db-config: check-test-env-file
	@docker compose --env-file $(TEST_ENV_FILE) -f deploy/compose.yaml config >/dev/null
	@printf '%s\n' 'Docker Compose config is valid.'

test-db-up: check-test-env-file
	@docker compose --env-file $(TEST_ENV_FILE) -f deploy/compose.yaml up -d db

test-db-down: check-test-env-file
	@docker compose --env-file $(TEST_ENV_FILE) -f deploy/compose.yaml down

run-migrate-test-db: check-test-database-url
	@GOCACHE=$(GO_CACHE) TEST_DATABASE_URL="$(TEST_DATABASE_URL)" go run ./cmd/migrate up

run-seed-test-db: check-test-database-url
	@GOCACHE=$(GO_CACHE) DATABASE_URL="$(TEST_DATABASE_URL)" go run ./cmd/seed

run-api-test-db: check-test-database-url
	@GOCACHE=$(GO_CACHE) DATABASE_URL="$(TEST_DATABASE_URL)" go run ./cmd/api

test-backend: check-test-database-url
	@GOCACHE=$(GO_CACHE) TEST_DATABASE_URL="$(TEST_DATABASE_URL)" go test ./... -count=1

test-all: test-backend test-frontend

test: test-all

run-api: check-database-url
	@GOCACHE=$(GO_CACHE) go run ./cmd/api

run-migrate: check-database-url check-allow-write
	@GOCACHE=$(GO_CACHE) TEST_DATABASE_URL="$(DATABASE_URL)" go run ./cmd/migrate up

run-seed: check-database-url check-allow-write
	@GOCACHE=$(GO_CACHE) go run ./cmd/seed

run-frontend:
	@npm --prefix web run dev

api: run-api

migrate: run-migrate

seed: run-seed

test-cmd-api:
	@GOCACHE=$(GO_CACHE) go test ./cmd/api -count=1

test-cmd-seed:
	@GOCACHE=$(GO_CACHE) go test ./cmd/seed -count=1

test-task:
	@GOCACHE=$(GO_CACHE) go test ./internal/task -count=1

test-capture:
	@GOCACHE=$(GO_CACHE) go test ./internal/capture -count=1
	@GOCACHE=$(GO_CACHE) go test ./internal/httpapi -run '^TestRouterCreateCapture' -count=1

test-candidate:
	@GOCACHE=$(GO_CACHE) go test ./internal/candidate -count=1
	@GOCACHE=$(GO_CACHE) go test ./internal/httpapi -run '^TestRouterListInbox' -count=1

test-httpapi:
	@GOCACHE=$(GO_CACHE) go test ./internal/httpapi -count=1

test-go-core: test-task test-capture test-candidate test-httpapi test-cmd-api test-cmd-seed

test-postgres-migration: check-test-database-url
	@GOCACHE=$(GO_CACHE) TEST_DATABASE_URL="$(TEST_DATABASE_URL)" go test ./internal/postgres -run '^TestMigration|^TestRunner' -count=1

test-postgres-schema: check-test-database-url
	@GOCACHE=$(GO_CACHE) TEST_DATABASE_URL="$(TEST_DATABASE_URL)" go test ./internal/postgres -run '^TestSchema' -count=1

test-postgres-task: check-test-database-url
	@GOCACHE=$(GO_CACHE) TEST_DATABASE_URL="$(TEST_DATABASE_URL)" go test ./internal/postgres -run '^TestTaskRepositoryList$$' -count=1

test-postgres-capture: check-test-database-url
	@GOCACHE=$(GO_CACHE) TEST_DATABASE_URL="$(TEST_DATABASE_URL)" go test ./internal/postgres -run '^TestCaptureRepositoryCreate' -count=1

test-postgres-candidate: check-test-database-url
	@GOCACHE=$(GO_CACHE) TEST_DATABASE_URL="$(TEST_DATABASE_URL)" go test ./internal/postgres -run '^TestCandidateRepositoryListPending' -count=1

test-postgres: check-test-database-url
	@GOCACHE=$(GO_CACHE) TEST_DATABASE_URL="$(TEST_DATABASE_URL)" go test ./internal/postgres -count=1

test-frontend:
	@npm --prefix web run lint
	@npm --prefix web run typecheck

test-web: test-frontend
