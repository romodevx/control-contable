// Package taxengine_test is the tax engine's golden gate (PLAN M5-07,
// MODULE-tax-engine §3, US-X3). It is part of `go test ./...`, so every
// mismatch fails CI; there is no flag that updates an expected value.
package taxengine_test

import (
	"fmt"
	"io/fs"
	"os"
	"regexp"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain/taxparams"
	"github.com/romodevx/control-contable/backend/internal/taxengine/golden"
	"github.com/romodevx/control-contable/backend/internal/taxengine/params"
)

const (
	goldenRoot   = "golden" // under testdata/
	paramsFile   = "params/data/2026/v1.yaml"
	researchFile = "../../../docs/research/tax-regimes-mx-2026.md"
	baselineFile = "testdata/golden_baseline.yaml"
	moduleRoot   = "../.." // backend/
	writeScope   = "internal/taxengine/"
)

// gateStrategies are the strategies the gate runs, through the real M5-10
// registry. A strategy PR (M5-11..M5-19) adds its strategy here and removes
// pending_strategy from its cases in the same PR, until M5-22 wires them all
// at the composition root and the gate can take that list.
func gateStrategies() []domain.Strategy {
	return []domain.Strategy{
		domain.NewStrategy612PF(),
		domain.NewRESICOPFStrategy(), // M5-13
		domain.NewRESICOPMStrategy(), // M5-16
	}
}

func registry(t *testing.T) domain.Registry {
	t.Helper()
	r, err := domain.NewRegistry(gateStrategies()...)
	require.NoError(t, err)
	return r
}

func catalog(t *testing.T) *taxparams.Catalog {
	t.Helper()
	c, err := params.Embedded()
	require.NoError(t, err)
	return c
}

func rawParams(t *testing.T) golden.RawParams {
	t.Helper()
	data, err := os.ReadFile(paramsFile)
	require.NoError(t, err)
	rp, err := golden.ReadRawParams(data)
	require.NoError(t, err)
	return rp
}

func authoredCases(t *testing.T) []golden.Case {
	t.Helper()
	cases, err := golden.Load(os.DirFS("testdata"), goldenRoot)
	require.NoError(t, err)
	return cases
}

// TestGoldenCases runs every authored and generated case. Skipped (⚠️) and
// pending_strategy cases are listed in the log with their reasons.
func TestGoldenCases(t *testing.T) {
	authored := authoredCases(t)
	generated := generatedCases(t, rawParams(t))
	cases := slices.Concat(authored, generated)

	s := golden.Run(t, cases, registry(t), catalog(t))

	t.Logf("authored %d, generated %d", len(authored), len(generated))
	assert.Equal(t, len(cases), s.Total())
	assert.Empty(t, s.Ratchet(baseline(t)),
		"skipped and pending_strategy cases may only go down; edit %s by hand in the PR that changes them", baselineFile)
}

func baseline(t *testing.T) golden.Baseline {
	t.Helper()
	data, err := os.ReadFile(baselineFile)
	require.NoError(t, err)
	b, err := golden.ReadBaseline(data)
	require.NoError(t, err)
	return b
}

// TestGoldenSourcesAreTheResearchWorkedExamples ties every research case to
// a worked example that exists, and every worked example to a case.
func TestGoldenSourcesAreTheResearchWorkedExamples(t *testing.T) {
	doc, err := os.ReadFile(researchFile)
	require.NoError(t, err)
	examples := map[string]bool{}
	for _, m := range regexp.MustCompile(`(?m)^\| ((?:PF|PM)-[0-9A-Z-]+) \|`).FindAllStringSubmatch(string(doc), -1) {
		examples[m[1]] = false
	}
	require.NotEmpty(t, examples)

	for _, c := range authoredCases(t) {
		id, research := strings.CutPrefix(c.Source, "research:")
		if !research {
			continue
		}
		_, known := examples[id]
		assert.True(t, known, "%s: source %s is not a worked example of %s", c.File, id, researchFile)
		examples[id] = true
	}
	for id, seeded := range examples {
		assert.True(t, seeded, "worked example %s has no golden case", id)
	}
}

// TestGoldensAreNeverRewritten is the no-update-flag gate: nothing in the
// module may regenerate an expected value, and every case file still
// carries its hand-authored header (golden.Load refuses one without it).
func TestGoldensAreNeverRewritten(t *testing.T) {
	violations, err := golden.FindRewriters(os.DirFS(moduleRoot), writeScope)
	require.NoError(t, err)
	assert.Empty(t, violations, "goldens are hand-authored from official examples; remove the rewrite path")

	err = fs.WalkDir(os.DirFS("testdata"), goldenRoot, func(name string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || strings.HasSuffix(name, "README.md") {
			return err
		}
		data, err := fs.ReadFile(os.DirFS("testdata"), name)
		if err != nil {
			return err
		}
		assert.True(t, strings.HasPrefix(string(data), golden.HandAuthoredHeader), "%s lost its header", name)
		return nil
	})
	require.NoError(t, err)
}

// TestReferenceReproducesTheResearchArithmetic checks the reference
// against the research's own worked computations before it judges the
// engine (docs/research/tax-regimes-mx-2026.md §A.5).
func TestReferenceReproducesTheResearchArithmetic(t *testing.T) {
	rp := rawParams(t)
	tests := []struct {
		example string
		key     string
		month   int
		base    string
		want    string
	}{
		{"PF-612-ISR-01", "isr.art106.accumulated", 3, "60000.00", "7150.96"},
		{"PF-612-ISR-02", "isr.art106.accumulated", 10, "250000.00", "34516.54"},
		{"PF-606-ISR-01", "isr.art116.monthly", 0, "13000.00", "1076.00"},
		{"PF-626-ISR-01", "isr.resico_pf.monthly", 0, "60000.00", "900.00"},
		{"PF-626-ISR-02 a", "isr.resico_pf.monthly", 0, "25000.00", "250.00"},
		{"PF-626-ISR-02 b", "isr.resico_pf.monthly", 0, "25000.01", "275.00"},
		{"PF-621-01 before the reduction", "isr.rif.bimonthly", 0, "40000.00", "4767.31"},
	}
	for _, tt := range tests {
		t.Run(tt.example, func(t *testing.T) {
			table := findTable(t, rp, tt.key, tt.month)

			got, ok := refTax(table, mustCents(t, tt.base))

			require.True(t, ok)
			assert.Equal(t, tt.want, got.String())
		})
	}
}

// TestTariffEdgesMatchTheReference runs the engine's tariff lookup at
// every row edge of every 2026 table: lower, upper and upper + 0.01.
func TestTariffEdgesMatchTheReference(t *testing.T) {
	p := params2026(t)
	total := 0
	for _, raw := range rawParams(t).Tables {
		points, err := golden.TariffEdges(raw)
		require.NoError(t, err)
		table := engineTable(t, p, raw, mustDate(t, string(raw.ValidFrom)))
		for _, pt := range points {
			total++
			want, inTable := refTax(raw, pt.Amount)
			exact, err := pt.Amount.Exact()
			require.NoError(t, err)
			got, err := table.Tax(exact)
			if !inTable {
				assert.ErrorIs(t, err, taxparams.ErrAboveTable, "%s month %d %s %s", raw.Key, raw.Month, pt.Label, pt.Amount)
				continue
			}
			if assert.NoError(t, err, "%s month %d %s", raw.Key, raw.Month, pt.Label) {
				assert.Equal(t, want, got.Round(), "%s month %d %s at %s", raw.Key, raw.Month, pt.Label, pt.Amount)
			}
		}
	}
	t.Logf("tariff edges checked: %d", total)
	assert.Positive(t, total)
}

// TestValidityEdgesMatchTheReference resolves every value and table on the
// day before, the first day, the last day and the day after each entry
// (rate_8 on 2026-12-31 vs 2027-01-01, UMA January vs February, ...).
func TestValidityEdgesMatchTheReference(t *testing.T) {
	p := params2026(t)
	rp := rawParams(t)
	points, err := golden.ValidityEdges(rp)
	require.NoError(t, err)
	for _, pt := range points {
		name := fmt.Sprintf("%s month %d on %s", pt.Key, pt.Month, pt.Date)
		if pt.Table {
			checkTableValidity(t, p, rp, pt, name)
			continue
		}
		checkValueValidity(t, p, rp, pt, name)
	}
	t.Logf("validity edges checked: %d", len(points))
	assert.NotEmpty(t, points)
}

func checkValueValidity(t *testing.T, p *taxparams.Params, rp golden.RawParams, pt golden.ValidityPoint, name string) {
	t.Helper()
	var entries []golden.RawValue
	var froms, tos []golden.RawDate
	for _, v := range rp.Values {
		if v.Key == pt.Key {
			entries, froms, tos = append(entries, v), append(froms, v.ValidFrom), append(tos, v.ValidTo)
		}
	}
	got, err := p.At(taxparams.Key(pt.Key), pt.Date)
	i := refInForce(froms, tos, pt.Date.String())
	if i < 0 {
		assert.ErrorIs(t, err, taxparams.ErrNotInForce, name)
		return
	}
	if !assert.NoError(t, err, name) {
		return
	}
	want := entries[i]
	assert.Equal(t, string(want.ValidFrom), got.Validity.From.String(), name)
	assert.Equal(t, string(want.ValidTo), dateOrEmpty(got.Validity.To), name)
	switch {
	case want.Rate != "":
		assert.Equal(t, 0, refDecimal(want.Rate).Cmp(refRate(got.Rate)), name)
	case want.Money != "":
		assert.Equal(t, 0, refDecimal(want.Money).Cmp(refPesos(got.Money)), name)
	}
}

func checkTableValidity(t *testing.T, p *taxparams.Params, rp golden.RawParams, pt golden.ValidityPoint, name string) {
	t.Helper()
	var entries []golden.RawTable
	var froms, tos []golden.RawDate
	for _, tb := range rp.Tables {
		if tb.Key == pt.Key && tb.Month == pt.Month {
			entries, froms, tos = append(entries, tb), append(froms, tb.ValidFrom), append(tos, tb.ValidTo)
		}
	}
	got, err := lookupTable(p, pt.Key, pt.Month, pt.Date)
	i := refInForce(froms, tos, pt.Date.String())
	if i < 0 {
		assert.ErrorIs(t, err, taxparams.ErrNotInForce, name)
		return
	}
	if !assert.NoError(t, err, name) {
		return
	}
	want := entries[i]
	assert.Equal(t, string(want.ValidFrom), got.Validity.From.String(), name)
	assert.Equal(t, string(want.ValidTo), dateOrEmpty(got.Validity.To), name)
	assert.Len(t, got.Rows, len(want.Rows), name)
}

func dateOrEmpty(d fiscal.Date) string {
	if d.IsZero() {
		return ""
	}
	return d.String()
}

func params2026(t *testing.T) *taxparams.Params {
	t.Helper()
	sel, err := catalog(t).Select(2026)
	require.NoError(t, err)
	require.False(t, sel.Fallback)
	return sel.Params
}

func lookupTable(p *taxparams.Params, key string, month int, d fiscal.Date) (taxparams.Table, error) {
	if month == 0 {
		return p.Table(taxparams.Key(key), d)
	}
	return p.MonthlyTable(taxparams.Key(key), time.Month(month), d)
}

func engineTable(t *testing.T, p *taxparams.Params, raw golden.RawTable, d fiscal.Date) taxparams.Table {
	t.Helper()
	table, err := lookupTable(p, raw.Key, raw.Month, d)
	require.NoError(t, err, raw.Key)
	return table
}

func findTable(t *testing.T, rp golden.RawParams, key string, month int) golden.RawTable {
	t.Helper()
	for _, tb := range rp.Tables {
		if tb.Key == key && tb.Month == month {
			return tb
		}
	}
	t.Fatalf("no table %s month %d in %s", key, month, paramsFile)
	return golden.RawTable{}
}

func mustCents(t *testing.T, s string) money.Cents {
	t.Helper()
	c, err := money.ParseCents(s)
	require.NoError(t, err)
	return c
}

func mustDate(t *testing.T, s string) fiscal.Date {
	t.Helper()
	d, err := fiscal.ParseDate(s)
	require.NoError(t, err)
	return d
}
