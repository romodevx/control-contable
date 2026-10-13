package router_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/adapter/http/gen"
	"github.com/romodevx/control-contable/backend/adapter/http/router"
	"github.com/romodevx/control-contable/backend/internal/platform/apperr"
	"github.com/romodevx/control-contable/backend/internal/platform/httperr"
	"github.com/romodevx/control-contable/backend/internal/platform/logging"
)

const (
	validTaxpayer = "0199d4a2-7c1e-7b3a-9f10-2c5e8a1b4d01"
	incomesURL    = "/v1/taxpayers/" + validTaxpayer + "/incomes?period=2026-09"
	validLogin    = `{"email":"ana@example.mx","password":"x"}`

	validNewPassword = "correct-horse-battery"
	termsAccepted    = `{"kind":"terms","version":"2026-10-01"}`
	privacyAccepted  = `{"kind":"privacy_notice_integral","version":"2026-10-01"}`
)

// fakeAPI implements only the operations the tests probe; any other
// operation would hit the nil embedded interface.
type fakeAPI struct {
	gen.StrictServerInterface
	calls     *[]string
	loginBody *gen.LoginJSONRequestBody
	incomeErr error
}

func (f fakeAPI) GetHealth(context.Context, gen.GetHealthRequestObject) (gen.GetHealthResponseObject, error) {
	*f.calls = append(*f.calls, "getHealth")
	return gen.GetHealth200JSONResponse{Status: gen.HealthStatusStatusOk, Database: gen.HealthStatusDatabaseOk, Version: "t"}, nil
}

func (f fakeAPI) Login(_ context.Context, req gen.LoginRequestObject) (gen.LoginResponseObject, error) {
	*f.calls = append(*f.calls, "login")
	*f.loginBody = *req.Body
	return gen.Login401JSONResponse{}, nil
}

func (f fakeAPI) ListIncomes(context.Context, gen.ListIncomesRequestObject) (gen.ListIncomesResponseObject, error) {
	*f.calls = append(*f.calls, "listIncomes")
	return nil, f.incomeErr
}

type flagSet map[string]bool

func (f flagSet) Enabled(_ context.Context, module string) bool { return f[module] }

// recordingAuthn records every request it sees and rejects them when deny.
type recordingAuthn struct {
	deny  bool
	paths *[]string
}

func (a recordingAuthn) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*a.paths = append(*a.paths, r.URL.Path)
		if a.deny {
			_ = httperr.Write(w, r, apperr.Mapper{}, apperr.New(apperr.CodeUnauthorized, "Authentication required."))
			return
		}
		next.ServeHTTP(w, r)
	})
}

type harness struct {
	handler   http.Handler
	calls     []string
	authnSeen []string
	loginBody gen.LoginJSONRequestBody
	logs      bytes.Buffer
}

type setup struct {
	flags     flagSet
	denyAuthn bool
	incomeErr error
	maxBody   int64
}

func newHarness(t testing.TB, s setup) *harness {
	t.Helper()
	spec, err := gen.GetSpec()
	require.NoError(t, err)
	h := &harness{}
	api := fakeAPI{calls: &h.calls, loginBody: &h.loginBody, incomeErr: s.incomeErr}
	h.handler, err = router.New(api, spec, router.Options{
		Flags:        s.flags,
		Authn:        recordingAuthn{deny: s.denyAuthn, paths: &h.authnSeen},
		Logger:       logging.New(&h.logs, nil),
		MaxBodyBytes: s.maxBody,
	})
	require.NoError(t, err)
	return h
}

func (h *harness) do(method, target, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, target, strings.NewReader(body))
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	rec := httptest.NewRecorder()
	h.handler.ServeHTTP(rec, req)
	return rec
}

func envelope(t *testing.T, rec *httptest.ResponseRecorder) httperr.Body {
	t.Helper()
	require.Equal(t, "application/json", rec.Header().Get("Content-Type"), rec.Body.String())
	var env httperr.Response
	dec := json.NewDecoder(rec.Body)
	dec.DisallowUnknownFields()
	require.NoError(t, dec.Decode(&env))
	require.True(t, env.Error.Code.Valid())
	return env.Error
}

func TestFlagGate_disabledModuleIs404BeforeAnythingElse(t *testing.T) {
	tests := []struct {
		name, method, target, body string
	}{
		{name: "valid request", method: http.MethodPost, target: "/v1/auth/login", body: validLogin},
		{name: "malformed JSON body", method: http.MethodPost, target: "/v1/auth/login", body: "{not json"},
		{name: "schema-invalid body", method: http.MethodPost, target: "/v1/auth/login", body: `{"email":1}`},
		{name: "malformed path parameter", method: http.MethodGet, target: "/v1/taxpayers/not-a-uuid/incomes?period=2026-09"},
		{name: "missing required query", method: http.MethodGet, target: "/v1/taxpayers/" + validTaxpayer + "/incomes"},
		{name: "secured operation without credentials", method: http.MethodGet, target: incomesURL},
		{name: "other method on a gated path", method: http.MethodPut, target: "/v1/auth/login"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHarness(t, setup{flags: flagSet{}, denyAuthn: true})

			rec := h.do(tt.method, tt.target, tt.body)

			assert.Equal(t, http.StatusNotFound, rec.Code)
			assert.Equal(t, apperr.CodeNotFound, envelope(t, rec).Code)
			assert.Empty(t, h.calls, "the handler never runs")
			assert.Empty(t, h.authnSeen, "authentication never runs for a disabled module")
		})
	}
}

func TestNotFound_isIndistinguishableFromADisabledModule(t *testing.T) {
	h := newHarness(t, setup{flags: flagSet{}})
	disabled := h.do(http.MethodGet, incomesURL, "")
	missing := h.do(http.MethodGet, "/v1/nope", "")
	wrongMethodOnPlatform := h.do(http.MethodDelete, "/healthz", "")

	for _, rec := range []*httptest.ResponseRecorder{disabled, missing, wrongMethodOnPlatform} {
		assert.Equal(t, http.StatusNotFound, rec.Code)
		assert.Empty(t, rec.Header().Get("Allow"), "no Allow header reveals sibling methods")
		body := envelope(t, rec)
		assert.Equal(t, apperr.CodeNotFound, body.Code)
		assert.Equal(t, "Not found.", body.Message)
	}
}

func TestPlatformOperationsAreNeverGated(t *testing.T) {
	h := newHarness(t, setup{flags: flagSet{}, denyAuthn: true})

	rec := h.do(http.MethodGet, "/healthz", "")

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, []string{"getHealth"}, h.calls)
	assert.Empty(t, h.authnSeen, "getHealth declares security: [] so authentication is skipped")
}

func TestAuthn_runsForSecuredOperationsBeforeValidation(t *testing.T) {
	h := newHarness(t, setup{flags: flagSet{"module_incomes": true}, denyAuthn: true})

	rec := h.do(http.MethodGet, "/v1/taxpayers/not-a-uuid/incomes", "")

	assert.Equal(t, http.StatusUnauthorized, rec.Code, "an anonymous caller learns nothing about the input schema")
	assert.Equal(t, []string{"/v1/taxpayers/not-a-uuid/incomes"}, h.authnSeen)
	assert.Empty(t, h.calls)
}

func TestAuthn_isSkippedForPublicOperations(t *testing.T) {
	h := newHarness(t, setup{flags: flagSet{"module_auth": true}, denyAuthn: true})

	rec := h.do(http.MethodPost, "/v1/auth/login", validLogin)

	assert.Equal(t, http.StatusUnauthorized, rec.Code, "the fake login answers 401 itself")
	assert.Equal(t, []string{"login"}, h.calls)
	assert.Empty(t, h.authnSeen)
}

func TestValidation_rejectsInputThatBreaksTheContract(t *testing.T) {
	tests := []struct {
		name, method, target, body string
		flags                      flagSet
		wantFields                 []string
		wantCodes                  []string
	}{
		{name: "path parameter not a uuid", method: http.MethodGet, target: "/v1/taxpayers/not-a-uuid/incomes?period=2026-09",
			flags: flagSet{"module_incomes": true}, wantFields: []string{"taxpayerId"}, wantCodes: []string{"format"}},
		{name: "required query missing", method: http.MethodGet, target: "/v1/taxpayers/" + validTaxpayer + "/incomes",
			flags: flagSet{"module_incomes": true}, wantFields: []string{"period"}, wantCodes: []string{"required"}},
		{name: "query does not match its pattern", method: http.MethodGet, target: "/v1/taxpayers/" + validTaxpayer + "/incomes?period=2026-13",
			flags: flagSet{"module_incomes": true}, wantFields: []string{"period"}, wantCodes: []string{"pattern"}},
		{name: "required body property missing", method: http.MethodPost, target: "/v1/auth/login", body: `{"email":"ana@example.mx"}`,
			flags: flagSet{"module_auth": true}, wantFields: []string{"password"}, wantCodes: []string{"required"}},
		{name: "body property of the wrong type", method: http.MethodPost, target: "/v1/auth/login", body: `{"email":"ana@example.mx","password":7}`,
			flags: flagSet{"module_auth": true}, wantFields: []string{"password"}, wantCodes: []string{"type"}},
		{name: "body property too short", method: http.MethodPost, target: "/v1/auth/login", body: `{"email":"ana@example.mx","password":""}`,
			flags: flagSet{"module_auth": true}, wantFields: []string{"password"}, wantCodes: []string{"minLength"}},
		{name: "body format checked by the generated decoder", method: http.MethodPost, target: "/v1/auth/login", body: `{"email":"not-an-email","password":"x"}`,
			flags: flagSet{"module_auth": true}, wantFields: []string{"body"}, wantCodes: []string{"format"}},
		{name: "malformed JSON", method: http.MethodPost, target: "/v1/auth/login", body: "{not json",
			flags: flagSet{"module_auth": true}, wantFields: []string{"body"}, wantCodes: []string{"invalid"}},
		{name: "required body missing", method: http.MethodPost, target: "/v1/auth/login",
			flags: flagSet{"module_auth": true}, wantFields: []string{"body"}, wantCodes: []string{"required"}},
		// RegisterRequest reaches DocumentAcceptance and LegalDocumentKind through
		// $ref; the field must still be named, never collapsed into body/invalid.
		{name: "$ref body: required property missing", method: http.MethodPost, target: "/v1/auth/register",
			body:  `{"email":"ana@example.mx","accepted_documents":[` + termsAccepted + `,` + privacyAccepted + `]}`,
			flags: flagSet{"module_auth": true}, wantFields: []string{"password"}, wantCodes: []string{"required"}},
		{name: "$ref body: property too short", method: http.MethodPost, target: "/v1/auth/register",
			body:  `{"email":"ana@example.mx","password":"short","accepted_documents":[` + termsAccepted + `,` + privacyAccepted + `]}`,
			flags: flagSet{"module_auth": true}, wantFields: []string{"password"}, wantCodes: []string{"minLength"}},
		{name: "$ref body: array too short", method: http.MethodPost, target: "/v1/auth/register",
			body:  `{"email":"ana@example.mx","password":"` + validNewPassword + `","accepted_documents":[` + termsAccepted + `]}`,
			flags: flagSet{"module_auth": true}, wantFields: []string{"accepted_documents"}, wantCodes: []string{"minItems"}},
		{name: "$ref body: enum inside a referenced item", method: http.MethodPost, target: "/v1/auth/register",
			body:  `{"email":"ana@example.mx","password":"` + validNewPassword + `","accepted_documents":[` + termsAccepted + `,{"kind":"nope","version":"1"}]}`,
			flags: flagSet{"module_auth": true}, wantFields: []string{"accepted_documents.1.kind"}, wantCodes: []string{"enum"}},
		{name: "$ref body: property of the wrong type", method: http.MethodPost, target: "/v1/auth/register",
			body:  `{"email":"ana@example.mx","password":7,"accepted_documents":[` + termsAccepted + `,` + privacyAccepted + `]}`,
			flags: flagSet{"module_auth": true}, wantFields: []string{"password"}, wantCodes: []string{"type"}},
		{name: "$ref body: unknown property", method: http.MethodPost, target: "/v1/auth/register",
			body:  `{"email":"ana@example.mx","password":"` + validNewPassword + `","accepted_documents":[` + termsAccepted + `,` + privacyAccepted + `],"role":"admin"}`,
			flags: flagSet{"module_auth": true}, wantFields: []string{"role"}, wantCodes: []string{"additionalProperties"}},
		{name: "$ref body: several errors keep their own fields", method: http.MethodPost, target: "/v1/auth/register",
			body:  `{"email":"ana@example.mx","accepted_documents":[` + termsAccepted + `]}`,
			flags: flagSet{"module_auth": true}, wantFields: []string{"accepted_documents", "password"}, wantCodes: []string{"minItems", "required"}},
		{name: "deleteMe without password is a 400 on the field", method: http.MethodDelete, target: "/v1/me", body: `{}`,
			flags: flagSet{"module_auth": true}, wantFields: []string{"password"}, wantCodes: []string{"required"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHarness(t, setup{flags: tt.flags})

			rec := h.do(tt.method, tt.target, tt.body)

			require.Equal(t, http.StatusBadRequest, rec.Code, rec.Body.String())
			body := envelope(t, rec)
			assert.Equal(t, apperr.CodeValidationFailed, body.Code)
			var fields, codes []string
			for _, d := range body.Details {
				fields = append(fields, d.Field)
				codes = append(codes, d.Code)
				assert.NotEmpty(t, d.Message)
			}
			assert.Equal(t, tt.wantFields, fields)
			assert.Equal(t, tt.wantCodes, codes)
			assert.Empty(t, h.calls, "the handler never sees invalid input")
		})
	}
}

func TestValidation_neverEchoesTheRejectedValue(t *testing.T) {
	h := newHarness(t, setup{flags: flagSet{"module_auth": true}})

	rec := h.do(http.MethodPost, "/v1/auth/login", `{"email":"AAAA010101AAA-not-an-email","password":"x"}`)

	require.Equal(t, http.StatusBadRequest, rec.Code)
	assert.NotContains(t, rec.Body.String(), "AAAA010101AAA")
}

func TestValidation_rejectsOversizedBodies(t *testing.T) {
	h := newHarness(t, setup{flags: flagSet{"module_auth": true}, maxBody: 16})

	rec := h.do(http.MethodPost, "/v1/auth/login", validLogin)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	body := envelope(t, rec)
	assert.Equal(t, apperr.CodeValidationFailed, body.Code)
	require.Len(t, body.Details, 1)
	assert.Equal(t, "too_large", body.Details[0].Code)
	assert.Empty(t, h.calls)
}

func TestValidation_handlerStillReadsTheBody(t *testing.T) {
	h := newHarness(t, setup{flags: flagSet{"module_auth": true}})

	h.do(http.MethodPost, "/v1/auth/login", validLogin)

	require.Equal(t, []string{"login"}, h.calls)
	assert.Equal(t, "ana@example.mx", string(h.loginBody.Email))
	assert.Equal(t, "x", h.loginBody.Password)
}

func TestHandlerErrors_useTheEnvelope(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantCode   apperr.Code
	}{
		{name: "typed application error", err: apperr.New(apperr.CodeNotFound, "Not found."),
			wantStatus: http.StatusNotFound, wantCode: apperr.CodeNotFound},
		{name: "unexpected error is a generic 500", err: assert.AnError,
			wantStatus: http.StatusInternalServerError, wantCode: apperr.CodeInternalError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHarness(t, setup{flags: flagSet{"module_incomes": true}, incomeErr: tt.err})

			rec := h.do(http.MethodGet, incomesURL, "")

			assert.Equal(t, tt.wantStatus, rec.Code)
			body := envelope(t, rec)
			assert.Equal(t, tt.wantCode, body.Code)
			assert.NotContains(t, rec.Body.String(), assert.AnError.Error())
		})
	}
}

func TestHandlerErrors_internalCausesAreLogged(t *testing.T) {
	h := newHarness(t, setup{flags: flagSet{"module_incomes": true}, incomeErr: assert.AnError})

	h.do(http.MethodGet, incomesURL, "")

	assert.Contains(t, h.logs.String(), assert.AnError.Error())
}

func TestNew_failsClosedWhenARouteHasNoOperation(t *testing.T) {
	spec, err := gen.GetSpec()
	require.NoError(t, err)
	pruned := *spec
	pruned.Paths = openapi3.NewPaths()
	for path, item := range spec.Paths.Map() {
		if path != "/healthz" {
			pruned.Paths.Set(path, item)
		}
	}

	_, err = router.New(fakeAPI{}, &pruned, router.Options{Flags: flagSet{}, Authn: router.NoAuthentication{}, Logger: logging.New(&bytes.Buffer{}, nil)})

	assert.ErrorContains(t, err, "GET /healthz")
}

func TestDefaults(t *testing.T) {
	assert.False(t, router.AllModulesOff{}.Enabled(context.Background(), "module_auth"), "flags default OFF")
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusTeapot) })
	rec := httptest.NewRecorder()
	router.NoAuthentication{}.Authenticate(next).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	assert.Equal(t, http.StatusTeapot, rec.Code)
}
