package domain

import "slices"

// Treatment is a regime's support level (contract Treatment): the catalog's
// level for Strategy.Treatment, and what one estimate covered for
// RegimeEstimate.Coverage (625 without the definitive option is a full
// regime whose estimate is partial).
type Treatment string

// The contract's Treatment enum.
const (
	TreatmentFull          Treatment = "full"
	TreatmentPartial       Treatment = "partial"
	TreatmentInformational Treatment = "informational"
)

// Valid reports a contract Treatment.
func (t Treatment) Valid() bool {
	return slices.Contains([]Treatment{TreatmentFull, TreatmentPartial, TreatmentInformational}, t)
}

// Status is a RegimeEstimate's status (contract RegimeEstimate.status).
type Status string

// The contract's status enum (ADR-0002 Amendment 1).
const (
	StatusComplete      Status = "complete"
	StatusIncomplete    Status = "incomplete"
	StatusPartial       Status = "partial"
	StatusInformational Status = "informational"
)

// Valid reports a contract status.
func (s Status) Valid() bool {
	return slices.Contains([]Status{StatusComplete, StatusIncomplete, StatusPartial, StatusInformational}, s)
}

// TaxKind is the tax a breakdown line belongs to (contract BreakdownLine.tax).
type TaxKind string

// The contract's tax enum; withholding lines are the withholdings to remit.
const (
	TaxISR         TaxKind = "isr"
	TaxIVA         TaxKind = "iva"
	TaxWithholding TaxKind = "withholding"
)

// Valid reports a contract tax.
func (k TaxKind) Valid() bool { return slices.Contains([]TaxKind{TaxISR, TaxIVA, TaxWithholding}, k) }

// Operation is what a breakdown line does (contract BreakdownLine.operation).
type Operation string

// The contract's operation enum.
const (
	OpInput              Operation = "input"
	OpAdd                Operation = "add"
	OpSubtract           Operation = "subtract"
	OpMultiply           Operation = "multiply"
	OpLookup             Operation = "lookup"
	OpResult             Operation = "result"
	OpRoundingAdjustment Operation = "rounding_adjustment"
	OpInfo               Operation = "info"
)

// Valid reports a contract operation.
func (o Operation) Valid() bool {
	return slices.Contains([]Operation{
		OpInput, OpAdd, OpSubtract, OpMultiply, OpLookup, OpResult, OpRoundingAdjustment, OpInfo,
	}, o)
}

// Outcome is the sign of a TaxResult (contract TaxResult.outcome).
type Outcome string

// The contract's outcome enum.
const (
	OutcomePayable Outcome = "payable"
	OutcomeInFavor Outcome = "in_favor"
	OutcomeZero    Outcome = "zero"
)

// CPATopic is a rule a CPA has not validated yet (contract CpaTopic,
// append-only, MODULE-tax-engine §2.9).
type CPATopic string

// The contract's CpaTopic enum.
const (
	TopicRIFReduction                CPATopic = "rif_reduction"
	TopicRIFBimonthlyIVA             CPATopic = "rif_bimonthly_iva"
	TopicPlatformDefinitiveIVA       CPATopic = "platform_definitive_iva"
	TopicCrossRegimeIVANetting       CPATopic = "cross_regime_iva_netting"
	TopicRESICOPMInterestWithholding CPATopic = "resico_pm_interest_withholding"
)

// Valid reports a contract CpaTopic.
func (c CPATopic) Valid() bool {
	return slices.Contains([]CPATopic{
		TopicRIFReduction, TopicRIFBimonthlyIVA, TopicPlatformDefinitiveIVA,
		TopicCrossRegimeIVANetting, TopicRESICOPMInterestWithholding,
	}, c)
}

// Severity is how strongly a disclaimer is shown (contract
// Disclaimer.severity).
type Severity string

// The contract's severity enum.
const (
	SeverityInfo    Severity = "info"
	SeverityWarning Severity = "warning"
	SeverityStrong  Severity = "strong"
)

// Valid reports a contract severity.
func (s Severity) Valid() bool {
	return slices.Contains([]Severity{SeverityInfo, SeverityWarning, SeverityStrong}, s)
}

// DisclaimerCode names a disclaimer (contract DisclaimerCode, append-only).
// Its text is versioned with the parameter set, not here.
type DisclaimerCode string

// The contract's DisclaimerCode enum (MODULE-tax-engine §2.3).
const (
	DisclaimerNotTaxAdvice           DisclaimerCode = "not_tax_advice"
	DisclaimerIncompleteInputs       DisclaimerCode = "incomplete_inputs"
	DisclaimerProvisionalParameters  DisclaimerCode = "provisional_parameters"
	DisclaimerInformationalRegime    DisclaimerCode = "informational_regime"
	DisclaimerPartialRegime          DisclaimerCode = "partial_regime"
	DisclaimerDeductibilityClaimed   DisclaimerCode = "deductibility_claimed"
	DisclaimerNoCoefficientFirstYear DisclaimerCode = "no_coefficient_first_year"
	DisclaimerIVAMixedActivity       DisclaimerCode = "iva_mixed_activity"
	DisclaimerCPAReviewPending       DisclaimerCode = "cpa_review_pending"
	// DisclaimerRESICOCapExceeded: a 626 regime collected more than its
	// annual cap; the taxpayer leaves RESICO (PF next month, LISR 113-E and
	// 113-I; PM next fiscal year, LISR 214).
	DisclaimerRESICOCapExceeded DisclaimerCode = "resico_cap_exceeded"
)

// disclaimerSeverity is the fixed severity of each code (§2.3).
var disclaimerSeverity = map[DisclaimerCode]Severity{
	DisclaimerNotTaxAdvice:           SeverityInfo,
	DisclaimerIncompleteInputs:       SeverityWarning,
	DisclaimerProvisionalParameters:  SeverityWarning,
	DisclaimerInformationalRegime:    SeverityInfo,
	DisclaimerPartialRegime:          SeverityInfo,
	DisclaimerDeductibilityClaimed:   SeverityInfo,
	DisclaimerNoCoefficientFirstYear: SeverityInfo,
	DisclaimerIVAMixedActivity:       SeverityWarning,
	DisclaimerCPAReviewPending:       SeverityStrong,
	DisclaimerRESICOCapExceeded:      SeverityWarning,
}

// Severity returns the code's severity, or "" for an unknown code.
func (c DisclaimerCode) Severity() Severity { return disclaimerSeverity[c] }

// Valid reports a contract DisclaimerCode.
func (c DisclaimerCode) Valid() bool { return c.Severity() != "" }
