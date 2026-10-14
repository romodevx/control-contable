//go:build integration

package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/adapter/http/contracttest"
	"github.com/romodevx/control-contable/backend/internal/auth/domain"
	"github.com/romodevx/control-contable/backend/internal/platform/config"
	"github.com/romodevx/control-contable/backend/internal/platform/flags/flagtest"
	"github.com/romodevx/control-contable/backend/internal/platform/logging"
	"github.com/romodevx/control-contable/backend/internal/platform/pgtest"
)

// Rate limits of MODULE-auth §5 (M1-10) through the real composition root
// over Postgres: per client per operation (router), per email, per token
// family and per account (use cases). Every 429 goes through ServeOp, so the
// contract's TooManyRequests (Retry-After required) is checked each time.

// fromClient sets the request's client address to the n-th address of a
// documentation range (TRUSTED_PROXY_HOPS=0: the TCP peer is the client).
func fromClient(req *http.Request, n int) *http.Request {
	req.RemoteAddr = fmt.Sprintf("198.51.100.%d:4321", n%250+1)
	return req
}

// assertRetryAfter checks a 429: rate_limited and a Retry-After of at least
// 1 second and at most max.
func assertRetryAfter(t *testing.T, rec *httptest.ResponseRecorder, maxWait time.Duration) {
	t.Helper()
	require.Equal(t, http.StatusTooManyRequests, rec.Code, rec.Body.String())
	code, message := errorOf(t, rec)
	assert.Equal(t, "rate_limited", string(code))
	assert.Equal(t, "Too many requests. Try again later.", message)
	seconds, err := strconv.Atoi(rec.Header().Get("Retry-After"))
	require.NoError(t, err, "Retry-After is whole seconds")
	assert.GreaterOrEqual(t, seconds, 1)
	assert.LessOrEqual(t, seconds, int(maxWait/time.Second))
}

func newLimitsEnv(t *testing.T) contractEnv {
	t.Helper()
	env := newContractEnv(t, pingOK{})
	env.apply(t, []string{"module_auth"}, flagtest.Static{"module_auth": true})
	return env
}

func TestContract_authLoginLimits(t *testing.T) {
	c := contracttest.Load(t)

	// US-A2a + §5: 5 failures per email, known or unknown alike, then 429
	// (even with the right password); the 429 is the same for both.
	t.Run("5 failures per email, then 429 whether or not an account holds it", func(t *testing.T) {
		env := newLimitsEnv(t)
		env.signUp(t, signUpEmail, acceptableSample)
		var answers []string
		for _, email := range []string{signUpEmail, "nadie@example.mx"} {
			for i := range 5 {
				rec := env.do(fromClient(env.post("/v1/auth/login", loginJSON(email, "otra frase equivocada")), i))
				require.Equalf(t, http.StatusUnauthorized, rec.Code, "failure %d of %s", i+1, email)
			}
			rec := c.ServeOp(t, contracttest.Op("login"), env.handler,
				fromClient(env.post("/v1/auth/login", loginJSON(email, acceptableSample)), 9))
			assertRetryAfter(t, rec, 3*time.Minute)
			code, message := errorOf(t, rec)
			answers = append(answers, string(code)+" "+message+" "+rec.Header().Get("Content-Type"))
		}
		assert.Equal(t, answers[0], answers[1], "the 429 does not enumerate accounts")
		assert.NotContains(t, env.logs.String(), signUpEmail)
	})

	t.Run("a successful login refills the email's budget", func(t *testing.T) {
		env := newLimitsEnv(t)
		env.signUp(t, signUpEmail, acceptableSample)
		for i := range 4 {
			require.Equal(t, http.StatusUnauthorized, env.do(fromClient(env.post("/v1/auth/login", loginJSON(signUpEmail, "mal")), i)).Code)
		}
		require.Equal(t, http.StatusOK, env.do(fromClient(env.post("/v1/auth/login", loginJSON(signUpEmail, acceptableSample)), 5)).Code)

		for i := range 5 {
			assert.Equal(t, http.StatusUnauthorized, env.do(fromClient(env.post("/v1/auth/login", loginJSON(signUpEmail, "mal")), 10+i)).Code)
		}
		assertRetryAfter(t, env.do(fromClient(env.post("/v1/auth/login", loginJSON(signUpEmail, "mal")), 20)), 3*time.Minute)
	})

	// ADR-0005 (M1-17): the failure budget is per (email, client network),
	// so a stranger spending a victim's budget from their network does not
	// lock the victim out of theirs.
	t.Run("another network's failures do not lock the account out", func(t *testing.T) {
		env := newLimitsEnv(t)
		env.signUp(t, signUpEmail, acceptableSample)
		for i := range 10 {
			env.do(fromClient(env.post("/v1/auth/login", loginJSON(signUpEmail, "mal")), i))
		}
		assertRetryAfter(t, env.do(fromClient(env.post("/v1/auth/login", loginJSON(signUpEmail, acceptableSample)), 11)), 3*time.Minute)

		victim := env.post("/v1/auth/login", loginJSON(signUpEmail, acceptableSample))
		victim.RemoteAddr = "203.0.113.5:4321"
		rec := env.do(victim)
		assert.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	})

	// ADR-0005: 50 failures / h per email from all networks together bound
	// distributed guessing (then one per 72 s).
	t.Run("the email's ceiling holds across networks", func(t *testing.T) {
		env := newLimitsEnv(t)
		for n := range 10 {
			for i := range 5 {
				req := env.post("/v1/auth/login", loginJSON("nadie@example.mx", "mal"))
				req.RemoteAddr = fmt.Sprintf("100.64.%d.%d:4321", n, i+1)
				require.Equalf(t, http.StatusUnauthorized, env.do(req).Code, "network %d failure %d", n, i+1)
			}
		}
		req := env.post("/v1/auth/login", loginJSON("nadie@example.mx", "mal"))
		req.RemoteAddr = "100.64.99.1:4321"
		assertRetryAfter(t, c.ServeOp(t, contracttest.Op("login"), env.handler, req), 72*time.Second)
	})

	// ADR-0005: the budget lives in Postgres, so a second API instance sees
	// the failures the first one counted.
	t.Run("the email's budget is shared by every API instance", func(t *testing.T) {
		env := newLimitsEnv(t)
		auth, err := newAuthModule(authWiring{
			env: config.EnvDev, keys: testAuthKeys(), pool: env.app, logger: logging.New(&syncBuffer{}, nil),
			now: time.Now, argon2: fastArgon2,
		})
		require.NoError(t, err)
		other := newHandlerWith(t, contractConfig(), pingOK{}, env.flags, auth, logging.New(&syncBuffer{}, nil))

		for i := range 5 {
			h := env.handler
			if i%2 == 1 {
				h = other
			}
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, fromClient(env.post("/v1/auth/login", loginJSON("nadie@example.mx", "mal")), i))
			require.Equal(t, http.StatusUnauthorized, rec.Code)
		}
		rec := httptest.NewRecorder()
		other.ServeHTTP(rec, fromClient(env.post("/v1/auth/login", loginJSON("nadie@example.mx", "mal")), 7))
		assertRetryAfter(t, rec, 3*time.Minute)
	})

	// §5: 20 / 15 min per client, before validation or any credential work.
	t.Run("20 per client, then 429 even for a malformed body", func(t *testing.T) {
		env := newLimitsEnv(t)
		for i := range 20 {
			rec := env.do(fromClient(env.post("/v1/auth/login", loginJSON(fmt.Sprintf("persona%d@example.mx", i), "mal")), 1))
			require.Equalf(t, http.StatusUnauthorized, rec.Code, "attempt %d", i+1)
		}
		assertRetryAfter(t, c.ServeOp(t, contracttest.Op("login"), env.handler,
			fromClient(env.post("/v1/auth/login", `{"email":1}`), 1)), 15*time.Minute)
		assert.Equal(t, http.StatusUnauthorized,
			env.do(fromClient(env.post("/v1/auth/login", loginJSON("otra@example.mx", "mal")), 2)).Code, "another client is not affected")
	})
}

func TestContract_authSignUpLimits(t *testing.T) {
	c := contracttest.Load(t)

	// §5 + D7: register and resend share 5 / h per email; register then
	// answers 429, resend stays the silent 202 and sends nothing.
	t.Run("5 sign-ups per email, then register is 429 and resend is silent", func(t *testing.T) {
		env := newLimitsEnv(t)
		for i := range 5 {
			rec := env.do(fromClient(env.post("/v1/auth/register", registerJSON(signUpEmail, "v1")), i))
			require.Equalf(t, http.StatusAccepted, rec.Code, "register %d", i+1)
		}
		assertRetryAfter(t, c.ServeOp(t, contracttest.Op("register"), env.handler,
			fromClient(env.post("/v1/auth/register", registerJSON(signUpEmail, "v1")), 6)), time.Hour)

		codes := strings.Count(env.logs.String(), `"verification_code"`)
		rec := c.ServeOp(t, contracttest.Op("resendVerification"), env.handler,
			fromClient(env.post("/v1/auth/verify-email/resend", `{"email":"`+signUpEmail+`"}`), 7))
		assert.Equal(t, http.StatusAccepted, rec.Code)
		assert.JSONEq(t, verifiedSent, rec.Body.String())
		assert.Equal(t, codes, strings.Count(env.logs.String(), `"verification_code"`), "nothing sent")
	})

	t.Run("register: 10 per client, then 429", func(t *testing.T) {
		env := newLimitsEnv(t)
		for i := range 10 {
			rec := env.do(fromClient(env.post("/v1/auth/register", registerJSON(fmt.Sprintf("persona%d@example.mx", i), "v1")), 1))
			require.Equalf(t, http.StatusAccepted, rec.Code, "register %d", i+1)
		}
		assertRetryAfter(t, c.ServeOp(t, contracttest.Op("register"), env.handler,
			fromClient(env.post("/v1/auth/register", registerJSON("otra@example.mx", "v1")), 1)), time.Hour)
	})

	t.Run("resendVerification: 10 per client, then 429", func(t *testing.T) {
		env := newLimitsEnv(t)
		for i := range 10 {
			rec := env.do(fromClient(env.post("/v1/auth/verify-email/resend", fmt.Sprintf(`{"email":"persona%d@example.mx"}`, i)), 1))
			require.Equalf(t, http.StatusAccepted, rec.Code, "resend %d", i+1)
		}
		assertRetryAfter(t, c.ServeOp(t, contracttest.Op("resendVerification"), env.handler,
			fromClient(env.post("/v1/auth/verify-email/resend", `{"email":"otra@example.mx"}`), 1)), time.Hour)
	})

	t.Run("verifyEmail: 20 per client, then 429", func(t *testing.T) {
		env := newLimitsEnv(t)
		for i := range 20 {
			rec := env.do(fromClient(env.post("/v1/auth/verify-email", verifyJSON(fmt.Sprintf("persona%d@example.mx", i), "123456", acceptableSample)), 1))
			require.Equalf(t, http.StatusUnauthorized, rec.Code, "verify %d", i+1)
		}
		assertRetryAfter(t, c.ServeOp(t, contracttest.Op("verifyEmail"), env.handler,
			fromClient(env.post("/v1/auth/verify-email", verifyJSON(signUpEmail, "123456", acceptableSample)), 1)), 15*time.Minute)
	})
}

func TestContract_authSessionLimits(t *testing.T) {
	c := contracttest.Load(t)

	t.Run("refreshSession: 60 per client per minute, then 429", func(t *testing.T) {
		env := newLimitsEnv(t)
		for i := range 60 {
			rec := env.do(fromClient(env.post("/v1/auth/refresh", refreshJSON(fmt.Sprintf("%s.junk%d", sampleID, i))), 1))
			require.Equalf(t, http.StatusUnauthorized, rec.Code, "refresh %d", i+1)
		}
		assertRetryAfter(t, c.ServeOp(t, contracttest.Op("refreshSession"), env.handler,
			fromClient(env.post("/v1/auth/refresh", refreshJSON(sampleID+".junk")), 1)), time.Minute)
	})

	// §5: 20 rotations / h per family. Out of budget the token is not
	// rotated: it stays live for the retry after Retry-After.
	t.Run("20 rotations per family, then 429 and the token stays live", func(t *testing.T) {
		env := newLimitsEnv(t)
		token := env.signUp(t, signUpEmail, acceptableSample).Tokens.RefreshToken
		for i := range 20 {
			rec := env.do(fromClient(env.post("/v1/auth/refresh", refreshJSON(token)), i))
			require.Equalf(t, http.StatusOK, rec.Code, "rotation %d: %s", i+1, rec.Body.String())
			token = tokenPairOf(t, rec.Body.Bytes()).RefreshToken
		}
		assertRetryAfter(t, c.ServeOp(t, contracttest.Op("refreshSession"), env.handler,
			fromClient(env.post("/v1/auth/refresh", refreshJSON(token)), 30)), 3*time.Minute)

		parsed, err := domain.ParseRefreshToken(token)
		require.NoError(t, err)
		var live bool
		require.NoError(t, pgtest.AsUser(t.Context(), env.app, parsed.UserID().String(), func(tx *sql.Tx) error {
			return tx.QueryRowContext(t.Context(),
				"SELECT rotated_at IS NULL AND revoked_at IS NULL FROM refresh_tokens WHERE token_hash = $1",
				parsed.Hash().Bytes()).Scan(&live)
		}))
		assert.True(t, live, "a 429 neither rotates nor revokes")
	})

	t.Run("listLegalDocuments: 60 per client per minute, then 429", func(t *testing.T) {
		env := newLimitsEnv(t)
		for i := range 60 {
			require.Equalf(t, http.StatusOK, env.do(fromClient(env.request(http.MethodGet, "/v1/legal/documents"), 1)).Code, "read %d", i+1)
		}
		assertRetryAfter(t, c.ServeOp(t, contracttest.Op("listLegalDocuments"), env.handler,
			fromClient(env.request(http.MethodGet, "/v1/legal/documents"), 1)), time.Minute)
	})
}
