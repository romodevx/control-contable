package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain/taxparams"
)

func TestNewTaxResultRoundsOnce(t *testing.T) {
	tests := []struct {
		name    string
		exact   money.Exact
		amount  money.Cents
		payable money.Cents
		outcome domain.Outcome
	}{
		{name: "payable, half rounds up", exact: 150_955_000, amount: 150_96, payable: 150_96, outcome: domain.OutcomePayable},
		{name: "in favor, half rounds away from zero", exact: -924_005_000, amount: -924_01, payable: 0, outcome: domain.OutcomeInFavor},
		{name: "zero", exact: 0, amount: 0, payable: 0, outcome: domain.OutcomeZero},
		{name: "below half a centavo is zero", exact: 4_999, amount: 0, payable: 0, outcome: domain.OutcomeZero},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := domain.NewTaxResult(tt.exact)
			assert.Equal(t, tt.exact, r.Exact)
			assert.Equal(t, tt.amount, r.Amount)
			assert.Equal(t, tt.payable, r.Payable())
			assert.Equal(t, tt.outcome, r.Outcome())
		})
	}
}

func TestRegimeEstimateStatusPrecedence(t *testing.T) {
	missing := []domain.MissingInput{{Key: domain.InputCoeficienteUtilidad}}
	tests := []struct {
		name     string
		coverage domain.Treatment
		missing  []domain.MissingInput
		want     domain.Status
	}{
		{name: "full, all inputs", coverage: domain.TreatmentFull, want: domain.StatusComplete},
		{name: "full, missing", coverage: domain.TreatmentFull, missing: missing, want: domain.StatusIncomplete},
		{name: "partial", coverage: domain.TreatmentPartial, want: domain.StatusPartial},
		{name: "incomplete beats partial", coverage: domain.TreatmentPartial, missing: missing, want: domain.StatusIncomplete},
		{name: "informational beats incomplete", coverage: domain.TreatmentInformational, missing: missing, want: domain.StatusInformational},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := domain.RegimeEstimate{Coverage: tt.coverage, MissingInputs: tt.missing}
			assert.Equal(t, tt.want, e.Status())
		})
	}
}

func TestRegimeEstimateCPAReviewPending(t *testing.T) {
	assert.False(t, domain.RegimeEstimate{}.CPAReviewPending())
	e := domain.RegimeEstimate{CPATopics: []domain.CPATopic{domain.TopicRIFReduction}}
	assert.True(t, e.CPAReviewPending())
}

func TestStrategyKeyString(t *testing.T) {
	k := domain.StrategyKey{Regime: fiscal.RegimeRESICO, PersonType: fiscal.PersonMoral}
	assert.Equal(t, "626/moral", k.String())
}

func TestEnumsValid(t *testing.T) {
	valid := map[string][]interface{ Valid() bool }{
		"treatment": {domain.TreatmentFull, domain.TreatmentPartial, domain.TreatmentInformational},
		"status":    {domain.StatusComplete, domain.StatusIncomplete, domain.StatusPartial, domain.StatusInformational},
		"tax":       {domain.TaxISR, domain.TaxIVA, domain.TaxWithholding},
		"operation": {
			domain.OpInput, domain.OpAdd, domain.OpSubtract, domain.OpMultiply,
			domain.OpLookup, domain.OpResult, domain.OpRoundingAdjustment, domain.OpInfo,
		},
		"topic": {
			domain.TopicRIFReduction, domain.TopicRIFBimonthlyIVA, domain.TopicPlatformDefinitiveIVA,
			domain.TopicCrossRegimeIVANetting, domain.TopicRESICOPMInterestWithholding, domain.TopicWithheldIVACreditTiming,
		},
		"severity": {domain.SeverityInfo, domain.SeverityWarning, domain.SeverityStrong},
	}
	for name, values := range valid {
		for _, v := range values {
			assert.True(t, v.Valid(), "%s %v", name, v)
		}
	}
	for _, v := range []interface{ Valid() bool }{
		domain.Treatment(""), domain.Status("done"), domain.TaxKind("ieps"), domain.Operation("divide"),
		domain.CPATopic("other"), domain.Severity("loud"), domain.DisclaimerCode("other"),
	} {
		assert.False(t, v.Valid(), "%v", v)
	}
}

func TestDisclaimerSeverity(t *testing.T) {
	want := map[domain.DisclaimerCode]domain.Severity{
		domain.DisclaimerNotTaxAdvice:            domain.SeverityInfo,
		domain.DisclaimerIncompleteInputs:        domain.SeverityWarning,
		domain.DisclaimerProvisionalParameters:   domain.SeverityWarning,
		domain.DisclaimerInformationalRegime:     domain.SeverityInfo,
		domain.DisclaimerPartialRegime:           domain.SeverityInfo,
		domain.DisclaimerDeductibilityClaimed:    domain.SeverityInfo,
		domain.DisclaimerNoCoefficientFirstYear:  domain.SeverityInfo,
		domain.DisclaimerIVAMixedActivity:        domain.SeverityWarning,
		domain.DisclaimerCPAReviewPending:        domain.SeverityStrong,
		domain.DisclaimerRESICOCapExceeded:       domain.SeverityWarning,
		domain.DisclaimerCode("unknown_to_this"): "",
	}
	for code, severity := range want {
		assert.Equal(t, severity, code.Severity(), code)
		assert.Equal(t, severity != "", code.Valid(), code)
	}
}

func TestCitationValid(t *testing.T) {
	assert.True(t, domain.Citation{Law: "LISR", Article: "106", Note: "párrafo quinto"}.Valid())
	assert.True(t, domain.Citation{Law: "LIVA", Article: "5-D"}.Valid(), "note optional")
	assert.False(t, domain.Citation{Law: " ", Article: "106"}.Valid())
	assert.False(t, domain.Citation{Law: "LISR"}.Valid())
}

func TestSourceString(t *testing.T) {
	assert.Equal(t, "records", domain.SourceRecords().String())
	assert.Equal(t, "computed", domain.SourceComputed().String())
	assert.Equal(t, "tax_input:ptu_paid", domain.SourceTaxInput(domain.InputPTUPaid).String())
	assert.Equal(t, "params:isr.art106.accumulated", domain.SourceParams(taxparams.Key("isr.art106.accumulated")).String())
	assert.Empty(t, domain.Source{}.String())
}
