package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/romodevx/control-contable/backend/internal/taxengine/domain"
)

func disclaimerCodes(ds []domain.Disclaimer) []domain.DisclaimerCode {
	out := make([]domain.DisclaimerCode, 0, len(ds))
	for _, d := range ds {
		out = append(out, d.Code)
	}
	return out
}

// not_tax_advice is always present, and first (research Part C).
func TestEstimateDisclaimersAlwaysCarryNotTaxAdvice(t *testing.T) {
	got := domain.EstimateDisclaimers(domain.DisclaimerFacts{})

	assert.Equal(t, []domain.Disclaimer{{Code: domain.DisclaimerNotTaxAdvice, Severity: domain.SeverityInfo}}, got)
}

// Each status raises its code (§2.3); PR #78: incomplete_inputs whenever a
// regime is incomplete, so zero-filled figures always say "se usaron
// ceros".
func TestEstimateDisclaimersFromStatuses(t *testing.T) {
	missing := []domain.MissingInput{{Key: domain.InputPriorProvisionalPayments}}
	tests := []struct {
		name   string
		regime domain.RegimeEstimate
		want   domain.DisclaimerCode
	}{
		{"incomplete", domain.RegimeEstimate{Coverage: domain.TreatmentFull, MissingInputs: missing}, domain.DisclaimerIncompleteInputs},
		{"partial", domain.RegimeEstimate{Coverage: domain.TreatmentPartial}, domain.DisclaimerPartialRegime},
		{"informational", domain.RegimeEstimate{Coverage: domain.TreatmentInformational}, domain.DisclaimerInformationalRegime},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := domain.EstimateDisclaimers(domain.DisclaimerFacts{Regimes: []domain.RegimeEstimate{tt.regime}})

			assert.Equal(t, []domain.DisclaimerCode{domain.DisclaimerNotTaxAdvice, tt.want}, disclaimerCodes(got))
		})
	}
}

// A complete regime raises nothing beyond its own disclaimers.
func TestEstimateDisclaimersCompleteRegime(t *testing.T) {
	r := domain.RegimeEstimate{Coverage: domain.TreatmentFull, Disclaimers: []domain.DisclaimerCode{domain.DisclaimerDeductibilityClaimed}}

	got := domain.EstimateDisclaimers(domain.DisclaimerFacts{Regimes: []domain.RegimeEstimate{r}})

	assert.Equal(t, []domain.DisclaimerCode{domain.DisclaimerNotTaxAdvice, domain.DisclaimerDeductibilityClaimed}, disclaimerCodes(got))
}

// OQ-T8: a rule no CPA validated raises the strong cpa_review_pending,
// naming every topic once (regimes and the consolidated IVA), sorted.
func TestEstimateDisclaimersCPAReviewPendingIsStrongWithTopics(t *testing.T) {
	facts := domain.DisclaimerFacts{
		Regimes: []domain.RegimeEstimate{
			{Coverage: domain.TreatmentPartial, CPATopics: []domain.CPATopic{domain.TopicRIFReduction, domain.TopicRIFBimonthlyIVA}},
			{Coverage: domain.TreatmentFull, CPATopics: []domain.CPATopic{domain.TopicRIFReduction}},
		},
		IVA: domain.ConsolidatedIVA{CPATopics: []domain.CPATopic{domain.TopicCrossRegimeIVANetting}},
	}

	got := domain.EstimateDisclaimers(facts)

	assert.Equal(t, domain.Disclaimer{
		Code: domain.DisclaimerCPAReviewPending, Severity: domain.SeverityStrong,
		Topics: []domain.CPATopic{domain.TopicCrossRegimeIVANetting, domain.TopicRIFBimonthlyIVA, domain.TopicRIFReduction},
	}, got[1], "strong sits right after not_tax_advice")
}

// §2.3 order: not_tax_advice first, then strong, warning, info, then code;
// a code raised twice (a regime and the consolidation) appears once.
func TestEstimateDisclaimersOrderAndDeduplication(t *testing.T) {
	missing := []domain.MissingInput{{Key: domain.InputCoeficienteUtilidad}}
	facts := domain.DisclaimerFacts{
		Regimes: []domain.RegimeEstimate{
			{
				Coverage: domain.TreatmentFull, MissingInputs: missing,
				Disclaimers: []domain.DisclaimerCode{domain.DisclaimerIVAMixedActivity, domain.DisclaimerDeductibilityClaimed},
			},
			{Coverage: domain.TreatmentInformational},
			{Coverage: domain.TreatmentPartial, CPATopics: []domain.CPATopic{domain.TopicRIFReduction}},
		},
		IVA:                   domain.ConsolidatedIVA{Disclaimers: []domain.DisclaimerCode{domain.DisclaimerIVAMixedActivity}},
		ProvisionalParameters: true,
	}

	got := domain.EstimateDisclaimers(facts)

	assert.Equal(t, []domain.DisclaimerCode{
		domain.DisclaimerNotTaxAdvice,
		domain.DisclaimerCPAReviewPending,
		domain.DisclaimerIncompleteInputs,
		domain.DisclaimerIVAMixedActivity,
		domain.DisclaimerProvisionalParameters,
		domain.DisclaimerDeductibilityClaimed,
		domain.DisclaimerInformationalRegime,
		domain.DisclaimerPartialRegime,
	}, disclaimerCodes(got))
	for _, d := range got {
		assert.Equal(t, d.Code.Severity(), d.Severity, d.Code)
		if d.Code != domain.DisclaimerCPAReviewPending {
			assert.Empty(t, d.Topics, d.Code)
		}
	}
}
