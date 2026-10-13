package access_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/platform/access"
	"github.com/romodevx/control-contable/backend/internal/platform/apperr"
	"github.com/romodevx/control-contable/backend/internal/platform/clock"
	"github.com/romodevx/control-contable/backend/internal/platform/tx"
)

const userID = "0199d4a2-8f00-7c44-a1b2-3c4d5e6f7a80"

// registeredAt is the trial start of every fixture; trialEnd is exactly 7 days later.
var (
	registeredAt = time.Date(2026, 10, 13, 16, 0, 0, 0, time.UTC)
	trialEnd     = registeredAt.Add(access.TrialDuration)
)

func trial() access.Entitlement {
	return access.Entitlement{Kind: access.EntitlementTrial, StartsAt: registeredAt, EndsAt: trialEnd}
}

func pro(startsAt, endsAt time.Time) access.Entitlement {
	return access.Entitlement{Kind: access.EntitlementPro, StartsAt: startsAt, EndsAt: endsAt}
}

func comp(startsAt time.Time) access.Entitlement {
	return access.Entitlement{Kind: access.EntitlementComp, StartsAt: startsAt}
}

// fakeTx runs work in place and remembers the user of each unit of work.
type fakeTx struct {
	mu    sync.Mutex
	users []string
	modes []string
}

func (f *fakeTx) record(userID, mode string) error {
	if userID == "" {
		return tx.ErrNoUser
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	f.users = append(f.users, userID)
	f.modes = append(f.modes, mode)
	return nil
}

func (f *fakeTx) ReadWrite(ctx context.Context, userID string, work tx.Work) error {
	if err := f.record(userID, "rw"); err != nil {
		return err
	}
	return work(ctx)
}

func (f *fakeTx) ReadOnly(ctx context.Context, userID string, work tx.Work) error {
	if err := f.record(userID, "ro"); err != nil {
		return err
	}
	return work(ctx)
}

func (f *fakeTx) System(context.Context, tx.Work) error {
	return errors.New("fakeTx: System not expected")
}

// fakeStore keeps one user's entitlements and counters in memory. Reserve is
// the conditional increment the Postgres adapter runs as one UPDATE.
type fakeStore struct {
	mu           sync.Mutex
	entitlements []access.Entitlement
	usage        access.Usage
	limits       []int
	entErr       error
	reserveErr   error
	usageErr     error
}

func (s *fakeStore) Entitlements(_ context.Context, _ string) ([]access.Entitlement, error) {
	return s.entitlements, s.entErr
}

func (s *fakeStore) Reserve(_ context.Context, _ string, kind access.RecordKind, limit int) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.limits = append(s.limits, limit)
	if s.reserveErr != nil {
		return false, s.reserveErr
	}
	counter := &s.usage.Incomes
	if kind == access.RecordExpense {
		counter = &s.usage.Expenses
	}
	if limit != access.Unlimited && *counter >= limit {
		return false, nil
	}
	*counter++
	return true, nil
}

func (s *fakeStore) Usage(context.Context, string) (access.Usage, error) {
	return s.usage, s.usageErr
}

func newPolicy(store *fakeStore, now time.Time) (*access.Policy, *fakeTx) {
	txm := &fakeTx{}
	return access.NewPolicy(txm, store, store, clock.NewFixed(now)), txm
}

func TestDecide(t *testing.T) {
	expiredPro := pro(registeredAt.Add(-60*24*time.Hour), registeredAt.Add(-30*24*time.Hour))
	tests := []struct {
		name      string
		ents      []access.Entitlement
		now       time.Time
		wantLevel access.Level
		wantTrial time.Time
	}{
		{"trial at registration", []access.Entitlement{trial()}, registeredAt, access.LevelTrial, trialEnd},
		{"trial one nanosecond before 7 days", []access.Entitlement{trial()}, trialEnd.Add(-time.Nanosecond), access.LevelTrial, trialEnd},
		{"trial ends at exactly 7 days", []access.Entitlement{trial()}, trialEnd, access.LevelReadOnly, trialEnd},
		{"trial long over", []access.Entitlement{trial()}, trialEnd.Add(30 * 24 * time.Hour), access.LevelReadOnly, trialEnd},
		{"trial not started yet", []access.Entitlement{trial()}, registeredAt.Add(-time.Second), access.LevelReadOnly, trialEnd},
		{"no entitlement at all", nil, registeredAt, access.LevelReadOnly, time.Time{}},
		{"active pro wins over trial", []access.Entitlement{trial(), pro(registeredAt, registeredAt.Add(720*time.Hour))}, registeredAt.Add(time.Hour), access.LevelPro, time.Time{}},
		{"pro ends at its ends_at", []access.Entitlement{pro(registeredAt, trialEnd)}, trialEnd, access.LevelReadOnly, time.Time{}},
		{"lapsed pro is read-only", []access.Entitlement{expiredPro}, registeredAt, access.LevelReadOnly, time.Time{}},
		{"open-ended comp", []access.Entitlement{comp(registeredAt)}, registeredAt.Add(1000 * 24 * time.Hour), access.LevelComp, time.Time{}},
		{"comp after expired trial", []access.Entitlement{trial(), comp(trialEnd)}, trialEnd, access.LevelComp, time.Time{}},
		{"pro wins over comp", []access.Entitlement{comp(registeredAt), pro(registeredAt, trialEnd)}, registeredAt, access.LevelPro, time.Time{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := access.Decide(tt.ents, tt.now)

			assert.Equal(t, tt.wantLevel, got.Level)
			assert.True(t, tt.wantTrial.Equal(got.TrialEndsAt), "trial ends at %v, want %v", got.TrialEndsAt, tt.wantTrial)
		})
	}
}

func TestStatus_Predicates(t *testing.T) {
	tests := []struct {
		level         access.Level
		wantWritable  bool
		wantUnlimited bool
	}{
		{access.LevelTrial, true, false},
		{access.LevelPro, true, true},
		{access.LevelComp, true, true},
		{access.LevelReadOnly, false, false},
	}
	for _, tt := range tests {
		t.Run(string(tt.level), func(t *testing.T) {
			s := access.Status{Level: tt.level}
			assert.Equal(t, tt.wantWritable, s.Writable())
			assert.Equal(t, tt.wantUnlimited, s.Unlimited())
		})
	}
}

func assertCode(t *testing.T, err error, code apperr.Code, target error) {
	t.Helper()
	require.Error(t, err)
	appErr, ok := errors.AsType[*apperr.Error](err)
	require.True(t, ok, "want *apperr.Error, got %T: %v", err, err)
	assert.Equal(t, code, appErr.Code)
	assert.NotEmpty(t, appErr.Message)
	assert.ErrorIs(t, err, target)
}

func TestReserveRecord(t *testing.T) {
	tests := []struct {
		name       string
		ents       []access.Entitlement
		usage      access.Usage
		kind       access.RecordKind
		now        time.Time
		wantErr    error
		wantUsage  access.Usage
		wantLimits []int
	}{
		{
			name: "5th income in trial is allowed", ents: []access.Entitlement{trial()},
			usage: access.Usage{Incomes: 4}, kind: access.RecordIncome, now: registeredAt,
			wantUsage: access.Usage{Incomes: 5}, wantLimits: []int{access.TrialRecordCap},
		},
		{
			name: "6th income in trial is refused", ents: []access.Entitlement{trial()},
			usage: access.Usage{Incomes: 5}, kind: access.RecordIncome, now: registeredAt,
			wantErr: access.ErrTrialCapReached, wantUsage: access.Usage{Incomes: 5}, wantLimits: []int{access.TrialRecordCap},
		},
		{
			name: "5th expense in trial is allowed", ents: []access.Entitlement{trial()},
			usage: access.Usage{Incomes: 5, Expenses: 4}, kind: access.RecordExpense, now: registeredAt,
			wantUsage: access.Usage{Incomes: 5, Expenses: 5}, wantLimits: []int{access.TrialRecordCap},
		},
		{
			name: "6th expense in trial is refused", ents: []access.Entitlement{trial()},
			usage: access.Usage{Expenses: 5}, kind: access.RecordExpense, now: registeredAt,
			wantErr: access.ErrTrialCapReached, wantUsage: access.Usage{Expenses: 5}, wantLimits: []int{access.TrialRecordCap},
		},
		{
			name: "last instant of the trial still reserves", ents: []access.Entitlement{trial()},
			kind: access.RecordIncome, now: trialEnd.Add(-time.Nanosecond),
			wantUsage: access.Usage{Incomes: 1}, wantLimits: []int{access.TrialRecordCap},
		},
		{
			name: "exactly 7 days after registration is read-only", ents: []access.Entitlement{trial()},
			kind: access.RecordIncome, now: trialEnd,
			wantErr: access.ErrReadOnly,
		},
		{
			name: "no entitlement is read-only", kind: access.RecordExpense, now: registeredAt,
			wantErr: access.ErrReadOnly,
		},
		{
			name: "pro reserves past the trial cap", ents: []access.Entitlement{trial(), pro(registeredAt, registeredAt.Add(720*time.Hour))},
			usage: access.Usage{Incomes: 40}, kind: access.RecordIncome, now: trialEnd,
			wantUsage: access.Usage{Incomes: 41}, wantLimits: []int{access.Unlimited},
		},
		{
			name: "comp reserves past the trial cap", ents: []access.Entitlement{comp(registeredAt)},
			usage: access.Usage{Expenses: 5}, kind: access.RecordExpense, now: trialEnd,
			wantUsage: access.Usage{Expenses: 6}, wantLimits: []int{access.Unlimited},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &fakeStore{entitlements: tt.ents, usage: tt.usage}
			policy, txm := newPolicy(store, tt.now)

			err := policy.ReserveRecord(t.Context(), userID, tt.kind)

			if tt.wantErr != nil {
				assertCode(t, err, apperr.CodeEntitlementRequired, tt.wantErr)
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, tt.wantUsage, store.usage)
			assert.Equal(t, tt.wantLimits, store.limits)
			assert.Equal(t, []string{userID}, txm.users)
			assert.Equal(t, []string{"rw"}, txm.modes)
		})
	}
}

func TestReserveRecord_RejectsAnUnknownKind(t *testing.T) {
	store := &fakeStore{entitlements: []access.Entitlement{trial()}}
	policy, _ := newPolicy(store, registeredAt)

	err := policy.ReserveRecord(t.Context(), userID, access.RecordKind("invoice"))

	require.ErrorIs(t, err, access.ErrUnknownRecordKind)
	assert.Empty(t, store.limits, "nothing reserved")
}

func TestReserveRecord_PropagatesFailures(t *testing.T) {
	boom := errors.New("boom")
	tests := []struct {
		name  string
		store *fakeStore
	}{
		{"entitlement lookup", &fakeStore{entErr: boom}},
		{"reservation", &fakeStore{entitlements: []access.Entitlement{trial()}, reserveErr: boom}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			policy, _ := newPolicy(tt.store, registeredAt)

			err := policy.ReserveRecord(t.Context(), userID, access.RecordIncome)

			require.ErrorIs(t, err, boom)
			_, typed := errors.AsType[*apperr.Error](err)
			assert.False(t, typed, "infrastructure failures stay internal errors")
		})
	}
}

func TestReserveRecord_NeedsAUser(t *testing.T) {
	policy, _ := newPolicy(&fakeStore{}, registeredAt)

	err := policy.ReserveRecord(t.Context(), "", access.RecordIncome)

	require.ErrorIs(t, err, tx.ErrNoUser)
}

func TestEnsureWritable(t *testing.T) {
	tests := []struct {
		name    string
		ents    []access.Entitlement
		now     time.Time
		wantErr bool
	}{
		{"trial", []access.Entitlement{trial()}, registeredAt, false},
		{"trial at its cap is still writable (edits do not create)", []access.Entitlement{trial()}, registeredAt, false},
		{"after the trial", []access.Entitlement{trial()}, trialEnd, true},
		{"pro", []access.Entitlement{pro(registeredAt, trialEnd)}, registeredAt, false},
		{"comp", []access.Entitlement{comp(registeredAt)}, trialEnd, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &fakeStore{entitlements: tt.ents, usage: access.Usage{Incomes: 5, Expenses: 5}}
			policy, txm := newPolicy(store, tt.now)

			err := policy.EnsureWritable(t.Context(), userID)

			if tt.wantErr {
				assertCode(t, err, apperr.CodeEntitlementRequired, access.ErrReadOnly)
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, []string{"ro"}, txm.modes)
			assert.Empty(t, store.limits, "never reserves")
		})
	}
}

func TestCheckTaxpayerLimit(t *testing.T) {
	tests := []struct {
		name     string
		ents     []access.Entitlement
		now      time.Time
		existing int
		wantErr  bool
	}{
		{"pro below the cap", []access.Entitlement{pro(registeredAt, trialEnd)}, registeredAt, access.TaxpayerCap - 1, false},
		{"pro at the cap", []access.Entitlement{pro(registeredAt, trialEnd)}, registeredAt, access.TaxpayerCap, true},
		{"comp at the cap", []access.Entitlement{comp(registeredAt)}, registeredAt, access.TaxpayerCap, true},
		{"trial is not capped (PRD 7)", []access.Entitlement{trial()}, registeredAt, 3 * access.TaxpayerCap, false},
		{"read-only below the cap", []access.Entitlement{trial()}, trialEnd, access.TaxpayerCap - 1, false},
		{"read-only at the cap", []access.Entitlement{trial()}, trialEnd, access.TaxpayerCap, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			policy, _ := newPolicy(&fakeStore{entitlements: tt.ents}, tt.now)

			err := policy.CheckTaxpayerLimit(t.Context(), userID, tt.existing)

			if tt.wantErr {
				assertCode(t, err, apperr.CodeTaxpayerLimitReached, access.ErrTaxpayerLimitReached)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestSummary(t *testing.T) {
	tests := []struct {
		name          string
		ents          []access.Entitlement
		now           time.Time
		usage         access.Usage
		wantLevel     access.Level
		wantIncomes   int
		wantExpenses  int
		wantUnlimited bool
	}{
		{"fresh trial", []access.Entitlement{trial()}, registeredAt, access.Usage{}, access.LevelTrial, 5, 5, false},
		{"trial part used", []access.Entitlement{trial()}, registeredAt, access.Usage{Incomes: 2, Expenses: 5}, access.LevelTrial, 3, 0, false},
		{"read-only has nothing left", []access.Entitlement{trial()}, trialEnd, access.Usage{Incomes: 1}, access.LevelReadOnly, 0, 0, false},
		{"counter above the cap never goes negative", []access.Entitlement{trial()}, registeredAt, access.Usage{Incomes: 9}, access.LevelTrial, 0, 5, false},
		{"pro is unlimited", []access.Entitlement{pro(registeredAt, trialEnd)}, registeredAt, access.Usage{Incomes: 30}, access.LevelPro, 0, 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			policy, txm := newPolicy(&fakeStore{entitlements: tt.ents, usage: tt.usage}, tt.now)

			got, err := policy.Summary(t.Context(), userID)

			require.NoError(t, err)
			assert.Equal(t, tt.wantLevel, got.Status.Level)
			incomes, limited := got.Remaining(access.RecordIncome)
			assert.Equal(t, !tt.wantUnlimited, limited)
			assert.Equal(t, tt.wantIncomes, incomes)
			expenses, _ := got.Remaining(access.RecordExpense)
			assert.Equal(t, tt.wantExpenses, expenses)
			assert.Equal(t, []string{"ro"}, txm.modes)
		})
	}
}

func TestSummary_PropagatesUsageFailure(t *testing.T) {
	boom := errors.New("boom")
	policy, _ := newPolicy(&fakeStore{entitlements: []access.Entitlement{trial()}, usageErr: boom}, registeredAt)

	_, err := policy.Summary(t.Context(), userID)

	require.ErrorIs(t, err, boom)
}
