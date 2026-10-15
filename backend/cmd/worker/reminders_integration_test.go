//go:build integration

package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"database/sql"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/platform/clock"
	"github.com/romodevx/control-contable/backend/internal/platform/pgdb"
	"github.com/romodevx/control-contable/backend/internal/reminders/domain"
	"github.com/romodevx/control-contable/backend/internal/reminders/infra/store"
	"github.com/romodevx/control-contable/backend/internal/reminders/usecase"
	"github.com/romodevx/control-contable/backend/internal/reminders/usecase/pushtest"
)

// This suite's own fixture ids.
const (
	reminderUser     = "0199d4a2-0000-7000-8000-000000000c06"
	reminderTaxpayer = "0199d4a3-0000-7000-8000-000000000c61"
	reminderDevice   = "0199d4a4-0000-7000-8000-000000000c71"
)

const (
	seedUserSQL = `INSERT INTO users (id, email_hash, email_encrypted, email_key_id, password_hash, name)
VALUES ($1::text::uuid, sha256(convert_to($1::text, 'UTF8')),
        '\x0001'::bytea || sha256(convert_to($1::text, 'UTF8')) || sha256(convert_to($1::text, 'UTF8')), 1,
        '$argon2id$v=19$m=8,t=1,p=1$' || rtrim(encode(sha256(convert_to($1::text, 'UTF8')), 'base64'), '=') || '$' ||
          rtrim(encode(sha256(convert_to($1::text, 'UTF8')), 'base64'), '='), 'Fixture')`
	seedUsageSQL    = "INSERT INTO user_usage (user_id) VALUES ($1)"
	seedTrialSQL    = "INSERT INTO entitlements (user_id, kind, source, starts_at, ends_at) VALUES ($1, 'trial', 'registration', now(), now() + interval '7 days')"
	seedTaxpayerSQL = `INSERT INTO taxpayers (id, user_id, person_type, display_name, rfc_encrypted, rfc_key_id, rfc_hash, rfc_masked,
                       rfc_check_digit_valid)
VALUES ($2::text::uuid, $1, 'fisica', 'Ana', '\x0001'::bytea || sha256(convert_to($2::text, 'UTF8')) || sha256(convert_to($2::text, 'UTF8')),
        1, sha256(convert_to($2::text || '/rfc', 'UTF8')), 'GOMA******AB6', true)`
	seedRegimeSQL   = "INSERT INTO taxpayer_regimes (user_id, taxpayer_id, regime_code, person_type, start_period) VALUES ($1, $2, '612', 'fisica', '2026-01')"
	seedSettingsSQL = "INSERT INTO reminder_settings (user_id, enabled, early_day, on_due_date, send_time, show_amounts) VALUES ($1, true, 14, true, '10:00', true)"
)

func seedAsUser(t *testing.T, db *sql.DB, user, query string, args ...any) {
	t.Helper()
	tx, err := db.BeginTx(t.Context(), nil)
	require.NoError(t, err)
	defer func() { _ = tx.Rollback() }()
	_, err = tx.ExecContext(t.Context(), "SELECT set_config('app.current_user_id', $1, true)", user)
	require.NoError(t, err)
	_, err = tx.ExecContext(t.Context(), query, args...)
	require.NoError(t, err)
	require.NoError(t, tx.Commit())
}

// The worker's reminder job as cmd/worker wires it, over the real database
// as cc_worker: the flags read, the tokens the API sealed opened with the
// same keys, and the monthly Estimator run inside the user's transaction
// (show_amounts on, every module ON). The estimate must compute as
// cc_worker (its tax-inputs read is the M9-06 grant): no
// "estimate_unavailable".
func TestReminderJob_asWiredOverPostgres(t *testing.T) {
	app, workerDB, owner := reminderDatabases(t)
	_, err := owner.ExecContext(t.Context(), "UPDATE feature_flags SET enabled = true")
	require.NoError(t, err)
	for _, q := range []string{seedUserSQL, seedUsageSQL, seedTrialSQL, seedSettingsSQL} {
		seedAsUser(t, app, reminderUser, q, reminderUser)
	}
	seedAsUser(t, app, reminderUser, seedTaxpayerSQL, reminderUser, reminderTaxpayer)
	seedAsUser(t, app, reminderUser, seedRegimeSQL, reminderUser, reminderTaxpayer)

	keys := testReminderKeys(t)
	vault, err := newTokenVault(keys)
	require.NoError(t, err)
	token, err := domain.NewPushToken("fcm:" + rand.Text() + rand.Text())
	require.NoError(t, err)
	sealed, err := vault.Seal(token, usecase.TokenBinding{InstallationID: reminderDevice, UserID: reminderUser})
	require.NoError(t, err)
	appUnits, err := pgdb.NewTxManager(app)
	require.NoError(t, err)
	require.NoError(t, appUnits.ReadWrite(t.Context(), reminderUser, func(ctx context.Context) error {
		_, err := store.NewDevices().Claim(ctx, reminderUser, usecase.SealedDevice{
			InstallationID: reminderDevice, Platform: domain.PlatformIOS, Token: sealed,
		})
		return err
	}))

	units, err := pgdb.NewTxManager(workerDB)
	require.NoError(t, err)
	var logs bytes.Buffer
	sender := &pushtest.Sender{}
	now := clock.NewFixed(time.Date(2026, time.May, 14, 10, 5, 0, 0, clock.Location()))
	job, err := newReminderJob(workerDB, units, slog.New(slog.NewJSONHandler(&logs, nil)), sender, keys, now)
	require.NoError(t, err)

	require.NoError(t, job.Run(t.Context()))

	sends := sender.Sends()
	require.Len(t, sends, 1, logs.String())
	assert.Equal(t, "Tu pago mensual estimado vence el lunes 18 de mayo", sends[0].Notification.Title)
	assert.Equal(t, "Revisa tu estimado de abril para: Ana.", sends[0].Notification.Body, "no records: no figure")
	assert.NotContains(t, logs.String(), "estimate_unavailable")
	assert.NotContains(t, logs.String(), token.Reveal())
}
