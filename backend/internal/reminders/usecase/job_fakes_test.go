package usecase_test

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/platform/tx"
	"github.com/romodevx/control-contable/backend/internal/reminders/domain"
	"github.com/romodevx/control-contable/backend/internal/reminders/usecase"
)

// jobLogRow is one in-memory notification_log row.
type jobLogRow struct {
	id         string
	status     usecase.LogStatus
	attempts   int
	claimedAt  time.Time
	outcome    usecase.LogOutcome
	finishedAt int // how many Finish calls recorded it (exactly once: 1)
}

// jobDB is the in-memory database the job's per-user ports share: the log
// with the UNIQUE (user, period, kind) semantics of the migration, the
// taxpayers, the devices and the settings. Every per-user method reads the
// user of the fake unit of work (RLS). It is safe for concurrent use, so
// two jobs (two worker instances) can share it.
type jobDB struct {
	mu        sync.Mutex
	seq       int
	log       map[string]*jobLogRow // key: user|period|kind
	settings  map[string]domain.Settings
	taxpayers map[string][]usecase.RecipientTaxpayer
	devices   map[string][]usecase.SealedTarget
	deleted   []string

	claimErr  error
	finishErr error
}

func newJobDB() *jobDB {
	return &jobDB{
		log: map[string]*jobLogRow{}, settings: map[string]domain.Settings{},
		taxpayers: map[string][]usecase.RecipientTaxpayer{}, devices: map[string][]usecase.SealedTarget{},
	}
}

func logKey(user string, p domain.Period, k domain.Kind) string {
	return user + "|" + p.String() + "|" + string(k)
}

func requireUser(ctx context.Context) (string, error) {
	user := userOf(ctx)
	if user == "" {
		return "", errors.New("fake: per-user port called outside a user unit of work")
	}
	return user, nil
}

func (db *jobDB) ExpireStale(ctx context.Context, claimedBefore time.Time) (int, error) {
	user, err := requireUser(ctx)
	if err != nil {
		return 0, err
	}
	db.mu.Lock()
	defer db.mu.Unlock()
	n := 0
	for key, row := range db.log {
		if strings.HasPrefix(key, user+"|") && row.status == usecase.LogPending && row.claimedAt.Before(claimedBefore) {
			row.status, row.outcome = usecase.LogFailed, usecase.LogOutcome{Status: usecase.LogFailed, ErrorClass: "expired"}
			n++
		}
	}
	return n, nil
}

func (db *jobDB) Claim(ctx context.Context, r usecase.DueReminder, now, leaseCutoff time.Time) (usecase.LogClaim, bool, error) {
	user, err := requireUser(ctx)
	if err != nil {
		return usecase.LogClaim{}, false, err
	}
	if user != r.UserID {
		return usecase.LogClaim{}, false, errors.New("fake: RLS: claim for another user")
	}
	db.mu.Lock()
	defer db.mu.Unlock()
	if db.claimErr != nil {
		return usecase.LogClaim{}, false, db.claimErr
	}
	key := logKey(user, r.Period, r.Kind)
	row, ok := db.log[key]
	if !ok {
		db.seq++
		row = &jobLogRow{id: fmt.Sprintf("log-%d", db.seq), status: usecase.LogPending, attempts: 1, claimedAt: now}
		db.log[key] = row
		return usecase.LogClaim{ID: row.id, Attempts: 1}, true, nil
	}
	if row.status != usecase.LogPending || !row.claimedAt.Before(leaseCutoff) || row.attempts >= 3 {
		return usecase.LogClaim{}, false, nil
	}
	row.attempts++
	row.claimedAt = now
	return usecase.LogClaim{ID: row.id, Attempts: row.attempts}, true, nil
}

func (db *jobDB) Finish(ctx context.Context, c usecase.LogClaim, o usecase.LogOutcome) (bool, error) {
	if _, err := requireUser(ctx); err != nil {
		return false, err
	}
	db.mu.Lock()
	defer db.mu.Unlock()
	if db.finishErr != nil {
		return false, db.finishErr
	}
	for _, row := range db.log {
		if row.id == c.ID && row.attempts == c.Attempts && row.status == usecase.LogPending {
			row.status, row.outcome = o.Status, o
			row.finishedAt++
			return true, nil
		}
	}
	return false, nil
}

func (db *jobDB) Settings(ctx context.Context, userID string) (domain.Settings, bool, error) {
	if _, err := requireUser(ctx); err != nil {
		return domain.Settings{}, false, err
	}
	db.mu.Lock()
	defer db.mu.Unlock()
	s, ok := db.settings[userID]
	return s, ok, nil
}

func (db *jobDB) SaveSettings(context.Context, string, domain.Settings) error {
	return errors.New("fake: the job never saves settings")
}

func (db *jobDB) Taxpayers(ctx context.Context) ([]usecase.RecipientTaxpayer, error) {
	user, err := requireUser(ctx)
	if err != nil {
		return nil, err
	}
	db.mu.Lock()
	defer db.mu.Unlock()
	return slices.Clone(db.taxpayers[user]), nil
}

func (db *jobDB) Targets(ctx context.Context) ([]usecase.SealedTarget, error) {
	user, err := requireUser(ctx)
	if err != nil {
		return nil, err
	}
	db.mu.Lock()
	defer db.mu.Unlock()
	return slices.Clone(db.devices[user]), nil
}

func (db *jobDB) DeleteDevices(ctx context.Context, ids []string) (int, error) {
	user, err := requireUser(ctx)
	if err != nil {
		return 0, err
	}
	db.mu.Lock()
	defer db.mu.Unlock()
	kept := db.devices[user][:0:0]
	n := 0
	for _, d := range db.devices[user] {
		if slices.Contains(ids, d.DeviceID) {
			db.deleted = append(db.deleted, d.DeviceID)
			n++
			continue
		}
		kept = append(kept, d)
	}
	db.devices[user] = kept
	return n, nil
}

func (db *jobDB) row(t *testing.T, user string, p domain.Period, k domain.Kind) jobLogRow {
	t.Helper()
	db.mu.Lock()
	defer db.mu.Unlock()
	row, ok := db.log[logKey(user, p, k)]
	require.True(t, ok, "no log row for %s %s %s", user, p, k)
	return *row
}

func (db *jobDB) rows() int {
	db.mu.Lock()
	defer db.mu.Unlock()
	return len(db.log)
}

// jobUnits is the per-user unit of work: it hands work the user (RLS) and
// counts the transactions per user. Safe for concurrent use.
type jobUnits struct {
	mu    sync.Mutex
	users []string
}

func (u *jobUnits) ReadWrite(ctx context.Context, userID string, work tx.Work) error {
	u.mu.Lock()
	u.users = append(u.users, userID)
	u.mu.Unlock()
	return work(context.WithValue(ctx, userKey{}, userID))
}

func (u *jobUnits) ReadOnly(context.Context, string, tx.Work) error {
	return errors.New("fake: the job writes in every unit of work")
}

// systemKey marks the System unit of work.
type systemKey struct{}

type jobSystem struct{ calls int }

func (s *jobSystem) System(ctx context.Context, work tx.Work) error {
	s.calls++
	return work(context.WithValue(ctx, systemKey{}, true))
}

// dueCall is one DueReminders call.
type dueCall struct {
	now      time.Time
	dueToday bool
	after    string
	limit    int
	inSystem bool
}

// jobDue scripts due_reminder_users: it pages through reminders by user id.
type jobDue struct {
	mu        sync.Mutex
	reminders []usecase.DueReminder
	calls     []dueCall
	err       error
}

func (d *jobDue) DueReminders(ctx context.Context, now time.Time, dueToday bool, after string, limit int) ([]usecase.DueReminder, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.calls = append(d.calls, dueCall{now: now, dueToday: dueToday, after: after, limit: limit, inSystem: ctx.Value(systemKey{}) != nil})
	if d.err != nil {
		return nil, d.err
	}
	var out []usecase.DueReminder
	for _, r := range d.reminders {
		if r.UserID > after && len(out) < limit {
			out = append(out, r)
		}
	}
	return out, nil
}

// jobOpener opens the fake sealed tokens: ciphertext "sealed:<token>"
// bound to the row it was sealed for.
type jobOpener struct{}

func (jobOpener) Open(ciphertext []byte, _ usecase.TokenBinding) (domain.PushToken, error) {
	raw, ok := strings.CutPrefix(string(ciphertext), "sealed:")
	if !ok {
		return domain.PushToken{}, errors.New("fake: does not open")
	}
	return domain.NewPushToken(raw)
}

// jobEstimates answers scripted summaries per taxpayer and records calls.
type jobEstimates struct {
	mu      sync.Mutex
	answers map[string]usecase.EstimateSummary
	err     error
	calls   []string
}

func (e *jobEstimates) MonthlyEstimate(_ context.Context, userID, taxpayerID string, p domain.Period) (usecase.EstimateSummary, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.calls = append(e.calls, userID+"|"+taxpayerID+"|"+p.String())
	if e.err != nil {
		return usecase.EstimateSummary{}, e.err
	}
	return e.answers[taxpayerID], nil
}

// jobFlags answers fixed module states.
type jobFlags map[string]bool

func (f jobFlags) Enabled(_ context.Context, module string) bool { return f[module] }

// syncAlerts records the unseeded years, safe for concurrent use.
type syncAlerts struct {
	mu       sync.Mutex
	unseeded []int
}

func (a *syncAlerts) CalendarUnseeded(_ context.Context, year int) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.unseeded = append(a.unseeded, year)
}
