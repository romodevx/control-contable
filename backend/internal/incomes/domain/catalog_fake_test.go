package domain_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/incomes/domain"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
)

// datedRate is one parameter value with an inclusive validity range; a zero
// end means open.
type datedRate struct {
	from, to domain.Date
	rate     money.Rate
}

// fakeCatalog is the test double of the RateCatalog port M5-08 implements:
// a key maps to date-ranged values, as ADR-0002 §3 params do.
type fakeCatalog map[domain.RateKey][]datedRate

func (c fakeCatalog) Rate(key domain.RateKey, on domain.Date) (money.Rate, bool) {
	for _, v := range c[key] {
		if on.Compare(v.from) >= 0 && (v.to.IsZero() || on.Compare(v.to) <= 0) {
			return v.rate, true
		}
	}
	return 0, false
}

// date parses a YYYY-MM-DD literal or fails the test.
func date(t *testing.T, raw string) domain.Date {
	t.Helper()
	d, err := domain.ParseDate(raw)
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
		domain.KeyIVAGeneral: open("0.16"),
		domain.KeyIVAZero:    open("0"),
		domain.KeyIVABorder: {{
			from: date(t, "2022-01-01"), to: date(t, "2026-12-31"), rate: rate(t, "0.08"),
		}},
		domain.KeyISRWithholdingFees:      open("0.10"),
		domain.KeyISRWithholdingRent:      open("0.10"),
		domain.KeyISRWithholdingRESICO:    open("0.0125"),
		domain.KeyISRPlatformTransport:    open("0.021"),
		domain.KeyISRPlatformLodging:      open("0.04"),
		domain.KeyISRPlatformGoods:        open("0.025"),
		domain.KeyISRPlatformNoRFC:        open("0.20"),
		domain.KeyISRPlatformPersonaMoral: open("0.025"),
	}
}
