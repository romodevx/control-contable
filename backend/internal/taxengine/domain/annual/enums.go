package annual

import "slices"

// SectionKind groups the regimes of one annual estimate (MODULE-annual §2.2).
type SectionKind string

// The section kinds, in the fixed order an estimate lists them.
const (
	SectionPFArt152      SectionKind = "pf_art152"
	SectionPFRESICOCheck SectionKind = "pf_resico_check"
	SectionPM601         SectionKind = "pm_601"
	SectionPMRESICO      SectionKind = "pm_resico"
	SectionExplanation   SectionKind = "explanation"
)

// Valid reports a spec section kind.
func (k SectionKind) Valid() bool {
	return slices.Contains([]SectionKind{
		SectionPFArt152, SectionPFRESICOCheck, SectionPM601, SectionPMRESICO, SectionExplanation,
	}, k)
}

// Status is a section's status (MODULE-annual §2.3). Unlike the monthly
// estimate there is no "partial".
type Status string

// The section statuses; precedence informational > incomplete > complete.
const (
	StatusComplete      Status = "complete"
	StatusIncomplete    Status = "incomplete"
	StatusInformational Status = "informational"
)

// Valid reports a spec section status.
func (s Status) Valid() bool {
	return slices.Contains([]Status{StatusComplete, StatusIncomplete, StatusInformational}, s)
}

// Obligation is whether the taxpayer must file the section's return
// (MODULE-annual §2.4).
type Obligation string

// The obligation values.
const (
	ObligationObligated    Obligation = "obligated"
	ObligationOptional     Obligation = "optional"
	ObligationNotObligated Obligation = "not_obligated"
	ObligationRelieved     Obligation = "relieved"
)

// Valid reports a spec obligation.
func (o Obligation) Valid() bool {
	return slices.Contains([]Obligation{
		ObligationObligated, ObligationOptional, ObligationNotObligated, ObligationRelieved,
	}, o)
}

// BalanceKind is the sign of a section's balance (MODULE-annual §2.1).
type BalanceKind string

// The balance kinds.
const (
	BalanceACargo BalanceKind = "a_cargo"
	BalanceAFavor BalanceKind = "a_favor"
	BalanceCero   BalanceKind = "cero"
)

// Valid reports a spec balance kind.
func (b BalanceKind) Valid() bool {
	return slices.Contains([]BalanceKind{BalanceACargo, BalanceAFavor, BalanceCero}, b)
}

// CPATopic is an annual rule a CPA has not validated yet (MODULE-annual
// §2.10). The monthly topics are domain.CPATopic.
type CPATopic string

// The annual CPA topics.
const (
	TopicPlatformNonDefinitiveAnnual CPATopic = "platform_nondefinitive_annual"
	TopicUMAForMinimumWage           CPATopic = "uma_for_minimum_wage"
	TopicDeductionCapRESICOIncome    CPATopic = "deduction_cap_resico_income"
	TopicSalaryObligationBasis       CPATopic = "salary_obligation_basis"
	TopicCURounding                  CPATopic = "cu_rounding"
	TopicRESICOPMPTUBase             CPATopic = "resico_pm_ptu_base"
	TopicRESICOPFReliefRenewal       CPATopic = "resico_pf_relief_renewal"
	// TopicPTURate: the 10 % PTU rate is from a secondary source; the
	// Comisión Nacional PTU resolution (DOF) is not in params (ANN-05).
	TopicPTURate CPATopic = "ptu_rate"
)

// Valid reports a spec annual CPA topic.
func (c CPATopic) Valid() bool {
	return slices.Contains([]CPATopic{
		TopicPlatformNonDefinitiveAnnual, TopicUMAForMinimumWage, TopicDeductionCapRESICOIncome,
		TopicSalaryObligationBasis, TopicCURounding, TopicRESICOPMPTUBase, TopicRESICOPFReliefRenewal, TopicPTURate,
	}, c)
}

// Severity is how strongly a disclaimer is shown.
type Severity string

// The severities.
const (
	SeverityInfo    Severity = "info"
	SeverityWarning Severity = "warning"
	SeverityStrong  Severity = "strong"
)

// DisclaimerCode names an annual disclaimer (MODULE-annual §2.8). Its text
// is versioned with the parameter set, not here.
type DisclaimerCode string

// The annual disclaimer codes.
const (
	DisclaimerAnnualEstimate              DisclaimerCode = "annual_estimate"
	DisclaimerCPAReviewPending            DisclaimerCode = "cpa_review_pending"
	DisclaimerYearInProgress              DisclaimerCode = "year_in_progress"
	DisclaimerPaymentsEstimated           DisclaimerCode = "payments_estimated"
	DisclaimerIncompleteInputs            DisclaimerCode = "incomplete_inputs"
	DisclaimerDeductionsAsEntered         DisclaimerCode = "deductions_as_entered"
	DisclaimerRESICOPFDefinitive          DisclaimerCode = "resico_pf_definitive"
	DisclaimerInformativeIncome           DisclaimerCode = "informative_income"
	DisclaimerRMF2027Pending              DisclaimerCode = "rmf_2027_pending"
	DisclaimerInflationAdjustmentExcluded DisclaimerCode = "inflation_adjustment_excluded"
	DisclaimerNoAnnualIVA                 DisclaimerCode = "no_annual_iva"
)

// disclaimerSeverity is the fixed severity of each code (§2.8).
var disclaimerSeverity = map[DisclaimerCode]Severity{
	DisclaimerAnnualEstimate:              SeverityInfo,
	DisclaimerCPAReviewPending:            SeverityStrong,
	DisclaimerYearInProgress:              SeverityWarning,
	DisclaimerPaymentsEstimated:           SeverityWarning,
	DisclaimerIncompleteInputs:            SeverityWarning,
	DisclaimerDeductionsAsEntered:         SeverityInfo,
	DisclaimerRESICOPFDefinitive:          SeverityInfo,
	DisclaimerInformativeIncome:           SeverityInfo,
	DisclaimerRMF2027Pending:              SeverityInfo,
	DisclaimerInflationAdjustmentExcluded: SeverityInfo,
	DisclaimerNoAnnualIVA:                 SeverityInfo,
}

// Severity returns the code's severity, or "" for an unknown code.
func (c DisclaimerCode) Severity() Severity { return disclaimerSeverity[c] }

// Valid reports a spec annual disclaimer code.
func (c DisclaimerCode) Valid() bool { return c.Severity() != "" }
