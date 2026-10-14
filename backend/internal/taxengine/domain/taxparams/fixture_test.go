package taxparams_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain/taxparams"
)

// The fixture is synthetic on purpose: the real 2026 values are transcribed
// (and double-checked) in params/data/2026/v1.yaml by M5-04, never here.
const (
	keyBorderRate = taxparams.Key("iva.rate.border_8")
	keyUMA        = taxparams.Key("uma.monthly")
	keyRounding   = taxparams.Key("rounding.payment_whole_pesos")
	keyMonthly    = taxparams.Key("isr.art96.monthly")
	keyBrackets   = taxparams.Key("isr.resico_pf.monthly")
	keyAccum      = taxparams.Key("isr.art106.accumulated")
	keyCapUMAs    = taxparams.Key("deduction.cap_umas")
	keyDeadline   = taxparams.Key("annual.deadline.pf")
	sourceID      = "dof-synthetic"
)

func mustDate(t testing.TB, s string) fiscal.Date {
	t.Helper()
	d, err := fiscal.ParseDate(s)
	require.NoError(t, err)
	return d
}

func mustRate(t testing.TB, s string) money.Rate {
	t.Helper()
	r, err := money.ParseRate(s)
	require.NoError(t, err)
	return r
}

func mustCents(t testing.TB, s string) money.Cents {
	t.Helper()
	c, err := money.ParseCents(s)
	require.NoError(t, err)
	return c
}

func basis() []taxparams.Citation {
	return []taxparams.Citation{{Source: sourceID, Article: "Art. 1"}}
}

func year2026(t testing.TB) taxparams.Validity {
	return taxparams.Validity{From: mustDate(t, "2026-01-01"), To: mustDate(t, "2026-12-31")}
}

func from2026(t testing.TB) taxparams.Validity {
	return taxparams.Validity{From: mustDate(t, "2026-01-01")}
}

// accumulatedTable is month n of a synthetic accumulated tariff: every
// bound grows with n, like the published art. 106 tables, so a lookup that
// picks the wrong month is visible.
func accumulatedTable(t testing.TB, month time.Month) taxparams.Table {
	n := int(month)
	return taxparams.Table{
		Key: keyAccum, Kind: taxparams.TableMarginal, Month: month, Validity: from2026(t), Basis: basis(),
		Rows: []taxparams.Row{
			{Lower: mustCents(t, "0.01"), Upper: mustCents(t, fmt.Sprintf("%d.00", 100*n)), Rate: mustRate(t, "0.02")},
			{Lower: mustCents(t, fmt.Sprintf("%d.01", 100*n)), Unbounded: true, FixedFee: mustCents(t, fmt.Sprintf("%d.00", 2*n)), Rate: mustRate(t, "0.1")},
		},
	}
}

// validSpec is a complete, valid synthetic parameter set: dated values
// (a sunset rate, a mid-year money change, an open-ended flag, a count and
// a dated deadline), a plain
// marginal tariff, non-marginal brackets and twelve monthly tables.
func validSpec(t testing.TB) taxparams.Spec {
	t.Helper()
	spec := taxparams.Spec{
		FiscalYear:  2026,
		Version:     1,
		PublishedOn: mustDate(t, "2026-10-13"),
		Sources: []taxparams.Source{{
			ID: sourceID, Title: "Synthetic decree", URL: "https://example.test/dof/synthetic",
			DOFDate: mustDate(t, "2025-12-31"), Reference: "Art. 1",
		}},
		Values: []taxparams.Value{
			{Key: keyBorderRate, Kind: taxparams.KindRate, Rate: mustRate(t, "0.08"), Validity: year2026(t), Basis: basis()},
			{Key: keyUMA, Kind: taxparams.KindMoney, Money: mustCents(t, "100.00"), Basis: basis(),
				Validity: taxparams.Validity{From: mustDate(t, "2026-01-01"), To: mustDate(t, "2026-01-31")}},
			{Key: keyUMA, Kind: taxparams.KindMoney, Money: mustCents(t, "110.00"), Basis: basis(),
				Validity: taxparams.Validity{From: mustDate(t, "2026-02-01"), To: mustDate(t, "2027-01-31")}},
			{Key: keyRounding, Kind: taxparams.KindFlag, Flag: true, Validity: from2026(t), Basis: basis()},
			{Key: keyCapUMAs, Kind: taxparams.KindCount, Count: 5, Validity: from2026(t), Basis: basis()},
			{Key: keyDeadline, Kind: taxparams.KindDate, Date: mustDate(t, "2027-04-30"), Validity: year2026(t), Basis: basis()},
		},
		Tables: []taxparams.Table{
			{
				Key: keyMonthly, Kind: taxparams.TableMarginal, Validity: from2026(t), Basis: basis(),
				Rows: []taxparams.Row{
					{Lower: mustCents(t, "0.01"), Upper: mustCents(t, "100.00"), Rate: mustRate(t, "0.02")},
					{Lower: mustCents(t, "100.01"), Unbounded: true, FixedFee: mustCents(t, "2.00"), Rate: mustRate(t, "0.1")},
				},
			},
			{
				Key: keyBrackets, Kind: taxparams.TableFlat, Validity: from2026(t), Basis: basis(),
				Rows: []taxparams.Row{
					{Lower: mustCents(t, "0.01"), Upper: mustCents(t, "100.00"), Rate: mustRate(t, "0.01")},
					{Lower: mustCents(t, "100.01"), Upper: mustCents(t, "200.00"), Rate: mustRate(t, "0.011")},
				},
			},
		},
	}
	for m := time.January; m <= time.December; m++ {
		spec.Tables = append(spec.Tables, accumulatedTable(t, m))
	}
	return spec
}

func mustParams(t testing.TB, spec taxparams.Spec) *taxparams.Params {
	t.Helper()
	p, err := taxparams.NewParams(spec)
	require.NoError(t, err)
	return p
}

// withYear returns a valid spec for another fiscal year and version.
func withYear(t testing.TB, fiscalYear, version int, provisional bool) *taxparams.Params {
	t.Helper()
	spec := validSpec(t)
	spec.FiscalYear, spec.Version, spec.Provisional = fiscalYear, version, provisional
	return mustParams(t, spec)
}
