package usecase_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/platform/access"
	"github.com/romodevx/control-contable/backend/internal/platform/apperr"
	"github.com/romodevx/control-contable/backend/internal/taxpayers/domain"
	"github.com/romodevx/control-contable/backend/internal/taxpayers/usecase"
)

func fisicaInput(regimes ...usecase.RegimeInput) usecase.CreateInput {
	if len(regimes) == 0 {
		regimes = []usecase.RegimeInput{{Code: "612"}}
	}
	return usecase.CreateInput{PersonType: "fisica", RFC: rfcFisica, DisplayName: "Ana (personal)", Regimes: regimes}
}

func TestCreate_PersonaFisicaWithTwoRegimesAndDefaultedOptions(t *testing.T) {
	h := newHarness(t)
	in := fisicaInput(
		usecase.RegimeInput{Code: "606", Start: "2026-01"},
		usecase.RegimeInput{Code: "625", Start: "2026-01", Options: map[string]domain.OptionValue{
			"activities": domain.ListOption([]string{"goods_services"}),
		}},
	)

	got, err := h.taxpayers.Create(t.Context(), userA, in)

	require.NoError(t, err)
	assert.True(t, got.Created)
	tp := got.Taxpayer
	require.Len(t, h.ids, 1)
	assert.Equal(t, h.ids[0], tp.ID)
	assert.Equal(t, domain.PersonTypeFisica, tp.PersonType)
	assert.Equal(t, "Ana (personal)", tp.DisplayName)
	assert.Equal(t, rfcFisicaMasked, tp.RFCMasked)
	assert.Empty(t, tp.Warnings)
	assert.True(t, tp.ArchivedAt.IsZero())
	require.Len(t, tp.Regimes, 2)
	assert.Equal(t, "606", tp.Regimes[0].Code)
	assert.Equal(t, map[string]domain.OptionValue{"deduction_mode": domain.StringOption("blind_35")}, tp.Regimes[0].Options)
	assert.Equal(t, window(t, "2026-01", ""), tp.Regimes[0].Window)
	assert.Equal(t, "625", tp.Regimes[1].Code)
	assert.Equal(t, map[string]domain.OptionValue{
		"activities":            domain.ListOption([]string{"goods_services"}),
		"definitive_option":     domain.BoolOption(false),
		"rfc_given_to_platform": domain.BoolOption(true),
	}, tp.Regimes[1].Options)
}

func TestCreate_GeneratesTheRowIDBeforeSealingTheRFC(t *testing.T) {
	h := newHarness(t)

	got, err := h.taxpayers.Create(t.Context(), userA, fisicaInput())

	require.NoError(t, err)
	id := got.Taxpayer.ID
	require.Len(t, h.rfc.calls, 1)
	assert.Equal(t, usecase.RowBinding{TaxpayerID: id, UserID: userA}, h.rfc.calls[0].binding,
		"S-1: the ciphertext is bound to the row id that is inserted and to its owner")
	assert.Equal(t, []string{"id:" + id, "seal:" + id, "insert:" + id}, h.order)
	assert.Equal(t, []byte("ct:"+id+":"+userA), h.store.rows[id].rfc.Ciphertext)
}

func TestCreate_DefaultsTheStartToTheCurrentMexicoCityMonth(t *testing.T) {
	h := newHarness(t)

	got, err := h.taxpayers.Create(t.Context(), userA, fisicaInput())

	require.NoError(t, err)
	assert.Equal(t, window(t, "2026-10", ""), got.Taxpayer.Regimes[0].Window)
}

func TestCreate_PersonaMoral(t *testing.T) {
	h := newHarness(t)
	in := usecase.CreateInput{PersonType: "moral", RFC: rfcMoral, DisplayName: "Empresa", Regimes: []usecase.RegimeInput{{Code: "601"}}}

	got, err := h.taxpayers.Create(t.Context(), userA, in)

	require.NoError(t, err)
	assert.Equal(t, domain.PersonTypeMoral, got.Taxpayer.PersonType)
	assert.Equal(t, "A&B******XY2", got.Taxpayer.RFCMasked)
	assert.Equal(t, map[string]domain.OptionValue{"iva_fully_taxable": domain.BoolOption(true)}, got.Taxpayer.Regimes[0].Options)
}

func TestCreate_CheckDigitMismatchSucceedsWithAWarning(t *testing.T) {
	h := newHarness(t)
	in := fisicaInput()
	in.RFC = rfcFisicaMismatch

	got, err := h.taxpayers.Create(t.Context(), userA, in)

	require.NoError(t, err, "D-RFC: a check digit mismatch never rejects")
	require.Len(t, got.Taxpayer.Warnings, 1)
	w := got.Taxpayer.Warnings[0]
	assert.Equal(t, usecase.WarningRFCCheckDigitMismatch, w.Code)
	assert.Equal(t, "rfc", w.Field)
	assert.NotContains(t, w.Message, rfcFisicaMismatch)
	assert.False(t, h.store.rows[got.Taxpayer.ID].rfc.CheckDigitValid, "the verdict is stored")
}

func TestCreate_NormalizesTheTypedRFC(t *testing.T) {
	h := newHarness(t)
	in := fisicaInput()
	in.RFC = "  pexj-900215 hk8 "

	_, err := h.taxpayers.Create(t.Context(), userA, in)

	require.NoError(t, err)
	assert.Equal(t, rfcFisica, h.rfc.calls[0].rfc)
}

func TestCreate_RejectsBadRFCsNamingTheRule(t *testing.T) {
	tests := []struct {
		name       string
		personType string
		rfc        string
		detail     string
	}{
		{"moral RFC for a física", "fisica", rfcMoral, usecase.DetailRFCLengthMismatch},
		{"física RFC for a moral", "moral", rfcFisica, usecase.DetailRFCLengthMismatch},
		{"impossible date", "fisica", "ABCD991332XX1", usecase.DetailRFCDate},
		{"characters outside the alphabet", "fisica", "AB1D900215HK8", usecase.DetailRFCFormat},
		{"wrong length", "fisica", "PEXJ900215HK", usecase.DetailRFCFormat},
		{"público en general", "fisica", "XAXX010101000", usecase.DetailRFCGeneric},
		{"extranjero", "fisica", "XEXX010101000", usecase.DetailRFCGeneric},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHarness(t)
			in := fisicaInput()
			in.PersonType, in.RFC = tt.personType, tt.rfc
			if tt.personType == "moral" {
				in.Regimes = []usecase.RegimeInput{{Code: "601"}}
			}

			_, err := h.taxpayers.Create(t.Context(), userA, in)

			msg := requireAppErr(t, err, apperr.CodeValidationFailed, nil, [2]string{"rfc", tt.detail})
			assert.NotContains(t, msg+err.Error(), tt.rfc, "the RFC never appears in an error")
			assert.Empty(t, h.store.rows)
		})
	}
}

func TestCreate_RejectsAnInvalidDisplayNameAndPersonType(t *testing.T) {
	h := newHarness(t)
	in := fisicaInput()
	in.DisplayName = "   "
	_, err := h.taxpayers.Create(t.Context(), userA, in)
	requireAppErr(t, err, apperr.CodeValidationFailed, domain.ErrInvalidDisplayName, [2]string{"display_name", usecase.DetailInvalid})

	in = fisicaInput()
	in.PersonType = "extranjero"
	_, err = h.taxpayers.Create(t.Context(), userA, in)
	requireAppErr(t, err, apperr.CodeValidationFailed, domain.ErrInvalidPersonType, [2]string{"person_type", usecase.DetailInvalid})
}

func TestCreate_ReadOnlyUserGets402AndNothingIsWritten(t *testing.T) {
	h := newHarness(t)
	h.access.readOnly = true

	_, err := h.taxpayers.Create(t.Context(), userA, fisicaInput())

	requireAppErr(t, err, apperr.CodeEntitlementRequired, access.ErrReadOnly)
	assert.Empty(t, h.store.rows)
	assert.Empty(t, h.rfc.calls)
}

func TestCreate_ChecksTheAccessPolicyBeforeTheCap(t *testing.T) {
	h := newHarness(t)
	h.access.readOnly = true
	for i := range access.TaxpayerCap {
		h.store.seedTaxpayer(userA, fmt.Sprintf("0199d4a2-0000-7000-8000-1000000000%02d", i), domain.PersonTypeFisica, false)
	}

	_, err := h.taxpayers.Create(t.Context(), userA, fisicaInput())

	requireAppErr(t, err, apperr.CodeEntitlementRequired, access.ErrReadOnly)
}

func TestCreate_TheEleventhTaxpayerIsRefusedArchivedIncluded(t *testing.T) {
	h := newHarness(t)
	for i := range access.TaxpayerCap {
		h.store.seedTaxpayer(userA, fmt.Sprintf("0199d4a2-0000-7000-8000-1000000000%02d", i), domain.PersonTypeFisica, i%2 == 0)
	}

	_, err := h.taxpayers.Create(t.Context(), userA, fisicaInput())

	requireAppErr(t, err, apperr.CodeTaxpayerLimitReached, access.ErrTaxpayerLimitReached)
	assert.Contains(t, h.store.locks, "user:update", "the count runs under the user lock (S-5)")
	assert.Len(t, h.store.rows, access.TaxpayerCap)
}

func TestCreate_CapCountsOnlyTheCallersTaxpayers(t *testing.T) {
	h := newHarness(t)
	for i := range access.TaxpayerCap {
		h.store.seedTaxpayer(userB, fmt.Sprintf("0199d4a2-0000-7000-8000-2000000000%02d", i), domain.PersonTypeFisica, false)
	}

	_, err := h.taxpayers.Create(t.Context(), userA, fisicaInput())

	require.NoError(t, err)
}

func TestCreate_IdempotentRetryReturnsTheExistingTaxpayer(t *testing.T) {
	const clientID = "0199d4a2-8f00-7c44-a1b2-3c4d5e6f7a80"
	tests := []struct {
		name  string
		setup func(h *harness)
	}{
		{"plain retry", func(*harness) {}},
		{"read-only user", func(h *harness) { h.access.readOnly = true }},
		{"at the cap", func(h *harness) {
			for i := range access.TaxpayerCap - 1 {
				h.store.seedTaxpayer(userA, fmt.Sprintf("0199d4a2-0000-7000-8000-3000000000%02d", i), domain.PersonTypeFisica, false)
			}
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHarness(t)
			in := fisicaInput()
			in.ID = clientID
			first, err := h.taxpayers.Create(t.Context(), userA, in)
			require.NoError(t, err)
			require.True(t, first.Created)
			tt.setup(h)
			in.ID = strings.ToUpper(clientID) // same id, other spelling

			again, err := h.taxpayers.Create(t.Context(), userA, in)

			require.NoError(t, err)
			assert.False(t, again.Created)
			assert.Equal(t, clientID, again.Taxpayer.ID)
			assert.Empty(t, again.Taxpayer.Warnings)
			assert.Len(t, h.rfc.calls, 1, "a retry writes nothing")
		})
	}
}

func TestCreate_UsesTheCanonicalClientID(t *testing.T) {
	h := newHarness(t)
	in := fisicaInput()
	in.ID = "0199D4A2-8F00-7C44-A1B2-3C4D5E6F7A80"

	got, err := h.taxpayers.Create(t.Context(), userA, in)

	require.NoError(t, err)
	assert.Equal(t, "0199d4a2-8f00-7c44-a1b2-3c4d5e6f7a80", got.Taxpayer.ID)
	assert.Empty(t, h.ids, "a client id is never replaced")
	assert.Equal(t, got.Taxpayer.ID, h.rfc.calls[0].binding.TaxpayerID)
}

func TestCreate_RejectsAClientIDThatIsNotAUUID(t *testing.T) {
	h := newHarness(t)
	in := fisicaInput()
	in.ID = "not-a-uuid"

	_, err := h.taxpayers.Create(t.Context(), userA, in)

	requireAppErr(t, err, apperr.CodeValidationFailed, usecase.ErrInvalidID, [2]string{"id", usecase.DetailInvalid})
}

func TestCreate_AClientIDOfAnotherUserIsAGenericConflict(t *testing.T) {
	const clientID = "0199d4a2-8f00-7c44-a1b2-3c4d5e6f7a80"
	h := newHarness(t)
	h.store.seedTaxpayer(userB, clientID, domain.PersonTypeFisica, false)
	in := fisicaInput()
	in.ID = clientID

	_, err := h.taxpayers.Create(t.Context(), userA, in)

	requireAppErr(t, err, apperr.CodeConflict, usecase.ErrIDConflict)
}

func TestCreate_DuplicateRFCForTheSameUser(t *testing.T) {
	h := newHarness(t)
	_, err := h.taxpayers.Create(t.Context(), userA, fisicaInput())
	require.NoError(t, err)
	in := fisicaInput()
	in.RFC = strings.ToLower(rfcFisica) // another spelling, same RFC

	_, err = h.taxpayers.Create(t.Context(), userA, in)

	requireAppErr(t, err, apperr.CodeRFCAlreadyRegistered, usecase.ErrRFCAlreadyRegistered)
}

func TestCreate_SameRFCForAnotherUserIsAllowed(t *testing.T) {
	h := newHarness(t)
	_, err := h.taxpayers.Create(t.Context(), userB, fisicaInput())
	require.NoError(t, err)

	_, err = h.taxpayers.Create(t.Context(), userA, fisicaInput())

	require.NoError(t, err)
}

func TestCreate_IncompatibleRegimesAre422NamingTheLaterOne(t *testing.T) {
	h := newHarness(t)

	_, err := h.taxpayers.Create(t.Context(), userA, fisicaInput(usecase.RegimeInput{Code: "612"}, usecase.RegimeInput{Code: "626"}))

	msg := requireAppErr(t, err, apperr.CodeInvalidRegimeCombination, usecase.ErrIncompatibleRegimes,
		[2]string{"regimes.1.regime_code", usecase.DetailIncompatible})
	assert.Equal(t, "El RESICO sustituye al 612.", msg, "the catalog reason is shown")
	assert.Empty(t, h.store.rows)
}

func TestCreate_MoralHoldsOneRegimeAtATime(t *testing.T) {
	h := newHarness(t)
	in := usecase.CreateInput{PersonType: "moral", RFC: rfcMoral, DisplayName: "Empresa",
		Regimes: []usecase.RegimeInput{{Code: "601"}, {Code: "626"}}}

	_, err := h.taxpayers.Create(t.Context(), userA, in)

	requireAppErr(t, err, apperr.CodeInvalidRegimeCombination, usecase.ErrIncompatibleRegimes,
		[2]string{"regimes.1.regime_code", usecase.DetailIncompatible})
}

func TestCreate_NonOverlappingIncompatibleRegimesAreAllowed(t *testing.T) {
	h := newHarness(t)
	// Both open-ended: a create cannot end a regime, so 113-E exits are added later.
	_, err := h.taxpayers.Create(t.Context(), userA, fisicaInput(
		usecase.RegimeInput{Code: "612", Start: "2026-07"}, usecase.RegimeInput{Code: "606", Start: "2026-01"}))

	require.NoError(t, err)
}

func TestCreate_APairUnderCPAReviewIsAllowedWithAWarning(t *testing.T) {
	h := newHarness(t)

	got, err := h.taxpayers.Create(t.Context(), userA, fisicaInput(usecase.RegimeInput{Code: "605"}, usecase.RegimeInput{Code: "607"}))

	require.NoError(t, err)
	require.Equal(t, []usecase.Warning{{
		Code: usecase.WarningRegimeNeedsCPAReview, Field: "regimes.1.regime_code", Message: "Pendiente de revisión por un contador.",
	}}, got.Taxpayer.Warnings)
	id := got.Taxpayer.ID
	assert.False(t, h.store.regimes[id+"/605"].ack)
	assert.True(t, h.store.regimes[id+"/607"].ack, "the regime that answered the warning records it")
}

func TestCreate_ValidatesEveryRegimeAndReportsEveryField(t *testing.T) {
	h := newHarness(t)
	in := fisicaInput(
		usecase.RegimeInput{Code: "601"}, // moral only
		usecase.RegimeInput{Code: "621"}, // rif_year required, no default
		usecase.RegimeInput{Code: "606", Options: map[string]domain.OptionValue{
			"deduction_mode": domain.StringOption("half"), "colour": domain.StringOption("red"),
		}},
		usecase.RegimeInput{Code: "621"},                   // duplicate
		usecase.RegimeInput{Code: "612", Start: "2025-12"}, // before the first parameter year
		usecase.RegimeInput{Code: "605", Start: "2026-11"}, // after the current month
		usecase.RegimeInput{Code: "607", Start: "2026-1x"}, // not a period
	)

	_, err := h.taxpayers.Create(t.Context(), userA, in)

	requireAppErr(t, err, apperr.CodeValidationFailed, nil,
		[2]string{"regimes.0.regime_code", usecase.DetailNotForPersonType},
		[2]string{"regimes.1.options.rif_year", usecase.DetailRequired},
		[2]string{"regimes.2.options.colour", usecase.DetailUnknownOption},
		[2]string{"regimes.2.options.deduction_mode", usecase.DetailInvalidOption},
		[2]string{"regimes.3.regime_code", usecase.DetailDuplicate},
		[2]string{"regimes.4.start_period", usecase.DetailOutOfRange},
		[2]string{"regimes.5.start_period", usecase.DetailOutOfRange},
		[2]string{"regimes.6.start_period", usecase.DetailInvalid},
	)
	assert.Empty(t, h.store.rows)
}

func TestCreate_RequiresAtLeastOneRegime(t *testing.T) {
	h := newHarness(t)
	in := fisicaInput()
	in.Regimes = nil

	_, err := h.taxpayers.Create(t.Context(), userA, in)

	requireAppErr(t, err, apperr.CodeValidationFailed, nil, [2]string{"regimes", usecase.DetailRequired})
}

func TestCreate_TheCatalogsRaceBackstopIsAnInvalidCombination(t *testing.T) {
	h := newHarness(t)
	h.store.failNext = fmt.Errorf("infra: %w", usecase.ErrIncompatibleRegimes)

	_, err := h.taxpayers.Create(t.Context(), userA, fisicaInput())

	requireAppErr(t, err, apperr.CodeInvalidRegimeCombination, usecase.ErrIncompatibleRegimes)
}

func TestCreate_AuditLogNamesOnlyTheIDsAndNeverTheRFC(t *testing.T) {
	h := newHarness(t)
	in := fisicaInput()
	in.RFC = rfcFisicaMismatch

	got, err := h.taxpayers.Create(t.Context(), userA, in)
	require.NoError(t, err)
	_, err = h.taxpayers.Create(t.Context(), userA, in) // duplicate: an error path too
	require.Error(t, err)

	logs := h.logs.String()
	assert.Contains(t, logs, `"msg":"taxpayer.created"`)
	assert.Contains(t, logs, `"user_id":"`+userA+`"`)
	assert.Contains(t, logs, `"taxpayer_id":"`+got.Taxpayer.ID+`"`)
	for _, secret := range []string{rfcFisicaMismatch, strings.ToLower(rfcFisicaMismatch), "PEXJ900215"} {
		assert.NotContains(t, logs, secret, "S-4: the RFC is never logged")
	}
}

func TestList_HidesArchivedUnlessAskedAndLoadsRegimesInOneQuery(t *testing.T) {
	h := newHarness(t)
	h.store.seedTaxpayer(userA, "0199d4a2-0000-7000-8000-0000000000a1", domain.PersonTypeFisica, false,
		seedRegime{code: "612", window: window(t, "2026-01", "")})
	h.store.seedTaxpayer(userA, "0199d4a2-0000-7000-8000-0000000000a2", domain.PersonTypeFisica, true,
		seedRegime{code: "606", window: window(t, "2026-01", "")})
	h.store.seedTaxpayer(userA, "0199d4a2-0000-7000-8000-0000000000a3", domain.PersonTypeMoral, false,
		seedRegime{code: "601", window: window(t, "2026-01", "")})
	h.store.seedTaxpayer(userB, "0199d4a2-0000-7000-8000-0000000000b1", domain.PersonTypeFisica, false)

	active, err := h.taxpayers.List(t.Context(), userA, false)
	require.NoError(t, err)
	all, err := h.taxpayers.List(t.Context(), userA, true)
	require.NoError(t, err)

	require.Len(t, active, 2)
	assert.Equal(t, "0199d4a2-0000-7000-8000-0000000000a1", active[0].ID)
	assert.Equal(t, "612", active[0].Regimes[0].Code)
	assert.Equal(t, "0199d4a2-0000-7000-8000-0000000000a3", active[1].ID)
	assert.Equal(t, "601", active[1].Regimes[0].Code)
	require.Len(t, all, 3)
	assert.Equal(t, 2, h.store.regRuns, "one regimes query per list, never one per taxpayer")
	assert.Equal(t, []string{"ro", "ro"}, h.tx.modes)
}

func TestGet_ReadsArchivedToo_AndHidesOtherUsers(t *testing.T) {
	h := newHarness(t)
	const mine, theirs = "0199d4a2-0000-7000-8000-0000000000a1", "0199d4a2-0000-7000-8000-0000000000b1"
	h.store.seedTaxpayer(userA, mine, domain.PersonTypeFisica, true, seedRegime{code: "612", window: window(t, "2026-01", "")})
	h.store.seedTaxpayer(userB, theirs, domain.PersonTypeFisica, false)

	got, err := h.taxpayers.Get(t.Context(), userA, strings.ToUpper(mine))
	require.NoError(t, err)
	assert.Equal(t, mine, got.ID)
	assert.False(t, got.ArchivedAt.IsZero())
	assert.Len(t, got.Regimes, 1)

	for _, id := range []string{theirs, "0199d4a2-0000-7000-8000-0000000000ff", "not-a-uuid"} {
		_, err = h.taxpayers.Get(t.Context(), userA, id)
		requireAppErr(t, err, apperr.CodeNotFound, nil)
	}
}

func TestUpdate_RenamesTrimmedAndNFC(t *testing.T) {
	h := newHarness(t)
	const id = "0199d4a2-0000-7000-8000-0000000000a1"
	h.store.seedTaxpayer(userA, id, domain.PersonTypeFisica, false)
	name := "  José (negocio) "

	got, err := h.taxpayers.Update(t.Context(), userA, id, usecase.UpdateInput{DisplayName: &name})

	require.NoError(t, err)
	assert.Equal(t, "José (negocio)", got.DisplayName)
	assert.Empty(t, got.Warnings)
	assert.Contains(t, h.store.locks, fmt.Sprintf("taxpayer:%d", usecase.LockUpdate))
}

func TestUpdate_CorrectsTheRFCResealedForTheSameRow(t *testing.T) {
	h := newHarness(t)
	const id = "0199d4a2-0000-7000-8000-0000000000a1"
	h.store.seedTaxpayer(userA, id, domain.PersonTypeFisica, false)
	rfc := "aaaa-010101 aa0"

	got, err := h.taxpayers.Update(t.Context(), userA, id, usecase.UpdateInput{RFC: &rfc})

	require.NoError(t, err)
	assert.Equal(t, "AAAA******AA0", got.RFCMasked)
	require.Len(t, h.rfc.calls, 1)
	assert.Equal(t, usecase.RowBinding{TaxpayerID: id, UserID: userA}, h.rfc.calls[0].binding)
	assert.Equal(t, rfcFisicaOther, h.rfc.calls[0].rfc)
	assert.Equal(t, []byte("ct:"+id+":"+userA), h.store.rows[id].rfc.Ciphertext)
}

func TestUpdate_RFCCorrectionIsAuditedWithIDsOnly(t *testing.T) {
	h := newHarness(t)
	const id = "0199d4a2-0000-7000-8000-0000000000a1"
	h.store.seedTaxpayer(userA, id, domain.PersonTypeFisica, false)
	name, rfc := "Renamed", rfcFisicaOther

	_, err := h.taxpayers.Update(t.Context(), userA, id, usecase.UpdateInput{DisplayName: &name})
	require.NoError(t, err)
	assert.NotContains(t, h.logs.String(), "taxpayer.rfc_corrected", "a rename alone is not an RFC correction")

	_, err = h.taxpayers.Update(t.Context(), userA, id, usecase.UpdateInput{RFC: &rfc})
	require.NoError(t, err)

	logs := h.logs.String()
	assert.Equal(t, 1, strings.Count(logs, `"msg":"taxpayer.rfc_corrected"`))
	assert.Contains(t, logs, `"user_id":"`+userA+`"`)
	assert.Contains(t, logs, `"taxpayer_id":"`+id+`"`)
	assert.NotContains(t, logs, rfcFisicaOther, "the RFC is never logged (S-9)")
	assert.NotContains(t, logs, "AAAA******AA0", "not even masked")
}

func TestUpdate_RFCCorrectionWithAMismatchedCheckDigitWarns(t *testing.T) {
	h := newHarness(t)
	const id = "0199d4a2-0000-7000-8000-0000000000a1"
	h.store.seedTaxpayer(userA, id, domain.PersonTypeFisica, false)
	rfc := rfcFisicaMismatch

	got, err := h.taxpayers.Update(t.Context(), userA, id, usecase.UpdateInput{RFC: &rfc})

	require.NoError(t, err)
	require.Len(t, got.Warnings, 1)
	assert.Equal(t, usecase.WarningRFCCheckDigitMismatch, got.Warnings[0].Code)
}

func TestUpdate_Refusals(t *testing.T) {
	const id, archived, theirs = "0199d4a2-0000-7000-8000-0000000000a1", "0199d4a2-0000-7000-8000-0000000000a2", "0199d4a2-0000-7000-8000-0000000000b1"
	name, empty, moralRFC, dupRFC := "Otro", " ", rfcMoral, rfcFisicaOther
	tests := []struct {
		name     string
		readOnly bool
		id       string
		in       usecase.UpdateInput
		code     apperr.Code
		target   error
		fields   [][2]string
	}{
		{"read-only user", true, id, usecase.UpdateInput{DisplayName: &name}, apperr.CodeEntitlementRequired, access.ErrReadOnly, nil},
		{"archived taxpayer", false, archived, usecase.UpdateInput{DisplayName: &name}, apperr.CodeBusinessRuleViolation, usecase.ErrTaxpayerArchived,
			[][2]string{{"taxpayerId", usecase.DetailTaxpayerArchived}}},
		{"another user's taxpayer", false, theirs, usecase.UpdateInput{DisplayName: &name}, apperr.CodeNotFound, usecase.ErrTaxpayerNotFound, nil},
		{"blank name", false, id, usecase.UpdateInput{DisplayName: &empty}, apperr.CodeValidationFailed, domain.ErrInvalidDisplayName,
			[][2]string{{"display_name", usecase.DetailInvalid}}},
		{"RFC of the other person type", false, id, usecase.UpdateInput{RFC: &moralRFC}, apperr.CodeValidationFailed, domain.ErrPersonTypeMismatch,
			[][2]string{{"rfc", usecase.DetailRFCLengthMismatch}}},
		{"RFC already registered", false, id, usecase.UpdateInput{RFC: &dupRFC}, apperr.CodeRFCAlreadyRegistered, usecase.ErrRFCAlreadyRegistered, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHarness(t)
			h.store.seedTaxpayer(userA, id, domain.PersonTypeFisica, false)
			h.store.seedTaxpayer(userA, archived, domain.PersonTypeFisica, true)
			h.store.seedTaxpayer(userB, theirs, domain.PersonTypeFisica, false)
			_, err := h.taxpayers.Create(context.Background(), userA, usecase.CreateInput{
				PersonType: "fisica", RFC: rfcFisicaOther, DisplayName: "Dup", Regimes: []usecase.RegimeInput{{Code: "612"}},
			})
			require.NoError(t, err)
			h.access.readOnly = tt.readOnly

			_, err = h.taxpayers.Update(t.Context(), userA, tt.id, tt.in)

			fields := tt.fields
			requireAppErr(t, err, tt.code, tt.target, fields...)
			assert.Equal(t, "Seed a1", h.store.rows[id].t.DisplayName, "nothing changed")
		})
	}
}

func TestArchive_IsAllowedReadOnlyAndIdempotent(t *testing.T) {
	h := newHarness(t)
	const id = "0199d4a2-0000-7000-8000-0000000000a1"
	h.store.seedTaxpayer(userA, id, domain.PersonTypeFisica, false)
	h.access.readOnly = true

	got, err := h.taxpayers.Archive(t.Context(), userA, id)
	require.NoError(t, err)
	assert.True(t, got.ArchivedAt.Equal(today))

	again, err := h.taxpayers.Archive(t.Context(), userA, id)
	require.NoError(t, err)
	assert.True(t, again.ArchivedAt.Equal(today), "archiving an archived taxpayer changes nothing")
	assert.Zero(t, h.access.checks, "archive never asks the access policy")
	assert.Equal(t, 1, strings.Count(h.logs.String(), `"msg":"taxpayer.archived"`))

	_, err = h.taxpayers.Archive(t.Context(), userB, id)
	requireAppErr(t, err, apperr.CodeNotFound, usecase.ErrTaxpayerNotFound)
}

func TestUnarchive(t *testing.T) {
	h := newHarness(t)
	const id = "0199d4a2-0000-7000-8000-0000000000a1"
	h.store.seedTaxpayer(userA, id, domain.PersonTypeFisica, true)

	h.access.readOnly = true
	_, err := h.taxpayers.Unarchive(t.Context(), userA, id)
	requireAppErr(t, err, apperr.CodeEntitlementRequired, access.ErrReadOnly)
	assert.True(t, h.store.rows[id].t.Archived())

	h.access.readOnly = false
	got, err := h.taxpayers.Unarchive(t.Context(), userA, id)
	require.NoError(t, err)
	assert.True(t, got.ArchivedAt.IsZero())
	assert.Contains(t, h.logs.String(), `"msg":"taxpayer.unarchived"`)

	again, err := h.taxpayers.Unarchive(t.Context(), userA, id)
	require.NoError(t, err, "unarchiving an active taxpayer answers it unchanged")
	assert.True(t, again.ArchivedAt.IsZero())
}

func TestUnarchive_AtTheCapSucceedsBecauseArchivedAlreadyCounts(t *testing.T) {
	h := newHarness(t)
	for i := range access.TaxpayerCap {
		h.store.seedTaxpayer(userA, fmt.Sprintf("0199d4a2-0000-7000-8000-1000000000%02d", i), domain.PersonTypeFisica, i == 0)
	}

	_, err := h.taxpayers.Unarchive(t.Context(), userA, "0199d4a2-0000-7000-8000-100000000000")

	require.NoError(t, err)
}

func TestDelete(t *testing.T) {
	h := newHarness(t)
	const id = "0199d4a2-0000-7000-8000-0000000000a1"
	h.store.seedTaxpayer(userA, id, domain.PersonTypeFisica, true, seedRegime{code: "612", window: window(t, "2026-01", "")})
	h.access.readOnly = true // delete is allowed read-only

	err := h.taxpayers.Delete(t.Context(), userA, id, false)
	requireAppErr(t, err, apperr.CodeValidationFailed, usecase.ErrConfirmRequired, [2]string{"confirm", usecase.DetailConfirmRequired})
	require.Contains(t, h.store.rows, id)

	err = h.taxpayers.Delete(t.Context(), userB, id, true)
	requireAppErr(t, err, apperr.CodeNotFound, usecase.ErrTaxpayerNotFound)

	require.NoError(t, h.taxpayers.Delete(t.Context(), userA, id, true))
	assert.NotContains(t, h.store.rows, id)
	assert.Empty(t, h.store.regimes, "regimes cascade")
	assert.Contains(t, h.logs.String(), `"msg":"taxpayer.deleted"`)

	err = h.taxpayers.Delete(t.Context(), userA, id, true)
	requireAppErr(t, err, apperr.CodeNotFound, usecase.ErrTaxpayerNotFound)
}
