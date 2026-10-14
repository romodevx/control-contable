package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/platform/money"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain"
)

func isr(c money.Cents) *domain.TaxResult {
	r := domain.NewTaxResult(money.Exact(c) * money.ExactPerCent)
	return &r
}

func ivaResult(c money.Cents) domain.TaxResult {
	return domain.NewTaxResult(money.Exact(c) * money.ExactPerCent)
}

// The totals rule (ADR-0002 §1, PR #89): each amount counts once. ISR sums
// the payable per-regime ISR (a favor balance is not offset against another
// regime, OQ-T10), IVA is the consolidated result, withholdings to remit
// sum each regime's own, and the total adds the three payable parts.
func TestNewTotals(t *testing.T) {
	tests := []struct {
		name      string
		regimes   []domain.RegimeEstimate
		iva       domain.TaxResult
		wantISR   money.Cents
		wantRemit money.Cents
		wantTotal money.Cents
	}{
		{
			name:    "TX-612-TOTAL-01: ISR 150.96 + IVA 2,133.33",
			regimes: []domain.RegimeEstimate{{ISR: isr(150_96)}},
			iva:     ivaResult(2_133_33), wantISR: 150_96, wantTotal: 2_284_29,
		},
		{
			name:    "OQ-T10: 606 in favor 924.00 does not offset 612",
			regimes: []domain.RegimeEstimate{{ISR: isr(150_96)}, {ISR: isr(-924_00)}},
			iva:     ivaResult(200_00), wantISR: 150_96, wantTotal: 350_96,
		},
		{
			name: "TX-601-REMIT-01: to remit once, never in ISR or IVA",
			regimes: []domain.RegimeEstimate{{
				ISR: isr(36_000_00), IVA: &domain.IVAComponents{ChargedCollected: 320_000_00}, WithheldToRemit: 10_333_33,
			}},
			iva: ivaResult(120_000_00), wantISR: 36_000_00, wantRemit: 10_333_33, wantTotal: 166_333_33,
		},
		{
			name:    "IVA in favor is not payable",
			regimes: []domain.RegimeEstimate{{ISR: isr(100_00)}},
			iva:     ivaResult(-500_00), wantISR: 100_00, wantTotal: 100_00,
		},
		{
			name:    "no ISR figure (partial, informational)",
			regimes: []domain.RegimeEstimate{{Coverage: domain.TreatmentPartial}, {Coverage: domain.TreatmentInformational}},
			iva:     ivaResult(0),
		},
		{name: "no regime held", iva: ivaResult(0)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := domain.NewTotals(tt.regimes, tt.iva)

			require.NoError(t, err)
			assert.Equal(t, tt.wantISR, got.ISR.Amount)
			assert.Equal(t, tt.wantISR, got.ISR.Payable())
			assert.Equal(t, tt.iva, got.IVA)
			assert.Equal(t, tt.wantRemit, got.WithholdingsToRemit)
			assert.Equal(t, tt.wantTotal, got.TotalPayable)
		})
	}
}

func TestNewTotalsOverflowFails(t *testing.T) {
	huge := money.Cents(1 << 62)

	_, err := domain.NewTotals([]domain.RegimeEstimate{{WithheldToRemit: huge}, {WithheldToRemit: huge}}, ivaResult(0))

	assert.ErrorIs(t, err, money.ErrOverflow)
}
