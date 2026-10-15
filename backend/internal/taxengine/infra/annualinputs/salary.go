package annualinputs

import (
	"context"
	"time"

	"github.com/romodevx/control-contable/backend/internal/platform/money"
	"github.com/romodevx/control-contable/backend/internal/platform/pgdb"
	"github.com/romodevx/control-contable/backend/internal/taxengine/usecase/annual"
)

const (
	statementsTable      = "salary_statements"
	statementsPrimaryKey = "salary_statements_pkey"
)

const statementSelect = `SELECT id, taxpayer_id, fiscal_year, employer_label, accumulable_cents, exempt_cents,
       isr_withheld_cents, simultaneous, left_before_year_end, payer_not_withholding, created_at, updated_at
FROM salary_statements`

// The read paths, served by salary_statements_year_idx (user_id from RLS,
// taxpayer_id, fiscal_year).
const (
	findStatementQuery  = statementSelect + " WHERE id = ?"
	listStatementsQuery = statementSelect + " WHERE taxpayer_id = ? AND fiscal_year = ? ORDER BY created_at, id"
	countStatementsSQL  = "SELECT count(*) FROM salary_statements WHERE taxpayer_id = ? AND fiscal_year = ?"
	deleteStatementSQL  = "DELETE FROM salary_statements WHERE id = ?"
)

type statementRow struct {
	ID                  string    `gorm:"column:id"`
	TaxpayerID          string    `gorm:"column:taxpayer_id"`
	FiscalYear          int       `gorm:"column:fiscal_year"`
	EmployerLabel       *string   `gorm:"column:employer_label"`
	AccumulableCents    int64     `gorm:"column:accumulable_cents"`
	ExemptCents         int64     `gorm:"column:exempt_cents"`
	ISRWithheldCents    int64     `gorm:"column:isr_withheld_cents"`
	Simultaneous        bool      `gorm:"column:simultaneous"`
	LeftBeforeYearEnd   bool      `gorm:"column:left_before_year_end"`
	PayerNotWithholding bool      `gorm:"column:payer_not_withholding"`
	CreatedAt           time.Time `gorm:"column:created_at"`
	UpdatedAt           time.Time `gorm:"column:updated_at"`
}

func (r statementRow) toStored() annual.StoredStatement {
	return annual.StoredStatement{
		ID: r.ID, TaxpayerID: r.TaxpayerID,
		SalaryFields: annual.SalaryFields{
			FiscalYear: r.FiscalYear, EmployerLabel: deref(r.EmployerLabel), Accumulable: money.Cents(r.AccumulableCents),
			Exempt: money.Cents(r.ExemptCents), ISRWithheld: money.Cents(r.ISRWithheldCents), Simultaneous: r.Simultaneous,
			LeftBeforeYearEnd: r.LeftBeforeYearEnd, PayerNotWithholding: r.PayerNotWithholding,
		},
		CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt,
	}
}

// newStatementRow is exactly the columns cc_app may INSERT.
type newStatementRow struct {
	ID                  string  `gorm:"column:id"`
	UserID              string  `gorm:"column:user_id"`
	TaxpayerID          string  `gorm:"column:taxpayer_id"`
	FiscalYear          int     `gorm:"column:fiscal_year"`
	EmployerLabel       *string `gorm:"column:employer_label"`
	AccumulableCents    int64   `gorm:"column:accumulable_cents"`
	ExemptCents         int64   `gorm:"column:exempt_cents"`
	ISRWithheldCents    int64   `gorm:"column:isr_withheld_cents"`
	Simultaneous        bool    `gorm:"column:simultaneous"`
	LeftBeforeYearEnd   bool    `gorm:"column:left_before_year_end"`
	PayerNotWithholding bool    `gorm:"column:payer_not_withholding"`
}

func (newStatementRow) TableName() string { return statementsTable }

// statementUpdates are the columns cc_app may UPDATE (fiscal_year since
// 20261017102000).
func statementUpdates(f annual.SalaryFields) map[string]any {
	return map[string]any{
		"fiscal_year": f.FiscalYear, "employer_label": nullable(f.EmployerLabel), "accumulable_cents": int64(f.Accumulable),
		"exempt_cents": int64(f.Exempt), "isr_withheld_cents": int64(f.ISRWithheld), "simultaneous": f.Simultaneous,
		"left_before_year_end": f.LeftBeforeYearEnd, "payer_not_withholding": f.PayerNotWithholding,
	}
}

// LockStatementID implements annual.SalaryStore.
func (Store) LockStatementID(ctx context.Context, id string) error {
	return advisoryLock(ctx, "lock statement id", statementLockClass, id)
}

// FindStatement implements annual.SalaryStore.
func (Store) FindStatement(ctx context.Context, id string, lock annual.Lock) (annual.StoredStatement, error) {
	query := findStatementQuery
	if lock == annual.LockWrite {
		query += lockNoKeyUpdt
	}
	var rows []statementRow
	if err := raw(ctx, "find statement", &rows, query, id); err != nil {
		return annual.StoredStatement{}, err
	}
	if len(rows) == 0 {
		return annual.StoredStatement{}, annual.ErrStatementNotFound
	}
	return rows[0].toStored(), nil
}

// InsertStatement implements annual.SalaryStore.
func (Store) InsertStatement(ctx context.Context, s annual.NewStatement) error {
	row := newStatementRow{
		ID: s.ID, UserID: s.UserID, TaxpayerID: s.TaxpayerID, FiscalYear: s.FiscalYear,
		EmployerLabel: nullable(s.EmployerLabel), AccumulableCents: int64(s.Accumulable), ExemptCents: int64(s.Exempt),
		ISRWithheldCents: int64(s.ISRWithheld), Simultaneous: s.Simultaneous, LeftBeforeYearEnd: s.LeftBeforeYearEnd,
		PayerNotWithholding: s.PayerNotWithholding,
	}
	return insert(ctx, "insert statement", statementsPrimaryKey, &row)
}

// UpdateStatement implements annual.SalaryStore.
func (Store) UpdateStatement(ctx context.Context, id string, f annual.SalaryFields) error {
	db, err := pgdb.Conn(ctx)
	if err != nil {
		return err
	}
	result := db.Table(statementsTable).Where("id = ?", id).Updates(statementUpdates(f))
	if result.Error != nil {
		return wrap("update statement", result.Error)
	}
	if result.RowsAffected == 0 {
		return annual.ErrStatementNotFound
	}
	return nil
}

// DeleteStatement implements annual.SalaryStore.
func (Store) DeleteStatement(ctx context.Context, id string) error {
	n, err := exec(ctx, "delete statement", deleteStatementSQL, id)
	if err == nil && n == 0 {
		return annual.ErrStatementNotFound
	}
	return err
}

// CountStatements implements annual.SalaryStore.
func (Store) CountStatements(ctx context.Context, taxpayerID string, fiscalYear int) (int, error) {
	var n int
	err := raw(ctx, "count statements", &n, countStatementsSQL, taxpayerID, fiscalYear)
	return n, err
}

// ListStatements implements annual.SalaryStore.
func (Store) ListStatements(ctx context.Context, taxpayerID string, fiscalYear int) ([]annual.StoredStatement, error) {
	var rows []statementRow
	if err := raw(ctx, "list statements", &rows, listStatementsQuery, taxpayerID, fiscalYear); err != nil {
		return nil, err
	}
	out := make([]annual.StoredStatement, 0, len(rows))
	for _, r := range rows {
		out = append(out, r.toStored())
	}
	return out, nil
}
