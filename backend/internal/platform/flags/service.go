// Package flags is the module feature-flag service (P0-15, ADR-0001,
// feature-flags rule). The API asks it whether a module is ON before routing a
// request to that module, and GET /v1/flags lists the enabled modules for the
// client (UX only). Flags live in the feature_flags table; this package sees
// them only through the Reader port, so the storage is a replaceable detail.
//
// The service caches the enabled set for a short TTL and fails closed: when
// the set cannot be read, every module is treated as OFF.
package flags

import (
	"context"
	"fmt"
	"log/slog"
	"slices"
	"sync/atomic"
	"time"

	"github.com/romodevx/control-contable/backend/internal/platform/clock"
)

// DefaultTTL bounds how long a toggle made by cmd/flags (another process) takes
// to reach every API instance. A process-local Invalidate cannot cross
// instances, so the TTL is the propagation guarantee.
const DefaultTTL = 5 * time.Second

// DefaultLoadTimeout bounds one read of the enabled set. A database that hangs
// instead of refusing must not hold requests: the gate fails closed (404) once
// the load times out.
const DefaultLoadTimeout = 2 * time.Second

// Reader is the port to the flag storage: it returns the keys of the flags
// that are ON. The Postgres adapter reads them as the runtime role.
type Reader interface {
	EnabledKeys(ctx context.Context) ([]string, error)
}

// Option configures a Service.
type Option func(*Service)

// WithTTL sets how long a loaded snapshot is served. It must be positive.
func WithTTL(ttl time.Duration) Option { return func(s *Service) { s.ttl = ttl } }

// WithLoadTimeout bounds each read of the enabled set. It must be positive.
func WithLoadTimeout(d time.Duration) Option { return func(s *Service) { s.loadTimeout = d } }

// WithClock replaces the wall clock (tests).
func WithClock(c clock.Clock) Option { return func(s *Service) { s.clock = c } }

// WithLogger sets where failed reloads are reported. A nil logger keeps the
// default (discard): the fail-closed path must never panic.
func WithLogger(l *slog.Logger) Option {
	return func(s *Service) {
		if l != nil {
			s.logger = l
		}
	}
}

// Service answers "is this module ON?" from a cached snapshot of the enabled
// flags. It is safe for concurrent use.
type Service struct {
	reader Reader
	clock  clock.Clock
	logger *slog.Logger
	ttl    time.Duration
	// loadTimeout bounds one reader call.
	loadTimeout time.Duration

	snap atomic.Pointer[snapshot]
	// gen is bumped by Invalidate; a snapshot of an older generation is stale.
	gen atomic.Uint64
	// loading is a one-slot semaphore: one caller reloads while concurrent
	// misses wait for its snapshot (or give up when their context ends).
	loading chan struct{}
	// loads counts finished reader calls; failed holds the latest one when it
	// failed, so callers that waited on it share its error instead of each
	// retrying behind the semaphore.
	loads  atomic.Uint64
	failed atomic.Pointer[loadFailure]
}

// loadFailure is the error of the reader call numbered seq.
type loadFailure struct {
	seq uint64
	err error
}

// snapshot is an immutable view of the effectively enabled modules: the ON
// flags whose prerequisites are ON too.
type snapshot struct {
	enabled map[string]struct{}
	sorted  []string
	expires time.Time
	gen     uint64
}

// NewService returns a Service reading through reader. It panics on a
// non-positive TTL or load timeout: that is a wiring bug, caught at startup.
func NewService(reader Reader, opts ...Option) *Service {
	s := &Service{
		reader:      reader,
		clock:       clock.NewSystem(),
		logger:      slog.New(slog.DiscardHandler),
		ttl:         DefaultTTL,
		loadTimeout: DefaultLoadTimeout,
		loading:     make(chan struct{}, 1),
	}
	for _, opt := range opts {
		opt(s)
	}
	if s.ttl <= 0 {
		panic(fmt.Sprintf("flags: TTL must be positive, got %s", s.ttl))
	}
	if s.loadTimeout <= 0 {
		panic(fmt.Sprintf("flags: load timeout must be positive, got %s", s.loadTimeout))
	}
	return s
}

// Enabled reports whether module is effectively ON: its flag and every
// prerequisite's (Prerequisites) are ON. It fails closed: when the
// flags cannot be read (database error, canceled context) it returns false.
func (s *Service) Enabled(ctx context.Context, module string) bool {
	snap, err := s.current(ctx)
	if err != nil {
		s.logger.WarnContext(ctx, "feature flags unavailable; treating module as OFF",
			slog.String("module", module), slog.Any("error", err))
		return false
	}
	_, on := snap.enabled[module]
	return on
}

// EnabledModules returns the keys of the effectively enabled modules (see
// Enabled), sorted. It never
// returns nil on success, so the JSON is an array even when nothing is ON.
func (s *Service) EnabledModules(ctx context.Context) ([]string, error) {
	snap, err := s.current(ctx)
	if err != nil {
		return nil, err
	}
	return slices.Clone(snap.sorted), nil
}

// Invalidate drops the cached snapshot; the next call reloads. It only
// affects this process (see DefaultTTL).
func (s *Service) Invalidate() { s.gen.Add(1) }

func (s *Service) current(ctx context.Context) (*snapshot, error) {
	if snap := s.snap.Load(); s.fresh(snap) {
		return snap, nil
	}
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("flags: %w", err)
	}
	seen := s.loads.Load()
	select {
	case s.loading <- struct{}{}:
	case <-ctx.Done():
		return nil, fmt.Errorf("flags: wait for reload: %w", ctx.Err())
	}
	defer func() { <-s.loading }()
	// Another caller may have reloaded while this one waited.
	if snap := s.snap.Load(); s.fresh(snap) {
		return snap, nil
	}
	// Or failed to: share that error rather than queue one more slow load.
	if f := s.failed.Load(); f != nil && f.seq > seen {
		return nil, f.err
	}
	return s.reload(ctx)
}

// reload reads the enabled set once, bounded by loadTimeout. The read does
// not inherit the caller's cancellation: other callers may be waiting on it.
func (s *Service) reload(ctx context.Context) (*snapshot, error) {
	gen := s.gen.Load()
	loadCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), s.loadTimeout)
	defer cancel()
	keys, err := s.reader.EnabledKeys(loadCtx)
	seq := s.loads.Add(1)
	if err != nil {
		err = fmt.Errorf("flags: read enabled flags: %w", err)
		s.failed.Store(&loadFailure{seq: seq, err: err})
		return nil, err
	}
	snap := newSnapshot(EffectiveModules(prerequisites, keys), s.clock.Now().Add(s.ttl), gen)
	// An Invalidate during the read leaves gen behind, so fresh rejects this
	// snapshot on the next call; storing it is harmless.
	s.snap.Store(snap)
	return snap, nil
}

func (s *Service) fresh(snap *snapshot) bool {
	return snap != nil && snap.gen == s.gen.Load() && s.clock.Now().Before(snap.expires)
}

func newSnapshot(keys []string, expires time.Time, gen uint64) *snapshot {
	enabled := make(map[string]struct{}, len(keys))
	for _, k := range keys {
		enabled[k] = struct{}{}
	}
	sorted := make([]string, 0, len(enabled))
	for k := range enabled {
		sorted = append(sorted, k)
	}
	slices.Sort(sorted)
	return &snapshot{enabled: enabled, sorted: sorted, expires: expires, gen: gen}
}
