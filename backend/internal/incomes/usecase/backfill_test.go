package usecase_test

import (
	"context"
	"errors"
	"maps"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/incomes/domain"
	"github.com/romodevx/control-contable/backend/internal/incomes/usecase"
	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/clock"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
)

// Backfill fixture ids (WH-08).
const (
	resicoA      = "0199d4a2-1111-7000-8000-0000000000a6"
	legacyOwner  = "0199d4a3-0b0b-7000-8000-000000000001" // US-W1, typed before WH-06
	legacyPublic = "0199d4a3-0b0b-7000-8000-000000000002"
	olderTable   = "0199d4a3-0b0b-7000-8000-000000000003"
	current      = "0199d4a3-0b0b-7000-8000-000000000004"
	legacyRent   = "0199d4a3-0b0b-7000-8000-000000000005"
	notHeld      = "0199d4a3-0b0b-7000-8000-000000000006"
	deletedOne   = "0199d4a3-0b0b-7000-8000-000000000007"
)

// Underived implements usecase.BackfillStore over the fake rows: deleted
// ones included, ordered by id.
func (s *fakeStore) Underived(_ context.Context, version string, lock usecase.Lock) ([]usecase.StoredIncome, error) {
	if s.fail != nil {
		return nil, s.fail
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.backfillLocks = append(s.backfillLocks, lock)
	var out []usecase.StoredIncome
	for _, row := range s.rows {
		w := row.Withholding
		if w.TableVersion != version || w.Withholder == "" || row.Counterparty == "" || row.Operation == "" {
			out = append(out, row)
		}
	}
	slices.SortFunc(out, func(a, b usecase.StoredIncome) int { return strings.Compare(a.ID, b.ID) })
	return out, nil
}

// SaveDerivation implements usecase.BackfillStore.
func (s *fakeStore) SaveDerivation(ctx context.Context, id string, f usecase.Fields, w fiscal.WithholdingRecord) error {
	if s.saveErr != nil {
		return s.saveErr
	}
	deleted := s.rows[id].Deleted
	if err := s.Update(ctx, id, f, w); err != nil {
		return err
	}
	row := s.rows[id]
	row.Deleted = deleted
	s.rows[id] = row
	return nil
}

// backfillHarness is the backfill over the incomes fakes, with a RESICO
// taxpayer for the owner's example.
type backfillHarness struct {
	*harness
	backfill *usecase.Backfill
}

func newBackfillHarness(t *testing.T) *backfillHarness {
	t.Helper()
	h := newHarness(t)
	h.taxpayers.profiles[resicoA] = usecase.TaxpayerProfile{
		PersonType: fiscal.PersonFisica, Regimes: []usecase.RegimeProfile{{Code: "626"}},
	}
	h.taxpayers.windows[resicoA+"|626"] = [2]string{"2022-01", ""}
	return &backfillHarness{harness: h, backfill: usecase.NewBackfill(usecase.BackfillDeps{
		Tx: h.tx, Store: h.store, Taxpayers: h.taxpayers, Withholdings: h.catalog, Clock: clock.NewFixed(h.now),
	})}
}

// seed stores a row as it was written, with its typed amounts.
func (h *backfillHarness) seed(row usecase.StoredIncome) {
	h.store.rows[row.ID] = row
}

// legacy is an income written before WH-06: counterparty_type only, typed
// withholdings, no derivation.
func legacy(t testing.TB, id, taxpayer string, regime fiscal.RegimeCode, c domain.Counterparty, amount, iva, isr, ivaWithheld money.Cents) usecase.StoredIncome {
	return usecase.StoredIncome{
		ID: id, TaxpayerID: taxpayer, CounterpartyType: c,
		Fields: usecase.Fields{
			Regime: regime, Amount: amount, IVARate: fiscal.IVARate16, IVA: iva,
			Dates: domain.Dates{InvoicedOn: date(t, "2026-07-08"), CollectedOn: date(t, "2026-07-10")},
		},
		Withholding: fiscal.WithholdingRecord{Withholdings: fiscal.Withholdings{ISR: isr, IVA: ivaWithheld}},
	}
}

func TestBackfill_DryRunReportsAndWritesNothing(t *testing.T) {
	h := newBackfillHarness(t)
	h.seed(legacy(t, legacyOwner, resicoA, "626", domain.CounterpartyMoral, 100_000_00, 16_000_00, 0, 0))
	before := maps.Clone(h.store.rows)

	got, err := h.backfill.Run(t.Context(), userA, false)

	require.NoError(t, err)
	assert.Equal(t, []usecase.Rederived{{ID: legacyOwner, AmountsChanged: true, FactsAssumed: true}}, got.Rederived)
	assert.Empty(t, got.NeedsInput)
	assert.Equal(t, before, h.store.rows, "a dry run writes nothing")
	assert.Equal(t, []string{"ro"}, h.tx.modes, "in a read-only unit of work")
	assert.Equal(t, []usecase.Lock{usecase.LockNone}, h.store.backfillLocks)
}

// US-W1, the owner's case: RESICO PF bills a company $100,000.00 + IVA of
// professional services, typed with no withholding before WH-06.
func TestBackfill_ApplyDerivesTheOwnersExample(t *testing.T) {
	h := newBackfillHarness(t)
	h.seed(legacy(t, legacyOwner, resicoA, "626", domain.CounterpartyMoral, 100_000_00, 16_000_00, 0, 0))

	_, err := h.backfill.Run(t.Context(), userA, true)

	require.NoError(t, err)
	row := h.store.rows[legacyOwner]
	assert.Equal(t, fiscal.ReceptorPM, row.Counterparty)
	assert.Equal(t, fiscal.OperationProfessionalServices, row.Operation)
	assert.False(t, row.ViaPlatform)
	assert.Equal(t, money.Cents(1_250_00), row.Withholding.ISR)
	assert.Equal(t, money.Cents(10_666_67), row.Withholding.IVA)
	assert.Equal(t, []fiscal.RuleID{"ISR-4", "IVA-6"}, row.Withholding.RuleIDs)
	assert.Equal(t, fiscal.WithholdingTableVersion, row.Withholding.TableVersion)
	assert.Equal(t, []string{"rw"}, h.tx.modes)
	assert.Equal(t, []usecase.Lock{usecase.LockUpdate}, h.store.backfillLocks, "rows are locked while re-derived")
}

func TestBackfill_LegacyFactsFollowTheSpec(t *testing.T) {
	h := newBackfillHarness(t)
	h.seed(legacy(t, legacyPublic, taxpayerA, "612", domain.CounterpartyPublic, 5_000_00, 800_00, 0, 0))
	rent := legacy(t, legacyRent, taxpayerA, "606", domain.CounterpartyMoral, 30_000_00, 0, 3_000_00, 0)
	rent.IVARate = fiscal.IVAExempt
	rent.Dates = domain.Dates{InvoicedOn: date(t, "2026-09-02"), CollectedOn: date(t, "2026-09-03")}
	h.seed(rent)
	lodging := legacy(t, deletedOne, platformA, "625", domain.CounterpartyPlatform, 10_000_00, 1_600_00, 0, 0)
	lodging.Deleted = true
	h.seed(lodging)

	got, err := h.backfill.Run(t.Context(), userA, true)

	require.NoError(t, err)
	assert.Len(t, got.Rederived, 3)
	public := h.store.rows[legacyPublic]
	assert.Equal(t, fiscal.ReceptorPFNoActivityOrPublic, public.Counterparty)
	assert.Equal(t, fiscal.OperationProfessionalServices, public.Operation, "612's default")
	assert.Zero(t, public.Withholding.ISR+public.Withholding.IVA)
	derivedRent := h.store.rows[legacyRent]
	assert.Equal(t, fiscal.OperationRentRealEstate, derivedRent.Operation)
	assert.True(t, derivedRent.ResidentialRent, "606 exempt rent is housing")
	assert.Equal(t, money.Cents(3_000_00), derivedRent.Withholding.ISR, "WH-03b")
	platform := h.store.rows[deletedOne]
	assert.True(t, platform.ViaPlatform)
	assert.Equal(t, fiscal.ReceptorPM, platform.Counterparty)
	assert.True(t, platform.Deleted, "a deleted income is derived too, so a restore finds its facts")
	assert.Equal(t, fiscal.WithholdingTableVersion, platform.Withholding.TableVersion)
}

func TestBackfill_ReDerivesAnOlderTableVersionAndSkipsTheCurrentOne(t *testing.T) {
	h := newBackfillHarness(t)
	old := legacy(t, olderTable, taxpayerA, "612", "", 10_000_00, 1_600_00, 1, 1)
	old.Counterparty, old.Operation = fiscal.ReceptorPM, fiscal.OperationProfessionalServices
	old.CounterpartyType = domain.CounterpartyMoral
	old.Withholding.TableVersion, old.Withholding.RuleIDs = "2025.9", []fiscal.RuleID{"ISR-5", "IVA-6"}
	old.Withholding.Withholder = fiscal.WithholderReceptor
	h.seed(old)
	fresh := old
	fresh.ID, fresh.Withholding.TableVersion = current, fiscal.WithholdingTableVersion
	h.seed(fresh)

	got, err := h.backfill.Run(t.Context(), userA, true)

	require.NoError(t, err)
	assert.Equal(t, []usecase.Rederived{{ID: olderTable, AmountsChanged: true}}, got.Rederived)
	row := h.store.rows[olderTable]
	assert.Equal(t, money.Cents(1_000_00), row.Withholding.ISR)
	assert.Equal(t, money.Cents(1_066_67), row.Withholding.IVA)
	assert.Equal(t, fiscal.WithholdingTableVersion, row.Withholding.TableVersion)
	assert.Equal(t, money.Cents(1), h.store.rows[current].Withholding.ISR, "the current version is left alone")
}

func TestBackfill_SecondRunChangesNothing(t *testing.T) {
	h := newBackfillHarness(t)
	h.seed(legacy(t, legacyOwner, resicoA, "626", domain.CounterpartyMoral, 100_000_00, 16_000_00, 0, 0))
	_, err := h.backfill.Run(t.Context(), userA, true)
	require.NoError(t, err)
	updates := h.store.updates

	got, err := h.backfill.Run(t.Context(), userA, true)

	require.NoError(t, err)
	assert.Empty(t, got.Rederived)
	assert.Equal(t, updates, h.store.updates)
}

func TestBackfill_RowsTheRulesRefuseNeedUserInput(t *testing.T) {
	h := newBackfillHarness(t)
	gone := legacy(t, notHeld, taxpayerA, "626", domain.CounterpartyMoral, 1_000_00, 160_00, 0, 0)
	h.seed(gone) // taxpayerA never held 626
	unknown := legacy(t, legacyPublic, taxpayerA, "612", "", 1_000_00, 160_00, 0, 0)
	h.seed(unknown) // no counterparty_type to read
	future := legacy(t, legacyOwner, resicoA, "626", domain.CounterpartyMoral, 1_000_00, 160_00, 0, 0)
	future.Dates = domain.Dates{InvoicedOn: date(t, "2026-12-01"), CollectedOn: date(t, "2026-12-02")}
	h.seed(future) // after today: the domain refuses it

	got, err := h.backfill.Run(t.Context(), userA, true)

	require.NoError(t, err)
	assert.Empty(t, got.Rederived)
	assert.Equal(t, []string{legacyOwner, legacyPublic, notHeld}, got.NeedsInput)
	assert.Zero(t, h.store.updates, "left as stored")
}

func TestBackfill_Failures(t *testing.T) {
	boom := errors.New("boom")
	t.Run("store", func(t *testing.T) {
		h := newBackfillHarness(t)
		h.store.fail = boom
		_, err := h.backfill.Run(t.Context(), userA, false)
		require.ErrorIs(t, err, boom)
	})
	t.Run("profile", func(t *testing.T) {
		h := newBackfillHarness(t)
		h.seed(legacy(t, legacyOwner, resicoA, "626", domain.CounterpartyMoral, 100_000_00, 16_000_00, 0, 0))
		h.taxpayers.fail = boom
		_, err := h.backfill.Run(t.Context(), userA, false)
		require.ErrorIs(t, err, boom)
	})
	t.Run("save", func(t *testing.T) {
		h := newBackfillHarness(t)
		h.seed(legacy(t, legacyOwner, resicoA, "626", domain.CounterpartyMoral, 100_000_00, 16_000_00, 0, 0))
		h.store.saveErr = boom
		_, err := h.backfill.Run(t.Context(), userA, true)
		require.ErrorIs(t, err, boom)
	})
	t.Run("no user", func(t *testing.T) {
		h := newBackfillHarness(t)
		_, err := h.backfill.Run(t.Context(), "", true)
		require.Error(t, err)
	})
}
