package domain

import (
	"fmt"
	"slices"
	"time"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain/taxparams"
)

// The parameters RESICO PM reads (params/data/<year>).
const (
	paramPersonaMoralRate taxparams.Key = "isr.rate.persona_moral"
	paramRESICOPMCap      taxparams.Key = "regime.resico_pm.income_cap"
)

// The line codes of the RESICO PM ISR, in order. The cap line comes first
// when it applies; a tax input's line only when its value is not zero; the
// no-profit line only after a negative base; prior payments from February.
const (
	LineRESICOPMCapExceeded         LineCode = "isr.resico_cap_exceeded"
	LineRESICOPMIncomeYTD           LineCode = "isr.income_ytd"
	LineRESICOPMIncomeBeforeApp     LineCode = "isr.income_before_app"
	LineRESICOPMDeductionsYTD       LineCode = "isr.deductions_ytd"
	LineRESICOPMDeductionsBeforeApp LineCode = "isr.deductions_before_app"
	LineRESICOPMPTUPaid             LineCode = "isr.ptu_paid"
	LineRESICOPMPriorYearLosses     LineCode = "isr.prior_year_losses"
	LineRESICOPMBase                LineCode = "isr.base"
	LineRESICOPMNoProfit            LineCode = "isr.no_profit"
	LineRESICOPMRate                LineCode = "isr.rate"
	LineRESICOPMPriorPayments       LineCode = "isr.prior_payments"
	LineRESICOPMInterestWithholding LineCode = "isr.interest_withholding"
	LineRESICOPMPayable             LineCode = "isr.payable"
)

// The legal basis of each RESICO PM step (MODULE-tax-engine §2.5: LISR
// 207, 208, 211, 9; research B.2.2).
var (
	basisRESICOPMCap = []Citation{
		{Law: "LISR", Article: "206", Note: "párrafo segundo: ingresos del ejercicio mayores a 35,000,000"},
		{Law: "LISR", Article: "214", Note: "se tributa en el Título II a partir del ejercicio siguiente"},
	}
	basisRESICOPMIncome = []Citation{
		{Law: "LISR", Article: "207", Note: "ingresos efectivamente percibidos"},
		{Law: "LISR", Article: "211", Note: "ingresos del periodo comprendido desde el inicio del ejercicio"},
	}
	basisRESICOPMDeductions = []Citation{
		{Law: "LISR", Article: "208", Note: "deducciones autorizadas"},
		{Law: "LISR", Article: "211", Note: "deducciones efectivamente erogadas desde el inicio del ejercicio"},
	}
	basisRESICOPMPTU    = []Citation{{Law: "LISR", Article: "211", Note: "participación de los trabajadores en las utilidades pagada"}}
	basisRESICOPMLosses = []Citation{{Law: "LISR", Article: "211", Note: "pérdidas fiscales de ejercicios anteriores"}}
	basisRESICOPMBase   = []Citation{{Law: "LISR", Article: "211", Note: "utilidad fiscal para el pago provisional"}}
	basisRESICOPMRate   = []Citation{
		{Law: "LISR", Article: "9", Note: "tasa del 30 %"},
		{Law: "LISR", Article: "211", Note: "se aplica la tasa del artículo 9"},
	}
	basisRESICOPMPrior    = []Citation{{Law: "LISR", Article: "211", Note: "pagos provisionales efectuados con anterioridad"}}
	basisRESICOPMInterest = []Citation{
		{Law: "LISR", Article: "54", Note: "retención sobre intereses"},
		{Law: "LIF 2026", Article: "24", Note: "tasa anual de retención del 0.90 %"},
	}
	basisRESICOPMPayable = []Citation{{Law: "LISR", Article: "211", Note: "pago provisional mensual"}}
)

// RESICOPMStrategy is regime 626 for personas morales (LISR 206 to 215;
// research B.2.2): on a cash basis from January to the period, the income
// actually collected minus the deductions actually paid (plus what the year
// held before the app), minus PTU paid and prior-year losses, is the base;
// a base below zero is no profit. Base × 30 % minus the prior provisional
// payments and the art. 54 interest withholding (an open CPA question) is
// the month's payment. IVA follows the general rules through ComputeIVA.
type RESICOPMStrategy struct{}

// NewRESICOPMStrategy returns the RESICO PM strategy.
func NewRESICOPMStrategy() RESICOPMStrategy { return RESICOPMStrategy{} }

// Key is 626/moral.
func (RESICOPMStrategy) Key() StrategyKey {
	return StrategyKey{Regime: fiscal.RegimeRESICO, PersonType: fiscal.PersonMoral}
}

// Treatment is full: ISR and IVA are estimated.
func (RESICOPMStrategy) Treatment() Treatment { return TreatmentFull }

// RequiredInputs is the prior provisional payment of every month before the
// period in which the regime was active (MODULE-tax-engine §2.8); nil in
// January or in the regime's first active month.
func (RESICOPMStrategy) RequiredInputs(_ Options, p Period, f Facts) []RequiredInput {
	var months []time.Month
	for m := time.January; m < p.Month(); m++ {
		if slices.Contains(f.ActiveMonths, m) {
			months = append(months, m)
		}
	}
	if len(months) == 0 {
		return nil
	}
	return []RequiredInput{{Key: InputPriorProvisionalPayments, Kind: InputMoneyMonthly, Months: months}}
}

// Estimate computes the month's provisional ISR and IVA components. Past
// the annual cap (35,000,000 collected, 206) the year is still computed and
// an info line says the entity leaves RESICO from the next fiscal year.
func (s RESICOPMStrategy) Estimate(in Input, p *taxparams.Params) (RegimeEstimate, error) {
	est, err := s.estimate(in, p)
	if err != nil {
		return RegimeEstimate{}, fmt.Errorf("%s %s: %w", s.Key(), in.Period, err)
	}
	return est, nil
}

func (s RESICOPMStrategy) estimate(in Input, p *taxparams.Params) (RegimeEstimate, error) {
	n, date := in.Period.Month(), in.Period.End()
	ytd, err := in.Sums.Sum(time.January, n)
	if err != nil {
		return RegimeEstimate{}, err
	}
	rate, err := p.Rate(paramPersonaMoralRate, date)
	if err != nil {
		return RegimeEstimate{}, err
	}
	incomeCap, err := p.Money(paramRESICOPMCap, date)
	if err != nil {
		return RegimeEstimate{}, err
	}
	prior, err := in.Inputs.SumBefore(InputPriorProvisionalPayments, n)
	if err != nil {
		return RegimeEstimate{}, err
	}

	t := &resicoTrace{l: NewLedger()}
	before := resicoPMBeforeApp(in.Inputs)
	collected, err := ytd.Income.Cash.Add(before.income)
	if err != nil {
		return RegimeEstimate{}, err
	}
	if collected > incomeCap {
		t.info(resicoISRStep(LineRESICOPMCapExceeded, SourceRecords(), basisRESICOPMCap), collected)
	}
	t.resicoPMBase(ytd, before)
	t.l.Scale(resicoISRStep(LineRESICOPMRate, SourceParams(paramPersonaMoralRate), basisRESICOPMRate), rate)
	if n > time.January {
		t.l.Subtract(resicoISRStep(LineRESICOPMPriorPayments, SourceTaxInput(InputPriorProvisionalPayments), basisRESICOPMPrior), t.exact(prior))
	}
	est := RegimeEstimate{Key: s.Key(), Coverage: TreatmentFull}
	if before.interest != 0 {
		step := resicoISRStep(LineRESICOPMInterestWithholding, SourceTaxInput(InputInterestWithholding), basisRESICOPMInterest)
		step.NeedsCPAReview = true
		t.l.Subtract(step, t.exact(before.interest))
		est.CPATopics = []CPATopic{TopicRESICOPMInterestWithholding}
	}
	isr := NewTaxResult(t.l.Result(resicoISRStep(LineRESICOPMPayable, SourceComputed(), basisRESICOPMPayable)))
	est.ISR = &isr
	if ytd.Expenses.DeductiblePaid > 0 || before.deductions > 0 {
		est.Disclaimers = []DisclaimerCode{DisclaimerDeductibilityClaimed}
	}
	return t.finish(est, IVAInput{Sums: in.Sums.Month(n), Options: in.Options})
}

// resicoPMInputs are the optional money inputs of 626 PM (§2.8); an absent
// one is zero.
type resicoPMInputs struct {
	income, deductions, ptu, losses, interest money.Cents
}

func resicoPMBeforeApp(in TaxInputs) resicoPMInputs {
	value := func(k InputKey) money.Cents {
		v, _ := in.Money(k) // absent: zero
		return v
	}
	return resicoPMInputs{
		income: value(InputYTDBeforeAppIncome), deductions: value(InputYTDBeforeAppDeductions),
		ptu: value(InputPTUPaid), losses: value(InputPriorYearLosses), interest: value(InputInterestWithholding),
	}
}

// resicoPMBase traces the cash-flow profit: YTD income − YTD deductions −
// PTU − losses, and a zero base after a loss (no negative tax).
func (t *resicoTrace) resicoPMBase(ytd MonthSums, before resicoPMInputs) {
	t.l.Input(resicoISRStep(LineRESICOPMIncomeYTD, SourceRecords(), basisRESICOPMIncome), t.exact(ytd.Income.Cash))
	t.optional(t.l.Add, LineRESICOPMIncomeBeforeApp, InputYTDBeforeAppIncome, basisRESICOPMIncome, before.income)
	t.l.Subtract(resicoISRStep(LineRESICOPMDeductionsYTD, SourceRecords(), basisRESICOPMDeductions), t.exact(ytd.Expenses.DeductiblePaid))
	t.optional(t.l.Subtract, LineRESICOPMDeductionsBeforeApp, InputYTDBeforeAppDeductions, basisRESICOPMDeductions, before.deductions)
	t.optional(t.l.Subtract, LineRESICOPMPTUPaid, InputPTUPaid, basisRESICOPMPTU, before.ptu)
	t.optional(t.l.Subtract, LineRESICOPMPriorYearLosses, InputPriorYearLosses, basisRESICOPMLosses, before.losses)
	if base := t.l.Result(resicoISRStep(LineRESICOPMBase, SourceComputed(), basisRESICOPMBase)); base < 0 {
		t.l.Input(resicoISRStep(LineRESICOPMNoProfit, SourceComputed(), basisRESICOPMBase), 0)
	}
}

// optional traces a tax input with op (the Ledger's Add or Subtract), or
// nothing when it is zero.
func (t *resicoTrace) optional(op func(Step, money.Exact), code LineCode, key InputKey, basis []Citation, amount money.Cents) {
	if amount != 0 {
		op(resicoISRStep(code, SourceTaxInput(key), basis), t.exact(amount))
	}
}
