package ratelimit

import (
	"errors"
	"fmt"
	"regexp"
	"time"
)

// ErrInvalidPolicy reports a Policy that no bucket can enforce.
var ErrInvalidPolicy = errors.New("ratelimit: invalid policy")

// scopePattern is the rate_limit_buckets.scope CHECK (snake_case, ≤ 63 bytes).
var scopePattern = regexp.MustCompile(`^[a-z][a-z0-9_]{0,62}$`)

// Policy is one shared, persistent limit (ADR-0005): a bucket per key in
// Scope holding up to Burst events, refilled one event per Interval (GCRA).
// "5 per 15 minutes" is PerWindow(scope, 5, 15*time.Minute).
type Policy struct {
	// Scope names the limit; keys of different scopes never share a bucket.
	Scope    string
	Burst    int
	Interval time.Duration
}

// PerWindow is the policy admitting limit events at once and refilling the
// whole budget evenly over window.
func PerWindow(scope string, limit int, window time.Duration) Policy {
	return Policy{Scope: scope, Burst: limit, Interval: window / time.Duration(max(limit, 1))}
}

// Validate reports ErrInvalidPolicy for a scope that is not snake_case, a
// non-positive burst, or an interval under the database's microsecond
// resolution.
func (p Policy) Validate() error {
	switch {
	case !scopePattern.MatchString(p.Scope):
		return fmt.Errorf("%w: scope %q is not snake_case", ErrInvalidPolicy, p.Scope)
	case p.Burst < 1:
		return fmt.Errorf("%w: burst %d", ErrInvalidPolicy, p.Burst)
	case p.Interval < time.Microsecond:
		return fmt.Errorf("%w: interval %s", ErrInvalidPolicy, p.Interval)
	}
	return nil
}

// RetryAfterSeconds is wait as an HTTP Retry-After value: whole seconds,
// rounded up, at least 1 (the contract's TooManyRequests).
func RetryAfterSeconds(wait time.Duration) int {
	return max(int((wait+time.Second-1)/time.Second), 1)
}
