package annualinputs

import (
	"context"
	"time"

	"gorm.io/gorm"

	"github.com/romodevx/control-contable/backend/internal/platform/money"
	"github.com/romodevx/control-contable/backend/internal/platform/pgdb"
	adomain "github.com/romodevx/control-contable/backend/internal/taxengine/domain/annual"
	"github.com/romodevx/control-contable/backend/internal/taxengine/usecase/annual"
)

const (
	deductionsTable      = "personal_deductions"
	deductionsPrimaryKey = "personal_deductions_pkey"
)

// deductionSelect reads a row as the use cases see it; paid_on as text is a
// calendar day with no time zone.
const deductionSelect = `SELECT id, taxpayer_id, fiscal_year, category, amount_cents, paid_on::text AS paid_on,
       payment_method, beneficiary, donee_type, student_label, tuition_level, scholarship_cents, cfdi_uuid,
       deleted_at IS NOT NULL AS deleted, created_at, updated_at
FROM personal_deductions`

// The read paths. The live ones are served by the partial index
// personal_deductions_live_year_idx (user_id from RLS, taxpayer_id,
// fiscal_year, category).
const (
	findDeductionQuery  = deductionSelect + " WHERE id = ?"
	listDeductionsQuery = deductionSelect + " WHERE taxpayer_id = ? AND fiscal_year = ? AND deleted_at IS NULL"
	countDeductionsSQL  = "SELECT count(*) FROM personal_deductions WHERE taxpayer_id = ? AND fiscal_year = ? AND deleted_at IS NULL"
	byCategory          = " AND category = ?"
)

type deductionRow struct {
	ID               string    `gorm:"column:id"`
	TaxpayerID       string    `gorm:"column:taxpayer_id"`
	FiscalYear       int       `gorm:"column:fiscal_year"`
	Category         string    `gorm:"column:category"`
	AmountCents      int64     `gorm:"column:amount_cents"`
	PaidOn           string    `gorm:"column:paid_on"`
	PaymentMethod    string    `gorm:"column:payment_method"`
	Beneficiary      string    `gorm:"column:beneficiary"`
	DoneeType        *string   `gorm:"column:donee_type"`
	StudentLabel     *string   `gorm:"column:student_label"`
	TuitionLevel     *string   `gorm:"column:tuition_level"`
	ScholarshipCents *int64    `gorm:"column:scholarship_cents"`
	CFDIUUID         *string   `gorm:"column:cfdi_uuid"`
	Deleted          bool      `gorm:"column:deleted"`
	CreatedAt        time.Time `gorm:"column:created_at"`
	UpdatedAt        time.Time `gorm:"column:updated_at"`
}

func (r deductionRow) toStored() (annual.StoredDeduction, error) {
	paidOn, err := dateOf(r.PaidOn)
	if err != nil {
		return annual.StoredDeduction{}, err
	}
	f := annual.DeductionFields{
		Category: adomain.Category(r.Category), Amount: money.Cents(r.AmountCents), PaidOn: paidOn,
		PaymentMethod: adomain.PaymentMethod(r.PaymentMethod), Beneficiary: adomain.Beneficiary(r.Beneficiary),
		DoneeType: adomain.DoneeType(deref(r.DoneeType)), StudentLabel: deref(r.StudentLabel),
		TuitionLevel: adomain.TuitionLevel(deref(r.TuitionLevel)), CFDIUUID: deref(r.CFDIUUID),
	}
	if r.ScholarshipCents != nil {
		f.Scholarship = annual.NullCents{Cents: money.Cents(*r.ScholarshipCents), Valid: true}
	}
	return annual.StoredDeduction{
		ID: r.ID, TaxpayerID: r.TaxpayerID, FiscalYear: r.FiscalYear, DeductionFields: f, Deleted: r.Deleted,
		CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt,
	}, nil
}

// deductionColumns are the writable columns of fields: a zero nullable
// field is NULL.
type deductionColumns struct {
	Category         string  `gorm:"column:category"`
	AmountCents      int64   `gorm:"column:amount_cents"`
	PaidOn           string  `gorm:"column:paid_on"`
	PaymentMethod    string  `gorm:"column:payment_method"`
	Beneficiary      string  `gorm:"column:beneficiary"`
	DoneeType        *string `gorm:"column:donee_type"`
	StudentLabel     *string `gorm:"column:student_label"`
	TuitionLevel     *string `gorm:"column:tuition_level"`
	ScholarshipCents *int64  `gorm:"column:scholarship_cents"`
	CFDIUUID         *string `gorm:"column:cfdi_uuid"`
}

func columnsOf(f annual.DeductionFields) deductionColumns {
	c := deductionColumns{
		Category: string(f.Category), AmountCents: int64(f.Amount), PaidOn: f.PaidOn.String(),
		PaymentMethod: string(f.PaymentMethod), Beneficiary: string(f.Beneficiary),
		DoneeType: nullable(string(f.DoneeType)), StudentLabel: nullable(f.StudentLabel),
		TuitionLevel: nullable(string(f.TuitionLevel)), CFDIUUID: nullable(f.CFDIUUID),
	}
	if f.Scholarship.Valid {
		cents := int64(f.Scholarship.Cents)
		c.ScholarshipCents = &cents
	}
	return c
}

func (c deductionColumns) updates() map[string]any {
	return map[string]any{
		"category": c.Category, "amount_cents": c.AmountCents, "paid_on": c.PaidOn, "payment_method": c.PaymentMethod,
		"beneficiary": c.Beneficiary, "donee_type": c.DoneeType, "student_label": c.StudentLabel,
		"tuition_level": c.TuitionLevel, "scholarship_cents": c.ScholarshipCents, "cfdi_uuid": c.CFDIUUID,
	}
}

// newDeductionRow is exactly the columns cc_app may INSERT: fiscal_year is
// generated, person_type defaults to fisica. (Flat, not embedding
// deductionColumns: GORM skips an unexported embedded struct.)
type newDeductionRow struct {
	ID               string  `gorm:"column:id"`
	UserID           string  `gorm:"column:user_id"`
	TaxpayerID       string  `gorm:"column:taxpayer_id"`
	Category         string  `gorm:"column:category"`
	AmountCents      int64   `gorm:"column:amount_cents"`
	PaidOn           string  `gorm:"column:paid_on"`
	PaymentMethod    string  `gorm:"column:payment_method"`
	Beneficiary      string  `gorm:"column:beneficiary"`
	DoneeType        *string `gorm:"column:donee_type"`
	StudentLabel     *string `gorm:"column:student_label"`
	TuitionLevel     *string `gorm:"column:tuition_level"`
	ScholarshipCents *int64  `gorm:"column:scholarship_cents"`
	CFDIUUID         *string `gorm:"column:cfdi_uuid"`
}

func newDeduction(d annual.NewDeduction) newDeductionRow {
	c := columnsOf(d.DeductionFields)
	return newDeductionRow{
		ID: d.ID, UserID: d.UserID, TaxpayerID: d.TaxpayerID, Category: c.Category, AmountCents: c.AmountCents,
		PaidOn: c.PaidOn, PaymentMethod: c.PaymentMethod, Beneficiary: c.Beneficiary, DoneeType: c.DoneeType,
		StudentLabel: c.StudentLabel, TuitionLevel: c.TuitionLevel, ScholarshipCents: c.ScholarshipCents, CFDIUUID: c.CFDIUUID,
	}
}

func (newDeductionRow) TableName() string { return deductionsTable }

// LockDeductionID implements annual.DeductionStore.
func (Store) LockDeductionID(ctx context.Context, id string) error {
	return advisoryLock(ctx, "lock deduction id", deductionLockClass, id)
}

// FindDeduction implements annual.DeductionStore.
func (Store) FindDeduction(ctx context.Context, id string, lock annual.Lock) (annual.StoredDeduction, error) {
	query := findDeductionQuery
	if lock == annual.LockWrite {
		query += lockNoKeyUpdt
	}
	var rows []deductionRow
	if err := raw(ctx, "find deduction", &rows, query, id); err != nil {
		return annual.StoredDeduction{}, err
	}
	if len(rows) == 0 {
		return annual.StoredDeduction{}, annual.ErrDeductionNotFound
	}
	return rows[0].toStored()
}

// InsertDeduction implements annual.DeductionStore.
func (Store) InsertDeduction(ctx context.Context, d annual.NewDeduction) error {
	row := newDeduction(d)
	return insert(ctx, "insert deduction", deductionsPrimaryKey, &row)
}

// UpdateDeduction implements annual.DeductionStore; the database
// re-derives fiscal_year from paid_on.
func (Store) UpdateDeduction(ctx context.Context, id string, f annual.DeductionFields) error {
	return updateDeduction(ctx, "update deduction", id, columnsOf(f).updates())
}

// SetDeductionDeleted implements annual.DeductionStore: deleted_at is the
// database's now().
func (Store) SetDeductionDeleted(ctx context.Context, id string, deleted bool) error {
	var deletedAt any
	if deleted {
		deletedAt = gorm.Expr("now()")
	}
	return updateDeduction(ctx, "set deleted_at", id, map[string]any{"deleted_at": deletedAt})
}

func updateDeduction(ctx context.Context, op, id string, columns map[string]any) error {
	db, err := pgdb.Conn(ctx)
	if err != nil {
		return err
	}
	result := db.Table(deductionsTable).Where("id = ?", id).Updates(columns)
	if result.Error != nil {
		return wrap(op, result.Error)
	}
	if result.RowsAffected == 0 {
		return annual.ErrDeductionNotFound
	}
	return nil
}

// CountLiveDeductions implements annual.DeductionStore.
func (Store) CountLiveDeductions(ctx context.Context, taxpayerID string, fiscalYear int) (int, error) {
	var n int
	err := raw(ctx, "count deductions", &n, countDeductionsSQL, taxpayerID, fiscalYear)
	return n, err
}

// ListDeductions implements annual.DeductionStore.
func (Store) ListDeductions(ctx context.Context, taxpayerID string, fiscalYear int, category adomain.Category) ([]annual.StoredDeduction, error) {
	query, args := listDeductionsQuery, []any{taxpayerID, fiscalYear}
	if category != "" {
		query, args = query+byCategory, append(args, string(category))
	}
	var rows []deductionRow
	if err := raw(ctx, "list deductions", &rows, query, args...); err != nil {
		return nil, err
	}
	out := make([]annual.StoredDeduction, 0, len(rows))
	for _, r := range rows {
		d, err := r.toStored()
		if err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, nil
}
