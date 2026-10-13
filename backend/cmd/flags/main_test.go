package main

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/platform/flags"
)

func TestParseCommand(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		want    command
		wantErr bool
	}{
		{"list", []string{"list"}, command{verb: verbList}, false},
		{"presets", []string{"presets"}, command{verb: verbPresets}, false},
		{"enable", []string{"enable", "module_auth"}, command{verb: verbEnable, arg: "module_auth"}, false},
		{"disable", []string{"disable", "module_auth"}, command{verb: verbDisable, arg: "module_auth"}, false},
		{"apply-preset", []string{"apply-preset", "demo_core"}, command{verb: verbApplyPreset, arg: "demo_core"}, false},
		{"no args", nil, command{}, true},
		{"unknown verb", []string{"toggle", "module_auth"}, command{}, true},
		{"enable without a key", []string{"enable"}, command{}, true},
		{"list with an extra arg", []string{"list", "staging"}, command{}, true},
		{"apply-preset with two names", []string{"apply-preset", "a", "b"}, command{}, true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseCommand(tc.args)
			if tc.wantErr {
				require.ErrorIs(t, err, errUsage)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

// memStore is an in-memory flags.AdminStore.
type memStore struct {
	flags   []flags.Flag
	presets []flags.Preset
}

func newMemStore() *memStore {
	return &memStore{
		flags: []flags.Flag{
			{Key: "module_auth", Description: "M1 Auth & Account"},
			{Key: "module_incomes", Description: "M3 Incomes"},
			{Key: "module_taxpayers", Description: "M2 Taxpayers & Onboarding"},
		},
		presets: []flags.Preset{
			{Name: "demo_core", Description: "Core flow", Flags: []string{"module_auth", "module_taxpayers"}},
		},
	}
}

func (s *memStore) List(context.Context) ([]flags.Flag, error) {
	return append([]flags.Flag(nil), s.flags...), nil
}

func (s *memStore) SetEnabled(_ context.Context, key string, enabled bool) (bool, error) {
	for i := range s.flags {
		if s.flags[i].Key == key {
			previous := s.flags[i].Enabled
			s.flags[i].Enabled = enabled
			return previous, nil
		}
	}
	return false, flags.ErrUnknownFlag
}

func (s *memStore) ApplyPreset(ctx context.Context, name string) ([]flags.Flag, error) {
	for _, p := range s.presets {
		if p.Name != name {
			continue
		}
		for i := range s.flags {
			s.flags[i].Enabled = contains(p.Flags, s.flags[i].Key)
		}
		return s.List(ctx)
	}
	return nil, flags.ErrUnknownPreset
}

func (s *memStore) Presets(context.Context) ([]flags.Preset, error) { return s.presets, nil }

func contains(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}

func runOn(t *testing.T, store *memStore, args ...string) (string, error) {
	t.Helper()
	cmd, err := parseCommand(args)
	require.NoError(t, err)
	var out bytes.Buffer
	err = perform(t.Context(), cmd, flags.NewAdmin(store), &out)
	return out.String(), err
}

func TestListPrintsATableAndASummary(t *testing.T) {
	store := newMemStore()
	store.flags[0].Enabled = true

	out, err := runOn(t, store, "list")

	require.NoError(t, err)
	assert.Equal(t, ""+
		"KEY               STATE  DESCRIPTION\n"+
		"module_auth       ON     M1 Auth & Account\n"+
		"module_incomes    off    M3 Incomes\n"+
		"module_taxpayers  off    M2 Taxpayers & Onboarding\n"+
		"1 of 3 modules ON\n", out)
}

func TestEnableReportsTheTransition(t *testing.T) {
	store := newMemStore()

	out, err := runOn(t, store, "enable", "module_incomes")
	require.NoError(t, err)
	assert.Contains(t, out, "module_incomes: off -> ON\n")
	assert.Contains(t, out, "within 5s", "says when running API instances see it")

	out, err = runOn(t, store, "enable", "module_incomes")
	require.NoError(t, err)
	assert.Contains(t, out, "module_incomes: already ON (no change)\n")
}

func TestDisableReportsTheTransition(t *testing.T) {
	store := newMemStore()
	store.flags[0].Enabled = true

	out, err := runOn(t, store, "disable", "module_auth")
	require.NoError(t, err)
	assert.Contains(t, out, "module_auth: ON -> off\n")

	out, err = runOn(t, store, "disable", "module_auth")
	require.NoError(t, err)
	assert.Contains(t, out, "module_auth: already off (no change)\n")
}

func TestApplyPresetPrintsTheResultingState(t *testing.T) {
	store := newMemStore()
	store.flags[1].Enabled = true // module_incomes, not in demo_core

	out, err := runOn(t, store, "apply-preset", "demo_core")

	require.NoError(t, err)
	assert.Contains(t, out, "applied preset demo_core\n")
	assert.Contains(t, out, "module_incomes    off")
	assert.Contains(t, out, "module_auth       ON")
	assert.Contains(t, out, "2 of 3 modules ON\n")
}

func TestPresetsListsEachPresetWithItsFlags(t *testing.T) {
	out, err := runOn(t, newMemStore(), "presets")

	require.NoError(t, err)
	assert.Equal(t, "demo_core  Core flow\n  module_auth, module_taxpayers\n", out)
}

func TestErrorsAreActionable(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want error
	}{
		{"unknown flag", []string{"enable", "module_ghost"}, flags.ErrUnknownFlag},
		{"malformed key", []string{"enable", "Module-Auth"}, flags.ErrInvalidKey},
		{"unknown preset", []string{"apply-preset", "demo_ghost"}, flags.ErrUnknownPreset},
		{"malformed preset", []string{"apply-preset", "Demo Core"}, flags.ErrInvalidPresetName},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := runOn(t, newMemStore(), tc.args...)
			assert.ErrorIs(t, err, tc.want)
		})
	}
}

func TestRunNeedsTheOwnerURL(t *testing.T) {
	noEnv := func(string) (string, bool) { return "", false }

	err := run(t.Context(), []string{"list"}, noEnv, &bytes.Buffer{})

	require.Error(t, err)
	assert.Contains(t, err.Error(), envDatabaseURL)
}

func TestRunRejectsBadUsageBeforeConnecting(t *testing.T) {
	looked := false
	lookup := func(string) (string, bool) { looked = true; return "", false }

	err := run(t.Context(), []string{"frobnicate"}, lookup, &bytes.Buffer{})

	require.True(t, errors.Is(err, errUsage))
	assert.False(t, looked)
}
