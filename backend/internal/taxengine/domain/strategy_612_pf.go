package domain

import (
	"fmt"
	"time"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain/taxparams"
)

// Strategy612PF computes regime 612 for a persona física, Actividades
// Empresariales y Profesionales (LISR arts. 100-110; research A.2.3,
// MODULE-tax-engine §2.4):
//
//	base   = YTD cash income (+ ytd_before_app_income)
//	         − YTD deductions paid (+ ytd_before_app_deductions)
//	         − ptu_paid − prior_year_losses
//	caused = accumulated art. 106 tariff of month n on the base; 0 if base ≤ 0
//	ISR    = caused − prior provisional payments (months < n)
//	         − YTD ISR withheld by personas morales
//
// A negative ISR is in favor: no payment, credited later (research A.2.3
// step 8). The tariff is the stored table of month n, never derived. IVA is
// the month's, through the shared components (ComputeIVA). A missing input
// computes as zero; the registry marks the estimate incomplete (ADR-0002 §2).
type Strategy612PF struct{}

// NewStrategy612PF returns the 612 PF strategy.
func NewStrategy612PF() Strategy612PF { return Strategy612PF{} }

// pf612Tariff is the parameter key of the twelve accumulated tables.
const pf612Tariff taxparams.Key = "isr.art106.accumulated" // gitleaks:allow (a public parameter key, not a secret)

// The 612 ISR line codes, in the order Estimate appends them; the optional
// tax-input lines appear only when the input was entered.
const (
	pf612IncomeYTD           LineCode = "isr.income_ytd"
	pf612IncomeBeforeApp     LineCode = "isr.income_before_app"
	pf612DeductionsYTD       LineCode = "isr.deductions_ytd"
	pf612DeductionsBeforeApp LineCode = "isr.deductions_before_app"
	pf612PTUPaid             LineCode = "isr.ptu_paid"
	pf612PriorYearLosses     LineCode = "isr.prior_year_losses"
	pf612Base                LineCode = "isr.base"
	pf612TariffRow           LineCode = "isr.tariff_row"
	pf612TariffExcess        LineCode = "isr.tariff_excess"
	pf612Caused              LineCode = "isr.caused"
	pf612PriorPayments       LineCode = "isr.prior_payments"
	pf612WithheldByMoral     LineCode = "isr.withheld_by_moral"
	pf612Payable             LineCode = "isr.payable"
)

// The legal basis of each 612 ISR step (MODULE-tax-engine §2.5).
var (
	pf612BasisIncome = []Citation{
		{Law: "LISR", Article: "102", Note: "ingresos efectivamente cobrados"},
		{Law: "LISR", Article: "106", Note: "párrafo primero, desde el inicio del ejercicio"},
	}
	pf612BasisDeductions = []Citation{
		{Law: "LISR", Article: "103", Note: "deducciones autorizadas"},
		{Law: "LISR", Article: "105", Note: "efectivamente erogadas"},
		{Law: "LISR", Article: "106", Note: "párrafo primero"},
	}
	pf612BasisPTU    = []Citation{{Law: "LISR", Article: "106", Note: "párrafo primero, PTU pagada en el ejercicio"}}
	pf612BasisLosses = []Citation{{Law: "LISR", Article: "106", Note: "párrafo primero, pérdidas fiscales de ejercicios anteriores"}}
	pf612BasisBase   = []Citation{{Law: "LISR", Article: "106", Note: "párrafo primero"}}
	pf612BasisTariff = []Citation{
		{Law: "LISR", Article: "106", Note: "párrafos segundo y tercero, tarifa acumulada del mes"},
		{Law: "Anexo 8 RMF 2026", Article: "B.VI", Note: "tarifa del mes del pago"},
	}
	pf612BasisCaused   = []Citation{{Law: "LISR", Article: "106", Note: "párrafos primero a tercero"}}
	pf612BasisPrior    = []Citation{{Law: "LISR", Article: "106", Note: "párrafo cuarto, pagos provisionales anteriores"}}
	pf612BasisWithheld = []Citation{{Law: "LISR", Article: "106", Note: "párrafo quinto, retención del 10 % por personas morales"}}
	pf612BasisPayable  = []Citation{{Law: "LISR", Article: "106", Note: "pago provisional mensual"}}
)

// Key is 612 for a persona física.
func (Strategy612PF) Key() StrategyKey {
	return StrategyKey{Regime: fiscal.RegimeBusiness, PersonType: fiscal.PersonFisica}
}

// Treatment is full: ISR and IVA are both estimated.
func (Strategy612PF) Treatment() Treatment { return TreatmentFull }

// RequiredInputs is the prior provisional payment of every month before n
// in which the regime was active (MODULE-tax-engine §2.8): none in January
// or in the first month held. The other 612 inputs are optional.
func (Strategy612PF) RequiredInputs(_ Options, p Period, f Facts) []RequiredInput {
	var months []time.Month
	for _, m := range f.ActiveMonths {
		if m < p.Month() {
			months = append(months, m)
		}
	}
	if len(months) == 0 {
		return nil
	}
	return []RequiredInput{{Key: InputPriorProvisionalPayments, Kind: InputMoneyMonthly, Months: months}}
}

// Estimate computes the period's ISR and IVA components through one Ledger.
func (s Strategy612PF) Estimate(in Input, p *taxparams.Params) (RegimeEstimate, error) {
	l := NewLedger()
	isr, claimed, err := pf612ISR(l, in, p)
	if err != nil {
		return RegimeEstimate{}, fmt.Errorf("612 ISR: %w", err)
	}
	iva, err := ComputeIVA(l, IVAInput{Sums: in.Sums.Month(in.Period.Month()), Options: in.Options})
	if err != nil {
		return RegimeEstimate{}, fmt.Errorf("612: %w", err)
	}
	lines, err := l.Lines()
	if err != nil {
		return RegimeEstimate{}, fmt.Errorf("612: %w", err)
	}
	var disclaimers []DisclaimerCode
	if claimed {
		disclaimers = append(disclaimers, DisclaimerDeductibilityClaimed)
	}
	return RegimeEstimate{
		Key:             s.Key(),
		Coverage:        TreatmentFull,
		ISR:             &isr,
		IVA:             &iva.Components,
		WithheldToRemit: iva.WithheldToRemit,
		Lines:           lines,
		CPATopics:       iva.CPATopics,
		Disclaimers:     append(disclaimers, iva.Disclaimers...),
	}, nil
}

// pf612ISR traces the ISR steps and returns the result, and whether any
// deduction was claimed (the deductibility_claimed disclaimer).
func pf612ISR(l *Ledger, in Input, p *taxparams.Params) (TaxResult, bool, error) {
	n := in.Period.Month()
	ytd, err := in.Sums.Sum(time.January, n)
	if err != nil {
		return TaxResult{}, false, err
	}
	table, err := p.MonthlyTable(pf612Tariff, n, in.Period.End())
	if err != nil {
		return TaxResult{}, false, err
	}
	prior, err := in.Inputs.SumBefore(InputPriorProvisionalPayments, n)
	if err != nil {
		return TaxResult{}, false, err
	}

	t := pf612Trace{l: l, ytdArgs: []string{pf612MonthArg(in.Period, time.January), in.Period.String()}}
	base, claimed := t.base(ytd, in.Inputs)
	t.caused(table, base, in.Period.String())
	t.l.Subtract(pf612Step(pf612PriorPayments, SourceTaxInput(InputPriorProvisionalPayments), pf612BasisPrior), t.exact(prior))
	t.l.Subtract(pf612Step(pf612WithheldByMoral, SourceRecords(), pf612BasisWithheld), t.exact(ytd.Income.ISRWithheld.ByMoral))
	payable := t.l.Result(pf612Step(pf612Payable, SourceComputed(), pf612BasisPayable))
	if t.err != nil {
		return TaxResult{}, false, t.err
	}
	// A ledger error is sticky: ComputeIVA and Lines report it.
	return NewTaxResult(payable), claimed, nil
}

// pf612MonthArg renders month m of the period's year as "YYYY-MM".
func pf612MonthArg(p Period, m time.Month) string { return fmt.Sprintf("%04d-%02d", p.Year(), int(m)) }

func pf612Step(code LineCode, source Source, basis []Citation) Step {
	return Step{Tax: TaxISR, Code: code, Label: LabelKey(code), Source: source, Basis: basis}
}

// pf612Trace appends the 612 ISR steps, converting centavos to Exact with
// a sticky error; the Ledger keeps its own.
type pf612Trace struct {
	l       *Ledger
	ytdArgs []string // the YTD range, "YYYY-01".."YYYY-MM"
	err     error
}

func (t *pf612Trace) exact(c money.Cents) money.Exact {
	if t.err != nil {
		return 0
	}
	e, err := c.Exact()
	t.err = err
	return e
}

// optional subtracts (or, with add, adds) an entered money input; an input
// never entered has no line.
func (t *pf612Trace) optional(code LineCode, key InputKey, basis []Citation, in TaxInputs, add bool) money.Cents {
	v, ok := in.Money(key)
	if !ok {
		return 0
	}
	s := pf612Step(code, SourceTaxInput(key), basis)
	if add {
		t.l.Add(s, t.exact(v))
		return v
	}
	t.l.Subtract(s, t.exact(v))
	return v
}

// base traces income − deductions − PTU − losses and reports whether a
// deduction was claimed.
func (t *pf612Trace) base(ytd MonthSums, in TaxInputs) (money.Exact, bool) {
	income := pf612Step(pf612IncomeYTD, SourceRecords(), pf612BasisIncome)
	income.Args = t.ytdArgs
	t.l.Input(income, t.exact(ytd.Income.Cash))
	t.optional(pf612IncomeBeforeApp, InputYTDBeforeAppIncome, pf612BasisIncome, in, true)

	deductions := pf612Step(pf612DeductionsYTD, SourceRecords(), pf612BasisDeductions)
	deductions.Args = t.ytdArgs
	t.l.Subtract(deductions, t.exact(ytd.Expenses.DeductiblePaid))
	before := t.optional(pf612DeductionsBeforeApp, InputYTDBeforeAppDeductions, pf612BasisDeductions, in, false)

	t.optional(pf612PTUPaid, InputPTUPaid, pf612BasisPTU, in, false)
	t.optional(pf612PriorYearLosses, InputPriorYearLosses, pf612BasisLosses, in, false)
	base := t.l.Result(pf612Step(pf612Base, SourceComputed(), pf612BasisBase))
	return base, ytd.Expenses.DeductiblePaid > 0 || before > 0
}

// caused traces the ISR caused: the tariff of month n on a positive base,
// else zero with no lookup (LISR 106 ¶1).
func (t *pf612Trace) caused(table taxparams.Table, base money.Exact, month string) {
	if base <= 0 {
		t.l.Input(pf612Step(pf612Caused, SourceComputed(), pf612BasisCaused), 0)
		return
	}
	row := pf612Step(pf612TariffRow, SourceParams(pf612Tariff), pf612BasisTariff)
	row.Args = []string{month}
	b := t.l.Lookup(row, table, base)
	t.l.AddProduct(pf612Step(pf612TariffExcess, SourceParams(pf612Tariff), pf612BasisTariff), b.Excess(base), b.Row.Rate)
	t.l.Result(pf612Step(pf612Caused, SourceComputed(), pf612BasisCaused))
}
