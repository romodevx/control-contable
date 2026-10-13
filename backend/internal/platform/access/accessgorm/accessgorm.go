// Package accessgorm is the PostgreSQL adapter of the AccessPolicy ports
// (GORM, ADR-0001). It runs as cc_app inside the caller's unit of work
// (pgdb.Conn), so the RLS policies on entitlements and user_usage see the
// context user: another user's rows are invisible and untouchable.
package accessgorm

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/romodevx/control-contable/backend/internal/platform/access"
	"github.com/romodevx/control-contable/backend/internal/platform/pgdb"
)

// Store implements access.EntitlementReader and access.UsageCounter.
type Store struct{}

var (
	_ access.EntitlementReader = Store{}
	_ access.UsageCounter      = Store{}
)

// NewStore returns a Store. It holds no pool: every call uses the
// transaction of the unit of work in its ctx.
func NewStore() Store { return Store{} }

// entitlementRow maps the columns of entitlements the policy reads.
type entitlementRow struct {
	Kind     string
	StartsAt time.Time
	EndsAt   *time.Time
}

func (entitlementRow) TableName() string { return "entitlements" }

func (r entitlementRow) toDomain() access.Entitlement {
	e := access.Entitlement{Kind: access.EntitlementKind(r.Kind), StartsAt: r.StartsAt}
	if r.EndsAt != nil {
		e.EndsAt = *r.EndsAt
	}
	return e
}

// usageRow maps user_usage.
type usageRow struct {
	UserID          string `gorm:"primaryKey"`
	IncomesCreated  int
	ExpensesCreated int
}

func (usageRow) TableName() string { return "user_usage" }

// Entitlements returns every entitlement row of userID (at most one per kind,
// entitlements_user_kind_key).
func (Store) Entitlements(ctx context.Context, userID string) ([]access.Entitlement, error) {
	db, err := pgdb.Conn(ctx)
	if err != nil {
		return nil, err
	}
	var rows []entitlementRow
	if err := db.Select("kind", "starts_at", "ends_at").Where("user_id = ?", userID).Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("accessgorm: read entitlements: %w", err)
	}
	out := make([]access.Entitlement, 0, len(rows))
	for _, r := range rows {
		out = append(out, r.toDomain())
	}
	return out, nil
}

// reserveSQL is the trial-cap algorithm of the data model, one statement per
// counter column (a column name cannot be a bind parameter). The row lock
// taken by UPDATE serializes concurrent reservations, and under READ
// COMMITTED a waiter re-checks `< limit` against the committed winner's row,
// so the counter never passes the cap. ? = user id, unlimited, limit.
var reserveSQL = map[access.RecordKind]string{
	access.RecordIncome: `UPDATE user_usage SET incomes_created = incomes_created + 1
WHERE user_id = ? AND (? OR incomes_created < ?)
RETURNING incomes_created`,
	access.RecordExpense: `UPDATE user_usage SET expenses_created = expenses_created + 1
WHERE user_id = ? AND (? OR expenses_created < ?)
RETURNING expenses_created`,
}

// Reserve implements access.UsageCounter.
func (s Store) Reserve(ctx context.Context, userID string, kind access.RecordKind, limit int) (bool, error) {
	query, ok := reserveSQL[kind]
	if !ok {
		return false, fmt.Errorf("%w: %q", access.ErrUnknownRecordKind, kind)
	}
	db, err := pgdb.Conn(ctx)
	if err != nil {
		return false, err
	}
	var counts []int
	if err := db.Raw(query, userID, limit == access.Unlimited, limit).Scan(&counts).Error; err != nil {
		return false, fmt.Errorf("accessgorm: reserve %s: %w", kind, err)
	}
	if len(counts) == 1 {
		return true, nil
	}
	// No row updated: the cap, or no counter row at all (a data bug, never a 402).
	if _, err := s.read(db, userID); err != nil {
		return false, err
	}
	return false, nil
}

// Usage implements access.UsageCounter.
func (s Store) Usage(ctx context.Context, userID string) (access.Usage, error) {
	db, err := pgdb.Conn(ctx)
	if err != nil {
		return access.Usage{}, err
	}
	return s.read(db, userID)
}

func (Store) read(db *gorm.DB, userID string) (access.Usage, error) {
	var row usageRow
	err := db.Select("user_id", "incomes_created", "expenses_created").Where("user_id = ?", userID).Take(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return access.Usage{}, access.ErrUsageMissing
	}
	if err != nil {
		return access.Usage{}, fmt.Errorf("accessgorm: read usage: %w", err)
	}
	return access.Usage{Incomes: row.IncomesCreated, Expenses: row.ExpensesCreated}, nil
}
