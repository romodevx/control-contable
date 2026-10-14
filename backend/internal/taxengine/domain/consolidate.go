package domain

import (
	"fmt"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
)

// The line codes of the consolidated IVA, in the order ConsolidateIVA
// appends them; the definitive exclusion and the previous-month line appear
// only when non-zero.
const (
	LineIVATotalCharged    LineCode = "iva.total.charged"
	LineIVATotalExcluded   LineCode = "iva.total.definitive_excluded"
	LineIVATotalCreditable LineCode = "iva.total.creditable"
	LineIVATotalWithheld   LineCode = "iva.total.withheld"
	LineIVAConsolidated    LineCode = "iva.consolidated"
	LineIVAPreviousInFavor LineCode = "iva.previous_in_favor"
)

// minRegimesForNetting is the number of contributing regimes from which
// the consolidation nets one regime's IVA against another's (OQ-T2 ⚠️).
const minRegimesForNetting = 2

// The legal basis of each consolidation step (MODULE-tax-engine §2.5).
var (
	basisTotalWithheld = []Citation{
		{Law: "LIVA", Article: "1-A"}, {Law: "RLIVA", Article: "3", Note: "fracción I"},
		{Law: "LIVA", Article: "18-J", Note: "fracción II, inciso a)"},
	}
	basisConsolidated = []Citation{
		{Law: "LIVA", Article: "1", Note: "el contribuyente es el sujeto del impuesto"},
		{Law: "LIVA", Article: "5-D", Note: "pago mensual, no acumulado"},
	}
	basisPreviousInFavor = []Citation{{Law: "LIVA", Article: "6", Note: "saldo a favor no acreditado en esta estimación"}}
)

// IVAContribution is one regime's share of the taxpayer's monthly IVA: its
// components, and its exempt collections of the month (the signal of mixed
// activity across regimes).
type IVAContribution struct {
	Regime     fiscal.RegimeCode
	Components IVAComponents
	ExemptBase money.Cents
}

// ConsolidatedIVA is the taxpayer's one monthly IVA result across its
// regimes (LIVA 1, 5-D; MODULE-tax-engine §2.6), with its trace.
type ConsolidatedIVA struct {
	Result TaxResult
	Lines  []Line
	// Contributors counts the regimes with non-zero components.
	Contributors int
	CPATopics    []CPATopic
	Disclaimers  []DisclaimerCode
}

// ConsolidateIVA nets the regimes' components into one IVA (OQ-T2):
// Σ charged − Σ definitive excluded − Σ creditable − Σ withheld from the
// taxpayer, rounded once. Positive is payable, negative a saldo a favor.
//
//   - Two or more contributing regimes put the result under CPA review
//     (cross_regime_iva_netting, §2.9).
//   - Creditable IVA taken in full while some regime has exempt collections
//     raises iva_mixed_activity, which a single regime cannot see (§2.6).
//   - previousInFavor > 0 is the prior month's saldo a favor: it is not
//     carried forward in v1 (OQ-T5), only stated in an info line.
//
// Withholdings to remit are never netted here (LIVA 1-A). With no
// contribution the result is zero and there are no lines.
func ConsolidateIVA(contribs []IVAContribution, previousInFavor money.Cents) (ConsolidatedIVA, error) {
	if len(contribs) == 0 {
		return ConsolidatedIVA{Result: NewTaxResult(0)}, nil
	}
	total, exempt, contributors, err := sumContributions(contribs)
	if err != nil {
		return ConsolidatedIVA{}, fmt.Errorf("consolidate IVA: %w", err)
	}
	netting := contributors >= minRegimesForNetting
	exact, lines, err := traceConsolidation(total, netting, previousInFavor)
	if err != nil {
		return ConsolidatedIVA{}, fmt.Errorf("consolidate IVA: %w", err)
	}
	out := ConsolidatedIVA{Result: NewTaxResult(exact), Lines: lines, Contributors: contributors}
	if netting {
		out.CPATopics = []CPATopic{TopicCrossRegimeIVANetting}
	}
	if exempt > 0 && total.CreditablePaid > 0 {
		out.Disclaimers = []DisclaimerCode{DisclaimerIVAMixedActivity}
	}
	return out, nil
}

// sumContributions adds the components and exempt bases, and counts the
// regimes that contribute anything.
func sumContributions(contribs []IVAContribution) (total IVAComponents, exempt money.Cents, contributors int, err error) {
	var s summer
	for _, c := range contribs {
		if c.Components != (IVAComponents{}) {
			contributors++
		}
		total = IVAComponents{
			ChargedCollected:     s.add(total.ChargedCollected, c.Components.ChargedCollected),
			CreditablePaid:       s.add(total.CreditablePaid, c.Components.CreditablePaid),
			WithheldFromTaxpayer: s.add(total.WithheldFromTaxpayer, c.Components.WithheldFromTaxpayer),
			DefinitiveExcluded:   s.add(total.DefinitiveExcluded, c.Components.DefinitiveExcluded),
		}
		exempt = s.add(exempt, c.ExemptBase)
	}
	return total, exempt, contributors, s.err
}

// traceConsolidation appends the consolidation through a Ledger, so the
// lines are the math, and returns the exact result.
func traceConsolidation(total IVAComponents, netting bool, previousInFavor money.Cents) (money.Exact, []Line, error) {
	l := NewLedger()
	t := ivaTrace{l: l}
	l.Input(ivaStep(LineIVATotalCharged, SourceComputed(), basisCharged), t.exact(total.ChargedCollected))
	if total.DefinitiveExcluded != 0 {
		s := ivaStep(LineIVATotalExcluded, SourceComputed(), basisExclude)
		s.NeedsCPAReview = true
		l.Subtract(s, t.exact(total.DefinitiveExcluded))
	}
	l.Subtract(ivaStep(LineIVATotalCreditable, SourceComputed(), basisCreditable), t.exact(total.CreditablePaid))
	l.Subtract(ivaStep(LineIVATotalWithheld, SourceComputed(), basisTotalWithheld), t.exact(total.WithheldFromTaxpayer))
	result := ivaStep(LineIVAConsolidated, SourceComputed(), basisConsolidated)
	result.NeedsCPAReview = netting
	exact := l.Result(result)
	if previousInFavor > 0 {
		l.Info(ivaStep(LineIVAPreviousInFavor, SourceComputed(), basisPreviousInFavor), t.exact(previousInFavor))
	}
	if t.err != nil {
		return 0, nil, t.err
	}
	lines, err := l.Lines()
	return exact, lines, err
}
