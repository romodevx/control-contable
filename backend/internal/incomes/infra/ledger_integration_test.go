//go:build integration

package infra_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/incomes/domain"
	"github.com/romodevx/control-contable/backend/internal/incomes/infra"
	"github.com/romodevx/control-contable/backend/internal/incomes/usecase"
	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
	"github.com/romodevx/control-contable/backend/internal/platform/pgdb"
	"github.com/romodevx/control-contable/backend/internal/platform/pgtest"
)

// The tax engine's income aggregate (M3-07, MODULE-tax-engine §2.4): one statement per
// taxpayer and month range over the live incomes, by regime and month: the ISR income by
// booked_on (the regime's basis) and the IVA and the withholdings by collected_on (IVA is
// always cash, LIVA 1-B), as cc_app under the context user's RLS.

// The ledger fixtures' taxpayers, inserted directly (the composite FK and the booked_on
// trigger still check them). Ids are this file's own.
const (
	ledgerFisica = "0199d4a2-1111-7000-8000-00000004f0a1" // userA, física: 612 + 625 (cash)
	ledgerRESICO = "0199d4a2-1111-7000-8000-00000004f0a2" // userA, física: 626 (cash)
	ledgerMoral  = "0199d4a2-1111-7000-8000-00000004f0a3" // userA, moral: 601 (accrual)
	ledgerOther  = "0199d4a2-1111-7000-8000-00000004f0b1" // userB, física: 612
)

const insertLedgerTaxpayerSQL = `
INSERT INTO taxpayers (id, user_id, person_type, display_name, rfc_encrypted, rfc_key_id, rfc_hash, rfc_masked,
                       rfc_check_digit_valid)
VALUES ($2::text::uuid, $1, $3, 'Ana (ingresos)',
        '\x0001'::bytea || sha256(convert_to($2::text, 'UTF8')) || sha256(convert_to($2::text, 'UTF8')),
        1, sha256(convert_to($2::text || '/rfc', 'UTF8')),
        CASE $3 WHEN 'fisica' THEN 'GOMA******AB6' ELSE 'GOM******AB6' END, true)`

const insertLedgerRegimeSQL = `
INSERT INTO taxpayer_regimes (user_id, taxpayer_id, regime_code, person_type, options, start_period)
VALUES ($1, $2, $3, $4, '{}'::jsonb, '2022-01')`

// seedLedgerTaxpayer inserts taxpayer id of user holding regimes since January 2022.
func seedLedgerTaxpayer(t *testing.T, app *sql.DB, user, id, personType string, regimes ...string) {
	t.Helper()
	execAs(t, app, user, insertLedgerTaxpayerSQL, user, id, personType)
	for _, regime := range regimes {
		execAs(t, app, user, insertLedgerRegimeSQL, user, id, regime, personType)
	}
}

func seedLedgerTaxpayers(t *testing.T, e *env) {
	t.Helper()
	seedLedgerTaxpayer(t, e.app, userA, ledgerFisica, "fisica", "612", "625")
	seedLedgerTaxpayer(t, e.app, userA, ledgerRESICO, "fisica", "626")
	seedLedgerTaxpayer(t, e.app, userA, ledgerMoral, "moral", "601")
	seedLedgerTaxpayer(t, e.app, userB, ledgerOther, "fisica", "612")
}

// incomeFixture is one income row as the fixtures insert it; an empty date is NULL.
type incomeFixture struct {
	user, taxpayer, regime, counterparty, rate string
	amount, iva, isrWithheld, ivaWithheld      money.Cents
	invoicedOn, collectedOn                    string
	deleted                                    bool
}

const insertIncomeFixtureSQL = `
INSERT INTO incomes (user_id, taxpayer_id, regime_code, counterparty_type, amount_cents, iva_rate, iva_cents,
                     isr_withheld_cents, iva_withheld_cents, invoiced_on, collected_on)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10::date, $11::date)
RETURNING id`

func seedIncomes(t *testing.T, e *env, fixtures ...incomeFixture) {
	t.Helper()
	for _, f := range fixtures {
		var id string
		queryAs(t, e.app, f.user, insertIncomeFixtureSQL, []any{
			f.user, f.taxpayer, f.regime, f.counterparty, int64(f.amount), f.rate, int64(f.iva),
			int64(f.isrWithheld), int64(f.ivaWithheld), nullDate(f.invoicedOn), nullDate(f.collectedOn),
		}, &id)
		if f.deleted {
			execAs(t, e.app, f.user, "UPDATE incomes SET deleted_at = now() WHERE id = $1", id)
		}
	}
}

func nullDate(raw string) any {
	if raw == "" {
		return nil
	}
	return raw
}

func execAs(t *testing.T, app *sql.DB, userID, statement string, args ...any) {
	t.Helper()
	require.NoError(t, pgtest.AsUser(t.Context(), app, userID, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(t.Context(), statement, args...)
		return err
	}))
}

// queryAs runs one query under userID's RLS context and scans one row.
func queryAs(t *testing.T, app *sql.DB, userID, query string, args []any, dest ...any) {
	t.Helper()
	require.NoError(t, pgtest.AsUser(t.Context(), app, userID, func(tx *sql.Tx) error {
		return tx.QueryRowContext(t.Context(), query, args...).Scan(dest...)
	}))
}

// sale is a 16 % sale to the público en general by ledgerFisica's 612, invoiced and
// collected on the same day; the cases change what they test.
func sale(on string, amount money.Cents) incomeFixture {
	return incomeFixture{
		user: userA, taxpayer: ledgerFisica, regime: "612", counterparty: "publico_general",
		rate: "rate_16", amount: amount, iva: amount * 16 / 100, invoicedOn: on, collectedOn: on,
	}
}

func withIncome(f incomeFixture, change func(*incomeFixture)) incomeFixture {
	change(&f)
	return f
}

func readIncomeSums(t *testing.T, e *env, userID string, q infra.LedgerQuery) []infra.IncomeMonthSums {
	t.Helper()
	var rows []infra.IncomeMonthSums
	require.NoError(t, e.txm.ReadOnly(t.Context(), userID, func(ctx context.Context) error {
		var err error
		rows, err = infra.NewLedgerReader().IncomeSums(ctx, q)
		return err
	}))
	return rows
}

func rate16(base, iva money.Cents) infra.IVAByRate {
	return infra.IVAByRate{Rate16: infra.RateSums{Base: base, IVA: iva}}
}

// cashBasisFixtures cover every rate, the moral and platform withholdings, a soft delete,
// the period's edges and the same user's other taxpayer, for ledgerFisica in 2026.
func cashBasisFixtures() []incomeFixture {
	return []incomeFixture{
		// 612, January: a sale, fees paid by a persona moral (10 % ISR and 2/3 of the IVA
		// withheld) invoiced in December but collected in January (cash basis), the border
		// rate, and a soft-deleted sale that never counts.
		sale("2026-01-10", 1000_00),
		withIncome(sale("2026-01-20", 10000_00), func(f *incomeFixture) {
			f.counterparty, f.invoicedOn, f.isrWithheld, f.ivaWithheld = "persona_moral", "2025-12-28", 1000_00, 1066_67
		}),
		withIncome(sale("2026-01-25", 500_00), func(f *incomeFixture) { f.rate, f.iva = "rate_8", 40_00 }),
		withIncome(sale("2026-01-15", 9999_00), func(f *incomeFixture) { f.deleted = true }),
		// 612, March: the untaxed rates keep only their base.
		withIncome(sale("2026-03-31", 300_00), func(f *incomeFixture) { f.rate, f.iva = "rate_0", 0 }),
		withIncome(sale("2026-03-01", 400_00), func(f *incomeFixture) { f.rate, f.iva = "exempt", 0 }),
		// 612: invoiced in the period but collected after it: no ISR (cash) and no IVA yet.
		withIncome(sale("2026-10-02", 600_00), func(f *incomeFixture) { f.invoicedOn = "2026-09-15" }),
		// 625, February: a platform withholds ISR (1 %) and half of the IVA.
		withIncome(sale("2026-02-14", 2000_00), func(f *incomeFixture) {
			f.regime, f.counterparty, f.isrWithheld, f.ivaWithheld = "625", "plataforma", 20_00, 160_00
		}),
		// The period's edges: the last day counts; the day before the fiscal year and the
		// day after the period do not.
		sale("2026-09-30", 50_00),
		sale("2025-12-31", 777_00),
		sale("2026-10-01", 888_00),
		// The same user's other taxpayer: never in ledgerFisica's sums.
		withIncome(sale("2026-01-10", 123_00), func(f *incomeFixture) { f.taxpayer, f.regime = ledgerRESICO, "626" }),
	}
}

func TestLedgerReader_CashBasisExactSumsByRegimeAndMonth(t *testing.T) {
	e := newEnv(t)
	seedLedgerTaxpayers(t, e)
	seedIncomes(t, e, cashBasisFixtures()...)

	got := readIncomeSums(t, e, userA, infra.LedgerQuery{TaxpayerID: ledgerFisica, FiscalYear: 2026, Through: time.September})

	assert.Equal(t, []infra.IncomeMonthSums{
		{
			Regime: fiscal.RegimeBusiness, Year: 2026, Month: time.January, Cash: 11500_00,
			ByRate: infra.IVAByRate{
				Rate16: infra.RateSums{Base: 11000_00, IVA: 1760_00}, Rate8: infra.RateSums{Base: 500_00, IVA: 40_00},
			},
			IVAWithheld: infra.Withheld{ByMoral: 1066_67}, ISRWithheld: infra.Withheld{ByMoral: 1000_00},
		},
		{
			Regime: fiscal.RegimeBusiness, Year: 2026, Month: time.March, Cash: 700_00,
			ByRate: infra.IVAByRate{Rate0Base: 300_00, ExemptBase: 400_00},
		},
		{Regime: fiscal.RegimeBusiness, Year: 2026, Month: time.September, Cash: 50_00, ByRate: rate16(50_00, 8_00)},
		{
			Regime: fiscal.RegimePlatforms, Year: 2026, Month: time.February, Cash: 2000_00, ByRate: rate16(2000_00, 320_00),
			IVAWithheld: infra.Withheld{ByPlatform: 160_00}, ISRWithheld: infra.Withheld{ByPlatform: 20_00},
		},
	}, got, "cash basis: Accrual stays zero; ordered by regime, then month; months without records absent")
}

// An accrual regime (601) books ISR on invoiced_on, while its IVA and withholdings stay on
// collected_on: an invoice pending collection counts for ISR only, and a collection of last
// year's invoice for IVA only.
func TestLedgerReader_AccrualBasisSplitsISRFromIVA(t *testing.T) {
	e := newEnv(t)
	seedLedgerTaxpayers(t, e)
	moral := incomeFixture{
		user: userA, taxpayer: ledgerMoral, regime: "601", counterparty: "persona_moral", rate: "rate_16",
		amount: 10000_00, iva: 1600_00, invoicedOn: "2026-03-10", collectedOn: "2026-04-05",
	}
	seedIncomes(t, e,
		moral,
		withIncome(moral, func(f *incomeFixture) {
			f.amount, f.iva, f.invoicedOn, f.collectedOn = 5000_00, 800_00, "2026-03-20", ""
		}),
		withIncome(moral, func(f *incomeFixture) {
			f.amount, f.iva, f.invoicedOn, f.collectedOn = 3000_00, 480_00, "2025-12-15", "2026-01-10"
		}),
		withIncome(moral, func(f *incomeFixture) {
			f.counterparty, f.amount, f.iva, f.isrWithheld = "plataforma", 1000_00, 160_00, 25_00
			f.invoicedOn, f.collectedOn = "2026-04-02", "2026-04-02"
		}),
		withIncome(moral, func(f *incomeFixture) {
			f.amount, f.iva, f.invoicedOn, f.collectedOn = 700_00, 112_00, "2026-04-30", "2026-05-02"
		}),
		withIncome(moral, func(f *incomeFixture) { f.invoicedOn, f.deleted = "2026-04-01", true }),
	)

	got := readIncomeSums(t, e, userA, infra.LedgerQuery{TaxpayerID: ledgerMoral, FiscalYear: 2026, Through: time.April})

	assert.Equal(t, []infra.IncomeMonthSums{
		{Regime: fiscal.RegimeCode("601"), Year: 2026, Month: time.January, Cash: 3000_00, ByRate: rate16(3000_00, 480_00)},
		{Regime: fiscal.RegimeCode("601"), Year: 2026, Month: time.March, Accrual: 15000_00},
		{
			Regime: fiscal.RegimeCode("601"), Year: 2026, Month: time.April, Cash: 11000_00, Accrual: 1700_00,
			ByRate: rate16(11000_00, 1760_00), ISRWithheld: infra.Withheld{ByPlatform: 25_00},
		},
	}, got)
}

// The ISR withheld by a persona moral is the regime's kind: 10 % for 612, 1.25 % for 626.
func TestLedgerReader_RESICOWithholdingByMoral(t *testing.T) {
	e := newEnv(t)
	seedLedgerTaxpayers(t, e)
	seedIncomes(t, e, withIncome(sale("2026-02-03", 10000_00), func(f *incomeFixture) {
		f.taxpayer, f.regime, f.counterparty, f.isrWithheld, f.ivaWithheld = ledgerRESICO, "626", "persona_moral", 125_00, 1066_67
	}))

	got := readIncomeSums(t, e, userA, infra.LedgerQuery{TaxpayerID: ledgerRESICO, FiscalYear: 2026, Through: time.February})

	assert.Equal(t, []infra.IncomeMonthSums{{
		Regime: fiscal.RegimeRESICO, Year: 2026, Month: time.February, Cash: 10000_00, ByRate: rate16(10000_00, 1600_00),
		IVAWithheld: infra.Withheld{ByMoral: 1066_67}, ISRWithheld: infra.Withheld{ByMoral: 125_00},
	}}, got)
}

func TestLedgerReader_DecemberReachesTheYearEnd(t *testing.T) {
	e := newEnv(t)
	seedLedgerTaxpayers(t, e)
	seedIncomes(t, e, sale("2026-10-01", 888_00), sale("2026-12-31", 100_00), sale("2027-01-01", 999_00))

	got := readIncomeSums(t, e, userA, infra.LedgerQuery{TaxpayerID: ledgerFisica, FiscalYear: 2026, Through: time.December})

	require.Len(t, got, 2)
	assert.Equal(t, time.October, got[0].Month)
	assert.Equal(t, money.Cents(888_00), got[0].Cash)
	assert.Equal(t, time.December, got[1].Month)
	assert.Equal(t, money.Cents(100_00), got[1].Cash, "the next fiscal year never counts")
}

// M6-02: one statement reads January of FirstYear through Through of FiscalYear, each row
// naming its year, so the dashboard's two-fiscal-year window is one round trip.
func TestLedgerReader_RangeSpansTwoFiscalYears(t *testing.T) {
	e := newEnv(t)
	seedLedgerTaxpayers(t, e)
	seedIncomes(t, e, sale("2025-12-31", 777_00), sale("2026-01-02", 100_00), sale("2026-11-05", 200_00),
		sale("2027-02-03", 300_00), sale("2027-02-28", 1_00), sale("2027-03-01", 999_00))

	got := readIncomeSums(t, e, userA, infra.LedgerQuery{
		TaxpayerID: ledgerFisica, FirstYear: 2026, FiscalYear: 2027, Through: time.February,
	})

	assert.Equal(t, []infra.IncomeMonthSums{
		{Regime: fiscal.RegimeBusiness, Year: 2026, Month: time.January, Cash: 100_00, ByRate: rate16(100_00, 16_00)},
		{Regime: fiscal.RegimeBusiness, Year: 2026, Month: time.November, Cash: 200_00, ByRate: rate16(200_00, 32_00)},
		{Regime: fiscal.RegimeBusiness, Year: 2027, Month: time.February, Cash: 301_00, ByRate: rate16(301_00, 48_16)},
	}, got, "ordered by regime, year, month; nothing before January 2026 or after February 2027")
}

// M6-02 is_empty: any live income of the taxpayer, in any period; a soft-deleted one and
// another user's do not count.
func TestLedgerReader_HasLive(t *testing.T) {
	e := newEnv(t)
	seedLedgerTaxpayers(t, e)
	hasLive := func(userID, taxpayerID string) bool {
		t.Helper()
		var got bool
		require.NoError(t, e.txm.ReadOnly(t.Context(), userID, func(ctx context.Context) error {
			var err error
			got, err = infra.NewLedgerReader().HasLive(ctx, taxpayerID)
			return err
		}))
		return got
	}

	assert.False(t, hasLive(userA, ledgerFisica), "no income yet")
	seedIncomes(t, e, withIncome(sale("2026-01-10", 1_00), func(f *incomeFixture) { f.deleted = true }))
	assert.False(t, hasLive(userA, ledgerFisica), "a deleted income is not live")
	seedIncomes(t, e, sale("2026-02-10", 1_00))
	assert.True(t, hasLive(userA, ledgerFisica))
	assert.False(t, hasLive(userB, ledgerFisica), "RLS: another user's taxpayer has nothing")
	assert.False(t, hasLive(userA, ledgerRESICO), "the same user's other taxpayer")
}

func TestLedgerReader_OtherUsersRowsAreInvisible(t *testing.T) {
	e := newEnv(t)
	seedLedgerTaxpayers(t, e)
	seedIncomes(t, e, withIncome(sale("2026-01-10", 4000_00), func(f *incomeFixture) { f.user, f.taxpayer = userB, ledgerOther }))
	q := infra.LedgerQuery{TaxpayerID: ledgerOther, FiscalYear: 2026, Through: time.December}

	assert.Empty(t, readIncomeSums(t, e, userA, q), "RLS: another user's taxpayer reads as no records")
	owned := readIncomeSums(t, e, userB, q)
	require.Len(t, owned, 1)
	assert.Equal(t, money.Cents(4000_00), owned[0].Cash)
}

func TestLedgerReader_NoRecords(t *testing.T) {
	e := newEnv(t)
	seedLedgerTaxpayers(t, e)
	got := readIncomeSums(t, e, userA, infra.LedgerQuery{TaxpayerID: ledgerFisica, FiscalYear: 2026, Through: time.January})
	assert.Empty(t, got)
}

func TestLedgerReader_RefusesAnInvalidQuery(t *testing.T) {
	e := newEnv(t)
	for name, q := range map[string]infra.LedgerQuery{
		"month zero":        {TaxpayerID: ledgerFisica, FiscalYear: 2026, Through: 0},
		"month thirteen":    {TaxpayerID: ledgerFisica, FiscalYear: 2026, Through: 13},
		"before the floor":  {TaxpayerID: ledgerFisica, FiscalYear: 2021, Through: time.January},
		"malformed payer":   {TaxpayerID: "not-a-uuid", FiscalYear: 2026, Through: time.January},
		"no taxpayer given": {FiscalYear: 2026, Through: time.January},
		"first year after":  {TaxpayerID: ledgerFisica, FirstYear: 2027, FiscalYear: 2026, Through: time.January},
		"first year floor":  {TaxpayerID: ledgerFisica, FirstYear: 2021, FiscalYear: 2026, Through: time.January},
	} {
		t.Run(name, func(t *testing.T) {
			err := e.txm.ReadOnly(t.Context(), userA, func(ctx context.Context) error {
				_, err := infra.NewLedgerReader().IncomeSums(ctx, q)
				return err
			})
			require.ErrorIs(t, err, infra.ErrInvalidLedgerQuery)
		})
	}
}

func TestLedgerReader_NeedsAUnitOfWork(t *testing.T) {
	_, err := infra.NewLedgerReader().IncomeSums(t.Context(),
		infra.LedgerQuery{TaxpayerID: ledgerFisica, FiscalYear: 2026, Through: time.January})
	require.Error(t, err)
}

// PR #79's follow-up: the by-rate split keeps only the base of a 0 % or exempt income, so
// such an income with IVA would be dropped silently. The domain refuses it (US-I1,
// fiscal.CheckIVA) and the database backs it (incomes_zero_rate_has_no_iva).
func TestLedgerReader_UntaxedIncomeCannotCarryIVA(t *testing.T) {
	e := newEnv(t)
	seedLedgerTaxpayers(t, e)
	for _, rate := range []string{"rate_0", "exempt"} {
		t.Run(rate, func(t *testing.T) {
			f := withIncome(sale("2026-01-10", 1000_00), func(f *incomeFixture) { f.rate, f.iva = rate, 1_00 })
			err := pgtest.AsUser(t.Context(), e.app, userA, func(tx *sql.Tx) error {
				_, err := tx.ExecContext(t.Context(), insertIncomeFixtureSQL, f.user, f.taxpayer, f.regime, f.counterparty,
					int64(f.amount), f.rate, int64(f.iva), int64(0), int64(0), f.invoicedOn, f.collectedOn)
				return err
			})
			var pgErr *pgconn.PgError
			require.True(t, errors.As(err, &pgErr), "want a PostgreSQL error, got %v", err)
			assert.Equal(t, "incomes_zero_rate_has_no_iva", pgErr.ConstraintName)
		})
	}
}

// The store's writes (booked_on derived by the trigger) are what the aggregate sums.
func TestLedgerReader_SumsWhatTheStoreWrote(t *testing.T) {
	e := newEnv(t)
	seedLedgerTaxpayers(t, e)
	accrual := usecase.Fields{
		Regime: "601", Counterparty: fiscal.ReceptorPM, Operation: fiscal.OperationBusinessServices, Amount: 45_000_00,
		IVARate: fiscal.IVARate16, IVA: 7_200_00, Dates: domain.Dates{InvoicedOn: day(t, "2026-08-28")},
	}
	require.NoError(t, e.as(t, userA, func(ctx context.Context) error {
		if err := e.store.Insert(ctx, usecase.NewIncome{ID: id(71), UserID: userA, TaxpayerID: ledgerMoral, Fields: accrual}); err != nil {
			return err
		}
		accrual.Dates.CollectedOn = day(t, "2026-09-03")
		return e.store.Update(ctx, id(71), accrual, fiscal.WithholdingRecord{})
	}))

	got := readIncomeSums(t, e, userA, infra.LedgerQuery{TaxpayerID: ledgerMoral, FiscalYear: 2026, Through: time.September})

	assert.Equal(t, []infra.IncomeMonthSums{
		{Regime: fiscal.RegimeCode("601"), Year: 2026, Month: time.August, Accrual: 45_000_00},
		{Regime: fiscal.RegimeCode("601"), Year: 2026, Month: time.September, Cash: 45_000_00, ByRate: rate16(45_000_00, 7_200_00)},
	}, got)
}

// planUsers × planIncomesPerUser ≈ 50k rows, as the PLAN asks for the hot path.
const (
	planUsers          = 25
	planIncomesPerUser = 2000
)

func planUserID(i int) string     { return fmt.Sprintf("0199d4a2-0000-7000-8000-00000004f1%02d", i) }
func planTaxpayerID(i int) string { return fmt.Sprintf("0199d4a2-1111-7000-8000-00000004f1%02d", i) }

// seedPlanSQL writes planIncomesPerUser incomes for user $1 / moral taxpayer $2 (601,
// accrual) over 2025-01..2026-09, deterministic, with every rate and counterparty; one in
// seven is still pending collection.
var seedPlanSQL = fmt.Sprintf(`
WITH g AS (
  SELECT g, 10000 + (g * 37) %% 500000 AS amount, (ARRAY['rate_16', 'rate_8', 'rate_0', 'exempt'])[1 + g %% 4] AS rate,
         (ARRAY['persona_moral', 'publico_general', 'plataforma'])[1 + g %% 3] AS counterparty,
         date '2025-01-01' + (g * 7) %% 600 AS invoiced
  FROM generate_series(1, %d) AS g
)
INSERT INTO incomes (user_id, taxpayer_id, regime_code, counterparty_type, amount_cents, iva_rate, iva_cents,
                     isr_withheld_cents, iva_withheld_cents, invoiced_on, collected_on)
SELECT $1::uuid, $2::uuid, '601', counterparty, amount, rate,
       CASE rate WHEN 'rate_16' THEN amount * 16 / 100 WHEN 'rate_8' THEN amount * 8 / 100 ELSE 0 END,
       CASE counterparty WHEN 'plataforma' THEN amount * 25 / 1000 ELSE 0 END, 0,
       invoiced, CASE WHEN g %% 7 = 0 THEN NULL ELSE invoiced + g %% 40 END
FROM g`, planIncomesPerUser)

// EXPLAIN ANALYZE of the reader's own statement at ~50k rows, as cc_app under RLS: the two
// partial indexes serve the incomes, never a sequential scan (data-model.md: "Each
// critical query gets EXPLAIN ANALYZE in its module PR"). The basis lookup reads one
// taxpayer_regimes row: on a table this small (one row per plan taxpayer) the planner's
// cheapest read is the single heap page, so the test also proves that its unique keys
// serve the lookup once the table grows (enable_seqscan off: a plan without any
// sequential scan exists).
func TestLedgerReader_IndexesServeTheAggregateAt50kRows(t *testing.T) {
	e := newEnv(t)
	for i := range planUsers {
		user, taxpayer := planUserID(i), planTaxpayerID(i)
		for _, statement := range registerSQL {
			execAs(t, e.app, user, statement, user)
		}
		seedLedgerTaxpayer(t, e.app, user, taxpayer, "moral", "601")
		execAs(t, e.app, user, seedPlanSQL, user, taxpayer)
		// One in twenty is soft-deleted.
		execAs(t, e.app, user, "UPDATE incomes SET deleted_at = now() WHERE amount_cents % 20 = 0")
	}
	_, err := e.db.Connect(t, pgtest.RoleOwner).ExecContext(t.Context(), "ANALYZE")
	require.NoError(t, err)
	var total int
	queryAs(t, e.app, planUserID(0), "SELECT reltuples::int FROM pg_class WHERE oid = 'incomes'::regclass", nil, &total)
	require.GreaterOrEqual(t, total, planUsers*planIncomesPerUser*9/10, "~50k rows")

	q := infra.LedgerQuery{TaxpayerID: planTaxpayerID(7), FiscalYear: 2026, Through: time.September}
	plan := explainIncomeSums(t, e, planUserID(7), q, false)
	t.Logf("EXPLAIN (ANALYZE, BUFFERS) of the M3-07 aggregate, 2026-01..09, as cc_app:\n%s", plan)
	assert.True(t, strings.Contains(plan, "incomes_booked_idx"), "ISR by booked_on over its partial index")
	assert.True(t, strings.Contains(plan, "incomes_collected_idx"), "IVA and withholdings by collected_on over its partial index")
	assert.False(t, strings.Contains(plan, "Seq Scan on incomes"), "never a sequential scan of incomes")
	assert.True(t, strings.Contains(plan, "Aggregate"), "GROUP BY regime, month")

	indexOnly := explainIncomeSums(t, e, planUserID(7), q, true)
	t.Logf("The same statement with enable_seqscan off:\n%s", indexOnly)
	assert.True(t, strings.Contains(indexOnly, "Index Scan using taxpayer_regimes_"), "the basis lookup has an index path")
	assert.False(t, strings.Contains(indexOnly, "Seq Scan"), "a plan without any sequential scan exists")
}

// explainIncomeSums is the reader's plan for q as userID; noSeqScan disables sequential
// scans for the transaction first.
func explainIncomeSums(t *testing.T, e *env, userID string, q infra.LedgerQuery, noSeqScan bool) string {
	t.Helper()
	var plan string
	require.NoError(t, e.txm.ReadOnly(t.Context(), userID, func(ctx context.Context) error {
		if noSeqScan {
			db, err := pgdb.Conn(ctx)
			if err != nil {
				return err
			}
			if err := db.Exec("SET LOCAL enable_seqscan = off").Error; err != nil {
				return err
			}
		}
		var err error
		plan, err = infra.ExplainIncomeSums(ctx, q)
		return err
	}))
	return plan
}
