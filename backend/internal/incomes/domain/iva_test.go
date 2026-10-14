package domain_test

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/incomes/domain"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
)

// TestExpectedIVA uses the research worked examples (A.5): IVA is the
// amount before IVA times the rate, rounded once, half away from zero.
func TestExpectedIVA(t *testing.T) {
	tests := []struct {
		name    string
		amount  money.Cents
		ivaRate domain.IVARate
		on      string
		want    money.Cents
		wantErr error
	}{
		{"PF-612-IVA-01: 30,000 at 16 %", 30_000_00, domain.IVARate16, "2026-03-31", 4_800_00, nil},
		{"PF-626-IVA-01: 60,000 at 16 %", 60_000_00, domain.IVARate16, "2026-03-31", 9_600_00, nil},
		{"PF-606-IVA-01: rent 20,000 at 16 %", 20_000_00, domain.IVARate16, "2026-06-30", 3_200_00, nil},
		{"PF-625-01: platform sales 30,000", 30_000_00, domain.IVARate16, "2026-06-30", 4_800_00, nil},
		{"border 8 % on its last day", 10_000_00, domain.IVARate8, "2026-12-31", 800_00, nil},
		{"0.48 centavo rounds down", 3, domain.IVARate16, "2026-06-30", 0, nil},
		{"0.64 centavo rounds up", 4, domain.IVARate16, "2026-06-30", 1, nil},
		{"0.56 centavo at 8 % rounds up", 7, domain.IVARate8, "2026-06-30", 1, nil},
		{"zero rate", 10_000_00, domain.IVARate0, "2026-06-30", 0, nil},
		{"exempt needs no catalog", 10_000_00, domain.IVAExempt, "1999-01-01", 0, nil},
		{"border 8 % lapsed 2027-01-05", 10_000_00, domain.IVARate8, "2027-01-05", 0, domain.ErrRateNotValidOnDate},
		{"unknown treatment", 10_000_00, domain.IVARate("rate_20"), "2026-06-30", 0, domain.ErrInvalid},
		{"no parameter before the catalog starts", 10_000_00, domain.IVARate16, "2021-12-31", 0, domain.ErrRateNotValidOnDate},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := domain.ExpectedIVA(tt.amount, tt.ivaRate, date(t, tt.on), catalog2026(t))
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestExpectedIVAOverflowIsAParameterFault(t *testing.T) {
	absurd := fakeCatalog{domain.KeyIVAGeneral: {{from: date(t, "2022-01-01"), rate: math.MaxInt64}}}
	_, err := domain.ExpectedIVA(domain.MaxCents, domain.IVARate16, date(t, "2026-01-01"), absurd)
	require.ErrorIs(t, err, domain.ErrInvalidParameter)
}

// TestCheckIVA is MODULE-incomes §2.5 and US-I1: ±100 centavos (IvaTolerance)
// absorbs per-line CFDI rounding; anything else is iva_amount_mismatch.
func TestCheckIVA(t *testing.T) {
	tests := []struct {
		name    string
		amount  money.Cents
		ivaRate domain.IVARate
		iva     money.Cents
		on      string
		wantErr error
	}{
		{"US-I1: 10,000 at 16 % with 1,600.00", 10_000_00, domain.IVARate16, 1_600_00, "2026-09-12", nil},
		{"US-I1: 1,600.40 is inside the tolerance", 10_000_00, domain.IVARate16, 1_600_40, "2026-09-12", nil},
		{"exactly 100 centavos over", 10_000_00, domain.IVARate16, 1_601_00, "2026-09-12", nil},
		{"exactly 100 centavos under", 10_000_00, domain.IVARate16, 1_599_00, "2026-09-12", nil},
		{"101 centavos over", 10_000_00, domain.IVARate16, 1_601_01, "2026-09-12", domain.ErrIVAAmountMismatch},
		{"US-I1: 1,500.00 is a typo", 10_000_00, domain.IVARate16, 1_500_00, "2026-09-12", domain.ErrIVAAmountMismatch},
		{"US-I1: exempt with IVA 1.00", 10_000_00, domain.IVAExempt, 1_00, "2026-09-12", domain.ErrIVAAmountMismatch},
		{"zero rate with no IVA", 10_000_00, domain.IVARate0, 0, "2026-09-12", nil},
		{"zero rate never tolerates IVA", 10_000_00, domain.IVARate0, 1, "2026-09-12", domain.ErrIVAAmountMismatch},
		{"US-I1: 8 % on 2027-01-05", 10_000_00, domain.IVARate8, 800_00, "2027-01-05", domain.ErrRateNotValidOnDate},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := domain.CheckIVA(tt.amount, tt.ivaRate, tt.iva, date(t, tt.on), catalog2026(t))
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestCheckIVAParameterFault(t *testing.T) {
	absurd := fakeCatalog{domain.KeyIVAGeneral: {{from: date(t, "2022-01-01"), rate: math.MaxInt64}}}
	err := domain.CheckIVA(domain.MaxCents, domain.IVARate16, 0, date(t, "2026-01-01"), absurd)
	require.ErrorIs(t, err, domain.ErrInvalidParameter)
}

func TestIVATolerance(t *testing.T) {
	assert.Equal(t, money.Cents(100), domain.IVATolerance)
}
