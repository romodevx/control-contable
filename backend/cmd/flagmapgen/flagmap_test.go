package main

import (
	"go/parser"
	"go/token"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseOperations(t *testing.T) {
	tests := []struct {
		name    string
		spec    string
		want    []Operation
		wantErr []string // substrings the error must contain; nil means no error
	}{
		{
			name: "collects every operation sorted by operationId and ignores path-level keys",
			spec: `
openapi: 3.1.0
paths:
  /v1/b:
    summary: path-level summary is not an operation
    parameters: []
    post: { operationId: zeta, x-module: module_auth }
    get: { operationId: alpha, x-module: module_auth }
  /healthz:
    get: { operationId: getHealth, x-module: platform, x-public: true }
`,
			want: []Operation{
				{ID: "alpha", Method: "GET", Path: "/v1/b", Module: "module_auth"},
				{ID: "getHealth", Method: "GET", Path: "/healthz", Module: "platform"},
				{ID: "zeta", Method: "POST", Path: "/v1/b", Module: "module_auth"},
			},
		},
		{
			name: "every HTTP method is an operation",
			spec: `
paths:
  /x:
    put: { operationId: a, x-module: module_x }
    patch: { operationId: b, x-module: module_x }
    delete: { operationId: c, x-module: module_x }
    head: { operationId: d, x-module: module_x }
    options: { operationId: e, x-module: module_x }
    trace: { operationId: f, x-module: module_x }
`,
			want: []Operation{
				{ID: "a", Method: "PUT", Path: "/x", Module: "module_x"},
				{ID: "b", Method: "PATCH", Path: "/x", Module: "module_x"},
				{ID: "c", Method: "DELETE", Path: "/x", Module: "module_x"},
				{ID: "d", Method: "HEAD", Path: "/x", Module: "module_x"},
				{ID: "e", Method: "OPTIONS", Path: "/x", Module: "module_x"},
				{ID: "f", Method: "TRACE", Path: "/x", Module: "module_x"},
			},
		},
		{
			name: "operation without x-module fails",
			spec: `
paths:
  /v1/me:
    get: { operationId: getMe }
`,
			wantErr: []string{"GET /v1/me", "getMe", "x-module"},
		},
		{
			name: "public operation without x-module fails too",
			spec: `
paths:
  /v1/auth/login:
    post: { operationId: login, x-public: true, security: [] }
`,
			wantErr: []string{"POST /v1/auth/login", "x-module"},
		},
		{
			name: "x-module that is not platform nor a module_ flag key fails",
			spec: `
paths:
  /v1/me:
    get: { operationId: getMe, x-module: auth }
`,
			wantErr: []string{"GET /v1/me", `"auth"`},
		},
		{
			name: "x-module with uppercase fails",
			spec: `
paths:
  /v1/me:
    get: { operationId: getMe, x-module: module_Auth }
`,
			wantErr: []string{"module_Auth"},
		},
		{
			// P0-29: the feature_flags CHECK rejects these keys, so a route
			// tagged with one could never be enabled (silently 404 forever).
			name: "x-module the feature_flags key format rejects fails",
			spec: `
paths:
  /v1/a:
    get: { operationId: a, x-module: module_auth_ }
  /v1/b:
    get: { operationId: b, x-module: module_a__b }
`,
			wantErr: []string{"module_auth_", "module_a__b"},
		},
		{
			name: "x-module with a digit is a valid flag key",
			spec: `
paths:
  /v1/a:
    get: { operationId: a, x-module: module_m2 }
`,
			want: []Operation{{ID: "a", Method: "GET", Path: "/v1/a", Module: "module_m2"}},
		},
		{
			name: "operation without operationId fails",
			spec: `
paths:
  /v1/me:
    get: { x-module: module_auth }
`,
			wantErr: []string{"GET /v1/me", "operationId"},
		},
		{
			name: "duplicate operationId fails",
			spec: `
paths:
  /a:
    get: { operationId: same, x-module: module_a }
  /b:
    get: { operationId: same, x-module: module_a }
`,
			wantErr: []string{"duplicate operationId", "same"},
		},
		{
			name: "every violation is reported, not only the first",
			spec: `
paths:
  /a:
    get: { operationId: a }
  /b:
    get: { operationId: b, x-module: nope }
`,
			wantErr: []string{"GET /a", "GET /b"},
		},
		{
			// A $ref path item hides its operations from this parser; skipping it
			// would silently drop them from the flag map (fail open).
			name: "path item $ref fails instead of dropping its operations",
			spec: `
paths:
  /a:
    get: { operationId: a, x-module: module_a }
  /b:
    $ref: './paths/b.yaml'
`,
			wantErr: []string{"/b", "$ref"},
		},
		{
			name:    "a spec without operations fails",
			spec:    "openapi: 3.1.0\npaths: {}\n",
			wantErr: []string{"no operations"},
		},
		{
			name:    "invalid YAML fails",
			spec:    "paths: [unclosed",
			wantErr: []string{"parse"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseOperations([]byte(tt.spec))
			if tt.wantErr != nil {
				require.Error(t, err)
				for _, s := range tt.wantErr {
					assert.Contains(t, err.Error(), s)
				}
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestRender(t *testing.T) {
	ops := []Operation{
		{ID: "getHealth", Method: "GET", Path: "/healthz", Module: "platform"},
		{ID: "getMe", Method: "GET", Path: "/v1/me", Module: "module_auth"},
	}

	src, err := Render("gen", "openapi.yaml", ops)
	require.NoError(t, err)
	code := string(src)

	t.Run("is valid Go in the requested package", func(t *testing.T) {
		f, err := parser.ParseFile(token.NewFileSet(), "flagmap.gen.go", src, parser.ParseComments)
		require.NoError(t, err)
		assert.Equal(t, "gen", f.Name.Name)
	})
	t.Run("carries the generated-code marker linters recognise", func(t *testing.T) {
		assert.True(t, strings.HasPrefix(code, "// Code generated by flagmapgen from openapi.yaml. DO NOT EDIT.\n"))
	})
	t.Run("maps each operationId to its module", func(t *testing.T) {
		assert.Contains(t, code, `"getHealth": "platform",`)
		assert.Contains(t, code, `"getMe":     "module_auth",`)
	})
	t.Run("maps each ServeMux route pattern to its module", func(t *testing.T) {
		assert.Contains(t, code, `"GET /healthz": "platform",`)
		assert.Contains(t, code, `"GET /v1/me":   "module_auth",`)
	})
	t.Run("is deterministic", func(t *testing.T) {
		again, err := Render("gen", "openapi.yaml", ops)
		require.NoError(t, err)
		assert.Equal(t, src, again)
	})
}
