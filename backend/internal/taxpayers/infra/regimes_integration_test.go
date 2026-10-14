//go:build integration

package infra_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/platform/apperr"
	"github.com/romodevx/control-contable/backend/internal/taxpayers/domain"
	"github.com/romodevx/control-contable/backend/internal/taxpayers/usecase"
)

// createFisica onboards a física of userA holding regimes; it returns its id.
func createFisica(t *testing.T, e *env, regimes ...usecase.RegimeInput) string {
	t.Helper()
	created, err := e.taxpayers.Create(t.Context(), userA, fisica(rfcFisica, regimes...))
	require.NoError(t, err)
	return created.Taxpayer.ID
}

func TestAddRegime_CompatibilityAndPeriods(t *testing.T) {
	e := newEnv(t, userA).wire(t)
	id := createFisica(t, e, usecase.RegimeInput{Code: "612", Start: "2026-01"})

	_, err := e.regimes.Add(t.Context(), userA, id, usecase.RegimeInput{Code: "626", Start: "2026-01"})
	requireCode(t, err, apperr.CodeInvalidRegimeCombination)

	_, err = e.regimes.Add(t.Context(), userA, id, usecase.RegimeInput{Code: "612"})
	requireCode(t, err, apperr.CodeConflict)

	_, err = e.regimes.Add(t.Context(), userA, id, usecase.RegimeInput{Code: "601"})
	requireCode(t, err, apperr.CodeValidationFailed)

	added, err := e.regimes.Add(t.Context(), userA, id, usecase.RegimeInput{Code: "606", Start: "2026-03"})
	require.NoError(t, err)
	assert.Equal(t, "2026-03", added.Window.Start.String())
	assert.True(t, added.Window.IsOpen())
}

func TestAddRegime_ExitFromRESICOTheNextMonth(t *testing.T) {
	e := newEnv(t, userA).wire(t)
	id := createFisica(t, e, usecase.RegimeInput{Code: "626", Start: "2026-01"}, usecase.RegimeInput{Code: "605", Start: "2026-01"})
	_, err := e.regimes.End(t.Context(), userA, id, usecase.EndRegimeInput{Code: "626", End: "2026-06"})
	require.NoError(t, err)

	added, err := e.regimes.Add(t.Context(), userA, id, usecase.RegimeInput{Code: "612", Start: "2026-07"})

	require.NoError(t, err)
	assert.Equal(t, "2026-07", added.Window.Start.String())
}

func TestAddRegime_ReopenKeepsTheStoredStartSoNoFalseOverlap(t *testing.T) {
	// PR #57 follow-up: 626 held 2026-01..2026-02, 606 held 2026-03..2026-04, 612
	// since 2026-05. Re-adding 606 with a body start of 2026-02 reopens the row
	// from its stored 2026-03: no overlap with 626, so no 23P01.
	e := newEnv(t, userA).wire(t)
	id := createFisica(t, e,
		usecase.RegimeInput{Code: "626", Start: "2026-01"},
		usecase.RegimeInput{Code: "605", Start: "2026-01"})
	_, err := e.regimes.End(t.Context(), userA, id, usecase.EndRegimeInput{Code: "626", End: "2026-02"})
	require.NoError(t, err)
	_, err = e.regimes.Add(t.Context(), userA, id, usecase.RegimeInput{Code: "606", Start: "2026-03"})
	require.NoError(t, err)
	_, err = e.regimes.End(t.Context(), userA, id, usecase.EndRegimeInput{Code: "606", End: "2026-04"})
	require.NoError(t, err)
	_, err = e.regimes.Add(t.Context(), userA, id, usecase.RegimeInput{Code: "612", Start: "2026-05"})
	require.NoError(t, err)

	reopened, err := e.regimes.Add(t.Context(), userA, id, usecase.RegimeInput{Code: "606", Start: "2026-02",
		Options: map[string]domain.OptionValue{"deduction_mode": domain.StringOption("actual")}})

	require.NoError(t, err)
	assert.Equal(t, "2026-03", reopened.Window.Start.String())
	assert.True(t, reopened.Window.IsOpen())
	assert.Equal(t, domain.StringOption("actual"), reopened.Options["deduction_mode"])
	var rows int
	queryAs(t, e.app, userA, "SELECT count(*) FROM taxpayer_regimes WHERE taxpayer_id = $1 AND regime_code = '606'", []any{id}, &rows)
	assert.Equal(t, 1, rows, "OQ-5: the same row reopens")
}

func TestAddRegime_ReopenRefusesARealOverlap(t *testing.T) {
	e := newEnv(t, userA).wire(t)
	id := createFisica(t, e, usecase.RegimeInput{Code: "606", Start: "2026-01"}, usecase.RegimeInput{Code: "605", Start: "2026-01"})
	_, err := e.regimes.End(t.Context(), userA, id, usecase.EndRegimeInput{Code: "606", End: "2026-03"})
	require.NoError(t, err)
	_, err = e.regimes.Add(t.Context(), userA, id, usecase.RegimeInput{Code: "626", Start: "2026-04"})
	require.NoError(t, err)

	_, err = e.regimes.Add(t.Context(), userA, id, usecase.RegimeInput{Code: "606"})

	requireCode(t, err, apperr.CodeInvalidRegimeCombination)
}

func TestRegimeStore_MapsDatabaseRefusals(t *testing.T) {
	e := newEnv(t, userA).wire(t)
	id := createFisica(t, e, usecase.RegimeInput{Code: "612", Start: "2026-01"})
	base := usecase.NewRegime{TaxpayerID: id, UserID: userA, PersonType: domain.PersonTypeFisica, Start: mustPeriod(t, "2026-01")}
	insert := func(r usecase.NewRegime) error {
		return e.txm.ReadWrite(t.Context(), userA, func(ctx context.Context) error { return e.store.InsertRegime(ctx, r) })
	}

	clash := base
	clash.Code = "626"
	require.ErrorIs(t, insert(clash), usecase.ErrIncompatibleRegimes, "the trigger's 23P01 backstop")
	again := base
	again.Code = "612"
	require.ErrorIs(t, insert(again), usecase.ErrRegimeAlreadyActive, "23505 on the primary key")
	moralOnly := base
	moralOnly.Code = "601"
	require.ErrorIs(t, insert(moralOnly), usecase.ErrRegimeNotForPersonType, "the catalog foreign key")
	early := base
	early.Code, early.Start = "606", mustPeriod(t, "2021-12")
	require.ErrorIs(t, insert(early), usecase.ErrPeriodOutOfRange, "the start_period CHECK")
}

func TestUpdateRegimeOptions_Persists(t *testing.T) {
	e := newEnv(t, userA).wire(t)
	id := createFisica(t, e, usecase.RegimeInput{Code: "625", Start: "2026-01", Options: map[string]domain.OptionValue{
		"activities": domain.ListOption([]string{"lodging"})}})

	_, err := e.regimes.UpdateOptions(t.Context(), userA, id, "625", map[string]domain.OptionValue{"definitive_option": domain.BoolOption(true)})
	require.NoError(t, err)
	got, err := e.taxpayers.Get(t.Context(), userA, id)
	require.NoError(t, err)

	assert.Equal(t, map[string]domain.OptionValue{
		"activities":            domain.ListOption([]string{"lodging"}),
		"definitive_option":     domain.BoolOption(true),
		"rfc_given_to_platform": domain.BoolOption(true),
	}, got.Regimes[0].Options)
}

func TestEndRegime_RulesAgainstTheDatabase(t *testing.T) {
	e := newEnv(t, userA).wire(t)
	id := createFisica(t, e, usecase.RegimeInput{Code: "606", Start: "2026-01"}, usecase.RegimeInput{Code: "612", Start: "2026-01"})

	ended, err := e.regimes.End(t.Context(), userA, id, usecase.EndRegimeInput{Code: "606", End: "2026-09"})
	require.NoError(t, err)
	assert.Equal(t, "2026-09", ended.Window.End.String())

	_, err = e.regimes.End(t.Context(), userA, id, usecase.EndRegimeInput{Code: "606", End: "2026-09"})
	requireCode(t, err, apperr.CodeConflict)
	_, err = e.regimes.End(t.Context(), userA, id, usecase.EndRegimeInput{Code: "612", End: "2026-09"})
	requireCode(t, err, apperr.CodeBusinessRuleViolation) // last active regime
}

func TestEndRegime_AllowedForAReadOnlyUser(t *testing.T) {
	e := newEnv(t, userA).wire(t)
	id := createFisica(t, e, usecase.RegimeInput{Code: "606", Start: "2026-01"}, usecase.RegimeInput{Code: "612", Start: "2026-01"})
	e.wire(t, readOnlyAccess(e.txm))

	_, err := e.regimes.End(t.Context(), userA, id, usecase.EndRegimeInput{Code: "606", End: "2026-09"})
	require.NoError(t, err)
	_, err = e.taxpayers.Archive(t.Context(), userA, id)
	require.NoError(t, err)
	_, err = e.regimes.Add(t.Context(), userA, id, usecase.RegimeInput{Code: "605"})
	requireCode(t, err, apperr.CodeEntitlementRequired)
}

func TestQueries_CheckRegimeAndProfile(t *testing.T) {
	e := newEnv(t, userA).wire(t)
	id := createFisica(t, e, usecase.RegimeInput{Code: "626", Start: "2026-01"}, usecase.RegimeInput{Code: "605", Start: "2026-01"})
	_, err := e.regimes.End(t.Context(), userA, id, usecase.EndRegimeInput{Code: "626", End: "2026-06"})
	require.NoError(t, err)

	require.NoError(t, e.queries.CheckRegime(t.Context(), userA, id, "626", "2026-06"))
	requireCode(t, e.queries.CheckRegime(t.Context(), userA, id, "626", "2026-07"), apperr.CodeBusinessRuleViolation)
	// Nested in the record's own unit of work it joins it (FOR SHARE needs read-write).
	require.NoError(t, e.txm.ReadWrite(t.Context(), userA, func(ctx context.Context) error {
		return e.queries.CheckRegime(ctx, userA, id, "605", "2026-10")
	}))

	profile, err := e.queries.Profile(t.Context(), userA, id)
	require.NoError(t, err)
	assert.Equal(t, domain.PersonTypeFisica, profile.PersonType)
	require.Len(t, profile.Regimes, 2)
	assert.Equal(t, "2026-06", profile.Regimes[1].Window.End.String())

	_, err = e.taxpayers.Archive(t.Context(), userA, id)
	require.NoError(t, err)
	requireCode(t, e.queries.CheckRegime(t.Context(), userA, id, "605", "2026-10"), apperr.CodeBusinessRuleViolation)
	profile, err = e.queries.Profile(t.Context(), userA, id)
	require.NoError(t, err)
	assert.True(t, profile.Archived)
}

func mustPeriod(t *testing.T, raw string) domain.Period {
	t.Helper()
	p, err := domain.ParsePeriod(raw)
	require.NoError(t, err)
	return p
}
