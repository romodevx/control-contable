package domain

import (
	"cmp"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"slices"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
)

// FiguresInput is what one month's estimate shows: the taxpayer totals,
// the due date and each held regime's estimate.
type FiguresInput struct {
	Totals  Totals
	DueDate fiscal.Date
	Regimes []RegimeEstimate
}

// figuresScheme names the canonical text below. Changing the text marks
// every month a user viewed (M8), so it needs an ADR-0002 amendment and a
// re-baseline of estimate_views in the same migration (MODULE-history §2.4).
const figuresScheme = "figures/1"

// FiguresFingerprint is the result-figures hash (MODULE-history §2.4): the
// lower-case hex SHA-256 of a canonical text of the figures the user sees
// for the month. Totals (ISR, IVA, withholdings to remit, total payable),
// the due date, and per regime its key, status, ISR, IVA components,
// withholdings to remit and whether a CPA review is pending. Versions,
// texts, explanations, disclaimers, breakdown lines and sub-centavo
// precision are left out, and the regimes' order does not matter, so an
// engine bump or record edit that leaves every figure identical keeps the
// hash (DECIDED OQ-H4). O(regimes log regimes).
func FiguresFingerprint(in FiguresInput) string {
	t := in.Totals
	b := fmt.Appendf(nil, "%s\ndue %q\ntotals isr %d iva %d remit %d total %d\n",
		figuresScheme, in.DueDate.String(), t.ISR.Amount, t.IVA.Amount, t.WithholdingsToRemit, t.TotalPayable)
	regimes := slices.Clone(in.Regimes)
	slices.SortFunc(regimes, func(a, b RegimeEstimate) int { return cmp.Compare(a.Key.String(), b.Key.String()) })
	for _, r := range regimes {
		b = appendFigures(b, r)
	}
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

// appendFigures writes one regime's shown figures; an absent ISR or IVA is
// "-", distinct from a zero.
func appendFigures(b []byte, r RegimeEstimate) []byte {
	b = fmt.Appendf(b, "regime %q status %q cpa %t remit %d isr ", r.Key.String(), r.Status(), r.CPAReviewPending(), r.WithheldToRemit)
	if r.ISR == nil {
		b = append(b, '-')
	} else {
		b = fmt.Appendf(b, "%d", r.ISR.Amount)
	}
	b = append(b, " iva "...)
	if r.IVA == nil {
		b = append(b, '-')
	} else {
		v := r.IVA
		b = fmt.Appendf(b, "%d %d %d %d", v.ChargedCollected, v.CreditablePaid, v.WithheldFromTaxpayer, v.DefinitiveExcluded)
	}
	return append(b, '\n')
}
