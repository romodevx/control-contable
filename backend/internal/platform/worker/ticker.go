package worker

import "time"

// Ticker delivers ticks on C until stopped; *time.Ticker behind SystemTicker.
type Ticker interface {
	C() <-chan time.Time
	Stop()
}

// NewTicker makes the Ticker of one loop (the heartbeat or a job).
type NewTicker func(interval time.Duration) Ticker

// SystemTicker is the production NewTicker, on the wall clock.
func SystemTicker(interval time.Duration) Ticker {
	return systemTicker{time.NewTicker(interval)}
}

type systemTicker struct{ t *time.Ticker }

func (s systemTicker) C() <-chan time.Time { return s.t.C }
func (s systemTicker) Stop()               { s.t.Stop() }
