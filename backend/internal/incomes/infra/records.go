package infra

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/pgdb"
)

// LatestRecordDate returns the latest day a live income of the taxpayer's
// regime holds that regime on, or the zero Date when there is none. An
// income holds its regime on both of its dates (the use cases check the
// window on each), so the day is the later of invoiced_on and collected_on
// (greatest() skips a NULL). It serves the taxpayers module's record lookup
// (MODULE-taxpayers §2.6: a regime ends only after its last record; M2-14
// G-1), which cmd/api adapts. A soft-deleted income holds nothing: its
// restore re-checks the regime window. It runs on the unit of work in ctx,
// as its user (RLS), over incomes_taxpayer_regime_idx.
func (Store) LatestRecordDate(ctx context.Context, taxpayerID string, regime fiscal.RegimeCode) (fiscal.Date, error) {
	db, err := pgdb.Conn(ctx)
	if err != nil {
		return fiscal.Date{}, err
	}
	var latest sql.NullTime
	err = db.Table(table).
		Where("taxpayer_id = ? AND regime_code = ? AND deleted_at IS NULL", taxpayerID, string(regime)).
		Select("max(greatest(invoiced_on, collected_on))").Scan(&latest).Error
	if err != nil {
		return fiscal.Date{}, fmt.Errorf("incomes infra: latest income of regime: %w", err)
	}
	if !latest.Valid {
		return fiscal.Date{}, nil
	}
	return dateOf(&latest.Time)
}
