package taxengine_test

import (
	"flag"
	"math/rand/v2"
	"slices"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/clock"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
	"github.com/romodevx/control-contable/backend/internal/taxengine"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain/taxparams"
	"github.com/romodevx/control-contable/backend/internal/taxengine/params"
	"github.com/romodevx/control-contable/backend/internal/taxengine/usecase"
)

// The generators of the engine property suite (M5-24): realistic monthly
// ledgers for every registered regime (PF and PM) plus catalog
// informational ones, any month of 2026. Every case is drawn from its own
// PCG stream (propSeed, index), so a failure names the index that
// reproduces it alone.

const (
	// propSeed is the suite's fixed seed: changing it is a reviewed change.
	propSeed = 2026_0524
	// propDefaultCases keeps the default CI run well under a second per
	// property; -taxengine.property.cases raises it for a nightly run.
	propDefaultCases = 1000
	propYear         = 2026
	propPeso         = money.Cents(100)
	// propPercent is one percent of money.RateScale.
	propPercent = money.RateScale / 100
	// rifYears are the RIF years 1..10 of the rif_year option.
	rifYears = 10
)

var propCases = flag.Int("taxengine.property.cases", propDefaultCases,
	"cases per engine property (nightly: 20000)")

// propToday is 15 January 2027: every 2026 month is in the past.
var propToday = time.Date(propYear+1, time.January, 15, 12, 0, 0, 0, clock.Location())

// propEngine is the real registry and the embedded parameters, built once.
var propEngine = sync.OnceValues(func() (domain.Registry, *taxparams.Catalog) {
	registry, err := domain.NewRegistry(taxengine.Strategies()...)
	if err != nil {
		panic(err)
	}
	catalog, err := params.Embedded()
	if err != nil {
		panic(err)
	}
	return registry, catalog
})

// propRegime is one catalog regime the generator may hold.
type propRegime struct {
	code      fiscal.RegimeCode
	person    fiscal.PersonType
	treatment domain.Treatment
}

var propRegimes = []propRegime{
	{"601", fiscal.PersonMoral, domain.TreatmentFull},
	{"603", fiscal.PersonMoral, domain.TreatmentPartial},
	{"606", fiscal.PersonFisica, domain.TreatmentFull},
	{"612", fiscal.PersonFisica, domain.TreatmentFull},
	{"621", fiscal.PersonFisica, domain.TreatmentPartial},
	{"622", fiscal.PersonMoral, domain.TreatmentPartial},
	{"625", fiscal.PersonFisica, domain.TreatmentFull},
	{"626", fiscal.PersonFisica, domain.TreatmentFull},
	{"626", fiscal.PersonMoral, domain.TreatmentFull},
	// Catalog regimes without a strategy: informational fallback.
	{"605", fiscal.PersonFisica, domain.TreatmentInformational},
	{"614", fiscal.PersonFisica, domain.TreatmentInformational},
	{"620", fiscal.PersonMoral, domain.TreatmentInformational},
	{"624", fiscal.PersonMoral, domain.TreatmentInformational},
}

// propCase is one taxpayer's year and the month to estimate. Inputs keep
// their value lists, so a case can be rebuilt in another order.
type propCase struct {
	index    int
	person   fiscal.PersonType
	holds    []usecase.RegimeHold
	incomes  []usecase.IncomeMonthSums
	expenses []usecase.ExpenseMonthSums
	inputs   map[fiscal.RegimeCode][]domain.TaxInputValue
	period   string
}

func (c propCase) estimator(t testing.TB) *usecase.Estimator {
	t.Helper()
	registry, catalog := propEngine()
	inputs := txInputs{}
	for code, values := range c.inputs {
		in, err := domain.NewTaxInputs(values...)
		require.NoError(t, err, "case %d inputs %s", c.index, code)
		inputs[code] = in
	}
	est, err := usecase.NewEstimator(usecase.EstimateDeps{
		Tx: txUnit{}, Profiles: txProfiles{usecase.TaxpayerProfile{PersonType: c.person, Regimes: c.holds}},
		Ledger: txLedger{incomes: c.incomes, expenses: c.expenses}, Inputs: inputs, Texts: txNoTexts{}, Holidays: seededHolidays{},
		Strategies: registry, Params: catalog, Clock: clock.NewFixed(propToday),
	})
	require.NoError(t, err)
	return est
}

func (c propCase) estimate(t testing.TB) usecase.TaxpayerEstimate {
	t.Helper()
	got, err := c.estimator(t).EstimateMonth(t.Context(), txUser, txTaxpayer, c.period)
	require.NoError(t, err, "case %d (%s %s)", c.index, c.person, c.period)
	return got
}

// clone deep-copies the slices a property mutates.
func (c propCase) clone() propCase {
	c.holds = slices.Clone(c.holds)
	c.incomes = slices.Clone(c.incomes)
	c.expenses = slices.Clone(c.expenses)
	inputs := make(map[fiscal.RegimeCode][]domain.TaxInputValue, len(c.inputs))
	for code, values := range c.inputs {
		inputs[code] = slices.Clone(values)
	}
	c.inputs = inputs
	return c
}

// month is the period's month.
func (c propCase) month(t testing.TB) time.Month {
	t.Helper()
	p, err := domain.ParsePeriod(c.period)
	require.NoError(t, err)
	return p.Month()
}

// forEachCase runs check on every generated case, each from its own
// stream.
func forEachCase(t *testing.T, check func(t *testing.T, c propCase)) {
	t.Helper()
	for i := range *propCases {
		rng := rand.New(rand.NewPCG(propSeed, uint64(i))) //nolint:gosec // a fixed, reproducible seed is the point
		c := genCase(rng, i)
		check(t, c)
		if t.Failed() {
			t.Fatalf("first failing case: %d (rerun with -run %s and this index)", i, t.Name())
		}
	}
}

// propGen draws one case.
type propGen struct {
	rng    *rand.Rand
	border bool // the taxpayer also invoices at the border 8 %
}

func genCase(rng *rand.Rand, index int) propCase {
	g := propGen{rng: rng, border: rng.IntN(5) == 0}
	person := fiscal.PersonFisica
	if rng.IntN(3) == 0 {
		person = fiscal.PersonMoral
	}
	c := propCase{index: index, person: person, inputs: map[fiscal.RegimeCode][]domain.TaxInputValue{}}
	c.period = periodString(time.Month(1 + rng.IntN(12)))
	for _, r := range g.pickRegimes(person) {
		hold := g.hold(r)
		c.holds = append(c.holds, hold)
		if r.treatment == domain.TreatmentInformational {
			continue
		}
		for m := time.January; m <= time.December; m++ {
			if !holdCovers(hold, m) || rng.IntN(8) == 0 {
				continue // a month without records
			}
			c.incomes = append(c.incomes, usecase.IncomeMonthSums{Regime: r.code, Month: m, Sums: g.income(r)})
			c.expenses = append(c.expenses, usecase.ExpenseMonthSums{Regime: r.code, Month: m, Sums: g.expense(r)})
		}
		if values := g.taxInputs(r); len(values) > 0 {
			c.inputs[r.code] = values
		}
	}
	return c
}

// pickRegimes holds one to three distinct regime codes of person.
func (g propGen) pickRegimes(person fiscal.PersonType) []propRegime {
	var pool []propRegime
	for _, r := range propRegimes {
		if r.person == person {
			pool = append(pool, r)
		}
	}
	g.rng.Shuffle(len(pool), func(i, j int) { pool[i], pool[j] = pool[j], pool[i] })
	return pool[:1+g.rng.IntN(3)]
}

// hold opens the regime before or during 2026 and sometimes closes it.
func (g propGen) hold(r propRegime) usecase.RegimeHold {
	start := mustPeriod(propYear-1, time.July)
	if g.rng.IntN(3) == 0 {
		start = mustPeriod(propYear, time.Month(1+g.rng.IntN(12)))
	}
	var end domain.Period
	if g.rng.IntN(6) == 0 {
		first := time.January
		if start.Year() == propYear {
			first = start.Month()
		}
		end = mustPeriod(propYear, first+time.Month(g.rng.IntN(int(time.December-first)+1)))
	}
	return usecase.RegimeHold{Code: r.code, Treatment: r.treatment, Options: g.options(r), Start: start, End: end}
}

func holdCovers(h usecase.RegimeHold, m time.Month) bool {
	index := propYear*12 + int(m)
	startIndex := h.Start.Year()*12 + int(h.Start.Month())
	if index < startIndex {
		return false
	}
	return h.End.IsZero() || index <= h.End.Year()*12+int(h.End.Month())
}

// options are valid catalog options of the regime, defaults sometimes
// left unset.
func (g propGen) options(r propRegime) domain.Options {
	o := map[domain.OptionKey]domain.OptionValue{}
	if g.rng.IntN(2) == 0 {
		o[domain.OptionIVAFullyTaxable] = domain.BoolOption(g.rng.IntN(4) != 0)
	}
	switch r.code {
	case fiscal.RegimeRent:
		modes := []string{"blind_35", "actual"}
		if g.rng.IntN(4) != 0 {
			o[domain.OptionRentDeductionMode] = domain.TextOption(modes[g.rng.IntN(len(modes))])
		}
		if g.rng.IntN(10) == 0 {
			o[domain.OptionRentQuarterlyPayment] = domain.BoolOption(true)
		}
	case "621":
		o[domain.OptionRIFYear] = domain.TextOption(strconv.Itoa(1 + g.rng.IntN(rifYears)))
	case fiscal.RegimePlatforms:
		activities := []string{"transport_delivery", "lodging", "goods_services"}
		g.rng.Shuffle(len(activities), func(i, j int) { activities[i], activities[j] = activities[j], activities[i] })
		o["activities"] = domain.ListOption(activities[:1+g.rng.IntN(2)])
		o["definitive_option"] = domain.BoolOption(g.rng.IntN(2) == 0)
		if g.rng.IntN(2) == 0 {
			o["rfc_given_to_platform"] = domain.BoolOption(g.rng.IntN(5) != 0)
		}
	}
	return domain.NewOptions(o)
}

// maxMonthlyIncome is a realistic monthly ceiling: PF up to 250,000 pesos
// (past the RESICO PF monthly table now and then), PM up to 3,000,000.
func maxMonthlyIncome(r propRegime) money.Cents {
	if r.person == fiscal.PersonMoral {
		return 3_000_000 * propPeso
	}
	return 250_000 * propPeso
}

// share returns c × pct / 100, rounded down.
func share(c money.Cents, pct int64) money.Cents { return c * money.Cents(pct) / 100 }

func (g propGen) amount(ceiling money.Cents) money.Cents {
	if g.rng.IntN(10) == 0 {
		return 0
	}
	return money.Cents(g.rng.Int64N(int64(ceiling)))
}

// income splits the month's cash income by IVA rate and adds the
// withholdings a persona moral or a platform would make.
func (g propGen) income(r propRegime) domain.IncomeSums {
	cash := g.amount(maxMonthlyIncome(r))
	in := domain.IncomeSums{Cash: cash, Accrual: cash + share(cash, g.rng.Int64N(21)-10)}
	rest := cash
	if g.border {
		base := share(rest, g.rng.Int64N(101))
		in.ByRate.Rate8 = domain.RateSums{Base: base, IVA: share(base, 8)}
		rest -= base
	}
	if g.rng.IntN(4) == 0 {
		in.ByRate.Rate0Base = share(rest, g.rng.Int64N(51))
		rest -= in.ByRate.Rate0Base
	}
	if g.rng.IntN(5) == 0 {
		in.ByRate.ExemptBase = share(rest, g.rng.Int64N(51))
		rest -= in.ByRate.ExemptBase
	}
	in.ByRate.Rate16 = domain.RateSums{Base: rest, IVA: share(rest, 16)}
	g.withholdings(r, &in)
	return in
}

func (g propGen) withholdings(r propRegime, in *domain.IncomeSums) {
	charged := in.ByRate.Rate16.IVA + in.ByRate.Rate8.IVA
	switch {
	case r.code == fiscal.RegimePlatforms:
		in.IVAWithheld.ByPlatform = share(charged, 50)
		in.ISRWithheld.ByPlatform = share(in.Cash, 2+g.rng.Int64N(3))
	case r.person == fiscal.PersonFisica && g.rng.IntN(5) < 2:
		in.IVAWithheld.ByMoral = charged * 2 / 3
		in.ISRWithheld.ByMoral = share(in.ByRate.Rate16.Base, 10)
		if r.code == fiscal.RegimeRESICO {
			in.ISRWithheld.ByMoral = in.ByRate.Rate16.Base * 125 / 10_000
		}
	}
}

// expense draws deductible and other expenses below the income ceiling,
// their IVA, the 606 predial and the PM withholdings to remit.
func (g propGen) expense(r propRegime) domain.ExpenseSums {
	ceiling := share(maxMonthlyIncome(r), 60)
	deductible := g.amount(ceiling)
	e := domain.ExpenseSums{
		DeductiblePaid:    deductible,
		NonDeductiblePaid: g.amount(share(ceiling, 10)),
		CashOverLimitPaid: g.amount(share(ceiling, 2)),
	}
	e.IVAPaid.Rate16 = domain.RateSums{Base: deductible, IVA: share(deductible, 16)}
	creditable := 50 + g.rng.Int64N(51) // percent of the IVA that is creditable
	e.CreditableIVAPaid = share(e.IVAPaid.Rate16.IVA, creditable)
	if r.code == fiscal.RegimeRent && g.rng.IntN(3) == 0 {
		e.PredialPaid = g.amount(5_000 * propPeso)
	}
	if r.person == fiscal.PersonMoral && g.rng.IntN(5) < 2 {
		e.ISRWithheldToRemit = g.amount(share(ceiling, 2))
		e.IVAWithheldToRemit = min(g.amount(share(ceiling, 2)), e.IVAPaid.Rate16.IVA)
		// WH-09: the same creditable share applies to both parts (OQ-W5);
		// the paid part is credited this month, the withheld part the next
		// (LIVA 5 fr. IV).
		e.CreditableIVAPaid = share(e.IVAPaid.Rate16.IVA-e.IVAWithheldToRemit, creditable)
		e.CreditableIVAWithheld = share(e.IVAWithheldToRemit, creditable)
	}
	return e
}

// propInputKeys are the inputs each regime's definitions name (§2.8).
func propInputKeys(r propRegime) (monthly bool, optional []domain.InputKey) {
	switch {
	case r.code == "601":
		return true, []domain.InputKey{domain.InputPTUPaid, domain.InputPriorYearLosses, domain.InputInterestWithholding}
	case r.code == fiscal.RegimeBusiness:
		return true, []domain.InputKey{
			domain.InputYTDBeforeAppIncome, domain.InputYTDBeforeAppDeductions, domain.InputPTUPaid, domain.InputPriorYearLosses,
		}
	case r.code == fiscal.RegimeRESICO && r.person == fiscal.PersonMoral:
		return true, []domain.InputKey{
			domain.InputYTDBeforeAppIncome, domain.InputYTDBeforeAppDeductions, domain.InputPTUPaid,
			domain.InputPriorYearLosses, domain.InputInterestWithholding,
		}
	}
	return false, nil
}

// taxInputs enters some, all or none of the regime's inputs.
func (g propGen) taxInputs(r propRegime) []domain.TaxInputValue {
	var values []domain.TaxInputValue
	monthly, optional := propInputKeys(r)
	if monthly && g.rng.IntN(5) != 0 {
		for m := time.January; m <= time.November; m++ {
			if g.rng.IntN(4) != 0 {
				values = append(values, domain.TaxInputValue{
					Key: domain.InputPriorProvisionalPayments, Kind: domain.InputMoneyMonthly, Month: m,
					Money: g.amount(share(maxMonthlyIncome(r), 5)),
				})
			}
		}
	}
	for _, key := range optional {
		if g.rng.IntN(4) == 0 {
			values = append(values, domain.TaxInputValue{Key: key, Kind: domain.InputMoney, Money: g.amount(share(maxMonthlyIncome(r), 20))})
		}
	}
	switch r.code {
	case "601":
		values = append(values, g.generalPMInputs()...)
	case fiscal.RegimePlatforms:
		if g.rng.IntN(4) != 0 {
			values = append(values, domain.TaxInputValue{
				Key: domain.InputPriorYearIncome, Kind: domain.InputMoney, Money: g.amount(600_000 * propPeso),
			})
		}
	}
	return values
}

func (g propGen) generalPMInputs() []domain.TaxInputValue {
	var values []domain.TaxInputValue
	if g.rng.IntN(5) != 0 {
		values = append(values, domain.TaxInputValue{
			Key: domain.InputCoeficienteUtilidad, Kind: domain.InputRate, Rate: money.Rate(g.rng.Int64N(60 * propPercent)),
		})
	}
	if g.rng.IntN(3) == 0 {
		values = append(values, domain.TaxInputValue{
			Key: domain.InputFirstFiscalYear, Kind: domain.InputBoolean, Bool: g.rng.IntN(3) == 0,
		})
	}
	return values
}

func periodString(m time.Month) string { return mustPeriod(propYear, m).String() }

func mustPeriod(year int, m time.Month) domain.Period {
	p, err := domain.NewPeriod(year, m)
	if err != nil {
		panic(err)
	}
	return p
}
