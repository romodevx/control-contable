// Package explain is the annual estimate's informative part (MODULE-annual
// §1, §2.8–§2.10; PLAN ANN-12): the explanation sections of the regimes the
// app does not compute (no annual for this income, or "tu contador lo
// calcula" with the deadline), the deadlines read from the parameter set,
// the > $500,000 informative-income note (LISR 150 ¶3), the disclaimers
// with their fixed order, and the es-MX catalog of the legend, disclaimer
// texts, CPA-topic phrases and explanation texts.
//
// Like packages annual and pm it is pure: (input, parameters) only, no
// clock; moving a legal date to the next business day is the M5 calendar's
// job in the use case (ANN-14).
package explain

import (
	"cmp"
	"slices"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain/annual"
)

// Code is what an explanation section says; ExplanationText renders it.
type Code string

// The explanation codes.
const (
	// CodePlatformDefinitive: 625 with the definitive option, whose
	// withholdings were the tax (LISR 113-B, 152 ¶2).
	CodePlatformDefinitive Code = "explanation.platform_definitive"
	// CodePrizesDefinitive: 615, the prize withholding is definitive.
	CodePrizesDefinitive Code = "explanation.prizes_definitive"
	// CodeNoObligations: 616, no tax obligations.
	CodeNoObligations Code = "explanation.no_obligations"
	// CodeForeignResident: 610, definitive withholding (Título V).
	CodeForeignResident Code = "explanation.foreign_resident"
	// CodeNonprofitInformative: 603 files an informative return.
	CodeNonprofitInformative Code = "explanation.nonprofit_informative"
	// CodeCPAComputes: the regime files an annual return the app does not
	// compute; the CPA does.
	CodeCPAComputes Code = "explanation.cpa_computes"
	// CodeCPAConfirms: whether the regime files depends on facts the app
	// does not hold (611, 614, 621); the CPA confirms.
	CodeCPAConfirms Code = "explanation.cpa_confirms"
	// CodeAskCPA: a regime outside the research tables.
	CodeAskCPA Code = "explanation.ask_cpa"
)

// Held is one regime the taxpayer held in the fiscal year; DefinitiveOption
// is the 625 M5 tax input definitive_option.
type Held struct {
	Regime           fiscal.RegimeCode
	DefinitiveOption bool
}

// Section is one explanation section: no figure, informational status, the
// obligation when the law settles it ("" when it depends on facts the app
// does not hold) and the deadline to show.
type Section struct {
	Kind       annual.SectionKind
	Regime     fiscal.RegimeCode
	Status     annual.Status
	Obligation annual.Obligation
	Code       Code
	Deadline   DeadlineKind
	Basis      []domain.Citation
}

// explanation is one row of the regime tables.
type explanation struct {
	code       Code
	obligation annual.Obligation
	deadline   DeadlineKind
	basis      []domain.Citation
}

const lisr = "LISR"

// noAnnual are the regimes with no annual return for their income, for any
// person type (research Part A).
var noAnnual = map[fiscal.RegimeCode]explanation{
	"615": {CodePrizesDefinitive, annual.ObligationNotObligated, DeadlineNone, []domain.Citation{
		{Law: lisr, Article: "150", Note: "primer y tercer párrafos"}, {Law: lisr, Article: "152", Note: "segundo párrafo"},
	}},
	"616": {CodeNoObligations, annual.ObligationNotObligated, DeadlineNone, nil},
	"610": {CodeForeignResident, annual.ObligationNotObligated, DeadlineNone, []domain.Citation{
		{Law: lisr, Article: "153", Note: "Título V, retención definitiva"},
	}},
}

// platformDefinitive is 625 with the definitive option (research §A.1).
var platformDefinitive = explanation{CodePlatformDefinitive, annual.ObligationNotObligated, DeadlineNone, []domain.Citation{
	{Law: lisr, Article: "113-B", Note: "opción de pago definitivo"},
	{Law: lisr, Article: "150", Note: "primer párrafo"},
	{Law: lisr, Article: "152", Note: "segundo párrafo"},
}}

// byPerson are the explanation-only regimes of each person type (research
// Part A, MODULE-annual §1). 614 is explanation-only in v1 (OQ-A6).
var byPerson = map[fiscal.PersonType]map[fiscal.RegimeCode]explanation{
	fiscal.PersonFisica: {
		"607": {CodeCPAComputes, annual.ObligationObligated, DeadlinePF, []domain.Citation{
			{Law: lisr, Article: "150", Note: "primer párrafo"}, {Law: lisr, Article: "152", Note: "primer párrafo, Capítulo IV"},
		}},
		"608": {CodeCPAComputes, annual.ObligationObligated, DeadlinePF, []domain.Citation{
			{Law: lisr, Article: "152", Note: "primer párrafo, Capítulo IX"},
		}},
		"611": {CodeCPAConfirms, "", DeadlinePF, []domain.Citation{
			{Law: lisr, Article: "140"}, {Law: lisr, Article: "152", Note: "primer párrafo, Capítulo VIII"},
		}},
		"614": {CodeCPAConfirms, "", DeadlinePF, []domain.Citation{
			{Law: lisr, Article: "135", Note: "intereses"}, {Law: lisr, Article: "136"}, {Law: lisr, Article: "150", Note: "segundo párrafo"},
			{Law: "RMF 2026", Article: "3.17.2", Note: "intereses reales mayores a 100,000 pesos"},
		}},
		"621": {CodeCPAConfirms, "", DeadlinePF, []domain.Citation{
			{Law: lisr, Article: "111", Note: "texto vigente hasta 2021, Transitorio Segundo fr. IX DOF 12-11-2021"},
		}},
	},
	fiscal.PersonMoral: {
		"603": {CodeNonprofitInformative, annual.ObligationObligated, DeadlineNonprofitInformative, []domain.Citation{
			{Law: lisr, Article: "86", Note: "fracción III, declaración informativa"},
		}},
		"620": {CodeCPAComputes, annual.ObligationObligated, DeadlineNone, []domain.Citation{
			{Law: lisr, Article: "194", Note: "fracción I, impuesto por cada socio"},
		}},
		"622": {CodeCPAComputes, annual.ObligationObligated, DeadlinePM, []domain.Citation{
			{Law: lisr, Article: "74"}, {Law: lisr, Article: "76", Note: "fracción V"},
		}},
		"623": {CodeCPAComputes, annual.ObligationObligated, DeadlinePM, []domain.Citation{{Law: lisr, Article: "64"}}},
		"624": {CodeCPAComputes, annual.ObligationObligated, DeadlinePM, []domain.Citation{
			{Law: lisr, Article: "72"}, {Law: lisr, Article: "73"},
		}},
	},
}

// computed are the regimes another section covers (pf_art152,
// pf_resico_check, pm_601, pm_resico); 625 only without the definitive
// option.
var computed = map[fiscal.PersonType][]fiscal.RegimeCode{
	fiscal.PersonFisica: {"605", "606", "612", "625", fiscal.RegimeRESICO},
	fiscal.PersonMoral:  {"601", fiscal.RegimeRESICO},
}

var askCPA = explanation{code: CodeAskCPA}

// Sections returns one explanation section per explanation-only regime
// held, in regime-code order, each regime once. O(n log n).
func Sections(person fiscal.PersonType, held []Held) []Section {
	sorted := slices.SortedFunc(slices.Values(held), func(a, b Held) int { return cmp.Compare(a.Regime, b.Regime) })
	sorted = slices.CompactFunc(sorted, func(a, b Held) bool { return a.Regime == b.Regime })
	var out []Section
	for _, h := range sorted {
		e, ok := explanationOf(person, h)
		if !ok {
			continue
		}
		out = append(out, Section{
			Kind: annual.SectionExplanation, Regime: h.Regime, Status: annual.StatusInformational,
			Obligation: e.obligation, Code: e.code, Deadline: e.deadline, Basis: slices.Clone(e.basis),
		})
	}
	return out
}

// explanationOf returns h's explanation, or false when another section
// computes it.
func explanationOf(person fiscal.PersonType, h Held) (explanation, bool) {
	if h.Regime == "625" && h.DefinitiveOption && person == fiscal.PersonFisica {
		return platformDefinitive, true
	}
	if slices.Contains(computed[person], h.Regime) {
		return explanation{}, false
	}
	if e, ok := noAnnual[h.Regime]; ok {
		return e, true
	}
	if e, ok := byPerson[person][h.Regime]; ok {
		return e, true
	}
	return askCPA, true
}
