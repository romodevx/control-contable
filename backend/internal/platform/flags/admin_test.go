package flags_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/platform/flags"
)

// fakeStore records the calls the Admin use case lets through.
type fakeStore struct {
	flags   []flags.Flag
	presets []flags.Preset
	calls   []string
	err     error
}

func (s *fakeStore) List(context.Context) ([]flags.Flag, error) {
	s.calls = append(s.calls, "list")
	return s.flags, s.err
}

func (s *fakeStore) SetEnabled(_ context.Context, key string, enabled bool) (bool, error) {
	s.calls = append(s.calls, "set "+key)
	return !enabled, s.err
}

func (s *fakeStore) ApplyPreset(_ context.Context, name string) ([]flags.Flag, error) {
	s.calls = append(s.calls, "apply "+name)
	return s.flags, s.err
}

func (s *fakeStore) Presets(context.Context) ([]flags.Preset, error) {
	s.calls = append(s.calls, "presets")
	return s.presets, s.err
}

func TestValidKey(t *testing.T) {
	tests := []struct {
		key  string
		want bool
	}{
		{"module_auth", true},
		{"module_tax_engine", true},
		{"module_m2", true},
		{"module_", false},
		{"module_Auth", false},
		{"module__auth", false},
		{"module_auth_", false},
		{"module_2fa", false},
		{"auth", false},
		{"platform", false},
		{"", false},
		{"module_auth; DROP TABLE feature_flags", false},
	}
	for _, tc := range tests {
		t.Run(tc.key, func(t *testing.T) {
			assert.Equal(t, tc.want, flags.ValidKey(tc.key))
		})
	}
}

func TestValidPresetName(t *testing.T) {
	tests := []struct {
		name string
		want bool
	}{
		{"demo_core", true},
		{"demo", true},
		{"demo2", true},
		{"Demo", false},
		{"demo-core", false},
		{"_demo", false},
		{"", false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, flags.ValidPresetName(tc.name))
		})
	}
}

func TestAdminSetEnabledRejectsAMalformedKeyBeforeTheStore(t *testing.T) {
	store := &fakeStore{}
	admin := flags.NewAdmin(store)

	_, err := admin.SetEnabled(t.Context(), "Module-Auth", true)

	require.ErrorIs(t, err, flags.ErrInvalidKey)
	assert.Empty(t, store.calls)
}

func TestAdminSetEnabledPassesAValidKeyThrough(t *testing.T) {
	store := &fakeStore{}
	admin := flags.NewAdmin(store)

	previous, err := admin.SetEnabled(t.Context(), "module_auth", true)

	require.NoError(t, err)
	assert.False(t, previous)
	assert.Equal(t, []string{"set module_auth"}, store.calls)
}

func TestAdminApplyPresetRejectsAMalformedNameBeforeTheStore(t *testing.T) {
	store := &fakeStore{}
	admin := flags.NewAdmin(store)

	_, err := admin.ApplyPreset(t.Context(), "Demo Core")

	require.ErrorIs(t, err, flags.ErrInvalidPresetName)
	assert.Empty(t, store.calls)
}

func TestAdminDelegatesReadsAndPresetApplication(t *testing.T) {
	want := []flags.Flag{{Key: "module_auth", Enabled: true, Description: "M1"}}
	store := &fakeStore{flags: want, presets: []flags.Preset{{Name: "demo_core"}}}
	admin := flags.NewAdmin(store)

	listed, err := admin.List(t.Context())
	require.NoError(t, err)
	applied, err := admin.ApplyPreset(t.Context(), "demo_core")
	require.NoError(t, err)
	presets, err := admin.Presets(t.Context())
	require.NoError(t, err)

	assert.Equal(t, want, listed)
	assert.Equal(t, want, applied)
	assert.Equal(t, []flags.Preset{{Name: "demo_core"}}, presets)
	assert.Equal(t, []string{"list", "apply demo_core", "presets"}, store.calls)
}

func TestAdminWrapsStoreErrors(t *testing.T) {
	store := &fakeStore{err: flags.ErrUnknownFlag}
	admin := flags.NewAdmin(store)

	_, err := admin.SetEnabled(t.Context(), "module_ghost", true)

	assert.ErrorIs(t, err, flags.ErrUnknownFlag)
}
