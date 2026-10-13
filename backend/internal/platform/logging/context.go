package logging

import (
	"context"
	"log/slog"
	"slices"

	"github.com/romodevx/control-contable/backend/internal/platform/requestid"
)

// contextHandler adds the request id of the context to each record, at the
// top level even after WithGroup. It remembers the With* calls applied to it
// (ops) and, for a grouped handler, rebuilds base+request id+ops per record;
// an ungrouped handler just appends the attribute to the record.
type contextHandler struct {
	base    slog.Handler                      // the JSON handler, before any With*
	ops     []func(slog.Handler) slog.Handler // With* calls, in order
	built   slog.Handler                      // base with ops applied
	grouped bool
}

func (h *contextHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.built.Enabled(ctx, level)
}

func (h *contextHandler) Handle(ctx context.Context, r slog.Record) error {
	id := requestid.From(ctx)
	if id == "" {
		return h.built.Handle(ctx, r)
	}
	attr := slog.String(RequestIDKey, id)
	if !h.grouped {
		r = r.Clone()
		r.AddAttrs(attr)
		return h.built.Handle(ctx, r)
	}
	handler := h.base.WithAttrs([]slog.Attr{attr})
	for _, op := range h.ops {
		handler = op(handler)
	}
	return handler.Handle(ctx, r)
}

func (h *contextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h.with(func(next slog.Handler) slog.Handler { return next.WithAttrs(attrs) }, false)
}

func (h *contextHandler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}
	return h.with(func(next slog.Handler) slog.Handler { return next.WithGroup(name) }, true)
}

func (h *contextHandler) with(op func(slog.Handler) slog.Handler, group bool) slog.Handler {
	return &contextHandler{
		base:    h.base,
		ops:     append(slices.Clone(h.ops), op),
		built:   op(h.built),
		grouped: h.grouped || group,
	}
}
