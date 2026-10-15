package domain

import (
	"fmt"
	"strconv"
	"time"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain/taxparams"
)

// Names in this file carry a rif prefix: every regime strategy lives in
// package domain (PLAN PG-M5-S writes them in parallel), so a generic name
// would collide.

// OptionRIFYear is the 621 regime option "año de tributación en el RIF",
// "1".."10" (catalog option_defs, required).
const OptionRIFYear OptionKey = "rif_year"

// ErrInvalidRIFYear reports a 621 regime whose rif_year option is absent or
// outside 1-10: the catalog requires it, so this is a wiring defect.
const ErrInvalidRIFYear Error = "taxengine: rif_year option absent or outside 1-10"

// rifRegime is régimen 621, Incorporación Fiscal (closed to new taxpayers
// since 2022; Transitorio Segundo fr. IX, DOF 12-11-2021).
const rifRegime fiscal.RegimeCode = "621"

// rifTariff is the bimonthly RIF tariff (Anexo 8 B.VII), published identical
// to the February accumulated table of art. 106 (research A.2.7).
const rifTariff taxparams.Key = "isr.rif.bimonthly"

// rifReductionByYear is the ISR reduction for year 1..10 in RIF, 100 % down
// to 10 % (LISR 111 as in force until 2021). ⚠️ Not re-verified against the
// repealed text (CPA-REVIEW "RIF reduction by year"): every line using it
// is flagged and the estimate carries TopicRIFReduction.
var rifReductionByYear = [...]money.Rate{
	1_000_000, 900_000, 800_000, 700_000, 600_000, 500_000, 400_000, 300_000, 200_000, 100_000,
}

// The 621 line codes, in the order Estimate appends them.
const (
	rifLineBimesterPending LineCode = "isr.rif_bimester_pending"
	rifLineIncome          LineCode = "isr.income_bimester"
	rifLineDeductions      LineCode = "isr.deductions_bimester"
	rifLineBase            LineCode = "isr.base"
	rifLineTariffRow       LineCode = "isr.tariff_row"
	rifLineTariffExcess    LineCode = "isr.tariff_excess"
	rifLineCaused          LineCode = "isr.caused"
	rifLineReduction       LineCode = "isr.rif_reduction"
	rifLinePayable         LineCode = "isr.payable"
	rifLineIVABimester     LineCode = "iva.rif_bimester"
)

// The legal basis of each 621 step (MODULE-tax-engine §2.5: LISR 111 as in
// force until 31-12-2021, Transitorio Segundo fr. IX, Anexo 8 B.VII).
var (
	rifBasisLaw = []Citation{
		{Law: "LISR", Article: "111", Note: "texto vigente hasta el 31-12-2021"},
		{Law: "Decreto DOF 12-11-2021", Article: "Transitorio Segundo", Note: "fracción IX"},
	}
	rifBasisBase   = []Citation{{Law: "LISR", Article: "111", Note: "texto vigente hasta el 31-12-2021, ingresos menos deducciones del bimestre"}}
	rifBasisTariff = []Citation{
		{Law: "LISR", Article: "111", Note: "texto vigente hasta el 31-12-2021, tarifa bimestral"},
		{Law: "Anexo 8 RMF 2026", Article: "B.VII", Note: "tarifa bimestral del RIF"},
	}
	rifBasisReduction = []Citation{{Law: "LISR", Article: "111", Note: "texto vigente hasta el 31-12-2021, reducción por año en el RIF"}}
	rifBasisPayable   = []Citation{{Law: "LISR", Article: "111", Note: "texto vigente hasta el 31-12-2021, pago bimestral"}}
	rifBasisIVA       = []Citation{{Law: "LIVA", Article: "5-E", Note: "texto vigente hasta el 31-12-2021, pago bimestral"}}
)

// Strategy621PF is régimen 621, Incorporación Fiscal, for a persona física
// (research A.2.7, MODULE-tax-engine §2.4). It is partial: the RIF pays by
// bimester, so only the bimester's second month (an even month) has a
// figure:
//
//	base   = cash income − deductions paid, months n−1 and n
//	caused = bimonthly tariff (Anexo 8 B.VII) on the base; 0 if base ≤ 0
//	ISR    = caused × (100 % − reduction for rif_year)  ⚠️
//
// The ISR is rounded once, at the end (ADR-0002 §4); the research halves
// the rounded caused ISR instead, which can differ by a centavo
// (PF-621-01: 2,383.65 here, 2,383.66 there; CPA-REVIEW). The IVA is the
// bimester's (OQ-T7 ⚠️). The bimester's first month has no figure, only a
// line naming the month that has it. PTU is not deducted: the catalog
// defines no 621 tax input for it.
type Strategy621PF struct{}

// NewStrategy621PF returns the 621 PF strategy.
func NewStrategy621PF() Strategy621PF { return Strategy621PF{} }

// Key is 621 for a persona física.
func (Strategy621PF) Key() StrategyKey {
	return StrategyKey{Regime: rifRegime, PersonType: fiscal.PersonFisica}
}

// Treatment is partial: an unverified reduction and a bimonthly payment.
func (Strategy621PF) Treatment() Treatment { return TreatmentPartial }

// RequiredInputs is empty: the year in RIF is a regime option, and the
// bimonthly payment is definitive (no prior payments to credit).
func (Strategy621PF) RequiredInputs(Options, Period, Facts) []RequiredInput { return nil }

// Estimate computes the bimester's ISR and IVA in its second month, and
// only the pointer to that month in its first.
func (s Strategy621PF) Estimate(in Input, p *taxparams.Params) (RegimeEstimate, error) {
	year, reduction, err := rifYearOf(in.Options)
	if err != nil {
		return RegimeEstimate{}, fmt.Errorf("621 %s: %w", in.Period, err)
	}
	est := RegimeEstimate{Key: s.Key(), Coverage: TreatmentPartial, Explanation: RegimeExplanation(s.Key())}
	l := NewLedger()
	n := in.Period.Month()
	if n%2 == 1 {
		pending := rifStep(TaxISR, rifLineBimesterPending, SourceComputed(), rifBasisLaw)
		pending.Args = []string{rifMonthArg(in.Period, n+1)}
		l.Info(pending, 0)
		est.Lines, err = l.Lines()
		return est, rifWrap(in.Period, err)
	}

	bimester, err := in.Sums.Sum(n-1, n)
	if err != nil {
		return RegimeEstimate{}, rifWrap(in.Period, err)
	}
	table, err := p.Table(rifTariff, in.Period.End())
	if err != nil {
		return RegimeEstimate{}, rifWrap(in.Period, err)
	}
	t := rifTrace{l: l, bimester: []string{rifMonthArg(in.Period, n-1), in.Period.String()}}
	isr := t.isr(bimester, table, year, reduction)
	if t.err != nil {
		return RegimeEstimate{}, rifWrap(in.Period, t.err)
	}

	ivaNote := rifStep(TaxIVA, rifLineIVABimester, SourceComputed(), rifBasisIVA)
	ivaNote.Args, ivaNote.NeedsCPAReview = t.bimester, true
	l.Info(ivaNote, 0)
	// No WithheldPriorMonth: only a persona moral withholds IVA from its
	// suppliers (LIVA 1-A fr. II), so a RIF persona física has no withheld
	// part to credit (LIVA 5 fr. IV).
	iva, err := ComputeIVA(l, IVAInput{Sums: bimester, Options: in.Options})
	if err != nil {
		return RegimeEstimate{}, rifWrap(in.Period, err)
	}
	if est.Lines, err = l.Lines(); err != nil {
		return RegimeEstimate{}, rifWrap(in.Period, err)
	}

	est.ISR, est.IVA, est.WithheldToRemit = &isr, &iva.Components, iva.WithheldToRemit
	est.CPATopics = append([]CPATopic{TopicRIFReduction, TopicRIFBimonthlyIVA}, iva.CPATopics...)
	if bimester.Expenses.DeductiblePaid > 0 {
		est.Disclaimers = append(est.Disclaimers, DisclaimerDeductibilityClaimed)
	}
	est.Disclaimers = append(est.Disclaimers, iva.Disclaimers...)
	return est, nil
}

func rifWrap(p Period, err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("621 %s: %w", p, err)
}

// rifYearOf reads rif_year and returns it with its reduction.
func rifYearOf(o Options) (string, money.Rate, error) {
	raw, ok := o.Text(OptionRIFYear)
	if !ok {
		return "", 0, ErrInvalidRIFYear
	}
	year, err := strconv.Atoi(raw)
	if err != nil || year < 1 || year > len(rifReductionByYear) {
		return "", 0, fmt.Errorf("%q: %w", raw, ErrInvalidRIFYear)
	}
	return raw, rifReductionByYear[year-1], nil
}

// rifMonthArg renders month m of the period's year as "YYYY-MM".
func rifMonthArg(p Period, m time.Month) string { return fmt.Sprintf("%04d-%02d", p.Year(), int(m)) }

func rifStep(tax TaxKind, code LineCode, source Source, basis []Citation) Step {
	return Step{Tax: tax, Code: code, Label: LabelKey(code), Source: source, Basis: basis}
}

// rifTrace appends the 621 ISR steps, converting centavos to Exact with a
// sticky error; the Ledger keeps its own.
type rifTrace struct {
	l        *Ledger
	bimester []string // "YYYY-MM" of months n−1 and n
	err      error
}

func (t *rifTrace) exact(c money.Cents) money.Exact {
	if t.err != nil {
		return 0
	}
	e, err := c.Exact()
	t.err = err
	return e
}

// isr traces base → caused → reduced ISR and returns the result.
func (t *rifTrace) isr(bimester MonthSums, table taxparams.Table, year string, reduction money.Rate) TaxResult {
	income := rifStep(TaxISR, rifLineIncome, SourceRecords(), rifBasisBase)
	income.Args = t.bimester
	t.l.Input(income, t.exact(bimester.Income.Cash))
	deductions := rifStep(TaxISR, rifLineDeductions, SourceRecords(), rifBasisBase)
	deductions.Args = t.bimester
	t.l.Subtract(deductions, t.exact(bimester.Expenses.DeductiblePaid))
	base := t.l.Result(rifStep(TaxISR, rifLineBase, SourceComputed(), rifBasisBase))

	t.caused(table, base)

	reduce := rifStep(TaxISR, rifLineReduction, SourceComputed(), rifBasisReduction)
	reduce.Args, reduce.NeedsCPAReview = []string{year, reduction.String()}, true
	t.l.Scale(reduce, money.RateScale-reduction)
	payable := t.l.Result(rifStep(TaxISR, rifLinePayable, SourceComputed(), rifBasisPayable))
	if t.err == nil {
		t.err = t.l.err
	}
	return NewTaxResult(payable)
}

// caused traces the bimonthly tariff on a positive base, else zero with no
// lookup.
func (t *rifTrace) caused(table taxparams.Table, base money.Exact) {
	if base <= 0 {
		t.l.Input(rifStep(TaxISR, rifLineCaused, SourceComputed(), rifBasisTariff), 0)
		return
	}
	b := t.l.Lookup(rifStep(TaxISR, rifLineTariffRow, SourceParams(rifTariff), rifBasisTariff), table, base)
	t.l.AddProduct(rifStep(TaxISR, rifLineTariffExcess, SourceParams(rifTariff), rifBasisTariff), b.Excess(base), b.Row.Rate)
	t.l.Result(rifStep(TaxISR, rifLineCaused, SourceComputed(), rifBasisTariff))
}
