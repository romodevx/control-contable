package gen_test

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/adapter/http/gen"
)

// moduleKeyPattern mirrors rule/operation-x-module in redocly.yaml.
var moduleKeyPattern = regexp.MustCompile(`^(platform|module_[a-z][a-z_]*)$`)

type contractOperation struct {
	where  string // "METHOD /path"
	id     string
	module any // raw x-module extension value
}

// contractOperations walks the spec embedded in the generated code, which is
// what the server actually serves and validates against.
func contractOperations(t *testing.T) []contractOperation {
	t.Helper()
	swagger, err := gen.GetSwagger()
	require.NoError(t, err, "embedded spec must load")

	var ops []contractOperation
	for path, item := range swagger.Paths.Map() {
		for method, op := range item.Operations() {
			ops = append(ops, contractOperation{
				where:  method + " " + path,
				id:     op.OperationID,
				module: op.Extensions["x-module"],
			})
		}
	}
	require.NotEmpty(t, ops, "embedded spec has no operations")
	return ops
}

// TestEveryOperationDeclaresXModule is the ADR-0001 guard: an operation without
// x-module could not be gated by its module's flag, so it must not exist.
// Public operations (x-public: true) are not exempt; the always-on ones declare
// x-module: platform and are the only ones the flag gate lets through.
func TestEveryOperationDeclaresXModule(t *testing.T) {
	for _, op := range contractOperations(t) {
		module, isString := op.module.(string)
		if !assert.Truef(t, isString, "%s (%s): x-module missing or not a string", op.where, op.id) {
			continue
		}
		assert.Regexpf(t, moduleKeyPattern, module,
			"%s (%s): x-module %q is neither platform nor a module_<name> flag key", op.where, op.id, module)
	}
}

func TestFlagMapMatchesContract(t *testing.T) {
	ops := contractOperations(t)
	want := make(map[string]string, len(ops))
	for _, op := range ops {
		module, _ := op.module.(string)
		want[op.id] = module
	}

	assert.Equal(t, want, gen.OperationModules(), "flagmap.gen.go is stale: run `make generate`")

	for id, module := range want {
		got, ok := gen.OperationModule(id)
		assert.Truef(t, ok, "OperationModule(%q) unknown", id)
		assert.Equal(t, module, got)
	}
	_, ok := gen.OperationModule("notAnOperation")
	assert.False(t, ok, "an undeclared operationId must not resolve to a module")
}

// TestFlagMapCoversStrictServer ties the map to the interface the composition
// root implements: every strict handler has exactly one gated entry.
func TestFlagMapCoversStrictServer(t *testing.T) {
	modules := gen.OperationModules()
	handlers := reflect.TypeFor[gen.StrictServerInterface]()

	require.Equal(t, len(modules), handlers.NumMethod(), "one flag-map entry per strict handler")
	for id := range modules {
		_, found := handlers.MethodByName(strings.ToUpper(id[:1]) + id[1:])
		assert.Truef(t, found, "operationId %q has no StrictServerInterface method", id)
	}
}

// TestRouteModuleMatchesServeMuxPatterns proves the route lookup is keyed by the
// exact pattern the generated std-http server registers (what r.Pattern holds),
// so the flag gate can run before the strict handler decodes the body.
func TestRouteModuleMatchesServeMuxPatterns(t *testing.T) {
	mux := http.NewServeMux()
	gen.HandlerFromMux(gen.NewStrictHandler(nil, nil), mux)
	pathParam := regexp.MustCompile(`\{[^}]+\}`)

	ops := contractOperations(t)
	for _, op := range ops {
		method, path, _ := strings.Cut(op.where, " ")
		req := httptest.NewRequest(method, pathParam.ReplaceAllString(path, "x"), nil)

		_, pattern := mux.Handler(req)
		require.Equalf(t, op.where, pattern, "ServeMux pattern for %s", op.id)

		module, ok := gen.RouteModule(pattern)
		require.Truef(t, ok, "RouteModule(%q) unknown", pattern)
		assert.Equal(t, op.module, module, pattern)
	}
	_, ok := gen.RouteModule("GET /not-a-route")
	assert.False(t, ok)
}

func TestOperationModulesReturnsACopy(t *testing.T) {
	m := gen.OperationModules()
	m["getHealth"] = "module_tampered"

	got, ok := gen.OperationModule("getHealth")
	require.True(t, ok)
	assert.Equal(t, gen.ModulePlatform, got)
}
