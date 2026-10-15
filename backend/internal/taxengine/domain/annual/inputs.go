package annual

import (
	"slices"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain"
)

// RegimeSalary is 605 (sueldos y salarios): its income comes from the
// salary constancias, never from records.
const RegimeSalary = regimeSalary

// Categories returns every category in article order (LISR 151, then the
// decree): the order lists and totals show them in.
func Categories() []Category { return slices.Clone(categoryOrder) }

// InputAnnualDepreciation is the M5 tax input the annual sections read for
// the year's depreciation (612 PF, 601 PM, 626 PM; catalog data since
// ANN-03): typed from the CPA's fixed-asset schedule, no asset register.
const InputAnnualDepreciation domain.InputKey = "annual_depreciation"

// annualTaxInputKeys are the M5 tax inputs getAnnualInputs mirrors
// read-only (MODULE-annual §2.6, ADR-0009 "one source per fact"), in the
// contract's order.
var annualTaxInputKeys = []domain.InputKey{
	domain.InputPriorYearLosses, domain.InputPTUPaid, InputAnnualDepreciation, domain.InputInterestWithholding,
}

// TaxInputKeys returns the annual-relevant M5 tax input keys.
func TaxInputKeys() []domain.InputKey { return slices.Clone(annualTaxInputKeys) }

// IsTaxInput reports one of TaxInputKeys.
func IsTaxInput(key domain.InputKey) bool { return slices.Contains(annualTaxInputKeys, key) }

// provisionalISRRegimes pay provisional ISR every month: 601 (LISR 14),
// 606 (LISR 116), 612 (LISR 106), 625 without the definitive option
// (LISR 113-B) and 626 (LISR 113-E, 211). Their payments are annual
// credits (MODULE-annual §2.5).
var provisionalISRRegimes = []fiscal.RegimeCode{"601", fiscal.RegimeRent, fiscal.RegimeBusiness, fiscal.RegimePlatforms, fiscal.RegimeRESICO}

// PaysProvisionalISR reports a regime with monthly provisional ISR
// payments; 625 with the definitive option (platform withholdings are the
// final tax) has none, nor do 605 and the explanation-only regimes
// (getFiledPayments payments_not_applicable).
func PaysProvisionalISR(code fiscal.RegimeCode, definitiveOption bool) bool {
	if code == fiscal.RegimePlatforms && definitiveOption {
		return false
	}
	return slices.Contains(provisionalISRRegimes, code)
}
