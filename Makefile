# Root Makefile: repository-wide targets. Service targets live in backend/Makefile (P0-03).

CONTRACT_TOOLS := tools/contract
# npm rewrites this file on every install, so it marks "tooling matches the lockfile".
CONTRACT_TOOLS_INSTALLED := $(CONTRACT_TOOLS)/node_modules/.package-lock.json
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
# fillProperties=false: dynamic mode does not pad objects with filler keys to reach minProperties.
# errors=false: request violations are reported in the response, the mock keeps serving.
PRISM_MOCK_FLAGS := --host $(MOCK_HOST) --port $(MOCK_PORT) --dynamic=$(MOCK_DYNAMIC) \
	--json-schema-faker-fillProperties=false --errors=false

# Generated API reference (P0-32): static HTML rendered from the contract, never written by
# hand. build/ is git-ignored: publish the file, do not commit it.
API_DOCS_OUT ?= build/api-docs/index.html

# Generated error tables (M1-18): each module page's error table is rendered from the
# contract between BEGIN/END GENERATED markers. Add one <module>=<page> pair per module page.
API_BUNDLE := build/openapi/bundle.json
ERROR_TABLE_PAGES := module_auth=docs/modules/auth.md module_taxpayers=docs/modules/taxpayers.md \
	module_incomes=docs/modules/incomes.md module_expenses=docs/modules/expenses.md \
	module_tax_engine=docs/modules/tax-engine.md module_dashboard=docs/modules/dashboard.md \
	module_history=docs/modules/history.md
ERROR_TABLE := node $(CONTRACT_TOOLS)/error-table.js $(API_BUNDLE)

.PHONY: contract-tools contract-lint contract-lint-test mock api-docs api-bundle \
	api-error-tables api-error-tables-check

$(CONTRACT_TOOLS_INSTALLED): $(CONTRACT_TOOLS)/package.json $(CONTRACT_TOOLS)/package-lock.json
	npm ci --prefix $(CONTRACT_TOOLS) --ignore-scripts --no-audit --no-fund
	@touch $@

## contract-tools: install the pinned contract tooling (Redocly CLI, Prism)
contract-tools: $(CONTRACT_TOOLS_INSTALLED)

## contract-lint: lint docs/api/openapi.yaml against redocly.yaml (see docs/api/CONTRACT.md)
contract-lint: $(CONTRACT_TOOLS_INSTALLED)
	$(REDOCLY_ENV) $(REDOCLY) lint --config redocly.yaml

## contract-lint-test: prove every contract rule rejects what it should
contract-lint-test: $(CONTRACT_TOOLS_INSTALLED)
	node --test $(CONTRACT_TOOLS)/test.js $(CONTRACT_TOOLS)/error-table.test.js

## api-docs: render docs/api/openapi.yaml to one static HTML page (Redoc) at API_DOCS_OUT
api-docs: $(CONTRACT_TOOLS_INSTALLED)
	$(REDOCLY_ENV) $(REDOCLY) build-docs $(OPENAPI_SPEC) --config redocly.yaml \
		--output $(API_DOCS_OUT) --title "Control Contable API" --disableGoogleFont

## api-bundle: dereferenced JSON bundle of the contract at API_BUNDLE (input of the generators)
api-bundle: $(CONTRACT_TOOLS_INSTALLED)
	$(REDOCLY_ENV) $(REDOCLY) bundle $(OPENAPI_SPEC) --config redocly.yaml --dereferenced \
		--ext json --output $(API_BUNDLE)

## api-error-tables: regenerate the error tables of the module pages (docs/modules/) from the contract
api-error-tables: api-bundle
	$(ERROR_TABLE) $(ERROR_TABLE_PAGES)

## api-error-tables-check: fail if a module page's error table differs from the contract (CI)
api-error-tables-check: api-bundle
	$(ERROR_TABLE) --check $(ERROR_TABLE_PAGES)

## mock: serve docs/api/openapi.yaml with Prism on MOCK_HOST:MOCK_PORT (MOCK_DYNAMIC=true for generated bodies)
mock: $(CONTRACT_TOOLS_INSTALLED)
	$(PRISM) mock $(OPENAPI_SPEC) $(PRISM_MOCK_FLAGS)

# --- Mobile (Flutter, mobile/; see mobile/README.md) -------------------------------------------
MOBILE := mobile
FLAVOR ?= dev
# Required for staging/prod (HTTPS); optional for dev (e.g. http://10.0.2.2:3000 on an emulator).
API_BASE_URL ?=
MOBILE_DEFINES := $(if $(API_BASE_URL),--dart-define=API_BASE_URL=$(API_BASE_URL))
# Symbols stay on the build machine (uploaded to crash reporting later); never inside the artifact.
MOBILE_SYMBOLS := build/symbols/$(FLAVOR)
MOBILE_RELEASE_FLAGS := --release --flavor $(FLAVOR) -t lib/main_$(FLAVOR).dart \
	--obfuscate --split-debug-info=$(MOBILE_SYMBOLS) $(MOBILE_DEFINES)

# Generated Dart client (ADR-0003). Paths are relative to $(MOBILE). swagger_parser reads the
# normalized copy (mobile/swagger_parser.yaml), never the contract directly.
MOBILE_OPENAPI_JSON := build/openapi/openapi.json
MOBILE_API_DIR := lib/api/generated

# Coverage gate (P0-25): non-generated lib/ code only (mobile/tool/coverage_gate.dart).
MOBILE_COVER_MIN ?= 80
# Release outputs of FLAVOR that the artifact gate inspects (paths fixed by the Flutter tool).
MOBILE_APK := $(MOBILE)/build/app/outputs/flutter-apk/app-$(FLAVOR)-release.apk
MOBILE_AAB := $(MOBILE)/build/app/outputs/bundle/$(FLAVOR)Release/app-$(FLAVOR)-release.aab
MOBILE_ARTIFACT_GATE := tools/mobile/check-release-artifact.sh

.PHONY: mobile-check mobile-test-mock mobile-client mobile-client-check mobile-run mobile-run-mock \
	mobile-require-api-url mobile-build-apk mobile-build-aab mobile-build-ipa mobile-cover \
	mobile-artifact-check mobile-artifact-check-test

## mobile-check: format check, analyze (0 issues), widget/unit tests
mobile-check:
	cd $(MOBILE) && dart format --set-exit-if-changed . && flutter analyze && flutter test

## mobile-test-mock: integration tests against the Prism mock (start it first: make mock)
mobile-test-mock:
	cd $(MOBILE) && MOCK_API_URL=http://$(MOCK_HOST):$(MOCK_PORT) flutter test --tags mock

## mobile-client: regenerate the Dart API client from docs/api/openapi.yaml (normalize →
## swagger_parser → build_runner → kind guard → format). Fails on a kind collision.
mobile-client:
	cd $(MOBILE) && dart run tool/normalize_openapi.dart ../$(OPENAPI_SPEC) $(MOBILE_OPENAPI_JSON)
	cd $(MOBILE) && rm -rf $(MOBILE_API_DIR) && dart run swagger_parser
	cd $(MOBILE) && dart run build_runner build --delete-conflicting-outputs
	cd $(MOBILE) && dart run tool/check_generated_kinds.dart $(MOBILE_OPENAPI_JSON) $(MOBILE_API_DIR)
	cd $(MOBILE) && dart format $(MOBILE_API_DIR)

## mobile-client-check: drift check; fails when the committed client differs from the contract
mobile-client-check: mobile-client
	@drift="$$(git status --porcelain -- $(MOBILE)/$(MOBILE_API_DIR))"; \
	if [ -n "$$drift" ]; then \
	  echo "$$drift"; \
	  echo "Generated client is out of date: run 'make mobile-client' and commit $(MOBILE)/$(MOBILE_API_DIR)" >&2; \
	  exit 1; \
	fi

## mobile-run: run FLAVOR (dev|staging|prod) on the connected device
mobile-run:
	cd $(MOBILE) && flutter run --flavor $(FLAVOR) -t lib/main_$(FLAVOR).dart $(MOBILE_DEFINES)

## mobile-run-mock: run dev against the Prism mock (make mock, port 3001)
mobile-run-mock:
	cd $(MOBILE) && flutter run --flavor dev -t lib/main_dev.dart --dart-define=USE_MOCK=true $(MOBILE_DEFINES)

# staging/prod refuse to boot without an HTTPS API_BASE_URL (AppConfig); fail the build
# here instead of shipping an artifact that crashes on launch.
mobile-require-api-url:
	@case "$(FLAVOR)" in dev) ;; *) case "$(API_BASE_URL)" in https://?*) ;; \
	  *) echo "FLAVOR=$(FLAVOR) needs API_BASE_URL=https://..." >&2; exit 1 ;; esac ;; esac

## mobile-build-apk: obfuscated release APK for FLAVOR, symbols split out of the artifact
mobile-build-apk: mobile-require-api-url
	cd $(MOBILE) && flutter build apk $(MOBILE_RELEASE_FLAGS)

## mobile-build-aab: obfuscated release app bundle (Play upload) for FLAVOR, symbols split out
mobile-build-aab: mobile-require-api-url
	cd $(MOBILE) && flutter build appbundle $(MOBILE_RELEASE_FLAGS)

## mobile-cover: tests with coverage, then fail below MOBILE_COVER_MIN% on non-generated lib/ code
mobile-cover:
	cd $(MOBILE) && flutter test --coverage && dart run tool/coverage_gate.dart coverage/lcov.info $(MOBILE_COVER_MIN)

## mobile-artifact-check: fail if FLAVOR's release APK/AAB ships symbols, debug info or source maps
mobile-artifact-check:
	$(MOBILE_ARTIFACT_GATE) $(MOBILE_APK) $(MOBILE_AAB)

## mobile-artifact-check-test: prove the artifact gate rejects what it should (Linux; see the script for macOS)
mobile-artifact-check-test:
	tools/mobile/check-release-artifact_test.sh

## mobile-build-ipa: obfuscated release IPA for FLAVOR (needs Xcode + the iOS schemes, OB6)
mobile-build-ipa: mobile-require-api-url
	cd $(MOBILE) && flutter build ipa $(MOBILE_RELEASE_FLAGS)

.PHONY: hooks
## hooks: enable the repo's git hooks (gitleaks pre-commit secret scan)
hooks:
	git config core.hooksPath .githooks
