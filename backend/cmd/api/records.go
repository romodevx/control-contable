package main

import (
	"context"
	"fmt"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	tpdomain "github.com/romodevx/control-contable/backend/internal/taxpayers/domain"
	tpusecase "github.com/romodevx/control-contable/backend/internal/taxpayers/usecase"
)

// latestExpenses is the expenses store's lookup of a regime's latest live
// expense (expinfra.Store).
type latestExpenses interface {
	LatestPaidOn(ctx context.Context, taxpayerID string, regime fiscal.RegimeCode) (fiscal.Date, error)
}

// latestIncomes is the incomes store's lookup of the latest day a live
// income holds a regime on (incomesinfra.Store).
type latestIncomes interface {
	LatestRecordDate(ctx context.Context, taxpayerID string, regime fiscal.RegimeCode) (fiscal.Date, error)
}

// regimeRecords implements the taxpayers RegimeRecords port over the
// expenses and incomes tables (M2-14 G-1: it replaces infra.NoRecords): the
// later of both records' periods.
//
// It runs on the taxpayers unit of work in ctx, as its user, after the
// regime end path locked the taxpayer_regimes row FOR NO KEY UPDATE. Every
// record write path checks its window through CheckRegime in its own unit
// of work, which takes FOR SHARE on that row: a write that checked first
// commits before the lookup runs (READ COMMITTED: the lookup's statement
// sees it), and one that checks later waits and reads the ended window.
type regimeRecords struct {
	expenses latestExpenses
	incomes  latestIncomes
}

var _ tpusecase.RegimeRecords = regimeRecords{}

// LatestRecordPeriod implements tpusecase.RegimeRecords.
func (r regimeRecords) LatestRecordPeriod(ctx context.Context, taxpayerID, code string) (tpdomain.Period, error) {
	regime := fiscal.RegimeCode(code)
	expense, err := r.expenses.LatestPaidOn(ctx, taxpayerID, regime)
	if err != nil {
		return tpdomain.Period{}, err
	}
	income, err := r.incomes.LatestRecordDate(ctx, taxpayerID, regime)
	if err != nil {
		return tpdomain.Period{}, err
	}
	latest := expense
	if income.After(latest) {
		latest = income
	}
	if latest.IsZero() {
		return tpdomain.Period{}, nil
	}
	p, err := tpdomain.ParsePeriod(latest.Period())
	if err != nil {
		return tpdomain.Period{}, fmt.Errorf("records: period of %s: %w", latest, err)
	}
	return p, nil
}
