package router_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/romodevx/control-contable/backend/internal/platform/apperr"
)

// TestNotFound_uncleanPathsDoNotRevealDisabledRoutes pins that ServeMux's
// path-cleaning redirect (301 to the cleaned path) never fires. ServeMux
// redirects only when the cleaned path matches a registered pattern, so a 301
// for "/v1//auth/login" against a 404 for "/v1//nope" would tell an anonymous
// caller which disabled routes exist.
func TestNotFound_uncleanPathsDoNotRevealDisabledRoutes(t *testing.T) {
	targets := []string{
		"/v1//auth/login",
		"/v1/auth/./login",
		"/v1/x/../auth/login",
		"/v1//nope",
	}
	for _, target := range targets {
		t.Run(target, func(t *testing.T) {
			h := newHarness(t, setup{flags: flagSet{}})

			rec := h.do(http.MethodPost, target, validLogin)

			assert.Equal(t, http.StatusNotFound, rec.Code)
			assert.Empty(t, rec.Header().Get("Location"), "no redirect to the canonical route")
			assert.Equal(t, apperr.CodeNotFound, envelope(t, rec).Code)
			assert.Empty(t, h.calls)
		})
	}
}
