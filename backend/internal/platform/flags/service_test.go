package flags_test

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/platform/flags"
)

var errDatabaseDown = errors.New("database down")

// fakeReader is a scripted Reader: it returns keys (or err) and counts loads.
type fakeReader struct {
	mu    sync.Mutex
	keys  []string
	err   error
	loads atomic.Int64
}

func (r *fakeReader) EnabledKeys(context.Context) ([]string, error) {
	r.loads.Add(1)
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.err != nil {
		return nil, r.err
	}
	return append([]string(nil), r.keys...), nil
}

func (r *fakeReader) set(keys []string, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.keys, r.err = keys, err
}

// manualClock is a Clock the test advances by hand.
type manualClock struct {
	mu  sync.Mutex
	now time.Time
}

func (c *manualClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.now
}

func (c *manualClock) advance(d time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.now = c.now.Add(d)
}

const testTTL = 5 * time.Second

func newService(t *testing.T, reader flags.Reader) (*flags.Service, *manualClock) {
	t.Helper()
	clk := &manualClock{now: time.Date(2026, 10, 13, 12, 0, 0, 0, time.UTC)}
	return flags.NewService(reader, flags.WithClock(clk), flags.WithTTL(testTTL)), clk
}

func TestEnabled(t *testing.T) {
	tests := []struct {
		name   string
		keys   []string
		module string
		want   bool
	}{
		{"enabled module", []string{"module_auth", "module_incomes"}, "module_incomes", true},
		{"disabled module", []string{"module_auth"}, "module_incomes", false},
		{"no module enabled", nil, "module_auth", false},
		{"platform is not a flag", []string{"module_auth"}, "platform", false},
		{"empty key", []string{"module_auth"}, "", false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc, _ := newService(t, &fakeReader{keys: tc.keys})
			assert.Equal(t, tc.want, svc.Enabled(t.Context(), tc.module))
		})
	}
}

func TestEnabledModulesIsSortedAndDeduplicated(t *testing.T) {
	svc, _ := newService(t, &fakeReader{keys: []string{"module_taxpayers", "module_auth", "module_auth"}})

	got, err := svc.EnabledModules(t.Context())

	require.NoError(t, err)
	assert.Equal(t, []string{"module_auth", "module_taxpayers"}, got)
}

func TestEnabledModulesReturnsACopy(t *testing.T) {
	svc, _ := newService(t, &fakeReader{keys: []string{"module_auth"}})
	first, err := svc.EnabledModules(t.Context())
	require.NoError(t, err)
	first[0] = "module_tampered"

	second, err := svc.EnabledModules(t.Context())

	require.NoError(t, err)
	assert.Equal(t, []string{"module_auth"}, second)
}

func TestEnabledModulesWithNothingEnabledIsEmptyNotNil(t *testing.T) {
	svc, _ := newService(t, &fakeReader{})

	got, err := svc.EnabledModules(t.Context())

	require.NoError(t, err)
	assert.NotNil(t, got, "the contract requires an array, never null")
	assert.Empty(t, got)
}

func TestCacheServesWithinTTL(t *testing.T) {
	reader := &fakeReader{keys: []string{"module_auth"}}
	svc, clk := newService(t, reader)

	require.True(t, svc.Enabled(t.Context(), "module_auth"))
	reader.set(nil, nil) // turned OFF in the database
	clk.advance(testTTL - time.Nanosecond)

	assert.True(t, svc.Enabled(t.Context(), "module_auth"), "still cached")
	assert.Equal(t, int64(1), reader.loads.Load())
}

func TestCacheReloadsWhenTTLExpires(t *testing.T) {
	reader := &fakeReader{keys: []string{"module_auth"}}
	svc, clk := newService(t, reader)

	require.True(t, svc.Enabled(t.Context(), "module_auth"))
	reader.set(nil, nil)
	clk.advance(testTTL)

	assert.False(t, svc.Enabled(t.Context(), "module_auth"))
	assert.Equal(t, int64(2), reader.loads.Load())
}

func TestInvalidateForcesAReload(t *testing.T) {
	reader := &fakeReader{keys: []string{"module_auth"}}
	svc, _ := newService(t, reader)

	require.True(t, svc.Enabled(t.Context(), "module_auth"))
	reader.set([]string{"module_auth", "module_incomes"}, nil)
	svc.Invalidate()

	assert.True(t, svc.Enabled(t.Context(), "module_incomes"))
	assert.Equal(t, int64(2), reader.loads.Load())
}

func TestFailsClosedWhenTheReaderFails(t *testing.T) {
	reader := &fakeReader{keys: []string{"module_auth"}}
	svc, clk := newService(t, reader)
	require.True(t, svc.Enabled(t.Context(), "module_auth"))

	reader.set(nil, errDatabaseDown)
	clk.advance(testTTL)

	assert.False(t, svc.Enabled(t.Context(), "module_auth"), "a stale ON is never served after a failed reload")
	_, err := svc.EnabledModules(t.Context())
	assert.ErrorIs(t, err, errDatabaseDown)
}

func TestFailureIsNotCached(t *testing.T) {
	reader := &fakeReader{err: errDatabaseDown}
	svc, _ := newService(t, reader)
	require.False(t, svc.Enabled(t.Context(), "module_auth"))

	reader.set([]string{"module_auth"}, nil) // the database is back

	assert.True(t, svc.Enabled(t.Context(), "module_auth"))
}

// blockingReader honors cancellation the way a database driver does.
type blockingReader struct{}

func (blockingReader) EnabledKeys(ctx context.Context) ([]string, error) {
	<-ctx.Done()
	return nil, ctx.Err()
}

func TestFailsClosedWhenTheContextIsCanceled(t *testing.T) {
	svc, _ := newService(t, blockingReader{})
	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	assert.False(t, svc.Enabled(ctx, "module_auth"))
}

// gatedReader blocks every load until release is closed and signals the
// first load on started.
type gatedReader struct {
	release <-chan struct{}
	started chan struct{}
	once    sync.Once
	keys    []string
	loads   atomic.Int64
}

func (r *gatedReader) EnabledKeys(ctx context.Context) ([]string, error) {
	r.loads.Add(1)
	r.once.Do(func() { close(r.started) })
	select {
	case <-r.release:
		return r.keys, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func TestConcurrentMissesShareOneLoad(t *testing.T) {
	release := make(chan struct{})
	reader := &gatedReader{release: release, started: make(chan struct{}), keys: []string{"module_auth"}}
	svc, _ := newService(t, reader)

	const callers = 16
	var wg sync.WaitGroup
	results := make([]bool, callers)
	for i := range callers {
		wg.Go(func() { results[i] = svc.Enabled(t.Context(), "module_auth") })
	}
	<-reader.started
	close(release)
	wg.Wait()

	assert.Equal(t, int64(1), reader.loads.Load(), "waiters reuse the snapshot the first caller loaded")
	for i, got := range results {
		assert.True(t, got, "caller %d", i)
	}
}

func TestWaiterHonorsItsOwnCancellation(t *testing.T) {
	release := make(chan struct{})
	defer close(release)
	reader := &gatedReader{release: release, started: make(chan struct{}), keys: []string{"module_auth"}}
	svc, _ := newService(t, reader)
	go svc.Enabled(t.Context(), "module_auth") // holds the load
	<-reader.started

	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	assert.False(t, svc.Enabled(ctx, "module_auth"), "a canceled waiter gives up instead of queueing")
}

// hangGuard bounds how long a test waits for a call that must not hang.
const hangGuard = 5 * time.Second

// Regression (P0-28): a database that hangs instead of refusing (paused,
// blackholed, overloaded) made every gated request wait for the client to
// give up. The load is bounded, so the gate still answers 404 promptly.
func TestLoadIsBoundedByTheLoadTimeout(t *testing.T) {
	svc := flags.NewService(blockingReader{}, flags.WithLoadTimeout(20*time.Millisecond))

	done := make(chan bool, 1)
	go func() { done <- svc.Enabled(t.Context(), "module_auth") }()

	select {
	case got := <-done:
		assert.False(t, got, "a load that times out fails closed")
	case <-time.After(hangGuard):
		t.Fatal("Enabled hangs while the reader hangs: the load is not bounded")
	}
}

// failingGatedReader blocks every load until release is closed, then fails.
type failingGatedReader struct {
	release <-chan struct{}
	started chan struct{}
	once    sync.Once
	loads   atomic.Int64
}

func (r *failingGatedReader) EnabledKeys(ctx context.Context) ([]string, error) {
	r.loads.Add(1)
	r.once.Do(func() { close(r.started) })
	select {
	case <-r.release:
		return nil, errDatabaseDown
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// Regression (P0-28): callers queued behind a failed load each retried it in
// turn, so with a slow database the N-th request waited N load timeouts.
func TestWaitersShareAFailedLoad(t *testing.T) {
	release := make(chan struct{})
	reader := &failingGatedReader{release: release, started: make(chan struct{})}
	svc, _ := newService(t, reader)

	const callers = 16
	var wg sync.WaitGroup
	results := make([]bool, callers)
	for i := range callers {
		wg.Go(func() { results[i] = svc.Enabled(t.Context(), "module_auth") })
	}
	<-reader.started
	// Callers should be queued on the load before it fails. Enabled cannot
	// signal that, so the scheduler gets a moment, and the assertion tolerates
	// a straggler that arrives after the failure (it may legitimately reload).
	// Before the fix every caller retried: 16 loads.
	time.Sleep(50 * time.Millisecond)
	close(release)
	wg.Wait()

	assert.Less(t, reader.loads.Load(), int64(callers/2), "waiters reuse the failure of the load they waited on")
	for i, got := range results {
		assert.False(t, got, "caller %d", i)
	}
}

// The load must not depend on the caller that happened to start it: if that
// client disconnects, the callers waiting on the same load still get flags.
func TestLoadSurvivesTheCancellationOfTheCallerThatStartedIt(t *testing.T) {
	release := make(chan struct{})
	reader := &gatedReader{release: release, started: make(chan struct{}), keys: []string{"module_auth"}}
	svc, _ := newService(t, reader)

	first, cancelFirst := context.WithCancel(t.Context())
	firstDone := make(chan struct{})
	go func() { defer close(firstDone); svc.Enabled(first, "module_auth") }()
	<-reader.started

	second := make(chan bool, 1)
	go func() { second <- svc.Enabled(t.Context(), "module_auth") }()
	cancelFirst()
	close(release)
	<-firstDone

	select {
	case got := <-second:
		assert.True(t, got, "the load completes for the callers still waiting")
	case <-time.After(hangGuard):
		t.Fatal("second caller hangs")
	}
}

func TestDefaultLoadTimeoutIsShort(t *testing.T) {
	assert.Positive(t, flags.DefaultLoadTimeout)
	assert.LessOrEqual(t, flags.DefaultLoadTimeout, 2*time.Second,
		"a hung database must still let the gate answer 404 well inside the client's patience")
}

func TestNewServiceRejectsANonPositiveLoadTimeout(t *testing.T) {
	assert.Panics(t, func() { flags.NewService(&fakeReader{}, flags.WithLoadTimeout(0)) })
}

// Regression: a nil logger made the fail-closed path panic (500, not 404).
func TestNilLoggerStillFailsClosed(t *testing.T) {
	svc := flags.NewService(&fakeReader{err: errDatabaseDown}, flags.WithLogger(nil))

	assert.NotPanics(t, func() { assert.False(t, svc.Enabled(t.Context(), "module_auth")) })
}

func TestDefaultTTLIsShort(t *testing.T) {
	assert.Positive(t, flags.DefaultTTL)
	assert.LessOrEqual(t, flags.DefaultTTL, 10*time.Second,
		"a flag toggled by cmd/flags must reach every API instance within seconds")
}

func TestNewServiceRejectsANonPositiveTTL(t *testing.T) {
	assert.Panics(t, func() { flags.NewService(&fakeReader{}, flags.WithTTL(0)) })
}
