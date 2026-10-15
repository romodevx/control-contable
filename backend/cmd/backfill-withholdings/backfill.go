package main

import (
	"context"
	"fmt"

	expenses "github.com/romodevx/control-contable/backend/internal/expenses/usecase"
	incomes "github.com/romodevx/control-contable/backend/internal/incomes/usecase"
	"github.com/romodevx/control-contable/backend/internal/platform/tx"
)

// userLister lists every user to visit (users.id only, cc_worker's grant).
type userLister interface {
	UserIDs(ctx context.Context) ([]string, error)
}

// incomesBackfill and expensesBackfill are the modules' backfills
// (incomes.Backfill, expenses.Backfill): they join the unit of work in ctx.
type incomesBackfill interface {
	Run(ctx context.Context, userID string, apply bool) (incomes.BackfillResult, error)
}

type expensesBackfill interface {
	Run(ctx context.Context, userID string, apply bool) (expenses.BackfillResult, error)
}

// backfiller visits every user in one unit of work of their own
// (app.current_user_id set, so RLS scopes every read and write), runs both
// modules' backfills in it, and stops at the first error: the failing
// user's transaction rolls back, the users before it stay committed, and a
// re-run picks up where it stopped (the backfill is idempotent).
type backfiller struct {
	units    tx.Manager
	users    userLister
	incomes  incomesBackfill
	expenses expensesBackfill
}

// run backfills every user; apply false is a dry run in read-only units.
func (b *backfiller) run(ctx context.Context, apply bool) (report, error) {
	ids, err := b.users.UserIDs(ctx)
	if err != nil {
		return report{}, fmt.Errorf("list users: %w", err)
	}
	out := report{Apply: apply, Users: len(ids)}
	for _, userID := range ids {
		work := func(ctx context.Context) error { return b.user(ctx, userID, apply, &out) }
		if apply {
			err = b.units.ReadWrite(ctx, userID, work)
		} else {
			err = b.units.ReadOnly(ctx, userID, work)
		}
		if err != nil {
			return report{}, fmt.Errorf("user %s: %w", userID, err)
		}
	}
	return out, nil
}

// user backfills one user's incomes, then expenses, into out.
func (b *backfiller) user(ctx context.Context, userID string, apply bool, out *report) error {
	inc, err := b.incomes.Run(ctx, userID, apply)
	if err != nil {
		return err
	}
	exp, err := b.expenses.Run(ctx, userID, apply)
	if err != nil {
		return err
	}
	for _, r := range inc.Rederived {
		out.Incomes.record(r.ID, r.AmountsChanged, r.FactsAssumed)
	}
	out.Incomes.NeedsInput = append(out.Incomes.NeedsInput, inc.NeedsInput...)
	for _, r := range exp.Rederived {
		out.Expenses.record(r.ID, r.AmountsChanged, r.FactsAssumed)
	}
	out.Expenses.NeedsInput = append(out.Expenses.NeedsInput, exp.NeedsInput...)
	return nil
}
