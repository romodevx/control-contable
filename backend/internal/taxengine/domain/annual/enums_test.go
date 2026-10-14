package annual_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/romodevx/control-contable/backend/internal/taxengine/domain/annual"
)

type validator interface{ Valid() bool }

func TestEnumsAcceptOnlySpecValues(t *testing.T) {
	tests := []struct {
		name  string
		value validator
		want  bool
	}{
		{"section pf_art152", annual.SectionPFArt152, true},
		{"section pf_resico_check", annual.SectionPFRESICOCheck, true},
		{"section pm_601", annual.SectionPM601, true},
		{"section pm_resico", annual.SectionPMRESICO, true},
		{"section explanation", annual.SectionExplanation, true},
		{"section unknown", annual.SectionKind("pf_612"), false},
		{"status complete", annual.StatusComplete, true},
		{"status incomplete", annual.StatusIncomplete, true},
		{"status informational", annual.StatusInformational, true},
		{"status partial is monthly only", annual.Status("partial"), false},
		{"obligation obligated", annual.ObligationObligated, true},
		{"obligation optional", annual.ObligationOptional, true},
		{"obligation not_obligated", annual.ObligationNotObligated, true},
		{"obligation relieved", annual.ObligationRelieved, true},
		{"obligation unknown", annual.Obligation("maybe"), false},
		{"balance a_cargo", annual.BalanceACargo, true},
		{"balance a_favor", annual.BalanceAFavor, true},
		{"balance cero", annual.BalanceCero, true},
		{"balance unknown", annual.BalanceKind("zero"), false},
		{"topic platform", annual.TopicPlatformNonDefinitiveAnnual, true},
		{"topic uma", annual.TopicUMAForMinimumWage, true},
		{"topic resico cap", annual.TopicDeductionCapRESICOIncome, true},
		{"topic salary", annual.TopicSalaryObligationBasis, true},
		{"topic cu", annual.TopicCURounding, true},
		{"topic ptu", annual.TopicRESICOPMPTUBase, true},
		{"topic renewal", annual.TopicRESICOPFReliefRenewal, true},
		{"topic ptu rate", annual.TopicPTURate, true},
		{"topic monthly only", annual.CPATopic("rif_reduction"), false},
		{"disclaimer annual_estimate", annual.DisclaimerAnnualEstimate, true},
		{"disclaimer no_annual_iva", annual.DisclaimerNoAnnualIVA, true},
		{"disclaimer unknown", annual.DisclaimerCode("not_tax_advice"), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.value.Valid())
		})
	}
}

func TestDisclaimerSeverityFollowsTheSpec(t *testing.T) {
	tests := []struct {
		code annual.DisclaimerCode
		want annual.Severity
	}{
		{annual.DisclaimerAnnualEstimate, annual.SeverityInfo},
		{annual.DisclaimerCPAReviewPending, annual.SeverityStrong},
		{annual.DisclaimerYearInProgress, annual.SeverityWarning},
		{annual.DisclaimerPaymentsEstimated, annual.SeverityWarning},
		{annual.DisclaimerIncompleteInputs, annual.SeverityWarning},
		{annual.DisclaimerDeductionsAsEntered, annual.SeverityInfo},
		{annual.DisclaimerRESICOPFDefinitive, annual.SeverityInfo},
		{annual.DisclaimerInformativeIncome, annual.SeverityInfo},
		{annual.DisclaimerRMF2027Pending, annual.SeverityInfo},
		{annual.DisclaimerInflationAdjustmentExcluded, annual.SeverityInfo},
		{annual.DisclaimerNoAnnualIVA, annual.SeverityInfo},
		{annual.DisclaimerCode("unknown"), ""},
	}
	for _, tt := range tests {
		t.Run(string(tt.code), func(t *testing.T) {
			assert.Equal(t, tt.want, tt.code.Severity())
		})
	}
}
