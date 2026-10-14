package domain

import (
	"cmp"
	"slices"
)

// Disclaimer is one notice of an estimate (contract Disclaimer): its code,
// its fixed severity and, for cpa_review_pending, the topics it names. The
// text is versioned with the parameter set and rendered by the adapter.
type Disclaimer struct {
	Code     DisclaimerCode
	Severity Severity
	Topics   []CPATopic
}

// DisclaimerFacts are what an estimate's disclaimers depend on.
type DisclaimerFacts struct {
	Regimes []RegimeEstimate
	IVA     ConsolidatedIVA
	// ProvisionalParameters: the period's fiscal year has no published
	// set, or the set is provisional (ADR-0002 §3).
	ProvisionalParameters bool
}

// severityRank orders severities strong, warning, info (§2.3).
var severityRank = map[Severity]int{SeverityStrong: 0, SeverityWarning: 1, SeverityInfo: 2}

// statusDisclaimer is the code each non-complete status raises (§2.3).
var statusDisclaimer = map[Status]DisclaimerCode{
	StatusIncomplete:    DisclaimerIncompleteInputs,
	StatusPartial:       DisclaimerPartialRegime,
	StatusInformational: DisclaimerInformationalRegime,
}

// EstimateDisclaimers returns the disclaimers of an estimate, each code
// once, in the contract order: not_tax_advice first (always present), then
// by severity (strong, warning, info), then by code.
//
//   - incomplete_inputs, partial_regime, informational_regime: any regime
//     with that status (PR #78: zero-filled figures always carry
//     "se usaron ceros");
//   - each regime's own codes and the consolidated IVA's;
//   - provisional_parameters when the parameters are provisional;
//   - cpa_review_pending (strong, OQ-T8) when a regime or the consolidated
//     IVA applies a rule no CPA has validated, naming every topic once.
func EstimateDisclaimers(f DisclaimerFacts) []Disclaimer {
	raised := map[DisclaimerCode]bool{DisclaimerNotTaxAdvice: true}
	topics := slices.Clone(f.IVA.CPATopics)
	for _, r := range f.Regimes {
		if code, ok := statusDisclaimer[r.Status()]; ok {
			raised[code] = true
		}
		for _, code := range r.Disclaimers {
			raised[code] = true
		}
		topics = append(topics, r.CPATopics...)
	}
	for _, code := range f.IVA.Disclaimers {
		raised[code] = true
	}
	raised[DisclaimerProvisionalParameters] = f.ProvisionalParameters
	raised[DisclaimerCPAReviewPending] = len(topics) > 0

	out := make([]Disclaimer, 0, len(raised))
	for code, on := range raised {
		if on {
			out = append(out, Disclaimer{Code: code, Severity: code.Severity()})
		}
	}
	slices.SortFunc(out, compareDisclaimers)
	slices.Sort(topics)
	topics = slices.Compact(topics)
	for i := range out {
		if out[i].Code == DisclaimerCPAReviewPending {
			out[i].Topics = topics
		}
	}
	return out
}

// compareDisclaimers orders not_tax_advice first, then severity, then code.
func compareDisclaimers(a, b Disclaimer) int {
	first := func(d Disclaimer) int {
		if d.Code == DisclaimerNotTaxAdvice {
			return 0
		}
		return 1
	}
	return cmp.Or(
		cmp.Compare(first(a), first(b)),
		cmp.Compare(severityRank[a.Severity], severityRank[b.Severity]),
		cmp.Compare(a.Code, b.Code),
	)
}
