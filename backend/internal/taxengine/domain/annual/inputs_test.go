package annual_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain/annual"
)

func TestPaysProvisionalISR(t *testing.T) {
	tests := []struct {
		code       fiscal.RegimeCode
		definitive bool
		want       bool
	}{
		{"601", false, true},
		{"606", false, true},
		{"612", false, true},
		{"626", false, true},
		{"625", false, true},
		{"625", true, false},
		{"605", false, false},
		{"603", false, false},
		{"614", false, false},
		{"616", false, false},
		{"612", true, true},
	}
	for _, tt := range tests {
		assert.Equalf(t, tt.want, annual.PaysProvisionalISR(tt.code, tt.definitive), "%s definitive=%t", tt.code, tt.definitive)
	}
}

func TestCategoriesInArticleOrder(t *testing.T) {
	got := annual.Categories()
	assert.Equal(t, []annual.Category{
		annual.CategoryMedical, annual.CategoryMedicalDisability, annual.CategoryFuneral, annual.CategoryDonation,
		annual.CategoryMortgageRealInterest, annual.CategoryRetirementContribution, annual.CategoryHealthInsurance,
		annual.CategorySchoolTransport, annual.CategoryLocalPayrollTax, annual.CategoryTuition,
	}, got)
	got[0] = "changed"
	assert.Equal(t, annual.CategoryMedical, annual.Categories()[0], "a copy")
}

func TestTaxInputKeys(t *testing.T) {
	assert.Equal(t, domain.InputKey("annual_depreciation"), annual.InputAnnualDepreciation)
	assert.Equal(t, []domain.InputKey{
		domain.InputPriorYearLosses, domain.InputPTUPaid, annual.InputAnnualDepreciation, domain.InputInterestWithholding,
	}, annual.TaxInputKeys())
	assert.True(t, annual.IsTaxInput(domain.InputPTUPaid))
	assert.False(t, annual.IsTaxInput(domain.InputPriorProvisionalPayments))
	assert.Equal(t, fiscal.RegimeCode("605"), annual.RegimeSalary)
}
