// Package infra is the incomes module's PostgreSQL adapter (GORM, ADR-0001)
// over the incomes table of migration 20261013200000_incomes. It runs as
// cc_app inside the use case's unit of work (pgdb.Conn), so RLS shows and
// accepts the context user's rows only. It never begins, commits or rolls
// back, and never sets session state (P0-30 D-3). booked_on, deleted_at's
// clock and the audit columns belong to the database (trigger, now()).
package infra

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/romodevx/control-contable/backend/internal/incomes/domain"
	"github.com/romodevx/control-contable/backend/internal/incomes/usecase"
	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
	"github.com/romodevx/control-contable/backend/internal/platform/pgdb"
)

const table = "incomes"

// sqlstateUniqueViolation and the primary key it names on a client id taken
// by another user (the PK sees across RLS).
const (
	sqlstateUniqueViolation = "23505"
	primaryKey              = "incomes_pkey"
)

// incomeIDLockClass is the first key of the client-id advisory locks
// (two-int form, objsubid 2): it keeps them apart from the worker's
// single-key leader lock and from the expenses' id locks.
const incomeIDLockClass int32 = 0x494E4331 // "INC1"

// Store implements usecase.Store.
type Store struct{}

var _ usecase.Store = Store{}

// NewStore returns a Store. It holds no pool.
func NewStore() Store { return Store{} }

// incomeRow is what the use cases read.
type incomeRow struct {
	ID               string     `gorm:"column:id"`
	TaxpayerID       string     `gorm:"column:taxpayer_id"`
	RegimeCode       string     `gorm:"column:regime_code"`
	CounterpartyType string     `gorm:"column:counterparty_type"`
	AmountCents      int64      `gorm:"column:amount_cents"`
	IVARate          string     `gorm:"column:iva_rate"`
	IVACents         int64      `gorm:"column:iva_cents"`
	ISRWithheldCents int64      `gorm:"column:isr_withheld_cents"`
	IVAWithheldCents int64      `gorm:"column:iva_withheld_cents"`
	InvoicedOn       *time.Time `gorm:"column:invoiced_on"`
	CollectedOn      *time.Time `gorm:"column:collected_on"`
	BookedOn         time.Time  `gorm:"column:booked_on"`
	PlatformActivity *string    `gorm:"column:platform_activity"`
	CFDIUUID         *string    `gorm:"column:cfdi_uuid"`
	Description      *string    `gorm:"column:description"`
	DeletedAt        *time.Time `gorm:"column:deleted_at"`
	CreatedAt        time.Time  `gorm:"column:created_at"`
	UpdatedAt        time.Time  `gorm:"column:updated_at"`
}

var incomeColumns = []string{
	"id", "taxpayer_id", "regime_code", "counterparty_type", "amount_cents", "iva_rate", "iva_cents",
	"isr_withheld_cents", "iva_withheld_cents", "invoiced_on", "collected_on", "booked_on",
	"platform_activity", "cfdi_uuid", "description", "deleted_at", "created_at", "updated_at",
}

func (r incomeRow) toStored() (usecase.StoredIncome, error) {
	invoiced, err := dateOf(r.InvoicedOn)
	if err != nil {
		return usecase.StoredIncome{}, err
	}
	collected, err := dateOf(r.CollectedOn)
	if err != nil {
		return usecase.StoredIncome{}, err
	}
	booked, err := dateOf(&r.BookedOn)
	if err != nil {
		return usecase.StoredIncome{}, err
	}
	var folio fiscal.CFDIUUID
	if r.CFDIUUID != nil {
		if folio, err = fiscal.ParseCFDIUUID(*r.CFDIUUID); err != nil {
			return usecase.StoredIncome{}, fmt.Errorf("incomes infra: stored cfdi_uuid: %w", err)
		}
	}
	return usecase.StoredIncome{
		ID: r.ID, TaxpayerID: r.TaxpayerID,
		Fields: usecase.Fields{
			Regime: fiscal.RegimeCode(r.RegimeCode), Counterparty: domain.Counterparty(r.CounterpartyType),
			Amount: money.Cents(r.AmountCents), IVARate: fiscal.IVARate(r.IVARate), IVA: money.Cents(r.IVACents),
			Withheld:         fiscal.Withholdings{ISR: money.Cents(r.ISRWithheldCents), IVA: money.Cents(r.IVAWithheldCents)},
			Dates:            domain.Dates{InvoicedOn: invoiced, CollectedOn: collected},
			PlatformActivity: domain.PlatformActivity(deref(r.PlatformActivity)),
			CFDIUUID:         folio, Description: deref(r.Description),
		},
		BookedOn: booked, Deleted: r.DeletedAt != nil, CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt,
	}, nil
}

// newIncomeRow is exactly the columns cc_app may INSERT.
type newIncomeRow struct {
	ID               string  `gorm:"column:id"`
	UserID           string  `gorm:"column:user_id"`
	TaxpayerID       string  `gorm:"column:taxpayer_id"`
	RegimeCode       string  `gorm:"column:regime_code"`
	CounterpartyType string  `gorm:"column:counterparty_type"`
	AmountCents      int64   `gorm:"column:amount_cents"`
	IVARate          string  `gorm:"column:iva_rate"`
	IVACents         int64   `gorm:"column:iva_cents"`
	ISRWithheldCents int64   `gorm:"column:isr_withheld_cents"`
	IVAWithheldCents int64   `gorm:"column:iva_withheld_cents"`
	InvoicedOn       *string `gorm:"column:invoiced_on"`
	CollectedOn      *string `gorm:"column:collected_on"`
	PlatformActivity *string `gorm:"column:platform_activity"`
	CFDIUUID         *string `gorm:"column:cfdi_uuid"`
	Description      *string `gorm:"column:description"`
}

func (newIncomeRow) TableName() string { return table }

// LockID implements usecase.Store: a transaction-scoped advisory lock on
// the id, released at commit or rollback.
func (Store) LockID(ctx context.Context, id string) error {
	db, err := pgdb.Conn(ctx)
	if err != nil {
		return err
	}
	if err := db.Exec("SELECT pg_advisory_xact_lock(?, hashtext(?))", incomeIDLockClass, id).Error; err != nil {
		return fmt.Errorf("incomes infra: lock income id: %w", err)
	}
	return nil
}

// Find implements usecase.Store.
func (Store) Find(ctx context.Context, id string, lock usecase.Lock) (usecase.StoredIncome, error) {
	db, err := pgdb.Conn(ctx)
	if err != nil {
		return usecase.StoredIncome{}, err
	}
	query := db.Table(table).Select(incomeColumns).Where("id = ?", id).Limit(1)
	if lock == usecase.LockUpdate {
		query = query.Clauses(clause.Locking{Strength: "NO KEY UPDATE"})
	}
	var rows []incomeRow
	if err := query.Find(&rows).Error; err != nil {
		return usecase.StoredIncome{}, fmt.Errorf("incomes infra: find income: %w", err)
	}
	if len(rows) == 0 {
		return usecase.StoredIncome{}, usecase.ErrIncomeNotFound
	}
	return rows[0].toStored()
}

// Insert implements usecase.Store.
func (Store) Insert(ctx context.Context, in usecase.NewIncome) error {
	db, err := pgdb.Conn(ctx)
	if err != nil {
		return err
	}
	c := columnsOf(in.Fields)
	row := newIncomeRow{
		ID: in.ID, UserID: in.UserID, TaxpayerID: in.TaxpayerID,
		RegimeCode: c.regime, CounterpartyType: c.counterparty, AmountCents: c.amount, IVARate: c.ivaRate,
		IVACents: c.iva, ISRWithheldCents: c.isrWithheld, IVAWithheldCents: c.ivaWithheld,
		InvoicedOn: c.invoicedOn, CollectedOn: c.collectedOn, PlatformActivity: c.activity,
		CFDIUUID: c.cfdiUUID, Description: c.description,
	}
	if err := db.Create(&row).Error; err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok &&
			pgErr.Code == sqlstateUniqueViolation && pgErr.ConstraintName == primaryKey {
			return fmt.Errorf("%w (incomes infra: insert income: %w)", usecase.ErrIDTaken, err)
		}
		return fmt.Errorf("incomes infra: insert income: %w", err)
	}
	return nil
}

// Update implements usecase.Store. The trigger re-derives booked_on.
func (Store) Update(ctx context.Context, id string, f usecase.Fields) error {
	c := columnsOf(f)
	return update(ctx, "update income", id, map[string]any{
		"regime_code": c.regime, "counterparty_type": c.counterparty, "amount_cents": c.amount,
		"iva_rate": c.ivaRate, "iva_cents": c.iva, "isr_withheld_cents": c.isrWithheld,
		"iva_withheld_cents": c.ivaWithheld, "invoiced_on": c.invoicedOn, "collected_on": c.collectedOn,
		"platform_activity": c.activity, "cfdi_uuid": c.cfdiUUID, "description": c.description,
	})
}

// SetDeleted implements usecase.Store: deleted_at is the database's now().
func (Store) SetDeleted(ctx context.Context, id string, deleted bool) error {
	var deletedAt any
	if deleted {
		deletedAt = gorm.Expr("now()")
	}
	return update(ctx, "set deleted_at", id, map[string]any{"deleted_at": deletedAt})
}

func update(ctx context.Context, op, id string, columns map[string]any) error {
	db, err := pgdb.Conn(ctx)
	if err != nil {
		return err
	}
	result := db.Table(table).Where("id = ?", id).Updates(columns)
	if result.Error != nil {
		return fmt.Errorf("incomes infra: %s: %w", op, result.Error)
	}
	if result.RowsAffected == 0 {
		return usecase.ErrIncomeNotFound
	}
	return nil
}

// CFDIUUIDInUse implements usecase.Store (index incomes_cfdi_uuid_idx).
func (Store) CFDIUUIDInUse(ctx context.Context, taxpayerID string, uuid fiscal.CFDIUUID, exceptID string) (bool, error) {
	db, err := pgdb.Conn(ctx)
	if err != nil {
		return false, err
	}
	var inUse bool
	err = db.Raw(`SELECT EXISTS (
		SELECT 1 FROM incomes
		WHERE taxpayer_id = ? AND cfdi_uuid = ? AND deleted_at IS NULL AND id <> ?)`,
		taxpayerID, uuid.String(), exceptID).Scan(&inUse).Error
	if err != nil {
		return false, fmt.Errorf("incomes infra: find cfdi uuid: %w", err)
	}
	return inUse, nil
}

// List implements usecase.Store: keyset (booked_on DESC, id DESC) over the
// partial index incomes_booked_idx, scanned backwards.
func (Store) List(ctx context.Context, f usecase.ListFilter, after usecase.Keyset, limit int) ([]usecase.StoredIncome, error) {
	db, err := pgdb.Conn(ctx)
	if err != nil {
		return nil, err
	}
	query := filtered(db, f).Select(incomeColumns).Order("booked_on DESC, id DESC").Limit(limit)
	if after.ID != "" {
		query = query.Where("(booked_on, id) < (?::date, ?::uuid)", after.BookedOn.String(), after.ID)
	}
	var rows []incomeRow
	if err := query.Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("incomes infra: list incomes: %w", err)
	}
	out := make([]usecase.StoredIncome, len(rows))
	for i, r := range rows {
		if out[i], err = r.toStored(); err != nil {
			return nil, err
		}
	}
	return out, nil
}

// totalsRow is the sums of a filter. sum(bigint) is numeric: cast back.
type totalsRow struct {
	Count       int64 `gorm:"column:count"`
	Amount      int64 `gorm:"column:amount"`
	IVA         int64 `gorm:"column:iva"`
	ISRWithheld int64 `gorm:"column:isr_withheld"`
	IVAWithheld int64 `gorm:"column:iva_withheld"`
}

// Totals implements usecase.Store.
func (Store) Totals(ctx context.Context, f usecase.ListFilter) (usecase.PeriodTotals, error) {
	db, err := pgdb.Conn(ctx)
	if err != nil {
		return usecase.PeriodTotals{}, err
	}
	var t totalsRow
	err = filtered(db, f).Select(`count(*) AS count,
		COALESCE(sum(amount_cents), 0)::bigint AS amount,
		COALESCE(sum(iva_cents), 0)::bigint AS iva,
		COALESCE(sum(isr_withheld_cents), 0)::bigint AS isr_withheld,
		COALESCE(sum(iva_withheld_cents), 0)::bigint AS iva_withheld`).Scan(&t).Error
	if err != nil {
		return usecase.PeriodTotals{}, fmt.Errorf("incomes infra: sum incomes: %w", err)
	}
	return usecase.PeriodTotals{
		Count: t.Count, Amount: money.Cents(t.Amount), IVA: money.Cents(t.IVA),
		ISRWithheld: money.Cents(t.ISRWithheld), IVAWithheld: money.Cents(t.IVAWithheld),
	}, nil
}

// filtered is the live incomes of f's taxpayer booked in [From, Until).
func filtered(db *gorm.DB, f usecase.ListFilter) *gorm.DB {
	query := db.Table(table).
		Where("taxpayer_id = ? AND deleted_at IS NULL", f.TaxpayerID).
		Where("booked_on >= ?::date AND booked_on < ?::date", f.From.String(), f.Until.String())
	if f.Regime != "" {
		query = query.Where("regime_code = ?", string(f.Regime))
	}
	return query
}

// IncomeBasis implements usecase.Store: the catalog row of the regime for
// the person type (regimes.income_basis).
func (Store) IncomeBasis(ctx context.Context, person fiscal.PersonType, regime fiscal.RegimeCode) (domain.IncomeBasis, error) {
	db, err := pgdb.Conn(ctx)
	if err != nil {
		return "", err
	}
	var bases []string
	err = db.Table("regimes").Where("code = ? AND person_type = ?", string(regime), string(person)).
		Pluck("income_basis", &bases).Error
	if err != nil {
		return "", fmt.Errorf("incomes infra: read income basis: %w", err)
	}
	if len(bases) == 0 {
		return "", fmt.Errorf("incomes infra: regime %s/%s: %w", person, regime, errNoCatalogRow)
	}
	return domain.ParseIncomeBasis(bases[0])
}

// errNoCatalogRow reports a regime the catalog does not list for the person
// type: the composite FK makes it unreachable for a held regime.
var errNoCatalogRow = errors.New("no catalog row")

// columns are Fields as the database spells them; nil is NULL.
type columns struct {
	regime, counterparty, ivaRate         string
	amount, iva, isrWithheld, ivaWithheld int64
	invoicedOn, collectedOn               *string
	activity, cfdiUUID, description       *string
}

func columnsOf(f usecase.Fields) columns {
	return columns{
		regime: string(f.Regime), counterparty: string(f.Counterparty), ivaRate: string(f.IVARate),
		amount: int64(f.Amount), iva: int64(f.IVA), isrWithheld: int64(f.Withheld.ISR), ivaWithheld: int64(f.Withheld.IVA),
		invoicedOn: nullable(f.Dates.InvoicedOn.String()), collectedOn: nullable(f.Dates.CollectedOn.String()),
		activity: nullable(string(f.PlatformActivity)), cfdiUUID: nullable(f.CFDIUUID.String()),
		description: nullable(f.Description),
	}
}

// nullable is nil for "" (NULL), else a pointer to s.
func nullable(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// dateOf reads a date column (midnight UTC) as a calendar day; NULL is the
// zero Date.
func dateOf(t *time.Time) (fiscal.Date, error) {
	if t == nil {
		return fiscal.Date{}, nil
	}
	d, err := fiscal.ParseDate(t.UTC().Format(time.DateOnly))
	if err != nil {
		return fiscal.Date{}, fmt.Errorf("incomes infra: stored date: %w", err)
	}
	return d, nil
}
