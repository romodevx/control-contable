// Package archtest machine-checks the Dependency Rule of ADR-0001 over the
// backend's own packages, complementing the depguard rules in .golangci.yml
// (which guard third-party imports but cannot express "another module").
//
// Layout: internal/<module>/{domain,usecase,adapter,infra} and internal/platform.
// Rules, for non-test files:
//
//   - domain-imports: a domain imports only the standard library (never
//     net/http or database/sql), its own module's domain tree and the pure
//     platform packages in domainPlatformAllowlist.
//   - usecase-frameworks: a usecase never imports net/http, database/sql,
//     GORM or pgx.
//   - usecase-outer-layer: within its module a usecase imports only domain and
//     usecase packages (never adapter, infra or any other directory).
//   - usecase-platform-adapter: a usecase never imports a platform adapter
//     (platformAdapters: pgdb, pgtest); it depends on the port in
//     internal/platform/tx instead.
//   - cross-module: no module package imports another module. Modules talk
//     through ports owned by the consumer, wired in cmd/.
//   - platform-modules: platform never imports a module.
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
//
// Apart from test-harness, the composition root (cmd/) may import anything.
package archtest
