//go:build integration

package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/adapter/http/contracttest"
	"github.com/romodevx/control-contable/backend/adapter/http/gen"
	"github.com/romodevx/control-contable/backend/internal/platform/flags/flagtest"
)

// Contract tests of the 11 module_taxpayers operations (M2-07) through the
// real composition root over Postgres as cc_app, with the flag ON and OFF.
// module_auth stays ON in both runs: it is how the users sign up.

// Synthetic RFCs (internal/taxpayers/domain/rfc_test.go): valid check
// digits, except rfcMismatch.
const (
	rfcAna      = "PEXJ900215HK8"
	rfcMismatch = "PEXJ900215HK9"
	rfcOther    = "AAAA010101AA0"
	rfcMoral    = "A&B020304XY2"

	clientTaxpayerID = "0199d4a2-8f00-7c44-a1b2-3c4d5e6f7a80"
	unknownTaxpayer  = "0199d4a2-8f00-7c44-a1b2-3c4d5e6f7aff"
)

// taxpayerOps is one request per module_taxpayers operation, for the OFF
// sweep; the token is a placeholder (the gate answers first).
var taxpayerOps = []struct{ method, target, body string }{
	{http.MethodGet, "/v1/regimes?person_type=fisica", ""},
	{http.MethodGet, "/v1/taxpayers", ""},
	{http.MethodPost, "/v1/taxpayers", `{}`},
	{http.MethodGet, "/v1/taxpayers/" + clientTaxpayerID, ""},
	{http.MethodPatch, "/v1/taxpayers/" + clientTaxpayerID, `{"display_name":"x"}`},
	{http.MethodDelete, "/v1/taxpayers/" + clientTaxpayerID + "?confirm=true", ""},
	{http.MethodPost, "/v1/taxpayers/" + clientTaxpayerID + "/archive", ""},
	{http.MethodPost, "/v1/taxpayers/" + clientTaxpayerID + "/unarchive", ""},
	{http.MethodPost, "/v1/taxpayers/" + clientTaxpayerID + "/regimes", `{"regime_code":"612"}`},
	{http.MethodPatch, "/v1/taxpayers/" + clientTaxpayerID + "/regimes/612", `{"options":{}}`},
	{http.MethodDelete, "/v1/taxpayers/" + clientTaxpayerID + "/regimes/612?end_period=2026-09", ""},
}

func createJSON(id, personType, rfc, name string, regimes ...string) string {
	idField := ""
	if id != "" {
		idField = fmt.Sprintf(`"id":%q,`, id)
	}
	return fmt.Sprintf(`{%s"person_type":%q,"rfc":%q,"display_name":%q,"regimes":[%s]}`,
		idField, personType, rfc, name, strings.Join(regimes, ","))
}

func regimeJSON(code string) string { return fmt.Sprintf(`{"regime_code":%q}`, code) }

func decodeTaxpayer(t *testing.T, rec *httptest.ResponseRecorder) gen.Taxpayer {
	t.Helper()
	var tp gen.Taxpayer
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &tp), rec.Body.String())
	return tp
}

func decodeError(t *testing.T, rec *httptest.ResponseRecorder) gen.ErrorResponse {
	t.Helper()
	var e gen.ErrorResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &e), rec.Body.String())
	return e
}

// details returns the "field/code" pairs of an error envelope.
func details(t *testing.T, rec *httptest.ResponseRecorder) []string {
	t.Helper()
	e := decodeError(t, rec)
	var out []string
	if e.Error.Details != nil {
		for _, d := range *e.Error.Details {
			out = append(out, d.Field+"/"+d.Code)
		}
	}
	return out
}

func warningCodes(w *[]gen.Warning) []string {
	if w == nil {
		return nil
	}
	var out []string
	for _, x := range *w {
		out = append(out, x.Code+"@"+x.Field)
	}
	return out
}

// makeReadOnly ends userID's trial in the past, as an expired trial looks.
// entitlements is FORCE RLS and its owner policy covers comp rows only, so
// the fixture lifts FORCE for one statement in the owner's transaction and
// restores it before commit.
func (e contractEnv) makeReadOnly(t *testing.T, userID string) {
	t.Helper()
	tx, err := e.owner.BeginTx(t.Context(), nil)
	require.NoError(t, err)
	defer func() { _ = tx.Rollback() }()
	_, err = tx.ExecContext(t.Context(), "ALTER TABLE entitlements NO FORCE ROW LEVEL SECURITY")
	require.NoError(t, err)
	res, err := tx.ExecContext(t.Context(), `UPDATE entitlements
SET starts_at = now() - interval '200 hours', ends_at = now() - interval '100 hours'
WHERE user_id = $1 AND kind = 'trial'`, userID)
	require.NoError(t, err)
	n, err := res.RowsAffected()
	require.NoError(t, err)
	require.EqualValues(t, 1, n, "the trial row was updated")
	_, err = tx.ExecContext(t.Context(), "ALTER TABLE entitlements FORCE ROW LEVEL SECURITY")
	require.NoError(t, err)
	require.NoError(t, tx.Commit())
	var forced bool
	require.NoError(t, e.owner.QueryRowContext(t.Context(),
		"SELECT relforcerowsecurity FROM pg_class WHERE oid = 'entitlements'::regclass").Scan(&forced))
	require.True(t, forced, "FORCE RLS restored")
}

func TestContract_taxpayers(t *testing.T) {
	c := contracttest.Load(t)

	flagtest.OnOff(t, "module_taxpayers", func(t *testing.T, f flagtest.Static, on bool) {
		env := newContractEnv(t, pingOK{})
		env.apply(t, []string{"module_auth", "module_taxpayers"}, f)

		if !on {
			for _, op := range taxpayerOps {
				flagtest.AssertHidden(t, c.Serve(t, env.handler, env.send(op.method, op.target, op.body, "any-token")))
			}
			return
		}

		ana := env.signUp(t, signUpEmail, acceptableSample).Tokens.AccessToken
		bea := env.signUp(t, beaEmail, otherSample).Tokens.AccessToken
		serve := func(op contracttest.Operation, method, target, body, token string) *httptest.ResponseRecorder {
			t.Helper()
			return c.ServeOp(t, op, env.handler, env.send(method, target, body, token))
		}

		t.Run("listRegimes: the catalog with incompatible and CPA-review pairs", func(t *testing.T) {
			rec := serve(contracttest.Op("listRegimes"), http.MethodGet, "/v1/regimes?person_type=fisica", "", ana)
			require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
			var list gen.RegimeList
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &list))
			byCode := map[string]gen.Regime{}
			for _, r := range list.Items {
				byCode[r.Code] = r
			}
			require.Contains(t, byCode, "626")
			resico := byCode["626"]
			assert.Contains(t, codesOf(resico.IncompatibleWith), "612")
			require.NotNil(t, resico.ReviewPendingWith)
			assert.Contains(t, codesOf(*resico.ReviewPendingWith), "607", "PR #56 follow-up: the client can warn before submitting")

			rec = serve(contracttest.Op("listRegimes"), http.MethodGet, "/v1/regimes?person_type=moral", "", ana)
			require.Equal(t, http.StatusOK, rec.Code)

			rec = serve(contracttest.Op("listRegimes"), http.MethodGet, "/v1/regimes", "", ana)
			assert.Equal(t, http.StatusBadRequest, rec.Code)
			assert.Contains(t, details(t, rec), "person_type/required")
			rec = serve(contracttest.Op("listRegimes"), http.MethodGet, "/v1/regimes?person_type=juridica", "", ana)
			assert.Equal(t, http.StatusBadRequest, rec.Code)

			rec = serve(contracttest.Op("listRegimes"), http.MethodGet, "/v1/regimes?person_type=fisica", "", "")
			assert.Equal(t, http.StatusUnauthorized, rec.Code)
		})

		var anaID string
		t.Run("createTaxpayer: 201 with the check-digit warning, then an idempotent 200", func(t *testing.T) {
			body := createJSON(clientTaxpayerID, "fisica", rfcMismatch, "Ana (personal)", regimeJSON("612"))
			rec := serve(contracttest.Op("createTaxpayer"), http.MethodPost, "/v1/taxpayers", body, ana)
			require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())
			assert.Equal(t, "no-store", rec.Header().Get("Cache-Control"))
			created := decodeTaxpayer(t, rec)
			anaID = created.Id.String()
			assert.Equal(t, clientTaxpayerID, anaID)
			assert.Equal(t, "PEXJ******HK9", created.RfcMasked)
			assert.Equal(t, []string{"rfc_check_digit_mismatch@rfc"}, warningCodes(created.Warnings))
			assert.NotContains(t, rec.Body.String(), rfcMismatch, "the full RFC never leaves the server")
			assert.NotContains(t, env.logs.String(), rfcMismatch, "S-4: the RFC is never logged")
			assert.Contains(t, env.logs.String(), "taxpayer.created", "S-9: the audit event, ids only")

			rec = serve(contracttest.Op("createTaxpayer"), http.MethodPost, "/v1/taxpayers", body, ana)
			require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
			assert.Nil(t, decodeTaxpayer(t, rec).Warnings)
		})

		t.Run("createTaxpayer: refusals", func(t *testing.T) {
			rec := serve(contracttest.Op("createTaxpayer"), http.MethodPost, "/v1/taxpayers",
				createJSON("", "fisica", "pexj 900215-hk9", "Otra", regimeJSON("626")), ana)
			assert.Equal(t, http.StatusConflict, rec.Code, "the same RFC, typed differently")
			assert.Equal(t, gen.ErrorCodeRfcAlreadyRegistered, decodeError(t, rec).Error.Code)

			rec = serve(contracttest.Op("createTaxpayer"), http.MethodPost, "/v1/taxpayers",
				createJSON("", "fisica", rfcMoral, "Mal", regimeJSON("601")), ana)
			assert.Equal(t, http.StatusBadRequest, rec.Code)
			assert.ElementsMatch(t, []string{"rfc/rfc_length_mismatch", "regimes.0.regime_code/not_for_person_type"}, details(t, rec),
				"dotted fields, every failure at once")
			assert.NotContains(t, rec.Body.String(), rfcMoral)

			rec = serve(contracttest.Op("createTaxpayer"), http.MethodPost, "/v1/taxpayers",
				createJSON("", "fisica", rfcOther, "Mal", regimeJSON("612"), regimeJSON("626")), ana)
			assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
			assert.Equal(t, gen.ErrorCodeInvalidRegimeCombination, decodeError(t, rec).Error.Code)
			assert.Equal(t, []string{"regimes.1.regime_code/incompatible"}, details(t, rec))

			rec = serve(contracttest.Op("createTaxpayer"), http.MethodPost, "/v1/taxpayers",
				`{"person_type":"fisica","rfc":"`+rfcOther+`","display_name":"x","regimes":[],"extra":1}`, ana)
			assert.Equal(t, http.StatusBadRequest, rec.Code, "closed bodies (S-11)")

			rec = serve(contracttest.Op("createTaxpayer"), http.MethodPost, "/v1/taxpayers", createJSON("", "fisica", rfcOther, "x", regimeJSON("612")), "")
			assert.Equal(t, http.StatusUnauthorized, rec.Code)
		})

		t.Run("createTaxpayer: a pair under CPA review is allowed with a warning", func(t *testing.T) {
			rec := serve(contracttest.Op("createTaxpayer"), http.MethodPost, "/v1/taxpayers",
				createJSON("", "fisica", rfcOther, "Ana (renta)", regimeJSON("626"), regimeJSON("607")), ana)
			require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())
			assert.Equal(t, []string{"regime_combination_needs_cpa_review@regimes.1.regime_code"}, warningCodes(decodeTaxpayer(t, rec).Warnings))
		})

		t.Run("getTaxpayer and listTaxpayers: own data only (404 across users)", func(t *testing.T) {
			rec := serve(contracttest.Op("getTaxpayer"), http.MethodGet, "/v1/taxpayers/"+anaID, "", ana)
			require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
			assert.Nil(t, decodeTaxpayer(t, rec).Warnings, "never on reads")

			rec = serve(contracttest.Op("getTaxpayer"), http.MethodGet, "/v1/taxpayers/"+anaID, "", bea)
			assert.Equal(t, http.StatusNotFound, rec.Code, "another user's taxpayer is 404, never 403")
			rec = serve(contracttest.Op("getTaxpayer"), http.MethodGet, "/v1/taxpayers/"+unknownTaxpayer, "", ana)
			assert.Equal(t, http.StatusNotFound, rec.Code)
			rec = serve(contracttest.Op("getTaxpayer"), http.MethodGet, "/v1/taxpayers/not-a-uuid", "", ana)
			assert.Equal(t, http.StatusBadRequest, rec.Code)

			rec = serve(contracttest.Op("listTaxpayers"), http.MethodGet, "/v1/taxpayers", "", ana)
			require.Equal(t, http.StatusOK, rec.Code)
			var list gen.TaxpayerList
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &list))
			assert.Len(t, list.Items, 2)
			rec = serve(contracttest.Op("listTaxpayers"), http.MethodGet, "/v1/taxpayers", "", bea)
			require.Equal(t, http.StatusOK, rec.Code)
			assert.JSONEq(t, `{"items":[]}`, rec.Body.String())
		})

		t.Run("updateTaxpayer: rename, RFC correction, duplicate, cross-user", func(t *testing.T) {
			rec := serve(contracttest.Op("updateTaxpayer"), http.MethodPatch, "/v1/taxpayers/"+anaID, `{"display_name":"Ana (negocio)","rfc":"`+rfcAna+`"}`, ana)
			require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
			updated := decodeTaxpayer(t, rec)
			assert.Equal(t, "Ana (negocio)", updated.DisplayName)
			assert.Equal(t, "PEXJ******HK8", updated.RfcMasked)
			assert.Nil(t, updated.Warnings, "the corrected RFC has a valid check digit")

			rec = serve(contracttest.Op("updateTaxpayer"), http.MethodPatch, "/v1/taxpayers/"+anaID, `{"rfc":"`+rfcOther+`"}`, ana)
			assert.Equal(t, http.StatusConflict, rec.Code)
			assert.Equal(t, gen.ErrorCodeRfcAlreadyRegistered, decodeError(t, rec).Error.Code)

			rec = serve(contracttest.Op("updateTaxpayer"), http.MethodPatch, "/v1/taxpayers/"+anaID, `{"rfc":"`+rfcMoral+`"}`, ana)
			assert.Equal(t, http.StatusBadRequest, rec.Code)
			assert.Equal(t, []string{"rfc/rfc_length_mismatch"}, details(t, rec))

			rec = serve(contracttest.Op("updateTaxpayer"), http.MethodPatch, "/v1/taxpayers/"+anaID, `{"display_name":"Robo"}`, bea)
			assert.Equal(t, http.StatusNotFound, rec.Code)
		})

		t.Run("regimes: add, update options, end", func(t *testing.T) {
			base := "/v1/taxpayers/" + anaID + "/regimes"
			rec := serve(contracttest.Op("addTaxpayerRegime"), http.MethodPost, base, `{"regime_code":"606","options":{"deduction_mode":"actual"}}`, ana)
			require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())
			assert.Equal(t, "no-store", rec.Header().Get("Cache-Control"))

			rec = serve(contracttest.Op("addTaxpayerRegime"), http.MethodPost, base, `{"regime_code":"606"}`, ana)
			assert.Equal(t, http.StatusConflict, rec.Code, "already active")
			rec = serve(contracttest.Op("addTaxpayerRegime"), http.MethodPost, base, `{"regime_code":"626"}`, ana)
			assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
			assert.Equal(t, gen.ErrorCodeInvalidRegimeCombination, decodeError(t, rec).Error.Code)
			rec = serve(contracttest.Op("addTaxpayerRegime"), http.MethodPost, base, `{"regime_code":"621","options":{"rif_year":"2"}}`, ana)
			require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())
			var added gen.TaxpayerRegime
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &added))
			assert.Contains(t, warningCodes(added.Warnings), "regime_combination_needs_cpa_review@regime_code")
			rec = serve(contracttest.Op("addTaxpayerRegime"), http.MethodPost, base, `{"regime_code":"606"}`, bea)
			assert.Equal(t, http.StatusNotFound, rec.Code)

			rec = serve(contracttest.Op("updateTaxpayerRegime"), http.MethodPatch, base+"/606", `{"options":{"deduction_mode":"blind_35"}}`, ana)
			require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
			rec = serve(contracttest.Op("updateTaxpayerRegime"), http.MethodPatch, base+"/606", `{"options":{"deduction_mode":"nope"}}`, ana)
			assert.Equal(t, http.StatusBadRequest, rec.Code)
			assert.Equal(t, []string{"options.deduction_mode/invalid_option"}, details(t, rec))
			rec = serve(contracttest.Op("updateTaxpayerRegime"), http.MethodPatch, base+"/601", `{"options":{}}`, ana)
			assert.Equal(t, http.StatusNotFound, rec.Code, "a regime the taxpayer does not hold")

			rec = serve(contracttest.Op("archiveTaxpayerRegime"), http.MethodDelete, base+"/621?end_period=2026-10", "", ana)
			require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
			var ended gen.TaxpayerRegime
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &ended))
			require.NotNil(t, ended.EndPeriod)
			assert.Equal(t, "2026-10", *ended.EndPeriod)
			rec = serve(contracttest.Op("archiveTaxpayerRegime"), http.MethodDelete, base+"/621?end_period=2026-10", "", ana)
			assert.Equal(t, http.StatusConflict, rec.Code, "already ended")
			rec = serve(contracttest.Op("archiveTaxpayerRegime"), http.MethodDelete, base+"/606?end_period=2026-13", "", ana)
			assert.Equal(t, http.StatusBadRequest, rec.Code)
		})

		t.Run("archive, unarchive and the archived 422", func(t *testing.T) {
			rec := serve(contracttest.Op("archiveTaxpayer"), http.MethodPost, "/v1/taxpayers/"+anaID+"/archive", "", ana)
			require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
			assert.NotNil(t, decodeTaxpayer(t, rec).ArchivedAt)

			rec = serve(contracttest.Op("updateTaxpayer"), http.MethodPatch, "/v1/taxpayers/"+anaID, `{"display_name":"x"}`, ana)
			assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
			assert.Equal(t, []string{"taxpayerId/taxpayer_archived"}, details(t, rec))

			rec = serve(contracttest.Op("listTaxpayers"), http.MethodGet, "/v1/taxpayers", "", ana)
			var list gen.TaxpayerList
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &list))
			assert.Len(t, list.Items, 1, "archived ones are hidden by default")
			rec = serve(contracttest.Op("listTaxpayers"), http.MethodGet, "/v1/taxpayers?include_archived=true", "", ana)
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &list))
			assert.Len(t, list.Items, 2)

			rec = serve(contracttest.Op("unarchiveTaxpayer"), http.MethodPost, "/v1/taxpayers/"+anaID+"/unarchive", "", ana)
			require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
			assert.Nil(t, decodeTaxpayer(t, rec).ArchivedAt)
			rec = serve(contracttest.Op("unarchiveTaxpayer"), http.MethodPost, "/v1/taxpayers/"+anaID+"/unarchive", "", bea)
			assert.Equal(t, http.StatusNotFound, rec.Code)
			rec = serve(contracttest.Op("archiveTaxpayer"), http.MethodPost, "/v1/taxpayers/"+anaID+"/archive", "", bea)
			assert.Equal(t, http.StatusNotFound, rec.Code)
		})

		t.Run("ending the last active regime is 422", func(t *testing.T) {
			rec := serve(contracttest.Op("createTaxpayer"), http.MethodPost, "/v1/taxpayers",
				createJSON("", "moral", rfcMoral, "Empresa", regimeJSON("601")), bea)
			require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())
			id := decodeTaxpayer(t, rec).Id.String()
			rec = serve(contracttest.Op("archiveTaxpayerRegime"), http.MethodDelete, "/v1/taxpayers/"+id+"/regimes/601?end_period=2026-10", "", bea)
			assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
			assert.Equal(t, []string{"regimeCode/last_active_regime"}, details(t, rec))
		})

		t.Run("deleteTaxpayer: confirm required, then 204 and gone", func(t *testing.T) {
			rec := serve(contracttest.Op("deleteTaxpayer"), http.MethodDelete, "/v1/taxpayers/"+anaID, "", ana)
			assert.Equal(t, http.StatusBadRequest, rec.Code)
			assert.Equal(t, []string{"confirm/confirm_required"}, details(t, rec))
			rec = serve(contracttest.Op("deleteTaxpayer"), http.MethodDelete, "/v1/taxpayers/"+anaID+"?confirm=true", "", bea)
			assert.Equal(t, http.StatusNotFound, rec.Code)

			rec = serve(contracttest.Op("deleteTaxpayer"), http.MethodDelete, "/v1/taxpayers/"+anaID+"?confirm=true", "", ana)
			assert.Equal(t, http.StatusNoContent, rec.Code, rec.Body.String())
			rec = serve(contracttest.Op("getTaxpayer"), http.MethodGet, "/v1/taxpayers/"+anaID, "", ana)
			assert.Equal(t, http.StatusNotFound, rec.Code)
		})
	}, "module_auth")
}

func codesOf(in []gen.RegimeIncompatibility) []string {
	out := make([]string, 0, len(in))
	for _, x := range in {
		out = append(out, x.Code)
	}
	return out
}

// The cap (10 taxpayers, archived included) is 422 taxpayer_limit_reached;
// a read-only user (trial over) gets 402 on creates but may still list,
// archive and delete.
func TestContract_taxpayersCapAndReadOnly(t *testing.T) {
	c := contracttest.Load(t)
	env := newContractEnv(t, pingOK{})
	env.apply(t, []string{"module_auth", "module_taxpayers"}, flagtest.Static{"module_auth": true, "module_taxpayers": true})
	session := env.signUp(t, signUpEmail, acceptableSample)
	token := session.Tokens.AccessToken

	var first string
	for i := range 10 {
		rfc := fmt.Sprintf("AAAA0101%02dAA0", i+1)
		rec := c.ServeOp(t, contracttest.Op("createTaxpayer"), env.handler, env.send(http.MethodPost, "/v1/taxpayers",
			createJSON("", "fisica", rfc, fmt.Sprintf("Contribuyente %d", i+1), regimeJSON("612")), token))
		require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())
		if i == 0 {
			first = decodeTaxpayer(t, rec).Id.String()
		}
	}
	rec := c.ServeOp(t, contracttest.Op("createTaxpayer"), env.handler, env.send(http.MethodPost, "/v1/taxpayers",
		createJSON("", "fisica", "AAAA010111AA0", "Once", regimeJSON("612")), token))
	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
	assert.Equal(t, gen.ErrorCodeTaxpayerLimitReached, decodeError(t, rec).Error.Code)

	env.makeReadOnly(t, session.User.Id.String())

	rec = c.ServeOp(t, contracttest.Op("updateTaxpayer"), env.handler,
		env.send(http.MethodPatch, "/v1/taxpayers/"+first, `{"display_name":"x"}`, token))
	assert.Equal(t, http.StatusPaymentRequired, rec.Code)
	assert.Equal(t, gen.ErrorCodeEntitlementRequired, decodeError(t, rec).Error.Code)
	rec = c.ServeOp(t, contracttest.Op("addTaxpayerRegime"), env.handler,
		env.send(http.MethodPost, "/v1/taxpayers/"+first+"/regimes", `{"regime_code":"606"}`, token))
	assert.Equal(t, http.StatusPaymentRequired, rec.Code)
	rec = c.ServeOp(t, contracttest.Op("updateTaxpayerRegime"), env.handler,
		env.send(http.MethodPatch, "/v1/taxpayers/"+first+"/regimes/612", `{"options":{}}`, token))
	assert.Equal(t, http.StatusPaymentRequired, rec.Code)

	rec = c.ServeOp(t, contracttest.Op("archiveTaxpayer"), env.handler,
		env.send(http.MethodPost, "/v1/taxpayers/"+first+"/archive", "", token))
	assert.Equal(t, http.StatusOK, rec.Code, "archiving is allowed read-only")
	rec = c.ServeOp(t, contracttest.Op("unarchiveTaxpayer"), env.handler,
		env.send(http.MethodPost, "/v1/taxpayers/"+first+"/unarchive", "", token))
	assert.Equal(t, http.StatusPaymentRequired, rec.Code)
	rec = c.ServeOp(t, contracttest.Op("listRegimes"), env.handler,
		env.send(http.MethodGet, "/v1/regimes?person_type=fisica", "", token))
	assert.Equal(t, http.StatusOK, rec.Code, "the catalog is readable read-only")
	rec = c.ServeOp(t, contracttest.Op("deleteTaxpayer"), env.handler,
		env.send(http.MethodDelete, "/v1/taxpayers/"+first+"?confirm=true", "", token))
	assert.Equal(t, http.StatusNoContent, rec.Code, "deleting is allowed read-only")
}

// MODULE-taxpayers S-10: 30 writes per user per hour, shared by every
// mutation; the next one is 429 with Retry-After, reads still pass, and
// another user keeps their own budget.
func TestContract_taxpayersWriteBudget(t *testing.T) {
	c := contracttest.Load(t)
	env := newContractEnv(t, pingOK{})
	env.apply(t, []string{"module_auth", "module_taxpayers"}, flagtest.Static{"module_auth": true, "module_taxpayers": true})
	ana := env.signUp(t, signUpEmail, acceptableSample).Tokens.AccessToken
	bea := env.signUp(t, beaEmail, otherSample).Tokens.AccessToken

	rec := env.do(env.send(http.MethodPost, "/v1/taxpayers", createJSON("", "fisica", rfcAna, "Ana", regimeJSON("612")), ana))
	require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())
	id := decodeTaxpayer(t, rec).Id.String()
	for i := range 29 { // 1 create + 29 archive/unarchive = the 30 of the budget
		action := "/archive"
		if i%2 == 1 {
			action = "/unarchive"
		}
		rec = env.do(env.send(http.MethodPost, "/v1/taxpayers/"+id+action, "", ana))
		require.Equal(t, http.StatusOK, rec.Code, "write %d: %s", i+2, rec.Body.String())
	}

	rec = c.ServeOp(t, contracttest.Op("updateTaxpayer"), env.handler,
		env.send(http.MethodPatch, "/v1/taxpayers/"+id, `{"display_name":"x"}`, ana))
	assert.Equal(t, http.StatusTooManyRequests, rec.Code)
	assert.NotEmpty(t, rec.Header().Get("Retry-After"))
	assert.Equal(t, gen.ErrorCodeRateLimited, decodeError(t, rec).Error.Code)
	rec = c.ServeOp(t, contracttest.Op("deleteTaxpayer"), env.handler,
		env.send(http.MethodDelete, "/v1/taxpayers/"+id+"?confirm=true", "", ana))
	assert.Equal(t, http.StatusTooManyRequests, rec.Code, "one budget for every write")

	rec = c.ServeOp(t, contracttest.Op("getTaxpayer"), env.handler, env.send(http.MethodGet, "/v1/taxpayers/"+id, "", ana))
	assert.Equal(t, http.StatusOK, rec.Code, "reads spend no write budget")
	rec = c.ServeOp(t, contracttest.Op("createTaxpayer"), env.handler,
		env.send(http.MethodPost, "/v1/taxpayers", createJSON("", "fisica", rfcAna, "Bea", regimeJSON("612")), bea))
	assert.Equal(t, http.StatusCreated, rec.Code, "another user's budget is untouched")
}

// US-T5 (security gate, PR #60): every operation with a {taxpayerId} answers
// another user's taxpayer exactly as one that does not exist: same status,
// code, message and details, so no path parameter is an existence oracle.
// The regime ops are probed with a regime the victim really holds.
func TestContract_taxpayersForeignIsIndistinguishableFromMissing(t *testing.T) {
	c := contracttest.Load(t)
	env := newContractEnv(t, pingOK{})
	env.apply(t, []string{"module_auth", "module_taxpayers"}, flagtest.Static{"module_auth": true, "module_taxpayers": true})
	ana := env.signUp(t, signUpEmail, acceptableSample).Tokens.AccessToken
	bea := env.signUp(t, beaEmail, otherSample).Tokens.AccessToken

	rec := env.do(env.send(http.MethodPost, "/v1/taxpayers",
		createJSON("", "fisica", rfcAna, "Ana", regimeJSON("612")), ana))
	require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())
	victim := decodeTaxpayer(t, rec).Id.String()

	ops := []struct {
		name                 string
		op                   contracttest.Operation
		method, suffix, body string
	}{
		{"getTaxpayer", contracttest.Op("getTaxpayer"), http.MethodGet, "", ""},
		{"updateTaxpayer", contracttest.Op("updateTaxpayer"), http.MethodPatch, "", `{"display_name":"Robo","rfc":"` + rfcOther + `"}`},
		{"deleteTaxpayer", contracttest.Op("deleteTaxpayer"), http.MethodDelete, "?confirm=true", ""},
		{"archiveTaxpayer", contracttest.Op("archiveTaxpayer"), http.MethodPost, "/archive", ""},
		{"unarchiveTaxpayer", contracttest.Op("unarchiveTaxpayer"), http.MethodPost, "/unarchive", ""},
		{"addTaxpayerRegime", contracttest.Op("addTaxpayerRegime"), http.MethodPost, "/regimes", `{"regime_code":"606"}`},
		{"updateTaxpayerRegime", contracttest.Op("updateTaxpayerRegime"), http.MethodPatch, "/regimes/612", `{"options":{}}`},
		{"archiveTaxpayerRegime", contracttest.Op("archiveTaxpayerRegime"), http.MethodDelete, "/regimes/612?end_period=2026-10", ""},
	}
	for _, o := range ops {
		t.Run(o.name, func(t *testing.T) {
			foreign := c.ServeOp(t, o.op, env.handler,
				env.send(o.method, "/v1/taxpayers/"+victim+o.suffix, o.body, bea))
			missing := c.ServeOp(t, o.op, env.handler,
				env.send(o.method, "/v1/taxpayers/"+unknownTaxpayer+o.suffix, o.body, bea))

			require.Equal(t, http.StatusNotFound, foreign.Code, foreign.Body.String())
			require.Equal(t, http.StatusNotFound, missing.Code, missing.Body.String())
			fe, me := decodeError(t, foreign), decodeError(t, missing)
			fe.Error.RequestId, me.Error.RequestId = "", ""
			assert.Equal(t, me, fe, "a foreign taxpayer reads exactly like a missing one")
		})
	}

	rec = env.do(env.send(http.MethodGet, "/v1/taxpayers/"+victim, "", ana))
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	got := decodeTaxpayer(t, rec)
	assert.Equal(t, "Ana", got.DisplayName, "nothing the other user sent was written")
	assert.Nil(t, got.ArchivedAt)
	require.Len(t, got.Regimes, 1)
	assert.Nil(t, got.Regimes[0].EndPeriod)
}
