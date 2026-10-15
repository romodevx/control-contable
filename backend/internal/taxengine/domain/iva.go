package domain

import (
	"fmt"

	"github.com/romodevx/control-contable/backend/internal/platform/money"
)

// OptionIVAFullyTaxable is the regime option "every act of this regime is
// taxed" (612 PF, 601 PM, 626 PM; default true, MODULE-taxpayers §2.3).
const OptionIVAFullyTaxable OptionKey = "iva_fully_taxable"

// The line codes ComputeIVA appends, in this order; a zero optional line
// (8 %, 0 %, exempt, definitive exclusion, withheld IVA of the month
// before, platform withholding, to remit) is left out.
const (
	LineIVAChargedRate16      LineCode = "iva.charged.rate_16"
	LineIVAChargedRate8       LineCode = "iva.charged.rate_8"
	LineIVABaseRate0          LineCode = "iva.base.rate_0"
	LineIVABaseExempt         LineCode = "iva.base.exempt"
	LineIVACharged            LineCode = "iva.charged"
	LineIVADefinitiveExcluded LineCode = "iva.definitive_excluded"
	LineIVACreditable         LineCode = "iva.creditable"
	// LineIVACreditableWithheldPriorMonth is the IVA withheld from suppliers
	// the month before, credited this month (LIVA 5 fr. IV).
	LineIVACreditableWithheldPriorMonth LineCode = "iva.creditable.withheld_prior_month"
	LineIVAWithheldByMoral              LineCode = "iva.withheld.moral"
	LineIVAWithheldByPlatform           LineCode = "iva.withheld.platform"
	LineIVANet                          LineCode = "iva.regime_net"
	LineIVAWithheldToRemit              LineCode = "withholding.iva_to_remit"
)

// DefinitiveIVARule is the 625 hook (LIVA 18-L, 18-M ⚠️): given the
// month's income, it returns the IVA charged that the platform's
// withholding settles definitively, which leaves the monthly IVA. It must
// be pure; ComputeIVA refuses a value below zero or above the IVA charged.
type DefinitiveIVARule func(IncomeSums) (money.Cents, error)

// IVAInput is one regime's month as ComputeIVA reads it.
type IVAInput struct {
	Sums    MonthSums // this regime, the period's month only (IVA is never YTD, LIVA 5-D)
	Options Options   // reads OptionIVAFullyTaxable
	// WithheldPriorMonth is the creditable IVA the taxpayer withheld from
	// its suppliers the month before (YearSums.CreditableWithheldPriorMonth):
	// remitted with that month's return, it is credited in this one (LIVA 5
	// fr. IV). Sums.Expenses.CreditableIVAWithheld, this month's, waits for
	// the next month.
	WithheldPriorMonth money.Cents
	// Definitive is the 625 definitive-option rule; nil for every other
	// regime and for 625 without the option.
	Definitive DefinitiveIVARule
}

// IVAResult is one regime's IVA for the month: its components for the
// taxpayer's consolidation (M5-22, §2.6), never an amount to pay on its own.
type IVAResult struct {
	Components IVAComponents
	// Net is charged − excluded − creditable − withheld for this regime
	// alone: the regime's contribution to the consolidated IVA.
	Net money.Cents
	// WithheldToRemit is the IVA the taxpayer withheld from its suppliers:
	// remitted in full apart from its own IVA, never netted against it, and
	// credited only the month after (LIVA 5 fr. IV).
	WithheldToRemit money.Cents
	Disclaimers     []DisclaimerCode
	CPATopics       []CPATopic
}

// The legal basis of each IVA step (MODULE-tax-engine §2.5).
var (
	basisRate16 = []Citation{
		{Law: "LIVA", Article: "1", Note: "tasa general del 16 %"},
		{Law: "LIVA", Article: "1-B", Note: "IVA efectivamente cobrado"},
	}
	basisRate8 = []Citation{
		{
			Law: "Decreto de estímulos fiscales región fronteriza norte (DOF 31-12-2018)", Article: "Décimo Primero",
			Note: "crédito del 50 % de la tasa del IVA",
		},
		{
			Law: "Decreto de estímulos fiscales región fronteriza sur (DOF 30-12-2020)", Article: "Décimo",
			Note: "crédito del 50 % de la tasa del IVA",
		},
		{
			Law:     "Decreto que modifica los decretos de estímulos fiscales región fronteriza norte y sur (DOF 31-12-2025)",
			Article: "Primero y Segundo", Note: "vigencia hasta el 31-12-2026",
		},
		{Law: "LIVA", Article: "1-B", Note: "IVA efectivamente cobrado"},
	}
	basisRate0   = []Citation{{Law: "LIVA", Article: "2-A", Note: "tasa del 0 %"}}
	basisExempt  = []Citation{{Law: "LIVA", Article: "9"}, {Law: "LIVA", Article: "15"}, {Law: "LIVA", Article: "20"}}
	basisCharged = []Citation{{Law: "LIVA", Article: "1"}, {Law: "LIVA", Article: "1-B"}}
	basisExclude = []Citation{
		{Law: "LIVA", Article: "18-L"}, {Law: "LIVA", Article: "18-M", Note: "retención definitiva"},
	}
	basisCreditable = []Citation{{Law: "LIVA", Article: "5", Note: "fracciones I a V"}}
	basisWithheldIn = []Citation{{Law: "LIVA", Article: "5", Note: "fracción IV"}}
	basisByMoral    = []Citation{{Law: "LIVA", Article: "1-A"}, {Law: "RLIVA", Article: "3", Note: "fracción I"}}
	basisByPlatform = []Citation{{Law: "LIVA", Article: "18-J", Note: "fracción II, inciso a)"}}
	basisNet        = []Citation{{Law: "LIVA", Article: "5-D", Note: "pago mensual, no acumulado"}}
	basisToRemit    = []Citation{{Law: "LIVA", Article: "1-A", Note: "entero de la retención"}}
)

// ComputeIVA computes one regime's monthly IVA components through l, so
// the lines are the trace (§2.5, §2.6): the IVA charged and collected by
// rate (16 %, border 8 %, 0 % and exempt shown as bases), minus a 625
// definitive exclusion, minus the creditable IVA paid, minus the IVA it
// withheld from suppliers the month before (LIVA 5 fr. IV: credited in the
// return after the one that remitted it, assuming it was remitted on time,
// a CPA topic), minus the IVA withheld from the taxpayer; then the IVA it
// withheld from suppliers this month, to remit apart. Components'
// CreditablePaid is both creditable parts.
//
// Every figure is the records' own: rates and withholdings are never
// recomputed (§2.4). The creditable IVA is what the expenses module already
// resolved (category rules, the cash-payment limit). Proportional crediting
// for mixed taxed and exempt activity (LIVA 5 fr. V) is out of scope: the
// IVA is credited in full and iva_mixed_activity is added (§2.6).
func ComputeIVA(l *Ledger, in IVAInput) (IVAResult, error) {
	if l.err != nil {
		return IVAResult{}, fmt.Errorf("compute IVA: %w", l.err)
	}
	components, err := ivaComponents(in)
	if err != nil {
		return IVAResult{}, fmt.Errorf("compute IVA: %w", err)
	}
	t := ivaTrace{l: l}
	t.charged(in.Sums.Income.ByRate)
	net := t.net(components, ivaCredits{paid: in.Sums.Expenses.CreditableIVAPaid, withheldIn: in.WithheldPriorMonth},
		in.Sums.Income.IVAWithheld)
	t.toRemit(in.Sums.Expenses.IVAWithheldToRemit)
	if t.err == nil {
		t.err = l.err
	}
	if t.err != nil {
		return IVAResult{}, fmt.Errorf("compute IVA: %w", t.err)
	}

	result := IVAResult{
		Components: components, Net: net.Round(), WithheldToRemit: in.Sums.Expenses.IVAWithheldToRemit,
	}
	if mixedActivity(in, components.CreditablePaid) {
		result.Disclaimers = []DisclaimerCode{DisclaimerIVAMixedActivity}
	}
	if components.DefinitiveExcluded > 0 {
		result.CPATopics = append(result.CPATopics, TopicPlatformDefinitiveIVA)
	}
	if in.WithheldPriorMonth != 0 {
		result.CPATopics = append(result.CPATopics, TopicWithheldIVACreditTiming)
	}
	return result, nil
}

// monthIVA is the IVA input of in's period month: its sums and the IVA it
// withheld from suppliers the month before (LIVA 5 fr. IV).
func monthIVA(in Input) IVAInput {
	n := in.Period.Month()
	return IVAInput{Sums: in.Sums.Month(n), Options: in.Options, WithheldPriorMonth: in.Sums.CreditableWithheldPriorMonth(n)}
}

// ivaComponents sums the components and applies the definitive rule.
func ivaComponents(in IVAInput) (IVAComponents, error) {
	income := in.Sums.Income
	charged, err := income.ByRate.Rate16.IVA.Add(income.ByRate.Rate8.IVA)
	if err != nil {
		return IVAComponents{}, err
	}
	withheld, err := income.IVAWithheld.ByMoral.Add(income.IVAWithheld.ByPlatform)
	if err != nil {
		return IVAComponents{}, err
	}
	excluded, err := definitiveExclusion(in.Definitive, income, charged)
	if err != nil {
		return IVAComponents{}, err
	}
	creditable, err := in.Sums.Expenses.CreditableIVAPaid.Add(in.WithheldPriorMonth)
	if err != nil {
		return IVAComponents{}, err
	}
	return IVAComponents{
		ChargedCollected: charged, CreditablePaid: creditable,
		WithheldFromTaxpayer: withheld, DefinitiveExcluded: excluded,
	}, nil
}

func definitiveExclusion(rule DefinitiveIVARule, income IncomeSums, charged money.Cents) (money.Cents, error) {
	if rule == nil {
		return 0, nil
	}
	excluded, err := rule(income)
	if err != nil {
		return 0, fmt.Errorf("definitive IVA rule: %w", err)
	}
	if excluded < 0 || excluded > charged {
		return 0, fmt.Errorf("excluded %s of charged %s: %w", excluded, charged, ErrInvalidIVAExclusion)
	}
	return excluded, nil
}

// mixedActivity reports creditable IVA taken in full although the regime
// has exempt acts: the option says so, or the month has exempt collections.
func mixedActivity(in IVAInput, creditable money.Cents) bool {
	if creditable <= 0 {
		return false
	}
	fully, set := in.Options.Bool(OptionIVAFullyTaxable)
	return (set && !fully) || in.Sums.Income.ByRate.ExemptBase > 0
}

// ivaTrace appends the IVA steps, converting centavos to Exact with a
// sticky error; the Ledger keeps its own.
type ivaTrace struct {
	l   *Ledger
	err error
}

func (t *ivaTrace) exact(c money.Cents) money.Exact {
	if t.err != nil {
		return 0
	}
	e, err := c.Exact()
	t.err = err
	return e
}

func ivaStep(code LineCode, source Source, basis []Citation) Step {
	return Step{Tax: TaxIVA, Code: code, Label: LabelKey(code), Source: source, Basis: basis}
}

// charged traces the IVA charged and collected, by rate.
func (t *ivaTrace) charged(r IVAByRate) {
	t.l.Input(ivaStep(LineIVAChargedRate16, SourceRecords(), basisRate16), t.exact(r.Rate16.IVA))
	if r.Rate8 != (RateSums{}) {
		t.l.Add(ivaStep(LineIVAChargedRate8, SourceRecords(), basisRate8), t.exact(r.Rate8.IVA))
	}
	if r.Rate0Base != 0 {
		t.l.Info(ivaStep(LineIVABaseRate0, SourceRecords(), basisRate0), t.exact(r.Rate0Base))
	}
	if r.ExemptBase != 0 {
		t.l.Info(ivaStep(LineIVABaseExempt, SourceRecords(), basisExempt), t.exact(r.ExemptBase))
	}
	t.l.Result(ivaStep(LineIVACharged, SourceComputed(), basisCharged))
}

// ivaCredits are the two parts of the creditable IVA: the month's paid
// part and the IVA withheld from suppliers the month before.
type ivaCredits struct {
	paid, withheldIn money.Cents
}

// net traces charged − excluded − creditable − withheld and returns it.
func (t *ivaTrace) net(c IVAComponents, credits ivaCredits, withheld Withheld) money.Exact {
	if c.DefinitiveExcluded != 0 {
		s := ivaStep(LineIVADefinitiveExcluded, SourceComputed(), basisExclude)
		s.NeedsCPAReview = true
		t.l.Subtract(s, t.exact(c.DefinitiveExcluded))
	}
	t.l.Subtract(ivaStep(LineIVACreditable, SourceRecords(), basisCreditable), t.exact(credits.paid))
	if credits.withheldIn != 0 {
		s := ivaStep(LineIVACreditableWithheldPriorMonth, SourceRecords(), basisWithheldIn)
		s.NeedsCPAReview = true // assumes the withholding was remitted on time (OQ-W4)
		t.l.Subtract(s, t.exact(credits.withheldIn))
	}
	t.l.Subtract(ivaStep(LineIVAWithheldByMoral, SourceRecords(), basisByMoral), t.exact(withheld.ByMoral))
	if withheld.ByPlatform != 0 {
		t.l.Subtract(ivaStep(LineIVAWithheldByPlatform, SourceRecords(), basisByPlatform), t.exact(withheld.ByPlatform))
	}
	return t.l.Result(ivaStep(LineIVANet, SourceComputed(), basisNet))
}

// toRemit traces the IVA withheld from suppliers, a segment of its own.
func (t *ivaTrace) toRemit(amount money.Cents) {
	if amount == 0 {
		return
	}
	s := Step{
		Tax: TaxWithholding, Code: LineIVAWithheldToRemit, Label: LabelKey(LineIVAWithheldToRemit),
		Source: SourceRecords(), Basis: basisToRemit,
	}
	t.l.Input(s, t.exact(amount))
}
