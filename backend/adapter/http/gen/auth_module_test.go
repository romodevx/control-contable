package gen_test

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/romodevx/control-contable/backend/adapter/http/gen"
)

const moduleAuth = "module_auth"

// m1Routes are the M1 operations of MODULE-auth §1, written by hand: the
// generated map must agree with the spec, not merely with itself.
var m1Routes = map[string]string{
	"register":           "POST /v1/auth/register",
	"verifyEmail":        "POST /v1/auth/verify-email",
	"resendVerification": "POST /v1/auth/verify-email/resend",
	"login":              "POST /v1/auth/login",
	"refreshSession":     "POST /v1/auth/refresh",
	"logout":             "POST /v1/auth/logout",
	"getMe":              "GET /v1/me",
	"updateMe":           "PATCH /v1/me",
	"deleteMe":           "DELETE /v1/me",
	"listLegalDocuments": "GET /v1/legal/documents",
}

// TestFlagMap_everyM1OperationIsGatedByModuleAuth (M1-14): every M1
// operation, by operationId and by the ServeMux pattern the gate reads, is
// owned by module_auth, and module_auth owns nothing else. A route missing
// from here would ship ungated or hidden behind another module's flag.
func TestFlagMap_everyM1OperationIsGatedByModuleAuth(t *testing.T) {
	for operationID, pattern := range m1Routes {
		t.Run(operationID, func(t *testing.T) {
			module, ok := gen.OperationModule(operationID)
			assert.True(t, ok, "operationId is in the contract")
			assert.Equal(t, moduleAuth, module, "operationId -> flag")

			module, ok = gen.RouteModule(pattern)
			assert.True(t, ok, "route %q is registered", pattern)
			assert.Equal(t, moduleAuth, module, "route -> flag")
		})
	}

	var owned []string
	for operationID, module := range gen.OperationModules() {
		if module == moduleAuth {
			owned = append(owned, operationID)
		}
	}
	slices.Sort(owned)
	want := make([]string, 0, len(m1Routes))
	for operationID := range m1Routes {
		want = append(want, operationID)
	}
	slices.Sort(want)
	assert.Equal(t, want, owned, "module_auth owns exactly the M1 operations")
}
