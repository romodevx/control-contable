// Package remindershttp is the reminders module's HTTP adapter: it
// translates between the contract (docs/api/openapi.yaml, module_reminders)
// and the reminders use cases: getTaxCalendar (M9-04), registerDevice,
// unregisterDevice, getReminderSettings and putReminderSettings (M9-05).
package remindershttp

import (
	"context"
	"fmt"
	"time"

	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/romodevx/control-contable/backend/adapter/http/gen"
	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/principal"
	"github.com/romodevx/control-contable/backend/internal/platform/requestid"
	"github.com/romodevx/control-contable/backend/internal/reminders/domain"
	"github.com/romodevx/control-contable/backend/internal/reminders/usecase"
)

const (
	// bearerChallenge is the contract's fixed WWW-Authenticate value
	// (components.headers.WWWAuthenticate).
	bearerChallenge = `Bearer realm="control-contable"` //nolint:gosec // G101: an auth challenge, not a credential
	// msgAuthenticationRequired is the contract's Unauthorized message.
	msgAuthenticationRequired = "Authentication required."
	// taxCalendarCacheControl is the 200's Cache-Control (contract 0.13.0):
	// the same public data for every user, behind authentication, so only
	// the app's own cache may keep it, for an hour. Errors stay no-store
	// (middleware.NoStore).
	taxCalendarCacheControl = "private, max-age=3600"
)

// Calendar is the tax calendar use case (usecase.Calendar).
type Calendar interface {
	Year(ctx context.Context, year int) (usecase.Year, error)
}

// Devices is the devices use case (usecase.Devices).
type Devices interface {
	Register(ctx context.Context, userID string, r usecase.Registration) (usecase.Registered, error)
	Unregister(ctx context.Context, userID, installationID string) error
}

// Settings is the reminder settings use case (usecase.Settings).
type Settings interface {
	Get(ctx context.Context, userID string) (domain.Settings, error)
	Put(ctx context.Context, userID string, in usecase.SettingsInput) (domain.Settings, error)
}

// UseCases are the module's use cases the handler serves.
type UseCases struct {
	Calendar Calendar
	Devices  Devices
	Settings Settings
}

// Handler implements the module_reminders operations of
// gen.StrictServerInterface. cmd/api embeds it in the API server. The user
// id comes from the principal the router's authenticator verified; use case
// errors are client-ready *apperr.Error values returned as is (the router
// writes the envelope). The push token goes to the use case only: it is
// never logged, echoed or returned.
type Handler struct {
	calendar Calendar
	devices  Devices
	settings Settings
}

// NewHandler returns the handler over the use cases.
func NewHandler(uc UseCases) *Handler {
	return &Handler{calendar: uc.Calendar, devices: uc.Devices, settings: uc.Settings}
}

// GetTaxCalendar implements gen.StrictServerInterface (GET
// /v1/tax-calendar). The router has validated year against the contract.
func (h *Handler) GetTaxCalendar(ctx context.Context, req gen.GetTaxCalendarRequestObject) (gen.GetTaxCalendarResponseObject, error) {
	if _, err := principal.UserID(ctx); err != nil {
		return gen.GetTaxCalendar401JSONResponse{UnauthorizedJSONResponse: unauthorized(ctx)}, nil
	}
	year, err := h.calendar.Year(ctx, req.Params.Year)
	if err != nil {
		return nil, err
	}
	body, err := toCalendar(year)
	if err != nil {
		return nil, err
	}
	return gen.GetTaxCalendar200JSONResponse{
		Body:    body,
		Headers: gen.GetTaxCalendar200ResponseHeaders{CacheControl: taxCalendarCacheControl},
	}, nil
}

func toCalendar(y usecase.Year) (gen.TaxCalendar, error) {
	entries := make([]gen.TaxCalendarEntry, 0, len(y.Entries))
	for _, e := range y.Entries {
		nominal, err := contractDate(e.Nominal)
		if err != nil {
			return gen.TaxCalendar{}, err
		}
		due, err := contractDate(e.Due)
		if err != nil {
			return gen.TaxCalendar{}, err
		}
		entries = append(entries, gen.TaxCalendarEntry{
			Period:                e.Period,
			Obligation:            gen.MonthlyIsrIva,
			NominalDueDate:        nominal,
			DueDate:               due,
			MovedReason:           movedReason(e.MovedReason),
			SubjectToConfirmation: e.SubjectToConfirmation,
		})
	}
	return gen.TaxCalendar{Year: y.Year, Entries: entries}, nil
}

// movedReason is the contract's nullable moved_reason.
func movedReason(r fiscal.MovedReason) *gen.TaxCalendarMovedReason {
	if r == fiscal.MovedNone {
		return nil
	}
	out := gen.TaxCalendarMovedReason(r)
	return &out
}

// contractDate is d as the generated date type (format: date). The zero
// date is a use case bug.
func contractDate(d fiscal.Date) (openapi_types.Date, error) {
	t, err := time.Parse(time.DateOnly, d.String())
	if err != nil {
		return openapi_types.Date{}, fmt.Errorf("reminders: calendar date %q: %w", d, err)
	}
	return openapi_types.Date{Time: t}, nil
}

// unauthorized is the contract's Unauthorized response: the generic 401
// with the bearer challenge.
func unauthorized(ctx context.Context) gen.UnauthorizedJSONResponse {
	var body gen.ErrorResponse
	body.Error.Code = gen.ErrorCodeUnauthorized
	body.Error.Message = msgAuthenticationRequired
	body.Error.RequestId = requestid.From(ctx)
	return gen.UnauthorizedJSONResponse{Body: body, Headers: gen.UnauthorizedResponseHeaders{WWWAuthenticate: bearerChallenge}}
}
