//go:build integration

package taxinputs_test

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
	"github.com/romodevx/control-contable/backend/internal/taxengine/infra/taxinputs"
	"github.com/romodevx/control-contable/backend/internal/taxengine/usecase"
	"github.com/romodevx/control-contable/backend/migrations"
)

func TestMain(m *testing.M) {
	os.Exit(pgtest.Main(m, pgtest.WithMigrator(migrations.Up)))
}

// The tax-inputs store and use cases over the real migrations, as cc_app
// under RLS. Fixture ids are this suite's own; every test gets its own
// cloned database.
const (
	userA = "0199d4a2-0000-7000-8000-0000000005aa"
	userB = "0199d4a2-0000-7000-8000-0000000005bb"

	// fisicaA holds 612 since 2022-01; moralA held 626 from 2022-01 to
	// 2023-12 and holds 601 since 2024-01 (they may not overlap); fisicaB
	// (userB) holds 612.
	fisicaA = "0199d4a2-5555-7000-8000-0000000000a1"
	moralA  = "0199d4a2-5555-7000-8000-0000000000a2"
	fisicaB = "0199d4a2-5555-7000-8000-0000000000b1"
)

// today is the use cases' frozen clock: 14 October 2026 in Mexico City.
var today = time.Date(2026, time.October, 14, 12, 0, 0, 0, clock.Location())

// registerSQL is sign-up as cc_app does it under the new user's context.
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
INSERT INTO taxpayer_regimes (user_id, taxpayer_id, regime_code, person_type, start_period, end_period)
VALUES ($1, $2, $3, $4, $5, $6)`

type env struct {
	app    *sql.DB
	txm    *pgdb.TxManager
	store  taxinputs.Store
	inputs *usecase.Inputs
}

func newEnv(t *testing.T) *env {
	t.Helper()
	db := pgtest.NewDatabase(t)
	e := &env{app: db.Connect(t, pgtest.RoleApp), store: taxinputs.NewStore()}
	var err error
	e.txm, err = pgdb.NewTxManager(e.app)
	require.NoError(t, err)
	e.inputs = usecase.NewInputs(usecase.InputsDeps{
		Tx: e.txm, Access: access.NewPolicy(e.txm, accessgorm.NewStore(), accessgorm.NewStore(), clock.NewSystem()),
		Store: e.store, Clock: clock.NewFixed(today), Logger: slog.New(slog.DiscardHandler),
	})
	for _, u := range []string{userA, userB} {
		e.exec(t, u, registerSQL...)
	}
	e.seedTaxpayer(t, userA, fisicaA, "fisica", regime{"612", "2022-01", nil})
	e.seedTaxpayer(t, userA, moralA, "moral", regime{"626", "2022-01", "2023-12"}, regime{"601", "2024-01", nil})
	e.seedTaxpayer(t, userB, fisicaB, "fisica", regime{"612", "2022-01", nil})
	return e
}

type regime struct {
	code, start string
	end         any
}

func (e *env) seedTaxpayer(t *testing.T, userID, id, personType string, regimes ...regime) {
	t.Helper()
	require.NoError(t, pgtest.AsUser(t.Context(), e.app, userID, func(tx *sql.Tx) error {
		if _, err := tx.ExecContext(t.Context(), insertTaxpayerSQL, userID, id, personType); err != nil {
			return err
		}
		for _, r := range regimes {
			if _, err := tx.ExecContext(t.Context(), insertRegimeSQL, userID, id, r.code, personType, r.start, r.end); err != nil {
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

// archive archives taxpayer id of userID.
func (e *env) archive(t *testing.T, userID, id string) {
	t.Helper()
	require.NoError(t, pgtest.AsUser(t.Context(), e.app, userID, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(t.Context(), "UPDATE taxpayers SET archived_at = now() WHERE id = $1", id)
		return err
	}))
}
