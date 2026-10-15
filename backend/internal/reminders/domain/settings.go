// Package domain holds the reminders module's pure rules (MODULE-reminders
// §2.2, §2.3): the reminder settings and the device registration values.
// It imports nothing but the standard library.
package domain

import (
	"errors"
	"fmt"
)

var (
	// ErrEarlyDayOutOfRange reports an early reminder day outside 1-16.
	ErrEarlyDayOutOfRange = errors.New("reminders: early day out of range")
	// ErrInvalidSendTime reports a send time that is not HH:MM.
	ErrInvalidSendTime = errors.New("reminders: send time is not HH:MM")
	// ErrNoReminderSelected reports settings that are on and would send
	// nothing (422 no_reminder_selected).
	ErrNoReminderSelected = errors.New("reminders: enabled with no reminder selected")
)

const (
	// NoEarlyReminder is the EarlyDay of settings without an early reminder
	// (the contract's early_day null).
	NoEarlyReminder = 0
	// MinEarlyDay is the first day an early reminder may be set to.
	MinEarlyDay = 1
	// MaxEarlyDay is the last: always before the due date's day 17.
	MaxEarlyDay = 16

	defaultEarlyDay   = 14
	defaultSendHour   = 10
	hoursPerDay       = 24
	minutesPerHour    = 60
	sendTimeLength    = len("HH:MM")
	sendTimeSeparator = ':'
)

// SendTime is a minute of the day, America/Mexico_City (every user, §2.1).
// The zero value is not a valid send time: build one with ParseSendTime.
type SendTime struct {
	minuteOfDay int // 1 + minutes since midnight, so 0 means unset
}

// ParseSendTime parses "HH:MM" (00:00-23:59), exactly five ASCII
// characters.
func ParseSendTime(raw string) (SendTime, error) {
	if len(raw) != sendTimeLength || raw[2] != sendTimeSeparator {
		return SendTime{}, fmt.Errorf("%w: %q", ErrInvalidSendTime, raw)
	}
	hour, okHour := twoDigits(raw[0], raw[1])
	minute, okMinute := twoDigits(raw[3], raw[4])
	if !okHour || !okMinute || hour >= hoursPerDay || minute >= minutesPerHour {
		return SendTime{}, fmt.Errorf("%w: %q", ErrInvalidSendTime, raw)
	}
	return SendTime{minuteOfDay: 1 + hour*minutesPerHour + minute}, nil
}

func twoDigits(tens, units byte) (int, bool) {
	if tens < '0' || tens > '9' || units < '0' || units > '9' {
		return 0, false
	}
	return int(tens-'0')*10 + int(units-'0'), true
}

// IsZero reports the unset send time.
func (t SendTime) IsZero() bool { return t.minuteOfDay == 0 }

// Hour is the send time's hour, 0-23.
func (t SendTime) Hour() int { return (t.minuteOfDay - 1) / minutesPerHour }

// Minute is the send time's minute, 0-59.
func (t SendTime) Minute() int { return (t.minuteOfDay - 1) % minutesPerHour }

// String is "HH:MM".
func (t SendTime) String() string { return fmt.Sprintf("%02d:%02d", t.Hour(), t.Minute()) }

// Settings are a user's reminder preferences (§2.2).
type Settings struct {
	Enabled bool
	// EarlyDay is the nominal day of the due month for the early reminder,
	// MinEarlyDay-MaxEarlyDay, or NoEarlyReminder.
	EarlyDay  int
	OnDueDate bool
	SendTime  SendTime
	// ShowAmounts is the lock-screen privacy opt-in, stored as chosen; see
	// AmountsShown.
	ShowAmounts bool
}

// DefaultSettings are the settings of a user who never saved (§2.2).
func DefaultSettings() Settings {
	return Settings{
		Enabled:   false,
		EarlyDay:  defaultEarlyDay,
		OnDueDate: true,
		SendTime:  SendTime{minuteOfDay: 1 + defaultSendHour*minutesPerHour},
	}
}

// NewSettings validates a whole settings object (PUT replaces it): the
// early day in range or none, a send time, and, when enabled, at least one
// reminder selected.
func NewSettings(enabled bool, earlyDay int, onDueDate bool, sendTime SendTime, showAmounts bool) (Settings, error) {
	if earlyDay != NoEarlyReminder && (earlyDay < MinEarlyDay || earlyDay > MaxEarlyDay) {
		return Settings{}, fmt.Errorf("%w: %d", ErrEarlyDayOutOfRange, earlyDay)
	}
	if sendTime.IsZero() {
		return Settings{}, ErrInvalidSendTime
	}
	if enabled && earlyDay == NoEarlyReminder && !onDueDate {
		return Settings{}, ErrNoReminderSelected
	}
	return Settings{Enabled: enabled, EarlyDay: earlyDay, OnDueDate: onDueDate, SendTime: sendTime, ShowAmounts: showAmounts}, nil
}

// HasEarlyReminder reports whether an early reminder is selected.
func (s Settings) HasEarlyReminder() bool { return s.EarlyDay != NoEarlyReminder }

// AmountsShown reports whether a notification may carry amounts (§2.5):
// only when the user opted in and the tax engine is effectively ON. With
// the engine OFF the choice is kept (and the app hides the toggle), but no
// amount is ever sent.
func (s Settings) AmountsShown(taxEngineOn bool) bool { return s.ShowAmounts && taxEngineOn }
