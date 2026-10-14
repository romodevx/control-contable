package taxparams_test

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain/taxparams"
)

// tableIndex finds the first table of key in spec, so a case can mutate it.
func tableIndex(t testing.TB, spec taxparams.Spec, key taxparams.Key, month time.Month) int {
	t.Helper()
	for i, tbl := range spec.Tables {
		if tbl.Key == key && tbl.Month == month {
			return i
		}
	}
	t.Fatalf("fixture has no table %s month %d", key, month)
	return -1
}

// valueIndex finds the first value of key in spec, so a case can mutate it.
func valueIndex(t testing.TB, spec taxparams.Spec, key taxparams.Key) int {
	t.Helper()
	for i, v := range spec.Values {
		if v.Key == key {
			return i
		}
	}
	t.Fatalf("fixture has no value %s", key)
	return -1
}

func TestNewParamsAcceptsTheValidFixture(t *testing.T) {
	p := mustParams(t, validSpec(t))
	assert.Equal(t, 2026, p.FiscalYear())
	assert.Equal(t, 1, p.Version())
	assert.False(t, p.Provisional())
	assert.Equal(t, "2026-10-13", p.PublishedOn().String())
	require.Len(t, p.Sources(), 1)
	assert.Equal(t, sourceID, p.Sources()[0].ID)
}

func TestNewParamsRejects(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(t testing.TB, s *taxparams.Spec)
		want   error
	}{
		// Header.
		{"fiscal year before 2000", func(_ testing.TB, s *taxparams.Spec) { s.FiscalYear = 1999 }, taxparams.ErrInvalidHeader},
		{"fiscal year past 9999", func(_ testing.TB, s *taxparams.Spec) { s.FiscalYear = 10000 }, taxparams.ErrInvalidHeader},
		{"version zero", func(_ testing.TB, s *taxparams.Spec) { s.Version = 0 }, taxparams.ErrInvalidHeader},
		{"no published_on", func(_ testing.TB, s *taxparams.Spec) { s.PublishedOn = fiscal.Date{} }, taxparams.ErrInvalidHeader},

		// Sources.
		{"no sources", func(_ testing.TB, s *taxparams.Spec) { s.Sources = nil }, taxparams.ErrInvalidSource},
		{"source without id", func(_ testing.TB, s *taxparams.Spec) { s.Sources[0].ID = "" }, taxparams.ErrInvalidSource},
		{"source without title", func(_ testing.TB, s *taxparams.Spec) { s.Sources[0].Title = " " }, taxparams.ErrInvalidSource},
		{"source without reference", func(_ testing.TB, s *taxparams.Spec) { s.Sources[0].Reference = "" }, taxparams.ErrInvalidSource},
		{"source without dof date", func(_ testing.TB, s *taxparams.Spec) { s.Sources[0].DOFDate = fiscal.Date{} }, taxparams.ErrInvalidSource},
		{"source over plain http", func(_ testing.TB, s *taxparams.Spec) { s.Sources[0].URL = "http://example.test/dof" }, taxparams.ErrInvalidSource},
		{"source url without host", func(_ testing.TB, s *taxparams.Spec) { s.Sources[0].URL = "https:///dof" }, taxparams.ErrInvalidSource},
		{"source url unparsable", func(_ testing.TB, s *taxparams.Spec) { s.Sources[0].URL = "https://exa mple.test/%zz" }, taxparams.ErrInvalidSource},
		{"duplicate source id", func(_ testing.TB, s *taxparams.Spec) { s.Sources = append(s.Sources, s.Sources[0]) }, taxparams.ErrInvalidSource},

		// Values.
		{"value key not dotted", func(_ testing.TB, s *taxparams.Spec) { s.Values[0].Key = "iva" }, taxparams.ErrInvalidKey},
		{"value key upper case", func(_ testing.TB, s *taxparams.Spec) { s.Values[0].Key = "IVA.rate" }, taxparams.ErrInvalidKey},
		{"value key empty segment", func(_ testing.TB, s *taxparams.Spec) { s.Values[0].Key = "iva..rate" }, taxparams.ErrInvalidKey},
		{"value key trailing dot", func(_ testing.TB, s *taxparams.Spec) { s.Values[0].Key = "iva.rate." }, taxparams.ErrInvalidKey},
		{"value of unknown kind", func(_ testing.TB, s *taxparams.Spec) { s.Values[0].Kind = 0 }, taxparams.ErrInvalidKind},
		{"count of zero", func(t testing.TB, s *taxparams.Spec) { s.Values[valueIndex(t, *s, keyCapUMAs)].Count = 0 }, taxparams.ErrCountNotPositive},
		{"negative count", func(t testing.TB, s *taxparams.Spec) { s.Values[valueIndex(t, *s, keyCapUMAs)].Count = -5 }, taxparams.ErrCountNotPositive},
		{"date value without a date", func(t testing.TB, s *taxparams.Spec) {
			s.Values[valueIndex(t, *s, keyDeadline)].Date = fiscal.Date{}
		}, taxparams.ErrMissingDate},
		{"rate above one", func(t testing.TB, s *taxparams.Spec) { s.Values[0].Rate = mustRate(t, "1.000001") }, taxparams.ErrRateOutOfRange},
		{"negative rate", func(t testing.TB, s *taxparams.Spec) { s.Values[0].Rate = mustRate(t, "-0.01") }, taxparams.ErrRateOutOfRange},
		{"negative money", func(t testing.TB, s *taxparams.Spec) { s.Values[1].Money = mustCents(t, "-0.01") }, taxparams.ErrNegativeMoney},
		{"value without start", func(_ testing.TB, s *taxparams.Spec) { s.Values[0].Validity.From = fiscal.Date{} }, taxparams.ErrInvalidValidity},
		{"value ends before it starts", func(t testing.TB, s *taxparams.Spec) {
			s.Values[0].Validity = taxparams.Validity{From: mustDate(t, "2026-02-01"), To: mustDate(t, "2026-01-31")}
		}, taxparams.ErrInvalidValidity},
		{"value without basis", func(_ testing.TB, s *taxparams.Spec) { s.Values[0].Basis = nil }, taxparams.ErrInvalidBasis},
		{"basis citing an unknown source", func(_ testing.TB, s *taxparams.Spec) { s.Values[0].Basis[0].Source = "nope" }, taxparams.ErrInvalidBasis},
		{"basis without article", func(_ testing.TB, s *taxparams.Spec) { s.Values[0].Basis[0].Article = "" }, taxparams.ErrInvalidBasis},
		{"same key as rate and money", func(_ testing.TB, s *taxparams.Spec) {
			s.Values[2].Kind = taxparams.KindRate
		}, taxparams.ErrKindConflict},
		{"overlapping closed ranges", func(t testing.TB, s *taxparams.Spec) {
			s.Values[2].Validity.From = mustDate(t, "2026-01-31") // the January UMA ends that day
		}, taxparams.ErrOverlap},
		{"open range followed by another", func(_ testing.TB, s *taxparams.Spec) {
			s.Values[1].Validity.To = fiscal.Date{}
		}, taxparams.ErrOverlap},
		{"identical entries", func(_ testing.TB, s *taxparams.Spec) {
			s.Values = append(s.Values, s.Values[0])
		}, taxparams.ErrOverlap},

		// Tables.
		{"table key invalid", func(_ testing.TB, s *taxparams.Spec) { s.Tables[0].Key = "Tariff" }, taxparams.ErrInvalidKey},
		{"table key reused by a value", func(_ testing.TB, s *taxparams.Spec) { s.Tables[0].Key = keyUMA }, taxparams.ErrKindConflict},
		{"table of unknown kind", func(_ testing.TB, s *taxparams.Spec) { s.Tables[0].Kind = 0 }, taxparams.ErrInvalidKind},
		{"same table key marginal and flat", func(_ testing.TB, s *taxparams.Spec) {
			extra := s.Tables[1]
			extra.Kind = taxparams.TableMarginal
			extra.Validity.From = mustDateUnchecked("2030-01-01")
			s.Tables[1].Validity.To = mustDateUnchecked("2029-12-31")
			s.Tables = append(s.Tables, extra)
		}, taxparams.ErrKindConflict},
		{"table without basis", func(_ testing.TB, s *taxparams.Spec) { s.Tables[0].Basis = nil }, taxparams.ErrInvalidBasis},
		{"table without start", func(_ testing.TB, s *taxparams.Spec) { s.Tables[0].Validity.From = fiscal.Date{} }, taxparams.ErrInvalidValidity},
		{"table without rows", func(_ testing.TB, s *taxparams.Spec) { s.Tables[0].Rows = nil }, taxparams.ErrInvalidTable},
		{"first row not at one centavo", func(t testing.TB, s *taxparams.Spec) { s.Tables[0].Rows[0].Lower = mustCents(t, "0.00") }, taxparams.ErrInvalidTable},
		{"row upper below its lower", func(t testing.TB, s *taxparams.Spec) {
			s.Tables[1].Rows[1].Upper = mustCents(t, "100.00")
		}, taxparams.ErrInvalidTable},
		{"gap between rows", func(t testing.TB, s *taxparams.Spec) { s.Tables[0].Rows[1].Lower = mustCents(t, "100.02") }, taxparams.ErrNotContiguous},
		{"rows overlap", func(t testing.TB, s *taxparams.Spec) { s.Tables[0].Rows[1].Lower = mustCents(t, "100.00") }, taxparams.ErrNotContiguous},
		{"unbounded row before the last", func(_ testing.TB, s *taxparams.Spec) { s.Tables[1].Rows[0].Unbounded = true }, taxparams.ErrInvalidTable},
		{"marginal tariff with a bounded last row", func(t testing.TB, s *taxparams.Spec) {
			last := &s.Tables[0].Rows[1]
			last.Unbounded, last.Upper = false, mustCents(t, "200.00")
		}, taxparams.ErrInvalidTable},
		{"unbounded row with an upper bound", func(t testing.TB, s *taxparams.Spec) { s.Tables[0].Rows[1].Upper = mustCents(t, "999.99") }, taxparams.ErrInvalidTable},
		{"row rate above one", func(t testing.TB, s *taxparams.Spec) { s.Tables[0].Rows[0].Rate = mustRate(t, "1.5") }, taxparams.ErrRateOutOfRange},
		{"row rate negative", func(t testing.TB, s *taxparams.Spec) { s.Tables[1].Rows[1].Rate = money.Rate(-1) }, taxparams.ErrRateOutOfRange},
		{"negative fixed fee", func(t testing.TB, s *taxparams.Spec) { s.Tables[0].Rows[1].FixedFee = mustCents(t, "-1.00") }, taxparams.ErrNegativeMoney},
		{"fixed fee on a flat table", func(t testing.TB, s *taxparams.Spec) { s.Tables[1].Rows[0].FixedFee = mustCents(t, "1.00") }, taxparams.ErrInvalidTable},
		{"overlapping table versions", func(_ testing.TB, s *taxparams.Spec) { s.Tables = append(s.Tables, s.Tables[0]) }, taxparams.ErrOverlap},

		// Monthly tables.
		{"month 13", func(t testing.TB, s *taxparams.Spec) { s.Tables[tableIndex(t, *s, keyAccum, time.December)].Month = 13 }, taxparams.ErrInvalidMonth},
		{"negative month", func(t testing.TB, s *taxparams.Spec) { s.Tables[tableIndex(t, *s, keyAccum, time.March)].Month = -1 }, taxparams.ErrInvalidMonth},
		{"eleven monthly tables", func(t testing.TB, s *taxparams.Spec) {
			i := tableIndex(t, *s, keyAccum, time.December)
			s.Tables = append(s.Tables[:i], s.Tables[i+1:]...)
		}, taxparams.ErrMonthlyIncomplete},
		{"a month published twice", func(t testing.TB, s *taxparams.Spec) {
			s.Tables[tableIndex(t, *s, keyAccum, time.December)].Month = time.November
		}, taxparams.ErrOverlap},
		{"monthly tables with misaligned validity", func(t testing.TB, s *taxparams.Spec) {
			s.Tables[tableIndex(t, *s, keyAccum, time.June)].Validity = year2026(t)
		}, taxparams.ErrMonthlyIncomplete},
		{"monthly and plain tables under one key", func(t testing.TB, s *taxparams.Spec) {
			s.Tables[tableIndex(t, *s, keyMonthly, 0)].Key = keyAccum
		}, taxparams.ErrMonthlyMixed},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spec := validSpec(t)
			tt.mutate(t, &spec)
			p, err := taxparams.NewParams(spec)
			require.ErrorIs(t, err, tt.want)
			assert.Nil(t, p)
		})
	}
}

// mustDateUnchecked is for fixture literals that are valid by inspection.
func mustDateUnchecked(s string) fiscal.Date {
	d, err := fiscal.ParseDate(s)
	if err != nil {
		panic(err)
	}
	return d
}

func TestNewParamsReportsEveryProblemAtOnce(t *testing.T) {
	spec := validSpec(t)
	spec.Version = 0
	spec.Values[0].Rate = mustRate(t, "2")
	spec.Tables[0].Rows[1].Lower = mustCents(t, "500.00")

	_, err := taxparams.NewParams(spec)

	require.Error(t, err)
	for _, want := range []error{taxparams.ErrInvalidHeader, taxparams.ErrRateOutOfRange, taxparams.ErrNotContiguous} {
		assert.ErrorIs(t, err, want)
	}
	assert.Contains(t, err.Error(), "values[0] iva.rate.border_8")
	assert.Contains(t, err.Error(), "tables[0] isr.art96.monthly rows[1]")
}

// The kernel's shared rate keys (ADR-0007) must be keys a params file can
// define, or M5-08 could never resolve them.
func TestNewParamsAcceptsTheKernelRateKeys(t *testing.T) {
	keys := []fiscal.RateKey{
		fiscal.KeyIVAGeneral, fiscal.KeyIVABorder, fiscal.KeyIVAZero,
		fiscal.KeyISRWithholdingFees, fiscal.KeyISRWithholdingRent,
	}
	for _, key := range keys {
		t.Run(string(key), func(t *testing.T) {
			spec := validSpec(t)
			spec.Values[0].Key = taxparams.Key(key)
			p, err := taxparams.NewParams(spec)
			require.NoError(t, err)
			_, err = p.Rate(taxparams.Key(key), mustDate(t, "2026-06-30"))
			require.NoError(t, err)
		})
	}
}

func TestNewParamsAcceptsBoundaryRates(t *testing.T) {
	spec := validSpec(t)
	spec.Values[0].Rate = 0
	spec.Tables[0].Rows[1].Rate = money.RateScale
	_, err := taxparams.NewParams(spec)
	require.NoError(t, err)
}

func TestNewParamsAcceptsAMidYearTableReplacement(t *testing.T) {
	spec := validSpec(t)
	i := tableIndex(t, spec, keyMonthly, 0)
	replacement := spec.Tables[i]
	replacement.Validity = taxparams.Validity{From: mustDate(t, "2026-07-01")}
	replacement.Rows = []taxparams.Row{{Lower: mustCents(t, "0.01"), Unbounded: true, Rate: mustRate(t, "0.3")}}
	spec.Tables[i].Validity.To = mustDate(t, "2026-06-30")
	spec.Tables = append(spec.Tables, replacement)

	p := mustParams(t, spec)

	june, err := p.Table(keyMonthly, mustDate(t, "2026-06-30"))
	require.NoError(t, err)
	assert.Len(t, june.Rows, 2)
	july, err := p.Table(keyMonthly, mustDate(t, "2026-07-01"))
	require.NoError(t, err)
	assert.Len(t, july.Rows, 1)
}

func TestErrorStrings(t *testing.T) {
	assert.Equal(t, "taxparams: validity ranges overlap", taxparams.ErrOverlap.Error())
	assert.True(t, errors.Is(taxparams.ErrOverlap, taxparams.ErrOverlap))
}
