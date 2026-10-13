//go:build integration

package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/adapter/http/gen"
	"github.com/romodevx/control-contable/backend/internal/platform/crypto"
	"github.com/romodevx/control-contable/backend/internal/platform/flags/flagtest"
	"github.com/romodevx/control-contable/backend/internal/platform/pgtest"
)

// TestAuth_preRegistrationTakeover is the regression test of the
// pre-registration account takeover (MODULE-auth D8): an attacker starts a
// sign-up for the victim's address and tries to plant a password of the
// attacker's choosing; the victim then registers too (D7 makes that a
// resend), reads the code from their own inbox and verifies with their own
// password. The account must end up with the VICTIM's password: whoever
// proves inbox ownership sets it. Before D8 register took the password and
// this test failed (the stored hash verified the attacker's phrase).
func TestAuth_preRegistrationTakeover(t *testing.T) {
	const (
		victim = "victima@example.mx"
		// Made-up sample phrases, not credentials.
		attackerPhrase = "atacante conoce esta frase"
		victimPhrase   = "solo la victima la conoce"
	)
	env := newContractEnv(t, pingOK{})
	env.apply(t, []string{"module_auth"}, flagtest.Static{"module_auth": true})

	// The attacker cannot put a password into a sign-up at all.
	planted := `{"email":"` + victim + `","password":"` + attackerPhrase + `",` +
		`"accepted_documents":[{"kind":"privacy_notice_integral","version":"v1"},{"kind":"terms","version":"v1"}]}`
	rec := env.do(env.post("/v1/auth/register", planted))
	require.Equal(t, http.StatusBadRequest, rec.Code, "register refuses a password (D8)")
	assert.NotContains(t, rec.Body.String(), attackerPhrase)

	// Attacker, then victim, register the victim's address the only way left.
	for range 2 {
		rec = env.do(env.post("/v1/auth/register", registerJSON(victim, "v1")))
		require.Equal(t, http.StatusAccepted, rec.Code, rec.Body.String())
	}
	code := env.lastCode(t) // the victim's inbox

	rec = env.do(env.post("/v1/auth/verify-email", verifyJSON(victim, code, victimPhrase)))
	require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())
	var session gen.AuthSession
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &session))

	stored := env.passwordHash(t, session.User.Id.String())
	hasher, err := crypto.NewPasswordHasher(fastArgon2)
	require.NoError(t, err)
	attackerWins, err := hasher.Verify(attackerPhrase, stored)
	require.NoError(t, err)
	victimWins, err := hasher.Verify(victimPhrase, stored)
	require.NoError(t, err)
	assert.False(t, attackerWins, "the attacker's pre-registered password opens the victim's account")
	assert.True(t, victimWins, "the victim's password does not open the account they verified")
}

func (e contractEnv) do(req *http.Request) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	e.handler.ServeHTTP(rec, req)
	return rec
}

// passwordHash reads the account's stored password hash as that user (RLS).
func (e contractEnv) passwordHash(t *testing.T, userID string) string {
	t.Helper()
	var hash string
	require.NoError(t, pgtest.AsUser(t.Context(), e.app, userID, func(tx *sql.Tx) error {
		return tx.QueryRowContext(t.Context(), "SELECT password_hash FROM users WHERE id = $1", userID).Scan(&hash)
	}))
	return hash
}
