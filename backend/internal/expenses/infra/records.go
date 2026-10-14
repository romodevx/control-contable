package infra

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/pgdb"
)

// LatestPaidOn returns the latest paid_on of the taxpayer's live expenses
// of regime, or the zero Date when there is none. It serves the taxpayers
// module's record lookup (MODULE-taxpayers §2.6: a regime ends only after
// its last record; M2-14 G-1), which cmd/api adapts. A soft-deleted expense
// holds nothing: its restore re-checks the regime window. It runs on the
// unit of work in ctx, as its user (RLS), over
// expenses_taxpayer_regime_fkey_idx.
func (Store) LatestPaidOn(ctx context.Context, taxpayerID string, regime fiscal.RegimeCode) (fiscal.Date, error) {
	db, err := pgdb.Conn(ctx)
	if err != nil {
		return fiscal.Date{}, err
	}
	var latest sql.NullTime
	err = db.Model(&expenseRow{}).
		Where("taxpayer_id = ? AND regime_code = ? AND deleted_at IS NULL", taxpayerID, string(regime)).
		Select("max(paid_on)").Scan(&latest).Error
	if err != nil {
		return fiscal.Date{}, translate("latest expense of regime", err)
	}
	if !latest.Valid {
		return fiscal.Date{}, nil
	}
	day, err := fiscal.ParseDate(latest.Time.Format(dateLayout))
	if err != nil {
		return fiscal.Date{}, fmt.Errorf("expenses infra: latest paid_on: %w", err)
	}
	return day, nil
}
