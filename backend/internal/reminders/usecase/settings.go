package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/romodevx/control-contable/backend/internal/platform/apperr"
	"github.com/romodevx/control-contable/backend/internal/platform/tx"
	"github.com/romodevx/control-contable/backend/internal/reminders/domain"
)

// SettingsDeps are the ports of Settings. Store must run on Tx's units of
// work.
type SettingsDeps struct {
	Tx    tx.Manager
	Store SettingsStore
}

// Settings answers getReminderSettings and putReminderSettings
// (MODULE-reminders §2.2). Preferences, not records: no access policy, no
// 402 for a read-only user (§4). It is safe for concurrent use.
type Settings struct {
	deps SettingsDeps
}

// NewSettings returns the use case, or ErrMissingDependency.
func NewSettings(deps SettingsDeps) (*Settings, error) {
	switch {
	case deps.Tx == nil:
		return nil, fmt.Errorf("%w: SettingsDeps.Tx", ErrMissingDependency)
	case deps.Store == nil:
		return nil, fmt.Errorf("%w: SettingsDeps.Store", ErrMissingDependency)
	}
	return &Settings{deps: deps}, nil
}

// SettingsInput is putReminderSettings' body: the whole object.
type SettingsInput struct {
	Enabled bool
	// EarlyDay is nil for no early reminder (the contract's null).
	EarlyDay    *int
	OnDueDate   bool
	SendTime    string
	ShowAmounts bool
}

// Get answers the caller's saved settings, or the defaults when never saved
// (no row is created by a read).
func (s *Settings) Get(ctx context.Context, userID string) (domain.Settings, error) {
	var (
		out   domain.Settings
		found bool
	)
	err := s.deps.Tx.ReadOnly(ctx, userID, func(ctx context.Context) error {
		var err error
		out, found, err = s.deps.Store.Settings(ctx, userID)
		return err
	})
	if err != nil {
		return domain.Settings{}, fmt.Errorf("reminders: get settings: %w", err)
	}
	if !found {
		return domain.DefaultSettings(), nil
	}
	return out, nil
}

// Put replaces the caller's settings with in and answers them. Last write
// wins; saving never resends a reminder (§2.2). show_amounts is stored as
// chosen whatever module_tax_engine's state: the reminder job applies it
// (domain.Settings.AmountsShown).
func (s *Settings) Put(ctx context.Context, userID string, in SettingsInput) (domain.Settings, error) {
	settings, err := validateSettings(in)
	if err != nil {
		return domain.Settings{}, err
	}
	err = s.deps.Tx.ReadWrite(ctx, userID, func(ctx context.Context) error {
		return s.deps.Store.SaveSettings(ctx, userID, settings)
	})
	if err != nil {
		return domain.Settings{}, fmt.Errorf("reminders: put settings: %w", err)
	}
	return settings, nil
}

// validateSettings maps the body onto the domain: the schema backstops are
// 400s (the router answers them first), "on with nothing selected" is the
// 422 no_reminder_selected.
func validateSettings(in SettingsInput) (domain.Settings, error) {
	earlyDay := domain.NoEarlyReminder
	if in.EarlyDay != nil {
		switch day := *in.EarlyDay; {
		case day < domain.MinEarlyDay:
			return domain.Settings{}, invalidField(domain.ErrEarlyDayOutOfRange, fieldEarlyDay, detailMinimum, msgEarlyDayMinimum)
		case day > domain.MaxEarlyDay:
			return domain.Settings{}, invalidField(domain.ErrEarlyDayOutOfRange, fieldEarlyDay, detailMaximum, msgEarlyDayMaximum)
		default:
			earlyDay = day
		}
	}
	sendTime, err := domain.ParseSendTime(in.SendTime)
	if err != nil {
		return domain.Settings{}, invalidField(err, fieldSendTime, detailPattern, msgSendTimePattern)
	}
	settings, err := domain.NewSettings(in.Enabled, earlyDay, in.OnDueDate, sendTime, in.ShowAmounts)
	if errors.Is(err, domain.ErrNoReminderSelected) {
		return domain.Settings{}, apperr.Wrap(err, apperr.CodeBusinessRuleViolation, msgNoReminderSelected).
			WithFields(apperr.FieldError{Field: fieldBody, Code: DetailNoReminderSelected, Message: msgNoReminderSelectedF})
	}
	if err != nil {
		return domain.Settings{}, fmt.Errorf("reminders: settings: %w", err)
	}
	return settings, nil
}
