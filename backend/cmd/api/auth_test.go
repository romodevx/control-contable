package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/auth/domain"
	"github.com/romodevx/control-contable/backend/internal/platform/config"
	"github.com/romodevx/control-contable/backend/internal/platform/flags/flagtest"
	"github.com/romodevx/control-contable/backend/internal/platform/logging"
	"github.com/romodevx/control-contable/backend/internal/platform/principal"
)

const sessionUser = "0199d4a2-7c1e-7b3a-9f10-2c5e8a1b4d01"

var signedAt = time.Date(2026, 10, 13, 18, 30, 0, 0, time.UTC)

func fixtureKey(b byte) config.Secret { return config.NewSecret(strings.Repeat(string(rune(b)), 32)) }

// signWith signs an access token for sessionUser as the composition root
// does with keys.
func signWith(t *testing.T, keys config.Auth) string {
	t.Helper()
	c, err := newAuthCrypto(authWiring{keys: keys, argon2: fastArgon2})
	require.NoError(t, err)
	userID, err := domain.ParseUserID(sessionUser)
	require.NoError(t, err)
	token, _, err := c.access.Sign(userID, signedAt)
	require.NoError(t, err)
	return token
}

// authenticates reports whether the authenticator wired from keys admits
// token, and checks the principal it sets.
func authenticates(t *testing.T, keys config.Auth, token string) bool {
	t.Helper()
	authn, err := newAuthenticator(keys, func() time.Time { return signedAt.Add(time.Minute) }, nil)
	require.NoError(t, err)
	var reached bool
	h := authn.Authenticate(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reached = true
		userID, err := principal.UserID(r.Context())
		assert.NoError(t, err)
		assert.Equal(t, sessionUser, userID, "the user id tx.Manager will run as")
		w.WriteHeader(http.StatusNoContent)
	}))
	req := httptest.NewRequest(http.MethodGet, "/v1/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if !reached {
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	}
	return reached
}

// rotated is keys after a rotation: a new active key k2, the old k1 kept to
// verify the tokens it already signed.
func rotated(keys config.Auth) config.Auth {
	keys.PreviousAccessTokenKey, keys.PreviousAccessTokenKeyID = keys.AccessTokenKey, keys.AccessTokenKeyID
	keys.AccessTokenKey, keys.AccessTokenKeyID = fixtureKey(0x55), "k2"
	return keys
}

func TestAuthenticator_admitsTheTokensTheLoginSigns(t *testing.T) {
	keys := testAuthKeys()
	assert.True(t, authenticates(t, keys, signWith(t, keys)))
}

// Kid rotation: tokens signed before the rotation keep working until they
// expire; new ones are signed by the new key; once the previous key is
// dropped, its tokens stop working.
func TestAuthenticator_keyRotation(t *testing.T) {
	before := testAuthKeys()
	after := rotated(before)
	oldToken := signWith(t, before)
	newToken := signWith(t, after)

	assert.True(t, authenticates(t, after, oldToken), "signed by the previous key")
	assert.True(t, authenticates(t, after, newToken), "signed by the active key")
	assert.False(t, authenticates(t, before, newToken), "an instance without k2 refuses it")

	dropped := after
	dropped.PreviousAccessTokenKey, dropped.PreviousAccessTokenKeyID = config.Secret{}, ""
	assert.False(t, authenticates(t, dropped, oldToken), "the retired key verifies nothing any more")
}

func TestNewHandler_refusesToServeWithoutAnAuthenticator(t *testing.T) {
	module := stubAuthModule(t)
	module.authn = nil

	_, err := newHandler(testConfig(), deps{db: pingOK{}, flags: flagtest.Static{}, auth: module, now: time.Now,
		logger: logging.New(io.Discard, nil)})

	assert.ErrorContains(t, err, "no authenticator")
}

func TestAuthenticator_refusesAnotherDeploymentsKeyUnderAKnownKid(t *testing.T) {
	stranger := testAuthKeys()
	stranger.AccessTokenKey = fixtureKey(0x66) // same kid "k1", other key

	assert.False(t, authenticates(t, testAuthKeys(), signWith(t, stranger)))
}
