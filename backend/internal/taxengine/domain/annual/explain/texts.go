package explain

import (
	"strconv"
	"strings"
	"time"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain/annual"
)

// The es-MX v1 catalog of the annual estimate (MODULE-annual §2.8, §2.10):
// the legend, the disclaimer texts, the CPA-topic phrases and the
// explanation texts. Texts never carry an amount; a code the catalog does
// not know renders as the generic notice, never as its key.

// Legend is the frozen legend of every annual response (owner decision 1).
const Legend = "Esto es un aproximado; tu contador te dará las cifras reales."

// NoDate is the zero date: an explanation without a deadline.
var NoDate = fiscal.Date{}

// TextFacts are the variable parts of the disclaimer texts: the fiscal
// year, the last closed month of a year in progress and how many payments
// the app estimated.
type TextFacts struct {
	FiscalYear      int
	DataThrough     time.Month
	EstimatedMonths int
}

// fixedTexts are the disclaimers without variable parts.
var fixedTexts = map[annual.DisclaimerCode]string{
	annual.DisclaimerIncompleteInputs:            "Faltan datos; usamos ceros o dejamos fuera lo que no pudimos verificar.",
	annual.DisclaimerDeductionsAsEntered:         "Tomamos tus deducciones personales como las registraste; el SAT las valida con tus CFDI.",
	annual.DisclaimerInformativeIncome:           "Además debes informar en tu anual préstamos, donativos, herencias y premios que recibiste (LISR 150).",
	annual.DisclaimerInflationAdjustmentExcluded: "No incluimos el ajuste anual por inflación ni actualizamos pérdidas: pídeselos a tu contador.",
	annual.DisclaimerNoAnnualIVA:                 "El IVA no tiene declaración anual: se paga mes con mes.",
}

// DisclaimerText renders d with f.
func DisclaimerText(d Disclaimer, f TextFacts) string {
	fy := strconv.Itoa(f.FiscalYear)
	switch d.Code {
	case annual.DisclaimerAnnualEstimate:
		return annualEstimateText(fy)
	case annual.DisclaimerCPAReviewPending:
		return "Este cálculo usa reglas que un contador aún no valida: " + topicList(d.Topics) + "."
	case annual.DisclaimerYearInProgress:
		return "El año aún no termina: calculamos con tus registros hasta " + monthNames[f.DataThrough] + " de " + fy + "."
	case annual.DisclaimerPaymentsEstimated:
		return "Usamos lo que la app estimó para " + months(f.EstimatedMonths) +
			"; confírmalo con tus acuses para un resultado más preciso."
	case annual.DisclaimerRESICOPFDefinitive:
		return "Tus pagos mensuales de RESICO son definitivos; no presentas anual por estos ingresos (regla 3.13.7 RMF " + fy + ")."
	case annual.DisclaimerRMF2027Pending:
		return "Las facilidades para la anual " + fy + " (pago en parcialidades, exenciones, RESICO) se publican en la RMF " +
			strconv.Itoa(f.FiscalYear+1) + "; pueden cambiar."
	}
	if text, ok := fixedTexts[d.Code]; ok {
		return text
	}
	return annualEstimateText(fy)
}

func annualEstimateText(fy string) string {
	return "Estimación informativa de la declaración anual " + fy + ". No sustituye la declaración que presenta tu " +
		"contador ante el SAT. Basada en LISR arts. 9, 109, 113-F, 150–152, Anexo 8 RMF " + fy + " y RMF " + fy + "."
}

func months(n int) string {
	if n == 1 {
		return "1 mes"
	}
	return strconv.Itoa(n) + " meses"
}

// topicPhrases name each CPA topic inside the cpa_review_pending text.
var topicPhrases = map[annual.CPATopic]string{
	annual.TopicPlatformNonDefinitiveAnnual: "los ingresos de plataformas sin pago definitivo en la anual",
	annual.TopicUMAForMinimumWage:           "el uso de la UMA en lugar del salario mínimo en los topes",
	annual.TopicDeductionCapRESICOIncome:    "dejar fuera los ingresos de RESICO del tope de deducciones",
	annual.TopicSalaryObligationBasis:       "el límite de sueldos que obliga a presentar la anual",
	annual.TopicCURounding:                  "el redondeo del coeficiente de utilidad sugerido",
	annual.TopicRESICOPMPTUBase:             "la base de la PTU en RESICO de personas morales",
	annual.TopicRESICOPFReliefRenewal:       "la renovación de la facilidad de RESICO para no presentar anual",
	annual.TopicPTURate:                     "el porcentaje de PTU a repartir",
}

// genericTopic stands for a topic the catalog does not know yet.
const genericTopic = "una regla pendiente de revisión"

// TopicPhrase names t inside the cpa_review_pending text.
func TopicPhrase(t annual.CPATopic) string {
	if p, ok := topicPhrases[t]; ok {
		return p
	}
	return genericTopic
}

// topicList joins the topics' phrases the Spanish way: "a, b y c".
func topicList(topics []annual.CPATopic) string {
	phrases := make([]string, len(topics))
	for i, t := range topics {
		phrases[i] = TopicPhrase(t)
	}
	if len(phrases) <= 1 {
		return strings.Join(phrases, "")
	}
	return strings.Join(phrases[:len(phrases)-1], ", ") + " y " + phrases[len(phrases)-1]
}

// explanationTexts are the explanations without a deadline in the text.
var explanationTexts = map[Code]string{
	CodePlatformDefinitive: "Elegiste el pago definitivo en plataformas: tus retenciones de plataforma fueron definitivas " +
		"y estos ingresos no van en tu declaración anual.",
	CodePrizesDefinitive: "La retención sobre premios es pago definitivo: estos ingresos no van en tu declaración anual.",
	CodeNoObligations:    "Sin obligaciones fiscales: no presentas declaración anual por este régimen.",
	CodeForeignResident: "Como residente en el extranjero la retención es definitiva: no presentas declaración anual " +
		"por estos ingresos.",
	CodeAskCPA: "No estimamos la anual de este régimen; pregúntale a tu contador si debes presentarla.",
}

// ExplanationText renders an explanation; deadline is the date it shows,
// NoDate when it has none.
func ExplanationText(code Code, deadline fiscal.Date) string {
	switch {
	case code == CodeNonprofitInformative && !deadline.IsZero():
		return "Presentas una declaración informativa a más tardar el " + longDate(deadline) + "; tu contador la prepara."
	case code == CodeNonprofitInformative:
		return "Presentas una declaración informativa anual; tu contador la prepara y te indica el plazo."
	case code == CodeCPAComputes && !deadline.IsZero():
		return "Por estos ingresos presentas declaración anual a más tardar el " + longDate(deadline) +
			"; tu contador calcula el impuesto."
	case code == CodeCPAComputes:
		return "Tu contador calcula el impuesto anual de estos ingresos y te indica el plazo."
	case code == CodeCPAConfirms && !deadline.IsZero():
		return "Según tus demás ingresos, quizá debas declarar estos ingresos en la anual a más tardar el " +
			longDate(deadline) + "; tu contador lo confirma y calcula el impuesto."
	case code == CodeCPAConfirms:
		return "Según tus demás ingresos, quizá debas declarar estos ingresos en la anual; tu contador lo confirma."
	}
	if text, ok := explanationTexts[code]; ok {
		return text
	}
	return explanationTexts[CodeAskCPA]
}

// monthNames are the Spanish month names, lower case as written in a date.
var monthNames = map[time.Month]string{
	time.January: "enero", time.February: "febrero", time.March: "marzo", time.April: "abril",
	time.May: "mayo", time.June: "junio", time.July: "julio", time.August: "agosto",
	time.September: "septiembre", time.October: "octubre", time.November: "noviembre", time.December: "diciembre",
}

// longDate renders d as "15 de febrero de 2027".
func longDate(d fiscal.Date) string {
	t, err := time.Parse(time.DateOnly, d.String())
	if err != nil { // a valid fiscal.Date always parses
		return d.String()
	}
	return strconv.Itoa(t.Day()) + " de " + monthNames[t.Month()] + " de " + strconv.Itoa(t.Year())
}
