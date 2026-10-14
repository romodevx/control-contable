package usecase_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/incomes/domain"
	"github.com/romodevx/control-contable/backend/internal/incomes/usecase"
	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/apperr"
	"github.com/romodevx/control-contable/backend/internal/platform/clock"
)

func TestCreate_US1(t *testing.T) {
	h := newHarness(t)
	got, err := h.uc.Create(t.Context(), userA, create(incomeID, taxpayerA, us1(t)))
	require.NoError(t, err)

	assert.True(t, got.Created, "201")
	assert.Equal(t, incomeID, got.Income.ID)
	assert.Equal(t, taxpayerA, got.Income.TaxpayerID)
	assert.Equal(t, "2026-09-12", got.Income.BookedOn.String())
	assert.False(t, got.Income.PendingCollection)
	assert.Equal(t, cfdiUpper, got.Income.CFDIUUID.String(), "stored upper case")
	assert.Equal(t, "Diseño de logotipo", got.Income.Description)
	assert.Empty(t, got.Warnings)
	assert.Equal(t, 1, h.access.reserves, "one trial slot")
	assert.Equal(t, 1, h.access.used)
	assert.Equal(t, []regimeCall{{taxpayerA, "612", "2026-09"}}, h.taxpayers.calls, "one call per distinct period")
	assert.Equal(t, []string{"rw"}, h.tx.modes, "one unit of work")

	stored, err := h.uc.Get(t.Context(), userA, taxpayerA, incomeID)
	require.NoError(t, err)
	assert.Equal(t, got.Income, stored)
}

func TestCreate_EachDistinctPeriodIsChecked(t *testing.T) {
	h := newHarness(t)
	in := us1(t)
	in.Dates = domain.Dates{InvoicedOn: date(t, "2026-08-30"), CollectedOn: date(t, "2026-09-02")}
	_, err := h.uc.Create(t.Context(), userA, create(incomeID, taxpayerA, in))
	require.NoError(t, err)
	assert.ElementsMatch(t, []regimeCall{{taxpayerA, "612", "2026-08"}, {taxpayerA, "612", "2026-09"}}, h.taxpayers.calls)
}

func TestCreate_GeneratesAnIDWhenAbsentAndCanonicalizesTheClients(t *testing.T) {
	h := newHarness(t)
	got, err := h.uc.Create(t.Context(), userA, create("", taxpayerA, us1(t)))
	require.NoError(t, err)
	assert.Equal(t, generated, got.Income.ID)

	other := us1(t)
	other.CFDIUUID = ""
	got, err = h.uc.Create(t.Context(), userA, create(strings.ToUpper(incomeID), strings.ToUpper(taxpayerA), other))
	require.NoError(t, err)
	assert.Equal(t, incomeID, got.Income.ID, "\"ABC…\" and \"abc…\" name the same income")
	assert.Equal(t, taxpayerA, got.Income.TaxpayerID)
}

func TestCreate_Idempotency(t *testing.T) {
	tests := []struct {
		name     string
		taxpayer string
		edit     func(*usecase.IncomeInput)
		deleted  bool
		want     apperr.Code // empty: 200 replay
	}{
		{name: "same payload is the stored income", taxpayer: taxpayerA},
		{name: "same payload after normalization", taxpayer: taxpayerA, edit: func(in *usecase.IncomeInput) {
			in.CFDIUUID = cfdiUpper
			in.Description = "  Diseño de logotipo  "
		}},
		{name: "another amount", taxpayer: taxpayerA, edit: func(in *usecase.IncomeInput) { in.Amount++ }, want: apperr.CodeConflict},
		{name: "another taxpayer", taxpayer: taxpayerA2, want: apperr.CodeConflict},
		{name: "a soft-deleted income", taxpayer: taxpayerA, deleted: true, want: apperr.CodeConflict},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHarness(t)
			first, err := h.uc.Create(t.Context(), userA, create(incomeID, taxpayerA, us1(t)))
			require.NoError(t, err)
			if tt.deleted {
				require.NoError(t, h.uc.Delete(t.Context(), userA, taxpayerA, incomeID))
			}
			// The retry arrives at the trial cap: it must never see a 402.
			h.access.limit = 1
			in := us1(t)
			if tt.edit != nil {
				tt.edit(&in)
			}

			got, err := h.uc.Create(t.Context(), userA, create(incomeID, tt.taxpayer, in))

			assert.Equal(t, 1, h.access.reserves, "a retry never takes a slot")
			assert.Equal(t, 1, h.store.inserts)
			if tt.want != "" {
				requireApp(t, err, tt.want, "id", "id_reused")
				require.ErrorIs(t, err, usecase.ErrIDReused)
				return
			}
			require.NoError(t, err)
			assert.False(t, got.Created, "200")
			assert.Equal(t, first.Income, got.Income)
			assert.Equal(t, first.Warnings, got.Warnings)
		})
	}
}

// A retry replays the original body, warnings included.
func TestCreate_ReplayKeepsTheWarnings(t *testing.T) {
	h := newHarness(t)
	in := us1(t)
	in.Withheld.ISR = 0 // differs from the 10 % suggestion
	first, err := h.uc.Create(t.Context(), userA, create(incomeID, taxpayerA, in))
	require.NoError(t, err)
	require.NotEmpty(t, first.Warnings)

	again, err := h.uc.Create(t.Context(), userA, create(incomeID, taxpayerA, in))
	require.NoError(t, err)
	assert.Equal(t, first.Warnings, again.Warnings)
}

// An id another user holds collides on the primary key behind RLS: one
// retry, then 409 without details (it reveals no ownership).
func TestCreate_ForeignIDIsAGenericConflict(t *testing.T) {
	h := newHarness(t)
	h.store.foreign[incomeID] = true
	_, err := h.uc.Create(t.Context(), userA, create(incomeID, taxpayerA, us1(t)))
	assert.Empty(t, requireApp(t, err, apperr.CodeConflict), "no details")
	require.ErrorIs(t, err, usecase.ErrIDTaken)
	assert.Equal(t, 2, len(h.tx.modes), "the unit of work is retried once")
}

// Two submits of one id race: the loser's insert collides, its retry sees
// the winner's committed row and answers the replay.
func TestCreate_ConcurrentDoubleSubmitReplays(t *testing.T) {
	h := newHarness(t)
	winner := create(incomeID, taxpayerA, us1(t))
	h.store.onInsert = func(s *fakeStore, in usecase.NewIncome) error {
		s.onInsert = nil
		s.rows[in.ID] = usecase.StoredIncome{
			ID: in.ID, TaxpayerID: in.TaxpayerID, Fields: in.Fields, BookedOn: bookedOn(in.Fields),
		}
		return usecase.ErrIDTaken
	}
	got, err := h.uc.Create(t.Context(), userA, winner)
	require.NoError(t, err)
	assert.False(t, got.Created)
}

func TestCreate_AccessOrder(t *testing.T) {
	t.Run("read-only is a 402 before the regime check", func(t *testing.T) {
		h := newHarness(t)
		h.access.readOnly = true
		_, err := h.uc.Create(t.Context(), userA, create(incomeID, taxpayerA, us1(t)))
		requireApp(t, err, apperr.CodeEntitlementRequired)
		assert.Empty(t, h.taxpayers.calls)
		assert.Zero(t, h.store.inserts)
	})
	t.Run("trial cap names the incomes cap", func(t *testing.T) {
		h := newHarness(t)
		h.access.limit, h.access.used = 5, 5
		_, err := h.uc.Create(t.Context(), userA, create(incomeID, taxpayerA, us1(t)))
		requireApp(t, err, apperr.CodeEntitlementRequired, "incomes_remaining", "trial_cap_reached")
		assert.Zero(t, h.store.inserts)
	})
	t.Run("a refused rule never takes a slot", func(t *testing.T) {
		h := newHarness(t)
		in := us1(t)
		in.Amount = 0
		_, err := h.uc.Create(t.Context(), userA, create(incomeID, taxpayerA, in))
		requireApp(t, err, apperr.CodeValidationFailed, "amount_cents", "must_be_positive")
		assert.Zero(t, h.access.reserves)
	})
}

func TestCreate_TaxpayerAndRegime(t *testing.T) {
	tests := []struct {
		name     string
		taxpayer string
		edit     func(*usecase.IncomeInput)
		code     apperr.Code
		field    []string
	}{
		{name: "malformed taxpayer id", taxpayer: "not-a-uuid", code: apperr.CodeNotFound},
		{name: "unknown taxpayer", taxpayer: "0199d4a2-1111-7000-8000-0000000000ff", code: apperr.CodeNotFound},
		{name: "archived taxpayer", taxpayer: archivedA, code: apperr.CodeBusinessRuleViolation, field: []string{"taxpayerId", "taxpayer_archived"}},
		{name: "regime never held", taxpayer: taxpayerA, edit: func(in *usecase.IncomeInput) { in.Regime = "626" },
			code: apperr.CodeBusinessRuleViolation, field: []string{"regime_code", "regime_not_active"}},
		{name: "regime not held on the date", taxpayer: taxpayerA, edit: func(in *usecase.IncomeInput) {
			in.Regime = "606"
			in.Withheld.ISR = 1_000_00
			in.Dates.CollectedOn = date(t, "2026-10-01")
		}, code: apperr.CodeBusinessRuleViolation, field: []string{"regime_code", "regime_not_active"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHarness(t)
			in := us1(t)
			if tt.edit != nil {
				tt.edit(&in)
			}
			_, err := h.uc.Create(t.Context(), userA, create(incomeID, tt.taxpayer, in))
			requireApp(t, err, tt.code, tt.field...)
			assert.Zero(t, h.access.reserves)
			assert.Zero(t, h.store.inserts)
		})
	}
}

func TestCreate_DomainRules(t *testing.T) {
	long := strings.Repeat("a", 201)
	tests := []struct {
		name  string
		edit  func(*usecase.IncomeInput)
		code  apperr.Code
		field string
		cause error
	}{
		{"amount 0", func(in *usecase.IncomeInput) { in.Amount = 0 }, apperr.CodeValidationFailed, "amount_cents", fiscal.ErrMustBePositive},
		{"612 without collected_on", func(in *usecase.IncomeInput) { in.Dates.CollectedOn = fiscal.Date{} },
			apperr.CodeValidationFailed, "collected_on", fiscal.ErrRequired},
		{"collected tomorrow", func(in *usecase.IncomeInput) { in.Dates.CollectedOn = date(t, "2026-10-14") },
			apperr.CodeValidationFailed, "collected_on", fiscal.ErrOutOfRange},
		{"IVA off the rate", func(in *usecase.IncomeInput) { in.IVA = 1_500_00 },
			apperr.CodeBusinessRuleViolation, "iva_cents", fiscal.ErrIVAAmountMismatch},
		{"withholding by a persona física", func(in *usecase.IncomeInput) { in.Counterparty = domain.CounterpartyFisica },
			apperr.CodeBusinessRuleViolation, "isr_withheld_cents", fiscal.ErrWithholdingNotApplicable},
		{"malformed CFDI UUID", func(in *usecase.IncomeInput) { in.CFDIUUID = "{" + cfdiLower + "}" },
			apperr.CodeValidationFailed, "cfdi_uuid", fiscal.ErrInvalid},
		{"description over 200 characters", func(in *usecase.IncomeInput) { in.Description = long },
			apperr.CodeValidationFailed, "description", usecase.ErrTooLong},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHarness(t)
			in := us1(t)
			tt.edit(&in)
			_, err := h.uc.Create(t.Context(), userA, create(incomeID, taxpayerA, in))
			requireApp(t, err, tt.code, tt.field, detailOf(tt.cause))
			require.ErrorIs(t, err, tt.cause)
			assert.Zero(t, h.store.inserts)
		})
	}
}

func detailOf(err error) string {
	if e, ok := errors.AsType[fiscal.Error](err); ok {
		return e.Detail()
	}
	return "too_long"
}

// A date whose fiscal year has no parameter set (nor a previous one to fall
// back on) is a typed error, never a misleading rate_not_valid_on_date.
func TestCreate_DateOutsideTheLoadedParameters(t *testing.T) {
	h := newHarness(t)
	in := us1(t)
	in.Dates = domain.Dates{CollectedOn: date(t, "2025-06-10")}
	_, err := h.uc.Create(t.Context(), userA, create(incomeID, taxpayerA, in))
	requireApp(t, err, apperr.CodeBusinessRuleViolation, "collected_on", "parameters_unavailable")
	require.ErrorIs(t, err, domain.ErrParametersUnavailable)
	assert.Zero(t, h.access.reserves)
}

// Validation (400) still comes first: an out-of-range date is out_of_range,
// whatever the parameters.
func TestCreate_ValidationBeforeParameters(t *testing.T) {
	h := newHarness(t)
	in := us1(t)
	in.Dates = domain.Dates{CollectedOn: date(t, "2025-06-10")}
	in.Amount = 0
	_, err := h.uc.Create(t.Context(), userA, create(incomeID, taxpayerA, in))
	requireApp(t, err, apperr.CodeValidationFailed, "amount_cents", "must_be_positive")
}

// PR #77 follow-up: a date that reads the provisional fallback set warns.
func TestCreate_ProvisionalParametersWarn(t *testing.T) {
	h := newHarness(t).at(time.Date(2027, time.February, 1, 12, 0, 0, 0, clock.Location()))
	in := us1(t)
	in.Dates = domain.Dates{CollectedOn: date(t, "2027-01-05")}
	got, err := h.uc.Create(t.Context(), userA, create(incomeID, taxpayerA, in))
	require.NoError(t, err)
	assert.Equal(t, []domain.Warning{{Code: domain.WarningProvisionalParameters, Field: "collected_on"}}, got.Warnings)
}

// The rate_8 border decree lapses on 2026-12-31 (US-I1).
func TestCreate_RateNotValidOnDate(t *testing.T) {
	h := newHarness(t).at(time.Date(2027, time.February, 1, 12, 0, 0, 0, clock.Location()))
	in := us1(t)
	in.IVARate, in.IVA = fiscal.IVARate8, 800_00
	in.Counterparty, in.Withheld = domain.CounterpartyPublic, fiscal.Withholdings{}
	in.Dates = domain.Dates{CollectedOn: date(t, "2027-01-05")}
	_, err := h.uc.Create(t.Context(), userA, create(incomeID, taxpayerA, in))
	requireApp(t, err, apperr.CodeBusinessRuleViolation, "iva_rate", "rate_not_valid_on_date")
}

func TestCreate_DuplicateCFDIUUIDWarns(t *testing.T) {
	h := newHarness(t)
	_, err := h.uc.Create(t.Context(), userA, create(incomeID, taxpayerA, us1(t)))
	require.NoError(t, err)
	second := us1(t)
	second.CFDIUUID = cfdiUpper
	second.Amount, second.IVA = 5_000_00, 800_00
	second.Withheld = fiscal.Withholdings{ISR: 500_00, IVA: 533_33}

	got, err := h.uc.Create(t.Context(), userA, create("", taxpayerA, second))
	require.NoError(t, err, "a partial collection is legit (OQ-9)")
	assert.Equal(t, []domain.Warning{{Code: domain.WarningDuplicateCFDIUUID, Field: "cfdi_uuid"}}, got.Warnings)

	_, err = h.uc.Create(t.Context(), userA, create("0199d4a3-0000-7000-8000-0000000000c3", taxpayerA2, second))
	require.NoError(t, err)
	got, err = h.uc.Create(t.Context(), userA, create("0199d4a3-0000-7000-8000-0000000000c4", taxpayerA2, func() usecase.IncomeInput {
		in := us1(t)
		in.CFDIUUID = ""
		return in
	}()))
	require.NoError(t, err)
	assert.Empty(t, got.Warnings, "no folio, no duplicate")
}

func TestCreate_Normalization(t *testing.T) {
	tests := []struct {
		name, description, want string
	}{
		{"trimmed", "  Diseño  ", "Diseño"},
		{"NFC", "Diseño", "Diseño"},
		{"blank is null", "   ", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHarness(t)
			in := us1(t)
			in.Description = tt.description
			got, err := h.uc.Create(t.Context(), userA, create(incomeID, taxpayerA, in))
			require.NoError(t, err)
			assert.Equal(t, tt.want, got.Income.Description)
		})
	}
}

func TestCreate_PlatformActivityDefaultsAndReplays(t *testing.T) {
	h := newHarness(t)
	in := usecase.IncomeInput{
		Regime: "625", Counterparty: domain.CounterpartyPlatform, Amount: 15_000_00,
		IVARate: fiscal.IVARate16, IVA: 2_400_00, Withheld: fiscal.Withholdings{ISR: 600_00, IVA: 1_200_00},
		Dates: domain.Dates{CollectedOn: date(t, "2026-09-20")},
	}
	got, err := h.uc.Create(t.Context(), userA, create(incomeID, platformA, in))
	require.NoError(t, err)
	assert.Equal(t, domain.ActivityLodging, got.Income.PlatformActivity, "the taxpayer's one activity")

	again, err := h.uc.Create(t.Context(), userA, create(incomeID, platformA, in))
	require.NoError(t, err)
	assert.False(t, again.Created, "the omitted default is the same payload")
}

func TestCreate_AccrualPendingCollection(t *testing.T) {
	h := newHarness(t)
	in := usecase.IncomeInput{
		Regime: "601", Counterparty: domain.CounterpartyMoral, Amount: 45_000_00,
		IVARate: fiscal.IVARate16, IVA: 7_200_00, Dates: domain.Dates{InvoicedOn: date(t, "2026-09-05")},
	}
	got, err := h.uc.Create(t.Context(), userA, create(incomeID, pmA, in))
	require.NoError(t, err)
	assert.True(t, got.Income.PendingCollection, "Por cobrar")
	assert.Equal(t, "2026-09-05", got.Income.BookedOn.String())
}

func TestCreate_ClientIDMustBeAUUID(t *testing.T) {
	h := newHarness(t)
	_, err := h.uc.Create(t.Context(), userA, create("123", taxpayerA, us1(t)))
	requireApp(t, err, apperr.CodeValidationFailed, "id", "invalid")
}

func TestCreate_AuditLogsIDsOnly(t *testing.T) {
	h := newHarness(t)
	_, err := h.uc.Create(t.Context(), userA, create(incomeID, taxpayerA, us1(t)))
	require.NoError(t, err)
	logs := h.logs.String()
	assert.Contains(t, logs, `"msg":"income.created"`)
	assert.Contains(t, logs, incomeID)
	assert.Contains(t, logs, taxpayerA)
	for _, secret := range []string{"Diseño", cfdiUpper, cfdiLower, "1000000", "160000"} {
		assert.NotContains(t, logs, secret, "S-5: never logged")
	}
}

func TestCreate_InfrastructureFailureIsNotAClientError(t *testing.T) {
	h := newHarness(t)
	h.store.fail = errors.New("connection reset")
	_, err := h.uc.Create(t.Context(), userA, create(incomeID, taxpayerA, us1(t)))
	require.Error(t, err)
	_, isApp := errors.AsType[*apperr.Error](err)
	assert.False(t, isApp, "a 500, not a contract error")
	assert.Contains(t, err.Error(), "create income")
}

// The write happened: a retry after the taxpayer was archived still replays
// the income (without the warnings it can no longer compute).
func TestCreate_ReplayAfterArchive(t *testing.T) {
	h := seeded(t)
	p := h.taxpayers.profiles[taxpayerA]
	p.Archived = true
	h.taxpayers.profiles[taxpayerA] = p

	got, err := h.uc.Create(t.Context(), userA, create(incomeID, taxpayerA, us1(t)))
	require.NoError(t, err)
	assert.False(t, got.Created)
	assert.Empty(t, got.Warnings)
}

// Without NewID or Logger the use cases generate uuid v7 ids and discard
// the audit log.
func TestNewIncomesDefaults(t *testing.T) {
	h := newHarness(t)
	uc := usecase.NewIncomes(usecase.Deps{
		Tx: h.tx, Store: h.store, Taxpayers: h.taxpayers, Access: h.access, Rates: h.catalog, Clock: clock.NewFixed(today),
	})
	got, err := uc.Create(t.Context(), userA, create("", taxpayerA, us1(t)))
	require.NoError(t, err)
	parsed, err := uuid.Parse(got.Income.ID)
	require.NoError(t, err)
	assert.Equal(t, uuid.Version(7), parsed.Version())
}

func TestCreate_NoUser(t *testing.T) {
	h := newHarness(t)
	_, err := h.uc.Create(context.Background(), "", create(incomeID, taxpayerA, us1(t)))
	require.Error(t, err)
}
