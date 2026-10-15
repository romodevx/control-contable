# Runbook: local development

Brings up the whole stack on one machine, from a fresh clone to a running API, worker, mock
server and mobile client. Everything here is **LOCAL only**: no command in this runbook may
point at a shared, staging or production database.

| Service | Address | Started by |
|---|---|---|
| PostgreSQL 18 | `127.0.0.1:5434` | `make -C backend db-up` |
| API | `http://localhost:3000` | `make -C backend api` |
| Prism mock (contract) | `http://127.0.0.1:3001` | `make mock` |
| Worker | no port | `make -C backend worker` |

The ports come from the project registry: 5432 and 5433 belong to other local stacks.

## 1. Prerequisites

Check the toolchain against the [README](../../README.md#prerequisites):

```sh
go version          # go1.27.x (backend/go.mod pins toolchain go1.27.2)
docker compose version
node --version      # the version in .nvmrc (24.18.0 at least)
flutter --version   # 3.47.7
```

## 2. Database

```sh
make -C backend init-env
```

This writes `infra/local/.env` from `infra/local/.env.example` with random passwords for
`postgres`, `cc_owner`, `cc_app`, `cc_webhook` and `cc_worker` (mode 0600, git-ignored). It does
nothing if the file already exists. `db-up` runs it for you, so this step is optional.

```sh
make -C backend db-up
```

Expect `local target confirmed: docker=unix://… db_host=127.0.0.1`, then
`Container control-contable-postgres Healthy`. The container binds to loopback only, and the
role bootstrap (`infra/local/postgres`) runs the first time the data volume is created.

```sh
make -C backend migrate-up        # OK up 20261013120000_baseline.sql, OK up …_identity.sql
make -C backend migrate-status    # every migration "applied"
make -C backend db-roles-check    # NOTICE: OK: cc_app is NOSUPERUSER, NOBYPASSRLS, … (one per role)
```

Migrations run as `cc_owner`. The runtime roles (`cc_app`, `cc_webhook`, `cc_worker`) are never
owners and never BYPASSRLS; `db-roles-check` proves it over TCP.

## 3. API and worker

```sh
make -C backend api
```

Expect a JSON log line `"msg":"api listening","addr":"[::]:3000","env":"dev"`. In another shell:

```sh
curl -i http://localhost:3000/healthz   # 200 {"database":"ok","status":"ok","version":"0.0.0-dev"}
curl http://localhost:3000/v1/flags     # {"enabled":[]}  (every flag is OFF by default)
curl -i http://localhost:3000/v1/me     # 404: module_auth is OFF, so the route is hidden
```

```sh
make -C backend worker
```

Expect `"msg":"worker started"` and `"msg":"leadership acquired"`. No job is registered yet.
`Ctrl+C` stops either process gracefully (the API drains in-flight requests for up to 8 s).

Run on another port with `make -C backend api API_PORT=3010`.

### Access-token signing key (M1-07)

`make init-env` writes a random Ed25519 seed to `infra/local/.env` (`CC_AUTH_ACCESS_TOKEN_KEY`,
kid `CC_ACCESS_TOKEN_KEY_ID=k1`; the file is git-ignored, never commit it). Access tokens are
EdDSA JWTs with that `kid`, valid 15 min; the API accepts `Authorization: Bearer <token>` only.

```sh
make -C backend access-key-rotate   # new key + kid; the old pair stays as the verify-only previous key
```

Restart `make -C backend api` afterwards. Tokens signed by the previous key keep working until
they expire; a second rotation drops it. Shared environments set `AUTH_ACCESS_TOKEN_KEY`,
`AUTH_ACCESS_TOKEN_KEY_ID` and, while rotating, `AUTH_ACCESS_TOKEN_PREVIOUS_KEY` /
`AUTH_ACCESS_TOKEN_PREVIOUS_KEY_ID` from the secret manager.

## 4. Feature flags

```sh
make -C backend flags                                # all 10 module flags, "0 of 10 modules ON"
make -C backend flags ARGS="enable module_auth"
make -C backend flags ARGS="disable module_auth"
make -C backend flags ARGS="presets"                 # demo_core, demo_monetization, demo_full
make -C backend flags ARGS="apply-preset demo_core"
```

The CLI runs as `cc_owner`. A running API picks up a change within 5 s. Turning ON the flag of a
module that is not built yet changes nothing: its routes keep answering 404. Auth operations
(sessions, rate limits, signing keys) are in the [auth runbook](auth.md).

## 5. Mock server and mobile

The mock needs neither Docker nor the backend:

```sh
make mock                       # first run installs the pinned tools in tools/contract
curl http://127.0.0.1:3001/healthz
```

Details: [mock-server runbook](mock-server.md). Then, from the repository root:

```sh
make mobile-client              # regenerate mobile/lib/api/generated from the contract
make mobile-run-mock            # dev flavor against the mock
make mobile-run                 # dev flavor against the local API on port 3000
make mobile-run API_BASE_URL=http://10.0.2.2:3000   # Android emulator: the host is 10.0.2.2
```

The Android emulator reaches the host's `localhost` as `10.0.2.2`; the API listens on every
interface (`[::]:3000`), so no extra setup is needed. A debug APK of the dev flavor is about
200 MB: on an emulator low on storage `flutter run`/`adb install -r` fails with
`INSTALL_FAILED_INSUFFICIENT_STORAGE`. A profile build for the emulator's ABI is about 30 MB
and behaves the same against the local API:

```sh
cd mobile && flutter build apk --profile --flavor dev -t lib/main_dev.dart \
  --target-platform android-arm64 --dart-define=API_BASE_URL=http://10.0.2.2:3000
adb install -r build/app/outputs/flutter-apk/app-dev-profile.apk
```

With `module_auth` ON and `make -C backend api` running, the dev email adapter prints each
verification code in the API log (`"msg":"dev email: verification code"`; the address is
redacted). It is the only way to read a code locally.

## 6. API reference

```sh
make api-docs                   # build/api-docs/index.html, generated from docs/api/openapi.yaml
```

## 7. Stop, reset

```sh
make -C backend db-down         # stop the container, keep the data
make -C backend db-reset        # LOCAL ONLY: drop the volume, bootstrap the roles again
make -C backend migrate-up      # needed after every reset
```

`db-reset` deletes all local data. It refuses to run when the Docker engine is not a local
socket or when `CC_DB_HOST` or `DATABASE_URL` is not loopback.

## Troubleshooting

### Port 5434 is already allocated

`db-up` fails with `Bind for 127.0.0.1:5434 failed: port is already allocated`.

```sh
lsof -nP -iTCP:5434 -sTCP:LISTEN
docker ps --format '{{.Names}}\t{{.Ports}}' | grep 5434
```

If it is another project, stop it. If you must keep it, change `CC_DB_PORT` in
`infra/local/.env` (every target reads it from there) and register the new port.

### Port 3000 or 3001 is in use

`api` logs `bind: address already in use`; Prism reports `EADDRINUSE`.

```sh
lsof -nP -iTCP:3000 -iTCP:3001 -sTCP:LISTEN
```

Usually a previous `make api` or `make mock` is still running in another shell: stop it with
`Ctrl+C` or `kill <pid>`. Otherwise pick another port: `make -C backend api API_PORT=3010`,
`make mock MOCK_PORT=3011`. The mobile app then needs `API_BASE_URL` to match.

### Stale container from another checkout or worktree

The compose project (`control-contable`), the container (`control-contable-postgres`) and the
volume (`control-contable-pgdata`) have fixed names. Every clone and worktree on the machine
shares one database. Symptoms:

- `db-up` recreates a container that another checkout started;
- `password authentication failed for user "cc_owner"`: your `infra/local/.env` holds
  different passwords than the ones the volume was created with.

Find who started it:

```sh
docker inspect control-contable-postgres \
  --format '{{index .Config.Labels "com.docker.compose.project.working_dir"}}'
```

Then either copy that checkout's `infra/local/.env` into yours (keeps the data), or run
`make -C backend db-reset` followed by `migrate-up` (new volume, your passwords). If you delete
`.env` but keep the volume, the roles keep their old passwords: reset.

### `db-roles-check` fails on a maintenance database

`ERROR: cc_app can CONNECT to a maintenance database` means the volume was bootstrapped before the
current role hardening. The bootstrap runs only when the volume is created, so reset it:
`make -C backend db-reset && make -C backend migrate-up`.

### `REFUSED (not local)`

The guard found a remote Docker host (`DOCKER_HOST`, a non-local `docker context`), a
non-loopback `CC_DB_HOST`, or a remote `DATABASE_URL` in your environment. Unset it. The guard
exists so these targets can never touch a shared database.

### The API answers 404 for a route in the contract

Its module flag is OFF. Check with `make -C backend flags` and `curl localhost:3000/v1/flags`.
After `enable`, wait up to 5 s for the cache.

### `/healthz` reports `"database":"unreachable"`

The API is up but cannot reach PostgreSQL: `make -C backend db-up`. While the database is down
every flagged module is hidden (fail closed).

### `make mock` or `make api-docs` fails during `npm ci`

Use the Node version in `.nvmrc` (`nvm use`). The tooling is installed from the lockfile in
`tools/contract`; delete `tools/contract/node_modules` and retry if the install was interrupted.

## Test secrets (read before writing any test)

The repository is public and scanned by GitGuardian on every push. A fake value that
*looks* like a secret still raises an incident. Rules:

- Generate passwords, keys, tokens and JWTs at test time (`crypto/rand`, `ed25519.GenerateKey`,
  Dart `Random.secure()`); never commit literals that look like credentials, even fake ones.
- If a fixed vector is unavoidable (e.g. a published crypto test vector), build it from parts
  at runtime or hex-encode it with a comment explaining why.
- `make hooks` enables the gitleaks pre-commit hook (`.githooks/pre-commit`); commits with a
  finding are refused locally, before anything reaches GitHub. CI re-runs the scan (`secrets.yml`).
