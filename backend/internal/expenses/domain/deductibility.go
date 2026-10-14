package domain

// CategoryCode is an expense category code (expense_categories.code).
// Whether it exists is the use case's check (ErrUnknownCategory).
type CategoryCode string

// The categories with a rule in this package.
const (
	CategoryPredial          CategoryCode = "predial"                 // added to the 606 blind deduction
	CategoryProfessionalFees CategoryCode = "servicios_profesionales" // PM withholds ISR 10 % + 2/3 IVA
	CategoryRent             CategoryCode = "renta_local"             // PM withholds ISR 10 % + 2/3 IVA
	CategoryFreight          CategoryCode = "fletes"                  // PM withholds 4 % IVA
)

// CategoryRule is one category_regime_rules row for the payer's regime and
// person type (MODULE-expenses §2.2), passed in by the use case.
type CategoryRule struct {
	Category          CategoryCode
	Deductibility     Deductibility
	DefaultDeductible bool
	// ElectronicPaymentOnly marks a category never deductible when paid in
	// cash, whatever the amount (fuel, LISR 27 fr. III ⚠️ CPA).
	ElectronicPaymentOnly bool
}

// Options are the payer's regime options that change a hint (§2.3).
type Options struct {
	DeductionMode    DeductionMode // 606 deduction_mode
	DefinitiveOption bool          // 625 definitive_option
}

// Payer is the taxpayer who paid the expense.
type Payer struct {
	PersonType PersonType
	Regime     RegimeCode
	Options    Options
}

// Hint is a category's deductibility for the payer: what the form shows and
// the boolean a new expense starts with.
type Hint struct {
	Deductibility     Deductibility
	DefaultDeductible bool
}

// ResolveDeductibility applies MODULE-expenses §2.3 to a category rule:
//
//   - 626 física (RESICO PF): always isr_not_applicable; its IVA can still be
//     creditable.
//   - 625 with the definitive option: isr_not_applicable (LISR 113-B).
//   - 606 with the blind 35 % deduction: isr_not_applicable, except predial,
//     which is added to the 35 % (LISR 115 ¶2).
//
// A hint that does not lower ISR never defaults to deductible.
func ResolveDeductibility(rule CategoryRule, p Payer) Hint {
	if isrNotApplicable(rule.Category, p) {
		return Hint{Deductibility: ISRNotApplicable}
	}
	hint := Hint{Deductibility: rule.Deductibility, DefaultDeductible: rule.DefaultDeductible}
	if !hint.lowersISR() {
		hint.DefaultDeductible = false
	}
	return hint
}

func isrNotApplicable(category CategoryCode, p Payer) bool {
	switch {
	case p.Regime == RegimeRESICO && p.PersonType == PersonFisica:
		return true
	case p.Regime == RegimePlatforms && p.Options.DefinitiveOption:
		return true
	case p.Regime == RegimeRent && p.Options.DeductionMode == DeductionBlind35:
		return category != CategoryPredial
	}
	return false
}

// lowersISR reports a hint under which the expense may reduce ISR.
func (h Hint) lowersISR() bool {
	return h.Deductibility == Deductible || h.Deductibility == PartiallyDeductible
}
