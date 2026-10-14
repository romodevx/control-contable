package annualgolden

import (
	"cmp"
	"fmt"
	"slices"

	"github.com/romodevx/control-contable/backend/internal/platform/money"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain/annual"
)

// compare lists every way got departs from want: the section kinds in
// order first (nothing else is compared when they differ), then each
// section, then the estimate-level fields.
func compare(want Expected, got Result) []string {
	wantKinds := kindsOf(want.Sections, func(s SectionWant) annual.SectionKind { return s.Kind })
	gotKinds := kindsOf(got.Sections, func(s SectionResult) annual.SectionKind { return s.Kind })
	if !slices.Equal(wantKinds, gotKinds) {
		return []string{fmt.Sprintf("sections: want %v, got %v", wantKinds, gotKinds)}
	}
	var out []string
	for i, w := range want.Sections {
		out = append(out, compareSection(w, got.Sections[i])...)
	}
	if want.InformativeIncomeOver500k != nil && *want.InformativeIncomeOver500k != got.InformativeIncomeOver500k {
		out = append(out, fmt.Sprintf("informative_income_over_500k: want %t, got %t", *want.InformativeIncomeOver500k, got.InformativeIncomeOver500k))
	}
	if want.Disclaimers != nil && !sameSet(want.Disclaimers, got.Disclaimers) {
		out = append(out, fmt.Sprintf("disclaimers: want %v, got %v", want.Disclaimers, got.Disclaimers))
	}
	return out
}

func kindsOf[T any](sections []T, kind func(T) annual.SectionKind) []annual.SectionKind {
	out := make([]annual.SectionKind, len(sections))
	for i, s := range sections {
		out[i] = kind(s)
	}
	return out
}

func compareSection(w SectionWant, g SectionResult) []string {
	m := mismatches{prefix: string(w.Kind)}
	check(&m, "status", w.Status, g.Status)
	if w.Obligation != "" {
		check(&m, "obligation", w.Obligation, g.Obligation)
	}
	if w.Regimes != nil && !sameSet(w.Regimes, g.Regimes) {
		m.add("regimes", w.Regimes, g.Regimes)
	}
	if w.CPATopics != nil && !sameSet(w.CPATopics, g.CPATopics) {
		m.add("cpa_topics", w.CPATopics, g.CPATopics)
	}
	if w.MissingInputs != nil && !sameSet(w.MissingInputs, g.MissingInputs) {
		m.add("missing_inputs", w.MissingInputs, g.MissingInputs)
	}
	m.tax(w.Tax, g.Tax)
	m.optionalCents("difference", w.Difference, g.Difference)
	m.pmExtras(w.PMExtras, g.PMExtras)
	if w.Lines != nil && !subsequence(w.Lines, g.Lines) {
		m.add("lines (in this order)", w.Lines, g.Lines)
	}
	return m.out
}

// mismatches collects one section's differences, each named
// "<section> <field>".
type mismatches struct {
	prefix string
	out    []string
}

func (m *mismatches) add(field string, want, got any) {
	m.out = append(m.out, fmt.Sprintf("%s %s: want %v, got %v", m.prefix, field, want, got))
}

func check[T comparable](m *mismatches, field string, want, got T) {
	if want != got {
		m.add(field, want, got)
	}
}

func (m *mismatches) wantCents(field string, want *money.Cents, got money.Cents) {
	if want != nil {
		check(m, field, *want, got)
	}
}

func (m *mismatches) optionalCents(field string, want, got *money.Cents) {
	switch {
	case want == nil:
	case got == nil:
		m.add(field, *want, "none")
	default:
		check(m, field, *want, *got)
	}
}

func (m *mismatches) tax(w *TaxWant, g *TaxFigures) {
	if w == nil {
		return
	}
	if g == nil {
		m.add("tax", "figures", "none")
		return
	}
	m.wantCents("tax.accumulable_income", w.AccumulableIncome, g.AccumulableIncome)
	m.wantCents("tax.personal_deductions", w.PersonalDeductions, g.PersonalDeductions)
	m.wantCents("tax.base", w.Base, g.Base)
	m.wantCents("tax.isr", w.ISR, g.ISR)
	m.wantCents("tax.payments", w.Payments, g.Payments)
	m.wantCents("tax.withholdings", w.Withholdings, g.Withholdings)
	m.wantCents("tax.salary_isr", w.SalaryISR, g.SalaryISR)
	if w.Balance != nil {
		check(m, "tax.balance", *w.Balance, g.Balance)
	}
}

func (m *mismatches) pmExtras(w *PMExtras, g *PMExtrasResult) {
	if w == nil {
		return
	}
	if g == nil {
		m.add("pm_extras", "extras", "none")
		return
	}
	m.wantCents("pm_extras.ptu_base", w.PTUBase, g.PTUBase)
	m.wantCents("pm_extras.ptu_to_distribute", w.PTUToDistribute, g.PTUToDistribute)
	if w.SuggestedCU != nil {
		check(m, "pm_extras.suggested_cu_next_year", *w.SuggestedCU, g.SuggestedCU)
	}
}

// sameSet reports equal multisets.
func sameSet[T cmp.Ordered](want, got []T) bool {
	return slices.Equal(slices.Sorted(slices.Values(want)), slices.Sorted(slices.Values(got)))
}

// subsequence reports whether every wanted line appears in got, in order.
func subsequence(want, got []Line) bool {
	i := 0
	for _, g := range got {
		if i < len(want) && g == want[i] {
			i++
		}
	}
	return i == len(want)
}
