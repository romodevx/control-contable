package taxenginehttp

import (
	"strconv"
	"strings"
	"time"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain"
)

// The es-MX texts of the estimate and breakdown responses (MODULE-tax-engine
// §2.3, §2.5). The engine returns keys (label keys, disclaimer codes, CPA
// topics); this file is their one rendering. Texts never carry an amount.

// disclaimerTexts are the v1 texts of each disclaimer code (§2.3). The ones
// with variable parts are rendered by disclaimerText.
var disclaimerTexts = map[domain.DisclaimerCode]string{
	domain.DisclaimerNotTaxAdvice: "Estimación informativa. No sustituye la declaración ante el SAT ni la asesoría de un contador. " +
		"Cálculo basado en LISR, LIVA, LIF 2026 y Anexo 8 RMF 2026.",
	domain.DisclaimerIncompleteInputs:       "Faltan datos para este cálculo; se usaron ceros. Complétalos para un resultado más preciso.",
	domain.DisclaimerInformationalRegime:    "Algunos regímenes son informativos: la app no estima su impuesto.",
	domain.DisclaimerPartialRegime:          "En algunos regímenes sólo estimamos parte de los impuestos.",
	domain.DisclaimerDeductibilityClaimed:   "Tomamos como deducibles los gastos que marcaste así; el SAT puede no aceptarlos todos.",
	domain.DisclaimerNoCoefficientFirstYear: "En el primer ejercicio no hay pagos provisionales de ISR (LISR art. 14).",
	domain.DisclaimerIVAMixedActivity: "Tienes actividades exentas: el IVA acreditable debe ser proporcional " +
		"y aquí se acreditó completo.",
}

// resicoCapTexts are resico_cap_exceeded by person type (LISR 113-E, 113-I;
// LISR 206, 214).
var resicoCapTexts = map[fiscal.PersonType]string{
	fiscal.PersonFisica: "Tus ingresos cobrados en el año ya pasan de $3,500,000, el límite de RESICO. Desde el mes siguiente " +
		"dejas RESICO y pagas en el régimen general de tu actividad (612 o 606); revisa con tu contador.",
	fiscal.PersonMoral: "Los ingresos cobrados de la empresa en el año ya pasan de $35,000,000, el límite de RESICO. Desde el " +
		"siguiente ejercicio tributa en el régimen general (601, Título II) y debe avisarlo al SAT a más tardar el 31 de enero.",
}

// topicPhrases name each CPA topic inside the cpa_review_pending text.
var topicPhrases = map[domain.CPATopic]string{
	domain.TopicRIFReduction:                "la reducción del RIF",
	domain.TopicRIFBimonthlyIVA:             "el IVA bimestral del RIF",
	domain.TopicPlatformDefinitiveIVA:       "el IVA definitivo de plataformas",
	domain.TopicCrossRegimeIVANetting:       "compensación del IVA entre regímenes",
	domain.TopicRESICOPMInterestWithholding: "la retención por intereses en RESICO de personas morales",
	domain.TopicWithheldIVACreditTiming:     "acreditar el mes siguiente el IVA que retuviste a tus proveedores",
}

// disclaimerFacts are what the variable disclaimer texts read.
type disclaimerFacts struct {
	personType fiscal.PersonType
	// periodYear is the period's fiscal year; paramsYear the parameter
	// set's (an earlier one when provisional).
	periodYear, paramsYear int
}

// disclaimerText renders d. A code without a text (a new code the adapter
// does not know yet) falls back to the generic notice, never to the code.
func disclaimerText(d domain.Disclaimer, f disclaimerFacts) string {
	switch d.Code {
	case domain.DisclaimerCPAReviewPending:
		return "Este cálculo usa reglas que un contador aún no valida: " + topicList(d.Topics) + ". Tómalo con mayor reserva."
	case domain.DisclaimerProvisionalParameters:
		return "Aún no se publican las tarifas de " + strconv.Itoa(f.periodYear) + "; usamos las de " + strconv.Itoa(f.paramsYear) + "."
	case domain.DisclaimerRESICOCapExceeded:
		if text, ok := resicoCapTexts[f.personType]; ok {
			return text
		}
	}
	if text, ok := disclaimerTexts[d.Code]; ok {
		return text
	}
	return disclaimerTexts[domain.DisclaimerNotTaxAdvice]
}

// topicList joins the topics' phrases the Spanish way: "a, b y c".
func topicList(topics []domain.CPATopic) string {
	phrases := make([]string, 0, len(topics))
	for _, t := range topics {
		if p, ok := topicPhrases[t]; ok {
			phrases = append(phrases, p)
		}
	}
	switch len(phrases) {
	case 0:
		return "reglas pendientes de revisión"
	case 1:
		return phrases[0]
	}
	return strings.Join(phrases[:len(phrases)-1], ", ") + " y " + phrases[len(phrases)-1]
}

// strategyExplanations are the explanations a strategy names by its own
// key rather than the catalog's (625, 626 PF).
var strategyExplanations = map[domain.LabelKey]string{
	"platform_provisional_withholding": "Sin el pago definitivo, el ISR se calcula en 612 o RESICO; calcula tu pago ahí y " +
		"acredita estas retenciones.",
	"platform_mixed_activities": "Tienes varias actividades en plataformas con tasas distintas y tus ingresos del mes no se " +
		"separan por actividad, así que no estimamos el ISR; acredita estas retenciones en tu cálculo.",
	domain.LabelRESICOPFAboveTable: "Tus ingresos del mes pasan del último renglón de la tabla mensual de RESICO, así que no " +
		"hay tasa que aplicar; revisa con tu contador.",
}

// msgNoExplanation stands in for a catalog row without its text.
const msgNoExplanation = "La app no estima el impuesto de este régimen."

// explanation renders a regime's explanation key: the catalog's text for
// the regime's own key, else the strategy's; nil when there is none.
func explanation(key domain.StrategyKey, label domain.LabelKey, catalog string) *string {
	if label == "" {
		return nil
	}
	text, ok := strategyExplanations[label]
	switch {
	case label == domain.RegimeExplanation(key) && catalog != "":
		text = catalog
	case !ok:
		text = msgNoExplanation
	}
	return &text
}

// conceptContext is what a line's concept may name besides its own args.
type conceptContext struct {
	regime fiscal.RegimeCode
	period domain.Period
	line   domain.Line
}

type conceptFunc func(c conceptContext) string

// fixed is a concept with no variable part.
func fixed(text string) conceptFunc { return func(conceptContext) string { return text } }

// ranged names the args' month range ("ene–mar") after prefix, or falls
// back when the line has none.
func ranged(prefix, fallback string) conceptFunc {
	return func(c conceptContext) string {
		if r, ok := monthRange(c.line.Args); ok {
			return prefix + " " + r
		}
		return fallback
	}
}

// tariffRow names the tariff and the row a lookup line read.
func tariffRow(table string) conceptFunc {
	return func(c conceptContext) string {
		name := table
		if len(c.line.Args) > 0 {
			if month, ok := monthName(c.line.Args[0]); ok {
				name = "tarifa acumulada de " + month
			}
		}
		return "Cuota fija (" + name + rowSuffix(c.line.Row) + ")"
	}
}

func rowSuffix(row int) string {
	if row <= 0 {
		return ""
	}
	return ", renglón " + strconv.Itoa(row)
}

// concepts render each label key (§2.5). conceptOverrides take precedence
// for one regime, where the same step reads differently.
var concepts = map[domain.LabelKey]conceptFunc{
	domain.LabelRoundingAdjustment: fixed("Ajuste por redondeo"),

	// IVA of one regime (ComputeIVA) and the taxpayer's (ConsolidateIVA).
	"iva.charged.rate_16":     fixed("IVA trasladado cobrado al 16 %"),
	"iva.charged.rate_8":      fixed("IVA trasladado cobrado al 8 % (región fronteriza)"),
	"iva.base.rate_0":         fixed("Cobros a la tasa del 0 %"),
	"iva.base.exempt":         fixed("Cobros exentos de IVA"),
	"iva.charged":             fixed("IVA trasladado cobrado"),
	"iva.definitive_excluded": fixed("IVA que la plataforma retuvo como pago definitivo"),
	"iva.creditable":          fixed("IVA acreditable pagado"),
	"iva.creditable.withheld_prior_month": fixed(
		"IVA retenido el mes pasado (acreditable este mes; supone que enteraste la retención a tiempo)"),
	"iva.withheld.moral":            fixed("IVA que te retuvieron personas morales"),
	"iva.withheld.platform":         fixed("IVA que te retuvieron las plataformas"),
	"iva.regime_net":                fixed("IVA de este régimen"),
	"iva.creditable_not_applied":    fixed("IVA acreditable que no se aplica (pago definitivo)"),
	"iva.rif_bimester":              ranged("IVA del bimestre", "IVA del bimestre"),
	"iva.total.charged":             fixed("IVA trasladado cobrado (todos tus regímenes)"),
	"iva.total.definitive_excluded": fixed("IVA retenido como pago definitivo (todos tus regímenes)"),
	"iva.total.creditable":          fixed("IVA acreditable (todos tus regímenes)"),
	"iva.total.withheld":            fixed("IVA que te retuvieron (todos tus regímenes)"),
	"iva.consolidated":              fixed("IVA a cargo del mes (todos tus regímenes)"),
	"iva.previous_in_favor":         fixed("Saldo a favor de IVA del mes anterior (no se aplica en esta estimación)"),

	// Withholdings to remit.
	"withholding.iva_to_remit": fixed("IVA que retuviste a tus proveedores (a enterar)"),
	"withholding.isr_to_remit": fixed("ISR que retuviste a tus proveedores (a enterar)"),
	"withholding.to_remit":     fixed("Retenciones a enterar"),

	// ISR steps shared by several regimes.
	"isr.income_ytd":            ranged("Ingresos cobrados acumulados", "Ingresos acumulados del ejercicio"),
	"isr.income_before_app":     fixed("Ingresos del ejercicio previos a la app"),
	"isr.deductions_ytd":        ranged("Deducciones pagadas acumuladas", "Deducciones acumuladas del ejercicio"),
	"isr.deductions_before_app": fixed("Deducciones del ejercicio previas a la app"),
	"isr.ptu_paid":              fixed("PTU pagada en el ejercicio"),
	"isr.ptu":                   fixed("PTU pagada en el ejercicio"),
	"isr.prior_year_losses":     fixed("Pérdidas fiscales de ejercicios anteriores"),
	"isr.base":                  fixed("Base gravable"),
	"isr.no_base":               fixed("Sin base gravable"),
	"isr.no_profit":             fixed("Sin utilidad fiscal"),
	"isr.tariff_row":            tariffRow("tarifa"),
	"isr.tariff_excess":         fixed("Excedente del límite inferior por la tasa"),
	"isr.caused":                fixed("ISR causado"),
	"isr.prior_payments":        priorPayments,
	"isr.withheld_by_moral":     fixed("ISR que te retuvieron personas morales"),
	"isr.payable":               fixed("ISR a pagar"),
	"isr.rate":                  fixed("ISR a la tasa del impuesto"),
	"isr.interest_withholding":  fixed("ISR que te retuvieron por intereses"),

	// 606 PF.
	"isr.rent_month":        rentMonth,
	"isr.blind_deduction":   blindDeduction,
	"isr.predial":           fixed("Impuesto predial pagado"),
	"isr.actual_deductions": fixed("Deducciones pagadas del mes"),

	// 626 PF.
	"isr.income_month": fixed("Ingresos cobrados del mes"),
	"isr.resico_rate":  resicoRate,
	"isr.resico_excess_to_annual": fixed(
		"Retención mayor al ISR del mes: este mes pagas $0 y el resto queda a favor en tu declaración anual"),

	// 601 PM.
	"isr.coeficiente":       fixed("Utilidad estimada con el coeficiente de utilidad"),
	"isr.first_fiscal_year": fixed("Primer ejercicio: sin pagos provisionales de ISR"),

	// 625 PF.
	"isr.platform_income":                  fixed("Ingresos cobrados por plataformas"),
	"isr.platform_rate":                    fixed("Tasa de retención de la plataforma"),
	"isr.withheld_by_platform":             fixed("ISR que te retuvieron las plataformas"),
	"isr.definitive_no_favor":              fixed("La retención es pago definitivo: no genera saldo a favor"),
	"isr.withheld_by_platform_provisional": fixed("ISR retenido por plataformas (pago provisional, acredítalo)"),
	"isr.withheld_by_platform_definitive":  fixed("ISR retenido por plataformas (pago definitivo)"),

	// 621 PF.
	"isr.rif_bimester_pending": rifPending,
	"isr.income_bimester":      ranged("Ingresos cobrados del bimestre", "Ingresos cobrados del bimestre"),
	"isr.deductions_bimester":  ranged("Deducciones pagadas del bimestre", "Deducciones pagadas del bimestre"),
	"isr.rif_reduction":        rifReduction,
}

// conceptOverrides are a regime's own wording of a shared step, keyed by
// "<regime>/<label>".
var conceptOverrides = map[string]conceptFunc{
	"612/isr.caused":            fixed("ISR causado acumulado"),
	"612/isr.withheld_by_moral": fixed("Retenciones de ISR (10 %) acumuladas"),
	"606/isr.withheld_by_moral": fixed("Retención de ISR (10 %) de personas morales"),
	"606/isr.tariff_row":        tariffRow("tarifa mensual"),
	"621/isr.tariff_row":        tariffRow("tarifa bimestral"),
	"626/isr.withheld_by_moral": fixed("Retención de ISR (1.25 %) de personas morales"),
}

// msgUnknownConcept is the concept of a label no entry renders: the
// coverage test keeps it unreachable for every shipped strategy.
const msgUnknownConcept = "Paso del cálculo"

// Concept renders the Spanish concept of line, a step of regime's
// computation for period (contract BreakdownLine.concept).
func Concept(regime fiscal.RegimeCode, period domain.Period, line domain.Line) string {
	c := conceptContext{regime: regime, period: period, line: line}
	if f, ok := conceptOverrides[string(regime)+"/"+string(line.Label)]; ok {
		return f(c)
	}
	if f, ok := concepts[line.Label]; ok {
		return f(c)
	}
	return msgUnknownConcept
}

// HasConcept reports a label the catalog renders (the coverage test).
func HasConcept(label domain.LabelKey) bool {
	_, ok := concepts[label]
	return ok
}

func priorPayments(c conceptContext) string {
	const text = "Pagos provisionales anteriores"
	if c.period.IsZero() || c.period.Month() == time.January {
		return text
	}
	return text + " (" + monthAbbr(time.January) + "–" + monthAbbr(c.period.Month()-1) + ")"
}

func rentMonth(c conceptContext) string {
	if len(c.line.Args) > 0 {
		if month, ok := monthName(c.line.Args[0]); ok {
			return "Rentas cobradas en " + month
		}
	}
	return "Rentas cobradas del mes"
}

func blindDeduction(c conceptContext) string {
	if len(c.line.Args) > 0 {
		if pct, ok := percent(c.line.Args[0]); ok {
			return "Deducción ciega (" + pct + " de las rentas)"
		}
	}
	return "Deducción ciega"
}

func resicoRate(c conceptContext) string {
	if len(c.line.Args) > 0 {
		return "Tasa de la tabla de RESICO (renglón " + c.line.Args[0] + ")"
	}
	return "Tasa de la tabla de RESICO"
}

func rifPending(c conceptContext) string {
	if len(c.line.Args) > 0 {
		if month, ok := monthName(c.line.Args[0]); ok {
			return "El RIF se paga por bimestre; consulta " + month
		}
	}
	return "El RIF se paga por bimestre"
}

// rifReduction names the year in RIF and its reduction ([year, rate]).
func rifReduction(c conceptContext) string {
	const text = "Reducción del RIF"
	if len(c.line.Args) < 2 {
		return text
	}
	if pct, ok := percent(c.line.Args[1]); ok {
		return text + " (año " + c.line.Args[0] + ", " + pct + ")"
	}
	return text + " (año " + c.line.Args[0] + ")"
}

// monthNames and monthAbbrs are the Spanish month names, January first.
var (
	monthNames = [...]string{"enero", "febrero", "marzo", "abril", "mayo", "junio", "julio", "agosto",
		"septiembre", "octubre", "noviembre", "diciembre"}
	monthAbbrs = [...]string{"ene", "feb", "mar", "abr", "may", "jun", "jul", "ago", "sep", "oct", "nov", "dic"}
)

func monthAbbr(m time.Month) string { return monthAbbrs[m-1] }

// monthName names the month of a "YYYY-MM" arg.
func monthName(arg string) (string, bool) {
	p, err := domain.ParsePeriod(arg)
	if err != nil {
		return "", false
	}
	return monthNames[p.Month()-1], true
}

// monthRange renders the first two "YYYY-MM" args as "ene–mar" (one month:
// "mar").
func monthRange(args []string) (string, bool) {
	const rangeArgs = 2
	if len(args) < rangeArgs {
		return "", false
	}
	from, errFrom := domain.ParsePeriod(args[0])
	to, errTo := domain.ParsePeriod(args[1])
	if errFrom != nil || errTo != nil {
		return "", false
	}
	if from.Month() == to.Month() {
		return monthAbbr(to.Month()), true
	}
	return monthAbbr(from.Month()) + "–" + monthAbbr(to.Month()), true
}

// percent renders a decimal fraction ("0.35", "0.0125") as a percentage
// ("35 %", "1.25 %") by moving the point two places, never through a float.
func percent(decimal string) (string, bool) {
	const places = 2
	whole, frac, _ := strings.Cut(decimal, ".")
	if whole == "" || !allDigits(whole) || !allDigits(frac) {
		return "", false
	}
	frac += strings.Repeat("0", max(places-len(frac), 0))
	whole = strings.TrimLeft(whole+frac[:places], "0")
	if whole == "" {
		whole = "0"
	}
	if rest := strings.TrimRight(frac[places:], "0"); rest != "" {
		whole += "." + rest
	}
	return whole + " %", true
}

func allDigits(s string) bool {
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}
