package annual_test

import (
	"bytes"
	"cmp"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"maps"
	"slices"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/apperr"
	"github.com/romodevx/control-contable/backend/internal/platform/clock"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
	"github.com/romodevx/control-contable/backend/internal/platform/tx"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain"
	adomain "github.com/romodevx/control-contable/backend/internal/taxengine/domain/annual"
	"github.com/romodevx/control-contable/backend/internal/taxengine/usecase/annual"
)

const (
	userA     = "0199d4a2-0000-7000-8000-00000000000a"
	userB     = "0199d4a2-0000-7000-8000-00000000000b"
	fisicaA   = "0199d4a2-1111-7000-8000-00000000000a"
	fisica2A  = "0199d4a2-1111-7000-8000-00000000002a"
	moralA    = "0199d4a2-2222-7000-8000-00000000000a"
	fisicaB   = "0199d4a2-1111-7000-8000-00000000000b"
	newRowID  = "0199d4b1-0000-7000-8000-000000000001"
	clientID  = "0199d4b1-2b3c-7d4e-8f5a-6b7c8d9e0f01"
	clientID2 = "0199d4b1-2b3c-7d4e-8f5a-6b7c8d9e0f02"
)

// today is the frozen "now": 14 October 2026, noon in Mexico City.
var today = time.Date(2026, time.October, 14, 12, 0, 0, 0, clock.Location())

// stamp is the fake database's now().
var stamp = time.Date(2026, time.October, 14, 18, 0, 0, 0, time.UTC)

// userKey carries the fake unit of work's user, standing in for RLS.
type userKey struct{}

func ctxUser(ctx context.Context) string { s, _ := ctx.Value(userKey{}).(string); return s }

// fakeTx runs work inline with the user in ctx and records the modes used.
type fakeTx struct{ modes []string }

func (f *fakeTx) ReadWrite(ctx context.Context, userID string, work tx.Work) error {
	f.modes = append(f.modes, "rw")
	return work(context.WithValue(ctx, userKey{}, userID))
}

func (f *fakeTx) ReadOnly(ctx context.Context, userID string, work tx.Work) error {
	f.modes = append(f.modes, "ro")
	return work(context.WithValue(ctx, userKey{}, userID))
}

// errReadOnly stands in for access.Policy's client-ready 402.
var errReadOnly = apperr.New(apperr.CodeEntitlementRequired, "Read-only.")

type fakeAccess struct {
	readOnly bool
	calls    int
}

func (f *fakeAccess) EnsureWritable(context.Context, string) error {
	f.calls++
	if f.readOnly {
		return errReadOnly
	}
	return nil
}

type ownedTaxpayer struct {
	owner string
	annual.Taxpayer
	regimes []annual.HeldRegime
}

type ownedDeduction struct {
	owner string
	row   annual.StoredDeduction
}

type ownedStatement struct {
	owner string
	row   annual.StoredStatement
}

type factsKey struct {
	taxpayerID string
	year       int
}

type ownedFacts struct {
	owner string
	row   annual.StoredFacts
}

type ownedFiled struct {
	owner  string
	months []annual.FiledMonth
}

// fakeDB is every annual store in memory; rows of another user than the
// unit of work's are invisible, as under RLS.
type fakeDB struct {
	taxpayers  map[string]ownedTaxpayer
	deductions map[string]ownedDeduction
	statements map[string]ownedStatement
	facts      map[factsKey]ownedFacts
	filed      map[annual.FiledKey]ownedFiled
	taxInputs  map[factsKey]map[fiscal.RegimeCode]map[domain.InputKey]money.Cents
	// calls records the lock and write calls, in order.
	calls    []string
	failWith error
	seq      int
}

func newFakeDB() *fakeDB {
	return &fakeDB{
		taxpayers: map[string]ownedTaxpayer{}, deductions: map[string]ownedDeduction{},
		statements: map[string]ownedStatement{}, facts: map[factsKey]ownedFacts{},
		filed: map[annual.FiledKey]ownedFiled{}, taxInputs: map[factsKey]map[fiscal.RegimeCode]map[domain.InputKey]money.Cents{},
	}
}

func (f *fakeDB) record(format string, args ...any) {
	f.calls = append(f.calls, fmt.Sprintf(format, args...))
}

func (f *fakeDB) taxpayer(ctx context.Context, id string) (ownedTaxpayer, bool) {
	t, ok := f.taxpayers[id]
	return t, ok && t.owner == ctxUser(ctx)
}

func (f *fakeDB) Taxpayer(ctx context.Context, id string, lock annual.Lock) (annual.Taxpayer, error) {
	f.record("taxpayer:%s", lockName(lock))
	if f.failWith != nil {
		return annual.Taxpayer{}, f.failWith
	}
	t, ok := f.taxpayer(ctx, id)
	if !ok {
		return annual.Taxpayer{}, annual.ErrTaxpayerNotFound
	}
	return t.Taxpayer, nil
}

func (f *fakeDB) Regimes(ctx context.Context, id string) ([]annual.HeldRegime, error) {
	t, _ := f.taxpayer(ctx, id)
	return slices.Clone(t.regimes), nil
}

func (f *fakeDB) LockRegime(ctx context.Context, id string, code fiscal.RegimeCode) error {
	f.record("regime:%s", code)
	t, _ := f.taxpayer(ctx, id)
	if !slices.ContainsFunc(t.regimes, func(r annual.HeldRegime) bool { return r.Code == code }) {
		return annual.ErrRegimeNotHeld
	}
	return nil
}

func (f *fakeDB) LockList(_ context.Context, id string, list annual.List, year int) error {
	f.record("list:%d:%d", list, year)
	return nil
}

func (f *fakeDB) LockDeductionID(_ context.Context, id string) error {
	f.record("lock-id")
	return nil
}

func (f *fakeDB) FindDeduction(ctx context.Context, id string, lock annual.Lock) (annual.StoredDeduction, error) {
	f.record("find:%s", lockName(lock))
	d, ok := f.deductions[id]
	if !ok || d.owner != ctxUser(ctx) {
		return annual.StoredDeduction{}, annual.ErrDeductionNotFound
	}
	return d.row, nil
}

func (f *fakeDB) InsertDeduction(ctx context.Context, d annual.NewDeduction) error {
	f.record("insert")
	if f.failWith != nil {
		return f.failWith
	}
	if d.UserID != ctxUser(ctx) {
		return errors.New("fake: insert outside the user's unit of work")
	}
	if _, taken := f.deductions[d.ID]; taken {
		return annual.ErrIDTaken
	}
	f.deductions[d.ID] = ownedDeduction{owner: d.UserID, row: annual.StoredDeduction{
		ID: d.ID, TaxpayerID: d.TaxpayerID, FiscalYear: d.PaidOn.Year(), DeductionFields: d.DeductionFields,
		CreatedAt: stamp, UpdatedAt: stamp,
	}}
	return nil
}

func (f *fakeDB) UpdateDeduction(_ context.Context, id string, fields annual.DeductionFields) error {
	f.record("update")
	d := f.deductions[id]
	d.row.DeductionFields, d.row.FiscalYear, d.row.UpdatedAt = fields, fields.PaidOn.Year(), stamp.Add(time.Hour)
	f.deductions[id] = d
	return nil
}

func (f *fakeDB) SetDeductionDeleted(_ context.Context, id string, deleted bool) error {
	f.record("deleted:%t", deleted)
	d := f.deductions[id]
	d.row.Deleted = deleted
	f.deductions[id] = d
	return nil
}

func (f *fakeDB) liveDeductions(ctx context.Context, taxpayerID string, year int) []annual.StoredDeduction {
	var out []annual.StoredDeduction
	for _, d := range f.deductions {
		if d.owner == ctxUser(ctx) && d.row.TaxpayerID == taxpayerID && d.row.FiscalYear == year && !d.row.Deleted {
			out = append(out, d.row)
		}
	}
	return out
}

func (f *fakeDB) CountLiveDeductions(ctx context.Context, taxpayerID string, year int) (int, error) {
	f.record("count")
	return len(f.liveDeductions(ctx, taxpayerID, year)), nil
}

func (f *fakeDB) ListDeductions(ctx context.Context, taxpayerID string, year int, category adomain.Category) ([]annual.StoredDeduction, error) {
	if f.failWith != nil {
		return nil, f.failWith
	}
	rows := f.liveDeductions(ctx, taxpayerID, year)
	if category != "" {
		rows = slices.DeleteFunc(rows, func(d annual.StoredDeduction) bool { return d.Category != category })
	}
	// No particular order: reverse id order proves the use case sorts.
	slices.SortFunc(rows, func(a, b annual.StoredDeduction) int { return cmp.Compare(b.ID, a.ID) })
	return rows, nil
}

func (f *fakeDB) LockStatementID(_ context.Context, id string) error {
	f.record("lock-id")
	return nil
}

func (f *fakeDB) FindStatement(ctx context.Context, id string, lock annual.Lock) (annual.StoredStatement, error) {
	f.record("find:%s", lockName(lock))
	s, ok := f.statements[id]
	if !ok || s.owner != ctxUser(ctx) {
		return annual.StoredStatement{}, annual.ErrStatementNotFound
	}
	return s.row, nil
}

func (f *fakeDB) InsertStatement(ctx context.Context, s annual.NewStatement) error {
	f.record("insert")
	if s.UserID != ctxUser(ctx) {
		return errors.New("fake: insert outside the user's unit of work")
	}
	if _, taken := f.statements[s.ID]; taken {
		return annual.ErrIDTaken
	}
	f.seq++
	created := stamp.Add(time.Duration(f.seq) * time.Minute)
	f.statements[s.ID] = ownedStatement{owner: s.UserID, row: annual.StoredStatement{
		ID: s.ID, TaxpayerID: s.TaxpayerID, SalaryFields: s.SalaryFields, CreatedAt: created, UpdatedAt: created,
	}}
	return nil
}

func (f *fakeDB) UpdateStatement(_ context.Context, id string, fields annual.SalaryFields) error {
	f.record("update")
	s := f.statements[id]
	s.row.SalaryFields, s.row.UpdatedAt = fields, stamp.Add(time.Hour)
	f.statements[id] = s
	return nil
}

func (f *fakeDB) DeleteStatement(_ context.Context, id string) error {
	f.record("delete")
	delete(f.statements, id)
	return nil
}

func (f *fakeDB) yearStatements(ctx context.Context, taxpayerID string, year int) []annual.StoredStatement {
	var out []annual.StoredStatement
	for _, s := range f.statements {
		if s.owner == ctxUser(ctx) && s.row.TaxpayerID == taxpayerID && s.row.FiscalYear == year {
			out = append(out, s.row)
		}
	}
	slices.SortFunc(out, func(a, b annual.StoredStatement) int {
		return cmp.Or(a.CreatedAt.Compare(b.CreatedAt), cmp.Compare(a.ID, b.ID))
	})
	return out
}

func (f *fakeDB) CountStatements(ctx context.Context, taxpayerID string, year int) (int, error) {
	f.record("count")
	return len(f.yearStatements(ctx, taxpayerID, year)), nil
}

func (f *fakeDB) ListStatements(ctx context.Context, taxpayerID string, year int) ([]annual.StoredStatement, error) {
	if f.failWith != nil {
		return nil, f.failWith
	}
	return f.yearStatements(ctx, taxpayerID, year), nil
}

func (f *fakeDB) Facts(ctx context.Context, taxpayerID string, year int) (annual.StoredFacts, bool, error) {
	if f.failWith != nil {
		return annual.StoredFacts{}, false, f.failWith
	}
	row, ok := f.facts[factsKey{taxpayerID, year}]
	if !ok || row.owner != ctxUser(ctx) {
		return annual.StoredFacts{}, false, nil
	}
	return row.row, true, nil
}

func (f *fakeDB) UpsertFacts(ctx context.Context, userID, taxpayerID string, year int, facts annual.Facts) error {
	f.record("upsert")
	if userID != ctxUser(ctx) {
		return errors.New("fake: upsert outside the user's unit of work")
	}
	f.facts[factsKey{taxpayerID, year}] = ownedFacts{owner: userID, row: annual.StoredFacts{Facts: facts, UpdatedAt: stamp}}
	return nil
}

func (f *fakeDB) TaxInputMoney(_ context.Context, taxpayerID string, year int, keys []domain.InputKey) (map[fiscal.RegimeCode]map[domain.InputKey]money.Cents, error) {
	out := map[fiscal.RegimeCode]map[domain.InputKey]money.Cents{}
	for code, values := range f.taxInputs[factsKey{taxpayerID, year}] {
		out[code] = maps.Clone(values)
		maps.DeleteFunc(out[code], func(k domain.InputKey, _ money.Cents) bool { return !slices.Contains(keys, k) })
	}
	return out, nil
}

func (f *fakeDB) FiledMonths(ctx context.Context, key annual.FiledKey) ([]annual.FiledMonth, *time.Time, error) {
	if f.failWith != nil {
		return nil, nil, f.failWith
	}
	row, ok := f.filed[key]
	if !ok || row.owner != ctxUser(ctx) || len(row.months) == 0 {
		return nil, nil, nil
	}
	at := stamp
	return slices.Clone(row.months), &at, nil
}

func (f *fakeDB) ReplaceFiled(ctx context.Context, userID string, key annual.FiledKey, months []annual.FiledMonth) error {
	f.record("replace")
	if userID != ctxUser(ctx) {
		return errors.New("fake: replace outside the user's unit of work")
	}
	sorted := slices.Clone(months)
	slices.SortFunc(sorted, func(a, b annual.FiledMonth) int { return cmp.Compare(a.Month, b.Month) })
	f.filed[key] = ownedFiled{owner: userID, months: sorted}
	return nil
}

func lockName(l annual.Lock) string {
	if l == annual.LockWrite {
		return "write"
	}
	return "none"
}

func mustPeriod(raw string) domain.Period {
	p, err := domain.ParsePeriod(raw)
	if err != nil {
		panic(err)
	}
	return p
}

func mustDate(raw string) fiscal.Date {
	d, err := fiscal.ParseDate(raw)
	if err != nil {
		panic(err)
	}
	return d
}

func ptr[T any](v T) *T { return &v }

// world is userA with fisicaA (612 since 2022-01, 605 since 2026-01),
// fisica2A (612 only), moralA (601 since 2024-01) and userB with fisicaB.
type world struct {
	db     *fakeDB
	tx     *fakeTx
	access *fakeAccess
	logs   *bytes.Buffer
}

var defs612 = []domain.InputDef{
	{Key: domain.InputPriorProvisionalPayments, Label: "Pagos provisionales anteriores del ejercicio", Kind: domain.InputMoneyMonthly},
	{Key: domain.InputPTUPaid, Label: "PTU pagada en el ejercicio", Kind: domain.InputMoney},
	{Key: domain.InputPriorYearLosses, Label: "Pérdidas fiscales de ejercicios anteriores", Kind: domain.InputMoney},
	{Key: adomain.InputAnnualDepreciation, Label: "Depreciación del ejercicio", Kind: domain.InputMoney},
}

var defs601 = []domain.InputDef{
	{Key: domain.InputCoeficienteUtilidad, Label: "Coeficiente de utilidad", Kind: domain.InputRate},
	{Key: domain.InputInterestWithholding, Label: "Retenciones de intereses", Kind: domain.InputMoney},
	{Key: domain.InputPTUPaid, Label: "PTU pagada en el ejercicio", Kind: domain.InputMoney},
}

func newWorld() *world {
	w := &world{db: newFakeDB(), tx: &fakeTx{}, access: &fakeAccess{}, logs: &bytes.Buffer{}}
	w.db.taxpayers[fisicaA] = ownedTaxpayer{owner: userA, Taxpayer: annual.Taxpayer{PersonType: fiscal.PersonFisica}, regimes: []annual.HeldRegime{
		{Code: "605", Start: mustPeriod("2026-01")},
		{Code: "612", Start: mustPeriod("2022-01"), Definitions: defs612},
	}}
	w.db.taxpayers[fisica2A] = ownedTaxpayer{owner: userA, Taxpayer: annual.Taxpayer{PersonType: fiscal.PersonFisica}, regimes: []annual.HeldRegime{
		{Code: "612", Start: mustPeriod("2022-01"), Definitions: defs612},
	}}
	w.db.taxpayers[moralA] = ownedTaxpayer{owner: userA, Taxpayer: annual.Taxpayer{PersonType: fiscal.PersonMoral}, regimes: []annual.HeldRegime{
		{Code: "601", Start: mustPeriod("2024-01"), Definitions: defs601},
	}}
	w.db.taxpayers[fisicaB] = ownedTaxpayer{owner: userB, Taxpayer: annual.Taxpayer{PersonType: fiscal.PersonFisica}, regimes: []annual.HeldRegime{
		{Code: "612", Start: mustPeriod("2022-01"), Definitions: defs612},
	}}
	return w
}

func (w *world) archive(id string) {
	t := w.db.taxpayers[id]
	t.Archived = true
	w.db.taxpayers[id] = t
}

func (w *world) logger() *slog.Logger {
	return slog.New(slog.NewJSONHandler(w.logs, &slog.HandlerOptions{Level: slog.LevelDebug}))
}

func fixedID() (string, error) { return newRowID, nil }

// requireCode asserts err is an *apperr.Error of code and returns its
// fields as "field/code".
func requireCode(t *testing.T, err error, code apperr.Code) []string {
	t.Helper()
	e, ok := errors.AsType[*apperr.Error](err)
	require.Truef(t, ok, "want *apperr.Error, got %v", err)
	require.Equal(t, code, e.Code, e.Error())
	out := []string{}
	for _, fe := range e.Fields {
		out = append(out, fe.Field+"/"+fe.Code)
	}
	return out
}
