// Package health implements getHealth (GET /healthz), the platform's
// liveness and readiness probe.
//
// Any answer at all is liveness: the operation always returns 200 while the
// process can serve, so an orchestrator never restarts a healthy process
// because the database is down. Readiness is the body: status "ok" with
// database "ok" when a bounded ping succeeds, otherwise status "degraded" with
// database "unreachable" (the HealthStatus schema of the contract).
package health

import (
	"context"
	"log/slog"
	"time"

	"github.com/romodevx/control-contable/backend/adapter/http/gen"
)

// DefaultPingTimeout bounds the readiness ping so a hung database cannot
// hold probe requests open. It stays well under the 1 s default timeout of a
// Cloud Run liveness probe: a slower answer would fail the probe and restart
// a process that is alive while the database is not.
const DefaultPingTimeout = 500 * time.Millisecond

// Pinger is the readiness port: one cheap round trip to the database
// (pgdb.Pinger in production).
type Pinger interface {
	Ping(ctx context.Context) error
}

// Handler serves getHealth.
type Handler struct {
	db          Pinger
	version     string
	logger      *slog.Logger
	pingTimeout time.Duration
}

// Option customizes a Handler.
type Option func(*Handler)

// WithPingTimeout overrides DefaultPingTimeout.
func WithPingTimeout(d time.Duration) Option {
	return func(h *Handler) { h.pingTimeout = d }
}

// New returns a Handler reporting version and pinging db.
func New(db Pinger, version string, logger *slog.Logger, opts ...Option) *Handler {
	h := &Handler{db: db, version: version, logger: logger, pingTimeout: DefaultPingTimeout}
	for _, opt := range opts {
		opt(h)
	}
	return h
}

// GetHealth implements the getHealth operation of gen.StrictServerInterface.
func (h *Handler) GetHealth(ctx context.Context, _ gen.GetHealthRequestObject) (gen.GetHealthResponseObject, error) {
	ctx, cancel := context.WithTimeout(ctx, h.pingTimeout)
	defer cancel()
	if err := h.db.Ping(ctx); err != nil {
		h.logger.WarnContext(ctx, "readiness: database ping failed", slog.Any("error", err))
		return gen.GetHealth200JSONResponse{
			Status:   gen.HealthStatusStatusDegraded,
			Database: gen.HealthStatusDatabaseUnreachable,
			Version:  h.version,
		}, nil
	}
	return gen.GetHealth200JSONResponse{
		Status:   gen.HealthStatusStatusOk,
		Database: gen.HealthStatusDatabaseOk,
		Version:  h.version,
	}, nil
}
