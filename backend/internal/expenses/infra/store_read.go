package infra

import (
	"context"

	"gorm.io/gorm"

	"github.com/romodevx/control-contable/backend/internal/expenses/usecase"
	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
	"github.com/romodevx/control-contable/backend/internal/platform/pgdb"
)

// write runs a statement that returns one expenses row; no row means RLS
// (or the id) matched nothing.
func write(query *gorm.DB, op string) (usecase.StoredExpense, error) {
	var rows []expenseRow
	if err := query.Scan(&rows).Error; err != nil {
		return usecase.StoredExpense{}, translate(op, err)
	}
	if len(rows) == 0 {
		return usecase.StoredExpense{}, usecase.ErrExpenseNotFound
	}
	return rows[0].toStored()
}

// CFDIInUse implements usecase.ExpenseStore over the live-folio partial
// index (expenses_live_cfdi_uuid_idx).
func (Store) CFDIInUse(ctx context.Context, taxpayerID, exceptID string, folio fiscal.CFDIUUID) (bool, error) {
	db, err := pgdb.Conn(ctx)
	if err != nil {
		return false, err
	}
	query := db.Model(&expenseRow{}).Select("1").
		Where("taxpayer_id = ? AND cfdi_uuid = ? AND deleted_at IS NULL", taxpayerID, folio.String())
	if exceptID != "" {
		query = query.Where("id <> ?", exceptID)
	}
	var found []int
	if err := query.Limit(1).Scan(&found).Error; err != nil {
		return false, translate("find folio fiscal", err)
	}
	return len(found) > 0, nil
}

// live is the period's live expenses, over the expenses_live_paid_on_idx
// partial index (RLS adds user_id, its leading column).
func live(db *gorm.DB, q usecase.PeriodQuery) *gorm.DB {
	query := db.Model(&expenseRow{}).
		Where("taxpayer_id = ? AND deleted_at IS NULL AND paid_on >= ?::date AND paid_on < ?::date",
			q.TaxpayerID, q.From.String(), q.Until.String())
	if q.Regime != "" {
		query = query.Where("regime_code = ?", string(q.Regime))
	}
	return query
}

// List implements usecase.ExpenseStore: keyset order (paid_on, id) DESC.
func (Store) List(ctx context.Context, q usecase.ListQuery) ([]usecase.StoredExpense, error) {
	db, err := pgdb.Conn(ctx)
	if err != nil {
		return nil, err
	}
	query := live(db, q.PeriodQuery)
	if q.After != nil {
		query = query.Where("(paid_on, id) < (?::date, ?::uuid)", q.After.PaidOn.String(), q.After.ID)
	}
	var rows []expenseRow
	if err := query.Select(columns).Order("paid_on DESC, id DESC").Limit(q.Limit).Find(&rows).Error; err != nil {
		return nil, translate("list expenses", err)
	}
	out := make([]usecase.StoredExpense, 0, len(rows))
	for _, r := range rows {
		e, err := r.toStored()
		if err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, nil
}

// totalsRow is the period's sums; bigint sums are numeric, cast back.
type totalsRow struct {
	Count       int   `gorm:"column:count"`
	Amount      int64 `gorm:"column:amount"`
	IVA         int64 `gorm:"column:iva"`
	ISRWithheld int64 `gorm:"column:isr_withheld"`
	IVAWithheld int64 `gorm:"column:iva_withheld"`
}

// Totals implements usecase.ExpenseStore.
func (Store) Totals(ctx context.Context, q usecase.PeriodQuery) (usecase.Totals, error) {
	db, err := pgdb.Conn(ctx)
	if err != nil {
		return usecase.Totals{}, err
	}
	var row totalsRow
	err = live(db, q).Select(`count(*) AS count,
		coalesce(sum(amount_cents), 0)::bigint AS amount,
		coalesce(sum(iva_cents), 0)::bigint AS iva,
		coalesce(sum(isr_withheld_cents), 0)::bigint AS isr_withheld,
		coalesce(sum(iva_withheld_cents), 0)::bigint AS iva_withheld`).Scan(&row).Error
	if err != nil {
		return usecase.Totals{}, translate("sum expenses", err)
	}
	return usecase.Totals{
		Count: row.Count, Amount: money.Cents(row.Amount), IVA: money.Cents(row.IVA),
		ISRWithheld: money.Cents(row.ISRWithheld), IVAWithheld: money.Cents(row.IVAWithheld),
	}, nil
}
