package fiscal

import (
	"strconv"
	"strings"

	"github.com/romodevx/control-contable/backend/internal/platform/money"
)

// Side is which side of a record the taxpayer is on: the emisor of an
// income, or the receptor (a persona moral) of an expense. It picks the
// wording of an explanation ("te retiene" or "debes retener").
type Side string

// The record sides.
const (
	SideIncome  Side = "income"
	SideExpense Side = "expense"
)

// Explanation is one plain-words line (es-MX) for a matched rule, with its
// legal basis (contract WithholdingExplanation). It never holds a name, an
// RFC or a description.
type Explanation struct {
	RuleID RuleID
	Text   string
}

// ExplainInput is a derivation, fresh or stored, as Explain words it.
type ExplainInput struct {
	Side Side
	// Receptor is who was billed (income only): it names the withholder.
	Receptor ReceptorKind
	// RuleIDs are the ISR rule, then the IVA rule.
	RuleIDs []RuleID
	ISRRate money.Rate
	// ISR and IVA are the amounts withheld: a zero amount reads "no".
	ISR, IVA   money.Cents
	IVABasis   IVABasis
	Withholder Withholder
	// PlatformActivity names a persona física's platform sale (ISR-3).
	PlatformActivity PlatformActivity
}

// Explain returns one line per rule id, in order (ADR-0008: the server
// writes the explanation; clients never keep a rule table of their own).
// The wording depends only on the input, so a stored derivation reads the
// same on every request.
func Explain(in ExplainInput) []Explanation {
	out := make([]Explanation, 0, len(in.RuleIDs))
	for _, id := range in.RuleIDs {
		text := in.ivaLine(id)
		if strings.HasPrefix(string(id), isrRulePrefix) {
			text = in.isrLine(id)
		}
		out = append(out, Explanation{RuleID: id, Text: text})
	}
	return out
}

// Explain words d on side, for the receptor billed (income) and the 625
// activity of a platform sale.
func (d Derivation) Explain(side Side, receptor ReceptorKind, activity PlatformActivity) []Explanation {
	return Explain(ExplainInput{
		Side: side, Receptor: receptor, RuleIDs: d.RuleIDs, ISRRate: d.ISRRate, ISR: d.ISR, IVA: d.IVA,
		IVABasis: d.IVABasis, Withholder: d.Withholder, PlatformActivity: activity,
	})
}

const isrRulePrefix = "ISR-"

// incomeISRLines are the income ISR lines that name who withholds ({p}) and
// the rate ({r}).
var incomeISRLines = map[RuleID]string{
	RuleISR4: "{p} te retiene {r} de ISR porque estás en RESICO (LISR art. 113-J).",
	RuleISR5: "{p} te retiene {r} de ISR por honorarios (LISR art. 106).",
	RuleISR6: "{p} te retiene {r} de ISR por arrendamiento (LISR art. 116).",
	RuleISR7: "{p} no te retiene ISR en esta operación.",
}

// expenseISRLines are the expense ISR lines; any other rule withholds no ISR
// from a supplier.
var expenseISRLines = map[RuleID]string{
	RuleISR1: "No calculamos retenciones a proveedores en el extranjero.",
	RuleISR2: "No retienes ISR a una persona moral.",
	RuleISR4: "Debes retener {r} de ISR porque tu proveedor está en RESICO (LISR art. 113-J).",
	RuleISR5: "Debes retener {r} de ISR por honorarios de persona física (LISR art. 106).",
	RuleISR6: "Debes retener {r} de ISR por arrendamiento de persona física (LISR art. 116).",
}

// payers name the receptor that withholds from the taxpayer.
var payers = map[ReceptorKind]string{
	ReceptorPM:                "La empresa",
	ReceptorGovFederal:        "El gobierno federal",
	ReceptorGovStateMunicipal: "El gobierno estatal o municipal",
}

// platformSales word a persona física's platform sale.
var platformSales = map[PlatformActivity]string{
	PlatformGoodsServices:     "venta de bienes y servicios",
	PlatformTransportDelivery: "transporte o entrega",
	PlatformLodging:           "hospedaje",
}

// ivaBodies are what a positive IVA line withholds, by rule.
var ivaBodies = map[RuleID]string{
	RuleIVA4: "4 % del precio por flete de mercancía por carretera (LIVA art. 1-A, RLIVA art. 3).",
	RuleIVA5: "todo el IVA por desperdicios y reciclables (LIVA art. 1-A).",
	RuleIVA6: "2/3 del IVA por servicios de persona física (LIVA art. 1-A, RLIVA art. 3).",
	RuleIVA7: "2/3 del IVA por servicios de persona física al gobierno federal (LIVA art. 3, RLIVA art. 3).",
}

// The fixed lines and fragments.
const (
	platformName         = "La plataforma"
	defaultPayer         = "Quien te paga"
	defaultPlatformSale  = "tus ventas en la plataforma"
	defaultIVABody       = "parte del IVA (LIVA art. 1-A)."
	noIVALine            = "La operación no lleva IVA: no hay IVA que retener."
	platformHalfIVA      = "la mitad del IVA (LIVA art. 18-J)."
	platformAllIVA       = "todo el IVA porque no le diste tu RFC a la plataforma (LIVA art. 18-J)."
	pfNoActivityISRLine  = "Nadie te retiene ISR cuando le facturas a una persona sin negocio o al público en general."
	pfISRLine            = "Nadie te retiene ISR cuando le facturas a una persona física."
	foreignISRLine       = "Nadie te retiene ISR cuando le facturas a alguien en el extranjero."
	moralPlatformISRLine = "La plataforma te retiene {r} de ISR (LISR art. 113-A; LIF art. 25 fr. VI)."
	moralISRLine         = "Como persona moral, no te retienen ISR en esta operación."
	expenseNoISRLine     = "No retienes ISR en esta operación."
)

func (in ExplainInput) isrLine(id RuleID) string {
	if in.Side == SideExpense {
		line, ok := expenseISRLines[id]
		if !ok {
			line = expenseNoISRLine
		}
		return in.fill(line)
	}
	switch id {
	case RuleISR0:
		if in.Receptor == ReceptorPFNoActivityOrPublic {
			return pfNoActivityISRLine
		}
		return pfISRLine
	case RuleISR1:
		return foreignISRLine
	case RuleISR2:
		if in.ISR > 0 {
			return in.fill(moralPlatformISRLine)
		}
		return moralISRLine
	case RuleISR3:
		sale, ok := platformSales[in.PlatformActivity]
		if !ok {
			sale = defaultPlatformSale
		}
		return in.fill("La plataforma te retiene {r} de ISR por " + sale + " (LISR art. 113-A).")
	}
	line, ok := incomeISRLines[id]
	if !ok {
		line = incomeISRLines[RuleISR7]
	}
	return in.fill(line)
}

func (in ExplainInput) ivaLine(id RuleID) string {
	if id == RuleIVANoTax {
		return noIVALine
	}
	if in.IVA == 0 {
		return in.noIVAWithheld()
	}
	body := in.ivaBody(id)
	switch {
	case in.ISR > 0:
		return "Y " + body
	case in.Side == SideExpense:
		return "Debes retener " + body
	}
	return in.withholderName() + " te retiene " + body
}

// noIVAWithheld words a zero IVA withholding: "tampoco" after a zero ISR.
func (in ExplainInput) noIVAWithheld() string {
	switch {
	case in.Side == SideExpense && in.ISR > 0:
		return "No retienes IVA en esta operación."
	case in.Side == SideExpense:
		return "Tampoco retienes IVA en esta operación."
	case in.ISR > 0:
		return "No te retienen IVA en esta operación."
	}
	return "Tampoco te retienen IVA en esta operación."
}

func (in ExplainInput) ivaBody(id RuleID) string {
	if id == RuleIVA3 {
		if in.IVABasis == IVABasisAllOfIVA {
			return platformAllIVA
		}
		return platformHalfIVA
	}
	body, ok := ivaBodies[id]
	if !ok {
		return defaultIVABody
	}
	return body
}

// withholderName names who withholds from the taxpayer on an income.
func (in ExplainInput) withholderName() string {
	if in.Withholder == WithholderPlatform {
		return platformName
	}
	name, ok := payers[in.Receptor]
	if !ok {
		return defaultPayer
	}
	return name
}

// fill replaces {p} with the withholder and {r} with the ISR rate.
func (in ExplainInput) fill(line string) string {
	return strings.NewReplacer("{p}", in.withholderName(), "{r}", Percent(in.ISRRate)).Replace(line)
}

// percentScale is money.RateScale / 100: a Rate's millionths per percent
// point.
const percentScale = money.RateScale / 100

// Percent words a rate as a percentage without trailing zeros: 0.0125 is
// "1.25 %", 0.10 is "10 %".
func Percent(r money.Rate) string {
	whole := int64(r) / percentScale
	frac := int64(r) % percentScale
	text := strconv.FormatInt(whole, 10)
	if frac != 0 {
		digits := strconv.FormatInt(percentScale+frac, 10)[1:]
		text += "." + strings.TrimRight(digits, "0")
	}
	return text + " %"
}
