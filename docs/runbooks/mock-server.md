# Runbook: OpenAPI mock server (Prism)

Serves `docs/api/openapi.yaml` so mobile work can start before the backend exists (PLAN P0-26).
The contract is the only input: once a module's contract is frozen, the mock serves it.

## Start / stop

```sh
make mock                     # http://127.0.0.1:3001, static examples from the contract
make mock MOCK_DYNAMIC=true   # bodies generated from the schemas (random values)
make mock MOCK_PORT=3011      # another port if 3001 is taken
```

Stop with `Ctrl+C`. The first run installs the pinned tooling in `tools/contract` (same lockfile
as `make contract-lint`).

| Variable | Default | Meaning |
|---|---|---|
| `MOCK_HOST` | `127.0.0.1` | Bind address. Use `0.0.0.0` only to reach it from a physical device on the LAN. |
| `MOCK_PORT` | `3001` | Registry port for control-contable's mock (API is 3000, Postgres 5434). Check with `/projects check-port`. |
| `MOCK_DYNAMIC` | `false` | `false`: return the `example`/`examples` declared in the contract (deterministic; use for widget tests and demos). `true`: generate values from the schemas (use to shake out client parsing). |

Pinned: `@stoplight/prism-cli` 5.16.0. It declares Node >= 24.18; older 24.x prints an
`EBADENGINE` warning on install but runs.

## Smoke

```sh
curl -s http://127.0.0.1:3001/healthz
curl -s -X POST http://127.0.0.1:3001/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"demo@example.com","password":"Sup3rSecret!pass"}'
```

## Behaviour to know

- Requests are validated against the contract: an invalid body gets a `4xx` listing the violations.
- Force a documented response with `-H 'Prefer: code=401'`, or a named example with
  `-H 'Prefer: example=<name>'`.
- The mock has no state and no auth: any bearer token is accepted and nothing persists.
  Feature flags are not enforced; the server-side `404` for disabled modules exists only in
  the real backend.

## Mobile

- iOS simulator / desktop: `http://127.0.0.1:3001`.
- Android emulator: `http://10.0.2.2:3001` (the emulator's alias for the host loopback).
- Physical device: `make mock MOCK_HOST=0.0.0.0`, then the Mac's LAN IP. Stop it afterwards.
