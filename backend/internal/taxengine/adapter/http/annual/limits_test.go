package annualhttp_test

import (
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/adapter/http/gen"
	"github.com/romodevx/control-contable/backend/internal/platform/ratelimit"
	annualhttp "github.com/romodevx/control-contable/backend/internal/taxengine/adapter/http/annual"
)

const moduleAnnual = "module_annual"

// filedPaymentsWrite spends the estimate budget, not annual_writes: it runs
// the Estimator range to show the app's estimates (ADR-0009 "Budgets").
const filedPaymentsWrite = "putFiledPayments"

// patternRecorder is the generated server's base router: it records every
// ServeMux pattern the server registers and serves nothing.
type patternRecorder struct{ patterns []string }

func (m *patternRecorder) HandleFunc(pattern string, _ func(http.ResponseWriter, *http.Request)) {
	m.patterns = append(m.patterns, pattern)
}

func (*patternRecorder) ServeHTTP(http.ResponseWriter, *http.Request) {}

// annualRoutes are the module_annual operations the generated router
// registers, by operationId with their HTTP method. They come from the
// routes themselves and the flag map the gate reads, never from a list
// written by hand, so a new annual route shows up here on its own.
func annualRoutes(t *testing.T) map[string]string {
	t.Helper()
	spec, err := gen.GetSpec()
	require.NoError(t, err)
	rec := &patternRecorder{}
	gen.HandlerWithOptions(gen.NewStrictHandler(nil, nil), gen.StdHTTPServerOptions{BaseRouter: rec})

	routes := make(map[string]string)
	for _, pattern := range rec.patterns {
		module, ok := gen.RouteModule(pattern)
		require.True(t, ok, "route %q has a module", pattern)
		if module != moduleAnnual {
			continue
		}
		method, path, ok := strings.Cut(pattern, " ")
		require.True(t, ok, "route %q is METHOD /path", pattern)
		item := spec.Paths.Find(path)
		require.NotNil(t, item, "route %q is in the contract", pattern)
		op := item.GetOperation(method)
		require.NotNil(t, op, "route %q is in the contract", pattern)
		routes[op.OperationID] = method
	}
	require.NotEmpty(t, routes, "the router registers module_annual routes")
	return routes
}

// ANN-25 follow-up: every annual route the router serves spends a per-user
// budget, and the right one. Reads spend annual_reads (600/h) or the
// estimate's estimate_reads; writes spend annual_writes (300/h), but for the
// filed-payments write, which shows estimates. A route added without a
// budget, or with a budget of the wrong kind, fails here.
func TestUserLimits_coverEveryAnnualRouteOfTheRouter(t *testing.T) {
	reads := ratelimit.PerWindow("annual_reads", 600, time.Hour)
	writes := ratelimit.PerWindow("annual_writes", 300, time.Hour)
	estimates := annualhttp.RangePolicy
	require.Equal(t, "estimate_reads", estimates.Scope)

	routes := annualRoutes(t)
	for op, method := range routes {
		t.Run(op, func(t *testing.T) {
			policy, ok := annualhttp.UserLimits[op]
			require.True(t, ok, "%s %s has no rate-limit budget", method, op)
			require.NoError(t, policy.Validate())
			switch {
			case op == filedPaymentsWrite:
				assert.Equal(t, estimates, policy)
			case method == http.MethodGet:
				assert.Contains(t, []ratelimit.Policy{reads, estimates}, policy, "a read spends a read budget")
			default:
				assert.Equal(t, writes, policy, "a write spends annual_writes")
			}
		})
	}
	for op := range annualhttp.UserLimits {
		assert.Contains(t, routes, op, "budget %q limits no annual route", op)
	}
}
