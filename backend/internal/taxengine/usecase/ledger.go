package usecase

import (
	"fmt"
	"time"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain"
)

// regimeMonth keys one reader row.
type regimeMonth struct {
	regime fiscal.RegimeCode
	month  time.Month
}

// YearSumsByRegime folds the readers' rows for q into each regime's
// YearSums for January..q.Through: a strategy that accumulates from January
// (612, 626 PM, 601) takes YearSums.Sum(time.January, q.Through), a monthly
// one (606, 626 PF, 625, and IVA always) YearSums.Month(q.Through). The
// income and the expense rows of a (regime, month) meet in one MonthSums; a
// regime without rows is absent (the zero YearSums), a month without rows
// is zero.
//
// It fails with domain.ErrInvalidMonth for a Through outside 1..12, a row
// outside January..Through, or a (regime, month) repeated within a reader's
// rows (the readers GROUP BY regime, month, so either is a reader bug).
func YearSumsByRegime(q LedgerQuery, incomes []IncomeMonthSums, expenses []ExpenseMonthSums) (map[fiscal.RegimeCode]domain.YearSums, error) {
	if _, err := domain.NewYearSums(q.Through); err != nil {
		return nil, err
	}
	months := map[regimeMonth]domain.MonthSums{}
	seen := map[regimeMonth]bool{}
	for _, r := range incomes {
		key := regimeMonth{r.Regime, r.Month}
		if err := claim(seen, key, q.Through, "income"); err != nil {
			return nil, err
		}
		sums := months[key]
		sums.Income = r.Sums
		months[key] = sums
	}
	clear(seen)
	for _, r := range expenses {
		key := regimeMonth{r.Regime, r.Month}
		if err := claim(seen, key, q.Through, "expense"); err != nil {
			return nil, err
		}
		sums := months[key]
		sums.Expenses = r.Sums
		months[key] = sums
	}
	return foldYears(q.Through, months)
}

// claim records key as read, refusing a month outside January..through or
// a key the same reader already returned.
func claim(seen map[regimeMonth]bool, key regimeMonth, through time.Month, reader string) error {
	if key.month < time.January || key.month > through || seen[key] {
		return fmt.Errorf("%s row %s/%d through %s: %w", reader, key.regime, key.month, through, domain.ErrInvalidMonth)
	}
	seen[key] = true
	return nil
}

// foldYears groups the months by regime into YearSums.
func foldYears(through time.Month, months map[regimeMonth]domain.MonthSums) (map[fiscal.RegimeCode]domain.YearSums, error) {
	entries := map[fiscal.RegimeCode][]domain.MonthEntry{}
	for key, sums := range months {
		entries[key.regime] = append(entries[key.regime], domain.MonthEntry{Month: key.month, Sums: sums})
	}
	years := make(map[fiscal.RegimeCode]domain.YearSums, len(entries))
	for regime, es := range entries {
		y, err := domain.NewYearSums(through, es...)
		if err != nil {
			return nil, fmt.Errorf("regime %s: %w", regime, err)
		}
		years[regime] = y
	}
	return years, nil
}
