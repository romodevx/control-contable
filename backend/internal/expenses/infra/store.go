package infra

import (
	"context"
	"fmt"
	"time"

	"github.com/romodevx/control-contable/backend/internal/expenses/domain"
	"github.com/romodevx/control-contable/backend/internal/expenses/usecase"
	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
	"github.com/romodevx/control-contable/backend/internal/platform/pgdb"
)

// expenseIDLockClass is the first key of the client-id advisory locks
// (two-int form, objsubid 2): it keeps them apart from the worker's
// single-key leader lock and from any other module's id locks.
const expenseIDLockClass int32 = 0x45585031 // "EXP1"

// dateLayout is a paid_on as the stores pass and read it.
const dateLayout = "2006-01-02"

// Store implements usecase.ExpenseStore with GORM on the unit of work in
// ctx (pgdb.Conn), as cc_app: RLS shows only the context user's rows.
type Store struct{}

var _ usecase.ExpenseStore = Store{}

// NewStore returns a Store. It holds no pool.
func NewStore() Store { return Store{} }

// expenseRow is an expenses row as the stores read it.
type expenseRow struct {
	ID                   string     `gorm:"column:id"`
	TaxpayerID           string     `gorm:"column:taxpayer_id"`
	RegimeCode           string     `gorm:"column:regime_code"`
	CategoryCode         string     `gorm:"column:category_code"`
	Description          *string    `gorm:"column:description"`
	AmountCents          int64      `gorm:"column:amount_cents"`
	IVARate              string     `gorm:"column:iva_rate"`
	IVACents             int64      `gorm:"column:iva_cents"`
	PaymentMethod        string     `gorm:"column:payment_method"`
	HasCFDI              bool       `gorm:"column:has_cfdi"`
	CFDIUUID             *string    `gorm:"column:cfdi_uuid"`
	Deductible           bool       `gorm:"column:deductible"`
	DeductibleOverridden bool       `gorm:"column:deductible_overridden"`
	ISRWithheldCents     int64      `gorm:"column:isr_withheld_cents"`
	IVAWithheldCents     int64      `gorm:"column:iva_withheld_cents"`
	PaidOn               time.Time  `gorm:"column:paid_on"`
	DeletedAt            *time.Time `gorm:"column:deleted_at"`
	CreatedAt            time.Time  `gorm:"column:created_at"`
	UpdatedAt            time.Time  `gorm:"column:updated_at"`
}

func (expenseRow) TableName() string { return "expenses" }

// columns lists every column of an expenseRow.
const columns = `id, taxpayer_id, regime_code, category_code, description, amount_cents, iva_rate, iva_cents,
	payment_method, has_cfdi, cfdi_uuid, deductible, deductible_overridden, isr_withheld_cents, iva_withheld_cents,
	paid_on, deleted_at, created_at, updated_at`

// returning hands back the written row.
const returning = " RETURNING " + columns

// toStored maps a row to the use case's type. The CHECKs guarantee the
// enums, dates and folio formats; a value that still fails to parse is a
// server fault, never a client error.
func (r expenseRow) toStored() (usecase.StoredExpense, error) {
	paidOn, err := fiscal.ParseDate(r.PaidOn.Format(dateLayout))
	if err != nil {
		return usecase.StoredExpense{}, fmt.Errorf("expenses infra: paid_on of %s: %w", r.ID, err)
	}
	var folio fiscal.CFDIUUID
	if r.CFDIUUID != nil {
		if folio, err = fiscal.ParseCFDIUUID(*r.CFDIUUID); err != nil {
			return usecase.StoredExpense{}, fmt.Errorf("expenses infra: cfdi_uuid of %s: %w", r.ID, err)
		}
	}
	e := usecase.StoredExpense{
		ID: r.ID, TaxpayerID: r.TaxpayerID, Regime: fiscal.RegimeCode(r.RegimeCode),
		Category: domain.CategoryCode(r.CategoryCode), Amount: money.Cents(r.AmountCents),
		IVARate: fiscal.IVARate(r.IVARate), IVA: money.Cents(r.IVACents),
		PaymentMethod: domain.PaymentMethod(r.PaymentMethod), PaidOn: paidOn, HasCFDI: r.HasCFDI, CFDIUUID: folio,
		Deductible: r.Deductible, Overridden: r.DeductibleOverridden,
		Withheld:  fiscal.Withholdings{ISR: money.Cents(r.ISRWithheldCents), IVA: money.Cents(r.IVAWithheldCents)},
		Deleted:   r.DeletedAt != nil,
		CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt,
	}
	if r.Description != nil {
		e.Description = *r.Description
	}
	return e, nil
}

// nullable is "" as NULL.
func nullable(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// folioOf is a folio fiscal as its column value (NULL when absent).
func folioOf(u fiscal.CFDIUUID) *string { return nullable(u.String()) }

// LockID implements usecase.ExpenseStore: a transaction-scoped advisory
// lock on the id, released when the unit of work ends.
func (Store) LockID(ctx context.Context, id string) error {
	db, err := pgdb.Conn(ctx)
	if err != nil {
		return err
	}
	return translate("lock expense id", db.Exec("SELECT pg_advisory_xact_lock(?, hashtext(?))", expenseIDLockClass, id).Error)
}

// Find implements usecase.ExpenseStore.
func (Store) Find(ctx context.Context, id string, lock usecase.Lock) (usecase.StoredExpense, error) {
	db, err := pgdb.Conn(ctx)
	if err != nil {
		return usecase.StoredExpense{}, err
	}
	query := "SELECT " + columns + " FROM expenses WHERE id = ?"
	if lock == usecase.LockUpdate {
		query += " FOR NO KEY UPDATE"
	}
	var rows []expenseRow
	if err := db.Raw(query, id).Scan(&rows).Error; err != nil {
		return usecase.StoredExpense{}, translate("read expense", err)
	}
	if len(rows) == 0 {
		return usecase.StoredExpense{}, usecase.ErrExpenseNotFound
	}
	return rows[0].toStored()
}

// Insert implements usecase.ExpenseStore. It names exactly the columns
// cc_app may INSERT.
func (Store) Insert(ctx context.Context, userID string, e usecase.StoredExpense) (usecase.StoredExpense, error) {
	db, err := pgdb.Conn(ctx)
	if err != nil {
		return usecase.StoredExpense{}, err
	}
	statement := `INSERT INTO expenses (id, user_id, taxpayer_id, regime_code, category_code, description,
		amount_cents, iva_rate, iva_cents, payment_method, has_cfdi, cfdi_uuid, deductible, deductible_overridden,
		isr_withheld_cents, iva_withheld_cents, paid_on)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?::date)` + returning
	args := []any{e.ID, userID, e.TaxpayerID}
	return write(db.Raw(statement, append(args, editable(e)...)...), "insert expense")
}

// Update implements usecase.ExpenseStore. It writes every editable column.
func (Store) Update(ctx context.Context, e usecase.StoredExpense) (usecase.StoredExpense, error) {
	db, err := pgdb.Conn(ctx)
	if err != nil {
		return usecase.StoredExpense{}, err
	}
	statement := `UPDATE expenses SET regime_code = ?, category_code = ?, description = ?, amount_cents = ?,
		iva_rate = ?, iva_cents = ?, payment_method = ?, has_cfdi = ?, cfdi_uuid = ?, deductible = ?,
		deductible_overridden = ?, isr_withheld_cents = ?, iva_withheld_cents = ?, paid_on = ?::date
		WHERE id = ?` + returning
	return write(db.Raw(statement, append(editable(e), e.ID)...), "update expense")
}

// editable is the editable columns of e, in the order Insert and Update
// name them.
func editable(e usecase.StoredExpense) []any {
	return []any{
		string(e.Regime), string(e.Category), nullable(e.Description), int64(e.Amount), string(e.IVARate),
		int64(e.IVA), string(e.PaymentMethod), e.HasCFDI, folioOf(e.CFDIUUID), e.Deductible, e.Overridden,
		int64(e.Withheld.ISR), int64(e.Withheld.IVA), e.PaidOn.String(),
	}
}

// SetDeleted implements usecase.ExpenseStore: deleted_at is the
// database's now() (soft delete) or NULL (restore).
func (Store) SetDeleted(ctx context.Context, id string, deleted bool) (usecase.StoredExpense, error) {
	db, err := pgdb.Conn(ctx)
	if err != nil {
		return usecase.StoredExpense{}, err
	}
	statement := "UPDATE expenses SET deleted_at = CASE WHEN ? THEN now() END WHERE id = ?" + returning
	return write(db.Raw(statement, deleted, id), "set expense deleted")
}
