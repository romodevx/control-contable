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
	CounterpartyKind *string    `gorm:"column:counterparty_kind"`
	OperationKind    *string    `gorm:"column:operation_kind"`
	ResidentialRent  bool       `gorm:"column:residential_rent"`
	ViaPlatform      bool       `gorm:"column:via_platform"`
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
	// Withholding is embedded by GORM; an anonymous unexported embed would be
	// skipped (it cannot be set by reflection).
	Withholding withholdingColumns `gorm:"embedded"`
}

// incomeColumns is the select list of incomeRow.
const incomeColumns = "id, taxpayer_id, regime_code, counterparty_type, counterparty_kind, operation_kind, " +
	"residential_rent, via_platform, amount_cents, iva_rate, iva_cents, isr_withheld_cents, iva_withheld_cents, " +
	"invoiced_on, collected_on, booked_on, platform_activity, cfdi_uuid, description, deleted_at, created_at, " +
	"updated_at, " + withholdingSelect

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
	withholding, err := r.Withholding.record(r.ISRWithheldCents, r.IVAWithheldCents)
	if err != nil {
		return usecase.StoredIncome{}, err
	}
	return usecase.StoredIncome{
		ID: r.ID, TaxpayerID: r.TaxpayerID,
		Fields: usecase.Fields{
			Regime: fiscal.RegimeCode(r.RegimeCode), Counterparty: fiscal.ReceptorKind(deref(r.CounterpartyKind)),
			Operation: fiscal.OperationKind(deref(r.OperationKind)), ResidentialRent: r.ResidentialRent,
			ViaPlatform: r.ViaPlatform, Amount: money.Cents(r.AmountCents), IVARate: fiscal.IVARate(r.IVARate),
			IVA:              money.Cents(r.IVACents),
			Dates:            domain.Dates{InvoicedOn: invoiced, CollectedOn: collected},
			PlatformActivity: domain.PlatformActivity(deref(r.PlatformActivity)),
			CFDIUUID:         folio, Description: deref(r.Description),
		},
		Withholding: withholding, CounterpartyType: domain.Counterparty(r.CounterpartyType),
		BookedOn: booked, Deleted: r.DeletedAt != nil, CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt,
	}, nil
}

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
	row := columnsOf(in.Fields, in.Withholding)
	row["id"], row["user_id"], row["taxpayer_id"] = in.ID, in.UserID, in.TaxpayerID
	if err := db.Table(table).Create(row).Error; err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok &&
			pgErr.Code == sqlstateUniqueViolation && pgErr.ConstraintName == primaryKey {
			return fmt.Errorf("%w (incomes infra: insert income: %w)", usecase.ErrIDTaken, err)
		}
		return fmt.Errorf("incomes infra: insert income: %w", err)
	}
	return nil
}

// Update implements usecase.Store. The trigger re-derives booked_on.
func (Store) Update(ctx context.Context, id string, f usecase.Fields, w fiscal.WithholdingRecord) error {
	return update(ctx, "update income", id, columnsOf(f, w))
}

// lastFactsRow is the facts LastFacts reads.
type lastFactsRow struct {
	CounterpartyKind string  `gorm:"column:counterparty_kind"`
	OperationKind    *string `gorm:"column:operation_kind"`
	ResidentialRent  bool    `gorm:"column:residential_rent"`
	ViaPlatform      bool    `gorm:"column:via_platform"`
}

// LastFacts implements usecase.Store: the newest live income of the
// regime with facts, over incomes_taxpayer_regime_idx (created_at is the
// transaction's now(); the time-ordered id breaks a tie).
func (Store) LastFacts(ctx context.Context, taxpayerID string, regime fiscal.RegimeCode) (usecase.LastFacts, bool, error) {
	db, err := pgdb.Conn(ctx)
	if err != nil {
		return usecase.LastFacts{}, false, err
	}
	var rows []lastFactsRow
	err = db.Table(table).Select("counterparty_kind, operation_kind, residential_rent, via_platform").
		Where("taxpayer_id = ? AND regime_code = ? AND deleted_at IS NULL AND counterparty_kind IS NOT NULL",
			taxpayerID, string(regime)).
		Order("created_at DESC, id DESC").Limit(1).Find(&rows).Error
	if err != nil {
		return usecase.LastFacts{}, false, fmt.Errorf("incomes infra: last income facts: %w", err)
	}
	if len(rows) == 0 {
		return usecase.LastFacts{}, false, nil
	}
	r := rows[0]
	return usecase.LastFacts{
		Counterparty: fiscal.ReceptorKind(r.CounterpartyKind), Operation: fiscal.OperationKind(deref(r.OperationKind)),
		ResidentialRent: r.ResidentialRent, ViaPlatform: r.ViaPlatform,
	}, true, nil
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
