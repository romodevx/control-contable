package ratelimit_test

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/platform/ratelimit"
)

var epoch = time.Date(2026, 10, 13, 12, 0, 0, 0, time.UTC)

// fakeClock is a manually advanced clock; safe for concurrent reads.
type fakeClock struct {
	mu  sync.Mutex
	now time.Time
}

func newFakeClock() *fakeClock { return &fakeClock{now: epoch} }

func (c *fakeClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.now
}

func (c *fakeClock) Advance(d time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.now = c.now.Add(d)
}

func TestLimiter_burstThenRefill(t *testing.T) {
	tests := []struct {
		name      string
		perSecond int
		burst     int
		advance   time.Duration // after the burst is spent
		wantAllow bool
		wantRetry time.Duration // when denied
	}{
		{name: "spent burst is denied", perSecond: 2, burst: 3, wantRetry: 500 * time.Millisecond},
		{name: "partial refill is still denied", perSecond: 2, burst: 3, advance: 200 * time.Millisecond, wantRetry: 300 * time.Millisecond},
		{name: "one emission interval refills one token", perSecond: 2, burst: 3, advance: 500 * time.Millisecond, wantAllow: true},
		{name: "burst of one", perSecond: 1, burst: 1, wantRetry: time.Second},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clock := newFakeClock()
			l := ratelimit.New(tt.perSecond, tt.burst, clock.Now)
			for i := range tt.burst {
				allowed, _ := l.Allow("ip")
				require.True(t, allowed, "request %d of the burst", i+1)
			}
			clock.Advance(tt.advance)

			allowed, retry := l.Allow("ip")

			assert.Equal(t, tt.wantAllow, allowed)
			if !tt.wantAllow {
				assert.Equal(t, tt.wantRetry, retry)
			}
		})
	}
}

func TestLimiter_deniedRequestsDoNotConsume(t *testing.T) {
	clock := newFakeClock()
	l := ratelimit.New(1, 1, clock.Now)
	allowed, _ := l.Allow("ip")
	require.True(t, allowed)
	for range 5 {
		allowed, _ = l.Allow("ip")
		require.False(t, allowed)
	}

	clock.Advance(time.Second)
	allowed, _ = l.Allow("ip")

	assert.True(t, allowed, "a hammering client recovers after one interval, not after its denied requests")
}

func TestLimiter_keysAreIndependent(t *testing.T) {
	l := ratelimit.New(1, 1, newFakeClock().Now)
	a, _ := l.Allow("a")
	a2, _ := l.Allow("a")
	b, _ := l.Allow("b")

	assert.True(t, a)
	assert.False(t, a2)
	assert.True(t, b)
}

func TestLimiter_idleKeysAreEvicted(t *testing.T) {
	clock := newFakeClock()
	l := ratelimit.New(10, 10, clock.Now)
	for i := range 100 {
		l.Allow(fmt.Sprintf("10.0.0.%d", i))
	}
	require.Equal(t, 100, l.Len())

	clock.Advance(ratelimit.SweepInterval)
	l.Allow("fresh")

	assert.Equal(t, 1, l.Len(), "keys whose bucket refilled are dropped; only the new one stays")
}

func TestLimiter_busyKeysSurviveTheSweep(t *testing.T) {
	clock := newFakeClock()
	l := ratelimit.New(1, 1000, clock.Now) // 1000 requests need 1000 s to refill
	for range 1000 {
		l.Allow("busy")
	}

	clock.Advance(ratelimit.SweepInterval)
	l.Allow("other")

	assert.Equal(t, 2, l.Len())
}

func TestLimiter_concurrentUseNeverExceedsBurst(t *testing.T) {
	const burst = 50
	l := ratelimit.New(1, burst, newFakeClock().Now)
	var (
		wg      sync.WaitGroup
		mu      sync.Mutex
		allowed int
	)
	for range 4 * burst {
		wg.Go(func() {
			if ok, _ := l.Allow("ip"); ok {
				mu.Lock()
				allowed++
				mu.Unlock()
			}
		})
	}
	wg.Wait()

	assert.Equal(t, burst, allowed)
}

func TestNew_rejectsNonPositiveSettings(t *testing.T) {
	now := newFakeClock().Now
	assert.Panics(t, func() { ratelimit.New(0, 1, now) })
	assert.Panics(t, func() { ratelimit.New(1, 0, now) })
}

// NewEvery expresses slow rates (one per hour) that a per-second rate cannot.
func TestNewEvery_slowRate(t *testing.T) {
	clk := newFakeClock()
	l := ratelimit.NewEvery(time.Hour, 1, clk.Now)

	ok, _ := l.Allow("k")
	require.True(t, ok, "the first one passes")
	ok, wait := l.Allow("k")
	require.False(t, ok, "a second one within the hour is denied")
	assert.Equal(t, time.Hour, wait)
	ok, _ = l.Allow("other")
	assert.True(t, ok, "keys are independent")

	clk.Advance(time.Hour)
	ok, _ = l.Allow("k")
	assert.True(t, ok, "one more after the interval")
}

func TestNewEvery_burstThenOnePerInterval(t *testing.T) {
	clk := newFakeClock()
	l := ratelimit.NewEvery(12*time.Minute, 5, clk.Now)
	for i := range 5 {
		ok, _ := l.Allow("k")
		require.True(t, ok, "burst %d", i)
	}
	ok, wait := l.Allow("k")
	require.False(t, ok)
	assert.Equal(t, 12*time.Minute, wait)
}

func TestNewEvery_rejectsNonPositiveSettings(t *testing.T) {
	now := newFakeClock().Now
	assert.Panics(t, func() { ratelimit.NewEvery(0, 1, now) })
	assert.Panics(t, func() { ratelimit.NewEvery(time.Second, 0, now) })
}
