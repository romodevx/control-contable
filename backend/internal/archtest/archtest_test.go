package archtest

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const fixtureModule = "example.com/fx"

func TestCheckImport(t *testing.T) {
	tests := []struct {
		name     string
		importer string
		imported string
		wantRule Rule
	}{
		{"domain may import stdlib", "internal/ledger/domain", "time", ""},
		{"domain may import its own subpackages", "internal/ledger/domain", "internal/ledger/domain/vo", ""},
		{"domain may import platform money", "internal/ledger/domain", "internal/platform/money", ""},
		{"domain may not import third-party", "internal/ledger/domain", "gorm.io/gorm", RuleDomainImports},
		// ADR-0004: NFKC password normalization; only this x/text package.
		{"domain may import unicode/norm", "internal/ledger/domain", "golang.org/x/text/unicode/norm", ""},
		{"domain may not import the rest of x/text", "internal/ledger/domain", "golang.org/x/text/cases", RuleDomainImports},
		{"domain may not import a norm look-alike", "internal/ledger/domain", "golang.org/x/text/unicode/normx", RuleDomainImports},
		{"pure platform may not import unicode/norm", "internal/platform/money", "golang.org/x/text/unicode/norm", RulePurePlatformImports},
		{"domain may not import its own usecase", "internal/ledger/domain", "internal/ledger/usecase", RuleDomainImports},
		{"domain may not import another platform package", "internal/ledger/domain", "internal/platform/tx", RuleDomainImports},
		{"domain may not import another module's domain", "internal/ledger/domain", "internal/billing/domain", RuleDomainImports},
		{"domain subpackage follows domain rule", "internal/ledger/domain/vo", "net/http", RuleDomainImports},
		{"usecase may import its domain", "internal/ledger/usecase", "internal/ledger/domain", ""},
		{"usecase may import platform", "internal/ledger/usecase", "internal/platform/tx", ""},
		{"usecase may not import the postgres adapter", "internal/ledger/usecase", "internal/platform/pgdb", RuleUsecasePlatformAdapter},
		{"usecase may not import the test harness", "internal/ledger/usecase", "internal/platform/pgtest", RuleUsecasePlatformAdapter},
		{"usecase may not import a platform adapter subpackage", "internal/ledger/usecase", "internal/platform/pgdb/x", RuleUsecasePlatformAdapter},
		{"infra may import the postgres adapter", "internal/ledger/infra", "internal/platform/pgdb", ""},
		{"usecase may import neutral third-party", "internal/ledger/usecase", "github.com/google/uuid", ""},
		{"usecase may not import net/http", "internal/ledger/usecase", "net/http", RuleUsecaseFrameworks},
		{"usecase may not import database/sql", "internal/ledger/usecase", "database/sql", RuleUsecaseFrameworks},
		{"usecase may not import gorm", "internal/ledger/usecase", "gorm.io/driver/postgres", RuleUsecaseFrameworks},
		{"usecase may not import pgx", "internal/ledger/usecase", "github.com/jackc/pgx/v5", RuleUsecaseFrameworks},
		{"usecase may not import its infra", "internal/ledger/usecase", "internal/ledger/infra", RuleUsecaseOuterLayer},
		{"usecase may not import its adapter", "internal/ledger/usecase", "internal/ledger/adapter/http", RuleUsecaseOuterLayer},
		{"usecase may not import another module", "internal/ledger/usecase", "internal/billing/usecase", RuleCrossModule},
		{"usecase may import a usecase subpackage", "internal/ledger/usecase", "internal/ledger/usecase/port", ""},
		{"usecase may not import an unknown sibling layer", "internal/ledger/usecase", "internal/ledger/app", RuleUsecaseOuterLayer},
		{"usecase may not import module root package", "internal/ledger/usecase", "internal/ledger", RuleUsecaseOuterLayer},
		{"infra may import gorm", "internal/ledger/infra", "gorm.io/gorm", ""},
		{"infra may import its usecase", "internal/ledger/infra", "internal/ledger/usecase", ""},
		{"infra may not import another module", "internal/ledger/infra", "internal/billing/usecase", RuleCrossModule},
		{"adapter may not import another module", "internal/ledger/adapter/http", "internal/billing/domain", RuleCrossModule},
		{"platform may import platform", "internal/platform/tx", "internal/platform/clock", ""},
		{"platform may not import a module", "internal/platform/money", "internal/ledger/domain", RulePlatformModules},
		{"impure platform may import gorm", "internal/platform/tx", "gorm.io/gorm", ""},
		{"pure platform may import stdlib", "internal/platform/money", "math/bits", ""},
		{"pure platform may import its own subpackages", "internal/platform/money", "internal/platform/money/round", ""},
		{"pure platform subpackage is pure too", "internal/platform/money/round", "gorm.io/gorm", RulePurePlatformImports},
		{"pure platform may not import gorm", "internal/platform/money", "gorm.io/gorm", RulePurePlatformImports},
		{"pure platform may not import net/http", "internal/platform/money", "net/http", RulePurePlatformImports},
		{"pure platform may not import impure platform", "internal/platform/money", "internal/platform/tx", RulePurePlatformImports},
		{"composition root may import anything", "cmd/api", "internal/ledger/infra", ""},
		// PR #12 follow-up 3: the RLS harness logs in as superuser with fixed
		// passwords; it must never be linked into a production binary.
		{"composition root may not import the test harness", "cmd/api", "internal/platform/pgtest", RuleTestHarness},
		{"infra may not import the test harness", "internal/ledger/infra", "internal/platform/pgtest", RuleTestHarness},
		{"adapter may not import the test harness", "internal/ledger/adapter/http", "internal/platform/pgtest", RuleTestHarness},
		{"platform may not import the test harness", "internal/platform/pgdb", "internal/platform/pgtest", RuleTestHarness},
		{"top-level package may not import the test harness", "migrations", "internal/platform/pgtest", RuleTestHarness},
		{"the harness may import its own subpackages", "internal/platform/pgtest", "internal/platform/pgtest/fixture", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := checkImport(fixtureModule, fixtureModule+"/"+tt.importer, qualify(tt.imported))
			assert.Equal(t, tt.wantRule, got)
		})
	}
}

// qualify prefixes repo-relative "internal/..." paths with the fixture module path.
func qualify(path string) string {
	if !strings.HasPrefix(path, "internal/") {
		return path
	}
	return fixtureModule + "/" + path
}

func TestCheckFixtureReportsEveryViolation(t *testing.T) {
	got, err := Check(filepath.Join("testdata", "fixture"), fixtureModule)
	require.NoError(t, err)

	want := []Violation{
		{File: "internal/ledger/domain/entry.go", Imported: "example.com/fx/internal/ledger/infra", Rule: RuleDomainImports},
		{File: "internal/ledger/domain/entry.go", Imported: "gorm.io/gorm", Rule: RuleDomainImports},
		{File: "internal/ledger/infra/repo.go", Imported: "example.com/fx/internal/billing/usecase", Rule: RuleCrossModule},
		{File: "internal/ledger/usecase/post.go", Imported: "example.com/fx/internal/billing/domain", Rule: RuleCrossModule},
		{File: "internal/ledger/usecase/post.go", Imported: "example.com/fx/internal/ledger/infra", Rule: RuleUsecaseOuterLayer},
		{File: "internal/ledger/usecase/post.go", Imported: "gorm.io/gorm", Rule: RuleUsecaseFrameworks},
		{File: "internal/ledger/usecase/post.go", Imported: "net/http", Rule: RuleUsecaseFrameworks},
		{File: "internal/platform/money/money.go", Imported: "example.com/fx/internal/ledger/domain", Rule: RulePlatformModules},
		{File: "internal/platform/money/money.go", Imported: "gorm.io/gorm", Rule: RulePurePlatformImports},
	}
	assert.Equal(t, want, got)
}

// P0-29: the per-import rules only deny the adapters they name, so a usecase
// reached GORM through platform/flags/flagsgorm and net/http through the
// top-level adapter/ tree, directly or through a neutral-looking package.
func TestCheckReportsFrameworksReachedThroughModulePackages(t *testing.T) {
	got, err := Check(filepath.Join("testdata", "transitive"), fixtureModule)
	require.NoError(t, err)

	const usecase = "internal/ledger/usecase/post.go"
	want := []Violation{
		{File: usecase, Imported: "example.com/fx/adapter/http/router", Rule: RuleInnerTransitiveFramework},
		{File: usecase, Imported: "example.com/fx/internal/platform/flags/flagsgorm", Rule: RuleInnerTransitiveFramework},
		{File: usecase, Imported: "example.com/fx/internal/platform/ports", Rule: RuleInnerTransitiveFramework},
	}
	assert.Equal(t, want, got)
}

// P0-30: the test harness must not reach a production binary through a
// neutral-looking package either. cmd/api imports seed, which imports pgtest;
// pgtest importing its own sub-package is fine.
func TestCheckReportsTestHarnessReachedThroughModulePackages(t *testing.T) {
	got, err := Check(filepath.Join("testdata", "harness"), fixtureModule)
	require.NoError(t, err)

	want := []Violation{
		{File: "cmd/api/main.go", Imported: "example.com/fx/internal/platform/seed", Rule: RuleTestHarness},
		{File: "internal/platform/seed/seed.go", Imported: "example.com/fx/internal/platform/pgtest", Rule: RuleTestHarness},
	}
	assert.Equal(t, want, got)
}

func TestBackendModuleRespectsDependencyRule(t *testing.T) {
	root := filepath.Join("..", "..")
	modulePath, err := ReadModulePath(filepath.Join(root, "go.mod"))
	require.NoError(t, err)
	require.Equal(t, "github.com/romodevx/control-contable/backend", modulePath)

	violations, err := Check(root, modulePath)
	require.NoError(t, err)
	assert.Empty(t, violations, "Dependency Rule violations (see internal/archtest/doc.go)")
}
