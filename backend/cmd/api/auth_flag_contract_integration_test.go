//go:build integration

package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/adapter/http/contracttest"
	"github.com/romodevx/control-contable/backend/adapter/http/gen"
	"github.com/romodevx/control-contable/backend/internal/platform/flags/flagtest"
)

// Flag suite of M1 (MODULE-auth §7, M1-10): with module_auth OFF every M1
// route, whatever the request (a valid body and bearer, a malformed body,
// no bearer, a flood), answers exactly what an unknown route answers, before
// binding, decoding, authentication or any rate limit; nothing reaches the
// database. With the flag ON the same requests reach the module.

// authProbe is one request to an M1 operation.
type authProbe struct {
	op, name, method, path, body string
	bearer                       bool
}

// authProbes are, for every M1 operation, a request the module would
// accept and requests it would reject for different reasons.
func authProbes(refreshToken string) []authProbe {
	return []authProbe{
		{op: "register", name: "valid", method: http.MethodPost, path: "/v1/auth/register", body: registerJSON("nueva@example.mx", "v1")},
		{op: "register", name: "malformed JSON", method: http.MethodPost, path: "/v1/auth/register", body: "{not json"},
		{op: "verifyEmail", name: "valid", method: http.MethodPost, path: "/v1/auth/verify-email", body: verifyJSON(signUpEmail, "123456", acceptableSample)},
		{op: "verifyEmail", name: "schema-invalid", method: http.MethodPost, path: "/v1/auth/verify-email", body: `{"email":1}`},
		{op: "resendVerification", name: "valid", method: http.MethodPost, path: "/v1/auth/verify-email/resend", body: `{"email":"` + signUpEmail + `"}`},
		{op: "login", name: "valid", method: http.MethodPost, path: "/v1/auth/login", body: loginJSON(signUpEmail, acceptableSample)},
		{op: "login", name: "empty body", method: http.MethodPost, path: "/v1/auth/login"},
		{op: "refreshSession", name: "valid", method: http.MethodPost, path: "/v1/auth/refresh", body: refreshJSON(refreshToken)},
		{op: "logout", name: "valid", method: http.MethodPost, path: "/v1/auth/logout", body: refreshJSON(refreshToken), bearer: true},
		{op: "logout", name: "no bearer", method: http.MethodPost, path: "/v1/auth/logout", body: refreshJSON(refreshToken)},
		{op: "getMe", name: "valid", method: http.MethodGet, path: "/v1/me", bearer: true},
		{op: "getMe", name: "no bearer", method: http.MethodGet, path: "/v1/me"},
		{op: "updateMe", name: "valid", method: http.MethodPatch, path: "/v1/me", body: `{"name":"Ana"}`, bearer: true},
		{op: "deleteMe", name: "wrong password", method: http.MethodDelete, path: "/v1/me", body: `{"password":"otra frase equivocada"}`, bearer: true},
		{op: "deleteMe", name: "no body", method: http.MethodDelete, path: "/v1/me", bearer: true},
		{op: "listLegalDocuments", name: "valid", method: http.MethodGet, path: "/v1/legal/documents"},
	}
}

func (e contractEnv) probe(p authProbe, accessToken string) *http.Request {
	token := ""
	if p.bearer {
		token = accessToken
	}
	return e.send(p.method, p.path, p.body, token)
}

// hiddenAnswer is a 404 without what differs per request (the request id).
type hiddenAnswer struct {
	Status  int
	Headers []string
	Code    string
	Message string
	Fields  []string
}

func hiddenAnswerOf(t *testing.T, rec *httptest.ResponseRecorder) hiddenAnswer {
	t.Helper()
	var headers []string
	for name, values := range rec.Header() {
		if name == "X-Request-Id" {
			continue
		}
		headers = append(headers, name+": "+strings.Join(values, ","))
	}
	slices.Sort(headers)
	var body map[string]map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body), rec.Body.String())
	var fields []string
	for k := range body["error"] {
		fields = append(fields, k)
	}
	slices.Sort(fields)
	code, _ := body["error"]["code"].(string)
	message, _ := body["error"]["message"].(string)
	return hiddenAnswer{Status: rec.Code, Headers: headers, Code: code, Message: message, Fields: fields}
}

func TestContract_authFlagOffIsIndistinguishableFromAnUnknownRoute(t *testing.T) {
	c := contracttest.Load(t)
	env := newContractEnv(t, pingOK{})
	env.apply(t, []string{"module_auth"}, flagtest.Static{"module_auth": true})
	session := env.signUp(t, signUpEmail, acceptableSample)
	probes := authProbes(session.Tokens.RefreshToken)
	require.ElementsMatch(t, authOperationIDs(t), uniqueOps(probes), "every M1 operation is probed")

	flagtest.OnOff(t, "module_auth", func(t *testing.T, f flagtest.Static, on bool) {
		env.apply(t, []string{"module_auth"}, f)
		spentBefore := env.bucketRows(t)

		for _, p := range probes {
			t.Run(p.op+" "+p.name, func(t *testing.T) {
				req := env.probe(p, session.Tokens.AccessToken)
				if on {
					rec := env.do(req)
					assert.NotEqual(t, http.StatusNotFound, rec.Code, "ON: the module answers")
					return
				}
				rec := c.Serve(t, env.handler, req)
				unknown := env.do(env.send(p.method, "/v1/auth/nope", p.body, ""))
				require.Equal(t, http.StatusNotFound, unknown.Code)
				hidden := hiddenAnswerOf(t, rec)
				assert.Equal(t, string(gen.ErrorCodeNotFound), hidden.Code)
				assert.Equal(t, hiddenAnswerOf(t, unknown), hidden, "OFF is an unknown route, header for header")
			})
		}
		if on {
			return
		}

		t.Run("a flood never turns into a 429", func(t *testing.T) {
			for range 30 {
				rec := env.do(fromClient(env.post("/v1/auth/login", loginJSON(signUpEmail, "mal")), 1))
				require.Equal(t, http.StatusNotFound, rec.Code)
			}
		})
		t.Run("nothing reaches the database: no budget is spent", func(t *testing.T) {
			require.NotEmpty(t, spentBefore, "not vacuous: sign-up and the ON run spent budgets")
			assert.Equal(t, spentBefore, env.bucketRows(t))
		})
	})
}

// authOperationIDs are the contract's module_auth operations.
func authOperationIDs(t *testing.T) []string {
	t.Helper()
	var ids []string
	for _, op := range gatedOperations(t) {
		if op.module == "module_auth" {
			ids = append(ids, op.id)
		}
	}
	require.NotEmpty(t, ids)
	return ids
}

func uniqueOps(probes []authProbe) []string {
	var ops []string
	for _, p := range probes {
		if !slices.Contains(ops, p.op) {
			ops = append(ops, p.op)
		}
	}
	return ops
}

// bucketRows is the state of every rate-limit bucket (scope, digest and
// tat), read as cc_app: a snapshot, so a re-spent bucket shows too.
func (e contractEnv) bucketRows(t *testing.T) []string {
	t.Helper()
	rows, err := e.app.QueryContext(t.Context(),
		"SELECT scope || ':' || encode(key_hash, 'hex') || ':' || tat::text FROM rate_limit_buckets ORDER BY 1")
	require.NoError(t, err)
	defer func() { _ = rows.Close() }()
	var out []string
	for rows.Next() {
		var row string
		require.NoError(t, rows.Scan(&row))
		out = append(out, row)
	}
	require.NoError(t, rows.Err())
	return out
}
