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

.PHONY: contract-tools contract-lint contract-lint-test mock

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
	node --test $(CONTRACT_TOOLS)/test.js

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

.PHONY: mobile-check mobile-run mobile-run-mock mobile-require-api-url mobile-build-apk mobile-build-ipa

## mobile-check: format check, analyze (0 issues), widget/unit tests
mobile-check:
	cd $(MOBILE) && dart format --set-exit-if-changed . && flutter analyze && flutter test

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

## mobile-build-ipa: obfuscated release IPA for FLAVOR (needs Xcode + the iOS schemes, OB6)
mobile-build-ipa: mobile-require-api-url
	cd $(MOBILE) && flutter build ipa $(MOBILE_RELEASE_FLAGS)
