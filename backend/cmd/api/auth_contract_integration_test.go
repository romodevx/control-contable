//go:build integration

package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/adapter/http/contracttest"
	"github.com/romodevx/control-contable/backend/adapter/http/gen"
	"github.com/romodevx/control-contable/backend/internal/platform/flags/flagtest"
)

// Contract tests of M1-05's operations (register, verifyEmail,
// resendVerification) through the real composition root over Postgres as
// cc_app, with module_auth ON and OFF.

const (
	signUpEmail = "ana.gomez@example.mx"
	// Made-up sample phrases, not credentials.
	acceptableSample = "tlacuache montaña 2026"
	commonSample     = "passwordpassword"
	verifiedSent     = `{"status":"verification_sent"}`
)

var loggedCode = regexp.MustCompile(`"verification_code":"([0-9]{6})"`)

func registerJSON(email, phrase, version string) string {
	return `{"email":"` + email + `","password":"` + phrase + `","name":"Ana Gómez",` +
		`"accepted_documents":[{"kind":"privacy_notice_integral","version":"` + version + `"},{"kind":"terms","version":"v1"}]}`
}

func (e contractEnv) post(target, body string) *http.Request {
	req := httptest.NewRequest(http.MethodPost, target, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.RemoteAddr = contractClientAddr
	return req
}

// lastCode is the code the dev email adapter logged last (dev only).
func (e contractEnv) lastCode(t *testing.T) string {
	t.Helper()
	matches := loggedCode.FindAllStringSubmatch(e.logs.String(), -1)
	require.NotEmpty(t, matches, "the dev email adapter logged no code")
	return matches[len(matches)-1][1]
}

func fieldError(t *testing.T, rec *httptest.ResponseRecorder) gen.FieldError {
	t.Helper()
	var body gen.ErrorResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	require.Equal(t, gen.ErrorCodeValidationFailed, body.Error.Code)
	require.NotNil(t, body.Error.Details)
	require.Len(t, *body.Error.Details, 1)
	return (*body.Error.Details)[0]
}

func TestContract_authSignUp(t *testing.T) {
	c := contracttest.Load(t)

	flagtest.OnOff(t, "module_auth", func(t *testing.T, f flagtest.Static, on bool) {
		env := newContractEnv(t, pingOK{})
		env.apply(t, []string{"module_auth"}, f)

		if !on {
			for _, req := range []*http.Request{
				env.post("/v1/auth/register", registerJSON(signUpEmail, acceptableSample, "v1")),
				env.post("/v1/auth/verify-email", `{"email":"`+signUpEmail+`","code":"123456"}`),
				env.post("/v1/auth/verify-email/resend", `{"email":"`+signUpEmail+`"}`),
			} {
				flagtest.AssertHidden(t, c.Serve(t, env.handler, req))
			}
			return
		}

		t.Run("register answers the generic 202", func(t *testing.T) {
			rec := c.ServeOp(t, contracttest.Op("register"), env.handler,
				env.post("/v1/auth/register", registerJSON(signUpEmail, acceptableSample, "v1")))
			require.Equal(t, http.StatusAccepted, rec.Code, rec.Body.String())
			assert.JSONEq(t, verifiedSent, rec.Body.String())
		})

		t.Run("a common password names the field, never the value", func(t *testing.T) {
			rec := c.ServeOp(t, contracttest.Op("register"), env.handler,
				env.post("/v1/auth/register", registerJSON("jose@example.mx", commonSample, "v1")))
			require.Equal(t, http.StatusBadRequest, rec.Code)
			got := fieldError(t, rec)
			assert.Equal(t, "password", got.Field)
			assert.Equal(t, "password_policy", got.Code)
			assert.NotContains(t, rec.Body.String(), commonSample)
		})

		t.Run("a stale document version is refused", func(t *testing.T) {
			rec := c.ServeOp(t, contracttest.Op("register"), env.handler,
				env.post("/v1/auth/register", registerJSON("jose@example.mx", acceptableSample, "v0")))
			require.Equal(t, http.StatusBadRequest, rec.Code)
			got := fieldError(t, rec)
			assert.Equal(t, "accepted_documents", got.Field)
			assert.Equal(t, "document_version_not_current", got.Code)
		})

		t.Run("resend answers the generic 202", func(t *testing.T) {
			for _, email := range []string{signUpEmail, "nobody@example.mx"} {
				rec := c.ServeOp(t, contracttest.Op("resendVerification"), env.handler,
					env.post("/v1/auth/verify-email/resend", `{"email":"`+email+`"}`))
				require.Equal(t, http.StatusAccepted, rec.Code)
				assert.JSONEq(t, verifiedSent, rec.Body.String())
			}
		})

		t.Run("resend validates its body", func(t *testing.T) {
			rec := c.ServeOp(t, contracttest.Op("resendVerification"), env.handler,
				env.post("/v1/auth/verify-email/resend", `{"email":"x"}`))
			assert.Equal(t, http.StatusBadRequest, rec.Code)
		})

		code := env.lastCode(t)
		wrong := "000000"
		if code == wrong {
			wrong = "111111"
		}

		t.Run("a wrong code is the generic 401", func(t *testing.T) {
			rec := c.ServeOp(t, contracttest.Op("verifyEmail"), env.handler,
				env.post("/v1/auth/verify-email", `{"email":"`+signUpEmail+`","code":"`+wrong+`"}`))
			require.Equal(t, http.StatusUnauthorized, rec.Code)
			assert.Contains(t, rec.Body.String(), "The code is invalid or has expired.")
		})

		t.Run("the right code creates the account and a session", func(t *testing.T) {
			rec := c.ServeOp(t, contracttest.Op("verifyEmail"), env.handler,
				env.post("/v1/auth/verify-email", `{"email":"`+signUpEmail+`","code":"`+code+`"}`))
			require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())
			assert.Equal(t, "no-store", rec.Header().Get("Cache-Control"))
			var session gen.AuthSession
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &session))
			assert.Equal(t, signUpEmail, string(session.User.Email))
			require.NotNil(t, session.User.Name)
			assert.Equal(t, "Ana Gómez", *session.User.Name)
			assert.Equal(t, 900, session.Tokens.AccessExpiresIn)
			assert.Equal(t, 2592000, session.Tokens.RefreshExpiresIn)
			assert.True(t, strings.HasPrefix(session.Tokens.RefreshToken, session.User.Id.String()+"."))
		})

		t.Run("register of the verified email answers the same 202", func(t *testing.T) {
			rec := c.ServeOp(t, contracttest.Op("register"), env.handler,
				env.post("/v1/auth/register", registerJSON(signUpEmail, acceptableSample, "v1")))
			require.Equal(t, http.StatusAccepted, rec.Code)
			assert.JSONEq(t, verifiedSent, rec.Body.String())
		})

		t.Run("no log line carries the email address", func(t *testing.T) {
			assert.NotContains(t, env.logs.String(), signUpEmail)
			assert.NotContains(t, env.logs.String(), acceptableSample)
		})
	})
}
