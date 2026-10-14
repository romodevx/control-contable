package domain_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain"
)

// consolidatedShown renders the consolidated lines as "code op amount".
func consolidatedShown(lines []domain.Line) []string {
	out := make([]string, 0, len(lines))
	for _, l := range lines {
		out = append(out, fmt.Sprintf("%s %s %s", l.Code, l.Operation, l.Amount))
	}
	return out
}

// The research's PF-612-IVA-01 components (March) and the 606 side of
// TX-NET-01 (commercial rent to a PM with creditable IVA 3,000.00).
var (
	tx612IVA = domain.IVAComponents{ChargedCollected: 4_800_00, CreditablePaid: 1_600_00, WithheldFromTaxpayer: 1_066_67}
	tx606IVA = domain.IVAComponents{ChargedCollected: 3_200_00, CreditablePaid: 3_000_00, WithheldFromTaxpayer: 2_133_33}
)

// One regime: the consolidated IVA is its own net, and no cross-regime
// warning is raised (MODULE-tax-engine §2.6).
func TestConsolidateIVAOneRegime(t *testing.T) {
	got, err := domain.ConsolidateIVA([]domain.IVAContribution{{Regime: fiscal.RegimeBusiness, Components: tx612IVA}}, 0)

	require.NoError(t, err)
	assert.Equal(t, money.Cents(2_133_33), got.Result.Amount)
	assert.Equal(t, domain.OutcomePayable, got.Result.Outcome())
	assert.Equal(t, 1, got.Contributors)
	assert.Empty(t, got.CPATopics)
	assert.Empty(t, got.Disclaimers)
	assert.Equal(t, []string{
		"iva.total.charged input 4800.00",
		"iva.total.creditable subtract 1600.00",
		"iva.total.withheld subtract 1066.67",
		"iva.consolidated result 2133.33",
	}, consolidatedShown(got.Lines))
	for _, l := range got.Lines {
		assert.Equal(t, domain.TaxIVA, l.Tax, l.Code)
		assert.NotEmpty(t, l.Basis, "%s needs a legal basis", l.Code)
		assert.False(t, l.NeedsCPAReview, l.Code)
	}
}

// TX-NET-01 (OQ-T2): 612 (+2,133.33) and 606 (−1,933.33) net into one IVA
// of 200.00 payable; two contributing regimes put the consolidation line
// under CPA review (cross_regime_iva_netting).
func TestConsolidateIVANetsAcrossRegimesTXNET01(t *testing.T) {
	got, err := domain.ConsolidateIVA([]domain.IVAContribution{
		{Regime: fiscal.RegimeBusiness, Components: tx612IVA},
		{Regime: fiscal.RegimeRent, Components: tx606IVA},
	}, 0)

	require.NoError(t, err)
	assert.Equal(t, money.Cents(200_00), got.Result.Amount)
	assert.Equal(t, 2, got.Contributors)
	assert.Equal(t, []domain.CPATopic{domain.TopicCrossRegimeIVANetting}, got.CPATopics)
	assert.Equal(t, []string{
		"iva.total.charged input 8000.00",
		"iva.total.creditable subtract 4600.00",
		"iva.total.withheld subtract 3200.00",
		"iva.consolidated result 200.00",
	}, consolidatedShown(got.Lines))
	result := got.Lines[len(got.Lines)-1]
	assert.True(t, result.NeedsCPAReview, "the consolidated line carries the ⚠️")
}

// A regime with all-zero components does not count as contributing: no
// cross-regime warning for a 612 with a silent 606.
func TestConsolidateIVAZeroComponentsDoNotContribute(t *testing.T) {
	got, err := domain.ConsolidateIVA([]domain.IVAContribution{
		{Regime: fiscal.RegimeBusiness, Components: tx612IVA},
		{Regime: fiscal.RegimeRent},
	}, 0)

	require.NoError(t, err)
	assert.Equal(t, 1, got.Contributors)
	assert.Empty(t, got.CPATopics)
	assert.False(t, got.Lines[len(got.Lines)-1].NeedsCPAReview)
}

// A negative net is a saldo a favor: in_favor, never payable (US-X1).
func TestConsolidateIVAInFavor(t *testing.T) {
	got, err := domain.ConsolidateIVA([]domain.IVAContribution{{Regime: fiscal.RegimeRent, Components: tx606IVA}}, 0)

	require.NoError(t, err)
	assert.Equal(t, money.Cents(-1_933_33), got.Result.Amount)
	assert.Equal(t, domain.OutcomeInFavor, got.Result.Outcome())
	assert.Zero(t, got.Result.Payable())
}

// No regime contributes (informational only, or no regime held): zero IVA
// and no lines.
func TestConsolidateIVANothingToConsolidate(t *testing.T) {
	got, err := domain.ConsolidateIVA(nil, 0)

	require.NoError(t, err)
	assert.Equal(t, domain.OutcomeZero, got.Result.Outcome())
	assert.Empty(t, got.Lines)
	assert.Zero(t, got.Contributors)
}

// The 625 definitive exclusion is subtracted on its own ⚠️ line.
func TestConsolidateIVADefinitiveExclusion(t *testing.T) {
	platform := domain.IVAComponents{ChargedCollected: 4_800_00, DefinitiveExcluded: 2_400_00, WithheldFromTaxpayer: 2_400_00}

	got, err := domain.ConsolidateIVA([]domain.IVAContribution{{Regime: fiscal.RegimePlatforms, Components: platform}}, 0)

	require.NoError(t, err)
	assert.Equal(t, []string{
		"iva.total.charged input 4800.00",
		"iva.total.definitive_excluded subtract 2400.00",
		"iva.total.creditable subtract 0.00",
		"iva.total.withheld subtract 2400.00",
		"iva.consolidated result 0.00",
	}, consolidatedShown(got.Lines))
	assert.True(t, got.Lines[1].NeedsCPAReview)
}

// OQ-T5: a prior month's saldo a favor is not carried forward; an info
// line says so, and the result does not change.
func TestConsolidateIVAPreviousInFavorIsNotCarried(t *testing.T) {
	got, err := domain.ConsolidateIVA([]domain.IVAContribution{{Regime: fiscal.RegimeBusiness, Components: tx612IVA}}, 500_00)

	require.NoError(t, err)
	assert.Equal(t, money.Cents(2_133_33), got.Result.Amount)
	last := got.Lines[len(got.Lines)-1]
	assert.Equal(t, domain.LineIVAPreviousInFavor, last.Code)
	assert.Equal(t, domain.OpInfo, last.Operation)
	assert.Equal(t, money.Cents(500_00), last.Amount)
	assert.NotEmpty(t, last.Basis, "cites LIVA 6")
}

// PR #79 follow-up: a taxed regime with creditable IVA and another regime
// with exempt collections are mixed activity at the taxpayer level.
func TestConsolidateIVAMixedActivityAcrossRegimes(t *testing.T) {
	tests := []struct {
		name     string
		contribs []domain.IVAContribution
		want     []domain.DisclaimerCode
	}{
		{
			name: "612 credits, 606 residential rent is exempt",
			contribs: []domain.IVAContribution{
				{Regime: fiscal.RegimeBusiness, Components: tx612IVA},
				{Regime: fiscal.RegimeRent, ExemptBase: 15_000_00},
			},
			want: []domain.DisclaimerCode{domain.DisclaimerIVAMixedActivity},
		},
		{
			name:     "exempt collections, nothing credited",
			contribs: []domain.IVAContribution{{Regime: fiscal.RegimeRent, ExemptBase: 15_000_00}},
		},
		{
			name:     "credits, no exempt act",
			contribs: []domain.IVAContribution{{Regime: fiscal.RegimeBusiness, Components: tx612IVA}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := domain.ConsolidateIVA(tt.contribs, 0)

			require.NoError(t, err)
			assert.Equal(t, tt.want, got.Disclaimers)
		})
	}
}

func TestConsolidateIVAOverflowFails(t *testing.T) {
	huge := domain.IVAComponents{ChargedCollected: money.Cents(1 << 62)}

	_, err := domain.ConsolidateIVA([]domain.IVAContribution{
		{Regime: fiscal.RegimeBusiness, Components: huge}, {Regime: fiscal.RegimeRent, Components: huge},
	}, 0)

	assert.ErrorIs(t, err, money.ErrOverflow)
}
