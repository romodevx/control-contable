package annualhttp

import (
	"strconv"
	"strings"

	"github.com/romodevx/control-contable/backend/internal/taxengine/domain"
	adomain "github.com/romodevx/control-contable/backend/internal/taxengine/domain/annual"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain/annual/pm"
	usecaseannual "github.com/romodevx/control-contable/backend/internal/taxengine/usecase/annual"
)

// The es-MX texts of the annual responses (MODULE-annual §2.5, §2.6, §8):
// breakdown concepts, obligation reasons, notes, missing-input labels and
// write warnings. The engine returns keys; this file is their one
// rendering. A key without a text renders a generic phrase, never itself.

// conceptFunc renders the concept of one line.
type conceptFunc func(l domain.Line, fiscalYear int) string

func fixed(text string) conceptFunc { return func(domain.Line, int) string { return text } }

// concepts are the annual lines' concepts by label key.
var concepts = map[domain.LabelKey]conceptFunc{
	// pf_art152 (ANN-08)
	"annual_salary":                  salaryConcept,
	"annual_business_income":         businessIncome,
	"annual_business_deductions":     fixed("Deducciones autorizadas pagadas en el año"),
	"annual_depreciation":            fixed("Depreciación del ejercicio"),
	"annual_ptu_paid":                fixed("PTU pagada en el ejercicio"),
	"annual_prior_year_losses":       fixed("Pérdidas fiscales de ejercicios anteriores"),
	"annual_business_result":         fixed("Utilidad del ejercicio (actividad empresarial y profesional)"),
	"annual_business_loss":           fixed("Pérdida del ejercicio, amortizable en años siguientes"),
	"annual_business_taxable":        fixed("Ingresos acumulables de actividad empresarial y profesional"),
	"annual_rent_income":             fixed("Rentas cobradas en el año"),
	"annual_rent_blind_deduction":    blindDeduction,
	"annual_rent_predial":            fixed("Impuesto predial pagado"),
	"annual_rent_deductions":         fixed("Deducciones autorizadas de arrendamiento"),
	"annual_rent_result":             fixed("Resultado de arrendamiento"),
	"annual_rent_loss":               fixed("Pérdida de arrendamiento del ejercicio, amortizable en años siguientes"),
	"annual_rent_taxable":            fixed("Ingresos acumulables de arrendamiento"),
	"annual_accumulable_income":      fixed("Ingresos acumulables"),
	"annual_personal_deductions":     fixed("Deducciones personales"),
	"annual_base":                    fixed("Base del impuesto anual"),
	"annual_no_base":                 fixed("Sin base gravable: no hay impuesto"),
	"annual_tariff_row":              tariffRow,
	"annual_tariff_excess":           fixed("Excedente del límite inferior por la tasa"),
	"annual_isr":                     fixed("ISR del ejercicio"),
	"annual_credit_payments":         paymentsConcept,
	"annual_credit_withheld":         fixed("ISR que te retuvieron personas morales o plataformas"),
	"annual_credit_salary_isr":       fixed("ISR que te retuvieron tus patrones"), //nolint:misspell // es-MX "patrones" (employers)
	"annual_balance":                 balanceConcept,
	"annual_deduction_excluded_cash": excludedCash,
	"annual_deduction_excluded_unverified": fixed(
		"Donativos no aplicados: falta tu ingreso acumulable del año anterior para verificar el tope"),
	"annual_deduction_cap_applied": capApplied,
	// pf_resico_check (ANN-09)
	"annual_resico_income":      fixed("Ingresos cobrados en RESICO en el año"),
	"annual_resico_rate":        resicoRate,
	"annual_resico_above_table": fixed("Ingresos arriba de la tabla anual de RESICO: no hay tasa"),
	"annual_resico_payments":    paymentsConcept,
	"annual_resico_withheld":    fixed("ISR que te retuvieron personas morales (1.25 %)"),
	"annual_resico_difference":  fixed("Diferencia informativa (no se paga ni se devuelve)"),
	"annual_resico_definitive":  fixed("Tus pagos mensuales de RESICO son definitivos"),
	// pm_601, pm_resico (ANN-10, ANN-11)
	domain.LabelKey(pm.LineIncome):              fixed("Ingresos acumulables del ejercicio"),
	domain.LabelKey(pm.LineDeductions):          fixed("Deducciones autorizadas"),
	domain.LabelKey(pm.LineDepreciation):        fixed("Depreciación del ejercicio"),
	domain.LabelKey(pm.LineProfitBeforePTU):     fixed("Utilidad antes de PTU pagada"),
	domain.LabelKey(pm.LinePTUPaid):             fixed("PTU pagada en el ejercicio"),
	domain.LabelKey(pm.LineTaxableProfit):       fixed("Utilidad fiscal"),
	domain.LabelKey(pm.LineNoProfit):            fixed("Sin utilidad fiscal: pérdida amortizable en años siguientes"),
	domain.LabelKey(pm.LinePriorYearLosses):     fixed("Pérdidas fiscales de ejercicios anteriores aplicadas"),
	domain.LabelKey(pm.LineBase):                fixed("Resultado fiscal"),
	domain.LabelKey(pm.LineRate):                fixed("Tasa del ISR de personas morales"),
	domain.LabelKey(pm.LineAnnualTax):           fixed("ISR del ejercicio"),
	domain.LabelKey(pm.LinePaymentsFiled):       pmPayments("presentados"),
	domain.LabelKey(pm.LinePaymentsTaxInput):    pmPayments("registrados en tus datos fiscales"),
	domain.LabelKey(pm.LinePaymentsEstimated):   pmPayments("estimados por la app"),
	domain.LabelKey(pm.LineInterestWithholding): fixed("ISR retenido por intereses"),
	domain.LabelKey(pm.LineBalance):             balanceConcept,
	domain.LabelKey(pm.LineRefundOnlyPaid):      fixed("Solo se devuelve lo efectivamente pagado o retenido"),
	domain.LabelKey(pm.LinePTUBase):             fixed("Base de la PTU"),
	domain.LabelKey(pm.LinePTURate):             fixed("Porcentaje de PTU a repartir"),
	domain.LabelKey(pm.LinePTUToDistribute):     fixed("PTU a repartir"),
	// the Ledger's own
	domain.LabelRoundingAdjustment: fixed("Ajuste por redondeo"),
}

// genericConcept stands for a label the catalog does not know yet.
const genericConcept = "Paso del cálculo"

// concept renders a line's concept.
func concept(l domain.Line, fiscalYear int) string {
	if f, ok := concepts[l.Label]; ok {
		return f(l, fiscalYear)
	}
	return genericConcept
}

// HasConcept reports a label with its own concept (the coverage test).
func HasConcept(label domain.LabelKey) bool {
	_, ok := concepts[label]
	return ok
}

func arg(l domain.Line, i int) string {
	if i < len(l.Args) {
		return l.Args[i]
	}
	return ""
}

func salaryConcept(l domain.Line, _ int) string {
	switch n := arg(l, 0); n {
	case "", "0":
		return "Sueldos acumulables (sin constancias registradas)"
	case "1":
		return "Sueldos acumulables de tu constancia"
	default:
		return "Sueldos acumulables de tus " + n + " constancias"
	}
}

func businessIncome(l domain.Line, _ int) string {
	if arg(l, 0) == "625" {
		return "Ingresos cobrados en el año por plataformas"
	}
	return "Ingresos cobrados en el año (actividad empresarial y profesional)"
}

func blindDeduction(l domain.Line, _ int) string {
	if p, ok := percent(arg(l, 1)); ok {
		return "Deducción opcional del " + p + " de las rentas"
	}
	return "Deducción opcional de las rentas"
}

func tariffRow(l domain.Line, _ int) string {
	if l.Row > 0 {
		return "Cuota fija (tarifa anual, renglón " + strconv.Itoa(l.Row) + ")"
	}
	return "Cuota fija (tarifa anual)"
}

func resicoRate(l domain.Line, _ int) string {
	if row := arg(l, 0); row != "" {
		return "Tasa anual de RESICO (renglón " + row + ")"
	}
	return "Tasa anual de RESICO"
}

// paymentSourceTexts name a payment line's source.
var paymentSourceTexts = map[adomain.PaymentSource]string{
	adomain.PaymentFiled:          "presentados",
	adomain.PaymentTaxInput:       "registrados en tus datos fiscales",
	adomain.PaymentEstimatedByApp: "estimados por la app",
}

// paymentsConcept renders a PF payment line: args regime, source, months.
func paymentsConcept(l domain.Line, _ int) string {
	return "Pagos provisionales de ISR " + paymentSourceTexts[adomain.PaymentSource(arg(l, 1))] + months(arg(l, 2))
}

// pmPayments renders a PM payment line: args months.
func pmPayments(source string) conceptFunc {
	return func(l domain.Line, _ int) string { return "Pagos provisionales de ISR " + source + months(arg(l, 0)) }
}

func months(n string) string {
	switch n {
	case "":
		return ""
	case "1":
		return " (1 mes)"
	}
	return " (" + n + " meses)"
}

func balanceConcept(l domain.Line, _ int) string {
	switch {
	case l.Amount > 0:
		return "Saldo a cargo"
	case l.Amount < 0:
		return "Saldo a favor"
	}
	return "Sin saldo"
}

// categoryTexts name each deduction category.
var categoryTexts = map[string]string{
	string(adomain.CategoryMedical):                "gastos médicos",
	string(adomain.CategoryMedicalDisability):      "gastos médicos por incapacidad",
	string(adomain.CategoryFuneral):                "gastos funerales",
	string(adomain.CategoryDonation):               "donativos",
	string(adomain.CategoryMortgageRealInterest):   "intereses reales de crédito hipotecario",
	string(adomain.CategoryRetirementContribution): "aportaciones complementarias para el retiro",
	string(adomain.CategoryHealthInsurance):        "primas de seguros de gastos médicos",
	string(adomain.CategorySchoolTransport):        "transporte escolar",
	string(adomain.CategoryLocalPayrollTax):        "impuesto local sobre sueldos",
	string(adomain.CategoryTuition):                "colegiaturas",
}

// capTexts name each cap of a "Tope aplicado" line.
var capTexts = map[string]string{
	string(adomain.CapFuneral):             "gastos funerales (1 UMA anual)",
	string(adomain.CapDonations):           "donativos (7 % de tus ingresos del año anterior)",
	string(adomain.CapGovernmentDonations): "donativos a gobierno (4 % de tus ingresos del año anterior)",
	string(adomain.CapRetirement):          "aportaciones para el retiro (10 % de tus ingresos o 5 UMA anuales)",
	string(adomain.CapTuition):             "colegiaturas",
	string(adomain.CapGlobal):              "deducciones personales (15 % de tus ingresos totales o 5 UMA anuales)",
}

// excludedCash renders a cash exclusion: args category, concept, from, to.
func excludedCash(l domain.Line, _ int) string {
	return "No deducible por pago en efectivo: " + textOr(categoryTexts, arg(l, 0), "deducción") + ", " + pesos(arg(l, 2))
}

// capApplied renders "Tope aplicado: {concepto}, de {x} a {y}"; a tuition
// cap names its student.
func capApplied(l domain.Line, _ int) string {
	subject := textOr(capTexts, arg(l, 0), "deducción")
	if arg(l, 0) == string(adomain.CapTuition) && arg(l, 1) != "" {
		subject += " de " + arg(l, 1)
	}
	return "Tope aplicado: " + subject + ", de " + pesos(arg(l, 2)) + " a " + pesos(arg(l, 3))
}

func textOr(m map[string]string, key, fallback string) string {
	if t, ok := m[key]; ok {
		return t
	}
	return fallback
}

// pesos renders a decimal amount ("12900.00") as "$12,900.00".
func pesos(decimal string) string {
	whole, frac, _ := strings.Cut(decimal, ".")
	sign := ""
	if strings.HasPrefix(whole, "-") {
		sign, whole = "-", whole[1:]
	}
	if whole == "" || !allDigits(whole) {
		return decimal
	}
	var b strings.Builder
	for i, r := range whole {
		if i > 0 && (len(whole)-i)%3 == 0 {
			b.WriteByte(',')
		}
		b.WriteRune(r)
	}
	if frac != "" {
		return sign + "$" + b.String() + "." + frac
	}
	return sign + "$" + b.String()
}

// percent renders a decimal rate ("0.35") as "35 %".
func percent(decimal string) (string, bool) {
	whole, frac, _ := strings.Cut(decimal, ".")
	if whole == "" || !allDigits(whole) || (frac != "" && !allDigits(frac)) {
		return "", false
	}
	frac += strings.Repeat("0", max(0, 2-len(frac)))
	n, err := strconv.Atoi(whole + frac[:2])
	if err != nil {
		return "", false
	}
	out := strconv.Itoa(n)
	if rest := strings.TrimRight(frac[2:], "0"); rest != "" {
		out += "." + rest
	}
	return out + " %", true
}

func allDigits(s string) bool {
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return s != ""
}

// formatFixed renders n / scale with the scale's decimals ("0.1380").
func formatFixed(n, scale int64) string {
	digits := len(strconv.FormatInt(scale, 10)) - 1
	frac := strconv.FormatInt(n%scale, 10)
	return strconv.FormatInt(n/scale, 10) + "." + strings.Repeat("0", digits-len(frac)) + frac
}

// lineSource is the contract AnnualLineSource of a line: each payment by
// its source (filed_payment, tax_input:prior_provisional_payments,
// estimated_by_app), the constancias' and the deductions' own sources.
func lineSource(l domain.Line) string {
	switch {
	case l.Code == adomain.LineArt152Payments || l.Code == adomain.LineRESICOCheckPayments:
		return paymentSource(adomain.PaymentSource(arg(l, 1)))
	case l.Code == pm.LinePaymentsFiled:
		return paymentSource(adomain.PaymentFiled)
	case l.Code == pm.LinePaymentsTaxInput:
		return paymentSource(adomain.PaymentTaxInput)
	case l.Code == pm.LinePaymentsEstimated:
		return paymentSource(adomain.PaymentEstimatedByApp)
	case l.Code == adomain.LineArt152Salary || l.Code == adomain.LineArt152SalaryISR:
		return "salary_statements"
	case l.Code == adomain.LineArt152PersonalDeductions || strings.HasPrefix(string(l.Code), "deductions."):
		return "personal_deductions"
	}
	if s := l.Source.String(); s != "" {
		return s
	}
	return "computed"
}

func paymentSource(s adomain.PaymentSource) string {
	switch s {
	case adomain.PaymentFiled:
		return "filed_payment"
	case adomain.PaymentTaxInput:
		return "tax_input:" + string(domain.InputPriorProvisionalPayments)
	}
	return "estimated_by_app"
}

// reasonTexts are the obligation reasons in plain words.
var reasonTexts = map[usecaseannual.ReasonCode]string{
	usecaseannual.ReasonBusinessIncome:        "Tuviste ingresos por actividad empresarial o profesional.",
	usecaseannual.ReasonPlatformIncome:        "Tuviste ingresos por plataformas sin el pago definitivo.",
	usecaseannual.ReasonRentIncome:            "Tuviste ingresos por arrendamiento.",
	usecaseannual.ReasonSimultaneousEmployers: "Tuviste dos patrones al mismo tiempo.", //nolint:misspell // es-MX "patrones" (employers), US-A4
	usecaseannual.ReasonLeftBeforeYearEnd:     "Dejaste de trabajar con un patrón antes del 31 de diciembre.",
	usecaseannual.ReasonPayerNotWithholding:   "Un patrón no estaba obligado a retenerte.",
	usecaseannual.ReasonToldEmployerWillFile:  "Le avisaste a tu patrón que presentarías tu declaración.",
	usecaseannual.ReasonSalaryOverThreshold:   "Tus sueldos del año pasan de $400,000.",
	usecaseannual.ReasonSalaryOnly: "No estás obligado, pero puedes presentarla para deducir tus gastos personales " +
		"(una devolución es la razón usual).",
	usecaseannual.ReasonNoIncome:     "No tuviste ingresos que se declaren en la anual.",
	usecaseannual.ReasonRESICORelief: "Tus pagos mensuales de RESICO son definitivos: no presentas anual por estos ingresos.",
	usecaseannual.ReasonPersonaMoral: "Las personas morales presentan declaración anual.",
}

func reasonText(r usecaseannual.ReasonCode) string {
	if t, ok := reasonTexts[r]; ok {
		return t
	}
	return genericConcept
}

// noteText renders a note under a section's figure.
func noteText(n usecaseannual.NoteCode, fiscalYear int) string {
	switch n {
	case usecaseannual.NoteRefundOnlyPaid:
		return "Solo se devuelve lo efectivamente pagado o retenido: confirma tus pagos con tus acuses."
	case usecaseannual.NoteRefundNeedsEFirma:
		return "Para solicitar la devolución necesitas e.firma."
	case usecaseannual.NoteLossCarryforward:
		return "Tuviste pérdida en el ejercicio: puedes amortizarla en los años siguientes."
	case usecaseannual.NoteCUSuggestion:
		return "Sugerido para tus pagos de " + strconv.Itoa(fiscalYear+1) + "; confírmalo con tu contador."
	}
	return "Revisa este dato con tu contador."
}

// missingLabels are the CTA labels of the missing inputs.
var missingLabels = map[string]string{
	adomain.MissingSalaryStatement:            "Constancias de sueldos",
	adomain.MissingPriorYearAccumulableIncome: "Ingresos acumulables del año anterior",
	"filed_payment":                           "Pagos provisionales presentados",
}

func missingLabel(key string) string {
	return textOr(missingLabels, key, "Datos fiscales del régimen")
}

// warningText renders a write warning.
func warningText(code usecaseannual.WarningCode) string {
	switch code {
	case usecaseannual.WarningCashNotDeductible:
		return "En efectivo no es deducible: lo guardamos, pero no se resta en tu anual."
	case usecaseannual.WarningSalaryRegimeNotHeld:
		return "No tienes el régimen de sueldos y salarios (605) en ese año; agrégalo a tu contribuyente."
	}
	return "Revisa este dato."
}
