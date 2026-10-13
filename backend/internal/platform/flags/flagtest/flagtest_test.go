package flagtest_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/romodevx/control-contable/backend/adapter/http/flaggate"
	"github.com/romodevx/control-contable/backend/internal/platform/flags/flagtest"
)

func TestStatic(t *testing.T) {
	f := flagtest.Static{"module_auth": true, "module_incomes": false}

	assert.True(t, f.Enabled(t.Context(), "module_auth"))
	assert.False(t, f.Enabled(t.Context(), "module_incomes"))
	assert.False(t, f.Enabled(t.Context(), "module_unknown"))
}

func TestOnOffRunsTheSuiteOncePerState(t *testing.T) {
	var seen []bool
	flagtest.OnOff(t, "module_incomes", func(t *testing.T, f flagtest.Static, on bool) {
		assert.Equal(t, on, f.Enabled(t.Context(), "module_incomes"))
		assert.False(t, f.Enabled(t.Context(), "module_auth"), "only the module under test is ON")
		seen = append(seen, on)
	})
	assert.Equal(t, []bool{true, false}, seen)
}

func TestOnOffWithOtherModulesKeepsThemON(t *testing.T) {
	flagtest.OnOff(t, "module_dashboard", func(t *testing.T, f flagtest.Static, _ bool) {
		assert.True(t, f.Enabled(t.Context(), "module_tax_engine"))
	}, "module_tax_engine")
}

func TestOnOffDrivesAGatedHandler(t *testing.T) {
	flagtest.OnOff(t, "module_reminders", func(t *testing.T, f flagtest.Static, on bool) {
		h := flaggate.Wrap(f, "module_reminders", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/v1/reminder-settings", nil))

		if !on {
			flagtest.AssertHidden(t, rec)
			return
		}
		assert.Equal(t, http.StatusOK, rec.Code)
	})
}

func TestAssertHiddenFailsOnAVisibleResponse(t *testing.T) {
	rec := httptest.NewRecorder()
	rec.WriteHeader(http.StatusOK)
	probe := &recordingT{}

	assert.False(t, flagtest.CheckHidden(probe, rec))
	assert.Positive(t, probe.errors)
}

func TestCheckHiddenAcceptsTheGateResponse(t *testing.T) {
	rec := httptest.NewRecorder()
	flaggate.WriteNotFound(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	probe := &recordingT{}

	assert.True(t, flagtest.CheckHidden(probe, rec))
	assert.Zero(t, probe.errors)
}

// recordingT counts assertion failures instead of failing the test.
type recordingT struct{ errors int }

func (r *recordingT) Errorf(string, ...any) { r.errors++ }
