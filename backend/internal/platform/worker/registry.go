// Package worker is the background job runner of cmd/worker (ADR-0001 O8): an
// in-process ticker per job, run only by the instance that holds the leader
// lock, so a horizontally scaled worker does each run once.
//
// Leadership comes from an Elector (pgdb.AdvisoryLeader: pg_try_advisory_lock
// on a dedicated connection). The Runner campaigns at start and on every
// heartbeat; an Elector error is a lost leadership, never an assumed one.
// Every run gets a context that is cancelled when its timeout passes, when
// leadership is lost or when the worker shuts down, so a job never outlives
// the leadership it started under. A panicking job is recovered and logged;
// the worker keeps running.
//
// Exactly-once effects do not come from the scheduler: between a leader's
// connection dying and that leader noticing at its next heartbeat, two
// instances may briefly both believe they lead. Jobs stay idempotent (e.g.
// the UNIQUE key of notification_log).
package worker

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"time"
)

// Job is one kind of background work.
type Job interface {
	// Name identifies the job in logs: snake_case, unique in a Registry.
	Name() string
	// Interval is the time between two runs.
	Interval() time.Duration
	// Run does one run. It must honor ctx: the Runner cancels it on timeout,
	// on lost leadership and on shutdown, and waits for Run to return.
	Run(ctx context.Context) error
}

// ErrInvalidJob reports a job the Registry refuses.
var ErrInvalidJob = errors.New("worker: invalid job")

var jobName = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)

// entry is a registered job with its per-run timeout.
type entry struct {
	job     Job
	timeout time.Duration
}

// Registry is the set of jobs a Runner schedules. The zero value is empty and
// ready to use. It is not safe for concurrent use: fill it in cmd/ before the
// Runner starts.
type Registry struct {
	entries []entry
}

// Register adds job with the timeout every run of it gets. The timeout is
// positive and at most the job's interval, since runs of one job never overlap.
func (r *Registry) Register(job Job, timeout time.Duration) error {
	if err := r.validate(job, timeout); err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidJob, err)
	}
	r.entries = append(r.entries, entry{job: job, timeout: timeout})
	return nil
}

func (r *Registry) validate(job Job, timeout time.Duration) error {
	if job == nil {
		return errors.New("nil job")
	}
	name := job.Name()
	if !jobName.MatchString(name) {
		return fmt.Errorf("name %q is not snake_case", name)
	}
	for _, e := range r.entries {
		if e.job.Name() == name {
			return fmt.Errorf("%q is already registered", name)
		}
	}
	interval := job.Interval()
	if interval <= 0 {
		return fmt.Errorf("%s: interval %s is not positive", name, interval)
	}
	if timeout <= 0 || timeout > interval {
		return fmt.Errorf("%s: timeout %s is not in (0, %s]", name, timeout, interval)
	}
	return nil
}

// Names lists the registered jobs in registration order.
func (r *Registry) Names() []string {
	names := make([]string, 0, len(r.entries))
	for _, e := range r.entries {
		names = append(names, e.job.Name())
	}
	return names
}
