// Package ratelimit is an in-process, keyed token-bucket rate limiter
// (ADR-0001 O9: counters live in process until we scale past one instance).
//
// It uses GCRA, the generic cell rate algorithm: an exact token bucket that
// stores one timestamp per key (the theoretical arrival time, TAT) instead of
// a token count and a refill time. Allow is O(1); idle keys are swept every
// SweepInterval in O(n), so memory follows the number of recently active keys.
package ratelimit

import (
	"fmt"
	"sync"
	"time"
)

// SweepInterval is how often Allow drops the keys whose bucket is full again.
const SweepInterval = time.Minute

// Limiter admits, per key, PerSecond requests per second with bursts of up
// to Burst. It is safe for concurrent use.
type Limiter struct {
	interval  time.Duration // emission interval: time to refill one token
	tolerance time.Duration // interval × burst: how far TAT may run ahead of now
	now       func() time.Time

	mu        sync.Mutex
	tat       map[string]time.Time
	lastSweep time.Time
}

// New returns a Limiter refilling perSecond tokens per second up to burst.
// now is the clock (time.Now in production). It panics on non-positive
// settings: they are programming errors, config validates them first.
func New(perSecond, burst int, now func() time.Time) *Limiter {
	if perSecond < 1 || burst < 1 {
		panic(fmt.Sprintf("ratelimit: perSecond (%d) and burst (%d) must be positive", perSecond, burst))
	}
	return NewEvery(time.Second/time.Duration(perSecond), burst, now)
}

// NewEvery returns a Limiter refilling one token every interval up to burst:
// the form for slow rates such as "once per hour" (interval time.Hour,
// burst 1). It panics on non-positive settings, like New.
func NewEvery(interval time.Duration, burst int, now func() time.Time) *Limiter {
	if interval <= 0 || burst < 1 {
		panic(fmt.Sprintf("ratelimit: interval (%s) and burst (%d) must be positive", interval, burst))
	}
	return &Limiter{
		interval:  interval,
		tolerance: interval * time.Duration(burst),
		now:       now,
		tat:       make(map[string]time.Time),
		lastSweep: now(),
	}
}

// Allow takes one token from key's bucket. When the bucket is empty it
// returns false and how long until a token is available; a denied request
// consumes nothing.
func (l *Limiter) Allow(key string) (allowed bool, retryAfter time.Duration) {
	now := l.now()
	l.mu.Lock()
	defer l.mu.Unlock()

	l.sweep(now)
	tat := l.tat[key]
	if tat.Before(now) {
		tat = now
	}
	next := tat.Add(l.interval)
	if ahead := next.Sub(now); ahead > l.tolerance {
		return false, ahead - l.tolerance
	}
	l.tat[key] = next
	return true, 0
}

// Len returns the number of keys currently tracked.
func (l *Limiter) Len() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return len(l.tat)
}

// sweep drops keys whose TAT has passed: their bucket is full, which is the
// same state as an unknown key. Callers hold l.mu.
func (l *Limiter) sweep(now time.Time) {
	if now.Sub(l.lastSweep) < SweepInterval {
		return
	}
	l.lastSweep = now
	for key, tat := range l.tat {
		if !tat.After(now) {
			delete(l.tat, key)
		}
	}
}
