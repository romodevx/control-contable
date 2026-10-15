// Package infra is History's persistence: the estimate_views table
// (migration 20261014180000, MODULE-history §2.3) through GORM, as cc_app
// under RLS. It holds no pool: every method runs on the unit of work in
// ctx, whose transaction set app.current_user_id.
package infra

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgconn"

	"github.com/romodevx/control-contable/backend/internal/history/domain"
	"github.com/romodevx/control-contable/backend/internal/history/usecase"
	"github.com/romodevx/control-contable/backend/internal/platform/pgdb"
)

// ViewStore implements usecase.ViewStore.
type ViewStore struct{}

var _ usecase.ViewStore = ViewStore{}

// NewViewStore returns a ViewStore.
func NewViewStore() ViewStore { return ViewStore{} }

// viewsQuery is the fiscal-year read (§2.6 step 6): RLS adds user_id, so
// the primary key (user_id, taxpayer_id, period) serves it as one range
// scan (EXPLAIN pinned by the M8-02 migration test).
const viewsQuery = `SELECT period, records_fingerprint, figures_fingerprint, engine_version, parameter_set, viewed_at
FROM estimate_views WHERE taxpayer_id = ? AND period BETWEEN ? AND ?`

// upsertSQL is the one write (§2.3): idempotent, the last view wins. The
// key columns never change (cc_app has no UPDATE grant on them).
const upsertSQL = `INSERT INTO estimate_views (user_id, taxpayer_id, period, records_fingerprint, figures_fingerprint,
  engine_version, parameter_set, viewed_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT (user_id, taxpayer_id, period) DO UPDATE SET
  records_fingerprint = EXCLUDED.records_fingerprint,
  figures_fingerprint = EXCLUDED.figures_fingerprint,
  engine_version      = EXCLUDED.engine_version,
  parameter_set       = EXCLUDED.parameter_set,
  viewed_at           = EXCLUDED.viewed_at`

// The composite foreign key (user_id, taxpayer_id) -> taxpayers: it fails
// for a taxpayer that is not the caller's (FK checks bypass RLS, so this is
// the ownership proof of a write).
const (
	sqlstateForeignKeyViolation = "23503"
	constraintTaxpayer          = "estimate_views_taxpayer_fkey"
)

// dateLayout is a period's first day as PostgreSQL's date.
const dateLayout = "2006-01-02"

// viewRow is one estimate_views row as read. char(64) comes back unpadded
// (the CHECK keeps it exactly 64).
type viewRow struct {
	Period             time.Time `gorm:"column:period"`
	RecordsFingerprint string    `gorm:"column:records_fingerprint"`
	FiguresFingerprint string    `gorm:"column:figures_fingerprint"`
	EngineVersion      string    `gorm:"column:engine_version"`
	ParameterSet       string    `gorm:"column:parameter_set"`
	ViewedAt           time.Time `gorm:"column:viewed_at"`
}

// Views implements usecase.ViewStore.
func (ViewStore) Views(ctx context.Context, taxpayerID string, from, to domain.Month) ([]domain.View, error) {
	db, err := pgdb.Conn(ctx)
	if err != nil {
		return nil, err
	}
	var rows []viewRow
	if err := db.Raw(viewsQuery, taxpayerID, firstDay(from), firstDay(to)).Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("history infra: views: %w", err)
	}
	out := make([]domain.View, 0, len(rows))
	for _, r := range rows {
		out = append(out, domain.View{
			Month: domain.MonthOf(r.Period),
			Seen: domain.Seen{
				RecordsFingerprint: r.RecordsFingerprint, FiguresFingerprint: r.FiguresFingerprint,
				EngineVersion: r.EngineVersion, ParameterSet: r.ParameterSet,
			},
			ViewedAt: r.ViewedAt,
		})
	}
	return out, nil
}

// SaveView implements usecase.ViewStore.
func (ViewStore) SaveView(ctx context.Context, userID, taxpayerID string, v domain.View) error {
	db, err := pgdb.Conn(ctx)
	if err != nil {
		return err
	}
	err = db.Exec(upsertSQL, userID, taxpayerID, firstDay(v.Month), v.Seen.RecordsFingerprint, v.Seen.FiguresFingerprint,
		v.Seen.EngineVersion, v.Seen.ParameterSet, v.ViewedAt).Error
	if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok &&
		pgErr.Code == sqlstateForeignKeyViolation && pgErr.ConstraintName == constraintTaxpayer {
		return fmt.Errorf("%w (history infra: save view: %w)", usecase.ErrTaxpayerNotFound, err)
	}
	if err != nil {
		return fmt.Errorf("history infra: save view: %w", err)
	}
	return nil
}

// firstDay is m's first day, "YYYY-MM-01" (the period column).
func firstDay(m domain.Month) string {
	return time.Date(m.Year(), m.Month(), 1, 0, 0, 0, 0, time.UTC).Format(dateLayout)
}
