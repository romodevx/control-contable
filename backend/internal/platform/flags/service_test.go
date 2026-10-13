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

func TestDefaultTTLIsShort(t *testing.T) {
	assert.Positive(t, flags.DefaultTTL)
	assert.LessOrEqual(t, flags.DefaultTTL, 10*time.Second,
		"a flag toggled by cmd/flags must reach every API instance within seconds")
}

func TestNewServiceRejectsANonPositiveTTL(t *testing.T) {
	assert.Panics(t, func() { flags.NewService(&fakeReader{}, flags.WithTTL(0)) })
}
