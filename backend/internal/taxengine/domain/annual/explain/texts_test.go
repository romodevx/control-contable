package explain_test

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/romodevx/control-contable/backend/internal/taxengine/domain/annual"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain/annual/explain"
)

// allDisclaimers and allTopics are the spec's codes (MODULE-annual §2.8,
// §2.10); each must be Valid, so a code added to the enums without a text
// here fails TestEveryCodeHasAText below once listed.
var (
	allDisclaimers = []annual.DisclaimerCode{
		annual.DisclaimerAnnualEstimate, annual.DisclaimerCPAReviewPending, annual.DisclaimerYearInProgress,
		annual.DisclaimerPaymentsEstimated, annual.DisclaimerIncompleteInputs, annual.DisclaimerDeductionsAsEntered,
		annual.DisclaimerRESICOPFDefinitive, annual.DisclaimerInformativeIncome, annual.DisclaimerRMF2027Pending,
		annual.DisclaimerInflationAdjustmentExcluded, annual.DisclaimerNoAnnualIVA,
	}
	allTopics = []annual.CPATopic{
		annual.TopicPlatformNonDefinitiveAnnual, annual.TopicUMAForMinimumWage, annual.TopicDeductionCapRESICOIncome,
		annual.TopicSalaryObligationBasis, annual.TopicCURounding, annual.TopicRESICOPMPTUBase,
		annual.TopicRESICOPFReliefRenewal, annual.TopicPTURate,
	}
	allCodes = []explain.Code{
		explain.CodePlatformDefinitive, explain.CodePrizesDefinitive, explain.CodeNoObligations,
		explain.CodeForeignResident, explain.CodeNonprofitInformative, explain.CodeCPAComputes,
		explain.CodeCPAConfirms, explain.CodeAskCPA,
	}
)

func facts() explain.TextFacts {
	return explain.TextFacts{FiscalYear: fiscalYear, DataThrough: time.September, EstimatedMonths: 3}
}

// TestEveryCodeHasAText: the catalog renders every disclaimer, topic and
// explanation; none falls back to a key, a placeholder or an amount.
func TestEveryCodeHasAText(t *testing.T) {
	for _, code := range allDisclaimers {
		assert.True(t, code.Valid(), code)
		text := explain.DisclaimerText(explain.Disclaimer{Code: code, Topics: allTopics}, facts())
		assertRendered(t, string(code), text)
	}
	for _, topic := range allTopics {
		assert.True(t, topic.Valid(), topic)
		assertRendered(t, string(topic), explain.TopicPhrase(topic))
	}
	for _, code := range allCodes {
		assertRendered(t, string(code), explain.ExplanationText(code, date(t, "2027-04-30")))
		assertRendered(t, string(code), explain.ExplanationText(code, explain.NoDate))
	}
}

func assertRendered(t *testing.T, key, text string) {
	t.Helper()
	assert.NotEmpty(t, text, key)
	assert.NotContains(t, text, "{", key)
	assert.NotContains(t, text, key, "the key itself")
	assert.NotContains(t, text, "$", "%s: texts never carry an amount", key)
}

// TestDisclaimerTextsAreTheSpec pins the es-MX v1 texts of MODULE-annual
// §2.8 with their variable parts.
func TestDisclaimerTextsAreTheSpec(t *testing.T) {
	tests := []struct {
		d    explain.Disclaimer
		want string
	}{
		{explain.Disclaimer{Code: annual.DisclaimerAnnualEstimate},
			"Estimación informativa de la declaración anual 2026. No sustituye la declaración que presenta tu contador " +
				"ante el SAT. Basada en LISR arts. 9, 109, 113-F, 150–152, Anexo 8 RMF 2026 y RMF 2026."},
		{explain.Disclaimer{Code: annual.DisclaimerCPAReviewPending, Topics: []annual.CPATopic{annual.TopicCURounding, annual.TopicPTURate}},
			"Este cálculo usa reglas que un contador aún no valida: el redondeo del coeficiente de utilidad sugerido y " +
				"el porcentaje de PTU a repartir."},
		{explain.Disclaimer{Code: annual.DisclaimerYearInProgress}, "El año aún no termina: calculamos con tus registros hasta septiembre de 2026."},
		{explain.Disclaimer{Code: annual.DisclaimerPaymentsEstimated},
			"Usamos lo que la app estimó para 3 meses; confírmalo con tus acuses para un resultado más preciso."},
		{explain.Disclaimer{Code: annual.DisclaimerRMF2027Pending},
			"Las facilidades para la anual 2026 (pago en parcialidades, exenciones, RESICO) se publican en la RMF 2027; pueden cambiar."},
		{explain.Disclaimer{Code: annual.DisclaimerRESICOPFDefinitive},
			"Tus pagos mensuales de RESICO son definitivos; no presentas anual por estos ingresos (regla 3.13.7 RMF 2026)."},
		{explain.Disclaimer{Code: annual.DisclaimerNoAnnualIVA}, "El IVA no tiene declaración anual: se paga mes con mes."},
	}
	for _, tt := range tests {
		t.Run(string(tt.d.Code), func(t *testing.T) {
			assert.Equal(t, tt.want, explain.DisclaimerText(tt.d, facts()))
		})
	}
}

func TestPaymentsEstimatedSingularMonth(t *testing.T) {
	f := facts()
	f.EstimatedMonths = 1

	got := explain.DisclaimerText(explain.Disclaimer{Code: annual.DisclaimerPaymentsEstimated}, f)

	assert.Equal(t, "Usamos lo que la app estimó para 1 mes; confírmalo con tus acuses para un resultado más preciso.", got)
}

func TestTopicListJoinsTheSpanishWay(t *testing.T) {
	one := explain.DisclaimerText(explain.Disclaimer{Code: annual.DisclaimerCPAReviewPending, Topics: []annual.CPATopic{annual.TopicPTURate}}, facts())
	three := explain.DisclaimerText(explain.Disclaimer{Code: annual.DisclaimerCPAReviewPending, Topics: []annual.CPATopic{
		annual.TopicCURounding, annual.TopicPTURate, annual.TopicRESICOPMPTUBase,
	}}, facts())

	assert.True(t, strings.HasSuffix(one, ": el porcentaje de PTU a repartir."), one)
	assert.Contains(t, three, "sugerido, el porcentaje de PTU a repartir y la base")
}

// TestUnknownCodesFallBackToTheGenericNotice: a new code the catalog does
// not know yet never shows its key.
func TestUnknownCodesFallBackToTheGenericNotice(t *testing.T) {
	generic := explain.DisclaimerText(explain.Disclaimer{Code: annual.DisclaimerAnnualEstimate}, facts())

	assert.Equal(t, generic, explain.DisclaimerText(explain.Disclaimer{Code: "made_up"}, facts()))
	assert.NotContains(t, explain.ExplanationText("explanation.made_up", explain.NoDate), "made_up")
	assert.NotContains(t, explain.TopicPhrase("made_up"), "made_up")
}

// TestExplanationTexts pins the spec's 603 sentence (US-A11) and the 625
// definitive note (§2.2).
func TestExplanationTexts(t *testing.T) {
	assert.Equal(t, "Presentas una declaración informativa a más tardar el 15 de febrero de 2027; tu contador la prepara.",
		explain.ExplanationText(explain.CodeNonprofitInformative, date(t, "2027-02-15")))
	assert.Contains(t, explain.ExplanationText(explain.CodePlatformDefinitive, explain.NoDate),
		"tus retenciones de plataforma fueron definitivas")
	assert.Equal(t, "Por estos ingresos presentas declaración anual a más tardar el 31 de marzo de 2027; tu contador calcula el impuesto.",
		explain.ExplanationText(explain.CodeCPAComputes, date(t, "2027-03-31")))
	assert.Equal(t, "Tu contador calcula el impuesto anual de estos ingresos y te indica el plazo.",
		explain.ExplanationText(explain.CodeCPAComputes, explain.NoDate))
}

func TestLegendIsFrozen(t *testing.T) {
	assert.Equal(t, "Esto es un aproximado; tu contador te dará las cifras reales.", explain.Legend)
}
