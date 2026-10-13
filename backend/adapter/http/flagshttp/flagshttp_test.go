package flagshttp_test

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/adapter/http/flagshttp"
	"github.com/romodevx/control-contable/backend/adapter/http/gen"
	"github.com/romodevx/control-contable/backend/adapter/http/router"
	"github.com/romodevx/control-contable/backend/internal/platform/flags"
	"github.com/romodevx/control-contable/backend/internal/platform/flags/flagtest"
	"github.com/romodevx/control-contable/backend/internal/platform/requestid"
)

var _ flagshttp.Lister = (*flags.Service)(nil)

const testRequestID = "req-flags"

type stubLister struct {
	keys []string
	err  error
}

func (s stubLister) EnabledModules(context.Context) ([]string, error) { return s.keys, s.err }

// server embeds the flags handler the way the composition root does; the
// other operations are unreachable in these tests.
type server struct {
	gen.StrictServerInterface
	*flagshttp.Handler
}

func (s server) ListEnabledFlags(ctx context.Context, req gen.ListEnabledFlagsRequestObject) (gen.ListEnabledFlagsResponseObject, error) {
	return s.Handler.ListEnabledFlags(ctx, req)
}

func get(t *testing.T, lister flagshttp.Lister) *httptest.ResponseRecorder {
	t.Helper()
	spec, err := gen.GetSpec()
	require.NoError(t, err)
	h, err := router.New(server{Handler: flagshttp.New(lister, nil)}, spec, router.Options{
		Flags:  flagtest.Static{}, // every module OFF: /v1/flags is platform
		Authn:  router.NoAuthentication{},
		Logger: slog.New(slog.DiscardHandler),
	})
	require.NoError(t, err)
	req := httptest.NewRequest(http.MethodGet, "/v1/flags", nil)
	req = req.WithContext(requestid.With(req.Context(), testRequestID))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestListEnabledFlags(t *testing.T) {
	tests := []struct {
		name string
		keys []string
		want string
	}{
		{"some modules ON", []string{"module_auth", "module_incomes"}, `{"enabled":["module_auth","module_incomes"]}`},
		{"nothing ON is an empty array", []string{}, `{"enabled":[]}`},
		{"nil from the lister is still an array", nil, `{"enabled":[]}`},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rec := get(t, stubLister{keys: tc.keys})

			require.Equal(t, http.StatusOK, rec.Code)
			assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
			assert.JSONEq(t, tc.want, rec.Body.String())
			assertMatchesContract(t, rec)
		})
	}
}

func TestListEnabledFlagsAnswers500EnvelopeWhenFlagsAreUnreadable(t *testing.T) {
	rec := get(t, stubLister{err: errors.New("connection refused")})

	require.Equal(t, http.StatusInternalServerError, rec.Code)
	var body gen.ErrorResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.Equal(t, gen.ErrorCode("internal_error"), body.Error.Code)
	assert.Equal(t, testRequestID, body.Error.RequestId)
	assert.NotContains(t, rec.Body.String(), "connection refused", "the cause is logged, never sent")
	assertMatchesContract(t, rec)
}

func TestListEnabledFlagsServesTheRealService(t *testing.T) {
	svc := flags.NewService(readerFunc(func(context.Context) ([]string, error) {
		return []string{"module_taxpayers", "module_auth"}, nil
	}))

	rec := get(t, svc)

	require.Equal(t, http.StatusOK, rec.Code)
	assert.JSONEq(t, `{"enabled":["module_auth","module_taxpayers"]}`, rec.Body.String())
}

type readerFunc func(context.Context) ([]string, error)

func (f readerFunc) EnabledKeys(ctx context.Context) ([]string, error) { return f(ctx) }

// assertMatchesContract validates the response of GET /v1/flags against the
// embedded contract (status, content type and body schema).
func assertMatchesContract(t *testing.T, rec *httptest.ResponseRecorder) {
	t.Helper()
	require.NoError(t, validateAgainstContract(t, rec))
}

func validateAgainstContract(t *testing.T, rec *httptest.ResponseRecorder) error {
	t.Helper()
	spec, err := gen.GetSpec()
	require.NoError(t, err)
	item := spec.Paths.Value("/v1/flags")
	require.NotNil(t, item)
	req := httptest.NewRequest(http.MethodGet, "/v1/flags", nil)
	input := &openapi3filter.ResponseValidationInput{
		RequestValidationInput: &openapi3filter.RequestValidationInput{
			Request: req,
			Route:   &routers.Route{Spec: spec, Path: "/v1/flags", PathItem: item, Method: http.MethodGet, Operation: item.Get},
		},
		Status:  rec.Code,
		Header:  rec.Header(),
		Options: &openapi3filter.Options{IncludeResponseStatus: true},
	}
	input.SetBodyBytes(rec.Body.Bytes())
	return openapi3filter.ValidateResponse(t.Context(), input)
}

func TestContractValidatorRejectsAMalformedKey(t *testing.T) {
	rec := httptest.NewRecorder()
	rec.Header().Set("Content-Type", "application/json")
	rec.WriteHeader(http.StatusOK)
	_, _ = rec.WriteString(`{"enabled":["Not A Key"]}`)

	assert.Error(t, validateAgainstContract(t, rec))
}
