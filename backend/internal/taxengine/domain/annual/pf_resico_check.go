package annual

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain/taxparams"
)

// paramRESICOPFAnnual is the LISR 113-F annual table: a flat rate on the
// whole income of the year (not marginal).
const paramRESICOPFAnnual taxparams.Key = "isr.resico_pf.annual"

// The line codes of pf_resico_check, in order. Above the table there is no
// rate: the above-table info line replaces the rate and the difference.
const (
	LineRESICOCheckIncome     domain.LineCode = "annual.resico.income"
	LineRESICOCheckRate       domain.LineCode = "annual.resico.rate"
	LineRESICOCheckAboveTable domain.LineCode = "annual.resico.above_table"
	LineRESICOCheckPayments   domain.LineCode = "annual.resico.payments"
	LineRESICOCheckWithheld   domain.LineCode = "annual.resico.withheld"
	LineRESICOCheckDifference domain.LineCode = "annual.resico.difference"
	LineRESICOCheckDefinitive domain.LineCode = "annual.resico.definitive"
)

// The legal basis of each pf_resico_check step (research §B.3).
var (
	basisRESICOIncome = []domain.Citation{
		{Law: lawLISR, Article: "113-F", Note: "primer párrafo: ingresos cobrados en el ejercicio, sin IVA y sin deducciones"},
	}
	basisRESICORate = []domain.Citation{
		{Law: lawLISR, Article: "113-F", Note: "tabla anual: la tasa aplica a la totalidad de los ingresos"},
	}
	basisRESICOPayments = []domain.Citation{
		{Law: lawLISR, Article: "113-F", Note: "segundo párrafo: menos los pagos mensuales del ejercicio"},
	}
	basisRESICOWithheld = []domain.Citation{
		{Law: lawLISR, Article: "113-J", Note: "retención del 1.25 % por personas morales"},
	}
	basisRESICODifference = []domain.Citation{
		{Law: lawLISR, Article: "113-F", Note: "segundo párrafo: diferencia del ejercicio"},
		{Law: "RMF 2026", Article: "3.13.7", Note: "pago mensual con carácter definitivo: diferencia solo informativa"},
	}
	basisRESICODefinitive = []domain.Citation{
		{Law: "RMF 2026", Article: "3.13.7", Note: "no se presenta declaración anual por los ingresos de RESICO"},
	}
)

// PFRESICOCheck is the pf_resico_check section (MODULE-annual §2.7, research
// §B.3): an informative reconciliation of RESICO PF. The year's income
// collected, without deductions, times the 113-F annual rate of its row,
// minus the monthly ISR paid (by source) and the 1.25 % withheld by
// personas morales, is a "diferencia informativa": rule 3.13.7 makes the
// monthly payments definitive, so it is never payable nor refundable and
// the section has no balance.
type PFRESICOCheck struct{}

// NewPFRESICOCheck returns the strategy.
func NewPFRESICOCheck() PFRESICOCheck { return PFRESICOCheck{} }

// Kind is pf_resico_check.
func (PFRESICOCheck) Kind() SectionKind { return SectionPFRESICOCheck }

// Applies reports a year holding RESICO (626).
func (PFRESICOCheck) Applies(in PFYear) bool { return in.holds(fiscal.RegimeRESICO) }

// Compute traces the check; the year must hold 626. It is pure.
func (c PFRESICOCheck) Compute(in PFYear, payments PaymentsReader, p *taxparams.Params) (Section, error) {
	s, err := c.compute(in, payments, p)
	if err != nil {
		return Section{}, fmt.Errorf("%s %d: %w", SectionPFRESICOCheck, in.FiscalYear, err)
	}
	return s, nil
}

func (PFRESICOCheck) compute(in PFYear, payments PaymentsReader, p *taxparams.Params) (Section, error) {
	if err := in.validate(); err != nil {
		return Section{}, err
	}
	y, ok := in.regime(fiscal.RegimeRESICO)
	if !ok {
		return Section{}, fmt.Errorf("regime %s not held: %w", fiscal.RegimeRESICO, ErrInvalidYear)
	}
	day, err := fiscal.ParseDate(fmt.Sprintf("%04d-12-31", in.FiscalYear))
	if err != nil {
		return Section{}, fmt.Errorf("annual: fiscal year: %w", err)
	}
	table, err := p.Table(paramRESICOPFAnnual, day)
	if err != nil {
		return Section{}, err
	}
	paid, err := readPayments(payments, in.FiscalYear, y.Regime)
	if err != nil {
		return Section{}, err
	}
	s := Section{
		Kind: SectionPFRESICOCheck, Regimes: []fiscal.RegimeCode{y.Regime},
		Status: StatusInformational, Obligation: ObligationRelieved, MissingInputs: paid.missing,
		CPATopics:   []CPATopic{TopicRESICOPFReliefRenewal},
		Disclaimers: []DisclaimerCode{DisclaimerCPAReviewPending, DisclaimerRESICOPFDefinitive},
	}
	if paid.estimated {
		s.Disclaimers = append(s.Disclaimers, DisclaimerPaymentsEstimated)
	}
	t := newAnnualTrace()
	s.RESICOCheck, err = t.resicoDifference(y, table, paid)
	if err != nil {
		return Section{}, err
	}
	t.l.Info(annualStep(LineRESICOCheckDefinitive, "annual_resico_definitive", domain.SourceComputed(), basisRESICODefinitive), 0)
	if s.Lines, err = t.lines(); err != nil {
		return Section{}, err
	}
	return s, nil
}

// resicoDifference traces income × rate − payments − withheld. Above the
// table there is no 113-F rate: an info line and no reconciliation. Without
// income the rate is the first row's and nothing is multiplied.
func (t *annualTrace) resicoDifference(y RegimeYear, table taxparams.Table, paid regimePayments) (*RESICOCheck, error) {
	if len(table.Rows) == 0 {
		return nil, fmt.Errorf("%s: no rows: %w", table.Key, taxparams.ErrInvalidTable)
	}
	check := RESICOCheck{Income: y.Income, Rate: table.Rows[0].Rate, Withheld: y.ISRWithheld}
	income := t.exact(y.Income)
	t.l.Input(annualStep(LineRESICOCheckIncome, "annual_resico_income", domain.SourceRecords(), basisRESICOIncome, string(y.Regime)), income)
	if income > 0 {
		b, err := table.Bracket(income)
		if errors.Is(err, taxparams.ErrAboveTable) {
			t.l.Info(annualStep(LineRESICOCheckAboveTable, "annual_resico_above_table", domain.SourceParams(paramRESICOPFAnnual), basisRESICORate), income)
			return nil, nil
		}
		if err != nil {
			return nil, err
		}
		// The payments are subtracted from the ISR as shown (difference =
		// annual ISR − payments − withheld), not from its exact value.
		t.l.ScaleRounded(annualStep(LineRESICOCheckRate, "annual_resico_rate", domain.SourceParams(paramRESICOPFAnnual), basisRESICORate, strconv.Itoa(b.Number)), b.Row.Rate)
		check.Rate = b.Row.Rate
		check.AnnualISR = t.l.Running().Round()
	}
	check.Payments = t.subtractPayments(paid.groups, LineRESICOCheckPayments, "annual_resico_payments", basisRESICOPayments)
	if y.ISRWithheld > 0 {
		t.l.Subtract(annualStep(LineRESICOCheckWithheld, "annual_resico_withheld", domain.SourceRecords(), basisRESICOWithheld, string(y.Regime)), t.exact(y.ISRWithheld))
	}
	check.Difference = t.l.Result(annualStep(LineRESICOCheckDifference, "annual_resico_difference", domain.SourceComputed(), basisRESICODifference)).Round()
	return &check, t.err
}
