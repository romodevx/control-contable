package taxengine_test

import (
	"context"
	"flag"
	"maps"
	"math/rand/v2"
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/clock"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
	"github.com/romodevx/control-contable/backend/internal/taxengine"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain/annual"
	"github.com/romodevx/control-contable/backend/internal/taxengine/infra/annuallimits"
	"github.com/romodevx/control-contable/backend/internal/taxengine/usecase"
	usecaseannual "github.com/romodevx/control-contable/backend/internal/taxengine/usecase/annual"
)

// The generators of the annual property suite (ANN-16): a taxpayer's
// 2026 drawn by the engine suite's generator (genCase: holds, monthly
// ledger and tax inputs of every PF and PM regime), plus what only the
// annual reads — salary constancias, personal deductions, the per-year
// facts, filed payments and the annual tax inputs — and the day the
// estimate is asked (the year closed, or in progress). The default
// payments come from the real monthly Estimator (EstimateRange), so the
// whole annual path runs: only the ports are in memory. Every case is
// drawn from its own PCG stream (annPropSeed, index), so a failure names
// the index that reproduces it alone.

const (
	// annPropSeed is the suite's fixed seed: changing it is a reviewed
	// change.
	annPropSeed = 2026_1016
	// annPropDefaultCases keeps the default CI run near a second per
	// property (each case is a year of monthly estimates);
	// -taxengine.annual.property.cases raises it for a nightly run.
	annPropDefaultCases = 1000
	// annMaxDeductions and annMaxStatements are realistic list lengths.
	annMaxDeductions = 8
	annMaxStatements = 3
)

var annPropCases = flag.Int("taxengine.annual.property.cases", annPropDefaultCases,
	"cases per annual property (nightly: 10000)")

// annClosedToday is 15 January 2027: 2026 is closed.
var annClosedToday = time.Date(propYear+1, time.January, 15, 12, 0, 0, 0, clock.Location())

// annSections is the composition root's section list over the embedded
// parameters, built once.
var annSections = sync.OnceValues(func() (usecaseannual.SectionRegistry, error) {
	_, cat := propEngine()
	return usecaseannual.NewSectionRegistry(taxengine.AnnualSections(annuallimits.New(cat))...)
})

// annCase is one taxpayer's fiscal year as the annual estimate reads it.
type annCase struct {
	propCase
	today time.Time
	year  usecaseannual.YearInputs
	filed map[fiscal.RegimeCode]map[time.Month]money.Cents
}

// clone deep-copies what a property mutates.
func (c annCase) clone() annCase {
	c.propCase = c.propCase.clone()
	c.year.Deductions = slices.Clone(c.year.Deductions)
	c.year.SalaryStatements = slices.Clone(c.year.SalaryStatements)
	filed := make(map[fiscal.RegimeCode]map[time.Month]money.Cents, len(c.filed))
	for code, months := range c.filed {
		filed[code] = maps.Clone(months)
	}
	c.filed = filed
	return c
}

// estimator wires the annual use case over the case's ports, with the
// real monthly Estimator as its range and the same clock for both.
func (c annCase) estimator(t testing.TB) *usecaseannual.Estimator {
	t.Helper()
	registry, cat := propEngine()
	sections, err := annSections()
	require.NoError(t, err)
	inputs := txInputs{}
	for code, values := range c.inputs {
		in, err := domain.NewTaxInputs(values...)
		require.NoError(t, err, "case %d inputs %s", c.index, code)
		inputs[code] = in
	}
	ports := annPorts{c: c}
	profiles := txProfiles{usecase.TaxpayerProfile{PersonType: c.person, Regimes: c.holds}}
	monthly, err := usecase.NewEstimator(usecase.EstimateDeps{
		Tx: txUnit{}, Profiles: profiles, Ledger: ports, Inputs: inputs, Texts: txNoTexts{}, Holidays: txNoHolidays{},
		Strategies: registry, Params: cat, Clock: clock.NewFixed(c.today),
	})
	require.NoError(t, err)
	est, err := usecaseannual.NewEstimator(usecaseannual.EstimateDeps{
		Tx: txUnit{}, Profiles: profiles, Ledger: ports, TaxInputs: inputs, YearInputs: ports, Filed: ports,
		Monthly: monthly, MonthlyStrategies: registry, Sections: sections, Holidays: txNoHolidays{},
		Params: cat, Clock: clock.NewFixed(c.today),
	})
	require.NoError(t, err)
	return est
}

func (c annCase) estimate(t testing.TB) usecaseannual.Estimate {
	t.Helper()
	got, err := c.estimator(t).Estimate(t.Context(), txUser, txTaxpayer, propYear)
	require.NoError(t, err, "case %d (%s, today %s)", c.index, c.person, c.today.Format(time.DateOnly))
	return got
}

// annPorts serves the case's ledger (up to the query's month, as the
// store does), annual inputs and filed payments.
type annPorts struct{ c annCase }

func (p annPorts) IncomeSums(_ context.Context, q usecase.LedgerQuery) ([]usecase.IncomeMonthSums, error) {
	var out []usecase.IncomeMonthSums
	for _, r := range p.c.incomes {
		if r.Month <= q.Through {
			r.Year = q.FiscalYear
			out = append(out, r)
		}
	}
	return out, nil
}

func (p annPorts) ExpenseSums(_ context.Context, q usecase.LedgerQuery) ([]usecase.ExpenseMonthSums, error) {
	var out []usecase.ExpenseMonthSums
	for _, r := range p.c.expenses {
		if r.Month <= q.Through {
			r.Year = q.FiscalYear
			out = append(out, r)
		}
	}
	return out, nil
}

func (p annPorts) YearInputs(context.Context, string, int) (usecaseannual.YearInputs, error) {
	return p.c.year, nil
}

func (p annPorts) FiledPayments(context.Context, string, int) (map[fiscal.RegimeCode]map[time.Month]money.Cents, error) {
	return p.c.filed, nil
}

// forEachAnnCase runs check on every generated case, each from its own
// stream, and stops at the first failing one.
func forEachAnnCase(t *testing.T, check func(t *testing.T, c annCase)) {
	t.Helper()
	for i := range *annPropCases {
		check(t, genAnnCase(i))
		if t.Failed() {
			t.Fatalf("first failing case: %d (seed %d; rerun with -run %s)", i, annPropSeed, t.Name())
		}
	}
}

// annRand is case index's stream; extra separates a property's own draws
// from the case's.
func annRand(index int, extra uint64) *rand.Rand {
	return rand.New(rand.NewPCG(annPropSeed, uint64(index)+extra<<32)) //nolint:gosec // a fixed, reproducible seed is the point
}

func genAnnCase(index int) annCase {
	rng := annRand(index, 0)
	g := annGen{propGen: propGen{rng: rng, border: rng.IntN(5) == 0}}
	c := annCase{propCase: genCase(rng, index), today: annClosedToday, filed: map[fiscal.RegimeCode]map[time.Month]money.Cents{}}
	if rng.IntN(4) == 0 {
		c.today = time.Date(propYear, time.Month(1+rng.IntN(12)), 15, 12, 0, 0, 0, clock.Location())
	}
	for _, h := range c.holds {
		c.inputs[h.Code] = append(c.inputs[h.Code], g.annualTaxInputs(h.Code, c.person)...)
		if months := g.filed(h); len(months) > 0 {
			c.filed[h.Code] = months
		}
	}
	if c.person == fiscal.PersonFisica {
		c.year = g.yearInputs()
	}
	return c
}

// annGen draws what only the annual reads.
type annGen struct{ propGen }

// annualTaxInputs adds the annual depreciation of a business regime now
// and then (601, 612, 625), the only annual-only tax input.
func (g annGen) annualTaxInputs(code fiscal.RegimeCode, person fiscal.PersonType) []domain.TaxInputValue {
	business := code == "601" || code == fiscal.RegimeBusiness || code == fiscal.RegimePlatforms
	if !business || g.rng.IntN(3) != 0 {
		return nil
	}
	ceiling := 200_000 * propPeso
	if person == fiscal.PersonMoral {
		ceiling = 2_000_000 * propPeso
	}
	return []domain.TaxInputValue{{Key: annual.InputAnnualDepreciation, Kind: domain.InputMoney, Money: g.amount(ceiling)}}
}

// filed confirms none, some or all of a regime's held months.
func (g annGen) filed(h usecase.RegimeHold) map[time.Month]money.Cents {
	share := g.rng.IntN(4) // 0: none, 3: every month
	if share == 0 {
		return nil
	}
	out := map[time.Month]money.Cents{}
	for m := time.January; m <= time.December; m++ {
		if holdCovers(h, m) && g.rng.IntN(3) < share {
			out[m] = g.amount(20_000 * propPeso)
		}
	}
	return out
}

// yearInputs draws a persona física's constancias, deductions and facts.
func (g annGen) yearInputs() usecaseannual.YearInputs {
	var y usecaseannual.YearInputs
	for range g.rng.IntN(annMaxStatements + 1) {
		y.SalaryStatements = append(y.SalaryStatements, g.statement())
	}
	for range g.rng.IntN(annMaxDeductions + 1) {
		y.Deductions = append(y.Deductions, g.deduction())
	}
	if g.rng.IntN(5) > 0 {
		prior := g.amount(3_000_000 * propPeso)
		y.Facts.PriorYearAccumulableIncome = &prior
	}
	if g.rng.IntN(3) == 0 {
		y.Facts.OtherExemptIncome = g.amount(300_000 * propPeso)
	}
	y.Facts.ToldEmployerWillFile = g.rng.IntN(6) == 0
	return y
}

// statement is one constancia: up to 1,500,000 accumulable, with the ISR
// an employer withholds (up to 35 %).
func (g annGen) statement() annual.SalaryStatement {
	accumulable := g.amount(1_500_000 * propPeso)
	return annual.SalaryStatement{
		Accumulable: accumulable, Exempt: g.amount(100_000 * propPeso),
		ISRWithheld:         share(accumulable, g.rng.Int64N(36)),
		Simultaneous:        g.rng.IntN(3) == 0,
		LeftBeforeYearEnd:   g.rng.IntN(5) == 0,
		PayerNotWithholding: g.rng.IntN(10) == 0,
	}
}

var annCategories = []annual.Category{
	annual.CategoryMedical, annual.CategoryMedicalDisability, annual.CategoryFuneral, annual.CategoryDonation,
	annual.CategoryMortgageRealInterest, annual.CategoryRetirementContribution, annual.CategoryHealthInsurance,
	annual.CategorySchoolTransport, annual.CategoryLocalPayrollTax, annual.CategoryTuition,
}

var annLevels = []annual.TuitionLevel{
	annual.LevelPreescolar, annual.LevelPrimaria, annual.LevelSecundaria, annual.LevelProfesionalTecnico, annual.LevelBachillerato,
}

func annPick[T any](rng *rand.Rand, xs []T) T { return xs[rng.IntN(len(xs))] }

// deduction is one valid personal deduction of up to 150,000 pesos, paid
// in cash now and then (then ineligible when the category forbids it).
func (g annGen) deduction() annual.Deduction {
	d := annual.Deduction{
		Category: annPick(g.rng, annCategories), Amount: 1 + g.amount(150_000*propPeso),
		PaymentMethod: annPick(g.rng, []annual.PaymentMethod{annual.PaymentElectronic, annual.PaymentElectronic, annual.PaymentCash}),
		Beneficiary:   annPick(g.rng, []annual.Beneficiary{annual.BeneficiarySelf, annual.BeneficiarySpouse, annual.BeneficiaryDescendant}),
	}
	switch d.Category {
	case annual.CategoryDonation:
		d.DoneeType = annPick(g.rng, []annual.DoneeType{annual.DoneeAuthorized, annual.DoneeGovernment})
	case annual.CategoryTuition:
		d.Beneficiary = annual.BeneficiaryDescendant
		d.Student = annPick(g.rng, []string{"Ana", "Luis", "Eva"})
		d.TuitionLevel = annPick(g.rng, annLevels)
		if g.rng.IntN(4) == 0 {
			d.Scholarship = g.amount(15_000 * propPeso)
		}
	}
	return d
}
