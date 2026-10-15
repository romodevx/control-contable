package usecase_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/platform/apperr"
	"github.com/romodevx/control-contable/backend/internal/reminders/domain"
	"github.com/romodevx/control-contable/backend/internal/reminders/usecase"
)

func newSettingsUseCase(t *testing.T) (*usecase.Settings, *fakeTx, *fakeSettings) {
	t.Helper()
	txm, store := &fakeTx{}, &fakeSettings{}
	uc, err := usecase.NewSettings(usecase.SettingsDeps{Tx: txm, Store: store})
	require.NoError(t, err)
	return uc, txm, store
}

func intPtr(n int) *int { return &n }

func TestNewSettings_needsEveryPort(t *testing.T) {
	for name, deps := range map[string]usecase.SettingsDeps{
		"Tx":    {Store: &fakeSettings{}},
		"Store": {Tx: &fakeTx{}},
	} {
		t.Run(name, func(t *testing.T) {
			_, err := usecase.NewSettings(deps)

			require.ErrorIs(t, err, usecase.ErrMissingDependency)
		})
	}
}

// MODULE-reminders §2.2: a user who never saved gets the defaults, read in a
// read-only unit of work (no row is created by a read).
func TestGetSettings_defaultsWhenNeverSaved(t *testing.T) {
	uc, txm, store := newSettingsUseCase(t)

	got, err := uc.Get(t.Context(), userA)

	require.NoError(t, err)
	assert.Equal(t, domain.DefaultSettings(), got)
	assert.Equal(t, []string{"ro"}, txm.modes)
	assert.Equal(t, userA, store.txUser)
	assert.Equal(t, userA, store.readFor)
	assert.Nil(t, store.saved, "a read writes nothing")
}

func TestGetSettings_theSavedOnes(t *testing.T) {
	uc, _, store := newSettingsUseCase(t)
	saved := domain.DefaultSettings()
	saved.Enabled, saved.EarlyDay, saved.ShowAmounts = true, domain.NoEarlyReminder, true
	store.saved = &saved

	got, err := uc.Get(t.Context(), userA)

	require.NoError(t, err)
	assert.Equal(t, saved, got)
}

func TestGetSettings_storeFailure(t *testing.T) {
	uc, _, store := newSettingsUseCase(t)
	store.readErr = errPortDown

	_, err := uc.Get(t.Context(), userA)

	require.ErrorIs(t, err, errPortDown)
}

func optIn() usecase.SettingsInput {
	return usecase.SettingsInput{Enabled: true, EarlyDay: intPtr(14), OnDueDate: true, SendTime: "08:30", ShowAmounts: false}
}

// PUT replaces the whole object as the caller, in a read-write unit of
// work, and answers what was saved. show_amounts is stored as chosen,
// whatever module_tax_engine's state (the job applies it, §2.5).
func TestPutSettings_savesTheWholeObject(t *testing.T) {
	tests := map[string]struct {
		in   usecase.SettingsInput
		want func(t *testing.T) domain.Settings
	}{
		"opt in": {optIn(), func(t *testing.T) domain.Settings {
			return mustSettings(t, true, 14, true, "08:30", false)
		}},
		"only the due date, amounts on": {
			usecase.SettingsInput{Enabled: true, EarlyDay: nil, OnDueDate: true, SendTime: "10:00", ShowAmounts: true},
			func(t *testing.T) domain.Settings {
				return mustSettings(t, true, domain.NoEarlyReminder, true, "10:00", true)
			},
		},
		"off and nothing selected": {
			usecase.SettingsInput{Enabled: false, EarlyDay: nil, OnDueDate: false, SendTime: "23:59"},
			func(t *testing.T) domain.Settings {
				return mustSettings(t, false, domain.NoEarlyReminder, false, "23:59", false)
			},
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			uc, txm, store := newSettingsUseCase(t)

			got, err := uc.Put(t.Context(), userA, tc.in)

			require.NoError(t, err)
			want := tc.want(t)
			assert.Equal(t, want, got)
			require.NotNil(t, store.saved)
			assert.Equal(t, want, *store.saved)
			assert.Equal(t, userA, store.savedFor)
			assert.Equal(t, userA, store.txUser)
			assert.Equal(t, []string{"rw"}, txm.modes)
		})
	}
}

func mustSettings(t *testing.T, enabled bool, earlyDay int, onDue bool, at string, show bool) domain.Settings {
	t.Helper()
	st, err := domain.ParseSendTime(at)
	require.NoError(t, err)
	s, err := domain.NewSettings(enabled, earlyDay, onDue, st, show)
	require.NoError(t, err)
	return s
}

// §6: the schema backstops (400) and the one business rule (422).
func TestPutSettings_refusals(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*usecase.SettingsInput)
		code   apperr.Code
		field  string
		detail string
	}{
		{"early day 0 is not null", func(in *usecase.SettingsInput) { in.EarlyDay = intPtr(0) },
			apperr.CodeValidationFailed, "early_day", "minimum"},
		{"early day 17", func(in *usecase.SettingsInput) { in.EarlyDay = intPtr(17) },
			apperr.CodeValidationFailed, "early_day", "maximum"},
		{"send time without minutes", func(in *usecase.SettingsInput) { in.SendTime = "10" },
			apperr.CodeValidationFailed, "send_time", "pattern"},
		{"enabled with no reminder selected", func(in *usecase.SettingsInput) { in.EarlyDay, in.OnDueDate = nil, false },
			apperr.CodeBusinessRuleViolation, "body", usecase.DetailNoReminderSelected},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			uc, txm, store := newSettingsUseCase(t)
			in := optIn()
			tc.mutate(&in)

			_, err := uc.Put(t.Context(), userA, in)

			requireAppError(t, err, tc.code, tc.field, tc.detail)
			assert.Nil(t, store.saved)
			assert.Empty(t, txm.modes, "refused before any unit of work")
		})
	}
}

func TestPutSettings_storeFailure(t *testing.T) {
	uc, _, store := newSettingsUseCase(t)
	store.saveErr = errPortDown

	_, err := uc.Put(t.Context(), userA, optIn())

	require.ErrorIs(t, err, errPortDown)
}
