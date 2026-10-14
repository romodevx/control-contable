package annual_test

import (
	"math/rand/v2"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/platform/money"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain/annual"
)

// The deductions property suite (PLAN ANN-07): applied ≤ registered,
// applied ≤ every cap, and monotonic in every amount and income. Each case
// is drawn from its own PCG stream (propSeed, index), so a failure names
// the index that reproduces it alone; no clock, no global random state.
const (
	propSeed  = 2026_1015
	propCases = 2000
	// propMaxRows and propMaxPesos keep the cases realistic: up to 12
	// rows of up to 300,000 pesos, incomes up to 5,000,000.
	propMaxRows   = 12
	propMaxPesos  = 300_000
	propMaxIncome = 5_000_000
	propPeso      = money.Cents(100)
)

var propCategories = []annual.Category{
	annual.CategoryMedical, annual.CategoryMedicalDisability, annual.CategoryFuneral, annual.CategoryDonation,
	annual.CategoryMortgageRealInterest, annual.CategoryRetirementContribution, annual.CategoryHealthInsurance,
	annual.CategorySchoolTransport, annual.CategoryLocalPayrollTax, annual.CategoryTuition,
}

var propLevels = []annual.TuitionLevel{
	annual.LevelPreescolar, annual.LevelPrimaria, annual.LevelSecundaria, annual.LevelProfesionalTecnico,
	annual.LevelBachillerato,
}

func propRand(i int) *rand.Rand {
	return rand.New(rand.NewPCG(propSeed, uint64(i))) //nolint:gosec // a fixed, reproducible seed is the point
}

func pick[T any](r *rand.Rand, xs []T) T { return xs[r.IntN(len(xs))] }

func amount(r *rand.Rand, maxPesos int64) money.Cents {
	return money.Cents(r.Int64N(maxPesos*int64(propPeso)) + 1)
}

func genDeduction(r *rand.Rand) annual.Deduction {
	d := annual.Deduction{
		Category: pick(r, propCategories), Amount: amount(r, propMaxPesos),
		PaymentMethod: pick(r, []annual.PaymentMethod{annual.PaymentElectronic, annual.PaymentElectronic, annual.PaymentCash}),
		Beneficiary:   pick(r, []annual.Beneficiary{annual.BeneficiarySelf, annual.BeneficiaryDescendant}),
	}
	switch d.Category {
	case annual.CategoryDonation:
		d.DoneeType = pick(r, []annual.DoneeType{annual.DoneeAuthorized, annual.DoneeGovernment})
	case annual.CategoryTuition:
		d.Student = pick(r, []string{"Ana", "Luis", "Eva"})
		d.TuitionLevel = pick(r, propLevels)
		if r.IntN(4) == 0 {
			d.Scholarship = amount(r, propMaxPesos/10)
		}
	}
	return d
}

func genInput(r *rand.Rand) annual.DeductionsInput {
	rows := make([]annual.Deduction, r.IntN(propMaxRows+1))
	for i := range rows {
		rows[i] = genDeduction(r)
	}
	acc := amount(r, propMaxIncome)
	in := annual.DeductionsInput{
		FiscalYear: fiscalYear, Deductions: rows,
		AccumulableIncome: acc, TotalIncome: acc + amount(r, propMaxIncome/10),
	}
	if r.IntN(5) > 0 {
		prior := amount(r, propMaxIncome)
		in.PriorYearAccumulableIncome = &prior
	}
	return in
}

func forEachCase(t *testing.T, check func(t *testing.T, r *rand.Rand, in annual.DeductionsInput, got annual.Deductions)) {
	t.Helper()
	calc := calculator(t)
	for i := range propCases {
		r := propRand(i)
		in := genInput(r)
		got, err := calc.Compute(in)
		require.NoError(t, err, "case %d", i)
		t.Run(strconv.Itoa(i), func(t *testing.T) { check(t, r, in, got) })
		if t.Failed() {
			t.Fatalf("property failed at case %d (seed %d)", i, propSeed)
		}
	}
}

func TestPropertyAppliedNeverExceedsClaimed(t *testing.T) {
	forEachCase(t, func(t *testing.T, _ *rand.Rand, _ annual.DeductionsInput, got annual.Deductions) {
		require.LessOrEqual(t, got.Applied, got.Claimed)
		require.GreaterOrEqual(t, got.Applied, money.Cents(0))
		require.Equal(t, got.Applied, got.CappedApplied+got.UncappedApplied)
		var sum money.Cents
		for _, c := range got.Categories {
			require.LessOrEqual(t, c.Applied, c.Eligible, c.Category)
			require.LessOrEqual(t, c.Eligible, c.Claimed, c.Category)
			require.GreaterOrEqual(t, c.Applied, money.Cents(0), c.Category)
			sum += c.Applied
		}
		require.Equal(t, got.Applied, sum)
	})
}

func TestPropertyAppliedNeverExceedsACap(t *testing.T) {
	limits := limits2026(t)
	uma := limits.AnnualUMA
	forEachCase(t, func(t *testing.T, _ *rand.Rand, in annual.DeductionsInput, got annual.Deductions) {
		globalCap := min(5*uma, percentOf(t, in.TotalIncome, "0.15"))
		require.LessOrEqual(t, got.CappedApplied, globalCap)
		applied := map[annual.Category]money.Cents{}
		for _, c := range got.Categories {
			applied[c.Category] = c.Applied
		}
		require.LessOrEqual(t, applied[annual.CategoryFuneral], uma)
		require.LessOrEqual(t, applied[annual.CategoryRetirementContribution],
			min(5*uma, percentOf(t, in.AccumulableIncome, "0.10")))
		if in.PriorYearAccumulableIncome == nil {
			require.Zero(t, applied[annual.CategoryDonation])
		} else {
			require.LessOrEqual(t, applied[annual.CategoryDonation], percentOf(t, *in.PriorYearAccumulableIncome, "0.07"))
		}
		require.LessOrEqual(t, applied[annual.CategoryTuition], tuitionCeiling(in, limits))
	})
}

// tuitionCeiling is Σ over students of the highest cap of their levels.
func tuitionCeiling(in annual.DeductionsInput, l annual.Limits) money.Cents {
	caps := map[string]money.Cents{}
	for _, d := range in.Deductions {
		if d.Category == annual.CategoryTuition {
			caps[d.Student] = max(caps[d.Student], l.TuitionCaps[d.TuitionLevel])
		}
	}
	var total money.Cents
	for _, c := range caps {
		total += c
	}
	return total
}

func percentOf(t *testing.T, c money.Cents, r string) money.Cents {
	t.Helper()
	exact, err := c.MulRate(rate(t, r))
	require.NoError(t, err)
	return exact.Round()
}

// TestPropertyMonotonic raises one row's amount, adds a row, or raises an
// income: the applied total never goes down.
func TestPropertyMonotonic(t *testing.T) {
	calc := calculator(t)
	forEachCase(t, func(t *testing.T, r *rand.Rand, in annual.DeductionsInput, got annual.Deductions) {
		more := grow(r, in)

		after, err := calc.Compute(more)

		require.NoError(t, err)
		require.GreaterOrEqual(t, after.Applied, got.Applied)
	})
}

// grow returns a copy of in with one quantity increased.
func grow(r *rand.Rand, in annual.DeductionsInput) annual.DeductionsInput {
	out := in
	out.Deductions = append([]annual.Deduction(nil), in.Deductions...)
	bump := amount(r, propMaxPesos)
	switch r.IntN(5) {
	case 0:
		if len(out.Deductions) > 0 {
			out.Deductions[r.IntN(len(out.Deductions))].Amount += bump
			return out
		}
		fallthrough
	case 1:
		out.Deductions = append(out.Deductions, genDeduction(r))
	case 2:
		out.AccumulableIncome += bump
		out.TotalIncome += bump
	case 3:
		out.TotalIncome += bump
	default:
		if in.PriorYearAccumulableIncome != nil {
			prior := *in.PriorYearAccumulableIncome + bump
			out.PriorYearAccumulableIncome = &prior
		}
	}
	return out
}
