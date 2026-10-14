package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/incomes/domain"
)

// parseCase checks one enum parser: every contract value round-trips and
// near misses (case, spaces, empty) are ErrInvalid.
func parseCase[T ~string](t *testing.T, parse func(string) (T, error), valid []string) {
	t.Helper()
	for _, raw := range valid {
		got, err := parse(raw)
		require.NoError(t, err, raw)
		assert.Equal(t, raw, string(got))
	}
	for _, raw := range []string{"", " " + valid[0], "UNKNOWN", "Rate_16"} {
		_, err := parse(raw)
		require.ErrorIs(t, err, domain.ErrInvalid, raw)
	}
}

func TestParseEnums(t *testing.T) {
	t.Run("iva rate", func(t *testing.T) {
		parseCase(t, domain.ParseIVARate, []string{"rate_16", "rate_8", "rate_0", "exempt"})
	})
	t.Run("counterparty", func(t *testing.T) {
		parseCase(t, domain.ParseCounterparty,
			[]string{"persona_fisica", "persona_moral", "publico_general", "plataforma"})
	})
	t.Run("platform activity", func(t *testing.T) {
		parseCase(t, domain.ParsePlatformActivity, []string{"transport_delivery", "lodging", "goods_services"})
	})
	t.Run("income basis", func(t *testing.T) {
		parseCase(t, domain.ParseIncomeBasis, []string{"cash", "accrual"})
	})
	t.Run("person type", func(t *testing.T) {
		parseCase(t, domain.ParsePersonType, []string{"fisica", "moral"})
	})
}

func TestIVARateTaxable(t *testing.T) {
	assert.True(t, domain.IVARate16.Taxable())
	assert.True(t, domain.IVARate8.Taxable())
	assert.False(t, domain.IVARate0.Taxable())
	assert.False(t, domain.IVAExempt.Taxable())
}
