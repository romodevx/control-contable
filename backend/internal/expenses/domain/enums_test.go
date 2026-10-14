package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/expenses/domain"
)

func parseCase[T ~string](t *testing.T, parse func(string) (T, error), valid []string) {
	t.Helper()
	for _, raw := range valid {
		got, err := parse(raw)
		require.NoError(t, err, raw)
		assert.Equal(t, raw, string(got))
	}
	for _, raw := range []string{"", " " + valid[0], "UNKNOWN"} {
		_, err := parse(raw)
		require.ErrorIs(t, err, domain.ErrInvalid, raw)
	}
}

func TestParseEnums(t *testing.T) {
	t.Run("iva rate", func(t *testing.T) {
		parseCase(t, domain.ParseIVARate, []string{"rate_16", "rate_8", "rate_0", "exempt"})
	})
	t.Run("payment method", func(t *testing.T) {
		parseCase(t, domain.ParsePaymentMethod,
			[]string{"cash", "transfer", "card_debit", "card_credit", "check", "e_wallet", "other"})
	})
	t.Run("deductibility", func(t *testing.T) {
		parseCase(t, domain.ParseDeductibility,
			[]string{"deductible", "partially_deductible", "not_deductible", "isr_not_applicable"})
	})
	t.Run("person type", func(t *testing.T) {
		parseCase(t, domain.ParsePersonType, []string{"fisica", "moral"})
	})
	t.Run("deduction mode", func(t *testing.T) {
		parseCase(t, domain.ParseDeductionMode, []string{"actual", "blind_35"})
	})
}

func TestIVARateTaxable(t *testing.T) {
	assert.True(t, domain.IVARate16.Taxable())
	assert.True(t, domain.IVARate8.Taxable())
	assert.False(t, domain.IVARate0.Taxable())
	assert.False(t, domain.IVAExempt.Taxable())
}

func TestLegend(t *testing.T) {
	assert.Equal(t, "Esto es un aproximado. No todos los gastos aplican para tu giro comercial.", domain.Legend)
}
