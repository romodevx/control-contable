package annual

import (
	"errors"
	"fmt"
	"slices"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain"
)

// ErrInvalidYear reports an annual input that breaks MODULE-annual §2.6: a
// negative figure, a regime held twice, a payment outside the spec.
const ErrInvalidYear Error = "annual: invalid annual input"

// regimeSalary is Sueldos y salarios (LISR Título IV Cap. I); its figures
// come from the constancias, never from the ledger.
const regimeSalary fiscal.RegimeCode = "605"

// OptionPlatformDefinitive is the 625 option of LISR 113-B: platform
// withholdings as a definitive tax, excluded from the annual (152 ¶2).
const OptionPlatformDefinitive domain.OptionKey = "definitive_option"

// RegimeYear is one regime's fiscal year as the annual sections read it:
// income on the regime's basis (collected for PF), authorized deductions
// paid, predial (606), ISR withheld from the taxpayer (10 % by personas
// morales, platform withholdings, 1.25 % RESICO), the money tax inputs the
// annual uses and the regime's options.
type RegimeYear struct {
	Regime          fiscal.RegimeCode
	Income          money.Cents
	Deductions      money.Cents
	Predial         money.Cents
	ISRWithheld     money.Cents
	Depreciation    money.Cents // tax input annual_depreciation
	PTUPaid         money.Cents // tax input ptu_paid
	PriorYearLosses money.Cents // tax input prior_year_losses, already updated
	Options         domain.Options
}

func (r RegimeYear) validate() error {
	figures := []money.Cents{r.Income, r.Deductions, r.Predial, r.ISRWithheld, r.Depreciation, r.PTUPaid, r.PriorYearLosses}
	if slices.ContainsFunc(figures, func(v money.Cents) bool { return v < 0 }) {
		return fmt.Errorf("regime %s: figures must not be negative: %w", r.Regime, ErrInvalidYear)
	}
	return nil
}

// platformDefinitive reports a 625 under the definitive option.
func (r RegimeYear) platformDefinitive() bool {
	definitive, _ := r.Options.Bool(OptionPlatformDefinitive)
	return r.Regime == fiscal.RegimePlatforms && definitive
}

// SalaryStatement is one employer's constancia (MODULE-annual §2.6): its
// totals are trusted, never recomputed.
type SalaryStatement struct {
	Accumulable         money.Cents
	Exempt              money.Cents
	ISRWithheld         money.Cents
	Simultaneous        bool
	LeftBeforeYearEnd   bool
	PayerNotWithholding bool
}

func (s SalaryStatement) validate() error {
	if s.Accumulable < 0 || s.Exempt < 0 || s.ISRWithheld < 0 {
		return fmt.Errorf("constancia figures must not be negative: %w", ErrInvalidYear)
	}
	return nil
}

// PFAnnualInputs are a persona física's per-year annual inputs.
type PFAnnualInputs struct {
	// PriorYearAccumulableIncome: the prior year's acumulables before
	// personal deductions (the 7 % donation cap); nil when not given.
	PriorYearAccumulableIncome *money.Cents
	// OtherExemptIncome joins the global cap's total income.
	OtherExemptIncome    money.Cents
	ToldEmployerWillFile bool
}

// PFYear is a persona física's fiscal year: every regime held (only the
// months in its window counted), the constancias, the personal deductions
// and the annual inputs. The provisional payments come through the
// PaymentsReader port.
type PFYear struct {
	FiscalYear       int
	Regimes          []RegimeYear
	SalaryStatements []SalaryStatement
	Deductions       []Deduction
	Inputs           PFAnnualInputs
}

func (y PFYear) validate() error {
	var problems []error
	seen := map[fiscal.RegimeCode]bool{}
	for _, r := range y.Regimes {
		if seen[r.Regime] {
			problems = append(problems, fmt.Errorf("regime %s held twice: %w", r.Regime, ErrInvalidYear))
		}
		seen[r.Regime] = true
		problems = append(problems, r.validate())
	}
	for i, s := range y.SalaryStatements {
		if err := s.validate(); err != nil {
			problems = append(problems, fmt.Errorf("constancia %d: %w", i, err))
		}
	}
	if y.Inputs.OtherExemptIncome < 0 {
		problems = append(problems, fmt.Errorf("other exempt income must not be negative: %w", ErrInvalidYear))
	}
	return errors.Join(problems...)
}

// regime returns the year of code, if held.
func (y PFYear) regime(code fiscal.RegimeCode) (RegimeYear, bool) {
	for _, r := range y.Regimes {
		if r.Regime == code {
			return r, true
		}
	}
	return RegimeYear{}, false
}

// holds reports a regime held in the year.
func (y PFYear) holds(code fiscal.RegimeCode) bool {
	_, ok := y.regime(code)
	return ok
}
