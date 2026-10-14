// Package flagshttp serves GET /v1/flags (operation listEnabledFlags): the
// keys of the enabled modules, so the client can hide what is OFF. It is UX
// only; the flag gate (adapter/http/flaggate) is the security boundary.
package flagshttp

import (
	"context"
	"log/slog"

	"github.com/romodevx/control-contable/backend/adapter/http/gen"
	"github.com/romodevx/control-contable/backend/internal/platform/apperr"
	"github.com/romodevx/control-contable/backend/internal/platform/requestid"
)

// Lister returns the enabled module keys, sorted. *flags.Service implements it.
type Lister interface {
	EnabledModules(ctx context.Context) ([]string, error)
}

// Handler implements the listEnabledFlags operation of gen.StrictServerInterface.
// The composition root embeds it in the single strict server.
type Handler struct {
	flags  Lister
	logger *slog.Logger
}

// New returns a Handler. A nil logger discards.
func New(flags Lister, logger *slog.Logger) *Handler {
	if logger == nil {
		logger = slog.New(slog.DiscardHandler)
	}
	return &Handler{flags: flags, logger: logger}
}

// ListEnabledFlags answers 200 with the enabled module keys (an empty array
// when none is ON), or the 500 envelope when the flags cannot be read.
func (h *Handler) ListEnabledFlags(ctx context.Context, _ gen.ListEnabledFlagsRequestObject) (gen.ListEnabledFlagsResponseObject, error) {
	keys, err := h.flags.EnabledModules(ctx)
	if err != nil {
		h.logger.ErrorContext(ctx, "list enabled flags", slog.Any("error", err))
		return internalError(ctx), nil
	}
	enabled := make([]gen.ModuleKey, 0, len(keys))
	enabled = append(enabled, keys...)
	return gen.ListEnabledFlags200JSONResponse{Body: gen.EnabledFlags{Enabled: enabled}}, nil
}

func internalError(ctx context.Context) gen.ListEnabledFlags500JSONResponse {
	var body gen.ErrorResponse
	body.Error.Code = gen.ErrorCode(apperr.CodeInternalError)
	body.Error.Message = apperr.InternalMessage
	body.Error.RequestId = requestid.From(ctx)
	return gen.ListEnabledFlags500JSONResponse{InternalErrorJSONResponse: gen.InternalErrorJSONResponse{Body: body}}
}
