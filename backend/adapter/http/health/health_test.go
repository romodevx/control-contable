package health_test

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/adapter/http/gen"
	"github.com/romodevx/control-contable/backend/adapter/http/health"
	"github.com/romodevx/control-contable/backend/internal/platform/logging"
)

const version = "1.2.3"

type pingFunc func(context.Context) error

func (f pingFunc) Ping(ctx context.Context) error { return f(ctx) }

func TestGetHealth(t *testing.T) {
	tests := []struct {
		name string
		ping pingFunc
		want gen.GetHealth200JSONResponse
	}{
		{
			name: "database reachable",
			ping: func(context.Context) error { return nil },
			want: gen.GetHealth200JSONResponse{Status: gen.HealthStatusStatusOk, Database: gen.HealthStatusDatabaseOk, Version: version},
		},
		{
			name: "database down: alive but degraded",
			ping: func(context.Context) error { return errors.New("dial tcp: connection refused") },
			want: gen.GetHealth200JSONResponse{Status: gen.HealthStatusStatusDegraded, Database: gen.HealthStatusDatabaseUnreachable, Version: version},
		},
		{
			name: "database hangs: the ping is bounded",
			ping: func(ctx context.Context) error {
				<-ctx.Done()
				return ctx.Err()
			},
			want: gen.GetHealth200JSONResponse{Status: gen.HealthStatusStatusDegraded, Database: gen.HealthStatusDatabaseUnreachable, Version: version},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := health.New(tt.ping, version, logging.New(&bytes.Buffer{}, nil), health.WithPingTimeout(10*time.Millisecond))

			got, err := h.GetHealth(context.Background(), gen.GetHealthRequestObject{})

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

// cloudRunProbeTimeout is the default timeoutSeconds of a Cloud Run HTTP
// liveness probe (ADR-0001 hosting).
const cloudRunProbeTimeout = time.Second

// A hung database must not make /healthz slower than the liveness probe, or
// the orchestrator restarts a process that is alive (the opposite of what the
// liveness contract promises).
func TestGetHealth_defaultPingFitsInsideTheLivenessProbe(t *testing.T) {
	var budget time.Duration
	h := health.New(pingFunc(func(ctx context.Context) error {
		deadline, ok := ctx.Deadline()
		require.True(t, ok, "the readiness ping must be bounded")
		budget = time.Until(deadline)
		return nil
	}), version, logging.New(&bytes.Buffer{}, nil))

	_, err := h.GetHealth(context.Background(), gen.GetHealthRequestObject{})

	require.NoError(t, err)
	assert.Less(t, budget, cloudRunProbeTimeout)
}

func TestGetHealth_logsTheCauseNotTheResponse(t *testing.T) {
	var logs bytes.Buffer
	h := health.New(pingFunc(func(context.Context) error { return errors.New("connection refused") }),
		version, logging.New(&logs, nil))

	_, err := h.GetHealth(context.Background(), gen.GetHealthRequestObject{})

	require.NoError(t, err)
	assert.Contains(t, logs.String(), "connection refused")
}
