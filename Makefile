# Root Makefile: repository-wide targets. Service targets live in backend/Makefile (P0-03).

CONTRACT_TOOLS := tools/contract
REDOCLY := $(CONTRACT_TOOLS)/node_modules/.bin/redocly
REDOCLY_ENV := REDOCLY_SUPPRESS_UPDATE_NOTICE=true REDOCLY_TELEMETRY=off

.PHONY: contract-tools contract-lint contract-lint-test

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
