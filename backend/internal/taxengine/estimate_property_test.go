package taxengine_test

import (
	"bytes"
	"encoding/json"
	"math/rand/v2"
	"slices"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain"
	"github.com/romodevx/control-contable/backend/internal/taxengine/usecase"
)

// The engine property suite (M5-24, MODULE-tax-engine §2.7, §2.11, ADR-0002
// §5): invariants of the merged EstimateMonth use case over every
// registered strategy, on generated ledgers, fixed seed. Generators:
// estimate_property_gen_test.go.

// --- (1) lines + rounding adjustment = result, per regime and totals ---

// TestEstimatePropertyLinesReconcile: in every breakdown (each regime's and
// the consolidated IVA's) the shown lines since the last anchor, plus the
// rounding adjustment, add up to each result line; each tax figure is its
// result line; the totals are the sums of the regimes' figures.
func TestEstimatePropertyLinesReconcile(t *testing.T) {
	forEachCase(t, func(t *testing.T, c propCase) {
		got := c.estimate(t)
		for _, r := range got.Regimes {
			label := r.Key.String()
			assertLinesReconcile(t, c.index, label, r.Lines)
			assertISRIsItsResultLine(t, c.index, r)
			assertIVANetIsItsResultLine(t, c.index, r)
			assertToRemitIsItsLine(t, c.index, r)
		}
		assertLinesReconcile(t, c.index, "consolidated IVA", got.IVA.Lines)
		assertTotalsAddUp(t, c.index, got)
	})
}

// assertLinesReconcile replays a breakdown with centavos only. Input and
// lookup lines anchor a new running value, and so does a multiply that does
// not follow a lookup (Ledger.Scale); a multiply right after a lookup is a
// tariff excess (Ledger.AddProduct) and adds.
func assertLinesReconcile(t *testing.T, index int, label string, lines []domain.Line) {
	t.Helper()
	var (
		shown   money.Cents
		rounded int // lines whose rounding the next result may absorb
		prev    domain.Operation
	)
	for i, l := range lines {
		assert.Equal(t, l.Exact.Round(), l.Amount, "case %d %s line %d %s: shown amount is the exact one rounded", index, label, l.Seq, l.Code)
		switch l.Operation {
		case domain.OpInput, domain.OpLookup:
			shown, rounded = l.Amount, 1
		case domain.OpMultiply:
			if prev == domain.OpLookup {
				shown, rounded = shown+l.Amount, rounded+1
			} else {
				shown, rounded = l.Amount, 1
			}
		case domain.OpAdd:
			shown, rounded = shown+l.Amount, rounded+1
		case domain.OpSubtract:
			shown, rounded = shown-l.Amount, rounded+1
		case domain.OpRoundingAdjustment:
			require.Less(t, i+1, len(lines), "case %d %s: adjustment is the last line", index, label)
			assert.Equal(t, domain.OpResult, lines[i+1].Operation, "case %d %s line %d: adjustment precedes a result", index, label, l.Seq)
			assert.LessOrEqual(t, max(l.Amount, -l.Amount), money.Cents(rounded),
				"case %d %s line %d: adjustment beyond one centavo per rounded line", index, label, l.Seq)
			shown += l.Amount
		case domain.OpResult:
			assert.Equal(t, l.Amount, shown, "case %d %s line %d %s: shown lines + adjustment = result", index, label, l.Seq, l.Code)
			shown, rounded = l.Amount, 1
		case domain.OpInfo:
			continue // a fact, no arithmetic
		default:
			t.Errorf("case %d %s line %d: unknown operation %q", index, label, l.Seq, l.Operation)
		}
		prev = l.Operation
	}
}

// lastResult is the last result line of tax, if any.
func lastResult(lines []domain.Line, tax domain.TaxKind) (domain.Line, bool) {
	for _, l := range slices.Backward(lines) {
		if l.Tax == tax && l.Operation == domain.OpResult {
			return l, true
		}
	}
	return domain.Line{}, false
}

func assertISRIsItsResultLine(t *testing.T, index int, r usecase.RegimeResult) {
	t.Helper()
	if r.ISR == nil {
		return
	}
	assert.Equal(t, r.ISR.Exact.Round(), r.ISR.Amount, "case %d %s: ISR rounded once", index, r.Key)
	line, ok := lastResult(r.Lines, domain.TaxISR)
	if !ok {
		assert.Zero(t, r.ISR.Amount, "case %d %s: an ISR figure without an ISR result line", index, r.Key)
		return
	}
	assert.Equal(t, r.ISR.Amount, line.Amount, "case %d %s: ISR = its last result line", index, r.Key)
}

// ivaNet is a regime's IVA a cargo: charged − definitive − creditable −
// withheld from the taxpayer (LIVA 5-D, §2.6).
func ivaNet(c domain.IVAComponents) money.Cents {
	return c.ChargedCollected - c.DefinitiveExcluded - c.CreditablePaid - c.WithheldFromTaxpayer
}

func assertIVANetIsItsResultLine(t *testing.T, index int, r usecase.RegimeResult) {
	t.Helper()
	if r.IVA == nil {
		return
	}
	line, ok := lastResult(r.Lines, domain.TaxIVA)
	require.True(t, ok, "case %d %s: IVA components without an IVA result line", index, r.Key)
	assert.Equal(t, ivaNet(*r.IVA), line.Amount, "case %d %s: the regime's IVA net = its components", index, r.Key)
}

// assertToRemitIsItsLine: the withholdings to remit are the last
// withholding line (the input of a single one, or their result).
func assertToRemitIsItsLine(t *testing.T, index int, r usecase.RegimeResult) {
	t.Helper()
	var last *domain.Line
	for i := range r.Lines {
		if r.Lines[i].Tax == domain.TaxWithholding {
			last = &r.Lines[i]
		}
	}
	if last == nil {
		assert.Zero(t, r.WithheldToRemit, "case %d %s: withholdings to remit without a line", index, r.Key)
		return
	}
	assert.Equal(t, r.WithheldToRemit, last.Amount, "case %d %s: to remit = its last line", index, r.Key)
}

func assertTotalsAddUp(t *testing.T, index int, got usecase.TaxpayerEstimate) {
	t.Helper()
	var isr, toRemit money.Cents
	for _, r := range got.Regimes {
		if r.ISR != nil {
			isr += r.ISR.Payable()
		}
		toRemit += r.WithheldToRemit
	}
	last, ok := lastResult(got.IVA.Lines, domain.TaxIVA)
	if ok {
		assert.Equal(t, got.IVA.Result.Amount, last.Amount, "case %d: consolidated IVA = its result line", index)
	} else {
		assert.Zero(t, got.IVA.Result.Amount, "case %d: an IVA figure without lines", index)
	}
	tot := got.Totals
	assert.Equal(t, got.IVA.Result, tot.IVA, "case %d: totals IVA = consolidated IVA", index)
	assert.Equal(t, isr, tot.ISR.Amount, "case %d: totals ISR = Σ payable regime ISR", index)
	assert.Equal(t, toRemit, tot.WithholdingsToRemit, "case %d: totals to remit = Σ regimes", index)
	assert.Equal(t, tot.ISR.Payable()+tot.IVA.Payable()+tot.WithholdingsToRemit, tot.TotalPayable,
		"case %d: total = ISR + IVA payable + to remit", index)
}

// --- (2) ISR monotonic non-decreasing in income ---

// isrDropBound is how far a regime's ISR may go DOWN when its income goes
// up: zero, except across a row edge of a published progressive tariff,
// whose fixed fees are not chained to the centavo (params
// tariff_continuity_2026_test.go, PR #76 notes). The bound is that test's
// documented drift: a centavo, times the month for the Jan-Nov accumulated
// art. 106 tables, two for the RIF bimonthly table (before its reduction).
func isrDropBound(key domain.StrategyKey, month time.Month) money.Cents {
	switch key.Regime {
	case fiscal.RegimeBusiness:
		if month == time.December {
			return 1
		}
		return money.Cents(month)
	case fiscal.RegimeRent:
		return 1
	case "621":
		return rifBimonthlyDrift
	}
	return 0
}

// rifBimonthlyDrift: B.VII is published as two monthly tables.
const rifBimonthlyDrift = 2

// TestEstimatePropertyISRIsMonotonicInIncome: more income in the period's
// month, everything else fixed, never lowers a regime's ISR beyond the
// documented tariff drift. An ISR that stops being estimated (RESICO PF
// above its monthly table: partial) has nothing to compare.
func TestEstimatePropertyISRIsMonotonicInIncome(t *testing.T) {
	var compared, drops int
	forEachCase(t, func(t *testing.T, c propCase) {
		before := c.estimate(t)
		rng := rand.New(rand.NewPCG(propSeed, uint64(c.index)+1<<32)) //nolint:gosec // fixed seed
		month := c.month(t)
		for _, r := range before.Regimes {
			if r.ISR == nil {
				continue
			}
			bumped := c.withIncomeBump(r.Key.Regime, month, propDelta(rng))
			after, ok := regimeOf(bumped.estimate(t), r.Key)
			require.True(t, ok)
			if after.ISR == nil {
				continue
			}
			compared++
			drop := r.ISR.Amount - after.ISR.Amount
			if drop > 0 {
				drops++
				t.Logf("case %d %s %s: ISR %s -> %s with more income (documented tariff drift)", c.index, r.Key, c.period, r.ISR.Amount, after.ISR.Amount)
			}
			assert.LessOrEqual(t, drop, isrDropBound(r.Key, month),
				"case %d %s %s: more income lowered the ISR %s -> %s", c.index, r.Key, c.period, r.ISR.Amount, after.ISR.Amount)
		}
	})
	t.Logf("compared %d regime ISRs, %d within-drift drops", compared, drops)
	require.Positive(t, compared)
}

// propDelta is a centavo now and then (row edges), else up to 50,000 pesos.
func propDelta(rng *rand.Rand) money.Cents {
	if rng.IntN(4) == 0 {
		return money.Cents(1 + rng.Int64N(100))
	}
	return money.Cents(1 + rng.Int64N(int64(50_000*propPeso)))
}

// withIncomeBump adds delta to the regime's cash and accrual income of
// month m (a new row if it had none): the base every ISR reads.
func (c propCase) withIncomeBump(code fiscal.RegimeCode, m time.Month, delta money.Cents) propCase {
	out := c.clone()
	for i, row := range out.incomes {
		if row.Regime == code && row.Month == m {
			out.incomes[i].Sums.Cash += delta
			out.incomes[i].Sums.Accrual += delta
			return out
		}
	}
	out.incomes = append(out.incomes, usecase.IncomeMonthSums{
		Regime: code, Month: m, Sums: domain.IncomeSums{Cash: delta, Accrual: delta},
	})
	return out
}

func regimeOf(e usecase.TaxpayerEstimate, key domain.StrategyKey) (usecase.RegimeResult, bool) {
	for _, r := range e.Regimes {
		if r.Key == key {
			return r, true
		}
	}
	return usecase.RegimeResult{}, false
}

// --- (3) byte-identical JSON for identical inputs ---

// propJSON is the canonical encoding of an estimate: encoding/json over
// the domain result (fields in declaration order, map keys sorted), plus
// the two values whose fields are unexported. There is no HTTP estimate DTO
// yet (M5-23); swap this for it when it lands.
type propJSON struct {
	Period   string
	DueDate  string
	Estimate usecase.TaxpayerEstimate
}

func canonicalJSON(t *testing.T, e usecase.TaxpayerEstimate) []byte {
	t.Helper()
	out, err := json.Marshal(propJSON{Period: e.Period.String(), DueDate: e.DueDate.String(), Estimate: e})
	require.NoError(t, err)
	return out
}

// TestEstimatePropertyIdenticalInputsGiveByteIdenticalJSON: the same
// records, inputs and parameters give the same bytes, from a fresh
// estimator, and whatever order the readers return their rows, the regimes
// and the input values in (no map- or row-order dependence, §2.11).
func TestEstimatePropertyIdenticalInputsGiveByteIdenticalJSON(t *testing.T) {
	forEachCase(t, func(t *testing.T, c propCase) {
		first := canonicalJSON(t, c.estimate(t))
		again := canonicalJSON(t, c.clone().estimate(t))
		rng := rand.New(rand.NewPCG(propSeed, uint64(c.index)+2<<32)) //nolint:gosec // fixed seed
		shuffled := canonicalJSON(t, c.shuffled(rng).estimate(t))

		require.Contains(t, string(first), `"Lines"`, "case %d: the encoding carries the breakdown", c.index)
		assert.True(t, bytes.Equal(first, again), "case %d: same input, different JSON", c.index)
		assert.True(t, bytes.Equal(first, shuffled), "case %d: reordered rows, different JSON", c.index)
	})
}

// shuffled returns the same case with every list in another order.
func (c propCase) shuffled(rng *rand.Rand) propCase {
	out := c.clone()
	shuffle(rng, out.holds)
	shuffle(rng, out.incomes)
	shuffle(rng, out.expenses)
	for _, values := range out.inputs {
		shuffle(rng, values)
	}
	return out
}

func shuffle[T any](rng *rand.Rand, s []T) {
	rng.Shuffle(len(s), func(i, j int) { s[i], s[j] = s[j], s[i] })
}

// --- (4) withholdings to remit are never the taxpayer's own taxes ---

// TestEstimatePropertyWithholdingsToRemitNeverEnterOwnTaxes: removing every
// withholding the taxpayer made to its suppliers changes nothing but the
// withholdings to remit and the total, by exactly that amount; each
// regime's ISR and IVA, the consolidated IVA and the totals' ISR and IVA
// stay put (ADR-0002 §1, M5-22).
func TestEstimatePropertyWithholdingsToRemitNeverEnterOwnTaxes(t *testing.T) {
	forEachCase(t, func(t *testing.T, c propCase) {
		with := c.estimate(t)
		without := c.withoutWithholdingsToRemit().estimate(t)

		require.Len(t, without.Regimes, len(with.Regimes))
		for i, r := range with.Regimes {
			w := without.Regimes[i]
			assert.Equal(t, r.ISR, w.ISR, "case %d %s: ISR moved with the withholdings to remit", c.index, r.Key)
			assert.Equal(t, r.IVA, w.IVA, "case %d %s: IVA moved with the withholdings to remit", c.index, r.Key)
			assert.Zero(t, w.WithheldToRemit, "case %d %s: to remit without any withholding", c.index, r.Key)
		}
		assert.Equal(t, with.IVA.Result, without.IVA.Result, "case %d: consolidated IVA", c.index)
		assert.Equal(t, with.Totals.ISR, without.Totals.ISR, "case %d: totals ISR", c.index)
		assert.Equal(t, with.Totals.IVA, without.Totals.IVA, "case %d: totals IVA", c.index)
		assert.Zero(t, without.Totals.WithholdingsToRemit, "case %d", c.index)
		assert.Equal(t, with.Totals.WithholdingsToRemit, with.Totals.TotalPayable-without.Totals.TotalPayable,
			"case %d: the total moves by exactly the withholdings to remit", c.index)
	})
}

func (c propCase) withoutWithholdingsToRemit() propCase {
	out := c.clone()
	for i := range out.expenses {
		out.expenses[i].Sums.ISRWithheldToRemit = 0
		out.expenses[i].Sums.IVAWithheldToRemit = 0
	}
	return out
}

// --- (5) IVA netting never carries forward ---

// TestEstimatePropertyIVANeverCarriesForward: the consolidated IVA is the
// period's own components netted (OQ-T2), nothing from earlier months
// (OQ-T5): a previous favor balance is an info line only, and making every
// earlier month hugely in favor leaves the period's IVA unchanged. The
// bimonthly 621 reads its bimester's first month by law, so its earlier
// months are left alone.
func TestEstimatePropertyIVANeverCarriesForward(t *testing.T) {
	forEachCase(t, func(t *testing.T, c propCase) {
		got := c.estimate(t)
		var net money.Cents
		for _, r := range got.Regimes {
			if r.IVA != nil {
				net += ivaNet(*r.IVA)
			}
		}
		assert.Equal(t, net, got.IVA.Result.Amount, "case %d: consolidated IVA = Σ the period's regime nets", c.index)
		for _, l := range got.IVA.Lines {
			if l.Code == domain.LineIVAPreviousInFavor {
				assert.Equal(t, domain.OpInfo, l.Operation, "case %d: the previous favor balance is shown, not netted", c.index)
			}
		}

		inFavorBefore := c.withEarlierMonthsInFavor(c.month(t)).estimate(t)
		assert.Equal(t, got.IVA.Result, inFavorBefore.IVA.Result, "case %d: earlier favor balances moved the IVA", c.index)
		assert.Equal(t, got.Totals.IVA, inFavorBefore.Totals.IVA, "case %d", c.index)
	})
}

// earlierMonthCreditable makes any earlier month an IVA favor balance.
const earlierMonthCreditable = 50_000_000 * propPeso

func (c propCase) withEarlierMonthsInFavor(m time.Month) propCase {
	out := c.clone()
	for i, row := range out.expenses {
		if row.Month < m && row.Regime != "621" {
			out.expenses[i].Sums.CreditableIVAPaid += earlierMonthCreditable
		}
	}
	return out
}

// --- (6) incomplete whenever required inputs are missing ---

// TestEstimatePropertyIncompleteWhenRequiredInputsMissing checks every
// regime's missing inputs against an oracle written from the §2.8 table,
// not from the strategies' RequiredInputs, and the status precedence
// informational > incomplete > partial > complete.
func TestEstimatePropertyIncompleteWhenRequiredInputsMissing(t *testing.T) {
	var incomplete int
	forEachCase(t, func(t *testing.T, c propCase) {
		got := c.estimate(t)
		month := c.month(t)
		for _, r := range got.Regimes {
			hold := c.holdOf(r.Key.Regime)
			want := expectedMissing(r.Key, hold, month, c.inputValues(r.Key.Regime))
			informational := r.Treatment == domain.TreatmentInformational || r.Coverage == domain.TreatmentInformational
			switch {
			case informational:
				assert.Equal(t, domain.StatusInformational, r.Status, "case %d %s", c.index, r.Key)
			case len(want) > 0:
				incomplete++
				assert.Equal(t, domain.StatusIncomplete, r.Status, "case %d %s: missing %v", c.index, r.Key, want)
				assert.Contains(t, disclaimerCodes(got.Disclaimers), domain.DisclaimerIncompleteInputs, "case %d", c.index)
			default:
				assert.NotEqual(t, domain.StatusIncomplete, r.Status, "case %d %s: nothing is missing", c.index, r.Key)
			}
			assert.Equal(t, want, r.MissingInputs, "case %d %s %s: missing inputs", c.index, r.Key, c.period)
		}
	})
	t.Logf("%d incomplete regime estimates checked", incomplete)
	require.Positive(t, incomplete)
}

func disclaimerCodes(ds []domain.Disclaimer) []domain.DisclaimerCode {
	out := make([]domain.DisclaimerCode, 0, len(ds))
	for _, d := range ds {
		out = append(out, d.Code)
	}
	return out
}

func (c propCase) holdOf(code fiscal.RegimeCode) usecase.RegimeHold {
	for _, h := range c.holds {
		if h.Code == code {
			return h
		}
	}
	panic("no hold " + code)
}

func (c propCase) inputValues(code fiscal.RegimeCode) []domain.TaxInputValue { return c.inputs[code] }

// expectedMissing is §2.8: 612 PF, 626 PM and 601 PM need a prior payment
// for every earlier month of the year the regime was active in; 601 PM
// needs the coeficiente unless it is in its first fiscal year (when it
// makes no provisional payment, so it needs nothing); 625 PF needs the
// prior-year income with the definitive option. Nothing else is required.
func expectedMissing(key domain.StrategyKey, h usecase.RegimeHold, n time.Month, values []domain.TaxInputValue) []domain.MissingInput {
	has := func(key domain.InputKey, m time.Month) bool {
		return slices.ContainsFunc(values, func(v domain.TaxInputValue) bool { return v.Key == key && v.Month == m })
	}
	var missing []domain.MissingInput
	if h.Code == "601" {
		if slices.ContainsFunc(values, func(v domain.TaxInputValue) bool { return v.Key == domain.InputFirstFiscalYear && v.Bool }) {
			return nil
		}
		if !has(domain.InputCoeficienteUtilidad, 0) {
			missing = append(missing, domain.MissingInput{Key: domain.InputCoeficienteUtilidad})
		}
	}
	if needsPriorPayments(key) {
		var months []time.Month
		for m := time.January; m < n; m++ {
			if holdCovers(h, m) && !has(domain.InputPriorProvisionalPayments, m) {
				months = append(months, m)
			}
		}
		if len(months) > 0 {
			missing = append(missing, domain.MissingInput{Key: domain.InputPriorProvisionalPayments, Months: months})
		}
	}
	if h.Code == fiscal.RegimePlatforms {
		if definitive, _ := h.Options.Bool("definitive_option"); definitive && !has(domain.InputPriorYearIncome, 0) {
			missing = append(missing, domain.MissingInput{Key: domain.InputPriorYearIncome})
		}
	}
	return missing
}

// needsPriorPayments: the regimes whose provisional payments accumulate
// (612 PF, 601 PM, 626 PM); RESICO PF pays each month on its own.
func needsPriorPayments(key domain.StrategyKey) bool {
	switch key.Regime {
	case fiscal.RegimeBusiness, "601":
		return true
	case fiscal.RegimeRESICO:
		return key.PersonType == fiscal.PersonMoral
	}
	return false
}
