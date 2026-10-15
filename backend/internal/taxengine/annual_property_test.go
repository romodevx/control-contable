package taxengine_test

import (
	"bytes"
	"cmp"
	"encoding/json"
	"errors"
	"io"
	"math/rand/v2"
	"slices"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain/annual"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain/annual/pm"
	"github.com/romodevx/control-contable/backend/internal/taxengine/infra/annuallimits"
	usecaseannual "github.com/romodevx/control-contable/backend/internal/taxengine/usecase/annual"
)

// The annual property suite (ANN-16, MODULE-annual §2.11, ADR-0002 §5):
// invariants of the merged EstimateAnnual use case over every PF and PM
// section strategy, on generated years, fixed seed. Generators:
// annual_property_gen_test.go.

// --- (1) lines + rounding adjustment = result, per section ---

// TestAnnualPropertyLinesReconcile: every computed section's shown lines
// since the last anchor, plus the rounding adjustment, add up to each
// result line; the section's figures are its result lines (ISR, balance,
// RESICO difference) and its credits are the sum of their parts.
func TestAnnualPropertyLinesReconcile(t *testing.T) {
	forEachAnnCase(t, func(t *testing.T, c annCase) {
		for _, s := range c.estimate(t).Sections {
			label := string(s.Kind)
			assertLinesReconcile(t, c.index, label, domainLines(s.Lines))
			if s.Tax != nil {
				assertTaxIsItsResultLines(t, c.index, s)
			}
			if s.RESICOCheck != nil {
				assertRESICOCheckIsItsResultLine(t, c.index, s)
			}
		}
	})
}

func domainLines(lines []usecaseannual.Line) []domain.Line {
	out := make([]domain.Line, len(lines))
	for i, l := range lines {
		out[i] = l.Line
	}
	return out
}

// annISRLines and annBalanceLines are the result lines that close a
// section's annual ISR and its balance (PM sections trace the PTU after
// the balance).
var (
	annISRLines     = []domain.LineCode{annual.LineArt152ISR, pm.LineAnnualTax}
	annBalanceLines = []domain.LineCode{annual.LineArt152Balance, pm.LineBalance}
)

// resultLine is the single result line of s whose code is in codes.
func resultLine(t *testing.T, index int, s usecaseannual.Section, codes []domain.LineCode) domain.Line {
	t.Helper()
	var found []domain.Line
	for _, l := range s.Lines {
		if l.Operation == domain.OpResult && slices.Contains(codes, l.Code) {
			found = append(found, l.Line)
		}
	}
	require.Len(t, found, 1, "case %d %s: one result line of %v", index, s.Kind, codes)
	return found[0]
}

func assertTaxIsItsResultLines(t *testing.T, index int, s usecaseannual.Section) {
	t.Helper()
	tax := s.Tax
	credits := tax.Payments + tax.Withholdings + tax.SalaryISR
	assert.Equal(t, credits, s.Credits, "case %d %s: credits = payments + withholdings + salary ISR", index, s.Kind)
	assert.Equal(t, tax.ISR, resultLine(t, index, s, annISRLines).Amount, "case %d %s: ISR = its result line", index, s.Kind)
	assert.Equal(t, signed(tax.Balance), resultLine(t, index, s, annBalanceLines).Amount,
		"case %d %s: the balance is its result line", index, s.Kind)
}

func assertRESICOCheckIsItsResultLine(t *testing.T, index int, s usecaseannual.Section) {
	t.Helper()
	check := s.RESICOCheck
	last, ok := lastResult(domainLines(s.Lines), domain.TaxISR)
	require.True(t, ok, "case %d %s", index, s.Kind)
	assert.Equal(t, check.Difference, last.Amount, "case %d %s: the difference is the last result line", index, s.Kind)
	assert.Equal(t, check.AnnualISR-check.Payments-check.Withheld, check.Difference,
		"case %d %s: annual ISR − payments − withheld = difference", index, s.Kind)
}

// signed is a balance as a signed amount: a cargo positive, a favor
// negative.
func signed(b annual.Balance) money.Cents {
	if b.Kind == annual.BalanceAFavor {
		return -b.Amount
	}
	return b.Amount
}

// --- (2) byte-identical JSON for identical inputs ---

// annPropJSON is the canonical encoding of an estimate: encoding/json
// over the use case's result (fields in declaration order, no map), plus
// the values whose fields are unexported (periods, dates, line sources).
type annPropJSON struct {
	DataThrough      string
	Deadline         string
	SectionDeadlines []string
	LineSources      []string
	Estimate         usecaseannual.Estimate
}

func annCanonicalJSON(t *testing.T, e usecaseannual.Estimate) []byte {
	t.Helper()
	view := annPropJSON{DataThrough: e.DataThrough.String(), Deadline: e.Deadline.String(), Estimate: e}
	for _, s := range e.Sections {
		view.SectionDeadlines = append(view.SectionDeadlines, s.Deadline.String())
		for _, l := range s.Lines {
			view.LineSources = append(view.LineSources, l.Source.String())
		}
	}
	out, err := json.Marshal(view)
	require.NoError(t, err)
	return out
}

// TestAnnualPropertyIdenticalInputsGiveByteIdenticalJSON: the same
// records, inputs and parameters give the same bytes and fingerprint, from
// a fresh estimator, and whatever order the ports return their rows,
// regimes, input values and constancias in (no map- or row-order
// dependence, §2.11). Deductions keep their order: the reader sorts them
// (YearInputsOf), and the calculator's notes follow it.
func TestAnnualPropertyIdenticalInputsGiveByteIdenticalJSON(t *testing.T) {
	forEachAnnCase(t, func(t *testing.T, c annCase) {
		first := c.estimate(t)
		again := c.clone().estimate(t)
		shuffled := c.shuffled(annRand(c.index, 1)).estimate(t)

		firstJSON := annCanonicalJSON(t, first)
		require.Contains(t, string(firstJSON), `"Sections"`, "case %d: the encoding carries the sections", c.index)
		assert.True(t, bytes.Equal(firstJSON, annCanonicalJSON(t, again)), "case %d: same input, different JSON", c.index)
		assert.True(t, bytes.Equal(firstJSON, annCanonicalJSON(t, shuffled)), "case %d: reordered rows, different JSON", c.index)
		assert.Equal(t, first.Fingerprint, shuffled.Fingerprint, "case %d: reordered rows, different fingerprint", c.index)
	})
}

// shuffled returns the same case with every list in another order.
func (c annCase) shuffled(rng *rand.Rand) annCase {
	out := c.clone()
	shuffle(rng, out.holds)
	shuffle(rng, out.incomes)
	shuffle(rng, out.expenses)
	for _, values := range out.inputs {
		shuffle(rng, values)
	}
	shuffle(rng, out.year.SalaryStatements)
	return out
}

// --- (3) the balance's sign follows the credits ---

// TestAnnualPropertyBalanceSignFollowsTheCredits: a balance is a cargo
// when the ISR exceeds the credits, a favor when the credits exceed it,
// cero when they meet, its amount never negative; a favor needs credits.
// Metamorphic: confirming Δ more in a filed month moves the section's
// signed balance down by exactly Δ and leaves its ISR alone.
func TestAnnualPropertyBalanceSignFollowsTheCredits(t *testing.T) {
	var compared int
	kinds := map[annual.BalanceKind]int{}
	forEachAnnCase(t, func(t *testing.T, c annCase) {
		got := c.estimate(t)
		for _, s := range got.Sections {
			if s.Tax == nil {
				continue
			}
			kinds[s.Tax.Balance.Kind]++
			assertBalanceSign(t, c.index, s)
		}
		rng := annRand(c.index, 2)
		regime, month, ok := c.filedMonthToRaise(got, rng)
		if !ok {
			return
		}
		delta := money.Cents(1 + rng.Int64N(int64(10_000*propPeso)))
		raised := c.clone()
		raised.filed[regime][month] += delta
		after := raised.estimate(t)
		for i, s := range got.Sections {
			if s.Tax == nil || !slices.Contains(s.Regimes, regime) {
				continue
			}
			compared++
			a := after.Sections[i]
			require.Equal(t, s.Kind, a.Kind, "case %d", c.index)
			assert.Equal(t, s.Tax.ISR, a.Tax.ISR, "case %d %s: a filed payment moved the ISR", c.index, s.Kind)
			assert.Equal(t, s.Tax.Payments+delta, a.Tax.Payments, "case %d %s: payments credit the filed amount", c.index, s.Kind)
			assert.Equal(t, signed(s.Tax.Balance)-delta, signed(a.Tax.Balance),
				"case %d %s: %s more filed in %s %s moved the balance %v -> %v", c.index, s.Kind, delta, regime, month, s.Tax.Balance, a.Tax.Balance)
		}
	})
	t.Logf("balances by kind %v; %d sections re-estimated with a raised filed payment", kinds, compared)
	require.Positive(t, compared)
	for _, k := range []annual.BalanceKind{annual.BalanceACargo, annual.BalanceAFavor, annual.BalanceCero} {
		require.Positive(t, kinds[k], "no %s balance generated", k)
	}
}

func assertBalanceSign(t *testing.T, index int, s usecaseannual.Section) {
	t.Helper()
	b := s.Tax.Balance
	require.True(t, b.Kind.Valid(), "case %d %s: balance kind %q", index, s.Kind, b.Kind)
	assert.GreaterOrEqual(t, b.Amount, money.Cents(0), "case %d %s: a negative balance amount", index, s.Kind)
	assert.Equal(t, b.Kind == annual.BalanceCero, b.Amount == 0, "case %d %s: cero iff zero (%v)", index, s.Kind, b)
	assert.GreaterOrEqual(t, s.Tax.ISR, money.Cents(0), "case %d %s: a negative ISR", index, s.Kind)
	assert.Equal(t, s.Tax.ISR-s.Credits, signed(b), "case %d %s: ISR %s − credits %s ≠ balance %v", index, s.Kind, s.Tax.ISR, s.Credits, b)
	if b.Kind == annual.BalanceAFavor {
		assert.Positive(t, s.Credits, "case %d %s: a favor without credits", index, s.Kind)
	}
}

// filedMonthToRaise picks a confirmed month, counted by the estimate
// (through data_through), of a regime a computed section credits.
func (c annCase) filedMonthToRaise(e usecaseannual.Estimate, rng *rand.Rand) (fiscal.RegimeCode, time.Month, bool) {
	if e.DataThrough.IsZero() {
		return "", 0, false
	}
	type candidate struct {
		regime fiscal.RegimeCode
		month  time.Month
	}
	var pool []candidate
	for _, s := range e.Sections {
		if s.Tax == nil {
			continue
		}
		for _, code := range s.Regimes {
			if !c.paysProvisionalISR(code) {
				continue // 605: its ISR is the employer's (salary_isr)
			}
			for m := range c.filed[code] {
				if m <= e.DataThrough.Month() && c.noDefaultAfter(code, m, e.DataThrough.Month()) {
					pool = append(pool, candidate{code, m})
				}
			}
		}
	}
	if len(pool) == 0 {
		return "", 0, false
	}
	// The pool came from a map: sort it so the pick is reproducible.
	slices.SortFunc(pool, func(a, b candidate) int {
		return cmp.Or(cmp.Compare(a.regime, b.regime), cmp.Compare(a.month, b.month))
	})
	pick := pool[rng.IntN(len(pool))]
	return pick.regime, pick.month, true
}

// noDefaultAfter reports that no held month of code after m (through the
// data) takes the app's estimate: a cumulative regime's defaults are
// chained (ResolveRegimePayments), so a later default would absorb a
// raised filed month by design.
func (c annCase) noDefaultAfter(code fiscal.RegimeCode, m, through time.Month) bool {
	hold := c.holdOf(code)
	for later := m + 1; later <= through; later++ {
		if !holdCovers(hold, later) {
			continue
		}
		if _, filed := c.filed[code][later]; filed {
			continue
		}
		typed := slices.ContainsFunc(c.inputs[code], func(v domain.TaxInputValue) bool {
			return v.Key == domain.InputPriorProvisionalPayments && v.Month == later
		})
		if !typed {
			return false
		}
	}
	return true
}

// paysProvisionalISR reports a held regime whose provisional payments the
// annual credits.
func (c annCase) paysProvisionalISR(code fiscal.RegimeCode) bool {
	for _, h := range c.holds {
		if h.Code == code {
			definitive, _ := h.Options.Bool(annual.OptionPlatformDefinitive)
			return annual.PaysProvisionalISR(code, definitive)
		}
	}
	return false
}

// --- (4) personal deductions never exceed their caps ---

// TestAnnualPropertyPersonalDeductionsNeverExceedTheirCaps: a persona
// física's applied deductions are never negative, never above what was
// registered, and their capped part never above the global cap of 5 UMA
// anuales (LISR 151 last ¶; the 15 % of income is the domain suite's,
// ANN-07); the base is the acumulables less them, floored at zero; and
// registering deductions never raises the ISR. A persona moral has none.
func TestAnnualPropertyPersonalDeductionsNeverExceedTheirCaps(t *testing.T) {
	_, cat := propEngine()
	limits, err := annuallimits.New(cat).DeductionLimits(propYear)
	require.NoError(t, err)
	globalCap := limits.AnnualUMA * money.Cents(limits.GlobalCapUMAs)
	var applied int
	forEachAnnCase(t, func(t *testing.T, c annCase) {
		got := c.estimate(t)
		var claimed, uncapped money.Cents
		for _, d := range c.year.Deductions {
			claimed += d.Amount
			if !d.Category.Capped() {
				uncapped += d.Amount
			}
		}
		for _, s := range got.Sections {
			if s.Tax == nil {
				continue
			}
			if s.Kind != annual.SectionPFArt152 {
				if c.person == fiscal.PersonMoral {
					assert.Nil(t, s.Tax.PersonalDeductions, "case %d %s: a persona moral has no personal deductions", c.index, s.Kind)
				}
				continue
			}
			require.NotNil(t, s.Tax.PersonalDeductions, "case %d", c.index)
			pd := *s.Tax.PersonalDeductions
			if pd > 0 {
				applied++
			}
			assert.GreaterOrEqual(t, pd, money.Cents(0), "case %d", c.index)
			assert.LessOrEqual(t, pd, claimed, "case %d: applied above registered", c.index)
			assert.LessOrEqual(t, pd-uncapped, globalCap, "case %d: capped deductions above 5 UMA", c.index)
			assert.Equal(t, max(0, s.Tax.AccumulableIncome-pd), s.Tax.Base, "case %d: base = acumulables − deductions, floored", c.index)
		}
		assertDeductionsNeverRaiseTheISR(t, c, got)
	})
	t.Logf("%d sections applied personal deductions", applied)
	require.Positive(t, applied)
}

func assertDeductionsNeverRaiseTheISR(t *testing.T, c annCase, with usecaseannual.Estimate) {
	t.Helper()
	if len(c.year.Deductions) == 0 {
		return
	}
	none := c.clone()
	none.year.Deductions = nil
	without := none.estimate(t)
	for i, s := range with.Sections {
		if s.Kind != annual.SectionPFArt152 {
			continue
		}
		assert.LessOrEqual(t, s.Tax.ISR, without.Sections[i].Tax.ISR, "case %d: deductions raised the ISR", c.index)
	}
}

// --- (5) no floats in the estimate ---

// TestAnnualPropertyEstimateCarriesNoFloats: every number in the
// estimate's canonical JSON is an integer (centavos, rates in millionths,
// counts): nothing reached the output through a float. The source-level
// guard is TestNoFloatingPointInMoneyCode.
func TestAnnualPropertyEstimateCarriesNoFloats(t *testing.T) {
	var numbers int
	forEachAnnCase(t, func(t *testing.T, c annCase) {
		dec := json.NewDecoder(bytes.NewReader(annCanonicalJSON(t, c.estimate(t))))
		dec.UseNumber()
		for {
			tok, err := dec.Token()
			if errors.Is(err, io.EOF) {
				break
			}
			require.NoError(t, err)
			n, ok := tok.(json.Number)
			if !ok {
				continue
			}
			numbers++
			_, err = strconv.ParseInt(n.String(), 10, 64)
			assert.NoError(t, err, "case %d: %s is not an integer", c.index, n)
		}
	})
	require.Positive(t, numbers)
}

// --- (6) every PF and PM strategy is exercised ---

// annExpectedKinds are the sections the suite must reach: every computed
// strategy (PF and PM) and the explanation-only regimes.
var annExpectedKinds = []annual.SectionKind{
	annual.SectionPFArt152, annual.SectionPFRESICOCheck, annual.SectionPM601, annual.SectionPMRESICO, annual.SectionExplanation,
}

// TestAnnualPropertyGeneratorReachesEveryStrategy keeps the properties
// honest: the generated years produce every section kind, each art. 152
// chapter, both year statuses and every section status, so a property
// cannot pass by never meeting a strategy.
func TestAnnualPropertyGeneratorReachesEveryStrategy(t *testing.T) {
	kinds := map[annual.SectionKind]int{}
	statuses := map[annual.Status]int{}
	years := map[usecaseannual.YearStatus]int{}
	chapters := map[fiscal.RegimeCode]int{}
	categories := map[annual.Category]int{}
	beneficiaries := map[annual.Beneficiary]int{}
	forEachAnnCase(t, func(t *testing.T, c annCase) {
		for _, d := range c.year.Deductions {
			categories[d.Category]++
			beneficiaries[d.Beneficiary]++
		}
		got := c.estimate(t)
		years[got.YearStatus]++
		for _, s := range got.Sections {
			kinds[s.Kind]++
			statuses[s.Status]++
			if s.Kind == annual.SectionPFArt152 {
				for _, r := range s.Regimes {
					chapters[r]++
				}
			}
		}
	})
	t.Logf("sections %v; statuses %v; years %v; art. 152 chapters %v; deductions %v by %v",
		kinds, statuses, years, chapters, categories, beneficiaries)
	for _, k := range annExpectedKinds {
		assert.Positive(t, kinds[k], "no %s section generated", k)
	}
	for _, s := range []annual.Status{annual.StatusComplete, annual.StatusIncomplete, annual.StatusInformational} {
		assert.Positive(t, statuses[s], "no %s section generated", s)
	}
	assert.Positive(t, years[usecaseannual.YearClosed])
	assert.Positive(t, years[usecaseannual.YearInProgress])
	for _, code := range []fiscal.RegimeCode{annual.RegimeSalary, fiscal.RegimeRent, fiscal.RegimeBusiness, fiscal.RegimePlatforms} {
		assert.Positive(t, chapters[code], "no art. 152 section with chapter %s", code)
	}
	for _, category := range annual.Categories() {
		assert.Positive(t, categories[category], "no %s deduction generated", category)
	}
	for _, b := range annBeneficiaries {
		assert.Positive(t, beneficiaries[b], "no deduction for a %s generated", b)
	}
}

// annBeneficiaries are the spec's beneficiaries (annual.Beneficiary.Valid).
var annBeneficiaries = []annual.Beneficiary{
	annual.BeneficiarySelf, annual.BeneficiarySpouse, annual.BeneficiaryAscendant, annual.BeneficiaryDescendant,
}
