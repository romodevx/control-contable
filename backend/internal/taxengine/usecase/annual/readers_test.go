package annual_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/romodevx/control-contable/backend/internal/platform/money"
	adomain "github.com/romodevx/control-contable/backend/internal/taxengine/domain/annual"
	"github.com/romodevx/control-contable/backend/internal/taxengine/usecase/annual"
)

func TestYearInputsOf(t *testing.T) {
	tuitionRow := annual.StoredDeduction{ID: "b", DeductionFields: annual.DeductionFields{
		Category: adomain.CategoryTuition, Amount: 3_000_000, PaidOn: mustDate("2026-08-20"), PaymentMethod: adomain.PaymentElectronic,
		Beneficiary: adomain.BeneficiaryDescendant, StudentLabel: "Hija", TuitionLevel: adomain.LevelPrimaria,
		Scholarship: annual.NullCents{Cents: 100, Valid: true},
	}}
	medicalRow := annual.StoredDeduction{ID: "a", DeductionFields: medicalFields("2026-06-12", 4_000_000, adomain.PaymentCash)}
	statement := annual.StoredStatement{SalaryFields: mainFields()}
	facts := annual.Facts{PriorYearAccumulableIncome: ptr(money.Cents(1)), OtherExemptIncome: 2}

	got := annual.YearInputsOf([]annual.StoredDeduction{tuitionRow, medicalRow}, []annual.StoredStatement{statement}, facts)

	assert.Equal(t, annual.YearInputs{
		Deductions: []adomain.Deduction{
			{Category: adomain.CategoryMedical, Amount: 4_000_000, PaymentMethod: adomain.PaymentCash, Beneficiary: adomain.BeneficiarySelf},
			{Category: adomain.CategoryTuition, Amount: 3_000_000, PaymentMethod: adomain.PaymentElectronic, Beneficiary: adomain.BeneficiaryDescendant,
				Student: "Hija", TuitionLevel: adomain.LevelPrimaria, Scholarship: 100},
		},
		SalaryStatements: []adomain.SalaryStatement{{Accumulable: 30_000_000, ISRWithheld: 4_000_000, Simultaneous: true}},
		Facts:            adomain.PFAnnualInputs{PriorYearAccumulableIncome: ptr(money.Cents(1)), OtherExemptIncome: 2},
	}, got, "deductions in the list order (category, then paid_on desc, then id): a deterministic calculator input")

	empty := annual.YearInputsOf(nil, nil, annual.Facts{})
	assert.NotNil(t, empty.Deductions)
	assert.NotNil(t, empty.SalaryStatements)
}
