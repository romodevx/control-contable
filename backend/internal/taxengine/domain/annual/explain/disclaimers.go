package explain

import (
	"cmp"
	"fmt"
	"slices"

	"github.com/romodevx/control-contable/backend/internal/taxengine/domain/annual"
)

// Disclaimer is one notice of an annual estimate: its code, its fixed
// severity and, for cpa_review_pending, the topics it names.
type Disclaimer struct {
	Code     annual.DisclaimerCode
	Severity annual.Severity
	Topics   []annual.CPATopic
}

// DisclaimerFacts are what the estimate's disclaimers depend on: the codes
// its sections and facts raised (year_in_progress, payments_estimated,
// incomplete_inputs, deductions_as_entered, resico_pf_definitive,
// informative_income, inflation_adjustment_excluded), the CPA topics of
// its sections, and whether the next year's RMF, which sets the annual
// facilities, is still unpublished (rmf_2027_pending for fiscal year 2026).
type DisclaimerFacts struct {
	Raised     []annual.DisclaimerCode
	Topics     []annual.CPATopic
	RMFPending bool
}

// severityRank orders severities strong, warning, info.
var severityRank = map[annual.Severity]int{annual.SeverityStrong: 0, annual.SeverityWarning: 1, annual.SeverityInfo: 2}

// Disclaimers returns the estimate's disclaimers, each code once, in the
// spec order (MODULE-annual §2.8): annual_estimate first, then strong,
// warning and info, then by code. annual_estimate and no_annual_iva are
// always present; cpa_review_pending (strong) names every topic once, in
// code order. An unknown code or topic is ErrUnknownCode.
func Disclaimers(f DisclaimerFacts) ([]Disclaimer, error) {
	raised := map[annual.DisclaimerCode]bool{annual.DisclaimerAnnualEstimate: true, annual.DisclaimerNoAnnualIVA: true}
	for _, c := range f.Raised {
		if !c.Valid() || c == annual.DisclaimerCPAReviewPending {
			return nil, fmt.Errorf("disclaimer %q: %w", c, ErrUnknownCode)
		}
		raised[c] = true
	}
	topics := slices.Sorted(slices.Values(f.Topics))
	topics = slices.Compact(topics)
	for _, t := range topics {
		if !t.Valid() {
			return nil, fmt.Errorf("CPA topic %q: %w", t, ErrUnknownCode)
		}
	}
	raised[annual.DisclaimerCPAReviewPending] = len(topics) > 0
	raised[annual.DisclaimerRMF2027Pending] = f.RMFPending

	out := make([]Disclaimer, 0, len(raised))
	for code, on := range raised {
		if !on {
			continue
		}
		d := Disclaimer{Code: code, Severity: code.Severity()}
		if code == annual.DisclaimerCPAReviewPending {
			d.Topics = topics
		}
		out = append(out, d)
	}
	slices.SortFunc(out, compareDisclaimers)
	return out, nil
}

func compareDisclaimers(a, b Disclaimer) int {
	notFirst := func(d Disclaimer) bool { return d.Code != annual.DisclaimerAnnualEstimate }
	return cmp.Or(
		compareBool(notFirst(a), notFirst(b)),
		cmp.Compare(severityRank[a.Severity], severityRank[b.Severity]),
		cmp.Compare(a.Code, b.Code),
	)
}

func compareBool(a, b bool) int {
	switch {
	case a == b:
		return 0
	case a:
		return 1
	}
	return -1
}
