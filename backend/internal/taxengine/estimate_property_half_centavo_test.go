package taxengine_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain"
	"github.com/romodevx/control-contable/backend/internal/taxengine/usecase"
)

// The monthly twin of the annual half-centavo fix (ANN-16, PR #148): a
// balance is the tax as SHOWN less the credits as shown. A tax figure whose
// exact value ends in half a centavo (100.005, shown 100.01) must not round
// a later balance past it: 100.01 shown less 100.01 withheld is 0.00, never
// −0.01, and a RESICO excess is the credits less the shown ISR.

// halfCentavoCase is one monthly ISR strategy: the line its credits are
// subtracted from, and a single-regime taxpayer with cash income and a
// credit of the given amount.
type halfCentavoCase struct {
	key    domain.StrategyKey
	figure domain.LineCode // the ISR figure the credits are subtracted from
	floors bool            // the month floors at zero (RESICO PF, 625 definitive)
	// noCredit: the regime subtracts no credit (RIF: the payable is the
	// reduced figure itself), so any sub-centavo figure will do.
	noCredit bool
	start    money.Cents // the first income tried
	build    func(income, credit money.Cents) propCase
}

// halfCentavoMonth is January: no prior provisional payment is required,
// so the estimate is complete (621, bimonthly, uses February).
const halfCentavoMonth = time.January

// halfCentavoScan bounds the search for an income whose figure ends in
// half a centavo: every rate the cases use reaches one within 40 centavos.
const halfCentavoScan = 1_000

var halfCentavoCases = []halfCentavoCase{
	{
		key: domain.StrategyKey{Regime: fiscal.RegimeBusiness, PersonType: fiscal.PersonFisica}, figure: "isr.caused",
		start: 100_000 * propPeso,
		build: func(income, credit money.Cents) propCase {
			return singleRegimeCase(fiscal.PersonFisica, fiscal.RegimeBusiness, halfCentavoMonth, nil,
				domain.IncomeSums{Cash: income, Accrual: income, ISRWithheld: domain.Withheld{ByMoral: credit}}, nil)
		},
	},
	{
		key: domain.StrategyKey{Regime: fiscal.RegimeRent, PersonType: fiscal.PersonFisica}, figure: "isr.caused",
		start: 100_000 * propPeso,
		build: func(income, credit money.Cents) propCase {
			return singleRegimeCase(fiscal.PersonFisica, fiscal.RegimeRent, halfCentavoMonth,
				map[domain.OptionKey]domain.OptionValue{domain.OptionRentDeductionMode: domain.TextOption("actual")},
				domain.IncomeSums{Cash: income, Accrual: income, ISRWithheld: domain.Withheld{ByMoral: credit}}, nil)
		},
	},
	{
		key: domain.StrategyKey{Regime: fiscal.RegimeRESICO, PersonType: fiscal.PersonFisica}, figure: domain.LineRESICOPFRate,
		floors: true, start: 10_000 * propPeso,
		build: func(income, credit money.Cents) propCase {
			return singleRegimeCase(fiscal.PersonFisica, fiscal.RegimeRESICO, halfCentavoMonth, nil,
				domain.IncomeSums{Cash: income, Accrual: income, ISRWithheld: domain.Withheld{ByMoral: credit}}, nil)
		},
	},
	{
		key: domain.StrategyKey{Regime: fiscal.RegimePlatforms, PersonType: fiscal.PersonFisica}, figure: "isr.platform_rate",
		floors: true, start: 10_000 * propPeso,
		build: func(income, credit money.Cents) propCase {
			return singleRegimeCase(fiscal.PersonFisica, fiscal.RegimePlatforms, halfCentavoMonth,
				map[domain.OptionKey]domain.OptionValue{
					"activities":            domain.ListOption([]string{"goods_services"}),
					"definitive_option":     domain.BoolOption(true),
					"rfc_given_to_platform": domain.BoolOption(true),
				},
				domain.IncomeSums{Cash: income, Accrual: income, ISRWithheld: domain.Withheld{ByPlatform: credit}}, nil)
		},
	},
	{
		key: domain.StrategyKey{Regime: "621", PersonType: fiscal.PersonFisica}, figure: "isr.rif_reduction",
		noCredit: true, start: 50_000 * propPeso,
		build: func(income, _ money.Cents) propCase {
			return singleRegimeCase(fiscal.PersonFisica, "621", time.February,
				map[domain.OptionKey]domain.OptionValue{domain.OptionRIFYear: domain.TextOption("10")},
				domain.IncomeSums{Cash: income, Accrual: income}, nil)
		},
	},
	{
		key: domain.StrategyKey{Regime: "601", PersonType: fiscal.PersonMoral}, figure: "isr.rate",
		start: 200_000 * propPeso,
		build: func(income, credit money.Cents) propCase {
			return singleRegimeCase(fiscal.PersonMoral, "601", halfCentavoMonth, nil,
				domain.IncomeSums{Cash: income, Accrual: income}, []domain.TaxInputValue{
					{Key: domain.InputCoeficienteUtilidad, Kind: domain.InputRate, Rate: 50 * propPercent},
					{Key: domain.InputInterestWithholding, Kind: domain.InputMoney, Money: credit},
				})
		},
	},
	{
		key: domain.StrategyKey{Regime: fiscal.RegimeRESICO, PersonType: fiscal.PersonMoral}, figure: "isr.rate",
		start: 100_000 * propPeso,
		build: func(income, credit money.Cents) propCase {
			return singleRegimeCase(fiscal.PersonMoral, fiscal.RegimeRESICO, halfCentavoMonth, nil,
				domain.IncomeSums{Cash: income, Accrual: income},
				[]domain.TaxInputValue{{Key: domain.InputInterestWithholding, Kind: domain.InputMoney, Money: credit}})
		},
	},
}

// singleRegimeCase is a taxpayer holding one regime since the year before,
// with income in month m only.
func singleRegimeCase(person fiscal.PersonType, code fiscal.RegimeCode, m time.Month,
	options map[domain.OptionKey]domain.OptionValue, income domain.IncomeSums, inputs []domain.TaxInputValue,
) propCase {
	return propCase{
		person: person, period: periodString(m),
		holds: []usecase.RegimeHold{{
			Code: code, Treatment: domain.TreatmentFull, Options: domain.NewOptions(options),
			Start: mustPeriod(propYear-1, time.July),
		}},
		incomes: []usecase.IncomeMonthSums{{Regime: code, Month: m, Sums: income}},
		inputs:  map[fiscal.RegimeCode][]domain.TaxInputValue{code: inputs},
	}
}

// TestEstimateMonthlyBalanceIsShownISRLessShownCredits: for every monthly
// ISR strategy, on a figure ending in half a centavo, the payable is the
// shown figure less the shown credits (floored at zero where the regime
// floors, the excess line then being the credits less the shown figure),
// with no rounding adjustment in between: for a credit equal to the shown
// figure (balance 0.00, not −0.01) and one a hundred pesos above it.
func TestEstimateMonthlyBalanceIsShownISRLessShownCredits(t *testing.T) {
	covered := map[domain.StrategyKey]bool{}
	for _, hc := range halfCentavoCases {
		t.Run(hc.key.String(), func(t *testing.T) {
			income, figure := hc.halfCentavoFigure(t)
			for _, credit := range []money.Cents{figure.Amount, figure.Amount + 100*propPeso} {
				lines := hc.isrLines(t, income, credit)
				payable := lines[len(lines)-1]
				label := fmt.Sprintf("income %s, figure exact %d shown %s, credit %s", income, figure.Exact, figure.Amount, credit)
				assertBalanceIsShownFigureLessCredits(t, label, hc.figure, lines)

				want := figure.Amount - credit
				if hc.noCredit {
					want = figure.Amount
				}
				if hc.floors {
					want = max(want, 0)
				}
				assert.Equal(t, want, payable.Amount, "%s: payable = shown ISR − shown credits", label)
			}
			covered[hc.key] = true
		})
	}
	for _, s := range isrStrategies() {
		assert.True(t, covered[s], "no half-centavo case for %s", s)
	}
}

// isrStrategies are the monthly strategies that estimate an ISR figure:
// 622 and 603 have none (partial PM, IVA only), and IVA is whole centavos.
func isrStrategies() []domain.StrategyKey {
	var keys []domain.StrategyKey
	for _, r := range propRegimes {
		if r.treatment == domain.TreatmentInformational || r.code == "622" || r.code == "603" {
			continue
		}
		keys = append(keys, domain.StrategyKey{Regime: r.code, PersonType: r.person})
	}
	return keys
}

// halfCentavoFigure finds the first income from hc.start whose figure line
// ends in exactly half a centavo (any fraction of one, for noCredit), and
// returns it with that line.
func (hc halfCentavoCase) halfCentavoFigure(t *testing.T) (money.Cents, domain.Line) {
	t.Helper()
	const half = money.ExactPerCent / 2
	for income := hc.start; income < hc.start+halfCentavoScan; income++ {
		for _, l := range hc.isrLines(t, income, 0) {
			rest := l.Exact % money.ExactPerCent
			if l.Code == hc.figure && (rest == half || hc.noCredit && rest != 0) {
				return income, l
			}
		}
	}
	require.FailNow(t, "no half-centavo figure", "%s: none from %s in %d centavos", hc.key, hc.start, halfCentavoScan)
	return 0, domain.Line{}
}

// isrLines are the regime's ISR lines, its payable last.
func (hc halfCentavoCase) isrLines(t *testing.T, income, credit money.Cents) []domain.Line {
	t.Helper()
	c := hc.build(income, credit)
	r, ok := regimeOf(c.estimate(t), hc.key)
	require.True(t, ok, "%s: no regime result", hc.key)
	require.NotNil(t, r.ISR, "%s: no ISR figure", hc.key)
	var lines []domain.Line
	for _, l := range r.Lines {
		if l.Tax == domain.TaxISR {
			lines = append(lines, l)
		}
	}
	require.NotEmpty(t, lines)
	require.Equal(t, domain.OpResult, lines[len(lines)-1].Operation, "%s: the ISR ends in its payable", hc.key)
	require.Equal(t, r.ISR.Amount, lines[len(lines)-1].Amount, "%s: ISR = its payable line", hc.key)
	return lines
}

// assertBalanceIsShownFigureLessCredits replays, in centavos as shown, the
// lines from the figure to the payable (its last line): the payable is the
// shown figure plus the shown adds less the shown subtractions, and no
// rounding adjustment stands between them.
func assertBalanceIsShownFigureLessCredits(t *testing.T, label string, figure domain.LineCode, lines []domain.Line) {
	t.Helper()
	start := -1
	for i, l := range lines {
		if l.Code == figure {
			start = i
		}
	}
	require.GreaterOrEqual(t, start, 0, "%s: no %s line", label, figure)
	shown := lines[start].Amount
	payable := lines[len(lines)-1]
	for _, l := range lines[start+1 : len(lines)-1] {
		switch l.Operation {
		case domain.OpAdd:
			shown += l.Amount
		case domain.OpSubtract:
			shown -= l.Amount
		case domain.OpInfo:
		default:
			t.Errorf("%s: line %d %s (%s) between the figure and the payable", label, l.Seq, l.Code, l.Operation)
		}
	}
	assert.Equal(t, shown, payable.Amount, "%s: payable = the shown figure ± the shown lines after it", label)
}

// TestEstimatePropertyBalancesNeedNoRoundingAdjustment: over the generated
// suite, no figure's balance (each regime's last ISR, IVA and withholding
// result, and the consolidated IVA) is preceded by a rounding adjustment.
// Only the figure a balance starts from may be sub-centavo; what follows
// it is whole centavos, so an adjustment there means the balance was
// rounded from the exact figure instead of the shown one. The IVA, whole
// centavos throughout, is guarded here for every strategy (622 and 603
// included).
func TestEstimatePropertyBalancesNeedNoRoundingAdjustment(t *testing.T) {
	forEachCase(t, func(t *testing.T, c propCase) {
		got := c.estimate(t)
		for _, r := range got.Regimes {
			for _, tax := range []domain.TaxKind{domain.TaxISR, domain.TaxIVA, domain.TaxWithholding} {
				assertNoAdjustmentBeforeLastResult(t, fmt.Sprintf("case %d %s %s", c.index, r.Key, tax), r.Lines, tax)
			}
		}
		assertNoAdjustmentBeforeLastResult(t, fmt.Sprintf("case %d consolidated IVA", c.index), got.IVA.Lines, domain.TaxIVA)
	})
}

// assertNoAdjustmentBeforeLastResult checks tax's last result line, the
// balance, is not preceded by a rounding adjustment.
func assertNoAdjustmentBeforeLastResult(t *testing.T, label string, lines []domain.Line, tax domain.TaxKind) {
	t.Helper()
	for i := len(lines) - 1; i > 0; i-- {
		if lines[i].Tax == tax && lines[i].Operation == domain.OpResult {
			assert.NotEqual(t, domain.OpRoundingAdjustment, lines[i-1].Operation,
				"%s: line %d %s, the balance, needs a rounding adjustment", label, lines[i].Seq, lines[i].Code)
			return
		}
	}
}
