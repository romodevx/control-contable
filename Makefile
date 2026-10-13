# Root Makefile: repository-wide targets. Service targets live in backend/Makefile (P0-03).

CONTRACT_TOOLS := tools/contract
REDOCLY := $(CONTRACT_TOOLS)/node_modules/.bin/redocly
REDOCLY_ENV := REDOCLY_SUPPRESS_UPDATE_NOTICE=true REDOCLY_TELEMETRY=off
PRISM := $(CONTRACT_TOOLS)/node_modules/.bin/prism
OPENAPI_SPEC := docs/api/openapi.yaml

# Mock server (docs/runbooks/mock-server.md). Port 3001 is reserved for control-contable's
# mock in the project registry (API 3000, db 5434). MOCK_DYNAMIC=true generates bodies from
# the schemas instead of returning the contract's static examples.
MOCK_HOST ?= 127.0.0.1
MOCK_PORT ?= 3001
MOCK_DYNAMIC ?= false

.PHONY: contract-tools contract-lint contract-lint-test mock

$(REDOCLY): $(CONTRACT_TOOLS)/package.json $(CONTRACT_TOOLS)/package-lock.json
	npm ci --prefix $(CONTRACT_TOOLS) --ignore-scripts --no-audit --no-fund
	@touch $@

## contract-tools: install the pinned contract tooling (Redocly CLI)
contract-tools: $(REDOCLY)

## contract-lint: lint docs/api/openapi.yaml against redocly.yaml (see docs/api/CONTRACT.md)
contract-lint: $(REDOCLY)
	$(REDOCLY_ENV) $(REDOCLY) lint --config redocly.yaml

## contract-lint-test: prove every contract rule rejects what it should
contract-lint-test: $(REDOCLY)
	node --test $(CONTRACT_TOOLS)/test.js

## mock: serve docs/api/openapi.yaml with Prism on MOCK_HOST:MOCK_PORT (MOCK_DYNAMIC=true for generated bodies)
mock: $(REDOCLY)
	$(PRISM) mock $(OPENAPI_SPEC) --host $(MOCK_HOST) --port $(MOCK_PORT) --dynamic=$(MOCK_DYNAMIC) --json-schema-faker-fillProperties=false --errors=false
