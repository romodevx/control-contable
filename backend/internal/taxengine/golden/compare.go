package golden

import (
	"fmt"
	"slices"
	"strings"

	"github.com/romodevx/control-contable/backend/internal/platform/money"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain"
)

// compare lists every way got differs from want, to the centavo, plus every
// line that breaks the citation rule (MODULE-tax-engine §2.5: every line
// except info and rounding_adjustment has a legal basis). Empty: a match.
func compare(want Expected, got domain.RegimeEstimate) []string {
	var out []string
	if s := got.Status(); s != want.Status {
		out = append(out, fmt.Sprintf("status: want %s, got %s", want.Status, s))
	}
	if want.asserts(domain.TaxISR) {
		if w, g := isrText(want.ISR), resultText(got.ISR); w != g {
			out = append(out, fmt.Sprintf("isr: want %s, got %s", w, g))
		}
	}
	if want.asserts(domain.TaxIVA) {
		if w, g := ivaText(want.IVA), ivaText(got.IVA); w != g {
			out = append(out, fmt.Sprintf("iva: want %s, got %s", w, g))
		}
	}
	if want.asserts(domain.TaxWithholding) && want.WithheldToRemit != got.WithheldToRemit {
		out = append(out, fmt.Sprintf("withheld_to_remit: want %s, got %s", want.WithheldToRemit, got.WithheldToRemit))
	}
	if want.Lines != nil {
		out = append(out, compareLines(want, got.Lines)...)
	}
	if want.MissingInputs != nil {
		out = appendSetDiff(out, "missing_inputs", missingTexts(want.MissingInputs), missingTexts(got.MissingInputs))
	}
	if want.CPATopics != nil {
		out = appendSetDiff(out, "cpa_topics", texts(want.CPATopics), texts(got.CPATopics))
	}
	if want.Disclaimers != nil {
		out = appendSetDiff(out, "disclaimers", texts(want.Disclaimers), texts(got.Disclaimers))
	}
	return append(out, uncited(got.Lines)...)
}

func isrText(c *money.Cents) string {
	if c == nil {
		return "none"
	}
	return c.String()
}

func resultText(r *domain.TaxResult) string {
	if r == nil {
		return "none"
	}
	return r.Amount.String()
}

func ivaText(c *domain.IVAComponents) string {
	if c == nil {
		return "none"
	}
	return fmt.Sprintf("{charged_collected %s, creditable_paid %s, withheld_from_taxpayer %s, definitive_excluded %s}",
		c.ChargedCollected, c.CreditablePaid, c.WithheldFromTaxpayer, c.DefinitiveExcluded)
}

// compareLines compares, in order, the lines of the asserted taxes.
func compareLines(want Expected, lines []domain.Line) []string {
	got := make([]LineWant, 0, len(lines))
	for _, l := range lines {
		if want.asserts(l.Tax) {
			got = append(got, LineWant{Tax: l.Tax, Code: l.Code, Amount: l.Amount})
		}
	}
	var out []string
	for i := range max(len(want.Lines), len(got)) {
		switch {
		case i >= len(got):
			out = append(out, fmt.Sprintf("line %d: missing %s", i+1, want.Lines[i]))
		case i >= len(want.Lines):
			out = append(out, fmt.Sprintf("line %d: unexpected %s", i+1, got[i]))
		case got[i] != want.Lines[i]:
			out = append(out, fmt.Sprintf("line %d: want %s, got %s", i+1, want.Lines[i], got[i]))
		}
	}
	return out
}

// String renders "isr isr.base 60000.00".
func (l LineWant) String() string { return fmt.Sprintf("%s %s %s", l.Tax, l.Code, l.Amount) }

func uncited(lines []domain.Line) []string {
	var out []string
	for _, l := range lines {
		if l.Operation != domain.OpInfo && l.Operation != domain.OpRoundingAdjustment && len(l.Basis) == 0 {
			out = append(out, fmt.Sprintf("line %d %s: no legal basis", l.Seq, l.Code))
		}
	}
	return out
}

func missingTexts(ms []domain.MissingInput) []string {
	out := make([]string, 0, len(ms))
	for _, m := range ms {
		months := make([]string, 0, len(m.Months))
		for _, month := range m.Months {
			months = append(months, fmt.Sprint(int(month)))
		}
		out = append(out, fmt.Sprintf("%s%v", m.Key, months))
	}
	return out
}

func texts[T ~string](vs []T) []string {
	out := make([]string, 0, len(vs))
	for _, v := range vs {
		out = append(out, string(v))
	}
	return out
}

// appendSetDiff compares two lists as sorted sets: order is the strategy's
// choice, membership is the contract.
func appendSetDiff(out []string, field string, want, got []string) []string {
	slices.Sort(want)
	slices.Sort(got)
	if !slices.Equal(want, got) {
		out = append(out, fmt.Sprintf("%s: want [%s], got [%s]", field, strings.Join(want, " "), strings.Join(got, " ")))
	}
	return out
}
