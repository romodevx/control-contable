//go:build integration

package infra_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/platform/access"
	"github.com/romodevx/control-contable/backend/internal/platform/apperr"
	"github.com/romodevx/control-contable/backend/internal/platform/crypto"
	"github.com/romodevx/control-contable/backend/internal/taxpayers/domain"
	"github.com/romodevx/control-contable/backend/internal/taxpayers/usecase"
)

func fisica(rfc string, regimes ...usecase.RegimeInput) usecase.CreateInput {
	if len(regimes) == 0 {
		regimes = []usecase.RegimeInput{{Code: "612", Start: "2026-01"}}
	}
	return usecase.CreateInput{PersonType: "fisica", RFC: rfc, DisplayName: "Ana (personal)", Regimes: regimes}
}

// rfcFor returns a distinct, well-formed física RFC per n (1..28); its check
// digit may not match, which D-RFC accepts.
func rfcFor(n int) string { return fmt.Sprintf("PEXJ9002%02dHK8", n) }

func requireCode(t *testing.T, err error, code apperr.Code) {
	t.Helper()
	require.Error(t, err)
	app, ok := errors.AsType[*apperr.Error](err)
	require.True(t, ok, "want *apperr.Error, got %T: %v", err, err)
	require.Equal(t, code, app.Code, "error: %v", err)
}

func TestCreate_PersistsTheTaxpayerRegimesAndOptions(t *testing.T) {
	e := newEnv(t, userA).wire(t)
	in := fisica(rfcFisica,
		usecase.RegimeInput{Code: "606", Start: "2026-01"},
		usecase.RegimeInput{Code: "625", Start: "2026-02", Options: map[string]domain.OptionValue{
			"activities": domain.ListOption([]string{"goods_services", "lodging"}),
		}})

	created, err := e.taxpayers.Create(t.Context(), userA, in)
	require.NoError(t, err)
	got, err := e.taxpayers.Get(t.Context(), userA, created.Taxpayer.ID)
	require.NoError(t, err)

	assert.Equal(t, created.Taxpayer, got)
	assert.Equal(t, "PEXJ******HK8", got.RFCMasked)
	assert.False(t, got.CreatedAt.IsZero())
	require.Len(t, got.Regimes, 2)
	assert.Equal(t, "Arrendamiento", got.Regimes[0].Name)
	assert.Equal(t, "full", got.Regimes[0].Treatment)
	assert.Equal(t, domain.StringOption("blind_35"), got.Regimes[0].Options["deduction_mode"])
	assert.Equal(t, domain.BoolOption(false), got.Regimes[0].Options["iva_border_region"], "every option, defaults included")
	assert.Equal(t, domain.ListOption([]string{"lodging", "goods_services"}), got.Regimes[1].Options["activities"], "catalog order")
	assert.Equal(t, "2026-02", got.Regimes[1].Window.Start.String())

	var stored string
	queryAs(t, e.app, userA, "SELECT person_type || '|' || rfc_check_digit_valid::text || '|' || rfc_key_id::text FROM taxpayers WHERE id = $1",
		[]any{got.ID}, &stored)
	assert.Equal(t, "fisica|true|1", stored)
}

func TestCreate_StoresTheRFCSealedForItsRowAndOwner(t *testing.T) {
	e := newEnv(t, userA, userB).wire(t)
	a1, err := e.taxpayers.Create(t.Context(), userA, fisica(rfcFisica))
	require.NoError(t, err)
	a2, err := e.taxpayers.Create(t.Context(), userA, fisica(rfcFisicaOther))
	require.NoError(t, err)
	b1, err := e.taxpayers.Create(t.Context(), userB, fisica(rfcFisica))
	require.NoError(t, err)
	ctA1 := ciphertext(t, e, userA, a1.Taxpayer.ID)

	rfc, err := e.vault.Open(ctA1, usecase.RowBinding{TaxpayerID: a1.Taxpayer.ID, UserID: userA})
	require.NoError(t, err)
	assert.Equal(t, rfcFisica, rfc.String())
	assert.NotContains(t, string(ctA1), rfcFisica, "never stored in clear")

	// Same user: copy A1's ciphertext onto A2's row (cc_app may write the column).
	require.NoError(t, execAs(t.Context(), e.app, userA,
		"UPDATE taxpayers SET rfc_encrypted = (SELECT rfc_encrypted FROM taxpayers WHERE id = $1) WHERE id = $2",
		a1.Taxpayer.ID, a2.Taxpayer.ID))
	_, err = e.vault.Open(ciphertext(t, e, userA, a2.Taxpayer.ID), usecase.RowBinding{TaxpayerID: a2.Taxpayer.ID, UserID: userA})
	require.ErrorIs(t, err, crypto.ErrDecrypt, "S-1: a ciphertext moved to another row fails")

	// Another user: neither their row nor A1's id under their user opens it.
	_, err = e.vault.Open(ctA1, usecase.RowBinding{TaxpayerID: b1.Taxpayer.ID, UserID: userB})
	require.ErrorIs(t, err, crypto.ErrDecrypt)
	_, err = e.vault.Open(ctA1, usecase.RowBinding{TaxpayerID: a1.Taxpayer.ID, UserID: userB})
	require.ErrorIs(t, err, crypto.ErrDecrypt, "the owner is part of the binding")
}

func TestCreate_BlindIndexIsPerUser(t *testing.T) {
	e := newEnv(t, userA, userB).wire(t)
	a, err := e.taxpayers.Create(t.Context(), userA, fisica(rfcFisica))
	require.NoError(t, err)
	b, err := e.taxpayers.Create(t.Context(), userB, fisica(rfcFisica))
	require.NoError(t, err, "uniqueness is per user (US-T2)")

	var hashA, hashB []byte
	queryAs(t, e.app, userA, "SELECT rfc_hash FROM taxpayers WHERE id = $1", []any{a.Taxpayer.ID}, &hashA)
	queryAs(t, e.app, userB, "SELECT rfc_hash FROM taxpayers WHERE id = $1", []any{b.Taxpayer.ID}, &hashB)
	assert.NotEqual(t, hashA, hashB, "S-2: the same RFC of two users is unlinkable")

	_, err = e.taxpayers.Create(t.Context(), userA, fisica(" pexj-900215 hk8"))
	requireCode(t, err, apperr.CodeRFCAlreadyRegistered)
}

func TestCreate_FiveConcurrentCreatesAtNineAdmitExactlyOne(t *testing.T) {
	e := newEnv(t, userA).wire(t)
	for n := 1; n <= access.TaxpayerCap-1; n++ {
		_, err := e.taxpayers.Create(t.Context(), userA, fisica(rfcFor(n)))
		require.NoError(t, err)
	}
	const racers = 5
	// A blocker holds the user lock until every racer is waiting on it, so the
	// five creates really overlap: without the lock all five would count 9.
	blocker, err := e.app.BeginTx(t.Context(), nil)
	require.NoError(t, err)
	_, err = blocker.ExecContext(t.Context(), "SELECT set_config('app.current_user_id', $1, true)", userA)
	require.NoError(t, err)
	_, err = blocker.ExecContext(t.Context(), "SELECT taxpayer_count_for_create()")
	require.NoError(t, err)
	errs := make([]error, racers)
	var wg sync.WaitGroup
	for i := range racers {
		wg.Go(func() {
			_, errs[i] = e.taxpayers.Create(context.Background(), userA, fisica(rfcFor(20+i)))
		})
	}
	waitForLockWaiters(t, e, racers)
	require.NoError(t, blocker.Commit())
	wg.Wait()

	won, capped := 0, 0
	for _, err := range errs {
		if err == nil {
			won++
			continue
		}
		requireCode(t, err, apperr.CodeTaxpayerLimitReached)
		capped++
	}
	assert.Equal(t, 1, won)
	assert.Equal(t, racers-1, capped)
	all, err := e.taxpayers.List(t.Context(), userA, true)
	require.NoError(t, err)
	assert.Len(t, all, access.TaxpayerCap)
}

// waitForLockWaiters polls until n sessions of this database wait on a lock,
// or a few seconds pass (a broken lock never makes them wait; the caller's
// assertions then fail).
func waitForLockWaiters(t *testing.T, e *env, n int) {
	t.Helper()
	const query = "SELECT count(*) FROM pg_stat_activity WHERE datname = current_database() AND wait_event_type = 'Lock'"
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		var waiting int
		require.NoError(t, e.app.QueryRowContext(t.Context(), query).Scan(&waiting))
		if waiting >= n {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func TestCreate_IdempotentRetryAndForeignClientID(t *testing.T) {
	const clientID = "0199d4a2-8f00-7c44-a1b2-3c4d5e6f7a80"
	e := newEnv(t, userA, userB).wire(t)
	in := fisica(rfcFisica)
	in.ID = clientID

	first, err := e.taxpayers.Create(t.Context(), userA, in)
	require.NoError(t, err)
	assert.True(t, first.Created)
	again, err := e.taxpayers.Create(t.Context(), userA, in)
	require.NoError(t, err)
	assert.False(t, again.Created)
	assert.Equal(t, first.Taxpayer.ID, again.Taxpayer.ID)

	in.RFC = rfcFisicaOther
	_, err = e.taxpayers.Create(t.Context(), userB, in)
	requireCode(t, err, apperr.CodeConflict) // RLS hides the row; the primary key still collides
}

func TestCreate_AFailureMidTransactionPersistsNothing(t *testing.T) {
	e := newEnv(t, userA)
	e.wire(t, func(d *usecase.Deps) { d.Regimes = failingSecondInsert{RegimeStore: e.store} })

	_, err := e.taxpayers.Create(t.Context(), userA, fisica(rfcFisica,
		usecase.RegimeInput{Code: "606", Start: "2026-01"}, usecase.RegimeInput{Code: "612", Start: "2026-01"}))
	require.Error(t, err)

	var taxpayers, regimes int
	queryAs(t, e.app, userA, "SELECT (SELECT count(*) FROM taxpayers), (SELECT count(*) FROM taxpayer_regimes)", nil, &taxpayers, &regimes)
	assert.Zero(t, taxpayers, "no taxpayer without its regimes (US-T1)")
	assert.Zero(t, regimes)
}

// failingSecondInsert fails the second regime insert of a create.
type failingSecondInsert struct {
	usecase.RegimeStore
}

var errInjected = errors.New("injected failure")

func (f failingSecondInsert) InsertRegime(ctx context.Context, r usecase.NewRegime) error {
	if r.Code == "612" {
		return errInjected
	}
	return f.RegimeStore.InsertRegime(ctx, r)
}

func TestCreate_ReadOnlyUserGets402(t *testing.T) {
	e := newEnv(t, userA)
	e.wire(t, readOnlyAccess(e.txm))

	_, err := e.taxpayers.Create(t.Context(), userA, fisica(rfcFisica))

	requireCode(t, err, apperr.CodeEntitlementRequired)
}

func TestCreate_IncompatibleBodyRegimesAre422(t *testing.T) {
	e := newEnv(t, userA).wire(t)
	moral := usecase.CreateInput{PersonType: "moral", RFC: rfcMoral, DisplayName: "Empresa",
		Regimes: []usecase.RegimeInput{{Code: "601", Start: "2026-01"}, {Code: "626", Start: "2026-01"}}}

	_, err := e.taxpayers.Create(t.Context(), userA, moral)

	requireCode(t, err, apperr.CodeInvalidRegimeCombination)
}

func TestCreate_NeverLogsTheRFC(t *testing.T) {
	e := newEnv(t, userA).wire(t)
	for _, rfc := range []string{rfcFisicaMismatch, rfcFisicaMismatch, "ABCD991332XX1"} { // ok, duplicate, bad date
		_, _ = e.taxpayers.Create(t.Context(), userA, fisica(rfc))
	}

	logs := e.logs.String()
	assert.Contains(t, logs, `"msg":"taxpayer.created"`)
	for _, secret := range []string{rfcFisicaMismatch, "PEXJ900215", "ABCD991332"} {
		assert.NotContains(t, logs, secret, "S-4")
	}
}

func TestEveryOperationOnAnotherUsersTaxpayerIs404(t *testing.T) {
	e := newEnv(t, userA, userB).wire(t)
	created, err := e.taxpayers.Create(t.Context(), userA, fisica(rfcFisica, usecase.RegimeInput{Code: "606", Start: "2026-01"},
		usecase.RegimeInput{Code: "612", Start: "2026-01"}))
	require.NoError(t, err)
	id := created.Taxpayer.ID
	name, rfc := "Mine now", rfcFisicaOther

	calls := map[string]func() error{
		"get": func() error { _, err := e.taxpayers.Get(t.Context(), userB, id); return err },
		"update": func() error {
			_, err := e.taxpayers.Update(t.Context(), userB, id, usecase.UpdateInput{DisplayName: &name, RFC: &rfc})
			return err
		},
		"archive":   func() error { _, err := e.taxpayers.Archive(t.Context(), userB, id); return err },
		"unarchive": func() error { _, err := e.taxpayers.Unarchive(t.Context(), userB, id); return err },
		"delete":    func() error { return e.taxpayers.Delete(t.Context(), userB, id, true) },
		"add regime": func() error {
			_, err := e.regimes.Add(t.Context(), userB, id, usecase.RegimeInput{Code: "605"})
			return err
		},
		"update options": func() error {
			_, err := e.regimes.UpdateOptions(t.Context(), userB, id, "606", nil)
			return err
		},
		"end regime": func() error {
			_, err := e.regimes.End(t.Context(), userB, id, usecase.EndRegimeInput{Code: "606", End: "2026-09"})
			return err
		},
		"check regime": func() error { return e.queries.CheckRegime(t.Context(), userB, id, "606", "2026-09") },
		"profile":      func() error { _, err := e.queries.Profile(t.Context(), userB, id); return err },
	}
	for name, call := range calls {
		t.Run(name, func(t *testing.T) { requireCode(t, call(), apperr.CodeNotFound) })
	}
	got, err := e.taxpayers.Get(t.Context(), userA, id)
	require.NoError(t, err)
	assert.Equal(t, created.Taxpayer, got, "the owner's taxpayer is untouched")
	list, err := e.taxpayers.List(t.Context(), userB, true)
	require.NoError(t, err)
	assert.Empty(t, list)
}

func TestUpdate_RenamesAndCorrectsTheRFC(t *testing.T) {
	e := newEnv(t, userA).wire(t)
	created, err := e.taxpayers.Create(t.Context(), userA, fisica(rfcFisica))
	require.NoError(t, err)
	id := created.Taxpayer.ID
	var hashBefore []byte
	queryAs(t, e.app, userA, "SELECT rfc_hash FROM taxpayers WHERE id = $1", []any{id}, &hashBefore)
	name, rfc := " Ana (negocio) ", rfcFisicaMismatch

	got, err := e.taxpayers.Update(t.Context(), userA, id, usecase.UpdateInput{DisplayName: &name, RFC: &rfc})

	require.NoError(t, err)
	assert.Equal(t, "Ana (negocio)", got.DisplayName)
	assert.Equal(t, "PEXJ******HK9", got.RFCMasked)
	require.Len(t, got.Warnings, 1)
	opened, err := e.vault.Open(ciphertext(t, e, userA, id), usecase.RowBinding{TaxpayerID: id, UserID: userA})
	require.NoError(t, err, "re-encrypted with the same row binding")
	assert.Equal(t, rfcFisicaMismatch, opened.String())
	var hashAfter []byte
	var digitValid bool
	queryAs(t, e.app, userA, "SELECT rfc_hash, rfc_check_digit_valid FROM taxpayers WHERE id = $1", []any{id}, &hashAfter, &digitValid)
	assert.NotEqual(t, hashBefore, hashAfter, "re-hashed")
	assert.False(t, digitValid)
}

func TestUpdate_ADuplicateRFCIs409(t *testing.T) {
	e := newEnv(t, userA).wire(t)
	_, err := e.taxpayers.Create(t.Context(), userA, fisica(rfcFisica))
	require.NoError(t, err)
	other, err := e.taxpayers.Create(t.Context(), userA, fisica(rfcFisicaOther))
	require.NoError(t, err)
	rfc := rfcFisica

	_, err = e.taxpayers.Update(t.Context(), userA, other.Taxpayer.ID, usecase.UpdateInput{RFC: &rfc})

	requireCode(t, err, apperr.CodeRFCAlreadyRegistered)
}

func TestArchiveUnarchiveListAndDelete(t *testing.T) {
	e := newEnv(t, userA).wire(t)
	a, err := e.taxpayers.Create(t.Context(), userA, fisica(rfcFisica))
	require.NoError(t, err)
	b, err := e.taxpayers.Create(t.Context(), userA, fisica(rfcFisicaOther))
	require.NoError(t, err)

	archived, err := e.taxpayers.Archive(t.Context(), userA, a.Taxpayer.ID)
	require.NoError(t, err)
	assert.True(t, archived.ArchivedAt.Equal(today))
	active, err := e.taxpayers.List(t.Context(), userA, false)
	require.NoError(t, err)
	require.Len(t, active, 1)
	assert.Equal(t, b.Taxpayer.ID, active[0].ID)
	all, err := e.taxpayers.List(t.Context(), userA, true)
	require.NoError(t, err)
	assert.Len(t, all, 2)
	name := "x"
	_, err = e.taxpayers.Update(t.Context(), userA, a.Taxpayer.ID, usecase.UpdateInput{DisplayName: &name})
	requireCode(t, err, apperr.CodeBusinessRuleViolation)

	unarchived, err := e.taxpayers.Unarchive(t.Context(), userA, a.Taxpayer.ID)
	require.NoError(t, err)
	assert.True(t, unarchived.ArchivedAt.IsZero())

	require.NoError(t, e.taxpayers.Delete(t.Context(), userA, a.Taxpayer.ID, true))
	var regimes int
	queryAs(t, e.app, userA, "SELECT count(*) FROM taxpayer_regimes WHERE taxpayer_id = $1", []any{a.Taxpayer.ID}, &regimes)
	assert.Zero(t, regimes, "regimes cascade")
	_, err = e.taxpayers.Create(t.Context(), userA, fisica(rfcFisica))
	require.NoError(t, err, "deleting frees the RFC")
}

func ciphertext(t *testing.T, e *env, userID, id string) []byte {
	t.Helper()
	var ct []byte
	queryAs(t, e.app, userID, "SELECT rfc_encrypted FROM taxpayers WHERE id = $1", []any{id}, &ct)
	require.NotEmpty(t, ct)
	return ct
}

func TestList_Order(t *testing.T) {
	e := newEnv(t, userA).wire(t)
	var ids []string
	for n := 1; n <= 3; n++ {
		created, err := e.taxpayers.Create(t.Context(), userA, fisica(rfcFor(n)))
		require.NoError(t, err)
		ids = append(ids, created.Taxpayer.ID)
	}

	all, err := e.taxpayers.List(t.Context(), userA, false)

	require.NoError(t, err)
	got := make([]string, 0, len(all))
	for _, tp := range all {
		got = append(got, tp.ID)
		assert.Len(t, tp.Regimes, 1)
	}
	assert.Equal(t, ids, got, "creation order (uuidv7 ids within one created_at)")
	assert.True(t, strings.HasPrefix(all[0].RFCMasked, "PEXJ"))
}
