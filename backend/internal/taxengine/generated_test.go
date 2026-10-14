package taxengine_test

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain"
	"github.com/romodevx/control-contable/backend/internal/taxengine/golden"
)

const fiscalYear = 2026

// edgeStrategy generates strategy-level cases at every row edge of the
// tariff a regime applies to an income with nothing else in it: the base
// is the income, so the ISR is the reference tariff's tax on it.
type edgeStrategy struct {
	key    domain.StrategyKey
	table  string
	months []int // the monthly tables used; {0} for one plain table
	// options of the regime that make the base the plain income.
	options map[domain.OptionKey]domain.OptionValue
	// priorPayments: the regime requires prior provisional payments (given
	// as 0.00, so the case is complete).
	priorPayments bool
}

var edgeStrategies = []edgeStrategy{
	{ // 612 PF: YTD income (no deductions) on the accumulated table of month n.
		key: domain.StrategyKey{Regime: fiscal.RegimeBusiness, PersonType: fiscal.PersonFisica}, table: "isr.art106.accumulated",
		months: []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}, priorPayments: true,
	},
	{ // 606 PF: the month's rent with actual (zero) deductions on the art. 116 table.
		key: domain.StrategyKey{Regime: fiscal.RegimeRent, PersonType: fiscal.PersonFisica}, table: "isr.art116.monthly",
		months:  []int{0},
		options: map[domain.OptionKey]domain.OptionValue{"deduction_mode": domain.TextOption("actual")},
	},
	{ // 626 PF: the month's collected income on the flat RESICO table.
		key: domain.StrategyKey{Regime: fiscal.RegimeRESICO, PersonType: fiscal.PersonFisica}, table: "isr.resico_pf.monthly",
		months: []int{0},
	},
}

// generatedCases are the strategy-level boundary cases. They are pending
// while their strategy is not registered and run as soon as it is.
func generatedCases(t *testing.T, rp golden.RawParams) []golden.Case {
	t.Helper()
	var cases []golden.Case
	for _, es := range edgeStrategies {
		for _, month := range es.months {
			raw := findTable(t, rp, es.table, month)
			points, err := golden.TariffEdges(raw)
			require.NoError(t, err)
			for _, pt := range points {
				want, inTable := refTax(raw, pt.Amount)
				if !inTable {
					continue // above the table: outside the regime (e.g. the RESICO cap)
				}
				cases = append(cases, edgeCase(t, es, month, pt, want))
			}
		}
	}
	return cases
}

func edgeCase(t *testing.T, es edgeStrategy, tableMonth int, pt golden.TariffPoint, isr money.Cents) golden.Case {
	t.Helper()
	month := time.Month(max(tableMonth, int(time.January)))
	period, err := domain.NewPeriod(fiscalYear, month)
	require.NoError(t, err)
	sums, err := domain.NewYearSums(month, domain.MonthEntry{
		Month: month, Sums: domain.MonthSums{Income: domain.IncomeSums{Cash: pt.Amount}},
	})
	require.NoError(t, err)
	var inputs []domain.TaxInputValue
	active := make([]time.Month, 0, int(month))
	for m := time.January; m <= month; m++ {
		active = append(active, m)
		if es.priorPayments && m < month {
			inputs = append(inputs, domain.TaxInputValue{
				Key: domain.InputPriorProvisionalPayments, Kind: domain.InputMoneyMonthly, Month: m,
			})
		}
	}
	taxInputs, err := domain.NewTaxInputs(inputs...)
	require.NoError(t, err)
	label := strings.ReplaceAll(pt.Label, " ", "-")
	return golden.Case{
		ID:        fmt.Sprintf("GEN-%s-%s-%s", strings.ReplaceAll(es.key.String(), "/", "-"), period, label),
		Source:    fmt.Sprintf("generated:%s/%d:%s", es.table, tableMonth, pt.Label),
		Generated: true,
		Input: domain.Input{
			Key: es.key, Period: period, Options: domain.NewOptions(es.options), Sums: sums, Inputs: taxInputs,
		},
		ActiveMonths: active,
		Want: golden.Expected{
			Status: domain.StatusComplete,
			Taxes:  []domain.TaxKind{domain.TaxISR},
			ISR:    &isr,
		},
	}
}
