package usecase_test

import (
	"bytes"
	"cmp"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/incomes/domain"
	"github.com/romodevx/control-contable/backend/internal/incomes/usecase"
	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/access"
	"github.com/romodevx/control-contable/backend/internal/platform/apperr"
	"github.com/romodevx/control-contable/backend/internal/platform/clock"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
	"github.com/romodevx/control-contable/backend/internal/platform/tx"
)

// Fixture ids: synthetic, owned by this suite.
const (
	userA      = "0199d4a2-0000-7000-8000-00000000000a"
	taxpayerA  = "0199d4a2-1111-7000-8000-00000000000a"
	taxpayerA2 = "0199d4a2-1111-7000-8000-0000000000a2"
	pmA        = "0199d4a2-1111-7000-8000-0000000000aa"
	platformA  = "0199d4a2-1111-7000-8000-0000000000a5"
	archivedA  = "0199d4a2-1111-7000-8000-0000000000af"
	incomeID   = "0199d4a3-1a2b-7d3c-8e4f-5a6b7c8d9e01"
	generated  = "0199d4a3-ffff-7000-8000-000000000001"
	cfdiLower  = "6f3b2c1d-8a4e-4b7f-9c2d-1e5a7b9c3d4f"
	cfdiUpper  = "6F3B2C1D-8A4E-4B7F-9C2D-1E5A7B9C3D4F"
)

// today is the frozen "now": 13 October 2026, noon in Mexico City.
var today = time.Date(2026, time.October, 13, 12, 0, 0, 0, clock.Location())

func date(t testing.TB, raw string) fiscal.Date {
	t.Helper()
	d, err := fiscal.ParseDate(raw)
	require.NoError(t, err)
	return d
}

func rate(t testing.TB, raw string) money.Rate {
	t.Helper()
	r, err := money.ParseRate(raw)
	require.NoError(t, err)
	return r
}

// userKey carries the fake unit of work's user, standing in for RLS.
type userKey struct{}

// fakeTx runs work inline with the user in ctx and records the modes used.
type fakeTx struct{ modes []string }

func (f *fakeTx) ReadWrite(ctx context.Context, userID string, work tx.Work) error {
	return f.run(ctx, "rw", userID, work)
}

func (f *fakeTx) ReadOnly(ctx context.Context, userID string, work tx.Work) error {
	return f.run(ctx, "ro", userID, work)
}

func (f *fakeTx) run(ctx context.Context, mode, userID string, work tx.Work) error {
	if userID == "" {
		return tx.ErrNoUser
	}
	f.modes = append(f.modes, mode)
	return work(context.WithValue(ctx, userKey{}, userID))
}

// fakeAccess is the access policy: read-only or capped, and it counts calls.
type fakeAccess struct {
	readOnly bool
	used     int // incomes created (the trial counter)
	limit    int // 0 = unlimited
	ensures  int
	reserves int
}

var (
	errReadOnly = apperr.Wrap(access.ErrReadOnly, apperr.CodeEntitlementRequired, "Your trial has ended.")
	errCap      = apperr.Wrap(access.ErrTrialCapReached, apperr.CodeEntitlementRequired, "Cap.").
			WithFields(apperr.FieldError{Field: "incomes_remaining", Code: access.DetailTrialCapReached})
)

func (f *fakeAccess) EnsureWritable(context.Context, string) error {
	f.ensures++
	if f.readOnly {
		return errReadOnly
	}
	return nil
}

func (f *fakeAccess) ReserveRecord(_ context.Context, _ string, kind access.RecordKind) error {
	f.reserves++
	if kind != access.RecordIncome {
		return fmt.Errorf("unexpected kind %q", kind)
	}
	if f.readOnly {
		return errReadOnly
	}
	if f.limit > 0 && f.used >= f.limit {
		return errCap
	}
	f.used++
	return nil
}

// regimeCall is one CheckRegime call.
type regimeCall struct{ taxpayer, regime, period string }

// fakeTaxpayers serves profiles and answers CheckRegime from per-regime
// windows of periods ("YYYY-MM" inclusive bounds; an empty end is open).
type fakeTaxpayers struct {
	profiles map[string]usecase.TaxpayerProfile
	windows  map[string][2]string // taxpayer|regime -> [start, end]
	calls    []regimeCall
	fail     error // returned by Profile when set
}

var errNotActive = apperr.Wrap(errors.New("regime not active"), apperr.CodeBusinessRuleViolation, "Not active.").
	WithFields(apperr.FieldError{Field: "regime_code", Code: "regime_not_active"})

var errTaxpayerNotFound = apperr.Wrap(errors.New("taxpayer not found"), apperr.CodeNotFound, "Not found.")

func newFakeTaxpayers() *fakeTaxpayers {
	fisica := func(regimes ...usecase.RegimeProfile) usecase.TaxpayerProfile {
		return usecase.TaxpayerProfile{PersonType: fiscal.PersonFisica, Regimes: regimes}
	}
	return &fakeTaxpayers{
		profiles: map[string]usecase.TaxpayerProfile{
			taxpayerA:  fisica(usecase.RegimeProfile{Code: "612"}, usecase.RegimeProfile{Code: "606"}),
			taxpayerA2: fisica(usecase.RegimeProfile{Code: "612"}),
			pmA:        {PersonType: fiscal.PersonMoral, Regimes: []usecase.RegimeProfile{{Code: "601"}}},
			platformA: fisica(usecase.RegimeProfile{
				Code: "625", PlatformActivities: []domain.PlatformActivity{domain.ActivityLodging}, RFCGivenToPlatform: true,
			}),
			archivedA: {PersonType: fiscal.PersonFisica, Archived: true, Regimes: []usecase.RegimeProfile{{Code: "612"}}},
		},
		windows: map[string][2]string{
			taxpayerA + "|612": {"2022-01", ""}, taxpayerA + "|606": {"2026-09", "2026-09"},
			taxpayerA2 + "|612": {"2022-01", ""}, pmA + "|601": {"2022-01", ""},
			platformA + "|625": {"2022-01", ""}, archivedA + "|612": {"2022-01", ""},
		},
	}
}

func (f *fakeTaxpayers) Profile(_ context.Context, _ string, taxpayerID string) (usecase.TaxpayerProfile, error) {
	if f.fail != nil {
		return usecase.TaxpayerProfile{}, f.fail
	}
	p, ok := f.profiles[taxpayerID]
	if !ok {
		return usecase.TaxpayerProfile{}, errTaxpayerNotFound
	}
	return p, nil
}

func (f *fakeTaxpayers) CheckRegime(_ context.Context, _ string, taxpayerID, regime, period string) error {
	f.calls = append(f.calls, regimeCall{taxpayerID, regime, period})
	p, ok := f.profiles[taxpayerID]
	if !ok {
		return errTaxpayerNotFound
	}
	if p.Archived {
		return apperr.New(apperr.CodeBusinessRuleViolation, "Archived.")
	}
	w, ok := f.windows[taxpayerID+"|"+regime]
	if !ok || period < w[0] || (w[1] != "" && period > w[1]) {
		return errNotActive
	}
	return nil
}

// fakeCatalog is the RateCatalog port: open-ended 2026 rates, the border 8 %
// until 2026-12-31, and parameter sets covering 2026 (published) and 2027
// (the provisional fallback).
type fakeCatalog struct {
	rates       map[fiscal.RateKey]money.Rate
	provisional map[int]bool // fiscal year -> provisional; absent = not covered
}

func newFakeCatalog(t testing.TB) *fakeCatalog {
	return &fakeCatalog{
		rates: map[fiscal.RateKey]money.Rate{
			fiscal.KeyIVAGeneral: rate(t, "0.16"), fiscal.KeyIVAZero: rate(t, "0"), fiscal.KeyIVABorder: rate(t, "0.08"),
			fiscal.KeyISRWithholdingFees: rate(t, "0.10"), fiscal.KeyISRWithholdingRent: rate(t, "0.10"),
			domain.KeyISRPlatformLodging: rate(t, "0.04"),
		},
		provisional: map[int]bool{2026: false, 2027: true},
	}
}

func (c *fakeCatalog) Rate(key fiscal.RateKey, on fiscal.Date) (money.Rate, bool) {
	if _, covered := c.provisional[on.Year()]; !covered {
		return 0, false
	}
	if key == fiscal.KeyIVABorder && on.Year() > 2026 {
		return 0, false
	}
	r, ok := c.rates[key]
	return r, ok
}

var errNotCovered = errors.New("no parameter set covers the date")

func (c *fakeCatalog) Provisional(on fiscal.Date) (bool, error) {
	p, ok := c.provisional[on.Year()]
	if !ok {
		return false, errNotCovered
	}
	return p, nil
}

// fakeStore is the incomes store in memory. It stands in for RLS by owning
// the rows of one user only: another user's ids are in foreign.
type fakeStore struct {
	mu      sync.Mutex
	rows    map[string]usecase.StoredIncome
	foreign map[string]bool // ids another user holds (PK collision, invisible)
	// lockedIDs are the client ids LockID serialized, in order.
	lockedIDs []string
	inserts   int
	updates   int
	deletes   int
	lastList  struct {
		filter usecase.ListFilter
		after  usecase.Keyset
		limit  int
	}
	// onInsert, when set, runs before the insert and may return its error.
	onInsert func(s *fakeStore, in usecase.NewIncome) error
	fail     error // returned by Find when set
	clock    time.Time
}

func newFakeStore() *fakeStore {
	return &fakeStore{rows: map[string]usecase.StoredIncome{}, foreign: map[string]bool{}, clock: today}
}

func (s *fakeStore) LockID(_ context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.lockedIDs = append(s.lockedIDs, id)
	return nil
}

func (s *fakeStore) Find(_ context.Context, id string, _ usecase.Lock) (usecase.StoredIncome, error) {
	if s.fail != nil {
		return usecase.StoredIncome{}, s.fail
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	row, ok := s.rows[id]
	if !ok {
		return usecase.StoredIncome{}, usecase.ErrIncomeNotFound
	}
	return row, nil
}

func (s *fakeStore) Insert(_ context.Context, in usecase.NewIncome) error {
	if s.onInsert != nil {
		if err := s.onInsert(s, in); err != nil {
			return err
		}
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, taken := s.rows[in.ID]; taken || s.foreign[in.ID] {
		return usecase.ErrIDTaken
	}
	s.inserts++
	s.rows[in.ID] = usecase.StoredIncome{
		ID: in.ID, TaxpayerID: in.TaxpayerID, Fields: in.Fields,
		BookedOn: bookedOn(in.Fields), CreatedAt: s.clock, UpdatedAt: s.clock,
	}
	return nil
}

// bookedOn mimics the trigger: accrual (601 here) books on invoiced_on.
func bookedOn(f usecase.Fields) fiscal.Date {
	if f.Regime == "601" {
		return f.Dates.InvoicedOn
	}
	return f.Dates.CollectedOn
}

func (s *fakeStore) Update(_ context.Context, id string, f usecase.Fields) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	row, ok := s.rows[id]
	if !ok {
		return usecase.ErrIncomeNotFound
	}
	s.updates++
	row.Fields, row.BookedOn, row.UpdatedAt = f, bookedOn(f), s.clock.Add(time.Minute)
	s.rows[id] = row
	return nil
}

func (s *fakeStore) SetDeleted(_ context.Context, id string, deleted bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	row, ok := s.rows[id]
	if !ok {
		return usecase.ErrIncomeNotFound
	}
	s.deletes++
	row.Deleted = deleted
	s.rows[id] = row
	return nil
}

func (s *fakeStore) CFDIUUIDInUse(_ context.Context, taxpayerID string, uuid fiscal.CFDIUUID, exceptID string) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for id, row := range s.rows {
		if id != exceptID && !row.Deleted && row.TaxpayerID == taxpayerID && row.CFDIUUID == uuid {
			return true, nil
		}
	}
	return false, nil
}

func (s *fakeStore) List(_ context.Context, f usecase.ListFilter, after usecase.Keyset, limit int) ([]usecase.StoredIncome, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.lastList.filter, s.lastList.after, s.lastList.limit = f, after, limit
	var out []usecase.StoredIncome
	for _, row := range s.rows {
		if s.inFilter(row, f) && (after.ID == "" || keysetBefore(row, after)) {
			out = append(out, row)
		}
	}
	slices.SortFunc(out, func(a, b usecase.StoredIncome) int {
		if c := b.BookedOn.Compare(a.BookedOn); c != 0 {
			return c
		}
		return cmp.Compare(b.ID, a.ID)
	})
	return out[:min(limit, len(out))], nil
}

func keysetBefore(row usecase.StoredIncome, after usecase.Keyset) bool {
	if c := row.BookedOn.Compare(after.BookedOn); c != 0 {
		return c < 0
	}
	return row.ID < after.ID
}

func (s *fakeStore) inFilter(row usecase.StoredIncome, f usecase.ListFilter) bool {
	return !row.Deleted && row.TaxpayerID == f.TaxpayerID &&
		!row.BookedOn.Before(f.From) && row.BookedOn.Before(f.Until) &&
		(f.Regime == "" || row.Regime == f.Regime)
}

func (s *fakeStore) Totals(_ context.Context, f usecase.ListFilter) (usecase.PeriodTotals, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var t usecase.PeriodTotals
	for _, row := range s.rows {
		if s.inFilter(row, f) {
			t.Count++
			t.Amount += row.Amount
			t.IVA += row.IVA
			t.ISRWithheld += row.Withheld.ISR
			t.IVAWithheld += row.Withheld.IVA
		}
	}
	return t, nil
}

func (s *fakeStore) IncomeBasis(_ context.Context, p fiscal.PersonType, regime fiscal.RegimeCode) (domain.IncomeBasis, error) {
	if p == fiscal.PersonMoral && regime == "601" {
		return domain.BasisAccrual, nil
	}
	return domain.BasisCash, nil
}

// harness is the use case over the fakes.
type harness struct {
	tx        *fakeTx
	access    *fakeAccess
	taxpayers *fakeTaxpayers
	store     *fakeStore
	catalog   *fakeCatalog
	logs      *bytes.Buffer
	now       time.Time
	uc        *usecase.Incomes
}

func newHarness(t *testing.T) *harness {
	t.Helper()
	h := &harness{
		tx: &fakeTx{}, access: &fakeAccess{}, taxpayers: newFakeTaxpayers(), store: newFakeStore(),
		catalog: newFakeCatalog(t), logs: &bytes.Buffer{}, now: today,
	}
	return h.wire()
}

// at moves the harness clock and rewires the use case.
func (h *harness) at(now time.Time) *harness {
	h.now = now
	return h.wire()
}

func (h *harness) wire() *harness {
	h.uc = usecase.NewIncomes(usecase.Deps{
		Tx: h.tx, Store: h.store, Taxpayers: h.taxpayers, Access: h.access, Rates: h.catalog,
		Clock:  clock.NewFixed(h.now),
		NewID:  func() (string, error) { return generated, nil },
		Logger: slog.New(slog.NewJSONHandler(h.logs, &slog.HandlerOptions{Level: slog.LevelDebug})),
	})
	return h
}

// us1 is US-I1: 10,000.00 + IVA 1,600.00 billed by a 612 física to a persona
// moral, withholdings as suggested, collected on 2026-09-12.
func us1(t testing.TB) usecase.IncomeInput {
	return usecase.IncomeInput{
		Regime: "612", Counterparty: domain.CounterpartyMoral,
		Amount: 10_000_00, IVARate: fiscal.IVARate16, IVA: 1_600_00,
		Withheld: fiscal.Withholdings{ISR: 1_000_00, IVA: 1_066_67},
		Dates:    domain.Dates{InvoicedOn: date(t, "2026-09-05"), CollectedOn: date(t, "2026-09-12")},
		CFDIUUID: cfdiLower, Description: "Diseño de logotipo",
	}
}

func create(id, taxpayer string, in usecase.IncomeInput) usecase.CreateInput {
	return usecase.CreateInput{ID: id, TaxpayerID: taxpayer, Income: in}
}

// requireApp asserts an *apperr.Error with code and, when given, the first
// field error. It returns the error's fields.
func requireApp(t *testing.T, err error, code apperr.Code, field ...string) []apperr.FieldError {
	t.Helper()
	require.Error(t, err)
	app, ok := errors.AsType[*apperr.Error](err)
	require.True(t, ok, "want *apperr.Error, got %T: %v", err, err)
	require.Equal(t, code, app.Code, "error: %v", err)
	if len(field) == 2 {
		require.NotEmpty(t, app.Fields, "error: %v", err)
		require.Equal(t, field[0], app.Fields[0].Field)
		require.Equal(t, field[1], app.Fields[0].Code)
	}
	return app.Fields
}
