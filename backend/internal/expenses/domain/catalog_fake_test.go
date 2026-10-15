package domain_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/expenses/domain"
	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
)

// dated is one parameter value with an inclusive validity range; a zero end
// means open.
type dated[T any] struct {
	from, to fiscal.Date
	value    T
}

func inForce[T any](values []dated[T], on fiscal.Date) (T, bool) {
	for _, v := range values {
		if on.Compare(v.from) >= 0 && (v.to.IsZero() || on.Compare(v.to) <= 0) {
			return v.value, true
		}
	}
	var zero T
	return zero, false
}

// fakeCatalog is the test double of the RateCatalog port M5-08 implements.
type fakeCatalog struct {
	rates   map[fiscal.RateKey][]dated[money.Rate]
	amounts map[domain.AmountKey][]dated[money.Cents]
}

func (c fakeCatalog) Rate(key fiscal.RateKey, on fiscal.Date) (money.Rate, bool) {
	return inForce(c.rates[key], on)
}

func (c fakeCatalog) Amount(key domain.AmountKey, on fiscal.Date) (money.Cents, bool) {
	return inForce(c.amounts[key], on)
}

func date(t *testing.T, raw string) fiscal.Date {
	t.Helper()
	d, err := fiscal.ParseDate(raw)
	require.NoError(t, err)
	return d
}

func rate(t *testing.T, raw string) money.Rate {
	t.Helper()
	r, err := money.ParseRate(raw)
	require.NoError(t, err)
	return r
}

// catalog2026 holds the 2026 values: the border 8 % ends on 2026-12-31, the
// cash-payment limit is $2,000.00 (LISR 27 fr. III, ⚠️ CPA).
func catalog2026(t *testing.T) fakeCatalog {
	t.Helper()
	since2022 := date(t, "2022-01-01")
	open := func(r string) []dated[money.Rate] { return []dated[money.Rate]{{from: since2022, value: rate(t, r)}} }
	return fakeCatalog{
		rates: map[fiscal.RateKey][]dated[money.Rate]{
			fiscal.KeyIVAGeneral: open("0.16"),
			fiscal.KeyIVAZero:    open("0"),
			fiscal.KeyIVABorder: {{
				from: since2022, to: date(t, "2026-12-31"), value: rate(t, "0.08"),
			}},
			fiscal.KeyISRWithholdingFees:    open("0.10"),
			fiscal.KeyISRWithholdingRent:    open("0.10"),
			domain.KeyIVAWithholdingFreight: open("0.04"),
		},
		amounts: map[domain.AmountKey][]dated[money.Cents]{
			domain.KeyCashPaymentLimit: {{from: since2022, value: 2_000_00}},
		},
	}
}

// fakeWithholdingCatalog is the test double of fiscal.WithholdingCatalog:
// the dated rates plus the kernel's money parameters.
type fakeWithholdingCatalog struct {
	rates   map[fiscal.RateKey][]dated[money.Rate]
	amounts map[fiscal.AmountKey][]dated[money.Cents]
}

func (c fakeWithholdingCatalog) Rate(key fiscal.RateKey, on fiscal.Date) (money.Rate, bool) {
	return inForce(c.rates[key], on)
}

func (c fakeWithholdingCatalog) Amount(key fiscal.AmountKey, on fiscal.Date) (money.Cents, bool) {
	return inForce(c.amounts[key], on)
}

// withholdings2026 holds the withholding parameters of params/data/2026
// (MODULE-withholdings §3.2) the expense side reads.
func withholdings2026(t *testing.T) fakeWithholdingCatalog {
	t.Helper()
	since2022 := date(t, "2022-01-01")
	open := func(r string) []dated[money.Rate] { return []dated[money.Rate]{{from: since2022, value: rate(t, r)}} }
	return fakeWithholdingCatalog{
		rates: map[fiscal.RateKey][]dated[money.Rate]{
			fiscal.KeyISRWithholdingFees:    open("0.10"),
			fiscal.KeyISRWithholdingRent:    open("0.10"),
			fiscal.KeyISRWithholdingRESICO:  open("0.0125"),
			fiscal.KeyIVAWithholdingFreight: open("0.04"),
		},
		amounts: map[fiscal.AmountKey][]dated[money.Cents]{
			fiscal.KeyIVAFederalWithholdingThreshold: {{from: since2022, value: 2_000_00}},
		},
	}
}
