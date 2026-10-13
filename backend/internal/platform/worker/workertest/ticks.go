// Package workertest is test support for the worker runner: manual tickers the
// test fires by hand, so scheduling tests never wait on the wall clock.
package workertest

import (
	"sync"
	"testing"
	"time"

	"github.com/romodevx/control-contable/backend/internal/platform/worker"
)

// Ticks is a worker.NewTicker whose tickers fire only when the test calls Tick.
// Tickers are keyed by interval, so give the heartbeat and every job a distinct
// interval.
type Ticks struct {
	mu       sync.Mutex
	channels map[time.Duration]*manualTicker
}

// NewTicks returns an empty set of manual tickers.
func NewTicks() *Ticks {
	return &Ticks{channels: make(map[time.Duration]*manualTicker)}
}

// NewTicker implements worker.NewTicker.
func (t *Ticks) NewTicker(interval time.Duration) worker.Ticker {
	return t.get(interval)
}

// Tick fires the ticker of interval once. It blocks until the runner receives
// the tick, which also means the runner finished handling the previous tick of
// that ticker: two Ticks in a row are a barrier for the first one.
func (t *Ticks) Tick(tb testing.TB, interval time.Duration) {
	tb.Helper()
	select {
	case t.get(interval).c <- time.Time{}:
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
