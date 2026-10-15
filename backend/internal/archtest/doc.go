// Package archtest machine-checks the Dependency Rule of ADR-0001 over the
// backend's own packages, complementing the depguard rules in .golangci.yml
// (which guard third-party imports but cannot express "another module").
//
// Layout: internal/<module>/{domain,usecase,adapter,infra}, internal/platform
// and the shared kernel internal/kernel (ADR-0007).
// Rules, for non-test files:
//
//   - domain-imports: a domain imports only the standard library (never
//     net/http or database/sql), its own module's domain tree and the pure
//     platform packages in domainPlatformAllowlist, and the shared kernel.
//   - usecase-imports: from outside this module a usecase imports only the
//     packages in usecaseExternalAllowlist (pure standard library, uuid,
//     errgroup, unicode/norm). An allowlist: net/http, database/sql, GORM,
//     pgx, net, os/exec or any new SDK are refused until listed.
//   - usecase-outer-layer: within its module a usecase imports only domain and
//     usecase packages (never adapter, infra or any other directory).
//   - usecase-platform-adapter: a usecase never imports a platform adapter
//     (platformAdapters: pgdb, pgtest); it depends on the port in
//     internal/platform/tx instead.
//   - cross-module: no module package imports another module. Modules talk
//     through ports owned by the consumer, wired in cmd/.
//   - platform-modules: platform never imports a module or the shared kernel.
//   - kernel-imports: a shared-kernel package (internal/kernel/..., ADR-0007)
//     imports only the standard library minus frameworks, the pure platform
//     packages and other kernel packages. Every module's domain imports it,
//     so it is never a path from a domain to a framework or another module.
//   - pure-platform-imports: a package in domainPlatformAllowlist obeys the
//     domain rule (standard library minus frameworks, other pure platform
//     packages), so the domain cannot reach a framework through it.
//   - test-harness: no production file anywhere, cmd/ included, imports the
//     RLS test harness internal/platform/pgtest (superuser logins, fixed
//     passwords), directly or through other packages of this module. Only
//     _test.go files and pgtest itself may.
//   - inner-transitive-framework: a domain or usecase never imports a package
//     of this module (platform adapter sub-packages such as flagsgorm, the
//     top-level adapter/ tree, any helper) that reaches net/http,
//     database/sql, GORM or pgx through this module's own imports. Packages
//     outside the module are not followed.
//   - adapter-unit-of-work: an HTTP adapter (the top-level adapter/ tree or a
//     module's adapter layer) never imports internal/platform/tx. Use cases
//     own units of work, so no handler can open one, least of all a
//     tx.SystemManager transaction without a user context (P0-30 D-2).
//   - inward-adapter: only adapters (the top-level adapter/ tree, a module's
//     adapter layer) and cmd/ import the top-level adapter/ tree: the router,
//     the generated server and httperr (P0-29 FU3, ADR-0010). Delivery
//     details never flow inward, into platform, infra, use cases or
//     migrations.
//
// Apart from test-harness, the composition root (cmd/) may import anything.
//
// platform/access is the one platform package with business rules (trial,
// caps, entitlements), kept there as a documented shared port by ADR-0010:
// modules consume it inside their own transactions, which a module package
// could not offer them under cross-module. No other business logic goes
// into platform/.
//
// CheckRepositories adds one rule that is not about imports (P0-30 D-3): a
// module repository (internal/<module>/infra, non-test files) never calls
// Begin, BeginTx, Commit, Rollback, SavePoint, RollbackTo or Transaction and
// never sends SQL that sets session state (set_config, SET, RESET). The use
// case's unit of work owns the transaction and the RLS context.
package archtest
