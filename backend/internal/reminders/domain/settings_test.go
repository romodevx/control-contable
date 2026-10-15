package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/reminders/domain"
)

func sendTime(t *testing.T, raw string) domain.SendTime {
	t.Helper()
	st, err := domain.ParseSendTime(raw)
	require.NoError(t, err)
	return st
}

// MODULE-reminders §2.2: the defaults GET answers until the first save.
func TestDefaultSettings(t *testing.T) {
	got := domain.DefaultSettings()

	assert.False(t, got.Enabled)
	assert.Equal(t, 14, got.EarlyDay)
	assert.True(t, got.OnDueDate)
	assert.Equal(t, "10:00", got.SendTime.String())
	assert.False(t, got.ShowAmounts, "amounts are a privacy opt-in, never defaulted on")
}

func TestParseSendTime(t *testing.T) {
	for _, tc := range []struct {
		raw          string
		hour, minute int
	}{
		{"00:00", 0, 0}, {"10:00", 10, 0}, {"08:30", 8, 30}, {"23:59", 23, 59},
	} {
		t.Run(tc.raw, func(t *testing.T) {
			got := sendTime(t, tc.raw)

			assert.Equal(t, tc.hour, got.Hour())
			assert.Equal(t, tc.minute, got.Minute())
			assert.Equal(t, tc.raw, got.String())
		})
	}
}

func TestParseSendTime_refusesAnythingButHHMM(t *testing.T) {
	for _, raw := range []string{"", "8:30", "24:00", "23:60", "10:00:00", "10-00", "１0:00", " 10:00", "ab:cd", "99:99"} {
		t.Run(raw, func(t *testing.T) {
			_, err := domain.ParseSendTime(raw)

			require.ErrorIs(t, err, domain.ErrInvalidSendTime)
		})
	}
}

func TestNewSettings(t *testing.T) {
	tenAM := sendTime(t, "10:00")
	tests := []struct {
		name      string
		enabled   bool
		earlyDay  int
		onDueDate bool
		wantErr   error
	}{
		{"on with both reminders", true, 14, true, nil},
		{"on, only the due date", true, domain.NoEarlyReminder, true, nil},
		{"on, only the early day", true, 1, false, nil},
		{"early day 16 is the last before day 17", true, 16, false, nil},
		{"off may select nothing", false, domain.NoEarlyReminder, false, nil},
		{"on and nothing selected (422 no_reminder_selected)", true, domain.NoEarlyReminder, false, domain.ErrNoReminderSelected},
		{"early day 17", true, 17, true, domain.ErrEarlyDayOutOfRange},
		{"negative early day", false, -1, true, domain.ErrEarlyDayOutOfRange},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := domain.NewSettings(tc.enabled, tc.earlyDay, tc.onDueDate, tenAM, true)

			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, domain.Settings{
				Enabled: tc.enabled, EarlyDay: tc.earlyDay, OnDueDate: tc.onDueDate, SendTime: tenAM, ShowAmounts: true,
			}, got)
		})
	}
}

func TestNewSettings_refusesTheZeroSendTime(t *testing.T) {
	_, err := domain.NewSettings(false, 14, true, domain.SendTime{}, false)

	require.ErrorIs(t, err, domain.ErrInvalidSendTime)
}

func TestSettings_HasEarlyReminder(t *testing.T) {
	assert.True(t, domain.DefaultSettings().HasEarlyReminder())
	s := domain.DefaultSettings()
	s.EarlyDay = domain.NoEarlyReminder
	assert.False(t, s.HasEarlyReminder())
}

// MODULE-reminders §2.5, contract ReminderSettings.show_amounts: the saved
// choice is kept as is, but amounts reach a notification only when the user
// opted in AND the tax engine is effectively ON.
func TestSettings_AmountsShown(t *testing.T) {
	for _, tc := range []struct {
		name               string
		showAmounts, engOn bool
		want               bool
	}{
		{"opted in, engine ON", true, true, true},
		{"opted in, engine OFF", true, false, false},
		{"not opted in, engine ON", false, true, false},
		{"not opted in, engine OFF", false, false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := domain.DefaultSettings()
			s.ShowAmounts = tc.showAmounts

			assert.Equal(t, tc.want, s.AmountsShown(tc.engOn))
		})
	}
}
