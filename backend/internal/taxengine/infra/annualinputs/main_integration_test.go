//go:build integration

package annualinputs_test

import (
	"database/sql"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/platform/access"
	"github.com/romodevx/control-contable/backend/internal/platform/access/accessgorm"
	"github.com/romodevx/control-contable/backend/internal/platform/clock"
	"github.com/romodevx/control-contable/backend/internal/platform/pgdb"
	"github.com/romodevx/control-contable/backend/internal/platform/pgtest"
	"github.com/romodevx/control-contable/backend/internal/taxengine/infra/annualinputs"
	"github.com/romodevx/control-contable/backend/internal/taxengine/usecase/annual"
	"github.com/romodevx/control-contable/backend/migrations"
)

func TestMain(m *testing.M) {
	os.Exit(pgtest.Main(m, pgtest.WithMigrator(migrations.Up)))
}

// The annual input stores and use cases over the real migrations, as
// cc_app under RLS. Fixture ids are this suite's own; every test gets its
// own cloned database.
const (
	userA = "0199d4a2-0000-7000-8000-0000000006aa"
	userB = "0199d4a2-0000-7000-8000-0000000006bb"
	// userRO registered without an entitlement: read-only.
	userRO = "0199d4a2-0000-7000-8000-0000000006cc"

	// fisicaA holds 612 since 2022-01, 605 since 2026-01 and 625 (definitive
	// option) since 2026-01; moralA holds 601 since 2024-01; fisicaB (userB)
	// and fisicaRO (userRO) hold 612.
	fisicaA  = "0199d4a2-6666-7000-8000-0000000000a1"
	moralA   = "0199d4a2-6666-7000-8000-0000000000a2"
	fisicaB  = "0199d4a2-6666-7000-8000-0000000000b1"
	fisicaRO = "0199d4a2-6666-7000-8000-0000000000c1"
)

// today is the use cases' frozen clock: 14 October 2026 in Mexico City.
var today = time.Date(2026, time.October, 14, 12, 0, 0, 0, clock.Location())

// registerSQL is sign-up as cc_app does it under the new user's context;
// the last statement (the trial) is left out for a read-only user.
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

// insertTaxpayerSQL writes taxpayer $2 of user $1 with person type $3; the
// RFC columns derive from the id like crypto output would.
const insertTaxpayerSQL = `
INSERT INTO taxpayers (id, user_id, person_type, display_name, rfc_encrypted, rfc_key_id, rfc_hash, rfc_masked,
                       rfc_check_digit_valid)
VALUES ($2::text::uuid, $1, $3, 'Fixture',
        '\x0001'::bytea || sha256(convert_to($2::text, 'UTF8')) || sha256(convert_to($2::text, 'UTF8')),
        1, sha256(convert_to($2::text || '/rfc', 'UTF8')),
        CASE $3 WHEN 'fisica' THEN 'GOMA******AB6' ELSE 'GOM******AB6' END, true)`

const insertRegimeSQL = `
INSERT INTO taxpayer_regimes (user_id, taxpayer_id, regime_code, person_type, start_period, end_period, options)
VALUES ($1, $2, $3, $4, $5, $6, $7::jsonb)`

type regime struct {
	code, start string
	end         any
	options     string
}

type env struct {
	app   *sql.DB
	txm   *pgdb.TxManager
	store annualinputs.Store
	pol   *access.Policy
}

func newEnv(t *testing.T) *env {
	t.Helper()
	db := pgtest.NewDatabase(t)
	e := &env{app: db.Connect(t, pgtest.RoleApp), store: annualinputs.NewStore()}
	var err error
	e.txm, err = pgdb.NewTxManager(e.app)
	require.NoError(t, err)
	e.pol = access.NewPolicy(e.txm, accessgorm.NewStore(), accessgorm.NewStore(), clock.NewSystem())
	for _, u := range []string{userA, userB} {
		e.exec(t, u, registerSQL...)
	}
	e.exec(t, userRO, registerSQL[:2]...)
	e.seedTaxpayer(t, userA, fisicaA, "fisica",
		regime{"605", "2026-01", nil, "{}"}, regime{"612", "2022-01", nil, "{}"},
		regime{"625", "2026-01", nil, `{"definitive_option": true}`})
	e.seedTaxpayer(t, userA, moralA, "moral", regime{"601", "2024-01", nil, "{}"})
	e.seedTaxpayer(t, userB, fisicaB, "fisica", regime{"612", "2022-01", nil, "{}"})
	e.seedTaxpayer(t, userRO, fisicaRO, "fisica", regime{"612", "2022-01", nil, "{}"})
	return e
}

func (e *env) seedTaxpayer(t *testing.T, userID, id, personType string, regimes ...regime) {
	t.Helper()
	require.NoError(t, pgtest.AsUser(t.Context(), e.app, userID, func(tx *sql.Tx) error {
		if _, err := tx.ExecContext(t.Context(), insertTaxpayerSQL, userID, id, personType); err != nil {
			return err
		}
		for _, r := range regimes {
			if _, err := tx.ExecContext(t.Context(), insertRegimeSQL, userID, id, r.code, personType, r.start, r.end, r.options); err != nil {
				return err
			}
		}
		return nil
	}))
}

// exec runs statements (each taking $1 = userID) under userID's context.
func (e *env) exec(t *testing.T, userID string, statements ...string) {
	t.Helper()
	require.NoError(t, pgtest.AsUser(t.Context(), e.app, userID, func(tx *sql.Tx) error {
		for _, s := range statements {
			if _, err := tx.ExecContext(t.Context(), s, userID); err != nil {
				return err
			}
		}
		return nil
	}))
}

func (e *env) archive(t *testing.T, userID, id string) {
	t.Helper()
	require.NoError(t, pgtest.AsUser(t.Context(), e.app, userID, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(t.Context(), "UPDATE taxpayers SET archived_at = now() WHERE id = $1", id)
		return err
	}))
}

func (e *env) deductions(now time.Time) *annual.Deductions {
	return annual.NewDeductions(annual.DeductionsDeps{
		Tx: e.txm, Access: e.pol, Taxpayers: e.store, Store: e.store, Clock: clock.NewFixed(now), Logger: slog.New(slog.DiscardHandler),
	})
}

func (e *env) salary(now time.Time) *annual.SalaryStatements {
	return annual.NewSalaryStatements(annual.SalaryDeps{
		Tx: e.txm, Access: e.pol, Taxpayers: e.store, Store: e.store, Clock: clock.NewFixed(now), Logger: slog.New(slog.DiscardHandler),
	})
}

func (e *env) inputs() *annual.Inputs {
	return annual.NewInputs(annual.InputsDeps{
		Tx: e.txm, Access: e.pol, Taxpayers: e.store, Store: e.store, Clock: clock.NewFixed(today), Logger: slog.New(slog.DiscardHandler),
	})
}

func (e *env) filed(now time.Time) *annual.FiledPayments {
	return annual.NewFiledPayments(annual.FiledDeps{
		Tx: e.txm, Access: e.pol, Taxpayers: e.store, Store: e.store, Clock: clock.NewFixed(now), Logger: slog.New(slog.DiscardHandler),
	})
}
