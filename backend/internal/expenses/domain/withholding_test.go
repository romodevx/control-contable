package domain_test

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/expenses/domain"
	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
)

func TestMayWithhold(t *testing.T) {
	assert.True(t, domain.MayWithhold(pm601), "LIVA 1-A: a persona moral withholds")
	assert.False(t, domain.MayWithhold(pf612))
}

// TestSuggestWithholdings: the suggestions a persona moral sees for its
// supplier (MODULE-expenses §2.6, research B.2.3, PM-601-IVA-01).
func TestSuggestWithholdings(t *testing.T) {
	known := func(c money.Cents) fiscal.Suggestion { return fiscal.Suggestion{Cents: c, Known: true} }
	tests := []struct {
		name        string
		category    domain.CategoryCode
		payer       domain.Payer
		amount, iva money.Cents
		want        fiscal.Suggested
	}{
		{"PM-601-IVA-01: PF lawyer fees 50,000 + 8,000", domain.CategoryProfessionalFees, pm601, 50_000_00, 8_000_00,
			fiscal.Suggested{ISR: known(5_000_00), IVA: known(5_333_33)}},
		{"US-E1: fees 10,000 + 1,600", domain.CategoryProfessionalFees, pm601, 10_000_00, 1_600_00,
			fiscal.Suggested{ISR: known(1_000_00), IVA: known(1_066_67)}},
		{"rent 20,000 + 3,200", domain.CategoryRent, pm626, 20_000_00, 3_200_00,
			fiscal.Suggested{ISR: known(2_000_00), IVA: known(2_133_33)}},
		{"freight: 4 % of the price, no ISR (RLIVA 3 fr. II)", domain.CategoryFreight, pm601, 10_000_00, 1_600_00,
			fiscal.Suggested{IVA: known(400_00)}},
		{"software: no suggestion", "software", pm601, 10_000_00, 1_600_00, fiscal.Suggested{}},
		{"a persona física never withholds", domain.CategoryProfessionalFees, pf612, 10_000_00, 1_600_00, fiscal.Suggested{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := domain.SuggestWithholdings(tt.category, tt.payer, tt.amount, tt.iva, date(t, "2026-06-15"), catalog2026(t))
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestSuggestWithholdingsMissingRate(t *testing.T) {
	got, err := domain.SuggestWithholdings(domain.CategoryFreight, pm601, 10_000_00, 1_600_00, date(t, "2026-06-15"), fakeCatalog{})
	require.NoError(t, err)
	assert.Equal(t, fiscal.Suggested{}, got, "a rate the catalog lacks is no suggestion")
}

func TestSuggestWithholdingsParameterFaults(t *testing.T) {
	absurd := func(key fiscal.RateKey) fakeCatalog {
		c := catalog2026(t)
		c.rates[key] = []dated[money.Rate]{{from: date(t, "2022-01-01"), value: math.MaxInt64}}
		return c
	}
	tests := []struct {
		name     string
		category domain.CategoryCode
		iva      money.Cents
		catalog  fakeCatalog
	}{
		{"ISR rate overflows", domain.CategoryProfessionalFees, 0, absurd(fiscal.KeyISRWithholdingFees)},
		{"IVA share overflows", domain.CategoryProfessionalFees, math.MaxInt64, catalog2026(t)},
		{"freight rate overflows", domain.CategoryFreight, 0, absurd(domain.KeyIVAWithholdingFreight)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := domain.SuggestWithholdings(tt.category, pm601, fiscal.MaxCents, tt.iva, date(t, "2026-06-15"), tt.catalog)
			require.ErrorIs(t, err, fiscal.ErrInvalidParameter)
		})
	}
}

func TestShareIsZero(t *testing.T) {
	assert.True(t, fiscal.Share{}.IsZero())
	assert.False(t, fiscal.ShareTwoThirds.IsZero())
}
