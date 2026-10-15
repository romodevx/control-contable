package domain

import (
	"fmt"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain/taxparams"
)

// The 606 PF regime options (MODULE-taxpayers §2.3, ADR-0002 §2).
const (
	// OptionRentDeductionMode picks the deductions: "blind_35" (35 % of the
	// rent plus the predial, LISR 115 ¶2; the catalog default) or "actual"
	// (the deductible expenses, LISR 115 fr. I–VI).
	OptionRentDeductionMode OptionKey = "deduction_mode"
	// OptionRentQuarterlyPayment is the quarterly option of LISR 116 ¶3
	// (catalog migration 20261014160000). Its table is not transcribed, so a
	// quarterly payer is informational (ADR-0002 §2).
	OptionRentQuarterlyPayment OptionKey = "pays_quarterly"
)

// ErrRentDeductionMode reports a 606 deduction_mode outside the catalog's
// values: the taxpayers module validates options, so it is a wiring bug.
const ErrRentDeductionMode Error = "taxengine: invalid 606 deduction_mode"

const (
	rentModeBlind  = "blind_35"
	rentModeActual = "actual"

	rentParamBlind  = taxparams.Key("isr.rent.blind_deduction")
	rentParamTariff = taxparams.Key("isr.art116.monthly")
)

// The ISR line codes of the 606 PF strategy, in order; a mode shows either
// the blind deduction and the predial or the actual deductions, and a base
// at or below zero shows rentLineNoBase instead of the tariff.
const (
	rentLineRent      LineCode = "isr.rent_month"
	rentLineBlind     LineCode = "isr.blind_deduction"
	rentLinePredial   LineCode = "isr.predial"
	rentLineActual    LineCode = "isr.actual_deductions"
	rentLineBase      LineCode = "isr.base"
	rentLineNoBase    LineCode = "isr.no_base"
	rentLineTariffRow LineCode = "isr.tariff_row"
	rentLineExcess    LineCode = "isr.tariff_excess"
	rentLineCaused    LineCode = "isr.caused"
	rentLineWithheld  LineCode = "isr.withheld_by_moral"
	rentLinePayable   LineCode = "isr.payable"
)

// The legal basis of each 606 ISR step (MODULE-tax-engine §2.5; research
// A.2.4).
var (
	rentBasisIncome = []Citation{
		{Law: "LISR", Article: "114", Note: "ingresos por arrendamiento"},
		{Law: "LISR", Article: "116", Note: "párrafo segundo: ingresos del mes, no acumulados"},
	}
	rentBasisBlind   = []Citation{{Law: "LISR", Article: "115", Note: "párrafo segundo: deducción opcional del 35 %"}}
	rentBasisPredial = []Citation{{Law: "LISR", Article: "115", Note: "párrafo segundo: más el impuesto predial"}}
	rentBasisActual  = []Citation{{Law: "LISR", Article: "115", Note: "fracciones I a VI"}}
	rentBasisBase    = []Citation{{Law: "LISR", Article: "116", Note: "párrafo segundo"}}
	rentBasisNoBase  = []Citation{{Law: "LISR", Article: "116", Note: "párrafo segundo: sin base, no hay impuesto"}}
	rentBasisTariff  = []Citation{
		{Law: "LISR", Article: "116", Note: "párrafo segundo: tarifa mensual"},
		{Law: "Anexo 8 RMF 2026", Article: "Tarifa art. 116", Note: "pagos provisionales mensuales"},
	}
	rentBasisWithheld = []Citation{{Law: "LISR", Article: "116", Note: "párrafo quinto: retención del 10 % por personas morales"}}
	rentBasisPayable  = []Citation{{Law: "LISR", Article: "116", Note: "párrafos segundo y quinto"}}
)

// RentPFStrategy is regime 606, Arrendamiento, persona física (LISR 114–118;
// research A.2.4; MODULE-tax-engine §2.4). Its ISR is monthly, never year to
// date: the rent collected in the month, minus the blind deduction (35 %)
// and the predial paid, or minus the deductible expenses paid; the art. 116
// monthly tariff (Anexo 8, the art. 96 rows) on that base; minus the 10 %
// withheld by personas morales, which can leave a saldo a favor. Its IVA is
// the shared monthly components: residential rent is exempt and carries no
// IVA on its records, commercial rent is taxed at 16 %.
type RentPFStrategy struct{}

// NewRentPFStrategy returns the 606 PF strategy.
func NewRentPFStrategy() RentPFStrategy { return RentPFStrategy{} }

// Key is 606 persona física.
func (RentPFStrategy) Key() StrategyKey {
	return StrategyKey{Regime: fiscal.RegimeRent, PersonType: fiscal.PersonFisica}
}

// Treatment is full.
func (RentPFStrategy) Treatment() Treatment { return TreatmentFull }

// RequiredInputs is empty: rent, deductions, predial and withholdings are
// all records (ADR-0002 §2).
func (RentPFStrategy) RequiredInputs(Options, Period, Facts) []RequiredInput { return nil }

// ParamRequirements are the parameters Estimate reads, for the composition
// root's boot check (Params.Require).
func (RentPFStrategy) ParamRequirements() []taxparams.Requirement {
	return []taxparams.Requirement{taxparams.RequireRate(rentParamBlind), taxparams.RequireTable(rentParamTariff)}
}

// Estimate computes the period's month. A quarterly payer gets the
// informational estimate.
func (s RentPFStrategy) Estimate(in Input, p *taxparams.Params) (RegimeEstimate, error) {
	if quarterly, _ := in.Options.Bool(OptionRentQuarterlyPayment); quarterly {
		return NewInformationalStrategy(s.Key()).Estimate(in, p)
	}
	blind, err := rentBlindMode(in.Options)
	if err != nil {
		return RegimeEstimate{}, err
	}
	month := in.Sums.Month(in.Period.Month())
	l := NewLedger()
	isr, err := rentISR(l, rentFacts{period: in.Period, month: month, blind: blind}, p)
	if err != nil {
		return RegimeEstimate{}, fmt.Errorf("606 ISR: %w", err)
	}
	iva, err := ComputeIVA(l, monthIVA(in))
	if err != nil {
		return RegimeEstimate{}, fmt.Errorf("606: %w", err)
	}
	lines, err := l.Lines()
	if err != nil {
		return RegimeEstimate{}, fmt.Errorf("606: %w", err)
	}
	return RegimeEstimate{
		Key: s.Key(), Coverage: TreatmentFull, ISR: &isr, IVA: &iva.Components,
		WithheldToRemit: iva.WithheldToRemit, Lines: lines, CPATopics: iva.CPATopics,
		Disclaimers: append(rentDisclaimers(month.Expenses, blind), iva.Disclaimers...),
	}, nil
}

// rentBlindMode reads deduction_mode: absent is the catalog default.
func rentBlindMode(o Options) (bool, error) {
	mode, set := o.Text(OptionRentDeductionMode)
	switch {
	case !set || mode == rentModeBlind:
		return true, nil
	case mode == rentModeActual:
		return false, nil
	}
	return false, fmt.Errorf("%q: %w", mode, ErrRentDeductionMode)
}

// rentDisclaimers adds deductibility_claimed when an expense entered the
// base: the predial under the blind deduction, every deductible expense
// under the actual one.
func rentDisclaimers(e ExpenseSums, blind bool) []DisclaimerCode {
	if (blind && e.PredialPaid > 0) || (!blind && e.DeductiblePaid > 0) {
		return []DisclaimerCode{DisclaimerDeductibilityClaimed}
	}
	return nil
}

// rentFacts are what the ISR reads.
type rentFacts struct {
	period Period
	month  MonthSums
	blind  bool
}

// rentTrace appends the ISR steps, converting centavos to Exact with a
// sticky error; the Ledger keeps its own.
type rentTrace struct {
	l   *Ledger
	err error
}

func (t *rentTrace) exact(c money.Cents) money.Exact {
	if t.err != nil {
		return 0
	}
	e, err := c.Exact()
	t.err = err
	return e
}

func rentStep(code LineCode, source Source, basis []Citation, args ...string) Step {
	return Step{Tax: TaxISR, Code: code, Label: LabelKey(code), Args: args, Source: source, Basis: basis}
}

// rentISR traces rent − deductions → tariff − PM withholding and returns
// the payable result (negative: saldo a favor).
func rentISR(l *Ledger, f rentFacts, p *taxparams.Params) (TaxResult, error) {
	table, err := p.Table(rentParamTariff, f.period.End())
	if err != nil {
		return TaxResult{}, err
	}
	t := rentTrace{l: l}
	t.l.Input(rentStep(rentLineRent, SourceRecords(), rentBasisIncome, f.period.String()), t.exact(f.month.Income.Cash))
	if err := t.deductions(f, p); err != nil {
		return TaxResult{}, err
	}
	t.tax(table, t.l.Result(rentStep(rentLineBase, SourceComputed(), rentBasisBase)))
	t.l.Subtract(rentStep(rentLineWithheld, SourceRecords(), rentBasisWithheld), t.exact(f.month.Income.ISRWithheld.ByMoral))
	payable := t.l.Result(rentStep(rentLinePayable, SourceComputed(), rentBasisPayable))
	if t.err == nil {
		t.err = l.err
	}
	if t.err != nil {
		return TaxResult{}, t.err
	}
	return NewTaxResult(payable), nil
}

// deductions subtracts the blind deduction and the predial, or the actual
// deductions.
func (t *rentTrace) deductions(f rentFacts, p *taxparams.Params) error {
	expenses := f.month.Expenses
	if !f.blind {
		t.l.Subtract(rentStep(rentLineActual, SourceRecords(), rentBasisActual), t.exact(expenses.DeductiblePaid))
		return nil
	}
	rate, err := p.Rate(rentParamBlind, f.period.End())
	if err != nil {
		return err
	}
	blind, err := f.month.Income.Cash.MulRate(rate)
	if err != nil {
		return err
	}
	t.l.Subtract(rentStep(rentLineBlind, SourceParams(rentParamBlind), rentBasisBlind, rate.String()), blind)
	t.l.Subtract(rentStep(rentLinePredial, SourceRecords(), rentBasisPredial), t.exact(expenses.PredialPaid))
	return nil
}

// tax applies the tariff to a positive base, and no tax to any other: the
// caused ISR starts a new running value either way.
func (t *rentTrace) tax(table taxparams.Table, base money.Exact) {
	if base <= 0 {
		t.l.Input(rentStep(rentLineNoBase, SourceComputed(), rentBasisNoBase), 0)
	} else {
		b := t.l.Lookup(rentStep(rentLineTariffRow, SourceParams(rentParamTariff), rentBasisTariff), table, base)
		t.l.AddProduct(rentStep(rentLineExcess, SourceParams(rentParamTariff), rentBasisTariff), b.Excess(base), b.Row.Rate)
	}
	t.l.ResultRounded(rentStep(rentLineCaused, SourceComputed(), rentBasisBase))
}
