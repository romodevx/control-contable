package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/incomes/domain"
	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
)

// parseCase checks one enum parser: every contract value round-trips and
// near misses (case, spaces, empty) are fiscal.ErrInvalid.
func parseCase[T ~string](t *testing.T, parse func(string) (T, error), valid []string) {
	t.Helper()
	for _, raw := range valid {
		got, err := parse(raw)
		require.NoError(t, err, raw)
		assert.Equal(t, raw, string(got))
	}
	for _, raw := range []string{"", " " + valid[0], "UNKNOWN", "Rate_16"} {
		_, err := parse(raw)
		require.ErrorIs(t, err, fiscal.ErrInvalid, raw)
	}
}

// TestParseEnums covers the incomes module's own enums; the shared ones
// (IvaRate, PersonType) are the fiscal kernel's.
func TestParseEnums(t *testing.T) {
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
}
