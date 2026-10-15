// Package infra is the PostgreSQL adapter of the records retention job: it
// calls purge_soft_deleted (M9-08) on the unit of work in ctx.
package infra

import (
	"context"
	"fmt"

	"github.com/romodevx/control-contable/backend/internal/platform/pgdb"
	"github.com/romodevx/control-contable/backend/internal/retention/usecase"
)

// purgeSQL runs one bounded pass. The function is SECURITY DEFINER (owner
// cc_worker_definer) and EXECUTE is granted to cc_worker only; it computes
// its own cutoffs and returns one row per target, in a fixed order.
const purgeSQL = "SELECT target, purged FROM purge_soft_deleted(?)"

// Purger implements usecase.Purger over purge_soft_deleted.
type Purger struct{}

var _ usecase.Purger = Purger{}

// NewPurger returns the adapter.
func NewPurger() Purger { return Purger{} }

// purgeRow is one row of purge_soft_deleted.
type purgeRow struct {
	Target string
	Purged int
}

// PurgeBatch implements usecase.Purger, in the unit of work in ctx (normally
// System: the function needs no user context).
func (Purger) PurgeBatch(ctx context.Context, batchSize int) ([]usecase.TargetCount, error) {
	db, err := pgdb.Conn(ctx)
	if err != nil {
		return nil, err
	}
	var rows []purgeRow
	if err := db.Raw(purgeSQL, batchSize).Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("purge_soft_deleted: %w", err)
	}
	counts := make([]usecase.TargetCount, len(rows))
	for i, r := range rows {
		counts[i] = usecase.TargetCount{Target: r.Target, Purged: r.Purged}
	}
	return counts, nil
}
