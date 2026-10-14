package domain

import "fmt"

// partialPMEstimate is the estimate of a partial persona moral regime whose
// ISR the app does not estimate (622, 603; MODULE-tax-engine §2.4): the
// month's IVA through the shared components, the ISR and IVA the company
// withheld from its suppliers as withholdings to remit (as 601 does), no ISR
// figure, and the catalog's explanation of what is not estimated.
func partialPMEstimate(key StrategyKey, in Input) (RegimeEstimate, error) {
	l := NewLedger()
	month := in.Sums.Month(in.Period.Month())
	iva, err := ComputeIVA(l, IVAInput{Sums: month, Options: in.Options})
	if err != nil {
		return RegimeEstimate{}, fmt.Errorf("%s %s: %w", key, in.Period, err)
	}
	toRemit := generalPMWithholdingsToRemit(l, iva.WithheldToRemit, month.Expenses.ISRWithheldToRemit)
	lines, err := l.Lines()
	if err != nil {
		return RegimeEstimate{}, fmt.Errorf("%s %s: %w", key, in.Period, err)
	}
	return RegimeEstimate{
		Key:             key,
		Coverage:        TreatmentPartial,
		IVA:             &iva.Components,
		WithheldToRemit: toRemit,
		Lines:           lines,
		Explanation:     RegimeExplanation(key),
		CPATopics:       iva.CPATopics,
		Disclaimers:     iva.Disclaimers,
	}, nil
}
