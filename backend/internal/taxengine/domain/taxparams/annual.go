package taxparams

import (
	"errors"
	"fmt"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
)

// Keys of the annual return parameters (MODULE-annual §2.7, §2.9; ANN-04).
const (
	// KeyAnnualUMA is the UMA "valor anual" (INEGI notice in the DOF).
	KeyAnnualUMA Key = "uma.annual"
	// KeyGlobalCapUMAs and KeyGlobalCapIncomeShare: the personal deductions
	// are at most min(n UMA anuales, share of total income) (LISR 151 last ¶).
	KeyGlobalCapUMAs        Key = "deduction.personal.global_cap_umas"
	KeyGlobalCapIncomeShare Key = "deduction.personal.global_cap_income_share"
	// KeyFuneralCapUMAs: funeral expenses up to n UMA anuales (LISR 151 fr. II).
	KeyFuneralCapUMAs Key = "deduction.personal.funeral_cap_umas"
	// KeyDonationShare of the prior year's acumulables, of which
	// KeyGovernmentDonationShare may go to government donees (LISR 151 fr. III).
	KeyDonationShare           Key = "deduction.personal.donation_share"
	KeyGovernmentDonationShare Key = "deduction.personal.government_donation_share"
	// KeyRetirementIncomeShare of the year's acumulables, at most
	// KeyRetirementCapUMAs UMA anuales (LISR 151 fr. V).
	KeyRetirementIncomeShare Key = "deduction.personal.retirement_income_share"
	KeyRetirementCapUMAs     Key = "deduction.personal.retirement_cap_umas"
	// KeySalaryAnnualThreshold: salary income above it obliges the annual
	// return (LISR 98 fr. III e).
	KeySalaryAnnualThreshold Key = "annual.salary_obligation_threshold"
	// KeyInformativeIncomeThreshold: total income above it obliges reporting
	// exempt and definitively taxed income (LISR 150 ¶3).
	KeyInformativeIncomeThreshold Key = "annual.informative_income_threshold"
	// The annual return deadlines of the set's fiscal year (MODULE-annual §2.9).
	KeyDeadlinePF                   Key = "annual.deadline.pf"
	KeyDeadlinePM                   Key = "annual.deadline.pm"
	KeyDeadlineNonprofitInformative Key = "annual.deadline.nonprofit_informative"
)

// tuitionCapPrefix + level is the key of a colegiaturas cap per student and
// level (decree DOF 26-12-2013 art. 1.10).
const tuitionCapPrefix = "deduction.tuition_cap."

// TuitionLevels are the decree's school levels, spelled as the consumer's
// level values (annual.TuitionLevel, PR #127). Each needs a cap.
var TuitionLevels = []string{"preescolar", "primaria", "secundaria", "profesional_tecnico", "bachillerato"}

// TuitionCapKey returns the key of a level's colegiaturas cap.
func TuitionCapKey(level string) Key { return Key(tuitionCapPrefix + level) }

// DeductionLimits are the personal-deduction limits of one fiscal year. Its
// fields, types and order are those of annual.Limits, the port of the
// deductions calculator, so the composition root converts one into the
// other with a plain conversion: annual.Limits(limits). L is the consumer's
// tuition level type.
type DeductionLimits[L ~string] struct {
	AnnualUMA               money.Cents
	GlobalCapUMAs           int64
	GlobalCapIncomeShare    money.Rate
	FuneralCapUMAs          int64
	DonationShare           money.Rate
	GovernmentDonationShare money.Rate
	RetirementIncomeShare   money.Rate
	RetirementCapUMAs       int64
	TuitionCaps             map[L]money.Cents
}

// ReadDeductionLimits returns the limits of fiscalYear from the set the
// catalog selects for it, each resolved on 31 December of that year (the
// UMA in force when the annual tax is computed). Any missing or mistyped
// key, or one not in force that day, is an error naming it: a fallback set
// whose UMA lapsed never yields a silent estimate.
func ReadDeductionLimits[L ~string](c *Catalog, fiscalYear int) (DeductionLimits[L], error) {
	sel, err := c.Select(fiscalYear)
	if err != nil {
		return DeductionLimits[L]{}, err
	}
	yearEnd, err := fiscal.ParseDate(fmt.Sprintf("%04d-12-31", fiscalYear))
	if err != nil {
		return DeductionLimits[L]{}, fmt.Errorf("taxparams: fiscal year %d: %w", fiscalYear, err)
	}
	r := reader{p: sel.Params, on: yearEnd}
	limits := DeductionLimits[L]{
		AnnualUMA:               r.money(KeyAnnualUMA),
		GlobalCapUMAs:           r.count(KeyGlobalCapUMAs),
		GlobalCapIncomeShare:    r.rate(KeyGlobalCapIncomeShare),
		FuneralCapUMAs:          r.count(KeyFuneralCapUMAs),
		DonationShare:           r.rate(KeyDonationShare),
		GovernmentDonationShare: r.rate(KeyGovernmentDonationShare),
		RetirementIncomeShare:   r.rate(KeyRetirementIncomeShare),
		RetirementCapUMAs:       r.count(KeyRetirementCapUMAs),
		TuitionCaps:             make(map[L]money.Cents, len(TuitionLevels)),
	}
	for _, level := range TuitionLevels {
		limits.TuitionCaps[L(level)] = r.money(TuitionCapKey(level))
	}
	if err := errors.Join(r.problems...); err != nil {
		return DeductionLimits[L]{}, fmt.Errorf("taxparams: deduction limits %d: %w", fiscalYear, err)
	}
	return limits, nil
}

// DeductionLimitRequirements are the keys ReadDeductionLimits reads, for a
// per-year test of the embedded sets.
func DeductionLimitRequirements() []Requirement {
	reqs := []Requirement{
		RequireMoney(KeyAnnualUMA),
		RequireCount(KeyGlobalCapUMAs), RequireRate(KeyGlobalCapIncomeShare),
		RequireCount(KeyFuneralCapUMAs),
		RequireRate(KeyDonationShare), RequireRate(KeyGovernmentDonationShare),
		RequireRate(KeyRetirementIncomeShare), RequireCount(KeyRetirementCapUMAs),
	}
	for _, level := range TuitionLevels {
		reqs = append(reqs, RequireMoney(TuitionCapKey(level)))
	}
	return reqs
}

// reader resolves several keys on one day, collecting every failure.
type reader struct {
	p        *Params
	on       fiscal.Date
	problems []error
}

func (r *reader) keep(err error) {
	if err != nil {
		r.problems = append(r.problems, err)
	}
}

func (r *reader) money(key Key) money.Cents {
	v, err := r.p.Money(key, r.on)
	r.keep(err)
	return v
}

func (r *reader) rate(key Key) money.Rate {
	v, err := r.p.Rate(key, r.on)
	r.keep(err)
	return v
}

func (r *reader) count(key Key) int64 {
	v, err := r.p.Count(key, r.on)
	r.keep(err)
	return v
}
