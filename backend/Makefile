GO            ?= go
GOLANGCI_LINT ?= golangci-lint
# Pinned scanner; CI and local runs use the same version (bump via PR, two-week rule).
GOVULNCHECK   := golang.org/x/vuln/cmd/govulncheck@v1.8.0
PKGS          := ./...
GEN_DIR       := adapter/http/gen
GEN_FILES     := $(GEN_DIR)/*.gen.go
# Extra `go test` flags, e.g. GOTEST_FLAGS="-covermode=atomic -coverprofile=coverage.out" in CI.
GOTEST_FLAGS  ?=
# Coverage gate (P0-19): business logic only; generated code and commands are left out.
COVER_PROFILE ?= coverage.out
COVER_MIN     ?= 80
COVER_EXCLUDE ?= /adapter/http/gen/|/cmd/

.DEFAULT_GOAL := check

.PHONY: check fmt fmt-check lint test test-integration cover cover-check vuln tidy generate codegen-check db-up db-down db-reset db-roles-check migrate-up migrate-down migrate-status flags

## check: fmt-check + lint + unit tests (what CI runs)
check: fmt-check lint test

## fmt: format code (gofmt + goimports, as configured in .golangci.yml)
fmt:
	$(GOLANGCI_LINT) fmt

## fmt-check: fail if any file needs gofmt
fmt-check:
	@out="$$(gofmt -l .)"; if [ -n "$$out" ]; then echo "gofmt needed:"; echo "$$out"; exit 1; fi

## lint: golangci-lint v2 (depguard Dependency Rule, forbidigo float/time.Now bans, gosec)
lint:
	$(GOLANGCI_LINT) run $(PKGS)

## test: unit tests, including the internal/archtest Dependency Rule check
test:
	$(GO) test -race -count=1 $(GOTEST_FLAGS) $(PKGS)

## test-integration: tests behind the `integration` build tag (real PostgreSQL via testcontainers-go; needs Docker)
test-integration:
	$(GO) test -race -count=1 -tags=integration $(GOTEST_FLAGS) $(PKGS)

## cover: unit tests with a coverage summary
cover:
	$(GO) test -count=1 -coverprofile=coverage.out $(PKGS)
	$(GO) tool cover -func=coverage.out | tail -1

## cover-check: fail if COVER_PROFILE's business-logic coverage (COVER_EXCLUDE left out) is below COVER_MIN%
cover-check:
	$(GO) run ./cmd/covergate -profile '$(COVER_PROFILE)' -min $(COVER_MIN) -exclude '$(COVER_EXCLUDE)'

## vuln: govulncheck (pinned) over every package; fails on vulnerabilities reachable from our code
vuln:
	$(GO) run $(GOVULNCHECK) $(PKGS)

## tidy: go mod tidy
tidy:
	$(GO) mod tidy

## generate: regenerate adapter/http/gen from ../docs/api/openapi.yaml (oapi-codegen strict std-http server + flag map)
generate:
	$(GO) generate ./$(GEN_DIR)/...

## codegen-check: regenerate and fail if any *.gen.go differs from what is committed (drift check; hand-written files in GEN_DIR are not judged)
codegen-check: generate
	@out="$$(git status --porcelain -- '$(GEN_FILES)')"; if [ -n "$$out" ]; then \
		echo "adapter/http/gen is out of date with docs/api/openapi.yaml; run 'make generate' and commit:"; \
		echo "$$out"; git --no-pager diff --stat -- '$(GEN_FILES)'; exit 1; fi

# ---- Local stack (LOCAL ONLY; infra/local) ---------------------------------------------

LOCAL_DIR    := ../infra/local
LOCAL_ENV    := $(LOCAL_DIR)/.env
COMPOSE      := docker compose -f $(LOCAL_DIR)/compose.yaml --env-file $(LOCAL_ENV)
ASSERT_LOCAL := $(LOCAL_DIR)/scripts/assert-local.sh $(LOCAL_ENV)
RUNTIME_ROLES := cc_app cc_webhook cc_worker

$(LOCAL_ENV):
	cp $(LOCAL_DIR)/.env.example $@
	@echo "created $@ from .env.example (local-only placeholder passwords)"

## db-up: start local postgres 18 (loopback only) and wait until healthy
db-up: $(LOCAL_ENV)
	@$(ASSERT_LOCAL)
	$(COMPOSE) up -d --wait postgres

## db-down: stop the local stack, keeping its data volume
db-down: $(LOCAL_ENV)
	$(COMPOSE) down

## db-reset: LOCAL ONLY. Destroy the data volume and re-bootstrap roles from scratch
db-reset: $(LOCAL_ENV)
	@$(ASSERT_LOCAL)
	$(COMPOSE) down --volumes
	$(COMPOSE) up -d --wait postgres

## db-roles-check: connect over TCP as each runtime role and assert NOSUPERUSER/NOBYPASSRLS/not owner
db-roles-check: $(LOCAL_ENV)
	@set -e; . $(LOCAL_ENV); for role in $(RUNTIME_ROLES); do \
	  case $$role in cc_app) pw=$$CC_APP_PASSWORD;; cc_webhook) pw=$$CC_WEBHOOK_PASSWORD;; cc_worker) pw=$$CC_WORKER_PASSWORD;; esac; \
	  $(COMPOSE) exec -T -e PGPASSWORD="$$pw" postgres \
	    psql --no-psqlrc -h 127.0.0.1 -U $$role -d "$${CC_DB_NAME:-control_contable}" -f /bootstrap/check-runtime-role.sql; \
	done

# Owner URL for cmd/migrate on the LOCAL stack only, built from infra/local/.env inside the
# recipe so the password never appears in make's output. Shared environments set
# MIGRATE_DATABASE_URL from the secret manager instead.
MIGRATE = set -e; . $(LOCAL_ENV); \
	MIGRATE_DATABASE_URL="postgres://cc_owner:$${CC_OWNER_PASSWORD}@$${CC_DB_HOST}:$${CC_DB_PORT}/$${CC_DB_NAME}?sslmode=disable" \
	$(GO) run ./cmd/migrate

## migrate-up: LOCAL ONLY. Apply every pending goose migration as cc_owner
migrate-up: $(LOCAL_ENV)
	@$(ASSERT_LOCAL)
	@$(MIGRATE) up

## migrate-down: LOCAL ONLY. Roll back the most recent migration as cc_owner
migrate-down: $(LOCAL_ENV)
	@$(ASSERT_LOCAL)
	@$(MIGRATE) down

## migrate-status: LOCAL ONLY. List migrations as applied or pending
migrate-status: $(LOCAL_ENV)
	@$(ASSERT_LOCAL)
	@$(MIGRATE) status

# Owner URL for cmd/flags on the LOCAL stack only (same construction as MIGRATE). Shared
# environments set DATABASE_URL_OWNER from the secret manager instead.
FLAGS = set -e; . $(LOCAL_ENV); \
	DATABASE_URL_OWNER="postgres://cc_owner:$${CC_OWNER_PASSWORD}@$${CC_DB_HOST}:$${CC_DB_PORT}/$${CC_DB_NAME}?sslmode=disable" \
	$(GO) run ./cmd/flags

## flags: LOCAL ONLY. Module flag admin as cc_owner, e.g. make flags ARGS="list" | ARGS="enable module_auth" | ARGS="apply-preset demo_core"
flags: $(LOCAL_ENV)
	@$(ASSERT_LOCAL)
	@$(FLAGS) $(or $(ARGS),list)
