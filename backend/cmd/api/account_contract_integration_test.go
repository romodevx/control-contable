//go:build integration

package main

import (
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/adapter/http/contracttest"
	"github.com/romodevx/control-contable/backend/adapter/http/gen"
	"github.com/romodevx/control-contable/backend/internal/platform/flags/flagtest"
	"github.com/romodevx/control-contable/backend/internal/platform/pgtest"
)

// Contract tests of M1-09's operations (getMe, updateMe, deleteMe,
// listLegalDocuments) through the real composition root over Postgres as
// cc_app, with module_auth ON and OFF.

const (
	beaEmail = "bea.lopez@example.mx"
	// A second made-up sample phrase, not a credential.
	otherSample = "ajolote del lago 1987"
)

// send builds a request with an optional JSON body and bearer token.
func (e contractEnv) send(method, target, body, token string) *http.Request {
	var reader io.Reader
	if body != "" {
		reader = strings.NewReader(body)
	}
	req := httptest.NewRequest(method, target, reader)
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	req.RemoteAddr = contractClientAddr
	return req
}

func decodeUser(t *testing.T, rec *httptest.ResponseRecorder) gen.User {
	t.Helper()
	var user gen.User
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &user))
	return user
}

// ownedRows counts userID's rows in each user-owned table, as that user.
func (e contractEnv) ownedRows(t *testing.T, userID string) map[string]int {
	t.Helper()
	counts := map[string]int{}
	require.NoError(t, pgtest.AsUser(t.Context(), e.app, userID, func(tx *sql.Tx) error {
		for table, column := range map[string]string{
			"users": "id", "user_usage": "user_id", "entitlements": "user_id", "consents": "user_id", "refresh_tokens": "user_id",
		} {
			var n int
			query := "SELECT count(*) FROM " + table + " WHERE " + column + " = $1" // #nosec G202 -- fixed test identifiers
			if err := tx.QueryRowContext(t.Context(), query, userID).Scan(&n); err != nil {
				return err
			}
			counts[table] = n
		}
		return nil
	}))
	return counts
}

func TestContract_authAccount(t *testing.T) {
	c := contracttest.Load(t)

	flagtest.OnOff(t, "module_auth", func(t *testing.T, f flagtest.Static, on bool) {
		env := newContractEnv(t, pingOK{})
		env.apply(t, []string{"module_auth"}, f)

		if !on {
			for _, req := range []*http.Request{
				env.send(http.MethodGet, "/v1/legal/documents", "", ""),
				env.send(http.MethodGet, "/v1/me", "", "any-token"),
				env.send(http.MethodPatch, "/v1/me", `{"name":"Ana"}`, "any-token"),
				env.send(http.MethodDelete, "/v1/me", `{"password":"x"}`, "any-token"),
			} {
				flagtest.AssertHidden(t, c.Serve(t, env.handler, req))
			}
			return
		}

		t.Run("legal documents: the current version of each kind, public", func(t *testing.T) {
			rec := c.ServeOp(t, contracttest.Op("listLegalDocuments"), env.handler,
				env.send(http.MethodGet, "/v1/legal/documents", "", ""))
			require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
			var list gen.LegalDocumentList
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &list))
			require.Len(t, list.Items, 3)
			for _, d := range list.Items {
				assert.Equal(t, "v1", d.Version, d.Kind)
				assert.True(t, strings.HasPrefix(d.Url, "https://"), d.Url)
			}
		})

		ana := env.signUp(t, signUpEmail, acceptableSample)
		bea := env.signUp(t, beaEmail, otherSample)
		anaToken, beaToken := ana.Tokens.AccessToken, bea.Tokens.AccessToken

		t.Run("getMe answers the caller's own profile, no-store", func(t *testing.T) {
			rec := c.ServeOp(t, contracttest.Op("getMe"), env.handler, env.send(http.MethodGet, "/v1/me", "", anaToken))
			require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
			assert.Equal(t, "no-store", rec.Header().Get("Cache-Control"))
			assert.Equal(t, ana.User, decodeUser(t, rec))
		})

		t.Run("getMe without a bearer is the generic 401", func(t *testing.T) {
			rec := c.ServeOp(t, contracttest.Op("getMe"), env.handler, env.send(http.MethodGet, "/v1/me", "", ""))
			require.Equal(t, http.StatusUnauthorized, rec.Code)
			assert.Equal(t, realmChallenge, rec.Header().Get("WWW-Authenticate"))
		})

		t.Run("updateMe renames only the caller", func(t *testing.T) {
			rec := c.ServeOp(t, contracttest.Op("updateMe"), env.handler,
				env.send(http.MethodPatch, "/v1/me", `{"name":"  Bea López Ruiz "}`, beaToken))
			require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
			assert.Equal(t, "no-store", rec.Header().Get("Cache-Control"))
			got := decodeUser(t, rec)
			require.NotNil(t, got.Name)
			assert.Equal(t, "Bea López Ruiz", *got.Name)
			assert.Equal(t, beaEmail, string(got.Email))

			rec = env.do(env.send(http.MethodGet, "/v1/me", "", anaToken))
			assert.Equal(t, ana.User, decodeUser(t, rec), "Ana is untouched")
		})

		for name, body := range map[string]string{
			"an empty body":       `{}`,
			"an explicit null":    `{"name":null}`,
			"an unknown property": `{"name":"Ana","email":"otra@example.mx"}`,
			"a too long name":     `{"name":"` + strings.Repeat("a", 121) + `"}`,
		} {
			t.Run("updateMe refuses "+name, func(t *testing.T) {
				rec := c.ServeOp(t, contracttest.Op("updateMe"), env.handler, env.send(http.MethodPatch, "/v1/me", body, anaToken))
				assert.Equal(t, http.StatusBadRequest, rec.Code, rec.Body.String())
			})
		}

		t.Run("updateMe refuses a blank name on the name field", func(t *testing.T) {
			rec := c.ServeOp(t, contracttest.Op("updateMe"), env.handler, env.send(http.MethodPatch, "/v1/me", `{"name":"   "}`, anaToken))
			require.Equal(t, http.StatusBadRequest, rec.Code)
			assert.Equal(t, "name", fieldError(t, rec).Field)
		})

		// Not through ServeOp: the authenticator's 401 carries the bearer
		// challenge (RFC 9110 requires it on every 401), but the frozen
		// deleteMe 401 is inline and declares no WWW-Authenticate header,
		// unlike the shared Unauthorized response. A contract fix, not a
		// server one (reported in the M1-09 PR).
		t.Run("deleteMe without a bearer is a 401 before the body is read", func(t *testing.T) {
			rec := env.do(env.send(http.MethodDelete, "/v1/me", `{}`, ""))
			require.Equal(t, http.StatusUnauthorized, rec.Code)
			assert.Equal(t, realmChallenge, rec.Header().Get("WWW-Authenticate"))
			code, _ := errorOf(t, rec)
			assert.Equal(t, gen.ErrorCodeUnauthorized, code)
		})

		t.Run("deleteMe without the password is a 400 on password and deletes nothing", func(t *testing.T) {
			for _, body := range []string{`{}`, `{"password":""}`} {
				rec := c.ServeOp(t, contracttest.Op("deleteMe"), env.handler, env.send(http.MethodDelete, "/v1/me", body, anaToken))
				require.Equal(t, http.StatusBadRequest, rec.Code, body)
				assert.Equal(t, "password", fieldError(t, rec).Field, body)
			}
			assert.Equal(t, 1, env.ownedRows(t, ana.User.Id.String())["users"])
		})

		t.Run("deleteMe with a wrong password is a 401 and deletes nothing", func(t *testing.T) {
			rec := c.ServeOp(t, contracttest.Op("deleteMe"), env.handler,
				env.send(http.MethodDelete, "/v1/me", `{"password":"`+otherSample+`"}`, anaToken))
			require.Equal(t, http.StatusUnauthorized, rec.Code)
			code, message := errorOf(t, rec)
			assert.Equal(t, gen.ErrorCodeUnauthorized, code)
			assert.Equal(t, "The password is incorrect.", message)
			assert.Equal(t, 1, env.ownedRows(t, ana.User.Id.String())["users"])
		})

		t.Run("deleteMe with the right password deletes everything the user owns", func(t *testing.T) {
			beaBefore := env.ownedRows(t, bea.User.Id.String())

			rec := c.ServeOp(t, contracttest.Op("deleteMe"), env.handler,
				env.send(http.MethodDelete, "/v1/me", `{"password":"`+acceptableSample+`"}`, anaToken))

			require.Equal(t, http.StatusNoContent, rec.Code, rec.Body.String())
			assert.Empty(t, rec.Body.String())
			for table, n := range env.ownedRows(t, ana.User.Id.String()) {
				assert.Zerof(t, n, "%s still holds rows of the deleted user", table)
			}
			assert.Equal(t, beaBefore, env.ownedRows(t, bea.User.Id.String()), "Bea is untouched")
		})

		t.Run("the deleted account's access token is refused everywhere", func(t *testing.T) {
			rec := c.ServeOp(t, contracttest.Op("getMe"), env.handler, env.send(http.MethodGet, "/v1/me", "", anaToken))
			require.Equal(t, http.StatusUnauthorized, rec.Code)
			assert.Equal(t, realmChallenge, rec.Header().Get("WWW-Authenticate"))

			rec = c.ServeOp(t, contracttest.Op("updateMe"), env.handler, env.send(http.MethodPatch, "/v1/me", `{"name":"Ana"}`, anaToken))
			assert.Equal(t, http.StatusUnauthorized, rec.Code)

			rec = c.ServeOp(t, contracttest.Op("deleteMe"), env.handler,
				env.send(http.MethodDelete, "/v1/me", `{"password":"`+acceptableSample+`"}`, anaToken))
			assert.Equal(t, http.StatusUnauthorized, rec.Code)
		})

		// Sessions are revoked by the cascade, in the delete's own
		// transaction: every refresh_tokens row of the user is gone, so no
		// family can be refreshed (M1-08's RevokeFamily is not needed).
		t.Run("the deleted account's refresh token is refused", func(t *testing.T) {
			rec := c.ServeOp(t, contracttest.Op("refreshSession"), env.handler,
				env.post("/v1/auth/refresh", `{"refresh_token":"`+ana.Tokens.RefreshToken+`"}`))
			require.Equal(t, http.StatusUnauthorized, rec.Code, rec.Body.String())
		})

		t.Run("the deleted account can no longer sign in", func(t *testing.T) {
			rec := env.do(env.post("/v1/auth/login", loginJSON(signUpEmail, acceptableSample)))
			assert.Equal(t, http.StatusUnauthorized, rec.Code, "the account is gone")
		})

		t.Run("registering the same email again gets a new account with no trial", func(t *testing.T) {
			again := env.signUp(t, signUpEmail, otherSample)
			assert.NotEqual(t, ana.User.Id, again.User.Id)
			owned := env.ownedRows(t, again.User.Id.String())
			assert.Equal(t, 1, owned["users"])
			assert.Zero(t, owned["entitlements"], "trial_registry survived the deletion: no second trial")
		})

		t.Run("no log line carries an email or a password", func(t *testing.T) {
			logs := env.logs.String()
			for _, secret := range []string{signUpEmail, beaEmail, acceptableSample, otherSample} {
				assert.NotContains(t, logs, secret)
			}
			assert.Contains(t, logs, `"event":"account_deleted"`)
		})
	})
}

// realmChallenge is the contract's WWWAuthenticate const.
const realmChallenge = `Bearer realm="control-contable"`
