package params_test

import (
	"errors"
	"io/fs"
	"os"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/platform/money"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain/taxparams"
	"github.com/romodevx/control-contable/backend/internal/taxengine/params"
)

func date(t testing.TB, s string) taxparams.Date {
	t.Helper()
	d, err := taxparams.ParseDate(s)
	require.NoError(t, err)
	return d
}

func TestLoadSyntheticFixtures(t *testing.T) {
	catalog, err := params.Load(os.DirFS("testdata"), "valid")
	require.NoError(t, err)
	assert.Equal(t, []taxparams.SetID{{FiscalYear: 2025, Version: 1}, {FiscalYear: 2026, Version: 1}, {FiscalYear: 2026, Version: 2}}, catalog.Sets())

	sel, err := catalog.Select(2026)
	require.NoError(t, err)
	require.Equal(t, 2, sel.Params.Version())
	assert.False(t, sel.Provisional)
	p := sel.Params
	assert.Len(t, p.Sources(), 2)
	assert.Equal(t, "2025-12-31", p.Sources()[0].DOFDate.String())

	border, err := p.Rate("iva.rate.border_8", date(t, "2026-12-31"))
	require.NoError(t, err)
	assert.Equal(t, money.Rate(80_000), border)
	_, err = p.Rate("iva.rate.border_8", date(t, "2027-01-01"))
	require.ErrorIs(t, err, taxparams.ErrNotInForce)

	jan, err := p.Money("uma.monthly", date(t, "2026-01-31"))
	require.NoError(t, err)
	assert.Equal(t, money.Cents(10_000), jan)
	feb, err := p.Money("uma.monthly", date(t, "2026-02-01"))
	require.NoError(t, err)
	assert.Equal(t, money.Cents(11_100), feb)

	rounding, err := p.Flag("rounding.payment_whole_pesos", date(t, "2026-03-31"))
	require.NoError(t, err)
	assert.False(t, rounding)

	tariff, err := p.Table("isr.art96.monthly", date(t, "2026-03-31"))
	require.NoError(t, err)
	assert.Equal(t, taxparams.TableMarginal, tariff.Kind)
	assert.Equal(t, []taxparams.Row{
		{Lower: 1, Upper: 10_000, FixedFee: 0, Rate: 20_000},
		{Lower: 10_001, Unbounded: true, FixedFee: 200, Rate: 100_000},
	}, tariff.Rows)
	assert.Equal(t, []taxparams.Citation{{Source: "anexo-synthetic", Article: "B.V"}}, tariff.Basis)

	brackets, err := p.Table("isr.resico_pf.monthly", date(t, "2026-03-31"))
	require.NoError(t, err)
	assert.Equal(t, taxparams.TableFlat, brackets.Kind)

	december, err := p.MonthlyTable("isr.art106.accumulated", time.December, date(t, "2026-12-31"))
	require.NoError(t, err)
	assert.Equal(t, money.Cents(120_000), december.Rows[0].Upper)

	require.NoError(t, p.Require(taxparams.RequireMonthlyTables("isr.art106.accumulated")))

	older, err := catalog.Select(2025)
	require.NoError(t, err)
	assert.True(t, older.Provisional, "the 2025 fixture is marked provisional")
}

func TestEmbeddedHasNoSetsUntilM5_04(t *testing.T) {
	// The real 2026 parameters are transcribed by M5-04, which replaces this
	// test with one asserting the embedded catalog loads and selects 2026.
	_, err := params.Embedded()
	require.ErrorIs(t, err, taxparams.ErrNoParameterSets)
}

// fixture2026 is the valid synthetic 2026/v1 file, the base the invalid
// cases edit.
func fixture2026(t testing.TB) string {
	t.Helper()
	b, err := os.ReadFile("testdata/valid/2026/v1.yaml")
	require.NoError(t, err)
	return string(b)
}

func one(path, content string) fstest.MapFS {
	return fstest.MapFS{path: &fstest.MapFile{Data: []byte(content)}}
}

func TestLoadRejects(t *testing.T) {
	base := fixture2026(t)
	replace := func(old, repl string) string {
		if !strings.Contains(base, old) {
			t.Fatalf("fixture lacks %q", old)
		}
		return strings.Replace(base, old, repl, 1)
	}
	tests := []struct {
		name string
		fsys fstest.MapFS
		want error
	}{
		{"file outside a year directory", one("v1.yaml", base), params.ErrUnexpectedFile},
		{"file name not v<N>.yaml", one("2026/v1.yml", base), params.ErrUnexpectedFile},
		{"version zero in the name", one("2026/v0.yaml", base), params.ErrUnexpectedFile},
		{"nested directory", one("2026/drafts/v1.yaml", base), params.ErrUnexpectedFile},
		{"directory that is not a year", one("misc/v1.yaml", base), params.ErrUnexpectedFile},
		{"fiscal year disagrees with the path", one("2027/v1.yaml", base), params.ErrPathMismatch},
		{"version disagrees with the path", one("2026/v3.yaml", base), params.ErrPathMismatch},
		{"empty file", one("2026/v1.yaml", ""), params.ErrSyntax},
		{"not yaml", one("2026/v1.yaml", "fiscal_year: [2026"), params.ErrSyntax},
		{"unknown field", one("2026/v1.yaml", replace("version: 1\n", "version: 1\nrevision: 2\n")), params.ErrSyntax},
		{"two documents", one("2026/v1.yaml", base+"---\nfiscal_year: 2026\n"), params.ErrSyntax},
		{"unquoted rate is a float", one("2026/v1.yaml", replace(`rate: "0.08"`, `rate: 0.08`)), params.ErrInexactNumber},
		{"unquoted amount is a float", one("2026/v1.yaml", replace(`money: "100.00"`, `money: 100.00`)), params.ErrInexactNumber},
		{"integer amount", one("2026/v1.yaml", replace(`money: "100.00"`, `money: 100`)), params.ErrInexactNumber},
		{"unquoted tariff bound", one("2026/v1.yaml", replace(`upper: "100.00", fixed_fee`, `upper: 100.00, fixed_fee`)), params.ErrInexactNumber},
		{"decimal as a mapping", one("2026/v1.yaml", replace(`rate: "0.08"`, `rate: {value: "0.08"}`)), params.ErrInexactNumber},
		{"rate with seven decimals", one("2026/v1.yaml", replace(`rate: "0.08"`, `rate: "0.0800001"`)), money.ErrPrecision},
		{"amount with three decimals", one("2026/v1.yaml", replace(`money: "100.00"`, `money: "100.001"`)), money.ErrPrecision},
		{"bound that is not a decimal", one("2026/v1.yaml", replace(`lower: "100.01", fixed_fee: "2.00"`, `lower: "1e2", fixed_fee: "2.00"`)), money.ErrSyntax},
		{"fee that is not a decimal", one("2026/v1.yaml", replace(`fixed_fee: "2.00"`, `fixed_fee: "two"`)), money.ErrSyntax},
		{"upper that is not a decimal", one("2026/v1.yaml", replace(`upper: "200.00"`, `upper: "2,00"`)), money.ErrSyntax},
		{"row without rate", one("2026/v1.yaml", replace(`, rate: "0.011"}`, `}`)), money.ErrSyntax},
		{"value with no kind", one("2026/v1.yaml", replace(`    flag: false
`, ``)), params.ErrValueKind},
		{"value with two kinds", one("2026/v1.yaml", replace(`rate: "0.08"`, `rate: "0.08"
    money: "1.00"`)), params.ErrValueKind},
		{"marginal row without fixed_fee", one("2026/v1.yaml", replace(`lower: "100.01", fixed_fee: "2.00", rate: "0.1"`, `lower: "100.01", rate: "0.1"`)), params.ErrMissingFixedFee},
		{"unknown table kind", one("2026/v1.yaml", replace("kind: flat", "kind: progressive")), params.ErrTableKind},
		{"date that is not a calendar day", one("2026/v1.yaml", replace("valid_to: 2026-01-31", "valid_to: 2026-02-30")), taxparams.ErrInvalidDate},
		{"timestamp instead of a date", one("2026/v1.yaml", replace("published_on: 2026-10-13", "published_on: 2026-10-13T10:00:00Z")), taxparams.ErrInvalidDate},
		{"date as a number", one("2026/v1.yaml", replace("published_on: 2026-10-13", "published_on: 20261013")), taxparams.ErrInvalidDate},
		{"source date quoted but malformed", one("2026/v1.yaml", replace("dof_date: 2025-12-31", `dof_date: "31/12/2025"`)), taxparams.ErrInvalidDate},
		{"domain validation still applies", one("2026/v1.yaml", replace(`rate: "0.08"`, `rate: "1.08"`)), taxparams.ErrRateOutOfRange},
		{"monthly tariffs incomplete", one("2026/v1.yaml", replace("month: 12", "month: 11")), taxparams.ErrMonthlyIncomplete},
		{"no files at all", fstest.MapFS{"README.md": &fstest.MapFile{Data: []byte("x")}}, taxparams.ErrNoParameterSets},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			catalog, err := params.Load(tt.fsys, ".")
			require.ErrorIs(t, err, tt.want)
			assert.Nil(t, catalog)
		})
	}
}

func TestLoadReportsEveryBadFileWithItsPath(t *testing.T) {
	base := fixture2026(t)
	fsys := fstest.MapFS{
		"2026/v1.yaml": &fstest.MapFile{Data: []byte(strings.Replace(base, `rate: "0.08"`, `rate: 0.08`, 1))},
		"2026/v2.yaml": &fstest.MapFile{Data: []byte(base)},
		"notes.txt":    &fstest.MapFile{Data: []byte("x")},
	}
	_, err := params.Load(fsys, ".")
	require.ErrorIs(t, err, params.ErrInexactNumber)
	require.ErrorIs(t, err, params.ErrPathMismatch)
	require.ErrorIs(t, err, params.ErrUnexpectedFile)
	for _, path := range []string{"2026/v1.yaml", "2026/v2.yaml", "notes.txt"} {
		assert.Contains(t, err.Error(), path)
	}
}

// failingFS fails to open one file, to prove read errors surface.
type failingFS struct {
	fs.FS
	fail string
}

var errDisk = errors.New("disk on fire")

func (f failingFS) Open(name string) (fs.File, error) {
	if name == f.fail {
		return nil, errDisk
	}
	return f.FS.Open(name)
}

func TestLoadSurfacesFileSystemErrors(t *testing.T) {
	_, err := params.Load(failingFS{FS: one("2026/v1.yaml", fixture2026(t)), fail: "2026/v1.yaml"}, ".")
	require.ErrorIs(t, err, errDisk)

	_, err = params.Load(fstest.MapFS{}, "missing")
	require.ErrorIs(t, err, fs.ErrNotExist)
}

func TestLoadUnderARoot(t *testing.T) {
	fsys := fstest.MapFS{"data/2026/v1.yaml": &fstest.MapFile{Data: []byte(fixture2026(t))}}
	catalog, err := params.Load(fsys, "data")
	require.NoError(t, err)
	assert.Equal(t, []taxparams.SetID{{FiscalYear: 2026, Version: 1}}, catalog.Sets())
}
