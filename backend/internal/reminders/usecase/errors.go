package usecase

import (
	"errors"

	"github.com/romodevx/control-contable/backend/internal/platform/apperr"
)

var (
	// ErrDeviceNotFound reports a device that is unknown or another user's.
	ErrDeviceNotFound = errors.New("reminders: device not found")
	// ErrClaimContention reports a claim that lost a documented race in
	// claim_push_token (40001 or 23505): the unit of work is retried.
	ErrClaimContention = errors.New("reminders: concurrent device claim")
)

// DetailNoReminderSelected is the details[].code of settings that are on
// and would send nothing (contract putReminderSettings, 422).
const DetailNoReminderSelected = "no_reminder_selected"

// Contract fields and the JSON Schema keywords the backstops answer (the
// router validates the contract first).
const (
	fieldInstallationID = "id"
	fieldPlatform       = "platform"
	fieldPushToken      = "push_token"
	fieldAppVersion     = "app_version"
	fieldEarlyDay       = "early_day"
	fieldSendTime       = "send_time"
	fieldBody           = "body"

	detailFormat  = "format"
	detailEnum    = "enum"
	detailPattern = "pattern"
	detailMinimum = "minimum"
	detailMaximum = "maximum"
)

// Client-safe messages (es-MX for the fields). Values the client sent are
// never echoed: the push token above all.
const (
	msgInvalidInput        = "One or more fields are invalid." // the router's wording for every validation_failed
	msgNotFound            = "Not found."
	msgInvalidFormat       = "Formato no válido."
	msgInvalidPlatform     = "Plataforma no válida."
	msgEarlyDayMinimum     = "El día debe ser 1 o mayor."
	msgEarlyDayMaximum     = "El día debe ser 16 o menor."
	msgSendTimePattern     = "Usa el formato HH:MM."
	msgNoReminderSelected  = "Enabled reminders must select at least one reminder."
	msgNoReminderSelectedF = "Elige al menos un recordatorio o desactívalos."
)

func invalidField(cause error, field, detail, message string) error {
	return apperr.Wrap(cause, apperr.CodeValidationFailed, msgInvalidInput).
		WithFields(apperr.FieldError{Field: field, Code: detail, Message: message})
}

// notFound is the hidden 404: the same envelope as the module OFF (§3).
func notFound(cause error) error {
	return apperr.Wrap(errors.Join(ErrDeviceNotFound, cause), apperr.CodeNotFound, msgNotFound)
}
