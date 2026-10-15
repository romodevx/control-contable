package remindershttp

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"github.com/romodevx/control-contable/backend/adapter/http/gen"
	"github.com/romodevx/control-contable/backend/internal/platform/apperr"
	"github.com/romodevx/control-contable/backend/internal/platform/principal"
	"github.com/romodevx/control-contable/backend/internal/reminders/domain"
	"github.com/romodevx/control-contable/backend/internal/reminders/usecase"
)

// errMissingInput is the cause of a backstop 400: the contract validation
// answers first, so reaching it means the router let a body through.
var errMissingInput = errors.New("reminders: required input missing")

const (
	msgInvalidInput  = "One or more fields are invalid." // the router's wording for every validation_failed
	msgRequiredField = "Este campo es obligatorio."
	detailRequired   = "required"
	fieldBody        = "body"
	fieldPushToken   = "push_token"
)

// RegisterDevice implements gen.StrictServerInterface (POST /v1/devices):
// 201 for a new or claimed installation, 200 for the caller's own refreshed
// one, the Device (no token) either way.
func (h *Handler) RegisterDevice(ctx context.Context, req gen.RegisterDeviceRequestObject) (gen.RegisterDeviceResponseObject, error) {
	userID, err := principal.UserID(ctx)
	if err != nil {
		return gen.RegisterDevice401JSONResponse{UnauthorizedJSONResponse: unauthorized(ctx)}, nil
	}
	reg, err := registration(req.Body)
	if err != nil {
		return nil, err
	}
	out, err := h.devices.Register(ctx, userID, reg)
	if err != nil {
		return nil, err
	}
	device, err := toDevice(out.Device)
	if err != nil {
		return nil, err
	}
	if out.Created {
		return gen.RegisterDevice201JSONResponse{Body: device}, nil
	}
	return gen.RegisterDevice200JSONResponse{Body: device}, nil
}

// registration maps DeviceCreate onto the use case's input. push_token is
// a *string in the generated type (writeOnly, required): nil cannot pass
// the contract validation, so it is the 400 backstop, never a dereference.
func registration(body *gen.DeviceCreate) (usecase.Registration, error) {
	if body == nil {
		return usecase.Registration{}, required(fieldBody)
	}
	if body.PushToken == nil {
		return usecase.Registration{}, required(fieldPushToken)
	}
	return usecase.Registration{
		InstallationID: body.Id.String(),
		Platform:       string(body.Platform),
		Token:          *body.PushToken,
		AppVersion:     body.AppVersion,
	}, nil
}

func required(field string) error {
	return apperr.Wrap(errMissingInput, apperr.CodeValidationFailed, msgInvalidInput).
		WithFields(apperr.FieldError{Field: field, Code: detailRequired, Message: msgRequiredField})
}

// toDevice is the contract's Device: never the token. An id that is not a
// UUID is a store bug (a 500).
func toDevice(d domain.Device) (gen.Device, error) {
	id, err := uuid.Parse(d.ID)
	if err != nil {
		return gen.Device{}, fmt.Errorf("reminders: device id: %w", err)
	}
	out := gen.Device{Id: id, Platform: gen.DevicePlatform(d.Platform), CreatedAt: d.CreatedAt, LastSeenAt: d.LastSeenAt}
	if d.AppVersion != "" {
		version := d.AppVersion
		out.AppVersion = &version
	}
	return out, nil
}

// UnregisterDevice implements gen.StrictServerInterface (DELETE
// /v1/devices/{deviceId}): 204, or the use case's hidden 404.
func (h *Handler) UnregisterDevice(ctx context.Context, req gen.UnregisterDeviceRequestObject) (gen.UnregisterDeviceResponseObject, error) {
	userID, err := principal.UserID(ctx)
	if err != nil {
		return gen.UnregisterDevice401JSONResponse{UnauthorizedJSONResponse: unauthorized(ctx)}, nil
	}
	if err := h.devices.Unregister(ctx, userID, req.DeviceId.String()); err != nil {
		return nil, err
	}
	return gen.UnregisterDevice204Response{}, nil
}

// GetReminderSettings implements gen.StrictServerInterface (GET
// /v1/reminder-settings): the saved settings or the defaults.
func (h *Handler) GetReminderSettings(ctx context.Context, _ gen.GetReminderSettingsRequestObject) (gen.GetReminderSettingsResponseObject, error) {
	userID, err := principal.UserID(ctx)
	if err != nil {
		return gen.GetReminderSettings401JSONResponse{UnauthorizedJSONResponse: unauthorized(ctx)}, nil
	}
	s, err := h.settings.Get(ctx, userID)
	if err != nil {
		return nil, err
	}
	return gen.GetReminderSettings200JSONResponse{Body: toReminderSettings(s)}, nil
}

// PutReminderSettings implements gen.StrictServerInterface (PUT
// /v1/reminder-settings): replaces the whole object, answers what was saved.
func (h *Handler) PutReminderSettings(ctx context.Context, req gen.PutReminderSettingsRequestObject) (gen.PutReminderSettingsResponseObject, error) {
	userID, err := principal.UserID(ctx)
	if err != nil {
		return gen.PutReminderSettings401JSONResponse{UnauthorizedJSONResponse: unauthorized(ctx)}, nil
	}
	if req.Body == nil {
		return nil, required(fieldBody)
	}
	b := req.Body
	in := usecase.SettingsInput{Enabled: b.Enabled, OnDueDate: b.OnDueDate, SendTime: b.SendTime, ShowAmounts: b.ShowAmounts}
	if b.EarlyDay != nil {
		day := *b.EarlyDay
		in.EarlyDay = &day
	}
	s, err := h.settings.Put(ctx, userID, in)
	if err != nil {
		return nil, err
	}
	return gen.PutReminderSettings200JSONResponse{Body: toReminderSettings(s)}, nil
}

func toReminderSettings(s domain.Settings) gen.ReminderSettings {
	out := gen.ReminderSettings{Enabled: s.Enabled, OnDueDate: s.OnDueDate, SendTime: s.SendTime.String(), ShowAmounts: s.ShowAmounts}
	if s.HasEarlyReminder() {
		day := s.EarlyDay
		out.EarlyDay = &day
	}
	return out
}
