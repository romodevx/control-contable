package infra

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/romodevx/control-contable/backend/internal/expenses/domain"
	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
	"github.com/romodevx/control-contable/backend/internal/platform/pgdb"
)

// ErrInvalidLedgerQuery reports a LedgerQuery the reader cannot run: a
// malformed taxpayer id, a month outside 1..12 or a year before the
// paid_on floor. The tax engine builds the query, so it is a caller bug.
var ErrInvalidLedgerQuery = errors.New("expenses infra: invalid ledger query")

// firstFiscalYear is the year of the paid_on floor (expenses_paid_on_floor).
const firstFiscalYear = 2022

// LedgerQuery selects one taxpayer's expenses paid from January of
// FirstYear through Through of FiscalYear. A zero FirstYear is FiscalYear:
// one fiscal year, January through Through.
type LedgerQuery struct {
	TaxpayerID string
	// FirstYear opens the range on its January 1st (M6-02: the dashboard's
	// window spans up to two fiscal years in one statement).
	FirstYear  int
	FiscalYear int
	Through    time.Month
}

// RateSums are the base (amount_cents) and the IVA (iva_cents) paid at one
// taxable rate.
type RateSums struct {
	Base money.Cents
	IVA  money.Cents
}

// IVAByRate splits the expenses by the record's iva_rate; a 0 % or exempt
// record carries no IVA (expenses_untaxed_has_no_iva), so only its base.
type IVAByRate struct {
	Rate16     RateSums
	Rate8      RateSums
	Rate0Base  money.Cents
	ExemptBase money.Cents
}

// ExpenseMonthSums are one (regime, month)'s live expenses, by paid_on, in
// centavos. The composition root maps them to the tax engine's
// ExpenseMonthSums (ADR-0001: modules meet only through consumer ports).
//
// The deductibility classes partition amount_cents: DeductiblePaid is the
// stored effective verdict (an override included); a non-deductible
// expense counts in CashOverLimitPaid when it was paid in cash over the
// limit, else in NonDeductiblePaid. CreditableIVAPaid sums the stored
// creditable-IVA candidate; PredialPaid the deductible predial; the
// withholdings made to suppliers count whatever the deductibility.
type ExpenseMonthSums struct {
	Regime             fiscal.RegimeCode
	Year               int
	Month              time.Month
	DeductiblePaid     money.Cents
	NonDeductiblePaid  money.Cents
	CashOverLimitPaid  money.Cents
	IVAPaid            IVAByRate
	CreditableIVAPaid  money.Cents
	PredialPaid        money.Cents
	ISRWithheldToRemit money.Cents
	IVAWithheldToRemit money.Cents
}

// LedgerReader reads the tax engine's expense aggregates with GORM on the
// unit of work in ctx (pgdb.Conn), as cc_app: RLS limits it to the context
// user's rows. It reads whatever module_expenses' flag says (MODULE-incomes
// OQ-8).
type LedgerReader struct{}

// NewLedgerReader returns a LedgerReader. It holds no pool.
func NewLedgerReader() LedgerReader { return LedgerReader{} }

// expenseSumsSQL is the single round trip of ExpenseSums: one GROUP BY
// regime, month over expenses_live_paid_on_idx (user_id from the RLS
// policy, taxpayer_id, then the paid_on range; the partial predicate drops
// the soft-deleted rows). Each sum is cast back to bigint, so an overflow
// fails (22003) instead of wrapping. The dates use CAST(... AS date): GORM
// leaves a named parameter followed by "::" unbound.
const expenseSumsSQL = `
SELECT regime_code,
       extract(year FROM paid_on)::int                                                               AS year,
       extract(month FROM paid_on)::int                                                              AS month,
       coalesce(sum(amount_cents) FILTER (WHERE deductible), 0)::bigint                              AS deductible_paid,
       coalesce(sum(amount_cents) FILTER (WHERE NOT deductible AND NOT cash_over_limit), 0)::bigint  AS non_deductible_paid,
       coalesce(sum(amount_cents) FILTER (WHERE NOT deductible AND cash_over_limit), 0)::bigint      AS cash_over_limit_paid,
       coalesce(sum(amount_cents) FILTER (WHERE iva_rate = 'rate_16'), 0)::bigint                    AS rate16_base,
       coalesce(sum(iva_cents) FILTER (WHERE iva_rate = 'rate_16'), 0)::bigint                       AS rate16_iva,
       coalesce(sum(amount_cents) FILTER (WHERE iva_rate = 'rate_8'), 0)::bigint                     AS rate8_base,
       coalesce(sum(iva_cents) FILTER (WHERE iva_rate = 'rate_8'), 0)::bigint                        AS rate8_iva,
       coalesce(sum(amount_cents) FILTER (WHERE iva_rate = 'rate_0'), 0)::bigint                     AS rate0_base,
       coalesce(sum(amount_cents) FILTER (WHERE iva_rate = 'exempt'), 0)::bigint                     AS exempt_base,
       coalesce(sum(creditable_iva_cents), 0)::bigint                                                AS creditable_iva_paid,
       coalesce(sum(amount_cents) FILTER (WHERE category_code = @predial AND deductible), 0)::bigint AS predial_paid,
       coalesce(sum(isr_withheld_cents), 0)::bigint                                                  AS isr_withheld,
       coalesce(sum(iva_withheld_cents), 0)::bigint                                                  AS iva_withheld
FROM expenses
WHERE taxpayer_id = @taxpayer AND deleted_at IS NULL AND paid_on >= CAST(@from_date AS date) AND paid_on < CAST(@until_date AS date)
GROUP BY regime_code, year, month
ORDER BY regime_code, year, month`

// sumsRow is one row of expenseSumsSQL.
type sumsRow struct {
	RegimeCode        string `gorm:"column:regime_code"`
	Year              int    `gorm:"column:year"`
	Month             int    `gorm:"column:month"`
	DeductiblePaid    int64  `gorm:"column:deductible_paid"`
	NonDeductiblePaid int64  `gorm:"column:non_deductible_paid"`
	CashOverLimitPaid int64  `gorm:"column:cash_over_limit_paid"`
	Rate16Base        int64  `gorm:"column:rate16_base"`
	Rate16IVA         int64  `gorm:"column:rate16_iva"`
	Rate8Base         int64  `gorm:"column:rate8_base"`
	Rate8IVA          int64  `gorm:"column:rate8_iva"`
	Rate0Base         int64  `gorm:"column:rate0_base"`
	ExemptBase        int64  `gorm:"column:exempt_base"`
	CreditableIVAPaid int64  `gorm:"column:creditable_iva_paid"`
	PredialPaid       int64  `gorm:"column:predial_paid"`
	ISRWithheld       int64  `gorm:"column:isr_withheld"`
	IVAWithheld       int64  `gorm:"column:iva_withheld"`
}

func (r sumsRow) toSums() ExpenseMonthSums {
	return ExpenseMonthSums{
		Regime: fiscal.RegimeCode(r.RegimeCode), Year: r.Year, Month: time.Month(r.Month),
		DeductiblePaid:    money.Cents(r.DeductiblePaid),
		NonDeductiblePaid: money.Cents(r.NonDeductiblePaid),
		CashOverLimitPaid: money.Cents(r.CashOverLimitPaid),
		IVAPaid: IVAByRate{
			Rate16:     RateSums{Base: money.Cents(r.Rate16Base), IVA: money.Cents(r.Rate16IVA)},
			Rate8:      RateSums{Base: money.Cents(r.Rate8Base), IVA: money.Cents(r.Rate8IVA)},
			Rate0Base:  money.Cents(r.Rate0Base),
			ExemptBase: money.Cents(r.ExemptBase),
		},
		CreditableIVAPaid:  money.Cents(r.CreditableIVAPaid),
		PredialPaid:        money.Cents(r.PredialPaid),
		ISRWithheldToRemit: money.Cents(r.ISRWithheld),
		IVAWithheldToRemit: money.Cents(r.IVAWithheld),
	}
}

// ExpenseSums returns one row per (regime, month) with live expenses in
// q's range, ordered by regime then month; months without records are
// absent.
func (LedgerReader) ExpenseSums(ctx context.Context, q LedgerQuery) ([]ExpenseMonthSums, error) {
	statement, err := expenseSums(ctx, q, "")
	if err != nil {
		return nil, err
	}
	var rows []sumsRow
	if err := statement.Scan(&rows).Error; err != nil {
		return nil, translate("sum expenses", err)
	}
	out := make([]ExpenseMonthSums, 0, len(rows))
	for _, r := range rows {
		out = append(out, r.toSums())
	}
	return out, nil
}

// hasLiveSQL is HasLive's statement: an EXISTS the planner answers from
// expenses_live_paid_on_idx (user_id from the RLS policy, then
// taxpayer_id; its partial predicate is the live rows').
const hasLiveSQL = `SELECT EXISTS (SELECT 1 FROM expenses WHERE taxpayer_id = ? AND deleted_at IS NULL)`

// HasLive reports whether the taxpayer has any live expense, in any period
// (the dashboard's is_empty, M6-02). It runs on the unit of work in ctx,
// as its user (RLS): another user's taxpayer has none.
func (LedgerReader) HasLive(ctx context.Context, taxpayerID string) (bool, error) {
	db, err := pgdb.Conn(ctx)
	if err != nil {
		return false, err
	}
	var exists bool
	if err := db.Raw(hasLiveSQL, taxpayerID).Scan(&exists).Error; err != nil {
		return false, translate("live expenses exist", err)
	}
	return exists, nil
}

// expenseSums builds the statement for q on the unit of work in ctx,
// prefixed (EXPLAIN in the tests).
func expenseSums(ctx context.Context, q LedgerQuery, prefix string) (*gorm.DB, error) {
	from, until, err := q.dates()
	if err != nil {
		return nil, err
	}
	db, err := pgdb.Conn(ctx)
	if err != nil {
		return nil, err
	}
	return db.Raw(prefix+expenseSumsSQL, map[string]any{
		"taxpayer": q.TaxpayerID, "from_date": from, "until_date": until, "predial": string(domain.CategoryPredial),
	}), nil
}

// dates returns the paid_on range [from, until) of q.
func (q LedgerQuery) dates() (from, until string, err error) {
	if _, parseErr := uuid.Parse(q.TaxpayerID); parseErr != nil {
		return "", "", fmt.Errorf("%w: taxpayer id: %w", ErrInvalidLedgerQuery, parseErr)
	}
	first := q.FirstYear
	if first == 0 {
		first = q.FiscalYear
	}
	if q.Through < time.January || q.Through > time.December || first < firstFiscalYear || first > q.FiscalYear {
		return "", "", fmt.Errorf("%w: %d..%d through month %d", ErrInvalidLedgerQuery, first, q.FiscalYear, q.Through)
	}
	start := time.Date(first, time.January, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(q.FiscalYear, q.Through+1, 1, 0, 0, 0, 0, time.UTC) // December rolls into January 1st
	return start.Format(dateLayout), end.Format(dateLayout), nil
}

// explainExpenseSums returns the EXPLAIN (ANALYZE, BUFFERS) text of q's
// statement (export_test.go: the PR's plan evidence).
func explainExpenseSums(ctx context.Context, q LedgerQuery) (string, error) {
	statement, err := expenseSums(ctx, q, "EXPLAIN (ANALYZE, BUFFERS) ")
	if err != nil {
		return "", err
	}
	var lines []string
	if err := statement.Scan(&lines).Error; err != nil {
		return "", translate("explain expense sums", err)
	}
	return strings.Join(lines, "\n"), nil
}
