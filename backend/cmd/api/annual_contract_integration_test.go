//go:build integration

package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/adapter/http/contracttest"
	"github.com/romodevx/control-contable/backend/adapter/http/gen"
	"github.com/romodevx/control-contable/backend/internal/platform/flags/flagtest"
)

// Contract tests of module_annual (ANN-15) through the real composition
// root over Postgres as cc_app: the estimate of TX-NET-01's taxpayer in the
// year in progress ("today" is estToday, 20 October 2026), the inputs'
// writes, another user's taxpayer (404) and the flag with its
// prerequisites. The module_annual flag row is the migration's (ANN-22).

// annualModules are the flags these suites set.
var annualModules = []string{
	"module_auth", "module_taxpayers", "module_incomes", "module_expenses", "module_tax_engine", "module_annual",
}

func annualAllOn() flagtest.Static {
	on := flagtest.Static{}
	for _, m := range annualModules {
		on[m] = true
	}
	return on
}

func annualEstimateTarget(taxpayerID string, year int) string {
	return fmt.Sprintf("/v1/taxpayers/%s/annual-estimates/%d", taxpayerID, year)
}

func TestAnnualContract_estimateAndInputsThroughTheStack(t *testing.T) {
	c := contracttest.Load(t)
	env := newContractEnv(t, pingOK{}, withTaxEngineDay(estToday))
	env.apply(t, annualModules, annualAllOn())
	ana := env.txNet01Taxpayer(t, signUpEmail, acceptableSample, rfcAna)
	bea := env.txNet01Taxpayer(t, beaEmail, otherSample, "AAAA010101AA0")
	base := "/v1/taxpayers/" + ana.taxpayer

	rec := c.ServeOp(t, contracttest.Op("getAnnualEstimate"), env.handler,
		env.send(http.MethodGet, annualEstimateTarget(ana.taxpayer, 2026), "", ana.token))
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	var est gen.AnnualEstimate
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &est))
	assert.Equal(t, gen.AnnualYearStatus("in_progress"), est.YearStatus)
	require.NotNil(t, est.DataThrough)
	assert.Equal(t, "2026-09", *est.DataThrough)
	require.NotEmpty(t, est.Sections)
	assert.Equal(t, gen.AnnualSectionKind("pf_art152"), est.Sections[0].Kind)

	rec = c.ServeOp(t, contracttest.Op("getAnnualBreakdown"), env.handler,
		env.send(http.MethodGet, annualEstimateTarget(ana.taxpayer, 2026)+"/breakdown?section=pf_art152", "", ana.token))
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())

	// IDOR parity: another user's taxpayer is the unknown taxpayer's 404.
	rec = c.ServeOp(t, contracttest.Op("getAnnualEstimate"), env.handler,
		env.send(http.MethodGet, annualEstimateTarget(ana.taxpayer, 2026), "", bea.token))
	assert.Equal(t, http.StatusNotFound, rec.Code, rec.Body.String())

	for _, step := range []struct {
		op                   contracttest.Operation
		method, target, body string
		status               int
	}{
		{contracttest.Op("putAnnualInputs"), http.MethodPut, base + "/annual-inputs/2026",
			`{"prior_year_accumulable_income_cents":50000000,"other_exempt_income_cents":0,"told_employer_will_file":false}`, http.StatusOK},
		{contracttest.Op("getAnnualInputs"), http.MethodGet, base + "/annual-inputs/2026", "", http.StatusOK},
		{contracttest.Op("createSalaryStatement"), http.MethodPost, base + "/salary-statements",
			`{"fiscal_year":2026,"accumulable_cents":30000000,"exempt_cents":0,"isr_withheld_cents":4000000}`, http.StatusCreated},
		{contracttest.Op("listSalaryStatements"), http.MethodGet, base + "/salary-statements?fiscal_year=2026", "", http.StatusOK},
		{contracttest.Op("createPersonalDeduction"), http.MethodPost, base + "/personal-deductions",
			`{"category":"medical","amount_cents":120000,"paid_on":"2026-05-10","payment_method":"cash","beneficiary":"self"}`, http.StatusCreated},
		{contracttest.Op("listPersonalDeductions"), http.MethodGet, base + "/personal-deductions?fiscal_year=2026", "", http.StatusOK},
		{contracttest.Op("getFiledPayments"), http.MethodGet, base + "/regimes/612/filed-payments/2026", "", http.StatusOK},
		{contracttest.Op("putFiledPayments"), http.MethodPut, base + "/regimes/612/filed-payments/2026",
			`{"months":[{"month":3,"amount_cents":200000}]}`, http.StatusOK},
	} {
		rec := c.ServeOp(t, step.op, env.handler, env.send(step.method, step.target, step.body, ana.token))
		require.Equal(t, step.status, rec.Code, "%s %s: %s", step.method, step.target, rec.Body.String())
	}

	// The inputs move the estimate: the constancia joins pf_art152.
	rec = env.do(env.send(http.MethodGet, annualEstimateTarget(ana.taxpayer, 2026), "", ana.token))
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &est))
	assert.Contains(t, est.Sections[0].Regimes, "605")
}

// Implied OFF: module_annual ON with module_tax_engine OFF hides every
// annual route (flags.Prerequisites, through the real flags service).
func TestAnnualContract_impliedOffWithoutTheEngine(t *testing.T) {
	c := contracttest.Load(t)
	env := newContractEnv(t, pingOK{}, withTaxEngineDay(estToday))
	flagsOn := annualAllOn()
	env.apply(t, annualModules, flagsOn)
	ana := env.txNet01Taxpayer(t, signUpEmail, acceptableSample, rfcAna)
	flagsOn["module_tax_engine"] = false
	env.apply(t, annualModules, flagsOn)

	rec := c.ServeOp(t, contracttest.Op("getAnnualEstimate"), env.handler,
		env.send(http.MethodGet, annualEstimateTarget(ana.taxpayer, 2026), "", ana.token))

	flagtest.AssertHidden(t, rec)
}
