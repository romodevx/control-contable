package usecase_test

import (
	"bytes"
	"cmp"
	"context"
	"errors"
	"log/slog"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/expenses/domain"
	"github.com/romodevx/control-contable/backend/internal/expenses/usecase"
	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/access"
	"github.com/romodevx/control-contable/backend/internal/platform/apperr"
	"github.com/romodevx/control-contable/backend/internal/platform/clock"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
	"github.com/romodevx/control-contable/backend/internal/platform/tx"
)

// Fixtures: the ids are this suite's own.
const (
	userA     = "0199d4a2-0000-7000-8000-00000000000a"
	userB     = "0199d4a2-0000-7000-8000-00000000000b"
	taxpayerA = "0199d4a2-1111-7000-8000-00000000000a"
	taxpayerM = "0199d4a2-1111-7000-8000-0000000000aa" // userA's persona moral
	taxpayerB = "0199d4a2-1111-7000-8000-00000000000b"
	expense1  = "0199d4a3-2b3c-7e4d-9f50-6b7c8d9e0f12"
	expense2  = "0199d4a3-2b3c-7e4d-9f50-6b7c8d9e0f13"
	serverID  = "0199d4a3-ffff-7000-8000-000000000001"
	cfdiUUID  = "3A9C7E21-5B4D-4F6A-8E2C-9D1B7A3F5C60"
)

// today is the frozen "now": 13 October 2026, noon in Mexico City.
var today = time.Date(2026, time.October, 13, 12, 0, 0, 0, clock.Location())

func date(t *testing.T, raw string) fiscal.Date {
	t.Helper()
	d, err := fiscal.ParseDate(raw)
	require.NoError(t, err)
	return d
}

func uuidOf(t *testing.T, raw string) fiscal.CFDIUUID {
	t.Helper()
	u, err := fiscal.ParseCFDIUUID(raw)
	require.NoError(t, err)
	return u
}

// userKey carries the fake unit of work's user, standing in for RLS.
type userKey struct{}

// ctxAs is a context inside userID's fake unit of work.
func ctxAs(t *testing.T, userID string) context.Context {
	t.Helper()
	return context.WithValue(t.Context(), userKey{}, userID)
}

func currentUser(ctx context.Context) string {
	user, _ := ctx.Value(userKey{}).(string)
	return user
}

// fakeTx runs work inline with the user in ctx and records the modes used.
type fakeTx struct {
	mu    sync.Mutex
	modes []string
}

func (f *fakeTx) ReadWrite(ctx context.Context, userID string, work tx.Work) error {
	return f.run(ctx, "rw", userID, work)
}

func (f *fakeTx) ReadOnly(ctx context.Context, userID string, work tx.Work) error {
	return f.run(ctx, "ro", userID, work)
}

func (f *fakeTx) run(ctx context.Context, mode, userID string, work tx.Work) error {
	f.mu.Lock()
	f.modes = append(f.modes, mode)
	f.mu.Unlock()
	// Like pgdb.TxManager, a failed unit of work comes back joined with its
	// (here absent) rollback error.
	if err := work(context.WithValue(ctx, userKey{}, userID)); err != nil {
		return errors.Join(err, nil)
	}
	return nil
}

// fakeAccess records the access calls; refuse makes them fail.
type fakeAccess struct {
	readOnly  bool
	capSpent  bool
	writable  int
	reserved  []access.RecordKind
	lastCalls []string
}

var errFakeReadOnly = apperr.Wrap(access.ErrReadOnly, apperr.CodeEntitlementRequired, "read-only")

func (a *fakeAccess) EnsureWritable(context.Context, string) error {
	a.writable++
	a.lastCalls = append(a.lastCalls, "writable")
	if a.readOnly {
		return errFakeReadOnly
	}
	return nil
}

func (a *fakeAccess) ReserveRecord(_ context.Context, _ string, kind access.RecordKind) error {
	a.lastCalls = append(a.lastCalls, "reserve")
	if a.readOnly {
		return errFakeReadOnly
	}
	if a.capSpent {
		return apperr.Wrap(access.ErrTrialCapReached, apperr.CodeEntitlementRequired, "cap").
			WithFields(apperr.FieldError{Field: "expenses_remaining", Code: access.DetailTrialCapReached})
	}
	a.reserved = append(a.reserved, kind)
	return nil
}

// hold is one regime window of a fake taxpayer.
type hold struct {
	code       fiscal.RegimeCode
	start, end string
	options    domain.Options
}

// fakeTaxpayer is one taxpayer of the fake directory.
type fakeTaxpayer struct {
	owner      string
	personType fiscal.PersonType
	archived   bool
	regimes    []hold
}

// fakeTaxpayers implements both taxpayer ports over a map.
type fakeTaxpayers struct {
	byID    map[string]*fakeTaxpayer
	checked []string
}

func notFoundErr() error { return apperr.New(apperr.CodeNotFound, "The taxpayer was not found.") }

func (f *fakeTaxpayers) find(ctx context.Context, id string) (*fakeTaxpayer, error) {
	t, ok := f.byID[id]
	if !ok || t.owner != currentUser(ctx) {
		return nil, notFoundErr()
	}
	return t, nil
}

func (f *fakeTaxpayers) CheckRegime(ctx context.Context, _, taxpayerID, regimeCode, period string) error {
	f.checked = append(f.checked, regimeCode+"@"+period)
	t, err := f.find(ctx, taxpayerID)
	if err != nil {
		return err
	}
	if t.archived {
		return apperr.New(apperr.CodeBusinessRuleViolation, "archived").
			WithFields(apperr.FieldError{Field: "taxpayerId", Code: "taxpayer_archived"})
	}
	for _, h := range t.regimes {
		if string(h.code) == regimeCode && period >= h.start && (h.end == "" || period <= h.end) {
			return nil
		}
	}
	return apperr.New(apperr.CodeBusinessRuleViolation, "not active").
		WithFields(apperr.FieldError{Field: "regime_code", Code: "regime_not_active"})
}

func (f *fakeTaxpayers) Profile(ctx context.Context, _, taxpayerID string) (usecase.TaxpayerProfile, error) {
	t, err := f.find(ctx, taxpayerID)
	if err != nil {
		return usecase.TaxpayerProfile{}, err
	}
	p := usecase.TaxpayerProfile{PersonType: t.personType, Archived: t.archived}
	for _, h := range t.regimes {
		p.Regimes = append(p.Regimes, usecase.RegimeHold{Code: h.code, StartPeriod: h.start, EndPeriod: h.end, Options: h.options})
	}
	return p, nil
}

// fakeExpenses is the expense store over a map, scoped by the ctx user.
type fakeExpenses struct {
	mu      sync.Mutex
	rows    map[string]storedRow
	locked  []string
	clockAt time.Time
}

type storedRow struct {
	owner string
	e     usecase.StoredExpense
}

func newFakeExpenses() *fakeExpenses {
	return &fakeExpenses{rows: map[string]storedRow{}, clockAt: today.Add(-time.Hour)}
}

func (f *fakeExpenses) LockID(_ context.Context, id string) error {
	f.locked = append(f.locked, id)
	return nil
}

func (f *fakeExpenses) Find(ctx context.Context, id string, _ usecase.Lock) (usecase.StoredExpense, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	row, ok := f.rows[id]
	if !ok || row.owner != currentUser(ctx) {
		return usecase.StoredExpense{}, usecase.ErrExpenseNotFound
	}
	return row.e, nil
}

func (f *fakeExpenses) Insert(ctx context.Context, userID string, e usecase.StoredExpense) (usecase.StoredExpense, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if _, taken := f.rows[e.ID]; taken {
		return usecase.StoredExpense{}, usecase.ErrIDConflict
	}
	if userID != currentUser(ctx) {
		panic("insert for another user")
	}
	e.CreatedAt, e.UpdatedAt = f.clockAt, f.clockAt
	f.rows[e.ID] = storedRow{owner: userID, e: e}
	return e, nil
}

func (f *fakeExpenses) Update(ctx context.Context, e usecase.StoredExpense) (usecase.StoredExpense, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	row, ok := f.rows[e.ID]
	if !ok || row.owner != currentUser(ctx) {
		return usecase.StoredExpense{}, usecase.ErrExpenseNotFound
	}
	e.CreatedAt, e.UpdatedAt = row.e.CreatedAt, f.clockAt.Add(time.Minute)
	row.e = e
	f.rows[e.ID] = row
	return e, nil
}

func (f *fakeExpenses) SetDeleted(ctx context.Context, id string, deleted bool) (usecase.StoredExpense, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	row, ok := f.rows[id]
	if !ok || row.owner != currentUser(ctx) {
		return usecase.StoredExpense{}, usecase.ErrExpenseNotFound
	}
	row.e.Deleted = deleted
	f.rows[id] = row
	return row.e, nil
}

func (f *fakeExpenses) CFDIInUse(ctx context.Context, taxpayerID, exceptID string, u fiscal.CFDIUUID) (bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for id, row := range f.rows {
		e := row.e
		if row.owner == currentUser(ctx) && id != exceptID && e.TaxpayerID == taxpayerID && !e.Deleted && e.CFDIUUID == u {
			return true, nil
		}
	}
	return false, nil
}

func (f *fakeExpenses) live(ctx context.Context, q usecase.PeriodQuery) []usecase.StoredExpense {
	var out []usecase.StoredExpense
	for _, row := range f.rows {
		e := row.e
		switch {
		case row.owner != currentUser(ctx), e.Deleted, e.TaxpayerID != q.TaxpayerID:
		case e.PaidOn.Before(q.From), !e.PaidOn.Before(q.Until):
		case q.Regime != "" && e.Regime != q.Regime:
		default:
			out = append(out, e)
		}
	}
	slices.SortFunc(out, func(a, b usecase.StoredExpense) int {
		if c := b.PaidOn.Compare(a.PaidOn); c != 0 {
			return c
		}
		return cmp.Compare(b.ID, a.ID)
	})
	return out
}

func (f *fakeExpenses) List(ctx context.Context, q usecase.ListQuery) ([]usecase.StoredExpense, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var out []usecase.StoredExpense
	for _, e := range f.live(ctx, q.PeriodQuery) {
		if q.After != nil && !before(e, *q.After) {
			continue
		}
		out = append(out, e)
		if len(out) == q.Limit {
			break
		}
	}
	return out, nil
}

// before reports whether e sorts after the cursor in (paid_on, id) DESC order.
func before(e usecase.StoredExpense, c usecase.Cursor) bool {
	if c := e.PaidOn.Compare(c.PaidOn); c != 0 {
		return c < 0
	}
	return e.ID < c.ID
}

func (f *fakeExpenses) Totals(ctx context.Context, q usecase.PeriodQuery) (usecase.Totals, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var t usecase.Totals
	for _, e := range f.live(ctx, q) {
		t.Count++
		t.Amount += e.Amount
		t.IVA += e.IVA
		t.ISRWithheld += e.Withheld.ISR
		t.IVAWithheld += e.Withheld.IVA
	}
	return t, nil
}

// fakeCategories serves a few catalog rows: the business column for every
// regime except 626 física (isr_not_applicable) and 606.
type fakeCategories struct{}

var businessRules = []usecase.CategoryEntry{
	{Rule: domain.CategoryRule{Category: "software", Deductibility: domain.Deductible, DefaultDeductible: true, IVACreditable: true}, Name: "Software y suscripciones", LegalBasis: "LISR 103"},
	{Rule: domain.CategoryRule{Category: "papeleria", Deductibility: domain.Deductible, DefaultDeductible: true, IVACreditable: true}, Name: "Papelería", LegalBasis: "LISR 103"},
	{Rule: domain.CategoryRule{Category: "combustible", Deductibility: domain.Deductible, DefaultDeductible: true, IVACreditable: true, ElectronicPaymentOnly: true}, Name: "Combustibles", LegalBasis: "LISR 27", Note: "Sólo deducible si lo pagas con tarjeta"},
	{Rule: domain.CategoryRule{Category: "restaurantes", Deductibility: domain.PartiallyDeductible, IVACreditable: true}, Name: "Consumo en restaurantes", LegalBasis: "LISR 28"},
	{Rule: domain.CategoryRule{Category: "gastos_personales", Deductibility: domain.NotDeductible}, Name: "Gastos personales", LegalBasis: "LISR 151"},
	{Rule: domain.CategoryRule{Category: "predial", Deductibility: domain.Deductible, DefaultDeductible: true, IVACreditable: true}, Name: "Impuesto predial", LegalBasis: "LISR 103, 115"},
	{Rule: domain.CategoryRule{Category: "mantenimiento", Deductibility: domain.Deductible, DefaultDeductible: true, IVACreditable: true}, Name: "Mantenimiento", LegalBasis: "LISR 103, 115"},
	{Rule: domain.CategoryRule{Category: "servicios_profesionales", Deductibility: domain.Deductible, DefaultDeductible: true, IVACreditable: true}, Name: "Honorarios", LegalBasis: "LISR 25"},
}

var errCategoryStore = &storeError{"category store down"}

type storeError struct{ msg string }

func (e *storeError) Error() string { return e.msg }

func (fakeCategories) Rules(_ context.Context, regime fiscal.RegimeCode, person fiscal.PersonType) ([]usecase.CategoryEntry, error) {
	if regime == "999" {
		return nil, errCategoryStore
	}
	out := slices.Clone(businessRules)
	if regime == fiscal.RegimeRESICO && person == fiscal.PersonFisica {
		for i := range out {
			out[i].Rule.Deductibility, out[i].Rule.DefaultDeductible = domain.ISRNotApplicable, false
			out[i].Note = "En RESICO tus gastos no reducen tu ISR"
		}
	}
	return out, nil
}

func (c fakeCategories) Rule(ctx context.Context, regime fiscal.RegimeCode, person fiscal.PersonType, category domain.CategoryCode) (usecase.CategoryEntry, error) {
	rules, err := c.Rules(ctx, regime, person)
	if err != nil {
		return usecase.CategoryEntry{}, err
	}
	for _, r := range rules {
		if r.Rule.Category == category {
			return r, nil
		}
	}
	return usecase.CategoryEntry{}, usecase.ErrCategoryNotFound
}

// fakeParameters is the RateCatalog plus the dated form parameters of 2026.
type fakeParameters struct {
	provisional bool
	uncovered   bool
}

func (p fakeParameters) Rate(key fiscal.RateKey, on fiscal.Date) (money.Rate, bool) {
	if p.uncovered {
		return 0, false
	}
	rates := map[fiscal.RateKey]string{
		fiscal.KeyIVAGeneral: "0.16", fiscal.KeyIVAZero: "0", fiscal.KeyIVABorder: "0.08",
		fiscal.KeyISRWithholdingFees: "0.10", fiscal.KeyISRWithholdingRent: "0.10", domain.KeyIVAWithholdingFreight: "0.04",
	}
	raw, ok := rates[key]
	if !ok || (key == fiscal.KeyIVABorder && on.Year() > 2026) {
		return 0, false
	}
	r, err := money.ParseRate(raw)
	return r, err == nil
}

const cashLimit money.Cents = 2_000_00

func (p fakeParameters) Amount(key domain.AmountKey, _ fiscal.Date) (money.Cents, bool) {
	return cashLimit, key == domain.KeyCashPaymentLimit && !p.uncovered
}

func (p fakeParameters) On(fiscal.Date) (usecase.DatedParameters, error) {
	if p.uncovered {
		return usecase.DatedParameters{}, usecase.ErrParametersUnavailable
	}
	return usecase.DatedParameters{
		IVARatesAllowed:  []fiscal.IVARate{fiscal.IVARate16, fiscal.IVARate8, fiscal.IVARate0, fiscal.IVAExempt},
		CashPaymentLimit: cashLimit,
		Provisional:      p.provisional,
	}, nil
}

// harness wires the use cases over the fakes.
type harness struct {
	tx         *fakeTx
	access     *fakeAccess
	taxpayers  *fakeTaxpayers
	expenses   *fakeExpenses
	parameters fakeParameters
	logs       *bytes.Buffer
	uc         *usecase.Expenses
}

func newHarness(t *testing.T) *harness {
	t.Helper()
	h := &harness{
		tx: &fakeTx{}, access: &fakeAccess{}, expenses: newFakeExpenses(), logs: &bytes.Buffer{},
		taxpayers: &fakeTaxpayers{byID: map[string]*fakeTaxpayer{
			taxpayerA: {owner: userA, personType: fiscal.PersonFisica, regimes: []hold{
				{code: fiscal.RegimeBusiness, start: "2026-01"},
				{code: fiscal.RegimeRESICO, start: "2022-01", end: "2025-12"},
				{code: fiscal.RegimeRent, start: "2026-01", options: domain.Options{DeductionMode: domain.DeductionBlind35}},
			}},
			taxpayerM: {owner: userA, personType: fiscal.PersonMoral, regimes: []hold{{code: "601", start: "2026-01"}}},
			taxpayerB: {owner: userB, personType: fiscal.PersonFisica, regimes: []hold{{code: fiscal.RegimeBusiness, start: "2026-01"}}},
		}},
	}
	h.wire()
	return h
}

func (h *harness) wire() {
	h.uc = usecase.NewExpenses(usecase.Deps{
		Tx: h.tx, Access: h.access, Regimes: h.taxpayers, Taxpayers: h.taxpayers,
		Expenses: h.expenses, Categories: fakeCategories{}, Parameters: h.parameters,
		Clock:  clock.NewFixed(today),
		NewID:  func() (string, error) { return serverID, nil },
		Logger: slog.New(slog.NewJSONHandler(h.logs, &slog.HandlerOptions{Level: slog.LevelDebug})),
	})
}

// softwareDraft is a 612 software expense of 500.00 + IVA 80.00 paid by card.
func softwareDraft(t *testing.T) domain.Draft {
	t.Helper()
	return domain.Draft{
		Category: "software", Amount: 500_00, IVARate: fiscal.IVARate16, IVA: 80_00,
		PaymentMethod: domain.PaymentCardDebit, PaidOn: date(t, "2026-09-12"), HasCFDI: true,
	}
}

func createInput(t *testing.T, id string) usecase.CreateInput {
	t.Helper()
	return usecase.CreateInput{
		UserID: userA, TaxpayerID: taxpayerA, ID: id, Regime: fiscal.RegimeBusiness,
		Draft: softwareDraft(t), Description: "  Suscripción anual de diseño ",
	}
}

// requireAppErr asserts err is an *apperr.Error with code and, when given,
// the first detail's field and code.
func requireAppErr(t *testing.T, err error, code apperr.Code, fieldAndDetail ...string) {
	t.Helper()
	_ = appErrOf(t, err, code, fieldAndDetail...)
}

// appErrOf is requireAppErr returning the error for further assertions.
func appErrOf(t *testing.T, err error, code apperr.Code, fieldAndDetail ...string) *apperr.Error {
	t.Helper()
	require.Error(t, err)
	appErr, ok := err.(*apperr.Error) //nolint:errorlint // the use cases return the *apperr.Error itself
	require.Truef(t, ok, "want *apperr.Error, got %T: %v", err, err)
	require.Equal(t, code, appErr.Code, appErr.Error())
	if len(fieldAndDetail) == 2 {
		require.NotEmpty(t, appErr.Fields, appErr.Error())
		require.Equal(t, fieldAndDetail[0], appErr.Fields[0].Field)
		require.Equal(t, fieldAndDetail[1], appErr.Fields[0].Code)
	}
	return appErr
}

func warningCodes(ws []domain.Warning) []string {
	out := make([]string, 0, len(ws))
	for _, w := range ws {
		out = append(out, string(w.Code)+"@"+w.Field)
	}
	return out
}

func logsMention(h *harness, needle string) bool { return strings.Contains(h.logs.String(), needle) }
