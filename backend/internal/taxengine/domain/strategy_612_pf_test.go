package domain_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain/taxparams"
)

// The fixtures of this file are prefixed pf612: the regime strategies are
// written in parallel into one test package (PLAN M5-11..M5-19).

const pf612TableKey = taxparams.Key("isr.art106.accumulated")

var pf612Key = domain.StrategyKey{Regime: fiscal.RegimeBusiness, PersonType: fiscal.PersonFisica}

// pf612Params is a synthetic parameter set with the twelve accumulated
// tables. Month m has two rows split at m × 1,000.00, so a lookup on the
// wrong month is visible; March reproduces the shape of the real row 6 of
// PF-612-ISR-01 (lower 52,600.93, fee 5,570.52, 21.36 %). The real 2026
// tables are exercised by the golden gate, never transcribed here.
func pf612Params(t testing.TB) *taxparams.Params {
	t.Helper()
	d := func(s string) fiscal.Date {
		date, err := fiscal.ParseDate(s)
		require.NoError(t, err)
		return date
	}
	basis := []taxparams.Citation{{Source: "dof-synthetic", Article: "Anexo 8"}}
	spec := taxparams.Spec{
		FiscalYear: 2026, Version: 1, PublishedOn: d("2026-10-13"),
		Sources: []taxparams.Source{{
			ID: "dof-synthetic", Title: "Synthetic Anexo 8", URL: "https://example.test/anexo8",
			DOFDate: d("2025-12-31"), Reference: "Anexo 8, B.VI",
		}},
	}
	for m := time.January; m <= time.December; m++ {
		split := money.Cents(int64(m) * 1_000_00)
		rows := []taxparams.Row{
			{Lower: 1, Upper: split, Rate: rate(t, "0.10")},
			{Lower: split + 1, Unbounded: true, FixedFee: split / 10, Rate: rate(t, "0.20")},
		}
		if m == time.March {
			rows = []taxparams.Row{
				{Lower: 1, Upper: 52_600_92, Rate: rate(t, "0.10")},
				{Lower: 52_600_93, Unbounded: true, FixedFee: 5_570_52, Rate: rate(t, "0.2136")},
			}
		}
		spec.Tables = append(spec.Tables, taxparams.Table{
			Key: pf612TableKey, Kind: taxparams.TableMarginal, Month: m,
			Validity: taxparams.Validity{From: d("2026-01-01"), To: d("2026-12-31")}, Rows: rows, Basis: basis,
		})
	}
	p, err := taxparams.NewParams(spec)
	require.NoError(t, err)
	return p
}

// pf612Input is a 612 input for period through raw with the same month
// repeated January..period.
func pf612Input(t *testing.T, raw string, month domain.MonthSums, values ...domain.TaxInputValue) domain.Input {
	t.Helper()
	p := period(t, raw)
	var entries []domain.MonthEntry
	for m := time.January; m <= p.Month(); m++ {
		entries = append(entries, domain.MonthEntry{Month: m, Sums: month})
	}
	ys, err := domain.NewYearSums(p.Month(), entries...)
	require.NoError(t, err)
	return domain.Input{
		Key: pf612Key, Period: p, Sums: ys, Inputs: inputs(t, values...),
		Options: domain.NewOptions(map[domain.OptionKey]domain.OptionValue{domain.OptionIVAFullyTaxable: domain.BoolOption(true)}),
	}
}

func pf612Money(key domain.InputKey, c money.Cents) domain.TaxInputValue {
	return domain.TaxInputValue{Key: key, Kind: domain.InputMoney, Money: c}
}

// pf612Shown renders the lines of one tax as "code op amount".
func pf612Shown(lines []domain.Line, tax domain.TaxKind) []string {
	var out []string
	for _, l := range lines {
		if l.Tax == tax {
			out = append(out, fmt.Sprintf("%s %s %s", l.Code, l.Operation, l.Amount))
		}
	}
	return out
}

func TestStrategy612PFIdentity(t *testing.T) {
	s := domain.NewStrategy612PF()

	assert.Equal(t, pf612Key, s.Key())
	assert.Equal(t, "612/fisica", s.Key().String())
	assert.Equal(t, domain.TreatmentFull, s.Treatment())
}

// The prior provisional payments are required for every month before n in
// which the regime was active (MODULE-tax-engine §2.8); the other 612
// inputs are optional and never required.
func TestStrategy612PFRequiredInputs(t *testing.T) {
	months := func(ms ...time.Month) []time.Month { return ms }
	tests := []struct {
		name   string
		period string
		active []time.Month
		want   []time.Month // nil: nothing required
	}{
		{"january needs nothing", "2026-01", months(time.January), nil},
		{"march, active all year", "2026-03", months(time.January, time.February, time.March), months(time.January, time.February)},
		{"started in march", "2026-03", months(time.March), nil},
		{"held february to april", "2026-04", months(time.February, time.March, time.April), months(time.February, time.March)},
		{"a gap in the year", "2026-05", months(time.January, time.March, time.May), months(time.January, time.March)},
		{"months after the period are ignored", "2026-02", months(time.January, time.February, time.June), months(time.January)},
		{"no active months", "2026-06", nil, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := domain.NewStrategy612PF().RequiredInputs(domain.Options{}, period(t, tt.period), domain.Facts{ActiveMonths: tt.active})

			if tt.want == nil {
				assert.Empty(t, got)
				return
			}
			assert.Equal(t, []domain.RequiredInput{{
				Key: domain.InputPriorProvisionalPayments, Kind: domain.InputMoneyMonthly, Months: tt.want,
			}}, got)
		})
	}
}

// Each ISR step of LISR 106, in order, as a line (research A.2.3).
func TestStrategy612PFISRSteps(t *testing.T) {
	month := domain.MonthSums{
		Income:   domain.IncomeSums{Cash: 30_000_00, ISRWithheld: domain.Withheld{ByMoral: 1_000_00, ByPlatform: 500_00}},
		Expenses: domain.ExpenseSums{DeductiblePaid: 10_000_00},
	}
	tests := []struct {
		name    string
		in      domain.Input
		want    []string
		isr     money.Cents
		outcome domain.Outcome
	}{
		{
			name: "PF-612-ISR-01 on the march table",
			in:   pf612Input(t, "2026-03", month, monthly(time.January, 2_000_00), monthly(time.February, 2_000_00)),
			want: []string{
				"isr.income_ytd input 90000.00", "isr.deductions_ytd subtract 30000.00", "isr.base result 60000.00",
				"isr.tariff_row lookup 5570.52", "isr.tariff_excess multiply 1580.44", "isr.caused result 7150.96",
				"isr.prior_payments subtract 4000.00", "isr.withheld_by_moral subtract 3000.00", "isr.payable result 150.96",
			},
			isr: 150_96, outcome: domain.OutcomePayable,
		},
		{
			name: "missing prior payments compute as zero",
			in:   pf612Input(t, "2026-03", month),
			want: []string{
				"isr.income_ytd input 90000.00", "isr.deductions_ytd subtract 30000.00", "isr.base result 60000.00",
				"isr.tariff_row lookup 5570.52", "isr.tariff_excess multiply 1580.44", "isr.caused result 7150.96",
				"isr.prior_payments subtract 0.00", "isr.withheld_by_moral subtract 3000.00", "isr.payable result 4150.96",
			},
			isr: 4_150_96, outcome: domain.OutcomePayable,
		},
		{
			name: "optional inputs: before the app, PTU and losses",
			in: pf612Input(t, "2026-03", month,
				pf612Money(domain.InputYTDBeforeAppIncome, 20_000_00),
				pf612Money(domain.InputYTDBeforeAppDeductions, 5_000_00),
				pf612Money(domain.InputPTUPaid, 3_000_00),
				pf612Money(domain.InputPriorYearLosses, 2_000_00),
			),
			want: []string{
				"isr.income_ytd input 90000.00", "isr.income_before_app add 20000.00",
				"isr.deductions_ytd subtract 30000.00", "isr.deductions_before_app subtract 5000.00",
				"isr.ptu_paid subtract 3000.00", "isr.prior_year_losses subtract 2000.00", "isr.base result 70000.00",
				"isr.tariff_row lookup 5570.52", "isr.tariff_excess multiply 3716.44", "isr.caused result 9286.96",
				"isr.prior_payments subtract 0.00", "isr.withheld_by_moral subtract 3000.00", "isr.payable result 6286.96",
			},
			isr: 6_286_96, outcome: domain.OutcomePayable,
		},
		{
			name: "PF-612-ISR-03: no positive base, no lookup",
			in: pf612Input(t, "2026-03", domain.MonthSums{
				Income: domain.IncomeSums{Cash: 10_000_00}, Expenses: domain.ExpenseSums{DeductiblePaid: 15_000_00},
			}),
			want: []string{
				"isr.income_ytd input 30000.00", "isr.deductions_ytd subtract 45000.00", "isr.base result -15000.00",
				"isr.caused input 0.00", "isr.prior_payments subtract 0.00", "isr.withheld_by_moral subtract 0.00",
				"isr.payable result 0.00",
			},
			isr: 0, outcome: domain.OutcomeZero,
		},
		{
			name: "losses that wipe out the base",
			in:   pf612Input(t, "2026-01", domain.MonthSums{Income: domain.IncomeSums{Cash: 5_000_00}}, pf612Money(domain.InputPriorYearLosses, 5_000_00)),
			want: []string{
				"isr.income_ytd input 5000.00", "isr.deductions_ytd subtract 0.00", "isr.prior_year_losses subtract 5000.00",
				"isr.base result 0.00", "isr.caused input 0.00", "isr.prior_payments subtract 0.00",
				"isr.withheld_by_moral subtract 0.00", "isr.payable result 0.00",
			},
			isr: 0, outcome: domain.OutcomeZero,
		},
		{
			name: "payments and withholding above the ISR caused: in favor",
			in:   pf612Input(t, "2026-02", month, monthly(time.January, 9_000_00)),
			// February: base 40,000 on row 2: 200.00 + 37,999.99 × 20 % = 7,799.998.
			want: []string{
				"isr.income_ytd input 60000.00", "isr.deductions_ytd subtract 20000.00", "isr.base result 40000.00",
				"isr.tariff_row lookup 200.00", "isr.tariff_excess multiply 7600.00", "isr.caused result 7800.00",
				"isr.prior_payments subtract 9000.00", "isr.withheld_by_moral subtract 2000.00", "isr.payable result -3200.00",
			},
			isr: -3_200_00, outcome: domain.OutcomeInFavor,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			est, err := domain.NewStrategy612PF().Estimate(tt.in, pf612Params(t))

			require.NoError(t, err)
			assert.Equal(t, tt.want, pf612Shown(est.Lines, domain.TaxISR))
			require.NotNil(t, est.ISR)
			assert.Equal(t, tt.isr, est.ISR.Amount)
			assert.Equal(t, tt.outcome, est.ISR.Outcome())
			assertCited(t, est.Lines)
		})
	}
}

// The lines say where each figure comes from and cite LISR 106.
func TestStrategy612PFLineSourcesAndBasis(t *testing.T) {
	in := pf612Input(t, "2026-03", domain.MonthSums{
		Income: domain.IncomeSums{Cash: 30_000_00}, Expenses: domain.ExpenseSums{DeductiblePaid: 10_000_00},
	}, pf612Money(domain.InputPTUPaid, 1_00), monthly(time.January, 0), monthly(time.February, 0))

	est, err := domain.NewStrategy612PF().Estimate(in, pf612Params(t))
	require.NoError(t, err)

	tests := []struct {
		code    domain.LineCode
		source  string
		article string
	}{
		{"isr.income_ytd", "records", "102"},
		{"isr.deductions_ytd", "records", "103"},
		{"isr.ptu_paid", "tax_input:ptu_paid", "106"},
		{"isr.base", "computed", "106"},
		{"isr.tariff_row", "params:isr.art106.accumulated", "106"},
		{"isr.tariff_excess", "params:isr.art106.accumulated", "106"},
		{"isr.caused", "computed", "106"},
		{"isr.prior_payments", "tax_input:prior_provisional_payments", "106"},
		{"isr.withheld_by_moral", "records", "106"},
		{"isr.payable", "computed", "106"},
	}
	for _, tt := range tests {
		t.Run(string(tt.code), func(t *testing.T) {
			line := lineByCode(t, est.Lines, tt.code)

			assert.Equal(t, tt.source, line.Source.String())
			assert.Contains(t, articles(line.Basis), tt.article)
		})
	}
	row := lineByCode(t, est.Lines, "isr.tariff_row")
	assert.Equal(t, 2, row.Row, "the row of the march table")
	assert.Equal(t, rate(t, "0.2136"), lineByCode(t, est.Lines, "isr.tariff_excess").Rate)
	assert.Contains(t, row.Args, "2026-03", "the lookup names the month of the table")
}

// The table is the accumulated one of the period's month, never another.
func TestStrategy612PFUsesTheTableOfMonthN(t *testing.T) {
	month := domain.MonthSums{Income: domain.IncomeSums{Cash: 1_000_00}}
	for m := time.January; m <= time.December; m++ {
		t.Run(m.String(), func(t *testing.T) {
			in := pf612Input(t, fmt.Sprintf("2026-%02d", int(m)), month)

			est, err := domain.NewStrategy612PF().Estimate(in, pf612Params(t))

			require.NoError(t, err)
			// YTD income m × 1,000.00 is the last centavo of row 1 of table m.
			assert.Equal(t, 1, lineByCode(t, est.Lines, "isr.tariff_row").Row)
			assert.Equal(t, money.Cents(int64(m)*100_00), est.ISR.Amount)
		})
	}
}

// IVA is the month's, never YTD, through the shared components (M5-09).
func TestStrategy612PFIVA(t *testing.T) {
	march := domain.MonthSums{
		Income: domain.IncomeSums{
			Cash:        30_000_00,
			ByRate:      domain.IVAByRate{Rate16: domain.RateSums{Base: 30_000_00, IVA: 4_800_00}},
			IVAWithheld: domain.Withheld{ByMoral: 1_066_67},
		},
		Expenses: domain.ExpenseSums{DeductiblePaid: 10_000_00, CreditableIVAPaid: 1_600_00},
	}

	est, err := domain.NewStrategy612PF().Estimate(pf612Input(t, "2026-03", march), pf612Params(t))

	require.NoError(t, err)
	require.NotNil(t, est.IVA)
	assert.Equal(t, domain.IVAComponents{ChargedCollected: 4_800_00, CreditablePaid: 1_600_00, WithheldFromTaxpayer: 1_066_67}, *est.IVA)
	assert.Equal(t, []string{
		"iva.charged.rate_16 input 4800.00", "iva.charged result 4800.00", "iva.creditable subtract 1600.00",
		"iva.withheld.moral subtract 1066.67", "iva.regime_net result 2133.33",
	}, pf612Shown(est.Lines, domain.TaxIVA))
	assert.Zero(t, est.WithheldToRemit)
}

func TestStrategy612PFEstimateFields(t *testing.T) {
	deducted := domain.MonthSums{
		Income:   domain.IncomeSums{Cash: 1_000_00, ByRate: domain.IVAByRate{ExemptBase: 1_00}},
		Expenses: domain.ExpenseSums{DeductiblePaid: 1_00, CreditableIVAPaid: 1_00, IVAWithheldToRemit: 50_00},
	}
	tests := []struct {
		name        string
		month       domain.MonthSums
		inputs      []domain.TaxInputValue
		disclaimers []domain.DisclaimerCode
		toRemit     money.Cents
	}{
		{"no deductions: nothing claimed", domain.MonthSums{Income: domain.IncomeSums{Cash: 1_000_00}}, nil, nil, 0},
		{
			"deductions before the app are claimed too", domain.MonthSums{Income: domain.IncomeSums{Cash: 1_000_00}},
			[]domain.TaxInputValue{pf612Money(domain.InputYTDBeforeAppDeductions, 1_00)},
			[]domain.DisclaimerCode{domain.DisclaimerDeductibilityClaimed}, 0,
		},
		{
			"recorded deductions, mixed activity and IVA to remit", deducted, nil,
			[]domain.DisclaimerCode{domain.DisclaimerDeductibilityClaimed, domain.DisclaimerIVAMixedActivity}, 50_00,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			est, err := domain.NewStrategy612PF().Estimate(pf612Input(t, "2026-01", tt.month, tt.inputs...), pf612Params(t))

			require.NoError(t, err)
			assert.Equal(t, pf612Key, est.Key)
			assert.Equal(t, domain.TreatmentFull, est.Coverage)
			assert.Equal(t, tt.disclaimers, est.Disclaimers)
			assert.Equal(t, tt.toRemit, est.WithheldToRemit)
			assert.Empty(t, est.CPATopics)
			assert.Empty(t, est.Explanation)
			assert.Equal(t, domain.StatusComplete, est.Status())
		})
	}
}

func TestStrategy612PFIsDeterministic(t *testing.T) {
	in := pf612Input(t, "2026-03", domain.MonthSums{
		Income:   domain.IncomeSums{Cash: 33_333_33, ISRWithheld: domain.Withheld{ByMoral: 333_33}},
		Expenses: domain.ExpenseSums{DeductiblePaid: 1_111_11},
	}, monthly(time.January, 1_000_00))
	p := pf612Params(t)

	first, err := domain.NewStrategy612PF().Estimate(in, p)
	require.NoError(t, err)
	second, err := domain.NewStrategy612PF().Estimate(in, p)
	require.NoError(t, err)

	assert.Equal(t, first, second)
}

func TestStrategy612PFFailures(t *testing.T) {
	short := pf612Input(t, "2026-03", domain.MonthSums{Income: domain.IncomeSums{Cash: 1_000_00}})
	var err error
	short.Sums, err = domain.NewYearSums(time.February)
	require.NoError(t, err)
	overflow := pf612Input(t, "2026-02", domain.MonthSums{Income: domain.IncomeSums{Cash: money.Cents(1 << 62)}})
	outOfForce := pf612Input(t, "2027-01", domain.MonthSums{Income: domain.IncomeSums{Cash: 1_000_00}})

	tests := []struct {
		name string
		in   domain.Input
		p    *taxparams.Params
		want error
	}{
		{"sums end before the period", short, pf612Params(t), domain.ErrInvalidMonth},
		{"YTD income overflows", overflow, pf612Params(t), money.ErrOverflow},
		{"no table in force on the period's last day", outOfForce, pf612Params(t), taxparams.ErrNotInForce},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := domain.NewStrategy612PF().Estimate(tt.in, tt.p)

			require.ErrorIs(t, err, tt.want)
		})
	}
}
