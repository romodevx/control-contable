package infra

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
	"github.com/romodevx/control-contable/backend/internal/platform/pgdb"
)

// ErrInvalidLedgerQuery reports a LedgerQuery the reader cannot run: a
// malformed taxpayer id, a month outside 1..12 or a year before the
// incomes' date floor. The tax engine builds the query, so it is a caller
// bug.
var ErrInvalidLedgerQuery = errors.New("incomes infra: invalid ledger query")

// firstFiscalYear is the year of the incomes' date floor
// (incomes_invoiced_on_floor, incomes_collected_on_floor).
const firstFiscalYear = 2022

// LedgerQuery selects one taxpayer's incomes from January of FirstYear
// through Through of FiscalYear. A zero FirstYear is FiscalYear: one
// fiscal year, January through Through.
type LedgerQuery struct {
	TaxpayerID string
	// FirstYear opens the range on its January 1st (M6-02: the dashboard's
	// window spans up to two fiscal years in one statement).
	FirstYear  int
	FiscalYear int
	Through    time.Month
}

// RateSums are the base (amount_cents) and the IVA (iva_cents) collected at
// one taxable rate.
type RateSums struct {
	Base money.Cents
	IVA  money.Cents
}

// IVAByRate splits the incomes collected by the record's iva_rate; a 0 % or
// exempt income carries no IVA (incomes_zero_rate_has_no_iva), so only its
// base.
type IVAByRate struct {
	Rate16     RateSums
	Rate8      RateSums
	Rate0Base  money.Cents
	ExemptBase money.Cents
}

// Withheld splits a withholding by the counterparty that made it. Only a
// persona moral or a platform withholds (incomes_no_withholding_by_
// individuals), and the kind follows the row's regime: a moral withholds
// 10 % ISR from 612 and 606, 1.25 % from 626; a platform withholds from
// 625 by activity, and from a persona moral.
type Withheld struct {
	ByMoral    money.Cents
	ByPlatform money.Cents
}

// IncomeMonthSums are one (regime, month)'s live incomes, in centavos. The
// composition root maps them to the tax engine's IncomeMonthSums (ADR-0001:
// modules meet only through consumer ports).
//
// Accrual is the ISR income of an accrual-basis regime (regimes.income_
// basis), by booked_on, which is its invoiced_on; it is zero for a cash
// regime, whose ISR income is Cash. Cash, ByRate and both withholdings are
// by collected_on, for every regime: IVA is always cash (LIVA 1-B), and an
// income still pending collection counts in Accrual only.
type IncomeMonthSums struct {
	Regime      fiscal.RegimeCode
	Year        int
	Month       time.Month
	Cash        money.Cents
	Accrual     money.Cents
	ByRate      IVAByRate
	IVAWithheld Withheld
	ISRWithheld Withheld
}

// LedgerReader reads the tax engine's income aggregates with GORM on the
// unit of work in ctx (pgdb.Conn), as cc_app: RLS limits it to the context
// user's rows. It reads whatever module_incomes' flag says (MODULE-incomes
// OQ-8).
type LedgerReader struct{}

// NewLedgerReader returns a LedgerReader. It holds no pool.
func NewLedgerReader() LedgerReader { return LedgerReader{} }

// The withholder of an income, from its derived facts (WH-06):
// counterparty_kind and via_platform, the same facts domain.LegacyCounterparty
// mirrors into the deprecated counterparty_type. A platform sale is the
// platform's; otherwise a persona moral or a government withholds as a moral.
// A legacy income the WH-08 backfill has not derived yet (counterparty_kind
// NULL) still answers by its stored counterparty_type, the backfill's own
// §5.2 mapping (ADR-0008 Amendment 2, item 3: no deploy order). That
// fallback, the ledger's last read of counterparty_type, goes with the
// column's drop once every environment is backfilled (PLAN WH-08 contract
// step).
const (
	byPlatform = `via_platform OR (counterparty_kind IS NULL AND counterparty_type = 'plataforma')`
	byMoral    = `NOT via_platform AND coalesce(counterparty_kind IN ('pm', 'gov_federal', 'gov_state_municipal'),
                                         counterparty_type = 'persona_moral')`
)

// incomeSumsSQL is the single round trip of IncomeSums. Each CTE is one
// GROUP BY regime, month over its partial index (user_id from the RLS
// policy, taxpayer_id, then the date range; the predicates drop the
// soft-deleted rows): booked over incomes_booked_idx, collected over
// incomes_collected_idx. They meet per (regime, month), and the regime's
// catalog basis (through the taxpayer_regimes row the composite FK
// guarantees) decides whether booked is accrual income: a cash regime's
// booked_on is its collected_on, already in collected. Each sum is cast back
// to bigint, so an overflow fails (22003) instead of wrapping. The dates use
// CAST(... AS date): GORM leaves a named parameter followed by "::" unbound.
const incomeSumsSQL = `
WITH booked AS (
  SELECT regime_code, extract(year FROM booked_on)::int AS year, extract(month FROM booked_on)::int AS month,
         sum(amount_cents) AS amount
  FROM incomes
  WHERE taxpayer_id = @taxpayer AND deleted_at IS NULL
    AND booked_on >= CAST(@from_date AS date) AND booked_on < CAST(@until_date AS date)
  GROUP BY regime_code, year, month
), collected AS (
  SELECT regime_code,
         extract(year FROM collected_on)::int                                                           AS year,
         extract(month FROM collected_on)::int                                                          AS month,
         sum(amount_cents)                                                                              AS amount,
         coalesce(sum(amount_cents) FILTER (WHERE iva_rate = 'rate_16'), 0)                             AS rate16_base,
         coalesce(sum(iva_cents) FILTER (WHERE iva_rate = 'rate_16'), 0)                                AS rate16_iva,
         coalesce(sum(amount_cents) FILTER (WHERE iva_rate = 'rate_8'), 0)                              AS rate8_base,
         coalesce(sum(iva_cents) FILTER (WHERE iva_rate = 'rate_8'), 0)                                 AS rate8_iva,
         coalesce(sum(amount_cents) FILTER (WHERE iva_rate = 'rate_0'), 0)                              AS rate0_base,
         coalesce(sum(amount_cents) FILTER (WHERE iva_rate = 'exempt'), 0)                              AS exempt_base,
         coalesce(sum(iva_withheld_cents) FILTER (WHERE ` + byMoral + `), 0)                         AS iva_withheld_moral,
         coalesce(sum(iva_withheld_cents) FILTER (WHERE ` + byPlatform + `), 0)                      AS iva_withheld_platform,
         coalesce(sum(isr_withheld_cents) FILTER (WHERE ` + byMoral + `), 0)                         AS isr_withheld_moral,
         coalesce(sum(isr_withheld_cents) FILTER (WHERE ` + byPlatform + `), 0)                      AS isr_withheld_platform
  FROM incomes
  WHERE taxpayer_id = @taxpayer AND deleted_at IS NULL AND collected_on IS NOT NULL
    AND collected_on >= CAST(@from_date AS date) AND collected_on < CAST(@until_date AS date)
  GROUP BY regime_code, year, month
)
SELECT tr.regime_code,
       coalesce(b.year, c.year)                                                          AS year,
       coalesce(b.month, c.month)                                                        AS month,
       coalesce(c.amount, 0)::bigint                                                     AS cash,
       (CASE r.income_basis WHEN 'accrual' THEN coalesce(b.amount, 0) ELSE 0 END)::bigint AS accrual,
       coalesce(c.rate16_base, 0)::bigint                                                AS rate16_base,
       coalesce(c.rate16_iva, 0)::bigint                                                 AS rate16_iva,
       coalesce(c.rate8_base, 0)::bigint                                                 AS rate8_base,
       coalesce(c.rate8_iva, 0)::bigint                                                  AS rate8_iva,
       coalesce(c.rate0_base, 0)::bigint                                                 AS rate0_base,
       coalesce(c.exempt_base, 0)::bigint                                                AS exempt_base,
       coalesce(c.iva_withheld_moral, 0)::bigint                                         AS iva_withheld_moral,
       coalesce(c.iva_withheld_platform, 0)::bigint                                      AS iva_withheld_platform,
       coalesce(c.isr_withheld_moral, 0)::bigint                                         AS isr_withheld_moral,
       coalesce(c.isr_withheld_platform, 0)::bigint                                      AS isr_withheld_platform
FROM booked b
FULL JOIN collected c ON c.regime_code = b.regime_code AND c.year = b.year AND c.month = b.month
JOIN taxpayer_regimes tr ON tr.taxpayer_id = @taxpayer AND tr.regime_code = coalesce(b.regime_code, c.regime_code)
JOIN regimes r ON r.code = tr.regime_code AND r.person_type = tr.person_type
WHERE c.month IS NOT NULL OR r.income_basis = 'accrual'
ORDER BY tr.regime_code, year, month`

// sumsRow is one row of incomeSumsSQL.
type sumsRow struct {
	RegimeCode          string `gorm:"column:regime_code"`
	Year                int    `gorm:"column:year"`
	Month               int    `gorm:"column:month"`
	Cash                int64  `gorm:"column:cash"`
	Accrual             int64  `gorm:"column:accrual"`
	Rate16Base          int64  `gorm:"column:rate16_base"`
	Rate16IVA           int64  `gorm:"column:rate16_iva"`
	Rate8Base           int64  `gorm:"column:rate8_base"`
	Rate8IVA            int64  `gorm:"column:rate8_iva"`
	Rate0Base           int64  `gorm:"column:rate0_base"`
	ExemptBase          int64  `gorm:"column:exempt_base"`
	IVAWithheldMoral    int64  `gorm:"column:iva_withheld_moral"`
	IVAWithheldPlatform int64  `gorm:"column:iva_withheld_platform"`
	ISRWithheldMoral    int64  `gorm:"column:isr_withheld_moral"`
	ISRWithheldPlatform int64  `gorm:"column:isr_withheld_platform"`
}

func (r sumsRow) toSums() IncomeMonthSums {
	return IncomeMonthSums{
		Regime: fiscal.RegimeCode(r.RegimeCode), Year: r.Year, Month: time.Month(r.Month),
		Cash:    money.Cents(r.Cash),
		Accrual: money.Cents(r.Accrual),
		ByRate: IVAByRate{
			Rate16:     RateSums{Base: money.Cents(r.Rate16Base), IVA: money.Cents(r.Rate16IVA)},
			Rate8:      RateSums{Base: money.Cents(r.Rate8Base), IVA: money.Cents(r.Rate8IVA)},
			Rate0Base:  money.Cents(r.Rate0Base),
			ExemptBase: money.Cents(r.ExemptBase),
		},
		IVAWithheld: Withheld{ByMoral: money.Cents(r.IVAWithheldMoral), ByPlatform: money.Cents(r.IVAWithheldPlatform)},
		ISRWithheld: Withheld{ByMoral: money.Cents(r.ISRWithheldMoral), ByPlatform: money.Cents(r.ISRWithheldPlatform)},
	}
}

// IncomeSums returns one row per (regime, month) with live incomes in q's
// range, ordered by regime then month; months without records are absent.
func (LedgerReader) IncomeSums(ctx context.Context, q LedgerQuery) ([]IncomeMonthSums, error) {
	statement, err := incomeSums(ctx, q, "")
	if err != nil {
		return nil, err
	}
	var rows []sumsRow
	if err := statement.Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("incomes infra: sum incomes: %w", err)
	}
	out := make([]IncomeMonthSums, 0, len(rows))
	for _, r := range rows {
		out = append(out, r.toSums())
	}
	return out, nil
}

// hasLiveSQL is HasLive's statement: an EXISTS the planner answers from
// incomes_booked_idx (user_id from the RLS policy, then taxpayer_id; its
// partial predicate is the live rows', and booked_on is never NULL).
const hasLiveSQL = `SELECT EXISTS (SELECT 1 FROM incomes WHERE taxpayer_id = ? AND deleted_at IS NULL)`

// HasLive reports whether the taxpayer has any live income, in any period
// (the dashboard's is_empty, M6-02). It runs on the unit of work in ctx,
// as its user (RLS): another user's taxpayer has none.
func (LedgerReader) HasLive(ctx context.Context, taxpayerID string) (bool, error) {
	db, err := pgdb.Conn(ctx)
	if err != nil {
		return false, err
	}
	var exists bool
	if err := db.Raw(hasLiveSQL, taxpayerID).Scan(&exists).Error; err != nil {
		return false, fmt.Errorf("incomes infra: live incomes exist: %w", err)
	}
	return exists, nil
}

// incomeSums builds the statement for q on the unit of work in ctx,
// prefixed (EXPLAIN in the tests).
func incomeSums(ctx context.Context, q LedgerQuery, prefix string) (*gorm.DB, error) {
	from, until, err := q.dates()
	if err != nil {
		return nil, err
	}
	db, err := pgdb.Conn(ctx)
	if err != nil {
		return nil, err
	}
	return db.Raw(prefix+incomeSumsSQL, map[string]any{
		"taxpayer": q.TaxpayerID, "from_date": from, "until_date": until,
	}), nil
}

// dates returns the date range [from, until) of q.
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
	return start.Format(time.DateOnly), end.Format(time.DateOnly), nil
}

// explainIncomeSums returns the EXPLAIN (ANALYZE, BUFFERS) text of q's
// statement (export_test.go: the PR's plan evidence).
func explainIncomeSums(ctx context.Context, q LedgerQuery) (string, error) {
	statement, err := incomeSums(ctx, q, "EXPLAIN (ANALYZE, BUFFERS) ")
	if err != nil {
		return "", err
	}
	var lines []string
	if err := statement.Scan(&lines).Error; err != nil {
		return "", fmt.Errorf("incomes infra: explain income sums: %w", err)
	}
	return strings.Join(lines, "\n"), nil
}
