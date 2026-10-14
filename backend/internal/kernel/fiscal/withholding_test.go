package fiscal_test

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
)

func known(c money.Cents) fiscal.Suggestion { return fiscal.Suggestion{Cents: c, Known: true} }

// TestShareSuggest: the IVA withholding shares of the law, rounded once
// (research A.5 "2/3 × 1,600 = 1,066.67"; PM-601-IVA-01 5,333.33).
func TestShareSuggest(t *testing.T) {
	tests := []struct {
		name  string
		share fiscal.Share
		iva   money.Cents
		want  fiscal.Suggestion
	}{
		{"US-I1: 2/3 of 1,600", fiscal.ShareTwoThirds, 1_600_00, known(1_066_67)},
		{"PM-601-IVA-01: 2/3 of 8,000", fiscal.ShareTwoThirds, 8_000_00, known(5_333_33)},
		{"PF-625-01: half of 4,800", fiscal.ShareHalf, 4_800_00, known(2_400_00)},
		{"625 without RFC: all of 2,400", fiscal.ShareWhole, 2_400_00, known(2_400_00)},
		{"exempt income: 2/3 of nothing", fiscal.ShareTwoThirds, 0, known(0)},
		{"the zero Share suggests nothing", fiscal.Share{}, 1_600_00, fiscal.Suggestion{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.share.Suggest(tt.iva)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestShareOverflowIsAParameterFault(t *testing.T) {
	_, err := fiscal.ShareTwoThirds.Suggest(math.MaxInt64)
	require.ErrorIs(t, err, fiscal.ErrInvalidParameter)
}

func TestShareIsZero(t *testing.T) {
	assert.True(t, fiscal.Share{}.IsZero())
	assert.False(t, fiscal.ShareTwoThirds.IsZero())
}

// TestSuggestAtRate: ISR (or IVA on freight) at the catalog rate in force on
// the day; no key or no rate that day is no suggestion, never an error.
func TestSuggestAtRate(t *testing.T) {
	tests := []struct {
		name string
		key  fiscal.RateKey
		base money.Cents
		on   string
		want fiscal.Suggestion
	}{
		{"PF-612-ISR-01: 10 % of 30,000 of fees", fiscal.KeyISRWithholdingFees, 30_000_00, "2026-03-31", known(3_000_00)},
		{"PF-606-ISR-02: 10 % of 20,000 of rent", fiscal.KeyISRWithholdingRent, 20_000_00, "2026-06-30", known(2_000_00)},
		{"no key", "", 30_000_00, "2026-03-31", fiscal.Suggestion{}},
		{"no rate that day", fiscal.KeyISRWithholdingFees, 30_000_00, "2021-12-31", fiscal.Suggestion{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := fiscal.SuggestAtRate(tt.key, tt.base, date(t, tt.on), catalog2026(t))
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestSuggestAtRateOverflowIsAParameterFault(t *testing.T) {
	absurd := fakeCatalog{fiscal.KeyISRWithholdingFees: {{from: date(t, "2022-01-01"), rate: math.MaxInt64}}}
	_, err := fiscal.SuggestAtRate(fiscal.KeyISRWithholdingFees, fiscal.MaxCents, date(t, "2026-01-01"), absurd)
	require.ErrorIs(t, err, fiscal.ErrInvalidParameter)
}

// TestSuggestionDiffers: only a known suggestion farther than IVATolerance
// from the withheld amount warns; the CFDI's amount wins either way.
func TestSuggestionDiffers(t *testing.T) {
	assert.False(t, known(1_000_00).Differs(1_001_00))
	assert.True(t, known(1_000_00).Differs(1_001_01))
	assert.True(t, known(1_000_00).Differs(0))
	assert.False(t, fiscal.Suggestion{}.Differs(5_000_00))
}
