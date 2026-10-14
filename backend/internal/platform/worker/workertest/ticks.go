// Package workertest is test support for the worker runner: manual tickers the
// test fires by hand, so scheduling tests never wait on the wall clock.
package workertest

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/romodevx/control-contable/backend/internal/platform/worker"
)

// Ticks is a worker.NewTicker whose tickers fire only when the test calls Tick.
// Tickers are keyed by interval, so give the heartbeat and every job a distinct
// interval.
type Ticks struct {
	waitLimit time.Duration

	mu       sync.Mutex
	channels map[time.Duration]*manualTicker
}

// Option customizes NewTicks.
type Option func(*Ticks)

// WithWaitLimit replaces WaitLimit as the longest Tick waits for a runner.
func WithWaitLimit(d time.Duration) Option { return func(t *Ticks) { t.waitLimit = d } }

// NewTicks returns an empty set of manual tickers.
func NewTicks(opts ...Option) *Ticks {
	t := &Ticks{waitLimit: WaitLimit, channels: make(map[time.Duration]*manualTicker)}
	for _, opt := range opts {
		opt(t)
	}
	return t
}

// NewTicker implements worker.NewTicker.
func (t *Ticks) NewTicker(interval time.Duration) worker.Ticker {
	return t.get(interval)
}

// Tick fires the ticker of interval once. It blocks until the runner receives
// the tick, which also means the runner finished handling the previous tick of
// that ticker: two Ticks in a row are a barrier for the first one. A tick no
// runner receives within the wait limit fails tb with every goroutine's stack.
func (t *Ticks) Tick(tb testing.TB, interval time.Duration) {
	tb.Helper()
	timer := time.NewTimer(t.waitLimit)
	defer timer.Stop()
	select {
	case t.get(interval).c <- time.Time{}:
	case <-timer.C:
		failHung(tb, t.waitLimit, fmt.Sprintf("a runner to receive the %s tick", interval))
	case <-tb.Context().Done():
		tb.Fatalf("workertest: no runner received the %s tick", interval)
	}
}

// Stopped reports whether the ticker of interval was stopped.
func (t *Ticks) Stopped(interval time.Duration) bool {
	m := t.get(interval)
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.stopped
}

func (t *Ticks) get(interval time.Duration) *manualTicker {
	t.mu.Lock()
	defer t.mu.Unlock()
	m, ok := t.channels[interval]
	if !ok {
		m = &manualTicker{c: make(chan time.Time)}
		t.channels[interval] = m
	}
	return m
}

type manualTicker struct {
	c       chan time.Time
	mu      sync.Mutex
	stopped bool
}

func (m *manualTicker) C() <-chan time.Time { return m.c }

func (m *manualTicker) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.stopped = true
}
