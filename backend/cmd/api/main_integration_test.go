//go:build integration

package main

import (
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/platform/config"
	"github.com/romodevx/control-contable/backend/internal/platform/flags/flagsgorm"
	"github.com/romodevx/control-contable/backend/internal/platform/flags/flagtest"
	"github.com/romodevx/control-contable/backend/internal/platform/pgtest"
	"github.com/romodevx/control-contable/backend/migrations"
)

func TestMain(m *testing.M) {
	os.Exit(pgtest.Main(m, pgtest.WithMigrator(migrations.Up)))
}

// roomyConfig is testConfig with enough rate-limit burst for a multi-request
// scenario (testConfig's burst of 3 is sized for the 429 test).
func roomyConfig() config.Config {
	cfg := testConfig()
	cfg.RateLimit.Burst = 100
	return cfg
}

// End to end: the owner toggles a flag in Postgres (as cmd/flags does), the
// API reads it as cc_app through the composition root's flags service, and
// the real router hides or routes the module.
func TestFlags_endToEndOverPostgres(t *testing.T) {
	db := pgtest.NewDatabase(t)
	owner, err := flagsgorm.Open(db.Connect(t, pgtest.RoleOwner))
	require.NoError(t, err)
	store := flagsgorm.NewStore(owner)
	svc, err := newFlagService(db.Connect(t, pgtest.RoleApp), nil)
	require.NoError(t, err)
	h := newFlaggedHandler(t, roomyConfig(), pingOK{}, svc)

	req, rec := get(h, "/v1/me")
	flagtest.AssertHidden(t, rec)
	conforms(t, req, rec, http.MethodGet, "/v1/me")
	req, rec = get(h, "/v1/flags")
	require.Equal(t, http.StatusOK, rec.Code)
	assert.JSONEq(t, `{"enabled":[]}`, rec.Body.String())
	conforms(t, req, rec, http.MethodGet, "/v1/flags")

	_, err = store.SetEnabled(t.Context(), "module_auth", true)
	require.NoError(t, err)
	svc.Invalidate() // another process would wait for flags.DefaultTTL

	_, rec = get(h, "/v1/me")
	assert.NotEqual(t, http.StatusNotFound, rec.Code, "module_auth is ON")
	req, rec = get(h, "/v1/flags")
	require.Equal(t, http.StatusOK, rec.Code)
	assert.JSONEq(t, `{"enabled":["module_auth"]}`, rec.Body.String())
	conforms(t, req, rec, http.MethodGet, "/v1/flags")
}

// A runtime role without SELECT on feature_flags (cc_webhook) cannot read the
// flags: every module is hidden and GET /v1/flags is the 500 envelope.
func TestFlags_unreadableFlagsFailClosed(t *testing.T) {
	db := pgtest.NewDatabase(t)
	owner, err := flagsgorm.Open(db.Connect(t, pgtest.RoleOwner))
	require.NoError(t, err)
	_, err = flagsgorm.NewStore(owner).ApplyPreset(t.Context(), "demo_full")
	require.NoError(t, err)
	svc, err := newFlagService(db.Connect(t, pgtest.RoleWebhook), nil)
	require.NoError(t, err)
	h := newFlaggedHandler(t, roomyConfig(), pingOK{}, svc)

	_, rec := get(h, "/v1/me")
	flagtest.AssertHidden(t, rec)
	req, rec := get(h, "/v1/flags")
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	conforms(t, req, rec, http.MethodGet, "/v1/flags")
}
