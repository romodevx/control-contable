// Package usecase holds the tax engine's application layer. This file
// declares the read ports the estimate use case (M5-22) consumes, on the
// consumer side (ADR-0001): small interfaces the incomes (M3-07), expenses
// (M4-08) and tax-inputs (M5-21) adapters implement. Every call runs inside
// the request's RLS transaction, so a port only ever sees the context user's
// rows (MODULE-tax-engine S-2).
package usecase

import (
	"context"
	"time"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain"
)

// LedgerQuery selects one taxpayer's records from January of FirstYear
// through the period's month of FiscalYear. A zero FirstYear is
// FiscalYear: one fiscal year, January through Through.
type LedgerQuery struct {
	TaxpayerID string
	// FirstYear opens the range on its January (EstimateRange: a window
	// spanning two fiscal years is still one statement per table).
	FirstYear  int
	FiscalYear int
	Through    time.Month
}

// IncomeMonthSums is one (regime, year, month) row of income aggregates.
type IncomeMonthSums struct {
	Regime fiscal.RegimeCode
	Year   int
	Month  time.Month
	Sums   domain.IncomeSums
}

// ExpenseMonthSums is one (regime, year, month) row of expense aggregates.
type ExpenseMonthSums struct {
	Regime fiscal.RegimeCode
	Year   int
	Month  time.Month
	Sums   domain.ExpenseSums
}

// IncomeLedgerReader reads income aggregates: one indexed GROUP BY month,
// regime over non-deleted incomes (M3-07). It reads even when
// module_incomes is OFF (MODULE-incomes OQ-8).
type IncomeLedgerReader interface {
	// IncomeSums returns one row per (regime, year, month) with records,
	// ordered by regime, year, month; months without records are absent.
	IncomeSums(ctx context.Context, q LedgerQuery) ([]IncomeMonthSums, error)
}

// ExpenseLedgerReader reads expense aggregates by paid_on (M4-08).
type ExpenseLedgerReader interface {
	// ExpenseSums returns one row per (regime, year, month) with records,
	// ordered by regime, year, month; months without records are absent.
	ExpenseSums(ctx context.Context, q LedgerQuery) ([]ExpenseMonthSums, error)
}

// LedgerReader is both readers: the composition root wires the incomes and
// the expenses adapters into one value.
type LedgerReader interface {
	IncomeLedgerReader
	ExpenseLedgerReader
}

// TaxInputReader reads the taxpayer's tax inputs (taxpayer_tax_inputs,
// M5-20/M5-21).
type TaxInputReader interface {
	// TaxInputs returns each regime's inputs for the fiscal year in one
	// query; a regime without inputs is absent (the zero TaxInputs).
	TaxInputs(ctx context.Context, taxpayerID string, fiscalYear int) (map[fiscal.RegimeCode]domain.TaxInputs, error)
}
