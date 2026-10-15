package main

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/romodevx/control-contable/backend/cmd/internal/profiles"
	expdomain "github.com/romodevx/control-contable/backend/internal/expenses/domain"
	expinfra "github.com/romodevx/control-contable/backend/internal/expenses/infra"
	expenses "github.com/romodevx/control-contable/backend/internal/expenses/usecase"
	incinfra "github.com/romodevx/control-contable/backend/internal/incomes/infra"
	incomes "github.com/romodevx/control-contable/backend/internal/incomes/usecase"
	"github.com/romodevx/control-contable/backend/internal/platform/clock"
	"github.com/romodevx/control-contable/backend/internal/platform/pgdb"
	"github.com/romodevx/control-contable/backend/internal/platform/tx"
	"github.com/romodevx/control-contable/backend/internal/taxengine/infra/ratecatalog"
	"github.com/romodevx/control-contable/backend/internal/taxengine/params"
	tpinfra "github.com/romodevx/control-contable/backend/internal/taxpayers/infra"
	taxpayers "github.com/romodevx/control-contable/backend/internal/taxpayers/usecase"
)

// newBackfiller wires both modules' backfills over one cc_worker unit of
// work manager (pool), so the incomes and expenses of a user, and the
// taxpayer reads they make, share that user's transaction. now is the
// derivation's "today" (the domain refuses later dates).
func newBackfiller(pool *sql.DB, now clock.Clock) (*backfiller, error) {
	units, err := pgdb.NewTxManager(pool)
	if err != nil {
		return nil, err
	}
	sets, err := params.Embedded()
	if err != nil {
		return nil, fmt.Errorf("tax parameters: %w", err)
	}
	rates, err := ratecatalog.New[expdomain.AmountKey](sets)
	if err != nil {
		return nil, fmt.Errorf("rate catalog: %w", err)
	}
	store := tpinfra.NewStore()
	queries := taxpayers.NewQueries(taxpayers.Deps{Tx: units, Taxpayers: store, Regimes: store})
	return &backfiller{
		units: units,
		users: userStore{system: units},
		incomes: incomes.NewBackfill(incomes.BackfillDeps{
			Tx: units, Store: incinfra.NewStore(), Taxpayers: profiles.IncomesReader{Reader: queries},
			Withholdings: rates.Withholdings(), Clock: now,
		}),
		expenses: expenses.NewBackfill(expenses.BackfillDeps{
			Tx: units, Store: expinfra.NewStore(), Taxpayers: profiles.ExpensesReader{Reader: queries},
			Categories: expinfra.NewCategoryCatalog(), Parameters: rates, Withholdings: rates.Withholdings(), Clock: now,
		}),
	}, nil
}

// userStore lists the users through the System unit of work: no user
// context, and the users_select_worker policy shows every id (the only
// column cc_worker may read there).
type userStore struct {
	system tx.SystemManager
}

// UserIDs implements userLister, ordered by id.
func (s userStore) UserIDs(ctx context.Context) ([]string, error) {
	var ids []string
	err := s.system.System(ctx, func(ctx context.Context) error {
		db, err := pgdb.Conn(ctx)
		if err != nil {
			return err
		}
		return db.Raw("SELECT id::text FROM users ORDER BY id").Scan(&ids).Error
	})
	if err != nil {
		return nil, fmt.Errorf("read user ids: %w", err)
	}
	return ids, nil
}
