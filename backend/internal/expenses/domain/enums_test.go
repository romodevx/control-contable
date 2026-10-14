package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/expenses/domain"
	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
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
		require.ErrorIs(t, err, fiscal.ErrInvalid, raw)
	}
}

// TestParseEnums covers the expenses module's own enums; the shared ones
// (IvaRate, PersonType) are the fiscal kernel's.
func TestParseEnums(t *testing.T) {
	t.Run("payment method", func(t *testing.T) {
		parseCase(t, domain.ParsePaymentMethod,
			[]string{"cash", "transfer", "card_debit", "card_credit", "check", "e_wallet", "other"})
	})
	t.Run("deductibility", func(t *testing.T) {
		parseCase(t, domain.ParseDeductibility,
			[]string{"deductible", "partially_deductible", "not_deductible", "isr_not_applicable"})
	})
	t.Run("deduction mode", func(t *testing.T) {
		parseCase(t, domain.ParseDeductionMode, []string{"actual", "blind_35"})
	})
}

func TestOverrideValid(t *testing.T) {
	for _, o := range []domain.Override{domain.UseDefault, domain.ClaimDeductible, domain.ClaimNotDeductible} {
		assert.True(t, o.Valid())
	}
	assert.False(t, domain.Override(-1).Valid())
	assert.False(t, domain.Override(3).Valid())
}

func TestLegend(t *testing.T) {
	assert.Equal(t, "Esto es un aproximado. No todos los gastos aplican para tu giro comercial.", domain.Legend)
}
