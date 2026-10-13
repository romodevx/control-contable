GO            ?= go
GOLANGCI_LINT ?= golangci-lint
PKGS          := ./...
GEN_DIR       := adapter/http/gen

.DEFAULT_GOAL := check

.PHONY: check fmt fmt-check lint test test-integration cover tidy generate codegen-check

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
	$(GO) test -race -count=1 $(PKGS)

## test-integration: tests behind the `integration` build tag (real PostgreSQL via testcontainers-go; needs Docker)
test-integration:
	$(GO) test -race -count=1 -tags=integration $(PKGS)

## cover: unit tests with a coverage summary
cover:
	$(GO) test -count=1 -coverprofile=coverage.out $(PKGS)
	$(GO) tool cover -func=coverage.out | tail -1

## tidy: go mod tidy
tidy:
	$(GO) mod tidy

## generate: regenerate adapter/http/gen from ../docs/api/openapi.yaml (oapi-codegen strict std-http server + flag map)
generate:
	$(GO) generate ./$(GEN_DIR)/...

## codegen-check: regenerate and fail if the result differs from what is committed (drift check)
codegen-check: generate
	@out="$$(git status --porcelain -- $(GEN_DIR))"; if [ -n "$$out" ]; then \
		echo "adapter/http/gen is out of date with docs/api/openapi.yaml; run 'make generate' and commit:"; \
		echo "$$out"; git --no-pager diff --stat -- $(GEN_DIR); exit 1; fi
