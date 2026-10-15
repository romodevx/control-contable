package annualinputs

import (
	"context"
	"time"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
	"github.com/romodevx/control-contable/backend/internal/platform/pgdb"
	"github.com/romodevx/control-contable/backend/internal/taxengine/usecase/annual"
)

const filedTable = "annual_filed_payments"

// The read and replace paths, all served by annual_filed_payments_pkey
// (user_id from RLS, taxpayer_id, regime_code, fiscal_year, month).
const (
	filedMonthsQuery = `SELECT month, amount_cents, updated_at FROM annual_filed_payments
WHERE taxpayer_id = ? AND regime_code = ? AND fiscal_year = ? ORDER BY month`
	filedYearQuery = `SELECT regime_code, month, amount_cents, updated_at FROM annual_filed_payments
WHERE taxpayer_id = ? AND fiscal_year = ?`
	deleteFiledSQL = "DELETE FROM annual_filed_payments WHERE taxpayer_id = ? AND regime_code = ? AND fiscal_year = ?"
)

type filedRow struct {
	RegimeCode  string    `gorm:"column:regime_code"`
	Month       int       `gorm:"column:month"`
	AmountCents int64     `gorm:"column:amount_cents"`
	UpdatedAt   time.Time `gorm:"column:updated_at"`
}

// newFiledRow is exactly the columns cc_app may INSERT.
type newFiledRow struct {
	UserID      string `gorm:"column:user_id"`
	TaxpayerID  string `gorm:"column:taxpayer_id"`
	RegimeCode  string `gorm:"column:regime_code"`
	FiscalYear  int    `gorm:"column:fiscal_year"`
	Month       int    `gorm:"column:month"`
	AmountCents int64  `gorm:"column:amount_cents"`
}

func (newFiledRow) TableName() string { return filedTable }

// FiledMonths implements annual.FiledStore.
func (Store) FiledMonths(ctx context.Context, key annual.FiledKey) ([]annual.FiledMonth, *time.Time, error) {
	var rows []filedRow
	if err := raw(ctx, "read filed payments", &rows, filedMonthsQuery, key.TaxpayerID, string(key.Regime), key.FiscalYear); err != nil {
		return nil, nil, err
	}
	months := make([]annual.FiledMonth, 0, len(rows))
	stamps := make([]time.Time, 0, len(rows))
	for _, r := range rows {
		months = append(months, annual.FiledMonth{Month: time.Month(r.Month), Amount: money.Cents(r.AmountCents)})
		stamps = append(stamps, r.UpdatedAt)
	}
	return months, latest(stamps), nil
}

// ReplaceFiled implements annual.FiledStore: one DELETE of the set, then
// one multi-row INSERT. The caller holds the regime row lock, so no
// concurrent replace of the same set can interleave.
func (Store) ReplaceFiled(ctx context.Context, userID string, key annual.FiledKey, months []annual.FiledMonth) error {
	if _, err := exec(ctx, "clear filed payments", deleteFiledSQL, key.TaxpayerID, string(key.Regime), key.FiscalYear); err != nil {
		return err
	}
	if len(months) == 0 {
		return nil
	}
	rows := make([]newFiledRow, 0, len(months))
	for _, m := range months {
		rows = append(rows, newFiledRow{
			UserID: userID, TaxpayerID: key.TaxpayerID, RegimeCode: string(key.Regime), FiscalYear: key.FiscalYear,
			Month: int(m.Month), AmountCents: int64(m.Amount),
		})
	}
	db, err := pgdb.Conn(ctx)
	if err != nil {
		return err
	}
	if err := db.Create(&rows).Error; err != nil {
		return wrap("store filed payments", err)
	}
	return nil
}

// FiledPayments implements annual.FiledPaymentsReader: every confirmed
// month of the taxpayer's year, by regime, in one query.
func (Store) FiledPayments(ctx context.Context, taxpayerID string, fiscalYear int) (map[fiscal.RegimeCode]map[time.Month]money.Cents, error) {
	var rows []filedRow
	if err := raw(ctx, "read filed payments", &rows, filedYearQuery, taxpayerID, fiscalYear); err != nil {
		return nil, err
	}
	out := map[fiscal.RegimeCode]map[time.Month]money.Cents{}
	for _, r := range rows {
		code := fiscal.RegimeCode(r.RegimeCode)
		if out[code] == nil {
			out[code] = map[time.Month]money.Cents{}
		}
		out[code][time.Month(r.Month)] = money.Cents(r.AmountCents)
	}
	return out, nil
}

// YearInputs implements annual.YearInputsReader: the live deductions, the
// constancias and the facts (defaults when never saved) of the year.
func (s Store) YearInputs(ctx context.Context, taxpayerID string, fiscalYear int) (annual.YearInputs, error) {
	deductions, err := s.ListDeductions(ctx, taxpayerID, fiscalYear, "")
	if err != nil {
		return annual.YearInputs{}, err
	}
	statements, err := s.ListStatements(ctx, taxpayerID, fiscalYear)
	if err != nil {
		return annual.YearInputs{}, err
	}
	facts, _, err := s.Facts(ctx, taxpayerID, fiscalYear)
	if err != nil {
		return annual.YearInputs{}, err
	}
	return annual.YearInputsOf(deductions, statements, facts.Facts), nil
}
