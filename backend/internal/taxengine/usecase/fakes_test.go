package usecase_test

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"slices"
	"time"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/apperr"
	"github.com/romodevx/control-contable/backend/internal/platform/clock"
	"github.com/romodevx/control-contable/backend/internal/platform/tx"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain"
	"github.com/romodevx/control-contable/backend/internal/taxengine/usecase"
)

const (
	userA     = "0199d4a2-0000-7000-8000-00000000000a"
	taxpayerA = "0199d4a2-1111-7000-8000-00000000000a"
)

// today is the frozen "now": 14 October 2026, noon in Mexico City.
var today = time.Date(2026, time.October, 14, 12, 0, 0, 0, clock.Location())

// userKey carries the fake unit of work's user, standing in for RLS.
type userKey struct{}

// fakeTx runs work inline with the user in ctx and records the modes used.
type fakeTx struct {
	modes []string
}

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

// fakeAccess refuses writes when readOnly.
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

// fakeStore is an in-memory TaxInputStore for one user's taxpayers.
type fakeStore struct {
	taxpayers map[string]usecase.TaxpayerState
	regimes   map[string]map[fiscal.RegimeCode]usecase.HeldRegime
	values    map[usecase.InputsKey][]usecase.StoredInput
	locks     []string
	replaced  []replaceCall
	failWith  error
}

type replaceCall struct {
	userID string
	key    usecase.InputsKey
	values []usecase.StoredInput
}

func newFakeStore() *fakeStore {
	return &fakeStore{
		taxpayers: map[string]usecase.TaxpayerState{},
		regimes:   map[string]map[fiscal.RegimeCode]usecase.HeldRegime{},
		values:    map[usecase.InputsKey][]usecase.StoredInput{},
	}
}

func (f *fakeStore) hold(taxpayerID string, code fiscal.RegimeCode, r usecase.HeldRegime) {
	if f.regimes[taxpayerID] == nil {
		f.regimes[taxpayerID] = map[fiscal.RegimeCode]usecase.HeldRegime{}
	}
	f.regimes[taxpayerID][code] = r
}

func (f *fakeStore) Taxpayer(_ context.Context, id string, lock usecase.Lock) (usecase.TaxpayerState, error) {
	f.locks = append(f.locks, "taxpayer:"+lockName(lock))
	t, ok := f.taxpayers[id]
	if !ok {
		return usecase.TaxpayerState{}, usecase.ErrTaxpayerNotFound
	}
	return t, nil
}

func (f *fakeStore) Regime(_ context.Context, id string, code fiscal.RegimeCode, lock usecase.Lock) (usecase.HeldRegime, error) {
	f.locks = append(f.locks, "regime:"+lockName(lock))
	r, ok := f.regimes[id][code]
	if !ok {
		return usecase.HeldRegime{}, usecase.ErrRegimeNotHeld
	}
	return r, nil
}

func (f *fakeStore) Values(_ context.Context, key usecase.InputsKey) ([]usecase.StoredInput, error) {
	if f.failWith != nil {
		return nil, f.failWith
	}
	return slices.Clone(f.values[key]), nil
}

func (f *fakeStore) Replace(ctx context.Context, userID string, key usecase.InputsKey, values []usecase.StoredInput) error {
	if f.failWith != nil {
		return f.failWith
	}
	if ctx.Value(userKey{}) != userID {
		return errors.New("fake: replace outside the user's unit of work")
	}
	f.replaced = append(f.replaced, replaceCall{userID: userID, key: key, values: slices.Clone(values)})
	f.values[key] = slices.Clone(values)
	return nil
}

func lockName(l usecase.Lock) string {
	if l == usecase.LockWrite {
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

// defs612 are the 612 PF definitions (contract example regime612).
var defs612 = []domain.InputDef{
	{Key: domain.InputPriorProvisionalPayments, Label: "Pagos provisionales anteriores del ejercicio", Kind: domain.InputMoneyMonthly, Required: true},
	{Key: domain.InputYTDBeforeAppIncome, Label: "Ingresos del ejercicio antes de usar la app", Kind: domain.InputMoney},
	{Key: domain.InputPTUPaid, Label: "PTU pagada en el ejercicio", Kind: domain.InputMoney},
}

// defs601 are the 601 PM definitions (contract example regime601).
var defs601 = []domain.InputDef{
	{Key: domain.InputCoeficienteUtilidad, Label: "Coeficiente de utilidad", Kind: domain.InputRate, Required: true},
	{Key: domain.InputFirstFiscalYear, Label: "Primer ejercicio fiscal", Kind: domain.InputBoolean},
}

type fixture struct {
	tx     *fakeTx
	access *fakeAccess
	store  *fakeStore
	logs   *bytes.Buffer
	uc     *usecase.Inputs
}

// newFixture: taxpayerA of userA holds 612 since 2022-01 (open).
func newFixture() *fixture {
	f := &fixture{tx: &fakeTx{}, access: &fakeAccess{}, store: newFakeStore(), logs: &bytes.Buffer{}}
	f.store.taxpayers[taxpayerA] = usecase.TaxpayerState{}
	f.store.hold(taxpayerA, "612", usecase.HeldRegime{Start: mustPeriod("2022-01"), Definitions: defs612})
	f.uc = usecase.NewInputs(usecase.InputsDeps{
		Tx: f.tx, Access: f.access, Store: f.store, Clock: clock.NewFixed(today),
		Logger: slog.New(slog.NewJSONHandler(f.logs, &slog.HandlerOptions{Level: slog.LevelDebug})),
	})
	return f
}
