package ratecatalog_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	expenses "github.com/romodevx/control-contable/backend/internal/expenses/domain"
	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain/taxparams"
	"github.com/romodevx/control-contable/backend/internal/taxengine/infra/ratecatalog"
)

// The view is what the composition root hands fiscal.DeriveWithholdings.
var _ fiscal.WithholdingCatalog = ratecatalog.WithholdingView[expenses.AmountKey]{}

// RMF 2026 rule 4.1.3: the Federación withholds IVA on goods and services
// only above $2,000.00 (docs/reviews/WH-05-params-verification.md).
func TestIVAWithholdingFederalThresholdIn2026(t *testing.T) {
	c := embedded(t)
	for _, on := range []string{"2026-01-01", "2026-12-31"} {
		got, err := c.IVAWithholdingFederalThreshold(date(t, on))
		require.NoError(t, err, on)
		assert.Equal(t, cents(t, "2000.00"), got, on)
	}
	got, ok := c.Withholdings().Amount(fiscal.KeyIVAFederalWithholdingThreshold, date(t, "2026-06-30"))
	require.True(t, ok)
	assert.Equal(t, cents(t, "2000.00"), got)
}

// An RMF rule lives one year: no set covers 2025, and in 2027 the 2026 set
// served as fallback no longer holds it.
func TestIVAWithholdingFederalThresholdOutsideTheRMF2026(t *testing.T) {
	c := embedded(t)
	_, err := c.IVAWithholdingFederalThreshold(date(t, "2025-12-31"))
	require.ErrorIs(t, err, ratecatalog.ErrDateNotCovered)
	_, err = c.IVAWithholdingFederalThreshold(date(t, "2027-01-01"))
	require.ErrorIs(t, err, taxparams.ErrNotInForce)
}

// A set without the threshold fails at boot, as a missing rate does.
func TestNewRequiresTheFederalThreshold(t *testing.T) {
	_, err := newSynthetic(t, syntheticSetWithout(t, 2026, false, taxparams.Key(fiscal.KeyIVAFederalWithholdingThreshold)))
	require.ErrorIs(t, err, ratecatalog.ErrMissingWithholdingRate)
	require.ErrorIs(t, err, taxparams.ErrUnknownKey)
}

// The view reads the synthetic set, not a constant: its threshold differs.
func TestWithholdingViewReadsTheSet(t *testing.T) {
	got, ok := synthetic(t, syntheticSet(t, 2026, false)).Withholdings().
		Amount(fiscal.KeyIVAFederalWithholdingThreshold, date(t, "2026-06-30"))
	require.True(t, ok)
	assert.Equal(t, cents(t, syntheticFederalThreshold), got)
}

// The kernel decision table over the published 2026 parameters: the owner's
// golden (WM WH-01) and both sides of the RMF 4.1.3 threshold.
func TestDeriveWithholdingsOverThe2026Parameters(t *testing.T) {
	view := embedded(t).Withholdings()
	pf := func(regime fiscal.RegimeCode, op fiscal.OperationKind, receptor fiscal.ReceptorKind, subtotal string) fiscal.WithholdingInput {
		s := cents(t, subtotal)
		return fiscal.WithholdingInput{
			EmisorPersona: fiscal.PersonFisica, EmisorRegime: regime, Receptor: receptor, Operation: op,
			Subtotal: s, IVA: s * 16 / 100, IVARate: fiscal.IVARate16, On: date(t, "2026-07-10"),
		}
	}
	tests := []struct {
		name     string
		input    fiscal.WithholdingInput
		isr, iva string
		net      string
		rules    []fiscal.RuleID
	}{
		{
			"WH-01 owner: RESICO PF services to a PM",
			pf(fiscal.RegimeRESICO, fiscal.OperationProfessionalServices, fiscal.ReceptorPM, "100000.00"),
			"1250.00", "10666.67", "104083.33", []fiscal.RuleID{fiscal.RuleISR4, fiscal.RuleIVA6},
		},
		{
			"goods to the Federación of exactly $2,000.00: nothing",
			pf(fiscal.RegimeBusiness, fiscal.OperationSaleOfGoods, fiscal.ReceptorGovFederal, "2000.00"),
			"0", "0", "2320.00", []fiscal.RuleID{fiscal.RuleISR7, fiscal.RuleIVA8},
		},
		{
			"goods to the Federación of $2,000.01: 2/3 (A2)",
			pf(fiscal.RegimeBusiness, fiscal.OperationSaleOfGoods, fiscal.ReceptorGovFederal, "2000.01"),
			"0", "213.33", "2106.68", []fiscal.RuleID{fiscal.RuleISR7, fiscal.RuleIVA7},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := fiscal.DeriveWithholdings(tt.input, view)
			require.NoError(t, err)
			assert.Equal(t, cents(t, tt.isr), got.ISR)
			assert.Equal(t, cents(t, tt.iva), got.IVA)
			assert.Equal(t, cents(t, tt.net), got.Net)
			assert.Equal(t, tt.rules, got.RuleIDs)
		})
	}
}
