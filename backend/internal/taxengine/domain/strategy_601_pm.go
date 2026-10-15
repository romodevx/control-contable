package domain

import (
	"fmt"
	"slices"
	"time"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain/taxparams"
)

// Names in this file carry a generalPM prefix: every regime strategy lives
// in package domain (PLAN PG-M5-S writes them in parallel), so a generic
// name would collide.

// generalPMRegime is régimen 601, General de Ley Personas Morales.
const generalPMRegime fiscal.RegimeCode = "601"

// generalPMRateKey is the persona moral ISR rate (LISR 9): 30 %.
const generalPMRateKey taxparams.Key = "isr.rate.persona_moral"

// The PTU paid is deducted in equal parts from May to December (LISR 14
// fr. II a)): eight parts, the first in May.
const (
	generalPMPTUParts      = 8
	generalPMPTUMonthsLead = int64(time.April) // part k is deducted in month 4 + k
)

// The line codes of the 601 PM ISR and of its ISR withheld to remit.
const (
	generalPMLineIncomeYTD           LineCode = "isr.income_ytd"
	generalPMLineCoeficiente         LineCode = "isr.coeficiente"
	generalPMLinePTU                 LineCode = "isr.ptu"
	generalPMLineLosses              LineCode = "isr.prior_year_losses"
	generalPMLineBase                LineCode = "isr.base"
	generalPMLineRate                LineCode = "isr.rate"
	generalPMLinePriorPayments       LineCode = "isr.prior_payments"
	generalPMLineInterestWithholding LineCode = "isr.interest_withholding"
	generalPMLinePayable             LineCode = "isr.payable"
	generalPMLineFirstFiscalYear     LineCode = "isr.first_fiscal_year"
	generalPMLineISRWithheldToRemit  LineCode = "withholding.isr_to_remit"
	generalPMLineWithheldToRemit     LineCode = "withholding.to_remit"
)

// The legal basis of each 601 step (MODULE-tax-engine §2.5 citation
// catalog; research §B.2.1).
var (
	generalPMBasisCoeficiente = []Citation{{Law: "LISR", Article: "14", Note: "fracción I, coeficiente de utilidad"}}
	generalPMBasisIncome      = []Citation{{Law: "LISR", Article: "14", Note: "fracción II, ingresos nominales"}}
	generalPMBasisPTU         = []Citation{{Law: "LISR", Article: "14", Note: "fracción II, inciso a), PTU de mayo a diciembre"}}
	generalPMBasisLosses      = []Citation{{Law: "LISR", Article: "14", Note: "fracción II, inciso c), pérdidas fiscales"}}
	generalPMBasisBase        = []Citation{{Law: "LISR", Article: "14", Note: "fracción II, utilidad fiscal del pago"}}
	generalPMBasisRate        = []Citation{
		{Law: "LISR", Article: "9", Note: "tasa del 30 %"}, {Law: "LISR", Article: "14", Note: "fracción III"},
	}
	generalPMBasisPayments = []Citation{{Law: "LISR", Article: "14", Note: "fracción III, pagos provisionales anteriores"}}
	generalPMBasisInterest = []Citation{
		{Law: "LISR", Article: "54", Note: "retención sobre intereses"},
		{Law: "LIF 2026", Article: "24", Note: "tasa anual del 0.90 %"},
	}
	generalPMBasisPayable   = []Citation{{Law: "LISR", Article: "14", Note: "fracción III, pago provisional"}}
	generalPMBasisFirstYear = []Citation{{Law: "LISR", Article: "14", Note: "primer ejercicio: sin pagos provisionales"}}
	generalPMBasisISRRemit  = []Citation{
		{Law: "LISR", Article: "106", Note: "párrafo quinto"},
		{Law: "LISR", Article: "116", Note: "párrafo quinto"},
		{Law: "LISR", Article: "113-J"},
	}
	generalPMBasisToRemit = append(slices.Clone(generalPMBasisISRRemit),
		Citation{Law: "LIVA", Article: "1-A", Note: "entero de la retención"})
)

// GeneralPMStrategy is régimen 601, General de Ley Personas Morales
// (MODULE-tax-engine §2.4, research §B.2.1). The monthly provisional ISR
// (LISR 14) is
//
//	coeficiente de utilidad × nominal income January..n on accrual
//	− PTU paid (in eighths, May to December, capped at that profit)
//	− prior-year losses (applied up to what remains)
//	= base × 30 % − prior provisional payments − ISR withheld on interest.
//
// In the first fiscal year there is no coeficiente and no provisional
// payment: ISR 0 with no_coefficient_first_year. The coeficiente is used as
// given (art. 14 is silent on its decimals). The IVA is the shared
// monthly components (ComputeIVA); the IVA and ISR the company withheld
// from its suppliers are its withholdings to remit.
type GeneralPMStrategy struct{}

// NewGeneralPMStrategy returns the 601 PM strategy.
func NewGeneralPMStrategy() GeneralPMStrategy { return GeneralPMStrategy{} }

// Key is 601/moral.
func (GeneralPMStrategy) Key() StrategyKey {
	return StrategyKey{Regime: generalPMRegime, PersonType: fiscal.PersonMoral}
}

// Treatment is full: ISR and IVA are estimated.
func (GeneralPMStrategy) Treatment() Treatment { return TreatmentFull }

// RequiredInputs are the coeficiente and the provisional payment of every
// active month before the period's; nothing in a first fiscal year, which
// has neither (MODULE-tax-engine §2.8).
func (GeneralPMStrategy) RequiredInputs(_ Options, p Period, f Facts) []RequiredInput {
	if generalPMFirstFiscalYear(f.Inputs) {
		return nil
	}
	required := []RequiredInput{{Key: InputCoeficienteUtilidad, Kind: InputRate}}
	var prior []time.Month
	for _, m := range f.ActiveMonths {
		if m < p.Month() {
			prior = append(prior, m)
		}
	}
	if len(prior) > 0 {
		required = append(required, RequiredInput{
			Key: InputPriorProvisionalPayments, Kind: InputMoneyMonthly, Months: prior,
		})
	}
	return required
}

func generalPMFirstFiscalYear(in TaxInputs) bool {
	first, _ := in.Bool(InputFirstFiscalYear) // absent: not the first year
	return first
}

// Estimate computes the period's ISR, IVA and withholdings to remit, with
// zero for any missing input.
func (s GeneralPMStrategy) Estimate(in Input, p *taxparams.Params) (RegimeEstimate, error) {
	l := NewLedger()
	est := RegimeEstimate{Key: s.Key(), Coverage: TreatmentFull}

	isr, err := s.isr(l, in, p)
	if err != nil {
		return RegimeEstimate{}, fmt.Errorf("601 ISR %s: %w", in.Period, err)
	}
	est.ISR = &isr
	if generalPMFirstFiscalYear(in.Inputs) {
		est.Disclaimers = append(est.Disclaimers, DisclaimerNoCoefficientFirstYear)
	}

	iva, err := ComputeIVA(l, monthIVA(in))
	if err != nil {
		return RegimeEstimate{}, fmt.Errorf("601 %s: %w", in.Period, err)
	}
	est.IVA = &iva.Components
	est.Disclaimers = append(est.Disclaimers, iva.Disclaimers...)
	est.CPATopics = append(est.CPATopics, iva.CPATopics...)

	isrToRemit := in.Sums.Month(in.Period.Month()).Expenses.ISRWithheldToRemit
	est.WithheldToRemit = generalPMWithholdingsToRemit(l, iva.WithheldToRemit, isrToRemit)

	if est.Lines, err = l.Lines(); err != nil {
		return RegimeEstimate{}, fmt.Errorf("601 %s: %w", in.Period, err)
	}
	return est, nil
}

// isr traces the provisional payment, or the first fiscal year's zero.
func (GeneralPMStrategy) isr(l *Ledger, in Input, p *taxparams.Params) (TaxResult, error) {
	if generalPMFirstFiscalYear(in.Inputs) {
		l.Info(generalPMStep(generalPMLineFirstFiscalYear, SourceTaxInput(InputFirstFiscalYear), generalPMBasisFirstYear), 0)
		return NewTaxResult(0), nil
	}
	rate, err := p.Rate(generalPMRateKey, in.Period.End())
	if err != nil {
		return TaxResult{}, err
	}
	a, err := generalPMAmountsOf(in)
	if err != nil {
		return TaxResult{}, err
	}
	coeficiente, _ := in.Inputs.Rate(InputCoeficienteUtilidad) // absent: zero, the estimate is incomplete

	l.Input(generalPMStep(generalPMLineIncomeYTD, SourceRecords(), generalPMBasisIncome), a.income)
	l.Scale(generalPMStep(generalPMLineCoeficiente, SourceTaxInput(InputCoeficienteUtilidad), generalPMBasisCoeficiente), coeficiente)
	l.Subtract(generalPMStep(generalPMLinePTU, SourceTaxInput(InputPTUPaid), generalPMBasisPTU),
		generalPMUpTo(a.ptuShare, l.Running()))
	l.Subtract(generalPMStep(generalPMLineLosses, SourceTaxInput(InputPriorYearLosses), generalPMBasisLosses),
		generalPMUpTo(a.losses, l.Running()))
	l.Result(generalPMStep(generalPMLineBase, SourceComputed(), generalPMBasisBase))
	l.ScaleRounded(generalPMStep(generalPMLineRate, SourceParams(generalPMRateKey), generalPMBasisRate), rate)
	l.Subtract(generalPMStep(generalPMLinePriorPayments, SourceTaxInput(InputPriorProvisionalPayments), generalPMBasisPayments),
		a.priorPayments)
	l.Subtract(generalPMStep(generalPMLineInterestWithholding, SourceTaxInput(InputInterestWithholding), generalPMBasisInterest),
		a.interest)
	payable := l.Result(generalPMStep(generalPMLinePayable, SourceComputed(), generalPMBasisPayable))
	if l.err != nil {
		return TaxResult{}, l.err
	}
	return NewTaxResult(payable), nil
}

// generalPMAmounts are the ISR figures the ledger steps take, as Exact.
type generalPMAmounts struct {
	income        money.Exact // nominal income January..n, accrual
	ptuShare      money.Exact // the PTU parts due by month n, before the cap
	losses        money.Exact
	priorPayments money.Exact
	interest      money.Exact
}

func generalPMAmountsOf(in Input) (generalPMAmounts, error) {
	ytd, err := in.Sums.Sum(time.January, in.Period.Month())
	if err != nil {
		return generalPMAmounts{}, err
	}
	ptu, _ := in.Inputs.Money(InputPTUPaid)            // absent: zero
	losses, _ := in.Inputs.Money(InputPriorYearLosses) // absent: zero
	interest, _ := in.Inputs.Money(InputInterestWithholding)
	payments, err := in.Inputs.SumBefore(InputPriorProvisionalPayments, in.Period.Month())
	if err != nil {
		return generalPMAmounts{}, err
	}

	var c generalPMExact
	a := generalPMAmounts{
		income:        c.of(ytd.Income.Accrual),
		ptuShare:      c.ptuShare(c.of(ptu), in.Period.Month()),
		losses:        c.of(losses),
		priorPayments: c.of(payments),
		interest:      c.of(interest),
	}
	if c.err != nil {
		return generalPMAmounts{}, c.err
	}
	return a, nil
}

// generalPMExact converts with a sticky error, so the amounts read as one
// struct literal.
type generalPMExact struct{ err error }

func (c *generalPMExact) of(v money.Cents) money.Exact {
	if c.err != nil {
		return 0
	}
	e, err := v.Exact()
	c.err = err
	return e
}

// ptuShare is the PTU deductible in month n: ptu ÷ 8 × (n − 4) from May,
// cumulatively; nothing before May (LISR 14 fr. II a)).
func (c *generalPMExact) ptuShare(ptu money.Exact, n time.Month) money.Exact {
	parts := int64(n) - generalPMPTUMonthsLead
	if c.err != nil || parts <= 0 {
		return 0
	}
	share, err := ptu.MulRatio(parts, generalPMPTUParts)
	c.err = err
	return share
}

// generalPMUpTo caps a deduction at the profit it reduces (never below
// zero): the PTU is capped at the payment's profit (LISR 14 fr. II a)) and
// a loss is applied only up to what remains, the rest staying pending.
func generalPMUpTo(deduction, profit money.Exact) money.Exact {
	return max(min(deduction, profit), 0)
}

// generalPMWithholdingsToRemit closes the withholdings-to-remit group and
// returns its total, read from the ledger so the amount and the breakdown
// cannot drift (ADR-0002 §1). The group is apart from the company's own ISR
// and IVA: ComputeIVA has already anchored the IVA it withheld from PF
// suppliers (LIVA 1-A), and the ISR it withheld (fees, rent, RESICO PF) is
// added to it and closed by one total line. Without IVA to remit, the ISR
// starts the group itself, so it never lands on the IVA's running value;
// a group of one line needs no total. On a ledger error it returns zero and
// Lines reports the error.
func generalPMWithholdingsToRemit(l *Ledger, ivaToRemit, isrToRemit money.Cents) money.Cents {
	if isrToRemit == 0 {
		return ivaToRemit
	}
	isrStep := Step{
		Tax: TaxWithholding, Code: generalPMLineISRWithheldToRemit, Label: LabelKey(generalPMLineISRWithheldToRemit),
		Source: SourceRecords(), Basis: generalPMBasisISRRemit,
	}
	exact, err := isrToRemit.Exact()
	if l.fail(isrStep, err) {
		return 0
	}
	if ivaToRemit == 0 {
		l.Input(isrStep, exact)
		return isrToRemit
	}
	l.Add(isrStep, exact)
	return l.Result(Step{
		Tax: TaxWithholding, Code: generalPMLineWithheldToRemit, Label: LabelKey(generalPMLineWithheldToRemit),
		Source: SourceComputed(), Basis: generalPMBasisToRemit,
	}).Round()
}

func generalPMStep(code LineCode, source Source, basis []Citation) Step {
	return Step{Tax: TaxISR, Code: code, Label: LabelKey(code), Source: source, Basis: basis}
}
