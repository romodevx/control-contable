//go:build integration

package infra_test

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/expenses/domain"
	"github.com/romodevx/control-contable/backend/internal/expenses/infra"
	"github.com/romodevx/control-contable/backend/internal/expenses/usecase"
	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
)

// The tax engine's expense aggregate (M4-08, MODULE-tax-engine §2.4): one GROUP BY
// regime, month over the live expenses of one taxpayer, by paid_on, as cc_app under the
// context user's RLS.

// ledgerFixture is one expense row as the fixtures insert it.
type ledgerFixture struct {
	user, taxpayer, regime, category string
	amount, iva                      money.Cents
	rate, method                     string
	cfdi, deductible, cashOver       bool
	creditable, isrWithheld          money.Cents
	ivaWithheld                      money.Cents
	paidOn                           string
	deleted                          bool
}

const insertLedgerFixtureSQL = `
INSERT INTO expenses (user_id, taxpayer_id, regime_code, category_code, amount_cents, iva_rate, iva_cents,
                      payment_method, has_cfdi, deductible, deductible_overridden, cash_over_limit,
                      creditable_iva_cents, isr_withheld_cents, iva_withheld_cents, paid_on)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16::date)
RETURNING id`

func seedLedger(t *testing.T, e *env, fixtures ...ledgerFixture) {
	t.Helper()
	for _, f := range fixtures {
		var id string
		queryAs(t, e.app, f.user, insertLedgerFixtureSQL, []any{
			f.user, f.taxpayer, f.regime, f.category, int64(f.amount), f.rate, int64(f.iva), f.method, f.cfdi,
			f.deductible, f.cashOver && f.deductible, f.cashOver, int64(f.creditable), int64(f.isrWithheld),
			int64(f.ivaWithheld), f.paidOn,
		}, &id)
		if f.deleted {
			execAs(t, e.app, f.user, "UPDATE expenses SET deleted_at = now() WHERE id = $1", id)
		}
	}
}

// card is a deductible 16 % expense of taxpayerA's 612, paid by card with a CFDI and
// its IVA creditable; the cases change what they test.
func card(paidOn string, amount money.Cents) ledgerFixture {
	return ledgerFixture{
		user: userA, taxpayer: taxpayerA, regime: "612", category: "software", amount: amount, iva: amount * 16 / 100,
		rate: "rate_16", method: "card_credit", cfdi: true, deductible: true, creditable: amount * 16 / 100,
		paidOn: paidOn,
	}
}

func with(f ledgerFixture, change func(*ledgerFixture)) ledgerFixture {
	change(&f)
	return f
}

// taxpayerAFixtures cover every class, rate and boundary for taxpayerA in 2026.
func taxpayerAFixtures() []ledgerFixture {
	return []ledgerFixture{
		// 612, January: one of each deductibility class, plus a cash payment over the
		// limit the user claimed deductible (deductible class, IVA never creditable).
		card("2026-01-10", 1000_00),
		with(card("2026-01-20", 500_00), func(f *ledgerFixture) {
			f.category, f.deductible, f.creditable = "restaurantes", false, 0
		}),
		with(card("2026-01-25", 3000_00), func(f *ledgerFixture) {
			f.category, f.method, f.deductible, f.cashOver, f.creditable = "papeleria", "cash", false, true, 0
		}),
		with(card("2026-01-31", 2500_00), func(f *ledgerFixture) {
			f.category, f.method, f.cashOver, f.creditable = "papeleria", "cash", true, 0
		}),
		// Soft-deleted: never counted.
		with(card("2026-01-15", 9999_00), func(f *ledgerFixture) { f.deleted = true }),
		// 612, March: the other IVA rates.
		with(card("2026-03-01", 200_00), func(f *ledgerFixture) { f.rate, f.iva, f.creditable = "rate_8", 16_00, 16_00 }),
		with(card("2026-03-31", 300_00), func(f *ledgerFixture) { f.rate, f.iva, f.creditable = "rate_0", 0, 0 }),
		with(card("2026-03-15", 400_00), func(f *ledgerFixture) {
			f.rate, f.iva, f.creditable, f.cfdi, f.deductible = "exempt", 0, 0, false, false
		}),
		// 606, February: predial counts only when deductible; a small cash payment is
		// not over the limit.
		with(card("2026-02-05", 1200_00), func(f *ledgerFixture) {
			f.regime, f.category, f.rate, f.iva, f.creditable = "606", "predial", "exempt", 0, 0
		}),
		with(card("2026-02-06", 100_00), func(f *ledgerFixture) {
			f.regime, f.category, f.rate, f.iva, f.creditable = "606", "predial", "exempt", 0, 0
			f.method, f.cfdi, f.deductible = "cash", false, false
		}),
		with(card("2026-02-07", 700_00), func(f *ledgerFixture) {
			f.regime, f.category, f.deductible, f.creditable = "606", "mantenimiento", false, 0
		}),
		// The period's edges: the last day of the period counts; the day before the
		// fiscal year and the day after the period do not.
		card("2026-09-30", 50_00),
		card("2025-12-31", 777_00),
		card("2026-10-01", 888_00),
		// The same user's other taxpayer: never in taxpayerA's sums.
		with(card("2026-01-10", 123_00), func(f *ledgerFixture) { f.taxpayer, f.regime = taxpayerR, "626" }),
	}
}

func readExpenseSums(t *testing.T, e *env, userID string, q infra.LedgerQuery) []infra.ExpenseMonthSums {
	t.Helper()
	var rows []infra.ExpenseMonthSums
	require.NoError(t, e.txm.ReadOnly(t.Context(), userID, func(ctx context.Context) error {
		var err error
		rows, err = infra.NewLedgerReader().ExpenseSums(ctx, q)
		return err
	}))
	return rows
}

func rate16(base, iva money.Cents) infra.IVAByRate {
	return infra.IVAByRate{Rate16: infra.RateSums{Base: base, IVA: iva}}
}

func TestLedgerReader_ExactSumsByRegimeAndMonth(t *testing.T) {
	e := newEnv(t)
	seedLedger(t, e, taxpayerAFixtures()...)

	got := readExpenseSums(t, e, userA, infra.LedgerQuery{TaxpayerID: taxpayerA, FiscalYear: 2026, Through: time.September})

	assert.Equal(t, []infra.ExpenseMonthSums{
		{
			Regime: fiscal.RegimeRent, Year: 2026, Month: time.February,
			DeductiblePaid: 1200_00, NonDeductiblePaid: 800_00,
			IVAPaid:     infra.IVAByRate{Rate16: infra.RateSums{Base: 700_00, IVA: 112_00}, ExemptBase: 1300_00},
			PredialPaid: 1200_00,
		},
		{
			Regime: fiscal.RegimeBusiness, Year: 2026, Month: time.January,
			DeductiblePaid: 3500_00, NonDeductiblePaid: 500_00, CashOverLimitPaid: 3000_00,
			IVAPaid: rate16(7000_00, 1120_00), CreditableIVAPaid: 160_00,
		},
		{
			Regime: fiscal.RegimeBusiness, Year: 2026, Month: time.March,
			DeductiblePaid: 500_00, NonDeductiblePaid: 400_00,
			IVAPaid: infra.IVAByRate{
				Rate8: infra.RateSums{Base: 200_00, IVA: 16_00}, Rate0Base: 300_00, ExemptBase: 400_00,
			},
			CreditableIVAPaid: 16_00,
		},
		{
			Regime: fiscal.RegimeBusiness, Year: 2026, Month: time.September,
			DeductiblePaid: 50_00, IVAPaid: rate16(50_00, 8_00), CreditableIVAPaid: 8_00,
		},
	}, got, "ordered by regime, then month; months without records absent")
}

func TestLedgerReader_DecemberReachesTheYearEnd(t *testing.T) {
	e := newEnv(t)
	seedLedger(t, e, card("2026-10-01", 888_00), card("2026-12-31", 100_00), card("2027-01-01", 999_00))

	got := readExpenseSums(t, e, userA, infra.LedgerQuery{TaxpayerID: taxpayerA, FiscalYear: 2026, Through: time.December})

	require.Len(t, got, 2)
	assert.Equal(t, time.October, got[0].Month)
	assert.Equal(t, money.Cents(888_00), got[0].DeductiblePaid)
	assert.Equal(t, time.December, got[1].Month)
	assert.Equal(t, money.Cents(100_00), got[1].DeductiblePaid, "the next fiscal year never counts")
}

// M6-02: one statement reads January of FirstYear through Through of FiscalYear, each row
// naming its year, so the dashboard's two-fiscal-year window is one round trip.
func TestLedgerReader_RangeSpansTwoFiscalYears(t *testing.T) {
	e := newEnv(t)
	seedLedger(t, e, card("2025-12-31", 777_00), card("2026-01-02", 100_00), card("2026-11-05", 200_00),
		card("2027-02-03", 300_00), card("2027-02-28", 1_00), card("2027-03-01", 999_00))

	got := readExpenseSums(t, e, userA, infra.LedgerQuery{
		TaxpayerID: taxpayerA, FirstYear: 2026, FiscalYear: 2027, Through: time.February,
	})

	paid := func(year int, month time.Month, amount money.Cents) infra.ExpenseMonthSums {
		return infra.ExpenseMonthSums{
			Regime: fiscal.RegimeBusiness, Year: year, Month: month, DeductiblePaid: amount,
			IVAPaid: rate16(amount, amount*16/100), CreditableIVAPaid: amount * 16 / 100,
		}
	}
	assert.Equal(t, []infra.ExpenseMonthSums{
		paid(2026, time.January, 100_00), paid(2026, time.November, 200_00), paid(2027, time.February, 301_00),
	}, got, "ordered by regime, year, month; nothing before January 2026 or after February 2027")
}

func TestLedgerReader_WithholdingsMadeToSuppliers(t *testing.T) {
	e := newEnv(t)
	fees := ledgerFixture{
		user: userA, taxpayer: taxpayerM, regime: "601", category: "servicios_profesionales",
		amount: 10000_00, iva: 1600_00, rate: "rate_16", method: "transfer", cfdi: true, deductible: true,
		creditable: 1600_00, isrWithheld: 1000_00, ivaWithheld: 1066_67, paidOn: "2026-04-10",
	}
	notDeductible := with(fees, func(f *ledgerFixture) {
		f.category, f.deductible, f.creditable, f.amount, f.iva = "renta_local", false, 0, 5000_00, 800_00
		f.isrWithheld, f.ivaWithheld = 500_00, 533_33
	})
	seedLedger(t, e, fees, notDeductible)

	got := readExpenseSums(t, e, userA, infra.LedgerQuery{TaxpayerID: taxpayerM, FiscalYear: 2026, Through: time.April})

	assert.Equal(t, []infra.ExpenseMonthSums{{
		Regime: fiscal.RegimeCode("601"), Year: 2026, Month: time.April,
		DeductiblePaid: 10000_00, NonDeductiblePaid: 5000_00,
		IVAPaid: rate16(15000_00, 2400_00), CreditableIVAPaid: 1600_00,
		ISRWithheldToRemit: 1500_00, IVAWithheldToRemit: 1600_00,
	}}, got, "withholdings are remitted whatever the deductibility")
}

func TestLedgerReader_OtherUsersRowsAreInvisible(t *testing.T) {
	e := newEnv(t)
	seedLedger(t, e, with(card("2026-01-10", 4000_00), func(f *ledgerFixture) { f.user, f.taxpayer = userB, taxpayerB }))
	q := infra.LedgerQuery{TaxpayerID: taxpayerB, FiscalYear: 2026, Through: time.December}

	assert.Empty(t, readExpenseSums(t, e, userA, q), "RLS: another user's taxpayer reads as no records")
	owned := readExpenseSums(t, e, userB, q)
	require.Len(t, owned, 1)
	assert.Equal(t, money.Cents(4000_00), owned[0].DeductiblePaid)
}

func TestLedgerReader_NoRecords(t *testing.T) {
	e := newEnv(t)
	got := readExpenseSums(t, e, userA, infra.LedgerQuery{TaxpayerID: taxpayerA, FiscalYear: 2026, Through: time.January})
	assert.Empty(t, got)
}

func TestLedgerReader_RefusesAnInvalidQuery(t *testing.T) {
	e := newEnv(t)
	for name, q := range map[string]infra.LedgerQuery{
		"month zero":        {TaxpayerID: taxpayerA, FiscalYear: 2026, Through: 0},
		"month thirteen":    {TaxpayerID: taxpayerA, FiscalYear: 2026, Through: 13},
		"before the floor":  {TaxpayerID: taxpayerA, FiscalYear: 2021, Through: time.January},
		"malformed payer":   {TaxpayerID: "not-a-uuid", FiscalYear: 2026, Through: time.January},
		"no taxpayer given": {FiscalYear: 2026, Through: time.January},
		"first year after":  {TaxpayerID: taxpayerA, FirstYear: 2027, FiscalYear: 2026, Through: time.January},
		"first year floor":  {TaxpayerID: taxpayerA, FirstYear: 2021, FiscalYear: 2026, Through: time.January},
	} {
		t.Run(name, func(t *testing.T) {
			err := e.txm.ReadOnly(t.Context(), userA, func(ctx context.Context) error {
				_, err := infra.NewLedgerReader().ExpenseSums(ctx, q)
				return err
			})
			require.ErrorIs(t, err, infra.ErrInvalidLedgerQuery)
		})
	}
}

func TestLedgerReader_NeedsAUnitOfWork(t *testing.T) {
	_, err := infra.NewLedgerReader().ExpenseSums(t.Context(),
		infra.LedgerQuery{TaxpayerID: taxpayerA, FiscalYear: 2026, Through: time.January})
	require.Error(t, err)
}

// The use cases store the assessed facts the aggregate sums, end to end.
func TestLedgerReader_SumsWhatTheUseCasesStored(t *testing.T) {
	e := newEnv(t)
	draft := func(category domain.CategoryCode, method domain.PaymentMethod, amount, iva money.Cents, paidOn string) domain.Draft {
		return domain.Draft{
			Category: category, Amount: amount, IVARate: fiscal.IVARate16, IVA: iva, PaymentMethod: method,
			PaidOn: date(t, paidOn), HasCFDI: true,
		}
	}
	for i, d := range []domain.Draft{
		draft("software", domain.PaymentCardCredit, 1000_00, 160_00, "2026-09-01"),
		draft("papeleria", domain.PaymentCash, 2000_00, 320_00, "2026-09-02"), // over the $2,000 limit
		draft("combustible", domain.PaymentCash, 300_00, 48_00, "2026-09-03"), // fuel in cash, any amount
		draft("papeleria", domain.PaymentCash, 1000_00, 160_00, "2026-09-04"), // under the limit
	} {
		_, err := e.uc.Create(t.Context(), usecase.CreateInput{
			UserID: userA, TaxpayerID: taxpayerA, ID: fmt.Sprintf("0199d4a2-2222-7000-8000-%012d", i+1),
			Regime: fiscal.RegimeBusiness, Draft: d,
		})
		require.NoError(t, err)
	}

	got := readExpenseSums(t, e, userA, infra.LedgerQuery{TaxpayerID: taxpayerA, FiscalYear: 2026, Through: time.September})

	assert.Equal(t, []infra.ExpenseMonthSums{{
		Regime: fiscal.RegimeBusiness, Year: 2026, Month: time.September,
		DeductiblePaid: 2000_00, CashOverLimitPaid: 2300_00,
		IVAPaid: rate16(4300_00, 688_00), CreditableIVAPaid: 320_00,
	}}, got)
}

// planUsers × planExpensesPerUser ≈ 50k rows, as the PLAN asks for the hot path.
const (
	planUsers           = 25
	planExpensesPerUser = 2000
)

func planUserID(i int) string     { return fmt.Sprintf("0199d4a2-0000-7000-8000-0000000f00%02d", i) }
func planTaxpayerID(i int) string { return fmt.Sprintf("0199d4a2-1111-7000-8000-0000000f00%02d", i) }

// seedPlanSQL writes planExpensesPerUser expenses for user $1 / taxpayer $2 over
// 2025-01..2026-09, deterministic, in 612 and 606, with every class and rate.
var seedPlanSQL = fmt.Sprintf(`
INSERT INTO expenses (user_id, taxpayer_id, regime_code, category_code, amount_cents, iva_rate, iva_cents,
                      payment_method, has_cfdi, deductible, cash_over_limit, creditable_iva_cents,
                      isr_withheld_cents, iva_withheld_cents, paid_on)
SELECT $1::uuid, $2::uuid, (ARRAY['612', '606'])[1 + g %% 2],
       (ARRAY['software', 'papeleria', 'predial', 'restaurantes'])[1 + g %% 4],
       10000 + (g * 37) %% 500000, 'rate_16', round((10000 + (g * 37) %% 500000) * 0.16)::bigint,
       CASE WHEN g %% 5 = 0 THEN 'cash' ELSE 'transfer' END,
       g %% 10 <> 0, g %% 3 <> 0, g %% 5 = 0,
       CASE WHEN g %% 10 <> 0 AND g %% 5 <> 0 THEN round((10000 + (g * 37) %% 500000) * 0.16)::bigint ELSE 0 END,
       0, 0,
       date '2025-01-01' + (g * 7) %% 638
FROM generate_series(1, %d) AS g`, planExpensesPerUser)

// EXPLAIN ANALYZE of the reader's own statement at ~50k rows, as cc_app under RLS: the
// partial index serves it, never a sequential scan (data-model.md: "Each critical query
// gets EXPLAIN ANALYZE in its module PR").
func TestLedgerReader_IndexServesTheAggregateAt50kRows(t *testing.T) {
	e := newEnv(t)
	for i := range planUsers {
		user, taxpayer := planUserID(i), planTaxpayerID(i)
		for _, statement := range registerSQL {
			execAs(t, e.app, user, statement, user)
		}
		execAs(t, e.app, user, insertTaxpayerSQL, user, taxpayer, "fisica")
		execAs(t, e.app, user, insertRegimeSQL, user, taxpayer, "612", "fisica", "{}", "2025-01", nil)
		execAs(t, e.app, user, insertRegimeSQL, user, taxpayer, "606", "fisica", "{}", "2025-01", nil)
		execAs(t, e.app, user, seedPlanSQL, user, taxpayer)
		// One in twenty is soft-deleted.
		execAs(t, e.app, user, "UPDATE expenses SET deleted_at = now() WHERE amount_cents % 20 = 0")
	}
	analyze(t, e)
	var total int
	queryAs(t, e.app, planUserID(0), "SELECT reltuples::int FROM pg_class WHERE oid = 'expenses'::regclass", nil, &total)
	require.GreaterOrEqual(t, total, planUsers*planExpensesPerUser*9/10, "~50k rows")

	user := planUserID(7)
	var plan string
	require.NoError(t, e.txm.ReadOnly(t.Context(), user, func(ctx context.Context) error {
		var err error
		plan, err = infra.ExplainExpenseSums(ctx,
			infra.LedgerQuery{TaxpayerID: planTaxpayerID(7), FiscalYear: 2026, Through: time.September})
		return err
	}))
	t.Logf("EXPLAIN (ANALYZE, BUFFERS) of the M4-08 aggregate, 2026-01..09, as cc_app:\n%s", plan)
	assert.Contains(t, plan, "expenses_live_paid_on_idx")
	assert.NotContains(t, plan, "Seq Scan")
	assert.True(t, strings.Contains(plan, "Aggregate"), "one GROUP BY")
}
