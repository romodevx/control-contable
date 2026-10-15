package domain

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain/taxparams"
)

// The parameters RESICO PF reads (params/data/<year>).
const (
	paramRESICOPFMonthly taxparams.Key = "isr.resico_pf.monthly"
	paramRESICOPFCap     taxparams.Key = "regime.resico_pf.income_cap"
)

// The line codes of the RESICO PF ISR, in order; the withholding line only
// when something was withheld, the excess line only when the withholding is
// more than the month's ISR. The annual cap is a disclaimer
// (DisclaimerRESICOCapExceeded), never a line: lines carry money.
const (
	LineRESICOPFIncomeMonth     LineCode = "isr.income_month"
	LineRESICOPFRate            LineCode = "isr.resico_rate"
	LineRESICOPFWithheldByMoral LineCode = "isr.withheld_by_moral"
	// LineRESICOPFExcessToAnnual is the withholding above the month's ISR:
	// never refunded in the month, it is credited in the annual return.
	LineRESICOPFExcessToAnnual LineCode = "isr.resico_excess_to_annual"
	LineRESICOPFPayable        LineCode = "isr.payable"
)

// LabelRESICOPFAboveTable explains a partial RESICO PF estimate: the month's
// income is above the monthly table, which ends at the regime's cap, so
// there is no RESICO rate to apply.
const LabelRESICOPFAboveTable LabelKey = "resico_pf_month_above_table"

// The legal basis of each RESICO PF step (MODULE-tax-engine §2.5).
var (
	basisRESICOPFIncome = []Citation{{
		Law: "LISR", Article: "113-E", Note: "párrafo quinto: ingresos efectivamente cobrados, sin IVA y sin deducciones",
	}}
	basisRESICOPFRate = []Citation{{
		Law: "LISR", Article: "113-E", Note: "tabla mensual: la tasa aplica a la totalidad de los ingresos del mes",
	}}
	basisRESICOPFWithheld = []Citation{{
		Law: "LISR", Article: "113-J", Note: "retención del 1.25 % por personas morales",
	}}
	basisRESICOPFPayable = []Citation{{Law: "LISR", Article: "113-E", Note: "pago mensual"}}
	basisRESICOPFExcess  = []Citation{
		{Law: "LISR", Article: "113-F", Note: "la declaración anual disminuye las retenciones del 113-J"},
		{Law: "LISR", Article: "113-J", Note: "párrafo segundo"},
	}
)

// RESICOPFStrategy is regime 626 for personas físicas (LISR 113-E to 113-J;
// research A.2.5): each month on its own, the income collected in the
// month, without IVA and without deductions, times the rate of its bracket
// in the monthly table. The rate is not marginal: it applies to the whole
// income (PF-626-ISR-02: 25,000.00 at 1 %, 25,000.01 at 1.1 %). The ISR
// withheld by personas morales (1.25 %, 113-J), as recorded, is credited.
// IVA follows the general rules (research A.3) through ComputeIVA.
type RESICOPFStrategy struct{}

// NewRESICOPFStrategy returns the RESICO PF strategy.
func NewRESICOPFStrategy() RESICOPFStrategy { return RESICOPFStrategy{} }

// Key is 626/fisica.
func (RESICOPFStrategy) Key() StrategyKey {
	return StrategyKey{Regime: fiscal.RegimeRESICO, PersonType: fiscal.PersonFisica}
}

// Treatment is full: ISR and IVA are estimated.
func (RESICOPFStrategy) Treatment() Treatment { return TreatmentFull }

// RequiredInputs is empty: every monthly payment stands alone, so no prior
// payment or other input is ever needed.
func (RESICOPFStrategy) RequiredInputs(Options, Period, Facts) []RequiredInput { return nil }

// Estimate computes the month's ISR and IVA components. Past the annual
// cap (3,500,000 collected Jan..n, 113-E) the month is still computed and
// the estimate carries DisclaimerRESICOCapExceeded: the taxpayer leaves
// RESICO from the next month (113-E, 113-I). A month
// above the table itself has no RESICO rate: the ISR is not estimated and
// the estimate is partial (LabelRESICOPFAboveTable).
func (s RESICOPFStrategy) Estimate(in Input, p *taxparams.Params) (RegimeEstimate, error) {
	est, err := s.estimate(in, p)
	if err != nil {
		return RegimeEstimate{}, fmt.Errorf("%s %s: %w", s.Key(), in.Period, err)
	}
	return est, nil
}

func (s RESICOPFStrategy) estimate(in Input, p *taxparams.Params) (RegimeEstimate, error) {
	n, date := in.Period.Month(), in.Period.End()
	ytd, err := in.Sums.Sum(time.January, n)
	if err != nil {
		return RegimeEstimate{}, err
	}
	table, err := p.Table(paramRESICOPFMonthly, date)
	if err != nil {
		return RegimeEstimate{}, err
	}
	incomeCap, err := p.Money(paramRESICOPFCap, date)
	if err != nil {
		return RegimeEstimate{}, err
	}

	t := &resicoTrace{l: NewLedger()}
	month := in.Sums.Month(n)
	est := RegimeEstimate{Key: s.Key(), Coverage: TreatmentFull}
	if ytd.Income.Cash > incomeCap {
		est.Disclaimers = []DisclaimerCode{DisclaimerRESICOCapExceeded}
	}
	isr, err := t.resicoPFISR(table, month.Income)
	switch {
	case errors.Is(err, taxparams.ErrAboveTable):
		est.Coverage, est.Explanation = TreatmentPartial, LabelRESICOPFAboveTable
	case err != nil:
		return RegimeEstimate{}, err
	default:
		est.ISR = isr
	}
	return t.finish(est, monthIVA(in))
}

// resicoPFISR traces income × bracket rate − PM withholding = ISR. It
// fails with taxparams.ErrAboveTable, before any line, for an income above
// the monthly table.
func (t *resicoTrace) resicoPFISR(table taxparams.Table, income IncomeSums) (*TaxResult, error) {
	cash := t.exact(income.Cash)
	var bracket taxparams.Bracket
	if cash > 0 {
		var err error
		if bracket, err = table.Bracket(cash); err != nil {
			return nil, err
		}
	}
	t.l.Input(resicoISRStep(LineRESICOPFIncomeMonth, SourceRecords(), basisRESICOPFIncome), cash)
	if cash > 0 {
		rateStep := resicoISRStep(LineRESICOPFRate, SourceParams(paramRESICOPFMonthly), basisRESICOPFRate)
		rateStep.Args = []string{strconv.Itoa(bracket.Number)} // the table row, for the label
		t.l.Scale(rateStep, bracket.Row.Rate)
	}
	if withheld := income.ISRWithheld.ByMoral; withheld != 0 {
		t.l.Subtract(resicoISRStep(LineRESICOPFWithheldByMoral, SourceRecords(), basisRESICOPFWithheld), t.exact(withheld))
	}
	if over := t.l.Running(); over < 0 {
		// A4 ⚠️: the month floors at zero and the excess waits for the
		// annual return (MODULE-withholdings §4.1).
		excess := resicoISRStep(LineRESICOPFExcessToAnnual, SourceComputed(), basisRESICOPFExcess)
		excess.NeedsCPAReview = true
		t.l.Add(excess, -over)
	}
	result := NewTaxResult(t.l.Result(resicoISRStep(LineRESICOPFPayable, SourceComputed(), basisRESICOPFPayable)))
	return &result, t.failure()
}

// resicoTrace drives a RESICO (PF or PM) ledger: centavos convert to
// Exact with a sticky error, beside the Ledger's own.
type resicoTrace struct {
	l   *Ledger
	err error
}

func (t *resicoTrace) exact(c money.Cents) money.Exact {
	if t.err != nil {
		return 0
	}
	e, err := c.Exact()
	t.err = err
	return e
}

// failure is the first conversion or ledger error.
func (t *resicoTrace) failure() error {
	if t.err != nil {
		return t.err
	}
	return t.l.err
}

// finish adds the shared IVA components (§2.5, §2.6) and the lines.
func (t *resicoTrace) finish(est RegimeEstimate, iva IVAInput) (RegimeEstimate, error) {
	est, err := t.addIVA(est, iva)
	if err != nil {
		return RegimeEstimate{}, err
	}
	return t.addLines(est)
}

// addIVA traces the shared IVA components after the ISR; WithheldToRemit
// is the IVA withheld from suppliers (LIVA 1-A) until a PM adds its ISR.
func (t *resicoTrace) addIVA(est RegimeEstimate, iva IVAInput) (RegimeEstimate, error) {
	if err := t.failure(); err != nil {
		return RegimeEstimate{}, err
	}
	result, err := ComputeIVA(t.l, iva)
	if err != nil {
		return RegimeEstimate{}, err
	}
	est.IVA = &result.Components
	est.WithheldToRemit = result.WithheldToRemit
	est.Disclaimers = append(est.Disclaimers, result.Disclaimers...)
	est.CPATopics = append(est.CPATopics, result.CPATopics...)
	return est, nil
}

// addLines closes the trace into the estimate's breakdown.
func (t *resicoTrace) addLines(est RegimeEstimate) (RegimeEstimate, error) {
	lines, err := t.l.Lines()
	if err != nil {
		return RegimeEstimate{}, err
	}
	est.Lines = lines
	return est, nil
}

func resicoISRStep(code LineCode, source Source, basis []Citation) Step {
	return Step{Tax: TaxISR, Code: code, Label: LabelKey(code), Source: source, Basis: basis}
}
