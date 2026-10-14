package domain_test

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/expenses/domain"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
)

// TestExpectedIVA: IVA paid on expenses in the research worked examples.
func TestExpectedIVA(t *testing.T) {
	tests := []struct {
		name      string
		amount    money.Cents
		treatment domain.IVARate
		on        string
		want      money.Cents
		wantErr   error
	}{
		{"PF-612-IVA-01: IVA paid 1,600 on 10,000", 10_000_00, domain.IVARate16, "2026-03-15", 1_600_00, nil},
		{"PF-626-IVA-01: creditable 2,000 on 12,500", 12_500_00, domain.IVARate16, "2026-03-15", 2_000_00, nil},
		{"PM-601-IVA-01: lawyer fees 50,000", 50_000_00, domain.IVARate16, "2026-06-15", 8_000_00, nil},
		{"US-E1: software 500.00", 500_00, domain.IVARate16, "2026-09-12", 80_00, nil},
		{"border 8 % on its last day", 10_000_00, domain.IVARate8, "2026-12-31", 800_00, nil},
		{"0.64 centavo rounds up", 4, domain.IVARate16, "2026-06-30", 1, nil},
		{"zero rate", 10_000_00, domain.IVARate0, "2026-06-30", 0, nil},
		{"exempt needs no catalog", 10_000_00, domain.IVAExempt, "1999-01-01", 0, nil},
		{"unknown treatment", 10_000_00, domain.IVARate("rate_20"), "2026-06-30", 0, domain.ErrInvalid},
		{"border 8 % lapsed", 10_000_00, domain.IVARate8, "2027-01-05", 0, domain.ErrRateNotValidOnDate},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := domain.ExpectedIVA(tt.amount, tt.treatment, date(t, tt.on), catalog2026(t))
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCheckIVA(t *testing.T) {
	tests := []struct {
		name      string
		amount    money.Cents
		treatment domain.IVARate
		iva       money.Cents
		on        string
		wantErr   error
	}{
		{"exact", 500_00, domain.IVARate16, 80_00, "2026-09-12", nil},
		{"within 100 centavos", 10_000_00, domain.IVARate16, 1_601_00, "2026-09-12", nil},
		{"101 centavos off", 10_000_00, domain.IVARate16, 1_598_99, "2026-09-12", domain.ErrIVAAmountMismatch},
		{"exempt with IVA", 10_000_00, domain.IVAExempt, 1, "2026-09-12", domain.ErrIVAAmountMismatch},
		{"zero rate without IVA", 10_000_00, domain.IVARate0, 0, "2026-09-12", nil},
		{"border 8 % lapsed", 10_000_00, domain.IVARate8, 800_00, "2027-01-05", domain.ErrRateNotValidOnDate},
		{"difference beyond int64", -10_00, domain.IVARate16, math.MaxInt64, "2026-09-12", domain.ErrIVAAmountMismatch},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := domain.CheckIVA(tt.amount, tt.treatment, tt.iva, date(t, tt.on), catalog2026(t))
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestIVAParameterFault(t *testing.T) {
	absurd := catalog2026(t)
	absurd.rates[domain.KeyIVAGeneral] = []dated[money.Rate]{{from: date(t, "2022-01-01"), value: math.MaxInt64}}
	_, err := domain.ExpectedIVA(domain.MaxCents, domain.IVARate16, date(t, "2026-01-01"), absurd)
	require.ErrorIs(t, err, domain.ErrInvalidParameter)
	err = domain.CheckIVA(domain.MaxCents, domain.IVARate16, 0, date(t, "2026-01-01"), absurd)
	require.ErrorIs(t, err, domain.ErrInvalidParameter)
}

func TestConstants(t *testing.T) {
	assert.Equal(t, money.Cents(100), domain.IVATolerance)
	assert.Equal(t, money.Cents(99_999_999_999_999), domain.MaxCents)
}
