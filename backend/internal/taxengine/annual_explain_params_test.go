package taxengine_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/platform/money"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain/annual/explain"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain/annual/pm"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain/taxparams"
)

// The annual ports on the embedded parameter catalog (PLAN ANN-10, ANN-12):
// *taxparams.Params is the pm.Params and explain.Params the composition
// root (ANN-14) passes, and the 2026 set yields the spec's dates.
var (
	_ pm.Params      = (*taxparams.Params)(nil)
	_ explain.Params = (*taxparams.Params)(nil)
)

func TestAnnualDeadlines2026FromTheEmbeddedParams(t *testing.T) {
	sel, err := catalog(t).Select(2026)
	require.NoError(t, err)

	for kind, want := range map[explain.DeadlineKind]string{
		explain.DeadlinePF: "2027-04-30", explain.DeadlinePM: "2027-03-31", explain.DeadlineNonprofitInformative: "2027-02-15",
	} {
		got, err := explain.Deadline(sel.Params, 2026, kind)

		require.NoError(t, err, kind)
		assert.Equal(t, want, got.String(), kind)
	}
}

func TestAnnualInformativeIncome2026FromTheEmbeddedParams(t *testing.T) {
	sel, err := catalog(t).Select(2026)
	require.NoError(t, err)

	at, err := explain.InformativeIncome(sel.Params, 2026, money.Cents(500_000_00))
	require.NoError(t, err)
	above, err := explain.InformativeIncome(sel.Params, 2026, money.Cents(500_000_01))
	require.NoError(t, err)

	assert.False(t, at, "LISR 150 ¶3: superiores a $500,000")
	assert.True(t, above)
}
