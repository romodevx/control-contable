//go:build integration

package main

import (
	"bytes"
	"context"
	"database/sql"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/apperr"
	"github.com/romodevx/control-contable/backend/internal/platform/clock"
	"github.com/romodevx/control-contable/backend/internal/platform/config"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
	"github.com/romodevx/control-contable/backend/internal/platform/pgtest"
	"github.com/romodevx/control-contable/backend/internal/platform/ratelimit/ratelimitpg"
	taxdomain "github.com/romodevx/control-contable/backend/internal/taxengine/domain"
	"github.com/romodevx/control-contable/backend/internal/taxengine/infra/taxinputs"
	taxusecase "github.com/romodevx/control-contable/backend/internal/taxengine/usecase"
	tpdomain "github.com/romodevx/control-contable/backend/internal/taxpayers/domain"
	taxpayers "github.com/romodevx/control-contable/backend/internal/taxpayers/usecase"
)

// The estimate (M5-22) over the real readers, wired by the composition root
// as cc_app under RLS: the taxpayers profile, the incomes and expenses
// aggregates (M3-07, M4-08) and the tax inputs (M5-21), all in one read-only
// unit of work of the taxpayers module's manager. Fixture ids are this
// suite's own; every test gets its own cloned database.
const (
	estUserA = "0199d4a2-0000-7000-8000-0000000006a1"
	estUserB = "0199d4a2-0000-7000-8000-0000000006b1"

	estRFCA = "PEXJ900215HK8"
	estRFCB = "AAAA010101AA0"
)

// estToday is 20 October 2026, noon in Mexico City.
var estToday = time.Date(2026, time.October, 20, 12, 0, 0, 0, clock.Location())

type estimateEnv struct {
	app       *sql.DB
	deps      taxpayers.Deps
	estimator *taxusecase.Estimator
}

func newEstimateEnv(t *testing.T, users ...string) *estimateEnv {
	t.Helper()
	app := pgtest.NewDatabase(t).Connect(t, pgtest.RoleApp)
	for _, u := range users {
		require.NoError(t, pgtest.AsUser(t.Context(), app, u, func(tx *sql.Tx) error {
			for _, statement := range incRegisterSQL {
				if _, err := tx.ExecContext(t.Context(), statement, u); err != nil {
					return err
				}
			}
			return nil
		}))
	}
	logger := slog.New(slog.DiscardHandler)
	appTx := newAppTx(t, app)
	deps, err := newTaxpayersDeps(taxpayersWiring{env: config.EnvDev, keys: config.Taxpayers{RFCKeyID: 1}, tx: appTx, logger: logger})
	require.NoError(t, err)
	limits, err := ratelimitpg.NewStore(app, bytes.Repeat([]byte{0x5e}, 32), time.Now)
	require.NoError(t, err)
	holidays, err := loadHolidaySnapshots(t.Context(), appTx)
	require.NoError(t, err)
	module, err := newTaxEngineModule(taxEngineWiring{
		limits: limits, logger: logger,
		tx: deps.Tx, taxpayers: taxpayers.NewQueries(deps), holidays: holidays.taxEngine, clock: clock.NewFixed(estToday),
	})
	require.NoError(t, err)
	return &estimateEnv{app: app, deps: deps, estimator: module.estimator}
}

// taxpayer creates a física of user holding regimes since January 2026.
func (e *estimateEnv) taxpayer(t *testing.T, user, rfc string, regimes ...taxpayers.RegimeInput) string {
	t.Helper()
	created, err := taxpayers.NewTaxpayers(e.deps).Create(t.Context(), user, taxpayers.CreateInput{
		PersonType: "fisica", RFC: rfc, DisplayName: "Estimación", Regimes: regimes,
	})
	require.NoError(t, err)
	return created.Taxpayer.ID
}

const estInsertIncomeSQL = `
INSERT INTO incomes (user_id, taxpayer_id, regime_code, counterparty_type, amount_cents, iva_rate, iva_cents,
                     isr_withheld_cents, iva_withheld_cents, collected_on)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10::date)`

const estInsertExpenseSQL = `
INSERT INTO expenses (user_id, taxpayer_id, regime_code, category_code, amount_cents, iva_rate, iva_cents,
                      payment_method, has_cfdi, deductible, deductible_overridden, cash_over_limit,
                      creditable_iva_cents, isr_withheld_cents, iva_withheld_cents, paid_on)
VALUES ($1, $2, $3, $4, $5, 'rate_16', $6, 'card_credit', true, $7, false, false, $8, 0, 0, $9::date)`

type estIncome struct {
	regime, counterparty, rate string
	amount, iva, isr, ivaHeld  money.Cents
	on                         string
}

type estExpense struct {
	regime, category    string
	amount, iva, credit money.Cents
	deductible          bool
	on                  string
}

func (e *estimateEnv) seed(t *testing.T, user, taxpayer string, incomes []estIncome, expenses []estExpense) {
	t.Helper()
	require.NoError(t, pgtest.AsUser(t.Context(), e.app, user, func(tx *sql.Tx) error {
		for _, i := range incomes {
			if _, err := tx.ExecContext(t.Context(), estInsertIncomeSQL, user, taxpayer, i.regime, i.counterparty,
				int64(i.amount), i.rate, int64(i.iva), int64(i.isr), int64(i.ivaHeld), i.on); err != nil {
				return err
			}
		}
		for _, x := range expenses {
			if _, err := tx.ExecContext(t.Context(), estInsertExpenseSQL, user, taxpayer, x.regime, x.category,
				int64(x.amount), int64(x.iva), x.deductible, int64(x.credit), x.on); err != nil {
				return err
			}
		}
		return nil
	}))
}

// priorPayments stores 612's prior provisional payments for 2026.
func (e *estimateEnv) priorPayments(t *testing.T, user, taxpayer string, byMonth map[time.Month]money.Cents) {
	t.Helper()
	var values []taxusecase.StoredInput
	for m, c := range byMonth {
		values = append(values, taxusecase.StoredInput{
			Key: taxdomain.InputPriorProvisionalPayments, Kind: taxdomain.InputMoneyMonthly, Month: m, Cents: c,
		})
	}
	key := taxusecase.InputsKey{TaxpayerID: taxpayer, Regime: fiscal.RegimeBusiness, FiscalYear: 2026}
	require.NoError(t, e.deps.Tx.ReadWrite(t.Context(), user, func(ctx context.Context) error {
		return taxinputs.NewStore().Replace(ctx, user, key, values)
	}))
}

// txNet01 seeds TX-NET-01 (MODULE-tax-engine §3) as records: 612 with the
// research's PF-612-ISR-01 and PF-612-IVA-01 March, and 606 renting
// commercially to a PM in March with creditable IVA 3,000.00.
func (e *estimateEnv) txNet01(t *testing.T, user, taxpayer string) {
	t.Helper()
	e.txNet01Records(t, user, taxpayer)
	e.priorPayments(t, user, taxpayer, map[time.Month]money.Cents{time.January: 2_000_00, time.February: 2_000_00})
}

// txNet01Records seeds TX-NET-01's incomes and expenses only (it needs no
// more than e.app).
func (e *estimateEnv) txNet01Records(t *testing.T, user, taxpayer string) {
	t.Helper()
	fees := func(on string) estIncome {
		return estIncome{regime: "612", counterparty: "persona_moral", rate: "rate_0", amount: 30_000_00, isr: 1_000_00, on: on}
	}
	deduction := func(on string, credit money.Cents) estExpense {
		return estExpense{regime: "612", category: "software", amount: 10_000_00, iva: 1_600_00, credit: credit, deductible: true, on: on}
	}
	e.seed(t, user, taxpayer,
		[]estIncome{
			fees("2026-01-15"), fees("2026-02-15"),
			{regime: "612", counterparty: "publico_general", rate: "rate_16", amount: 20_000_00, iva: 3_200_00, on: "2026-03-10"},
			{
				regime: "612", counterparty: "persona_moral", rate: "rate_16", amount: 10_000_00, iva: 1_600_00,
				isr: 1_000_00, ivaHeld: 1_066_67, on: "2026-03-20",
			},
			{
				regime: "606", counterparty: "persona_moral", rate: "rate_16", amount: 20_000_00, iva: 3_200_00,
				isr: 2_000_00, ivaHeld: 2_133_33, on: "2026-03-05",
			},
		},
		[]estExpense{
			deduction("2026-01-20", 0), deduction("2026-02-20", 0), deduction("2026-03-20", 1_600_00),
			{regime: "606", category: "mantenimiento", amount: 18_750_00, iva: 3_000_00, credit: 3_000_00, on: "2026-03-25"},
		})
}

func estRegimes() []taxpayers.RegimeInput {
	return []taxpayers.RegimeInput{
		{Code: "612", Start: "2026-01"},
		{Code: "606", Start: "2026-01", Options: map[string]tpdomain.OptionValue{"deduction_mode": tpdomain.StringOption("blind_35")}},
	}
}

// TX-NET-01 end to end over PostgreSQL: one IVA of 200.00 across 612 and
// 606, 612's ISR 150.96, 606's favor balance kept apart, total 350.96.
func TestEstimateMonth_overPostgresTXNet01(t *testing.T) {
	e := newEstimateEnv(t, estUserA)
	id := e.taxpayer(t, estUserA, estRFCA, estRegimes()...)
	e.txNet01(t, estUserA, id)

	got, err := e.estimator.EstimateMonth(t.Context(), estUserA, id, "2026-03")

	require.NoError(t, err)
	require.Len(t, got.Regimes, 2)
	rent, business := got.Regimes[0], got.Regimes[1]
	assert.Equal(t, taxdomain.StatusComplete, business.Status)
	assert.Equal(t, money.Cents(150_96), business.ISR.Amount)
	assert.Equal(t, money.Cents(-924_00), rent.ISR.Amount)
	assert.Equal(t, money.Cents(200_00), got.Totals.IVA.Amount)
	assert.Equal(t, money.Cents(350_96), got.Totals.TotalPayable)
	assert.Equal(t, "2026-04-17", got.DueDate.String())
	assert.Equal(t, taxdomain.DisclaimerCPAReviewPending, got.Disclaimers[1].Code)
	assert.Equal(t, []taxdomain.CPATopic{taxdomain.TopicCrossRegimeIVANetting}, got.Disclaimers[1].Topics)

	again, err := e.estimator.EstimateMonth(t.Context(), estUserA, id, "2026-03")
	require.NoError(t, err)
	assert.Equal(t, got.RecordsFingerprint, again.RecordsFingerprint, "same records, same fingerprint")
	assert.Equal(t, got.FiguresFingerprint, again.FiguresFingerprint, "same figures, same figures fingerprint")
}

// Without the stored prior payments the estimate is incomplete and names
// the missing months (TX-612-INC-01's shape over the database).
func TestEstimateMonth_overPostgresIncompleteWithoutInputs(t *testing.T) {
	e := newEstimateEnv(t, estUserA)
	id := e.taxpayer(t, estUserA, estRFCA, estRegimes()...)
	e.txNet01(t, estUserA, id)
	e.priorPayments(t, estUserA, id, nil)

	got, err := e.estimator.EstimateMonth(t.Context(), estUserA, id, "2026-03")

	require.NoError(t, err)
	business := got.Regimes[1]
	assert.Equal(t, taxdomain.StatusIncomplete, business.Status)
	assert.Equal(t, money.Cents(4_150_96), business.ISR.Amount)
	assert.Equal(t, []taxdomain.MissingInput{{
		Key: taxdomain.InputPriorProvisionalPayments, Months: []time.Month{time.January, time.February},
	}}, business.MissingInputs)
}

// S-2: another user's taxpayer is 404, and that user's own estimate never
// sees the first user's records.
func TestEstimateMonth_overPostgresIsolation(t *testing.T) {
	e := newEstimateEnv(t, estUserA, estUserB)
	idA := e.taxpayer(t, estUserA, estRFCA, estRegimes()...)
	e.txNet01(t, estUserA, idA)
	idB := e.taxpayer(t, estUserB, estRFCB, estRegimes()...)

	_, err := e.estimator.EstimateMonth(t.Context(), estUserB, idA, "2026-03")
	requireAppCode(t, err, apperr.CodeNotFound)

	own, err := e.estimator.EstimateMonth(t.Context(), estUserB, idB, "2026-03")
	require.NoError(t, err)
	assert.Equal(t, taxdomain.OutcomeZero, own.Totals.IVA.Outcome())
	assert.Zero(t, own.Totals.TotalPayable)
}
