package fiscal_test

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
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
		require.ErrorIs(t, err, fiscal.ErrInvalid, raw)
	}
}

func TestParseEnums(t *testing.T) {
	t.Run("iva rate", func(t *testing.T) {
		parseCase(t, fiscal.ParseIVARate, []string{"rate_16", "rate_8", "rate_0", "exempt"})
	})
	t.Run("person type", func(t *testing.T) {
		parseCase(t, fiscal.ParsePersonType, []string{"fisica", "moral"})
	})
}

func TestIVARateTaxable(t *testing.T) {
	assert.True(t, fiscal.IVARate16.Taxable())
	assert.True(t, fiscal.IVARate8.Taxable())
	assert.False(t, fiscal.IVARate0.Taxable())
	assert.False(t, fiscal.IVAExempt.Taxable())
}

// colour is a module-owned enum: modules extend the kernel's enum helpers
// by giving their own type a Valid method.
type colour string

func (c colour) Valid() bool { return slices.Contains([]colour{"red", "green"}, c) }

func TestParseEnumOfAModuleType(t *testing.T) {
	parseCase(t, fiscal.ParseEnum[colour], []string{"red", "green"})
}

func TestCheckEnum(t *testing.T) {
	var vs fiscal.Violations
	fiscal.CheckEnum(&vs, "a", colour(""))
	fiscal.CheckEnum(&vs, "b", colour("blue"))
	fiscal.CheckEnum(&vs, "c", colour("red"))
	assert.Equal(t, fiscal.Violations{
		{Field: "a", Err: fiscal.ErrRequired},
		{Field: "b", Err: fiscal.ErrInvalid},
	}, vs)
}
