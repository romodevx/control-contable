//go:build integration

package infra_test

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/auth/domain"
	"github.com/romodevx/control-contable/backend/internal/platform/pgtest"
)

// queuedEmails counts the outbox rows as the owner (FORCE RLS lifted for a
// transaction that is rolled back: the owner has no policy of its own).
func (s *stack) queuedEmails(t *testing.T) int {
	t.Helper()
	tx, err := s.owner.BeginTx(t.Context(), nil)
	require.NoError(t, err)
	defer func() { _ = tx.Rollback() }()
	_, err = tx.ExecContext(t.Context(), "ALTER TABLE auth_email_outbox NO FORCE ROW LEVEL SECURITY")
	require.NoError(t, err)
	var n int
	require.NoError(t, tx.QueryRowContext(t.Context(), "SELECT count(*) FROM auth_email_outbox").Scan(&n))
	return n
}

// registerOnly runs register as the request does: no delivery.
func (s *stack) registerOnly(t *testing.T, email string) {
	t.Helper()
	require.NoError(t, s.register.execute(t.Context(), registration(email)))
}

// The timing oracle closed (MODULE-auth §4): register answers a new email and
// a verified account alike without sending anything in the request. Each
// queues one sealed row; the worker sends the code or the notice afterwards.
func TestOutbox_registerSendsNothingInTheRequest(t *testing.T) {
	s := newStack(t)
	s.signUp(t, anaEmail, firstSample)
	s.inbox.forget()

	s.registerOnly(t, "nuevo@example.mx")
	s.registerOnly(t, anaEmail)

	assert.Empty(t, s.inbox.codes, "no code sent in the request")
	assert.Empty(t, s.inbox.notices, "no notice sent in the request")
	assert.Equal(t, 2, s.queuedEmails(t), "both queued")

	require.NoError(t, s.delivery.Run(t.Context()))

	assert.Len(t, s.inbox.codes["nuevo@example.mx"], 1)
	assert.Equal(t, 1, s.inbox.notices[anaEmail])
	assert.Zero(t, s.queuedEmails(t), "deleted on delivery")
}

func TestOutbox_resendSendsNothingInTheRequest(t *testing.T) {
	s := newStack(t)
	s.registerOnly(t, anaEmail)
	s.ownerExec(t, pastCooldown)

	require.NoError(t, s.resend.execute(t.Context(), anaEmail))

	assert.Empty(t, s.inbox.codes)
	assert.Equal(t, 2, s.queuedEmails(t))
	require.NoError(t, s.delivery.Run(t.Context()))
	assert.Len(t, s.inbox.codes[anaEmail], 2)
}

// The payload is sealed: the row shows neither the address nor the code.
func TestOutbox_rowsHoldNoClearPII(t *testing.T) {
	s := newStack(t)
	s.registerOnly(t, anaEmail)

	tx, err := s.owner.BeginTx(t.Context(), nil)
	require.NoError(t, err)
	defer func() { _ = tx.Rollback() }()
	_, err = tx.ExecContext(t.Context(), "ALTER TABLE auth_email_outbox NO FORCE ROW LEVEL SECURITY")
	require.NoError(t, err)
	var row string
	require.NoError(t, tx.QueryRowContext(t.Context(),
		"SELECT row_to_json(o)::text || encode(o.payload, 'escape') FROM auth_email_outbox o").Scan(&row))
	require.NoError(t, tx.Rollback())

	assert.NotContains(t, row, "example.mx")
	require.NoError(t, s.delivery.Run(t.Context()))
	assert.NotContains(t, row, s.inbox.lastCode(t, anaEmail))
}

// countingInbox counts every send per address, across workers.
type countingInbox struct {
	mu    sync.Mutex
	sends map[string]int
}

func (c *countingInbox) SendVerificationCode(_ context.Context, to domain.Email, _ domain.VerificationCode) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.sends[to.String()]++
	return nil
}

func (c *countingInbox) SendSignUpAttemptNotice(_ context.Context, to domain.Email) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.sends[to.String()]++
	return nil
}

// Worker split brain: two instances (two cc_worker pools) running the job at
// the same time deliver every queued email exactly once.
func TestOutbox_twoWorkersDeliverEachEmailOnce(t *testing.T) {
	s := newStack(t)
	const queued = 60
	for i := range queued {
		s.registerOnly(t, fmt.Sprintf("persona%02d@example.mx", i))
	}
	require.Equal(t, queued, s.queuedEmails(t))
	inbox := &countingInbox{sends: map[string]int{}}
	logger := slog.New(slog.DiscardHandler)
	jobs := []interface{ Run(context.Context) error }{
		s.newDeliveryJob(t, s.worker, inbox, logger),
		s.newDeliveryJob(t, s.newWorkerPool(t), inbox, logger),
	}

	var wg sync.WaitGroup
	start := make(chan struct{})
	for _, job := range jobs {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			for range 3 {
				assert.NoError(t, job.Run(context.Background()))
			}
		}()
	}
	close(start)
	wg.Wait()

	assert.Len(t, inbox.sends, queued, "every email delivered")
	for to, n := range inbox.sends {
		assert.Equal(t, 1, n, "%s delivered once", to)
	}
	assert.Zero(t, s.queuedEmails(t))
}

// newWorkerPool is a second cc_worker pool on the stack's database: another
// worker instance.
func (s *stack) newWorkerPool(t *testing.T) *sql.DB {
	t.Helper()
	return s.db.Connect(t, pgtest.RoleWorker)
}
