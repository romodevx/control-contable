package domain_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/incomes/domain"
	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
)

// datedRate is one parameter value with an inclusive validity range; a zero
// end means open.
type datedRate struct {
	from, to fiscal.Date
	rate     money.Rate
}

// fakeCatalog is the test double of the RateCatalog port M5-08 implements:
// a key maps to date-ranged values, as ADR-0002 §3 params do.
type fakeCatalog map[fiscal.RateKey][]datedRate

func (c fakeCatalog) Rate(key fiscal.RateKey, on fiscal.Date) (money.Rate, bool) {
	for _, v := range c[key] {
		if on.Compare(v.from) >= 0 && (v.to.IsZero() || on.Compare(v.to) <= 0) {
			return v.rate, true
		}
	}
	return 0, false
}

// date parses a YYYY-MM-DD literal or fails the test.
func date(t *testing.T, raw string) fiscal.Date {
	t.Helper()
	d, err := fiscal.ParseDate(raw)
	require.NoError(t, err)
	return d
}

// rate parses a decimal rate literal or fails the test.
func rate(t *testing.T, raw string) money.Rate {
	t.Helper()
	r, err := money.ParseRate(raw)
	require.NoError(t, err)
	return r
}

// catalog2026 holds the 2026 values of research §0: the border 8 % ends on
// 2026-12-31 (DOF 31-12-2025) and platforms withhold 2.5 % on goods (LIF 2026).
func catalog2026(t *testing.T) fakeCatalog {
	t.Helper()
	open := func(r string) []datedRate {
		return []datedRate{{from: date(t, "2022-01-01"), rate: rate(t, r)}}
	}
	return fakeCatalog{
		fiscal.KeyIVAGeneral: open("0.16"),
		fiscal.KeyIVAZero:    open("0"),
		fiscal.KeyIVABorder: {{
			from: date(t, "2022-01-01"), to: date(t, "2026-12-31"), rate: rate(t, "0.08"),
		}},
		fiscal.KeyISRWithholdingFees:      open("0.10"),
		fiscal.KeyISRWithholdingRent:      open("0.10"),
		domain.KeyISRWithholdingRESICO:    open("0.0125"),
		domain.KeyISRPlatformTransport:    open("0.021"),
		domain.KeyISRPlatformLodging:      open("0.04"),
		domain.KeyISRPlatformGoods:        open("0.025"),
		domain.KeyISRPlatformNoRFC:        open("0.20"),
		domain.KeyISRPlatformPersonaMoral: open("0.025"),
	}
}

// fakeWithholdingCatalog adds the dated money parameters DeriveWithholdings
// reads (fiscal.WithholdingCatalog) to fakeCatalog.
type fakeWithholdingCatalog struct {
	fakeCatalog
	amounts map[fiscal.AmountKey]money.Cents
}

func (c fakeWithholdingCatalog) Amount(key fiscal.AmountKey, _ fiscal.Date) (money.Cents, bool) {
	v, ok := c.amounts[key]
	return v, ok
}

// withholdingCatalog2026 is catalog2026 plus the WH-05 withholding keys of
// params/data/2026/v1.yaml: freight 4 %, the platform no-RFC rates and the
// RMF 4.1.3 federal threshold of 2,000.00.
func withholdingCatalog2026(t *testing.T) fakeWithholdingCatalog {
	t.Helper()
	rates := catalog2026(t)
	open := func(r string) []datedRate {
		return []datedRate{{from: date(t, "2022-01-01"), rate: rate(t, r)}}
	}
	rates[fiscal.KeyIVAWithholdingFreight] = open("0.04")
	rates[fiscal.KeyISRPlatformPersonaMoralNoRFC] = open("0.20")
	return fakeWithholdingCatalog{
		fakeCatalog: rates,
		amounts:     map[fiscal.AmountKey]money.Cents{fiscal.KeyIVAFederalWithholdingThreshold: 2_000_00},
	}
}
