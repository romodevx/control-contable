package annuallimits_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/platform/money"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain/annual"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain/taxparams"
	"github.com/romodevx/control-contable/backend/internal/taxengine/infra/annuallimits"
	"github.com/romodevx/control-contable/backend/internal/taxengine/params"
)

// The reader is the calculator's own port, not a mirrored copy of it.
var _ annual.LimitsReader = (*annuallimits.Reader)(nil)

// TestEmbedded2026LimitsPassTheCalculatorValidation reads the published
// 2026 set through the port and runs the consumer's own validation, which
// requires a cap for every annual.TuitionLevel: a level added on either side
// without the other fails here.
func TestEmbedded2026LimitsPassTheCalculatorValidation(t *testing.T) {
	catalog, err := params.Embedded()
	require.NoError(t, err)

	limits, err := annuallimits.New(catalog).DeductionLimits(2026)

	require.NoError(t, err)
	require.NoError(t, limits.Validate())
	assert.Len(t, limits.TuitionCaps, 5)
	assert.Equal(t, money.Cents(4_279_464), limits.AnnualUMA, "UMA valor anual in force on 31-12-2026")
	assert.Equal(t, money.Cents(1_290_000), limits.TuitionCaps[annual.LevelPrimaria], "decree art. 1.10: Primaria $12,900.00")
}

func TestReaderReportsAYearWithoutLimits(t *testing.T) {
	catalog, err := params.Embedded()
	require.NoError(t, err)

	_, err = annuallimits.New(catalog).DeductionLimits(2019)

	require.ErrorIs(t, err, taxparams.ErrNoParameters)
}

func TestCalculatorRunsOnTheEmbeddedLimits(t *testing.T) {
	catalog, err := params.Embedded()
	require.NoError(t, err)
	calc := annual.NewCalculator(annuallimits.New(catalog))

	got, err := calc.Compute(annual.DeductionsInput{
		FiscalYear: 2026,
		Deductions: []annual.Deduction{{
			Category: annual.CategoryTuition, Amount: 2_000_000, PaymentMethod: annual.PaymentElectronic,
			Beneficiary: annual.BeneficiarySelf, Student: "Ana", TuitionLevel: annual.LevelPrimaria,
		}},
		AccumulableIncome: 50_000_000, TotalIncome: 50_000_000,
	})

	require.NoError(t, err)
	assert.Equal(t, money.Cents(1_290_000), got.UncappedApplied, "20,000 of primaria capped at 12,900")
}
