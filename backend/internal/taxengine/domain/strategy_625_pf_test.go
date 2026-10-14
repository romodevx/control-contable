package domain_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain/taxparams"
)

// The 625 PF tests name their helpers with a pf625 prefix: every regime
// strategy test shares package domain_test (PLAN PG-M5-S), so a generic
// name would collide.

var pf625Key = domain.StrategyKey{Regime: "625", PersonType: fiscal.PersonFisica}

// pf625Params is a synthetic parameter set with the 2026 platform rates and
// the $300,000 cap, as params/data/2026/v1.yaml transcribes them; the
// golden gate runs the strategy against the real file. Omit drops keys.
func pf625Params(t testing.TB, omit ...taxparams.Key) *taxparams.Params {
	t.Helper()
	from, err := fiscal.ParseDate("2026-01-01")
	require.NoError(t, err)
	basis := []taxparams.Citation{{Source: "lisr", Article: "113-A"}}
	rates := map[taxparams.Key]string{
		"isr.platform.transport_delivery": "0.021",
		"isr.platform.lodging":            "0.04",
		"isr.platform.goods_services":     "0.025",
		"isr.platform.no_rfc":             "0.20",
	}
	spec := taxparams.Spec{
		FiscalYear: 2026, Version: 1, PublishedOn: from,
		Sources: []taxparams.Source{{
			ID: "lisr", Title: "Ley del ISR", URL: "https://www.diputados.gob.mx/LeyesBiblio/pdf/LISR.pdf",
			DOFDate: from, Reference: "arts. 113-A a 113-C",
		}},
	}
	skip := func(k taxparams.Key) bool {
		for _, o := range omit {
			if o == k {
				return true
			}
		}
		return false
	}
	for _, k := range []taxparams.Key{
		"isr.platform.transport_delivery", "isr.platform.lodging", "isr.platform.goods_services", "isr.platform.no_rfc",
	} {
		if skip(k) {
			continue
		}
		spec.Values = append(spec.Values, taxparams.Value{
			Key: k, Kind: taxparams.KindRate, Rate: rate(t, rates[k]),
			Validity: taxparams.Validity{From: from}, Basis: basis,
		})
	}
	if !skip("regime.platform.definitive_option_cap") {
		spec.Values = append(spec.Values, taxparams.Value{
			Key: "regime.platform.definitive_option_cap", Kind: taxparams.KindMoney, Money: 300_000_00,
			Validity: taxparams.Validity{From: from}, Basis: basis,
		})
	}
	p, err := taxparams.NewParams(spec)
	require.NoError(t, err)
	return p
}

// pf625Opts are the 625 options: the activities, the definitive option and
// whether the platform has the RFC.
type pf625Opts struct {
	activities []string
	definitive bool
	rfc        bool
}

func (o pf625Opts) options() domain.Options {
	return domain.NewOptions(map[domain.OptionKey]domain.OptionValue{
		"activities":            domain.ListOption(o.activities),
		"definitive_option":     domain.BoolOption(o.definitive),
		"rfc_given_to_platform": domain.BoolOption(o.rfc),
	})
}

// pf625Input is a March 2026 input with month's sums and, when prior is
// not "", that prior-year income.
func pf625Input(t *testing.T, o pf625Opts, month domain.MonthSums, prior string) domain.Input {
	t.Helper()
	period, err := domain.ParsePeriod("2026-03")
	require.NoError(t, err)
	sums, err := domain.NewYearSums(time.March, domain.MonthEntry{Month: time.March, Sums: month})
	require.NoError(t, err)
	var values []domain.TaxInputValue
	if prior != "" {
		values = append(values, domain.TaxInputValue{
			Key: domain.InputPriorYearIncome, Kind: domain.InputMoney, Money: cents(t, prior),
		})
	}
	inputs, err := domain.NewTaxInputs(values...)
	require.NoError(t, err)
	return domain.Input{Key: pf625Key, Period: period, Options: o.options(), Sums: sums, Inputs: inputs}
}

// pf625Platform is a month of platform income: cash collected, IVA charged
// at 16 % and what the platform withheld.
func pf625Platform(t testing.TB, cash, iva, ivaWithheld, isrWithheld string) domain.MonthSums {
	return domain.MonthSums{Income: domain.IncomeSums{
		Cash:        cents(t, cash),
		ByRate:      domain.IVAByRate{Rate16: domain.RateSums{Base: cents(t, cash), IVA: cents(t, iva)}},
		IVAWithheld: domain.Withheld{ByPlatform: cents(t, ivaWithheld)},
		ISRWithheld: domain.Withheld{ByPlatform: cents(t, isrWithheld)},
	}}
}

type pf625Line struct {
	Code   domain.LineCode
	Op     domain.Operation
	Amount string
}

func pf625Lines(lines []domain.Line, tax domain.TaxKind) []pf625Line {
	var out []pf625Line
	for _, l := range lines {
		if l.Tax == tax {
			out = append(out, pf625Line{Code: l.Code, Op: l.Operation, Amount: l.Amount.String()})
		}
	}
	return out
}

var pf625Goods = []string{"goods_services"}

func TestStrategy625PFIdentity(t *testing.T) {
	s := domain.NewStrategy625PF()
	assert.Equal(t, pf625Key, s.Key())
	assert.Equal(t, domain.TreatmentFull, s.Treatment())
}

func TestStrategy625PFRequiredInputs(t *testing.T) {
	period, err := domain.ParsePeriod("2026-03")
	require.NoError(t, err)
	s := domain.NewStrategy625PF()

	definitive := pf625Opts{activities: pf625Goods, definitive: true, rfc: true}.options()
	assert.Equal(t, []domain.RequiredInput{{Key: domain.InputPriorYearIncome, Kind: domain.InputMoney}},
		s.RequiredInputs(definitive, period, domain.Facts{}), "MODULE-tax-engine §2.8")

	provisional := pf625Opts{activities: pf625Goods, rfc: true}.options()
	assert.Empty(t, s.RequiredInputs(provisional, period, domain.Facts{}))
	assert.Empty(t, s.RequiredInputs(domain.Options{}, period, domain.Facts{}), "definitive_option defaults to false")
}

func TestStrategy625PFDefinitiveISR(t *testing.T) {
	tests := []struct {
		name    string
		opts    pf625Opts
		month   domain.MonthSums
		want    []pf625Line
		payable string
	}{
		{
			name: "PF-625-01: goods 30,000, 2.5 % withheld is the ISR (LIF 2026 art. 25 fr. VI)",
			opts: pf625Opts{activities: pf625Goods, definitive: true, rfc: true}, month: pf625Platform(t, "30000.00", "4800.00", "2400.00", "750.00"),
			want: []pf625Line{
				{"isr.platform_income", domain.OpInput, "30000.00"},
				{"isr.platform_rate", domain.OpMultiply, "750.00"},
				{"isr.withheld_by_platform", domain.OpSubtract, "750.00"},
				{"isr.payable", domain.OpResult, "0.00"},
			},
			payable: "0.00",
		},
		{
			name: "PF-625-02: lodging 15,000 at 4 % (LISR 113-A fr. II)",
			opts: pf625Opts{activities: []string{"lodging"}, definitive: true, rfc: true}, month: pf625Platform(t, "15000.00", "0.00", "0.00", "600.00"),
			want: []pf625Line{
				{"isr.platform_income", domain.OpInput, "15000.00"},
				{"isr.platform_rate", domain.OpMultiply, "600.00"},
				{"isr.withheld_by_platform", domain.OpSubtract, "600.00"},
				{"isr.payable", domain.OpResult, "0.00"},
			},
			payable: "0.00",
		},
		{
			name: "transport and delivery at 2.1 % (LISR 113-A fr. I)",
			opts: pf625Opts{activities: []string{"transport_delivery"}, definitive: true, rfc: true}, month: pf625Platform(t, "10000.00", "1600.00", "800.00", "210.00"),
			want: []pf625Line{
				{"isr.platform_income", domain.OpInput, "10000.00"},
				{"isr.platform_rate", domain.OpMultiply, "210.00"},
				{"isr.withheld_by_platform", domain.OpSubtract, "210.00"},
				{"isr.payable", domain.OpResult, "0.00"},
			},
			payable: "0.00",
		},
		{
			name: "no RFC given to the platform: 20 % whatever the activity (LISR 113-C fr. IV)",
			opts: pf625Opts{activities: []string{"lodging", "goods_services"}, definitive: true}, month: pf625Platform(t, "10000.00", "1600.00", "1600.00", "2000.00"),
			want: []pf625Line{
				{"isr.platform_income", domain.OpInput, "10000.00"},
				{"isr.platform_rate", domain.OpMultiply, "2000.00"},
				{"isr.withheld_by_platform", domain.OpSubtract, "2000.00"},
				{"isr.payable", domain.OpResult, "0.00"},
			},
			payable: "0.00",
		},
		{
			name: "directly collected income is self-paid at the same rate (LISR 113-A last paragraph)",
			opts: pf625Opts{activities: pf625Goods, definitive: true, rfc: true}, month: pf625Platform(t, "40000.00", "6400.00", "2400.00", "750.00"),
			want: []pf625Line{
				{"isr.platform_income", domain.OpInput, "40000.00"},
				{"isr.platform_rate", domain.OpMultiply, "1000.00"},
				{"isr.withheld_by_platform", domain.OpSubtract, "750.00"},
				{"isr.payable", domain.OpResult, "250.00"},
			},
			payable: "250.00",
		},
		{
			name: "a withholding above the rate leaves no balance in favor: it is definitive (LISR 113-B)",
			opts: pf625Opts{activities: pf625Goods, definitive: true, rfc: true}, month: pf625Platform(t, "30000.00", "4800.00", "2400.00", "750.02"),
			want: []pf625Line{
				{"isr.platform_income", domain.OpInput, "30000.00"},
				{"isr.platform_rate", domain.OpMultiply, "750.00"},
				{"isr.withheld_by_platform", domain.OpSubtract, "750.02"},
				{"isr.definitive_no_favor", domain.OpAdd, "0.02"},
				{"isr.payable", domain.OpResult, "0.00"},
			},
			payable: "0.00",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			est, err := domain.NewStrategy625PF().Estimate(pf625Input(t, tc.opts, tc.month, "250000.00"), pf625Params(t))
			require.NoError(t, err)
			assert.Equal(t, domain.TreatmentFull, est.Coverage)
			assert.Equal(t, tc.want, pf625Lines(est.Lines, domain.TaxISR))
			require.NotNil(t, est.ISR)
			assert.Equal(t, tc.payable, est.ISR.Amount.String())
			assert.Equal(t, domain.OutcomeZero == est.ISR.Outcome(), tc.payable == "0.00")
		})
	}
}

func TestStrategy625PFDefinitiveEligibility(t *testing.T) {
	month := pf625Platform(t, "30000.00", "4800.00", "2400.00", "750.00")
	tests := []struct {
		name       string
		definitive bool
		prior      string
		full       bool
	}{
		{"prior-year income at the cap: definitive (LISR 113-B fr. I)", true, "300000.00", true},
		{"prior-year income a centavo over the cap: provisional", true, "300000.01", false},
		{"prior-year income missing: evaluated as 0, eligible (§2.8)", true, "", true},
		{"no definitive option: provisional", false, "100000.00", false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			o := pf625Opts{activities: pf625Goods, definitive: tc.definitive, rfc: true}
			est, err := domain.NewStrategy625PF().Estimate(pf625Input(t, o, month, tc.prior), pf625Params(t))
			require.NoError(t, err)
			require.NotNil(t, est.IVA)
			if tc.full {
				assert.Equal(t, domain.TreatmentFull, est.Coverage)
				assert.NotNil(t, est.ISR)
				assert.Empty(t, est.Explanation)
				assert.Equal(t, cents(t, "2400.00"), est.IVA.DefinitiveExcluded)
				return
			}
			assert.Equal(t, domain.TreatmentPartial, est.Coverage)
			assert.Nil(t, est.ISR, "ISR is computed under 612 or RESICO, not here")
			assert.Equal(t, domain.LabelKey("platform_provisional_withholding"), est.Explanation)
			assert.Equal(t, []pf625Line{{"isr.withheld_by_platform_provisional", domain.OpInfo, "750.00"}},
				pf625Lines(est.Lines, domain.TaxISR), "the withholding shows as a provisional credit")
			assert.Zero(t, est.IVA.DefinitiveExcluded, "general IVA rules")
			assert.Empty(t, est.CPATopics)
		})
	}
}

func TestStrategy625PFMixedActivitiesWithRFCLeaveISRUnestimated(t *testing.T) {
	// The month's sums do not split income by activity, so no single
	// rate applies; the IVA does not depend on the activity.
	o := pf625Opts{activities: []string{"lodging", "goods_services"}, definitive: true, rfc: true}
	est, err := domain.NewStrategy625PF().Estimate(
		pf625Input(t, o, pf625Platform(t, "30000.00", "4800.00", "2400.00", "750.00"), "100000.00"), pf625Params(t))
	require.NoError(t, err)
	assert.Equal(t, domain.TreatmentPartial, est.Coverage)
	assert.Nil(t, est.ISR)
	assert.Equal(t, domain.LabelKey("platform_mixed_activities"), est.Explanation)
	assert.Equal(t, []pf625Line{{"isr.withheld_by_platform_definitive", domain.OpInfo, "750.00"}},
		pf625Lines(est.Lines, domain.TaxISR))
	require.NotNil(t, est.IVA)
	assert.Equal(t, cents(t, "2400.00"), est.IVA.DefinitiveExcluded)
}

func TestStrategy625PFUnknownOrNoActivityIsMixed(t *testing.T) {
	for _, activities := range [][]string{nil, {"rides"}} {
		o := pf625Opts{activities: activities, definitive: true, rfc: true}
		est, err := domain.NewStrategy625PF().Estimate(
			pf625Input(t, o, pf625Platform(t, "100.00", "16.00", "8.00", "2.50"), "0.00"), pf625Params(t))
		require.NoError(t, err)
		assert.Equal(t, domain.TreatmentPartial, est.Coverage, "%v", activities)
		assert.Nil(t, est.ISR)
	}
}

func TestStrategy625PFDefinitiveIVA(t *testing.T) {
	tests := []struct {
		name  string
		rfc   bool
		month domain.MonthSums
		want  domain.IVAComponents
		lines []pf625Line
	}{
		{
			name: "PF-625-01: the unwithheld half is settled by the withholding (LIVA 18-M ⚠️)",
			rfc:  true, month: pf625Platform(t, "30000.00", "4800.00", "2400.00", "750.00"),
			want: domain.IVAComponents{ChargedCollected: 4_800_00, WithheldFromTaxpayer: 2_400_00, DefinitiveExcluded: 2_400_00},
			lines: []pf625Line{
				{"iva.charged.rate_16", domain.OpInput, "4800.00"},
				{"iva.charged", domain.OpResult, "4800.00"},
				{"iva.definitive_excluded", domain.OpSubtract, "2400.00"},
				{"iva.creditable", domain.OpSubtract, "0.00"},
				{"iva.withheld.moral", domain.OpSubtract, "0.00"},
				{"iva.withheld.platform", domain.OpSubtract, "2400.00"},
				{"iva.regime_net", domain.OpResult, "0.00"},
			},
		},
		{
			name: "no RFC: the platform withholds all of it, nothing left to exclude (LIVA 18-J fr. II a)",
			rfc:  false, month: pf625Platform(t, "30000.00", "4800.00", "4800.00", "6000.00"),
			want: domain.IVAComponents{ChargedCollected: 4_800_00, WithheldFromTaxpayer: 4_800_00},
		},
		{
			name: "directly collected IVA stays under the general rules",
			rfc:  true, month: pf625Platform(t, "40000.00", "6400.00", "2400.00", "750.00"),
			want: domain.IVAComponents{ChargedCollected: 6_400_00, WithheldFromTaxpayer: 2_400_00, DefinitiveExcluded: 2_400_00},
		},
		{
			name: "a withholding above half the IVA never excludes more than the rest",
			rfc:  true, month: pf625Platform(t, "30000.00", "1000.00", "2400.00", "750.00"),
			want: domain.IVAComponents{ChargedCollected: 1_000_00, WithheldFromTaxpayer: 2_400_00},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			o := pf625Opts{activities: pf625Goods, definitive: true, rfc: tc.rfc}
			est, err := domain.NewStrategy625PF().Estimate(pf625Input(t, o, tc.month, "250000.00"), pf625Params(t))
			require.NoError(t, err)
			require.NotNil(t, est.IVA)
			assert.Equal(t, tc.want, *est.IVA)
			if tc.lines != nil {
				assert.Equal(t, tc.lines, pf625Lines(est.Lines, domain.TaxIVA))
			}
			if tc.want.DefinitiveExcluded > 0 {
				assert.Equal(t, []domain.CPATopic{domain.TopicPlatformDefinitiveIVA}, est.CPATopics)
				for _, l := range est.Lines {
					if l.Code == "iva.definitive_excluded" {
						assert.True(t, l.NeedsCPAReview)
					}
				}
			}
		})
	}
}

func TestStrategy625PFDefinitiveIVATakesNoCreditableIVA(t *testing.T) {
	// PR #79 follow-up: under the definitive option the IVA paid on
	// expenses does not reduce the IVA (LIVA 18-L/18-M ⚠️); taking it
	// would turn the month in favor.
	month := pf625Platform(t, "30000.00", "4800.00", "2400.00", "750.00")
	month.Expenses = domain.ExpenseSums{CreditableIVAPaid: 1_600_00}
	o := pf625Opts{activities: pf625Goods, definitive: true, rfc: true}
	est, err := domain.NewStrategy625PF().Estimate(pf625Input(t, o, month, "250000.00"), pf625Params(t))
	require.NoError(t, err)
	require.NotNil(t, est.IVA)
	assert.Zero(t, est.IVA.CreditablePaid)
	assert.Equal(t, domain.IVAComponents{ChargedCollected: 4_800_00, WithheldFromTaxpayer: 2_400_00, DefinitiveExcluded: 2_400_00}, *est.IVA)
	var info *domain.Line
	for i, l := range est.Lines {
		if l.Code == "iva.creditable_not_applied" {
			info = &est.Lines[i]
		}
	}
	require.NotNil(t, info, "the IVA not credited is shown, never dropped silently")
	assert.Equal(t, domain.OpInfo, info.Operation)
	assert.Equal(t, cents(t, "1600.00"), info.Amount)
	assert.True(t, info.NeedsCPAReview)
	assert.NotContains(t, est.Disclaimers, domain.DisclaimerIVAMixedActivity)

	// Without the definitive option the general rules credit it.
	o.definitive = false
	est, err = domain.NewStrategy625PF().Estimate(pf625Input(t, o, month, ""), pf625Params(t))
	require.NoError(t, err)
	assert.Equal(t, cents(t, "1600.00"), est.IVA.CreditablePaid)
	assert.NotContains(t, codes(est.Lines), domain.LineCode("iva.creditable_not_applied"))
}

func TestStrategy625PFLinesCiteTheirBasis(t *testing.T) {
	o := pf625Opts{activities: pf625Goods, definitive: true, rfc: true}
	est, err := domain.NewStrategy625PF().Estimate(
		pf625Input(t, o, pf625Platform(t, "30000.00", "4800.00", "2400.00", "750.02"), "250000.00"), pf625Params(t))
	require.NoError(t, err)
	for _, l := range est.Lines {
		if l.Operation == domain.OpRoundingAdjustment {
			continue
		}
		require.NotEmpty(t, l.Basis, "%s cites a legal basis", l.Code)
		for _, c := range l.Basis {
			assert.True(t, c.Valid(), "%s: %+v", l.Code, c)
		}
	}
	rateLine := est.Lines[1]
	require.Equal(t, domain.LineCode("isr.platform_rate"), rateLine.Code)
	assert.Equal(t, "params:isr.platform.goods_services", rateLine.Source.String())
	assert.Equal(t, "0.025", rateLine.Rate.String())
	assert.Equal(t, []domain.Citation{{Law: "LIF 2026", Article: "25", Note: "fracción VI, tasa del 2.5 %"}}, rateLine.Basis)
}

func TestStrategy625PFRegistryStatus(t *testing.T) {
	r, err := domain.NewRegistry(domain.NewStrategy625PF())
	require.NoError(t, err)
	month := pf625Platform(t, "30000.00", "4800.00", "2400.00", "750.00")

	definitive := pf625Opts{activities: pf625Goods, definitive: true, rfc: true}
	est, err := r.Estimate(pf625Input(t, definitive, month, "250000.00"), []time.Month{time.March}, pf625Params(t))
	require.NoError(t, err)
	assert.Equal(t, domain.StatusComplete, est.Status())

	est, err = r.Estimate(pf625Input(t, definitive, month, ""), []time.Month{time.March}, pf625Params(t))
	require.NoError(t, err)
	assert.Equal(t, domain.StatusIncomplete, est.Status())
	assert.Equal(t, []domain.MissingInput{{Key: domain.InputPriorYearIncome}}, est.MissingInputs)

	provisional := pf625Opts{activities: pf625Goods, rfc: true}
	est, err = r.Estimate(pf625Input(t, provisional, month, ""), []time.Month{time.March}, pf625Params(t))
	require.NoError(t, err)
	assert.Equal(t, domain.StatusPartial, est.Status())
}

func TestStrategy625PFFailures(t *testing.T) {
	month := pf625Platform(t, "30000.00", "4800.00", "2400.00", "750.00")
	definitive := pf625Opts{activities: pf625Goods, definitive: true, rfc: true}
	tests := []struct {
		name   string
		opts   pf625Opts
		params *taxparams.Params
	}{
		{"activity rate missing", definitive, pf625Params(t, "isr.platform.goods_services")},
		{"no-RFC rate missing", pf625Opts{activities: pf625Goods, definitive: true}, pf625Params(t, "isr.platform.no_rfc")},
		{"definitive cap missing", definitive, pf625Params(t, "regime.platform.definitive_option_cap")},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := domain.NewStrategy625PF().Estimate(pf625Input(t, tc.opts, month, "250000.00"), tc.params)
			assert.ErrorIs(t, err, taxparams.ErrUnknownKey)
		})
	}

	overflow := pf625Platform(t, "0.00", "0.00", "0.00", "0.00")
	overflow.Income.Cash = money.Cents(1 << 62)
	_, err := domain.NewStrategy625PF().Estimate(pf625Input(t, definitive, overflow, "0.00"), pf625Params(t))
	assert.ErrorIs(t, err, money.ErrOverflow)
}

func TestStrategy625PFIsDeterministic(t *testing.T) {
	o := pf625Opts{activities: pf625Goods, definitive: true, rfc: true}
	in := pf625Input(t, o, pf625Platform(t, "40000.00", "6400.00", "2400.00", "750.00"), "250000.00")
	first, err := domain.NewStrategy625PF().Estimate(in, pf625Params(t))
	require.NoError(t, err)
	for range 5 {
		again, err := domain.NewStrategy625PF().Estimate(in, pf625Params(t))
		require.NoError(t, err)
		assert.Equal(t, first, again)
	}
}
