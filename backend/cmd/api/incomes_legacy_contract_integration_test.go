//go:build integration

package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/adapter/http/contracttest"
	"github.com/romodevx/control-contable/backend/adapter/http/gen"
	"github.com/romodevx/control-contable/backend/internal/platform/flags/flagtest"
	"github.com/romodevx/control-contable/backend/internal/platform/pgtest"
)

// unbackfillSQL turns an income back into one stored before WH-06 that the
// WH-08 backfill has not derived: no facts, no derivation, only its
// counterparty_type and typed amounts.
const unbackfillSQL = `
UPDATE incomes SET counterparty_kind = NULL, operation_kind = NULL, withholding_rule_ids = '{}',
                   withholding_table_version = NULL, withholding_cpa_unverified = false
WHERE id = $1`

// Contract 0.12.0 assumes no deploy order (ADR-0008 Amendment 2): an income
// the backfill has not derived is served, valid against the contract, with
// the backfill's facts and a null withholding; an edit derives it.
func TestContract_incomesLegacyRecordIsServed(t *testing.T) {
	c := contracttest.Load(t)
	env := newContractEnv(t, pingOK{})
	env.apply(t, []string{"module_auth", "module_taxpayers", "module_incomes"},
		flagtest.Static{"module_auth": true, "module_taxpayers": true, "module_incomes": true})
	actors := env.incomeActors(t, c)
	id := clientIncomeID(900)
	rec := c.ServeOp(t, contracttest.Op("createIncome"), env.handler,
		env.send(http.MethodPost, incomesPath(incomeTaxpayerA), incomeJSON(t, id, nil), actors.ana))
	require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())
	require.NoError(t, pgtest.AsUser(t.Context(), env.app, actors.anaID, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(t.Context(), unbackfillSQL, id)
		return err
	}))

	rec = c.ServeOp(t, contracttest.Op("getIncome"), env.handler,
		env.send(http.MethodGet, incomesPath(incomeTaxpayerA, "/", id), "", actors.ana))
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	got := decodeIncome(t, rec)
	assert.Equal(t, gen.CounterpartyKind("pm"), got.CounterpartyKind, "counterparty_type persona_moral (§5.2)")
	assert.Equal(t, gen.OperationKind("professional_services"), got.OperationKind, "612's default")
	assert.Nil(t, got.Withholding, "not derived yet: null, never a 500")
	assert.EqualValues(t, 100000, got.IsrWithheldCents, "the stored amount")

	rec = c.ServeOp(t, contracttest.Op("listIncomes"), env.handler,
		env.send(http.MethodGet, incomesPath(incomeTaxpayerA, "?period=2026-09"), "", actors.ana))
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	var page gen.IncomePage
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &page))
	require.Len(t, page.Items, 1)
	assert.Nil(t, page.Items[0].Withholding)

	rec = c.ServeOp(t, contracttest.Op("updateIncome"), env.handler,
		env.send(http.MethodPatch, incomesPath(incomeTaxpayerA, "/", id), `{"description":"editado"}`, actors.ana))
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	edited := decodeIncome(t, rec)
	require.NotNil(t, edited.Withholding, "an edit derives on the served facts")
	assert.Equal(t, []string{"ISR-5", "IVA-6"}, edited.Withholding.RuleIds)
}
