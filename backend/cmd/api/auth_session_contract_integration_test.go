//go:build integration

package main

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/adapter/http/contracttest"
	"github.com/romodevx/control-contable/backend/adapter/http/gen"
	"github.com/romodevx/control-contable/backend/internal/platform/flags/flagtest"
)

// Contract tests of M1-08's operations (refreshSession, logout) through the
// real composition root over Postgres as cc_app, with module_auth ON and OFF.

const sessionEnded = "Your session has ended. Sign in again."

func refreshJSON(token string) string { return `{"refresh_token":"` + token + `"}` }

// bearerPost is a POST carrying an access token.
func (e contractEnv) bearerPost(target, body, accessToken string) *http.Request {
	req := e.post(target, body)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	return req
}

func tokenPairOf(t *testing.T, body []byte) gen.TokenPair {
	t.Helper()
	var pair gen.TokenPair
	require.NoError(t, json.Unmarshal(body, &pair))
	return pair
}

func TestContract_authSession(t *testing.T) {
	c := contracttest.Load(t)

	flagtest.OnOff(t, "module_auth", func(t *testing.T, f flagtest.Static, on bool) {
		env := newContractEnv(t, pingOK{})
		env.apply(t, []string{"module_auth"}, f)

		if !on {
			for _, req := range []*http.Request{
				env.post("/v1/auth/refresh", refreshJSON("x")),
				env.bearerPost("/v1/auth/logout", refreshJSON("x"), "x"),
			} {
				flagtest.AssertHidden(t, c.Serve(t, env.handler, req))
			}
			return
		}
		session := env.signUp(t, signUpEmail, acceptableSample)
		first := session.Tokens.RefreshToken
		var rotated gen.TokenPair

		t.Run("a live token rotates", func(t *testing.T) {
			rec := c.ServeOp(t, contracttest.Op("refreshSession"), env.handler, env.post("/v1/auth/refresh", refreshJSON(first)))
			require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
			assert.Equal(t, "no-store", rec.Header().Get("Cache-Control"))
			rotated = tokenPairOf(t, rec.Body.Bytes())
			assert.NotEqual(t, first, rotated.RefreshToken)
			assert.Equal(t, 900, rotated.AccessExpiresIn)
			assert.Equal(t, 2592000, rotated.RefreshExpiresIn)
			assert.Equal(t, session.User.Id.String(), bearerUser(t, rotated.AccessToken))
		})

		t.Run("reuse of the rotated token is the generic 401 and ends the family", func(t *testing.T) {
			rec := c.ServeOp(t, contracttest.Op("refreshSession"), env.handler, env.post("/v1/auth/refresh", refreshJSON(first)))
			require.Equal(t, http.StatusUnauthorized, rec.Code)
			code, message := errorOf(t, rec)
			assert.Equal(t, gen.ErrorCodeUnauthorized, code)
			assert.Equal(t, sessionEnded, message)

			rec = c.ServeOp(t, contracttest.Op("refreshSession"), env.handler, env.post("/v1/auth/refresh", refreshJSON(rotated.RefreshToken)))
			require.Equal(t, http.StatusUnauthorized, rec.Code, "the successor was revoked with its family")
			_, message = errorOf(t, rec)
			assert.Equal(t, sessionEnded, message)
		})

		t.Run("a malformed or unknown token is the same 401", func(t *testing.T) {
			for _, token := range []string{"garbage", session.User.Id.String() + ".AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"} {
				rec := c.ServeOp(t, contracttest.Op("refreshSession"), env.handler, env.post("/v1/auth/refresh", refreshJSON(token)))
				require.Equal(t, http.StatusUnauthorized, rec.Code, token)
				_, message := errorOf(t, rec)
				assert.Equal(t, sessionEnded, message)
			}
		})

		t.Run("a body without the token is a 400", func(t *testing.T) {
			rec := c.ServeOp(t, contracttest.Op("refreshSession"), env.handler, env.post("/v1/auth/refresh", `{}`))
			assert.Equal(t, http.StatusBadRequest, rec.Code)
		})

		login := func(t *testing.T) gen.AuthSession {
			t.Helper()
			rec := env.do(env.post("/v1/auth/login", loginJSON(signUpEmail, acceptableSample)))
			require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
			var s gen.AuthSession
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &s))
			return s
		}

		t.Run("logout revokes the family, then refresh fails", func(t *testing.T) {
			s := login(t)
			rec := c.ServeOp(t, contracttest.Op("logout"), env.handler,
				env.bearerPost("/v1/auth/logout", refreshJSON(s.Tokens.RefreshToken), s.Tokens.AccessToken))
			require.Equal(t, http.StatusNoContent, rec.Code, rec.Body.String())
			assert.Empty(t, rec.Body.String())

			rec = env.do(env.post("/v1/auth/refresh", refreshJSON(s.Tokens.RefreshToken)))
			assert.Equal(t, http.StatusUnauthorized, rec.Code)

			rec = c.ServeOp(t, contracttest.Op("logout"), env.handler,
				env.bearerPost("/v1/auth/logout", refreshJSON(s.Tokens.RefreshToken), s.Tokens.AccessToken))
			assert.Equal(t, http.StatusNoContent, rec.Code, "idempotent")
		})

		t.Run("logout with another user's token is 204 and leaves it alive", func(t *testing.T) {
			victim := login(t)
			other := env.signUp(t, "jose@example.mx", acceptableSample)
			rec := c.ServeOp(t, contracttest.Op("logout"), env.handler,
				env.bearerPost("/v1/auth/logout", refreshJSON(victim.Tokens.RefreshToken), other.Tokens.AccessToken))
			require.Equal(t, http.StatusNoContent, rec.Code)

			rec = env.do(env.post("/v1/auth/refresh", refreshJSON(victim.Tokens.RefreshToken)))
			assert.Equal(t, http.StatusOK, rec.Code, "the victim's session still works")
		})

		t.Run("logout without a valid bearer is 401", func(t *testing.T) {
			s := login(t)
			for _, access := range []string{"", "not-a-jwt"} {
				req := env.post("/v1/auth/logout", refreshJSON(s.Tokens.RefreshToken))
				if access != "" {
					req.Header.Set("Authorization", "Bearer "+access)
				}
				rec := c.ServeOp(t, contracttest.Op("logout"), env.handler, req)
				require.Equal(t, http.StatusUnauthorized, rec.Code)
			}
			rec := env.do(env.post("/v1/auth/refresh", refreshJSON(s.Tokens.RefreshToken)))
			assert.Equal(t, http.StatusOK, rec.Code, "nothing was revoked")
		})

		t.Run("logs audit the reuse and never carry a token", func(t *testing.T) {
			logs := env.logs.String()
			assert.Contains(t, logs, `"event":"token_reuse"`)
			assert.Contains(t, logs, `"event":"logout"`)
			assert.NotContains(t, logs, first)
			assert.NotContains(t, logs, rotated.RefreshToken)
		})
	})
}
