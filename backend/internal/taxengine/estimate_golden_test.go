package taxengine_test

import (
	"context"
	"fmt"
	"slices"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/clock"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
	"github.com/romodevx/control-contable/backend/internal/platform/tx"
	"github.com/romodevx/control-contable/backend/internal/taxengine"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain"
	"github.com/romodevx/control-contable/backend/internal/taxengine/params"
	"github.com/romodevx/control-contable/backend/internal/taxengine/usecase"
)

// The composite cases of MODULE-tax-engine §3 (M5-22), end to end through
// the EstimateMonth use case with the real registry (taxengine.Strategies)
// and the embedded 2026 parameters; only the ports are in memory. Expected
// values are the spec's, to the centavo.

const (
	txUser     = "0199d4a2-0000-7000-8000-0000000005a1"
	txTaxpayer = "0199d4a2-1111-7000-8000-0000000005a1"
)

// txToday is 20 October 2026: every 2026 month up to October is estimable.
var txToday = time.Date(2026, time.October, 20, 12, 0, 0, 0, clock.Location())

type txUnit struct{}

func (txUnit) ReadWrite(ctx context.Context, _ string, work tx.Work) error { return work(ctx) }
func (txUnit) ReadOnly(ctx context.Context, _ string, work tx.Work) error  { return work(ctx) }

type txProfiles struct{ profile usecase.TaxpayerProfile }

func (f txProfiles) Profile(context.Context, string, string) (usecase.TaxpayerProfile, error) {
	return f.profile, nil
}

type txLedger struct {
	incomes  []usecase.IncomeMonthSums
	expenses []usecase.ExpenseMonthSums
}

// IncomeSums returns the case's rows, each in the query's fiscal year (the
// cases are single-year).
func (f txLedger) IncomeSums(_ context.Context, q usecase.LedgerQuery) ([]usecase.IncomeMonthSums, error) {
	out := slices.Clone(f.incomes)
	for i := range out {
		out[i].Year = q.FiscalYear
	}
	return out, nil
}

// ExpenseSums returns the case's rows, each in the query's fiscal year.
func (f txLedger) ExpenseSums(_ context.Context, q usecase.LedgerQuery) ([]usecase.ExpenseMonthSums, error) {
	out := slices.Clone(f.expenses)
	for i := range out {
		out[i].Year = q.FiscalYear
	}
	return out, nil
}

type txInputs map[fiscal.RegimeCode]domain.TaxInputs

func (f txInputs) TaxInputs(context.Context, string, int) (map[fiscal.RegimeCode]domain.TaxInputs, error) {
	return f, nil
}

// txNoTexts is a catalog without texts: the figures do not read them.
type txNoTexts struct{}

func (txNoTexts) RegimeTexts(context.Context, string) (map[fiscal.RegimeCode]usecase.RegimeText, error) {
	return nil, nil
}

type txNoHolidays struct{}

func (txNoHolidays) IsHoliday(context.Context, fiscal.Date) (bool, error) { return false, nil }

// txCase is one taxpayer's month: its regimes, records and inputs.
type txCase struct {
	personType fiscal.PersonType
	regimes    []usecase.RegimeHold
	ledger     txLedger
	inputs     txInputs
	period     string
}

func (c txCase) estimate(t testing.TB) usecase.TaxpayerEstimate {
	t.Helper()
	got, err := c.estimator(t).EstimateMonth(t.Context(), txUser, txTaxpayer, c.period)
	require.NoError(t, err)
	return got
}

func (c txCase) estimator(t testing.TB) *usecase.Estimator {
	t.Helper()
	registry, err := domain.NewRegistry(taxengine.Strategies()...)
	require.NoError(t, err)
	catalog, err := params.Embedded()
	require.NoError(t, err)
	est, err := usecase.NewEstimator(usecase.EstimateDeps{
		Tx: txUnit{}, Profiles: txProfiles{usecase.TaxpayerProfile{PersonType: c.personType, Regimes: c.regimes}},
		Ledger: c.ledger, Inputs: c.inputs, Texts: txNoTexts{}, Holidays: txNoHolidays{},
		Strategies: registry, Params: catalog, Clock: clock.NewFixed(txToday),
	})
	require.NoError(t, err)
	return est
}

func txHold(t testing.TB, code fiscal.RegimeCode, options map[domain.OptionKey]domain.OptionValue) usecase.RegimeHold {
	t.Helper()
	start, err := domain.ParsePeriod("2026-01")
	require.NoError(t, err)
	return usecase.RegimeHold{Code: code, Treatment: domain.TreatmentFull, Options: domain.NewOptions(options), Start: start}
}

func txValues(t testing.TB, values ...domain.TaxInputValue) domain.TaxInputs {
	t.Helper()
	in, err := domain.NewTaxInputs(values...)
	require.NoError(t, err)
	return in
}

func txPrior(month time.Month, c money.Cents) domain.TaxInputValue {
	return domain.TaxInputValue{Key: domain.InputPriorProvisionalPayments, Kind: domain.InputMoneyMonthly, Month: month, Money: c}
}

func txRate(t testing.TB, key domain.InputKey, s string) domain.TaxInputValue {
	t.Helper()
	r, err := money.ParseRate(s)
	require.NoError(t, err)
	return domain.TaxInputValue{Key: key, Kind: domain.InputRate, Rate: r}
}

func txMoney(key domain.InputKey, c money.Cents) domain.TaxInputValue {
	return domain.TaxInputValue{Key: key, Kind: domain.InputMoney, Money: c}
}

var fullyTaxable = map[domain.OptionKey]domain.OptionValue{domain.OptionIVAFullyTaxable: domain.BoolOption(true)}

// pf612March is PF-612-ISR-01 + PF-612-IVA-01 in one March (records of
// regime 612).
func pf612March() ([]usecase.IncomeMonthSums, []usecase.ExpenseMonthSums) {
	fees := domain.IncomeSums{Cash: 30_000_00, ISRWithheld: domain.Withheld{ByMoral: 1_000_00}}
	march := fees
	march.ByRate = domain.IVAByRate{Rate16: domain.RateSums{Base: 30_000_00, IVA: 4_800_00}}
	march.IVAWithheld = domain.Withheld{ByMoral: 1_066_67}
	incomes := []usecase.IncomeMonthSums{
		{Regime: fiscal.RegimeBusiness, Month: time.January, Sums: fees},
		{Regime: fiscal.RegimeBusiness, Month: time.February, Sums: fees},
		{Regime: fiscal.RegimeBusiness, Month: time.March, Sums: march},
	}
	deductions := domain.ExpenseSums{DeductiblePaid: 10_000_00}
	marchExpenses := deductions
	marchExpenses.CreditableIVAPaid = 1_600_00
	expenses := []usecase.ExpenseMonthSums{
		{Regime: fiscal.RegimeBusiness, Month: time.January, Sums: deductions},
		{Regime: fiscal.RegimeBusiness, Month: time.February, Sums: deductions},
		{Regime: fiscal.RegimeBusiness, Month: time.March, Sums: marchExpenses},
	}
	return incomes, expenses
}

func tx612Total(t testing.TB) txCase {
	incomes, expenses := pf612March()
	return txCase{
		personType: fiscal.PersonFisica,
		regimes:    []usecase.RegimeHold{txHold(t, fiscal.RegimeBusiness, fullyTaxable)},
		ledger:     txLedger{incomes: incomes, expenses: expenses},
		inputs:     txInputs{fiscal.RegimeBusiness: txValues(t, txPrior(time.January, 2_000_00), txPrior(time.February, 2_000_00))},
		period:     "2026-03",
	}
}

func shown(lines []domain.Line) []string {
	out := make([]string, 0, len(lines))
	for _, l := range lines {
		out = append(out, fmt.Sprintf("%s %s %s %s", l.Tax, l.Code, l.Operation, l.Amount))
	}
	return out
}

// TX-612-TOTAL-01 (US-X1): ISR 150.96, IVA 2,133.33, total 2,284.29, due
// 2026-04-17, complete; the contract example.
func TestEstimateGoldenTX612Total01(t *testing.T) {
	got := tx612Total(t).estimate(t)

	require.Len(t, got.Regimes, 1)
	r := got.Regimes[0]
	assert.Equal(t, domain.StatusComplete, r.Status)
	assert.Equal(t, money.Cents(150_96), r.ISR.Amount)
	assert.Equal(t, money.Cents(150_96), got.Totals.ISR.Payable())
	assert.Equal(t, money.Cents(2_133_33), got.Totals.IVA.Payable())
	assert.Zero(t, got.Totals.WithholdingsToRemit)
	assert.Equal(t, money.Cents(2_284_29), got.Totals.TotalPayable)
	assert.Equal(t, "2026-04-17", got.DueDate.String())
	assert.Equal(t, "2026/v1", got.ParameterSet.String())
	assert.Equal(t, []domain.DisclaimerCode{domain.DisclaimerNotTaxAdvice, domain.DisclaimerDeductibilityClaimed}, txCodes(got.Disclaimers))
	assert.Equal(t, []string{
		"isr isr.income_ytd input 90000.00",
		"isr isr.deductions_ytd subtract 30000.00",
		"isr isr.base result 60000.00",
		"isr isr.tariff_row lookup 5570.52",
		"isr isr.tariff_excess multiply 1580.44",
		"isr isr.caused result 7150.96",
		"isr isr.prior_payments subtract 4000.00",
		"isr isr.withheld_by_moral subtract 3000.00",
		"isr isr.payable result 150.96",
		"iva iva.charged.rate_16 input 4800.00",
		"iva iva.charged result 4800.00",
		"iva iva.creditable subtract 1600.00",
		"iva iva.withheld.moral subtract 1066.67",
		"iva iva.regime_net result 2133.33",
	}, shown(r.Lines))
	assert.Equal(t, []string{
		"iva iva.total.charged input 4800.00",
		"iva iva.total.creditable subtract 1600.00",
		"iva iva.total.withheld subtract 1066.67",
		"iva iva.consolidated result 2133.33",
	}, shown(got.IVA.Lines))
}

// TX-612-INC-01 (US-X1b): without the January and February prior payments,
// 612 is incomplete, computed with zero for them: 7,150.96 − 0 − 3,000.00.
func TestEstimateGoldenTX612Inc01(t *testing.T) {
	c := tx612Total(t)
	c.inputs = nil

	got := c.estimate(t)

	r := got.Regimes[0]
	assert.Equal(t, domain.StatusIncomplete, r.Status)
	assert.Equal(t, money.Cents(4_150_96), r.ISR.Amount)
	assert.Equal(t, []domain.MissingInput{{
		Key: domain.InputPriorProvisionalPayments, Months: []time.Month{time.January, time.February},
	}}, r.MissingInputs)
	assert.Contains(t, txCodes(got.Disclaimers), domain.DisclaimerIncompleteInputs)
}

// TX-NET-01 (US-X1d, OQ-T2 ⚠️): a PF with 612 and 606 in March. 612 nets
// +2,133.33; 606 (commercial rent to a PM: IVA 3,200.00, withheld 2,133.33,
// creditable 3,000.00) nets −1,933.33; one IVA of 200.00 payable with the
// strong cross-regime warning. The 606 ISR favor balance (924.00,
// PF-606-ISR-02) stays with 606 (OQ-T10).
func TestEstimateGoldenTXNet01(t *testing.T) {
	c := tx612Total(t)
	c.regimes = append(c.regimes, txHold(t, fiscal.RegimeRent, map[domain.OptionKey]domain.OptionValue{
		domain.OptionRentDeductionMode: domain.TextOption("blind_35"),
	}))
	c.ledger.incomes = append(c.ledger.incomes, usecase.IncomeMonthSums{Regime: fiscal.RegimeRent, Month: time.March, Sums: domain.IncomeSums{
		Cash:        20_000_00,
		ByRate:      domain.IVAByRate{Rate16: domain.RateSums{Base: 20_000_00, IVA: 3_200_00}},
		IVAWithheld: domain.Withheld{ByMoral: 2_133_33},
		ISRWithheld: domain.Withheld{ByMoral: 2_000_00},
	}})
	c.ledger.expenses = append(c.ledger.expenses, usecase.ExpenseMonthSums{Regime: fiscal.RegimeRent, Month: time.March, Sums: domain.ExpenseSums{
		CreditableIVAPaid: 3_000_00,
	}})

	got := c.estimate(t)

	require.Len(t, got.Regimes, 2)
	rent, business := got.Regimes[0], got.Regimes[1]
	assert.Equal(t, fiscal.RegimeRent, rent.Key.Regime)
	assert.Equal(t, money.Cents(-924_00), rent.ISR.Amount)
	assert.Equal(t, domain.IVAComponents{ChargedCollected: 3_200_00, CreditablePaid: 3_000_00, WithheldFromTaxpayer: 2_133_33}, *rent.IVA)
	assert.Equal(t, money.Cents(150_96), business.ISR.Amount)
	assert.Equal(t, money.Cents(200_00), got.Totals.IVA.Amount)
	assert.Equal(t, domain.OutcomePayable, got.Totals.IVA.Outcome())
	assert.Equal(t, money.Cents(150_96), got.Totals.ISR.Amount, "606's favor balance is not offset")
	assert.Equal(t, money.Cents(350_96), got.Totals.TotalPayable)
	assert.Equal(t, []string{
		"iva iva.total.charged input 8000.00",
		"iva iva.total.creditable subtract 4600.00",
		"iva iva.total.withheld subtract 3200.00",
		"iva iva.consolidated result 200.00",
	}, shown(got.IVA.Lines))
	assert.True(t, got.IVA.Lines[3].NeedsCPAReview)
	assert.Equal(t, domain.Disclaimer{
		Code: domain.DisclaimerCPAReviewPending, Severity: domain.SeverityStrong,
		Topics: []domain.CPATopic{domain.TopicCrossRegimeIVANetting},
	}, got.Disclaimers[1])
}

// pm601June is PM-601-ISR-01 with PM-601-IVA-01's June IVA, plus the 10 %
// ISR (5,000.00) the PM withheld on the 50,000 fees.
func pm601June(t testing.TB) txCase {
	accrual := func(m time.Month, c money.Cents) usecase.IncomeMonthSums {
		return usecase.IncomeMonthSums{Regime: "601", Month: m, Sums: domain.IncomeSums{Accrual: c}}
	}
	june := accrual(time.June, 900_000_00)
	june.Sums.Cash = 2_000_000_00
	june.Sums.ByRate = domain.IVAByRate{Rate16: domain.RateSums{Base: 2_000_000_00, IVA: 320_000_00}}
	priors := []domain.TaxInputValue{txRate(t, domain.InputCoeficienteUtilidad, "0.1500"), txMoney(domain.InputPTUPaid, 120_000_00)}
	for m := time.January; m < time.June; m++ {
		priors = append(priors, txPrior(m, 36_000_00))
	}
	return txCase{
		personType: fiscal.PersonMoral,
		regimes:    []usecase.RegimeHold{txHold(t, "601", fullyTaxable)},
		ledger: txLedger{
			incomes: []usecase.IncomeMonthSums{
				accrual(time.January, 800_000_00), accrual(time.February, 800_000_00), accrual(time.March, 800_000_00),
				accrual(time.April, 800_000_00), accrual(time.May, 900_000_00), june,
			},
			expenses: []usecase.ExpenseMonthSums{{Regime: "601", Month: time.June, Sums: domain.ExpenseSums{
				CreditableIVAPaid: 200_000_00, IVAWithheldToRemit: 5_333_33, ISRWithheldToRemit: 5_000_00,
			}}},
		},
		inputs: txInputs{"601": txValues(t, priors...)},
		period: "2026-06",
	}
}

// TX-601-REMIT-01 (US-X1d, the totals rule of PR #89): ISR 36,000.00 + IVA
// 120,000.00 + to remit 10,333.33 (ISR 5,000.00 + IVA 5,333.33), counted
// once: total 166,333.33.
func TestEstimateGoldenTX601Remit01(t *testing.T) {
	got := pm601June(t).estimate(t)

	r := got.Regimes[0]
	assert.Equal(t, domain.StatusComplete, r.Status)
	assert.Equal(t, money.Cents(36_000_00), r.ISR.Amount)
	assert.Equal(t, money.Cents(10_333_33), r.WithheldToRemit)
	assert.Equal(t, money.Cents(120_000_00), got.Totals.IVA.Amount)
	assert.Equal(t, money.Cents(36_000_00), got.Totals.ISR.Amount)
	assert.Equal(t, money.Cents(10_333_33), got.Totals.WithholdingsToRemit)
	assert.Equal(t, money.Cents(166_333_33), got.Totals.TotalPayable)
	assert.Equal(t, "2026-07-17", got.DueDate.String())
}

// TX-ROUND-01 (§2.7): a 601 June whose shown lines miss the rounded base by
// a centavo: coeficiente 0.15 × 1,000.03 = 150.0045 (shown 150.00) minus
// the PTU 0.03 / 8 × 2 = 0.0075 (shown 0.01) is 149.997, rounded once to
// 150.00. One "Ajuste por redondeo" of +0.01 reconciles the lines.
func TestEstimateGoldenTXRound01(t *testing.T) {
	c := txCase{
		personType: fiscal.PersonMoral,
		regimes:    []usecase.RegimeHold{txHold(t, "601", fullyTaxable)},
		ledger: txLedger{incomes: []usecase.IncomeMonthSums{{Regime: "601", Month: time.June, Sums: domain.IncomeSums{
			Accrual: 1_000_03,
		}}}},
		inputs: txInputs{"601": txValues(t, txRate(t, domain.InputCoeficienteUtilidad, "0.15"), txMoney(domain.InputPTUPaid, 3),
			txPrior(time.January, 0), txPrior(time.February, 0), txPrior(time.March, 0), txPrior(time.April, 0), txPrior(time.May, 0))},
		period: "2026-06",
	}

	got := c.estimate(t)

	lines := got.Regimes[0].Lines
	var adjustments []string
	for i, l := range lines {
		if l.Operation == domain.OpRoundingAdjustment {
			adjustments = append(adjustments, fmt.Sprintf("%s before %s", l.Amount, lines[i+1].Code))
		}
	}
	assert.Equal(t, []string{"0.01 before isr.base"}, adjustments)
	assert.Contains(t, shown(lines), "isr isr.coeficiente multiply 150.00")
	assert.Contains(t, shown(lines), "isr isr.ptu subtract 0.01")
	assert.Contains(t, shown(lines), "isr isr.base result 150.00")
	assert.Equal(t, money.Cents(45_00), got.Regimes[0].ISR.Amount)
}

// Byte-identical estimates (ADR-0002 §5): the same input twice gives the
// same estimate and fingerprint.
func TestEstimateGoldenIsDeterministic(t *testing.T) {
	c := tx612Total(t)

	first, second := c.estimate(t), c.estimate(t)

	assert.Equal(t, first, second)
	assert.Regexp(t, `^[0-9a-f]{64}$`, first.RecordsFingerprint)
}

func txCodes(ds []domain.Disclaimer) []domain.DisclaimerCode {
	out := make([]domain.DisclaimerCode, 0, len(ds))
	for _, d := range ds {
		out = append(out, d.Code)
	}
	return out
}
