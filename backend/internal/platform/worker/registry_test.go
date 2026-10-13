package worker_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/platform/worker"
)

// stubJob is a Job with fixed metadata and a no-op Run.
type stubJob struct {
	name     string
	interval time.Duration
}

func (j stubJob) Name() string                     { return j.name }
func (j stubJob) Interval() time.Duration          { return j.interval }
func (stubJob) Run(context.Context) error          { return nil }
func job(name string, every time.Duration) stubJob { return stubJob{name: name, interval: every} }

func TestRegistry_Register_rejectsInvalidJobs(t *testing.T) {
	tests := []struct {
		name    string
		job     worker.Job
		timeout time.Duration
	}{
		{"nil job", nil, time.Second},
		{"empty name", job("", time.Minute), time.Second},
		{"name not snake_case", job("Purge-Records", time.Minute), time.Second},
		{"zero interval", job("purge", 0), time.Second},
		{"negative interval", job("purge", -time.Minute), time.Second},
		{"zero timeout", job("purge", time.Minute), 0},
		{"timeout longer than the interval", job("purge", time.Minute), 2 * time.Minute},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var r worker.Registry

			err := r.Register(tt.job, tt.timeout)

			require.ErrorIs(t, err, worker.ErrInvalidJob)
			assert.Empty(t, r.Names())
		})
	}
}

func TestRegistry_Register_rejectsADuplicateName(t *testing.T) {
	var r worker.Registry
	require.NoError(t, r.Register(job("purge", time.Minute), time.Second))

	err := r.Register(job("purge", time.Hour), time.Second)

	require.ErrorIs(t, err, worker.ErrInvalidJob)
	assert.Equal(t, []string{"purge"}, r.Names())
}

func TestRegistry_Names_keepsRegistrationOrder(t *testing.T) {
	var r worker.Registry
	require.NoError(t, r.Register(job("reminders", 5*time.Minute), time.Minute))
	require.NoError(t, r.Register(job("purge_soft_deleted", time.Hour), time.Hour))

	assert.Equal(t, []string{"reminders", "purge_soft_deleted"}, r.Names())
}
