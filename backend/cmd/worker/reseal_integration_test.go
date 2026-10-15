//go:build integration

package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"database/sql"
	"database/sql/driver"
	"encoding/base64"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	authdomain "github.com/romodevx/control-contable/backend/internal/auth/domain"
	authinfra "github.com/romodevx/control-contable/backend/internal/auth/infra"
	authusecase "github.com/romodevx/control-contable/backend/internal/auth/usecase"
	"github.com/romodevx/control-contable/backend/internal/platform/config"
	"github.com/romodevx/control-contable/backend/internal/platform/crypto"
	"github.com/romodevx/control-contable/backend/internal/platform/logging"
	"github.com/romodevx/control-contable/backend/internal/platform/pgdb"
	remdomain "github.com/romodevx/control-contable/backend/internal/reminders/domain"
	"github.com/romodevx/control-contable/backend/internal/reminders/infra/store"
	remusecase "github.com/romodevx/control-contable/backend/internal/reminders/usecase"
	tpdomain "github.com/romodevx/control-contable/backend/internal/taxpayers/domain"
	tpinfra "github.com/romodevx/control-contable/backend/internal/taxpayers/infra"
	tpusecase "github.com/romodevx/control-contable/backend/internal/taxpayers/usecase"
)

const (
	resealUser     = "0199d4a2-0000-7000-8000-000000000e01"
	resealTaxpayer = "0199d4a3-0000-7000-8000-000000000e11"
	resealDevice   = "0199d4a4-0000-7000-8000-000000000e21"
	resealEmail    = "ana.reseal@example.com"
	resealRFC      = "PEXJ900215HK8"

	insertSealedTaxpayerSQL = `
INSERT INTO taxpayers (id, user_id, person_type, display_name, rfc_encrypted, rfc_key_id, rfc_hash, rfc_masked,
                       rfc_check_digit_valid)
VALUES ($1, $2, 'fisica', 'Ana', $3, $4, $5, $6, $7)`
)

// bytesArg passes a []byte as one bytea parameter (GORM expands a bare one).
type bytesArg []byte

func (b bytesArg) Value() (driver.Value, error) { return []byte(b), nil }

// rotationKeys are the old (id 1) and new (id 2) keys of all three columns.
type rotationKeys struct {
	email, rfc, token [2][]byte
	rfcPepper         []byte
	reminders         config.Reminders // old push token key, id 1
}

func newRotationKeys(t *testing.T) rotationKeys {
	t.Helper()
	random := func() []byte {
		b := make([]byte, crypto.KeySize)
		_, err := rand.Read(b)
		require.NoError(t, err)
		return b
	}
	k := rotationKeys{
		email: [2][]byte{random(), random()}, rfc: [2][]byte{random(), random()}, token: [2][]byte{random(), random()},
		rfcPepper: random(),
	}
	k.reminders = testReminderKeys(t)
	k.reminders.PushTokenKey = config.NewSecret(string(k.token[0]))
	return k
}

func onlyKey(t *testing.T, id crypto.KeyID, key []byte) *crypto.Keyring {
	t.Helper()
	kr, err := crypto.NewRotatingKeyring(crypto.KeyVersion{ID: id, Key: key})
	require.NoError(t, err)
	return kr
}

// seedSealedUser writes a user, a taxpayer and a device the way the API
// does, every value sealed with the old keys by the modules' own code.
func seedSealedUser(t *testing.T, app *sql.DB, k rotationKeys) {
	t.Helper()
	units, err := pgdb.NewTxManager(app)
	require.NoError(t, err)
	userID, err := authdomain.ParseUserID(resealUser)
	require.NoError(t, err)
	email, err := authdomain.NewEmail(resealEmail)
	require.NoError(t, err)
	index, err := crypto.NewRFCIndex(k.rfcPepper)
	require.NoError(t, err)
	rfc, err := tpdomain.NewRFCFor(tpdomain.PersonTypeFisica, resealRFC)
	require.NoError(t, err)
	binding := tpusecase.RowBinding{TaxpayerID: resealTaxpayer, UserID: resealUser}
	sealedRFC, err := tpinfra.NewRFCVault(onlyKey(t, 1, k.rfc[0]), index).Seal(rfc, binding)
	require.NoError(t, err)
	vault, err := newTokenVault(k.reminders)
	require.NoError(t, err)
	token, err := remdomain.NewPushToken("fcm:" + resealDevice)
	require.NoError(t, err)
	sealedToken, err := vault.Seal(token, remusecase.TokenBinding{InstallationID: resealDevice, UserID: resealUser})
	require.NoError(t, err)

	require.NoError(t, units.ReadWrite(t.Context(), resealUser, func(ctx context.Context) error {
		if _, err := authinfra.NewAccountStore(onlyKey(t, 1, k.email[0])).Create(ctx, authusecase.NewAccount{
			ID: userID, Email: email, EmailHash: bytes.Repeat([]byte{0xe1}, crypto.HashSize),
			PasswordHash: randomVerifier(t), VerifiedAt: time.Now(),
		}); err != nil {
			return err
		}
		db, err := pgdb.Conn(ctx)
		if err != nil {
			return err
		}
		if err := db.Exec(insertSealedTaxpayerSQL, resealTaxpayer, resealUser, bytesArg(sealedRFC.Ciphertext), sealedRFC.KeyID,
			bytesArg(sealedRFC.Hash), sealedRFC.Masked, sealedRFC.CheckDigitValid).Error; err != nil {
			return err
		}
		_, err = store.NewDevices().Claim(ctx, resealUser, remusecase.SealedDevice{
			InstallationID: resealDevice, Platform: remdomain.PlatformAndroid, Token: sealedToken,
		})
		return err
	}))
}

// ADR-0012 end to end: the key_reseal job as newJobs wires it moves the
// email, the RFC and the push token, each bound to its row by its own
// module, to the new keys; the new keys alone then open all three.
func TestKeyReseal_asWiredOverPostgres(t *testing.T) {
	app, workerDB, _ := reminderDatabases(t)
	k := newRotationKeys(t)
	seedSealedUser(t, app, k)
	during := k.reminders
	during.PushTokenKey, during.PushTokenKeyID = config.NewSecret(string(k.token[1])), 2
	during.PushTokenPrevious = config.PreviousKey{Key: config.NewSecret(string(k.token[0])), KeyID: 1}
	sealingKey := func(keys [2][]byte) config.SealingKey {
		return config.SealingKey{
			Key: config.NewSecret(string(keys[1])), KeyID: 2,
			Previous: config.PreviousKey{Key: config.NewSecret(string(keys[0])), KeyID: 1},
		}
	}
	var logs bytes.Buffer
	jobs, err := newJobs(jobWiring{
		pool: workerDB, logger: logging.New(&logs, nil), push: remusecase.DisabledPushSender{},
		reminderKeys: during, outbox: testOutboxKeys(t), env: config.EnvDev, now: reminderDay,
		resealKeys: config.ResealKeys{RFC: sealingKey(k.rfc), Email: sealingKey(k.email)},
	})
	require.NoError(t, err)
	reseal := jobs[len(jobs)-1].job
	require.Equal(t, "key_reseal", reseal.Name())

	require.NoError(t, reseal.Run(t.Context()))

	assert.NotContains(t, logs.String(), `"unreadable":1`)
	appUnits, err := pgdb.NewTxManager(app)
	require.NoError(t, err)
	workerUnits, err := pgdb.NewTxManager(workerDB)
	require.NoError(t, err)
	userID, err := authdomain.ParseUserID(resealUser)
	require.NoError(t, err)
	require.NoError(t, appUnits.ReadWrite(t.Context(), resealUser, func(ctx context.Context) error {
		account, err := authinfra.NewAccountStore(onlyKey(t, 2, k.email[1])).Find(ctx, userID)
		require.NoError(t, err)
		assert.Equal(t, resealEmail, account.Email.String())
		db, err := pgdb.Conn(ctx)
		require.NoError(t, err)
		var ciphertext []byte
		require.NoError(t, db.Raw("SELECT rfc_encrypted FROM taxpayers WHERE id = ?", resealTaxpayer).Row().Scan(&ciphertext))
		index, err := crypto.NewRFCIndex(k.rfcPepper)
		require.NoError(t, err)
		rfc, err := tpinfra.NewRFCVault(onlyKey(t, 2, k.rfc[1]), index).Open(ciphertext,
			tpusecase.RowBinding{TaxpayerID: resealTaxpayer, UserID: resealUser})
		require.NoError(t, err)
		assert.Equal(t, resealRFC, rfc.String())
		return nil
	}))
	afterTokens := during
	afterTokens.PushTokenPrevious = config.PreviousKey{}
	vault, err := newTokenVault(afterTokens)
	require.NoError(t, err)
	require.NoError(t, workerUnits.ReadWrite(t.Context(), resealUser, func(ctx context.Context) error {
		db, err := pgdb.Conn(ctx)
		require.NoError(t, err)
		var ciphertext []byte
		require.NoError(t, db.Raw("SELECT token_ciphertext FROM devices WHERE id = ?", resealDevice).Row().Scan(&ciphertext))
		token, err := vault.Open(ciphertext, remusecase.TokenBinding{InstallationID: resealDevice, UserID: resealUser})
		require.NoError(t, err)
		assert.Equal(t, "fcm:"+resealDevice, token.Reveal())
		return nil
	}))
	assert.Contains(t, logs.String(), `"rows_on_old_keys":0`)
}

// randomVerifier is a well-formed argon2id PHC string over random bytes,
// generated at run time so no credential-like literal lives in the repo.
func randomVerifier(t *testing.T) string {
	t.Helper()
	const saltBytes, sumBytes = 16, 32
	b := make([]byte, saltBytes+sumBytes)
	_, err := rand.Read(b)
	require.NoError(t, err)
	enc := base64.RawStdEncoding
	return "$argon2id$v=19$m=65536,t=3,p=4$" + enc.EncodeToString(b[:saltBytes]) + "$" + enc.EncodeToString(b[saltBytes:])
}
