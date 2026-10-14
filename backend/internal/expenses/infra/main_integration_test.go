//go:build integration

package infra_test

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/expenses/domain"
	"github.com/romodevx/control-contable/backend/internal/expenses/infra"
	"github.com/romodevx/control-contable/backend/internal/expenses/usecase"
	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/access"
	"github.com/romodevx/control-contable/backend/internal/platform/access/accessgorm"
	"github.com/romodevx/control-contable/backend/internal/platform/clock"
	"github.com/romodevx/control-contable/backend/internal/platform/pgdb"
	"github.com/romodevx/control-contable/backend/internal/platform/pgtest"
	"github.com/romodevx/control-contable/backend/internal/taxengine/infra/ratecatalog"
	"github.com/romodevx/control-contable/backend/internal/taxengine/params"
	tpinfra "github.com/romodevx/control-contable/backend/internal/taxpayers/infra"
	tpusecase "github.com/romodevx/control-contable/backend/internal/taxpayers/usecase"
	"github.com/romodevx/control-contable/backend/migrations"
)

func TestMain(m *testing.M) {
	os.Exit(pgtest.Main(m, pgtest.WithMigrator(migrations.Up)))
}

// The expense use cases over the real stores, migrations, access policy,
// taxpayers queries and tax parameters, as cc_app under RLS, wired the way
// cmd/api will wire them (M4-07). Fixture ids are this suite's own; every
// test gets its own cloned database.
const (
	userA = "0199d4a2-0000-7000-8000-0000000e000a"
	userB = "0199d4a2-0000-7000-8000-0000000e000b"

	taxpayerA   = "0199d4a2-1111-7000-8000-0000000e000a" // física: 612 + 606 (blind 35 %)
	taxpayerR   = "0199d4a2-1111-7000-8000-0000000e00a2" // física: RESICO (626)
	taxpayerM   = "0199d4a2-1111-7000-8000-0000000e00a3" // moral: 601
	taxpayerOld = "0199d4a2-1111-7000-8000-0000000e00a4" // física: 612 since 2022
	taxpayerB   = "0199d4a2-1111-7000-8000-0000000e000b" // userB's física: 612
)

// today is the use cases' frozen clock: 13 October 2026 in Mexico City.
var today = time.Date(2026, time.October, 13, 12, 0, 0, 0, clock.Location())

// registerSQL is sign-up as cc_app does it under the new user's context: the
// user, its counters and a 7-day trial starting at the database's now().
var registerSQL = []string{
	`WITH v(id) AS (SELECT $1::uuid)
	 INSERT INTO users (id, email_hash, email_encrypted, email_key_id, password_hash, name)
	 SELECT id, sha256(convert_to(id::text, 'UTF8')),
	        '\x0001'::bytea || sha256(convert_to(id::text, 'UTF8')) || sha256(convert_to(id::text, 'UTF8')),
	        1, '$argon2id$v=19$m=65536,t=3,p=4$c2FsdHNhbHQ$aGFzaGhhc2hoYXNo', 'Ana'
	 FROM v`,
	"INSERT INTO user_usage (user_id) VALUES ($1)",
	"INSERT INTO entitlements (user_id, kind, source, starts_at, ends_at) VALUES ($1, 'trial', 'registration', now(), now() + interval '168 hours')",
}

const insertTaxpayerSQL = `
INSERT INTO taxpayers (id, user_id, person_type, display_name, rfc_encrypted, rfc_key_id, rfc_hash, rfc_masked,
                       rfc_check_digit_valid)
VALUES ($2::text::uuid, $1, $3, 'Ana (gastos)',
        '\x0001'::bytea || sha256(convert_to($2::text, 'UTF8')) || sha256(convert_to($2::text, 'UTF8')),
        1, sha256(convert_to($2::text || '/rfc', 'UTF8')),
        CASE $3 WHEN 'fisica' THEN 'GOMA******AB6' ELSE 'GOM******AB6' END, true)`

const insertRegimeSQL = `
INSERT INTO taxpayer_regimes (user_id, taxpayer_id, regime_code, person_type, options, start_period, end_period)
VALUES ($1, $2, $3, $4, $5::jsonb, $6, $7)`

// regimeSeed is one regime window of a seeded taxpayer.
type regimeSeed struct {
	code, options, start string
	end                  any // nil = open
}

// taxpayerSeeds are the fixtures' taxpayers and their regimes.
var taxpayerSeeds = []struct {
	user, id, personType string
	regimes              []regimeSeed
}{
	{userA, taxpayerA, "fisica", []regimeSeed{
		{"612", "{}", "2026-01", nil},
		{"606", `{"deduction_mode": "blind_35"}`, "2026-01", nil},
	}},
	{userA, taxpayerR, "fisica", []regimeSeed{{"626", "{}", "2026-01", nil}}},
	{userA, taxpayerM, "moral", []regimeSeed{{"601", "{}", "2026-01", nil}}},
	{userA, taxpayerOld, "fisica", []regimeSeed{{"612", "{}", "2022-01", nil}}},
	{userB, taxpayerB, "fisica", []regimeSeed{{"612", "{}", "2026-01", nil}}},
}

type env struct {
	db   *pgtest.Database
	app  *sql.DB
	txm  *pgdb.TxManager
	logs *bytes.Buffer
	uc   *usecase.Expenses
}

type envOption func(*usecase.Deps)

// readOnlyAccess decides access a month after the wall clock: the trial rows
// start at the database's now(), so every trial is over by then.
func readOnlyAccess(txm *pgdb.TxManager) envOption {
	return func(d *usecase.Deps) {
		later := clock.NewFixed(time.Now().Add(30 * 24 * time.Hour))
		d.Access = access.NewPolicy(txm, accessgorm.NewStore(), accessgorm.NewStore(), later)
	}
}

// newEnv is a fresh database with both users, their taxpayers and the use
// cases wired over it.
func newEnv(t *testing.T, options ...envOption) *env {
	t.Helper()
	db := pgtest.NewDatabase(t)
	e := &env{db: db, app: db.Connect(t, pgtest.RoleApp), logs: &bytes.Buffer{}}
	var err error
	e.txm, err = pgdb.NewTxManager(e.app)
	require.NoError(t, err)
	for _, user := range []string{userA, userB} {
		for _, statement := range registerSQL {
			execAs(t, e.app, user, statement, user)
		}
	}
	for _, tp := range taxpayerSeeds {
		execAs(t, e.app, tp.user, insertTaxpayerSQL, tp.user, tp.id, tp.personType)
		for _, r := range tp.regimes {
			execAs(t, e.app, tp.user, insertRegimeSQL, tp.user, tp.id, r.code, tp.personType, r.options, r.start, r.end)
		}
	}
	return e.wire(t, options...)
}

// wire (re)builds the use cases; options override deps.
func (e *env) wire(t *testing.T, options ...envOption) *env {
	t.Helper()
	sets, err := params.Embedded()
	require.NoError(t, err)
	catalog, err := ratecatalog.New[domain.AmountKey](sets)
	require.NoError(t, err)
	queries := tpusecase.NewQueries(tpusecase.Deps{
		Tx: e.txm, Taxpayers: tpinfra.NewStore(), Regimes: tpinfra.NewStore(), Catalog: tpinfra.NewCatalog(),
		Directory: tpinfra.NewCatalog(), Records: tpinfra.NoRecords{}, Clock: clock.NewFixed(today),
	})
	deps := usecase.Deps{
		Tx:         e.txm,
		Access:     access.NewPolicy(e.txm, accessgorm.NewStore(), accessgorm.NewStore(), clock.NewSystem()),
		Regimes:    queries,
		Taxpayers:  profileAdapter{queries: queries},
		Expenses:   infra.NewStore(),
		Categories: infra.NewCategoryCatalog(),
		Parameters: parametersAdapter{Catalog: catalog},
		Clock:      clock.NewFixed(today),
		Logger:     slog.New(slog.NewJSONHandler(e.logs, &slog.HandlerOptions{Level: slog.LevelDebug})),
	}
	for _, o := range options {
		o(&deps)
	}
	e.uc = usecase.NewExpenses(deps)
	return e
}

// profileAdapter is the composition root's adapter of the taxpayers
// profile reader to the expenses port (M4-07 moves it to cmd/api).
type profileAdapter struct{ queries *tpusecase.Queries }

func (a profileAdapter) Profile(ctx context.Context, userID, taxpayerID string) (usecase.TaxpayerProfile, error) {
	p, err := a.queries.Profile(ctx, userID, taxpayerID)
	if err != nil {
		return usecase.TaxpayerProfile{}, err
	}
	out := usecase.TaxpayerProfile{PersonType: fiscal.PersonType(p.PersonType), Archived: p.Archived}
	for _, r := range p.Regimes {
		hold := usecase.RegimeHold{Code: fiscal.RegimeCode(r.Code), StartPeriod: r.Window.Start.String()}
		if !r.Window.End.IsZero() {
			hold.EndPeriod = r.Window.End.String()
		}
		if mode, ok := r.Options["deduction_mode"].Text(); ok {
			hold.Options.DeductionMode = domain.DeductionMode(mode)
		}
		if definitive, ok := r.Options["definitive_option"].Bool(); ok {
			hold.Options.DefinitiveOption = definitive
		}
		out.Regimes = append(out.Regimes, hold)
	}
	return out, nil
}

// parametersAdapter is the composition root's adapter of the tax engine's
// RateCatalog to the expenses Parameters port (M4-07 moves it to cmd/api).
type parametersAdapter struct {
	*ratecatalog.Catalog[domain.AmountKey]
}

func (a parametersAdapter) On(on fiscal.Date) (usecase.DatedParameters, error) {
	limit, err := a.CashPaymentLimit(on)
	if errors.Is(err, ratecatalog.ErrDateNotCovered) {
		return usecase.DatedParameters{}, usecase.ErrParametersUnavailable
	}
	if err != nil {
		return usecase.DatedParameters{}, err
	}
	rates, err := a.AllowedIVARates(on)
	if err != nil {
		return usecase.DatedParameters{}, err
	}
	return usecase.DatedParameters{
		IVARatesAllowed: rates.Value, CashPaymentLimit: limit.Value, Provisional: limit.Provisional || rates.Provisional,
	}, nil
}

// execAs runs one statement under userID's RLS context.
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

// analyze refreshes the planner's statistics on expenses, as the owner.
func analyze(t *testing.T, e *env) {
	t.Helper()
	_, err := e.db.Connect(t, pgtest.RoleOwner).ExecContext(t.Context(), "ANALYZE expenses")
	require.NoError(t, err)
}

// expensesCreated is the user's trial counter.
func expensesCreated(t *testing.T, app *sql.DB, userID string) int {
	t.Helper()
	var n int
	queryAs(t, app, userID, "SELECT expenses_created FROM user_usage WHERE user_id = $1", []any{userID}, &n)
	return n
}

func date(t *testing.T, raw string) fiscal.Date {
	t.Helper()
	d, err := fiscal.ParseDate(raw)
	require.NoError(t, err)
	return d
}
