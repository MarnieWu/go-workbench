TEST_ENV_FILE := deploy/test.env
GO_CACHE := /tmp/go-workbench-go-cache

-include $(TEST_ENV_FILE)
export

.PHONY: help check-test-env check-database-url check-allow-write \
	test-db-config test-db-up test-db-down test-migrate test-seed test-api test \
	api migrate seed test-cmd-api test-cmd-seed

help:
	@printf '%s\n' 'Test/local database targets:'
	@printf '%s\n' '  make test-db-config  Validate the local test Docker Compose config'
	@printf '%s\n' '  make test-db-up      Start the local test PostgreSQL container'
	@printf '%s\n' '  make test-db-down    Stop the local test PostgreSQL container'
	@printf '%s\n' '  make test-migrate    Apply migrations to TEST_DATABASE_URL'
	@printf '%s\n' '  make test-seed       Insert browser-verification data into TEST_DATABASE_URL'
	@printf '%s\n' '  make test-api        Start the API with DATABASE_URL=TEST_DATABASE_URL'
	@printf '%s\n' '  make test            Run the Go test suite with TEST_DATABASE_URL'
	@printf '%s\n' ''
	@printf '%s\n' 'DATABASE_URL targets:'
	@printf '%s\n' '  make api             Start the API with DATABASE_URL'
	@printf '%s\n' '  make migrate         Apply migrations to DATABASE_URL. Requires ALLOW_WRITE=1'
	@printf '%s\n' '  make seed            Insert seed data into DATABASE_URL. Requires ALLOW_WRITE=1'

check-test-env:
	@test -f "$(TEST_ENV_FILE)" || { \
		printf '%s\n' 'Missing deploy/test.env. Copy deploy/test.env.example and fill local-only values.'; \
		exit 1; \
	}
	@test -n "$(TEST_POSTGRES_PASSWORD)" || { \
		printf '%s\n' 'TEST_POSTGRES_PASSWORD is required in deploy/test.env.'; \
		exit 1; \
	}
	@test -n "$(TEST_DATABASE_URL)" || { \
		printf '%s\n' 'TEST_DATABASE_URL is required in deploy/test.env.'; \
		exit 1; \
	}

check-database-url:
	@test -n "$(DATABASE_URL)" || { \
		printf '%s\n' 'DATABASE_URL is required.'; \
		exit 1; \
	}

check-allow-write:
	@test "$(ALLOW_WRITE)" = "1" || { \
		printf '%s\n' 'Refusing DATABASE_URL write. Re-run with ALLOW_WRITE=1 after confirming the target.'; \
		exit 1; \
	}

test-db-config: check-test-env
	@docker compose --env-file $(TEST_ENV_FILE) -f deploy/compose.yaml config >/dev/null
	@printf '%s\n' 'Docker Compose config is valid.'

test-db-up: check-test-env
	@docker compose --env-file $(TEST_ENV_FILE) -f deploy/compose.yaml up -d db

test-db-down: check-test-env
	@docker compose --env-file $(TEST_ENV_FILE) -f deploy/compose.yaml down

test-migrate: check-test-env
	@GOCACHE=$(GO_CACHE) TEST_DATABASE_URL="$(TEST_DATABASE_URL)" go run ./cmd/migrate up

test-seed: check-test-env
	@GOCACHE=$(GO_CACHE) DATABASE_URL="$(TEST_DATABASE_URL)" go run ./cmd/seed

test-api: check-test-env
	@GOCACHE=$(GO_CACHE) DATABASE_URL="$(TEST_DATABASE_URL)" go run ./cmd/api

test: check-test-env
	@GOCACHE=$(GO_CACHE) TEST_DATABASE_URL="$(TEST_DATABASE_URL)" go test ./... -count=1

api: check-database-url
	@GOCACHE=$(GO_CACHE) go run ./cmd/api

migrate: check-database-url check-allow-write
	@GOCACHE=$(GO_CACHE) TEST_DATABASE_URL="$(DATABASE_URL)" go run ./cmd/migrate up

seed: check-database-url check-allow-write
	@GOCACHE=$(GO_CACHE) go run ./cmd/seed

test-cmd-api:
	@GOCACHE=$(GO_CACHE) go test ./cmd/api -count=1

test-cmd-seed:
	@GOCACHE=$(GO_CACHE) go test ./cmd/seed -count=1
