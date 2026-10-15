package annual_test

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/platform/apperr"
	"github.com/romodevx/control-contable/backend/internal/platform/clock"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
	"github.com/romodevx/control-contable/backend/internal/taxengine/usecase/annual"
)

func (w *world) salary(now time.Time) *annual.SalaryStatements {
	return annual.NewSalaryStatements(annual.SalaryDeps{
		Tx: w.tx, Access: w.access, Taxpayers: w.db, Store: w.db, Clock: clock.NewFixed(now),
		NewID: fixedID, Logger: w.logger(),
	})
}

// mainEmployer is the first AN-05 constancia.
func mainEmployer(id string) annual.SalaryInput {
	return annual.SalaryInput{
		ID: id, TaxpayerID: fisicaA, FiscalYear: 2026, EmployerLabel: ptr("Empleo principal"),
		Accumulable: 30_000_000, Exempt: 0, ISRWithheld: 4_000_000, Simultaneous: true,
	}
}

func mainFields() annual.SalaryFields {
	return annual.SalaryFields{
		FiscalYear: 2026, EmployerLabel: "Empleo principal", Accumulable: 30_000_000, ISRWithheld: 4_000_000, Simultaneous: true,
	}
}

func (w *world) seedStatement(id, taxpayerID string, f annual.SalaryFields) {
	w.db.statements[id] = ownedStatement{owner: userA, row: annual.StoredStatement{
		ID: id, TaxpayerID: taxpayerID, SalaryFields: f, CreatedAt: stamp, UpdatedAt: stamp,
	}}
}

func (w *world) fillStatements(n, year int) {
	for i := range n {
		f := mainFields()
		f.FiscalYear = year
		w.seedStatement(fmt.Sprintf("0199d4b0-ffff-7000-8000-%012d", i), fisicaA, f)
	}
}

func TestSalary_CreateStoresAndAudits(t *testing.T) {
	w := newWorld()

	res, err := w.salary(today).Create(t.Context(), userA, mainEmployer(clientID))
	require.NoError(t, err)
	assert.True(t, res.Created)
	assert.Empty(t, res.Warnings, "fisicaA holds 605 in 2026")
	assert.Equal(t, clientID, res.Statement.ID)
	assert.Equal(t, fisicaA, res.Statement.TaxpayerID)
	assert.Equal(t, mainFields(), res.Statement.SalaryFields)
	assert.Equal(t, []string{"lock-id", "find:none", "taxpayer:write", "list:2:2026", "count", "insert", "find:none"}, w.db.calls)
	assert.Contains(t, w.logs.String(), `"msg":"salary_statement.created"`)
	assert.NotContains(t, w.logs.String(), "30000000")
	assert.NotContains(t, w.logs.String(), "Empleo")
}

func TestSalary_WarnsWithout605(t *testing.T) {
	w := newWorld()
	in := mainEmployer(clientID)
	in.TaxpayerID = fisica2A

	res, err := w.salary(today).Create(t.Context(), userA, in)
	require.NoError(t, err, "allowed for any persona física (OQ-A2)")
	assert.Equal(t, []annual.Warning{{Code: annual.WarningSalaryRegimeNotHeld, Field: annual.FieldFiscalYear}}, res.Warnings)

	again, err := w.salary(today).Create(t.Context(), userA, in)
	require.NoError(t, err)
	assert.False(t, again.Created)
	assert.Equal(t, res.Warnings, again.Warnings, "a replay carries the write's warnings")
}

func TestSalary_LabelIsOptionalAndNormalized(t *testing.T) {
	w := newWorld()
	in := mainEmployer(clientID)
	in.EmployerLabel = nil
	res, err := w.salary(today).Create(t.Context(), userA, in)
	require.NoError(t, err)
	assert.Empty(t, res.Statement.EmployerLabel)

	in = mainEmployer(clientID2)
	in.EmployerLabel = ptr("Compañía")
	res, err = w.salary(today).Create(t.Context(), userA, in)
	require.NoError(t, err)
	assert.Equal(t, "Compañía", res.Statement.EmployerLabel)
}

func TestSalary_CreateIdempotency(t *testing.T) {
	w := newWorld()
	uc := w.salary(today)
	first, err := uc.Create(t.Context(), userA, mainEmployer(clientID))
	require.NoError(t, err)
	w.access.readOnly = true

	again, err := uc.Create(t.Context(), userA, mainEmployer(clientID))
	require.NoError(t, err)
	assert.False(t, again.Created)
	assert.Equal(t, first.Statement, again.Statement)

	other := mainEmployer(clientID)
	other.Exempt = 1
	_, err = uc.Create(t.Context(), userA, other)
	assert.Equal(t, []string{"id/id_reused"}, requireCode(t, err, apperr.CodeConflict))

	w.db.statements[clientID2] = ownedStatement{owner: userB, row: annual.StoredStatement{ID: clientID2, TaxpayerID: fisicaB}}
	w.access.readOnly = false
	_, err = uc.Create(t.Context(), userA, mainEmployer(clientID2))
	assert.Empty(t, requireCode(t, err, apperr.CodeConflict))
}

func TestSalary_CreateRefusals(t *testing.T) {
	tests := []struct {
		name   string
		setup  func(w *world)
		in     func() annual.SalaryInput
		code   apperr.Code
		fields []string
	}{
		{"malformed taxpayer", nil, func() annual.SalaryInput { in := mainEmployer(clientID); in.TaxpayerID = "x"; return in }, apperr.CodeNotFound, []string{}},
		{"malformed id", nil, func() annual.SalaryInput { return mainEmployer("x") }, apperr.CodeValidationFailed, []string{"id/invalid"}},
		{"read-only", func(w *world) { w.access.readOnly = true }, func() annual.SalaryInput { return mainEmployer(clientID) }, apperr.CodeEntitlementRequired, []string{}},
		{"another user's taxpayer", nil, func() annual.SalaryInput { in := mainEmployer(clientID); in.TaxpayerID = fisicaB; return in }, apperr.CodeNotFound, []string{}},
		{"archived", func(w *world) { w.archive(fisicaA) }, func() annual.SalaryInput { return mainEmployer(clientID) },
			apperr.CodeBusinessRuleViolation, []string{"taxpayerId/taxpayer_archived"}},
		{"persona moral", nil, func() annual.SalaryInput { in := mainEmployer(clientID); in.TaxpayerID = moralA; return in },
			apperr.CodeBusinessRuleViolation, []string{"taxpayerId/person_type_not_allowed"}},
		{"before 2026", nil, func() annual.SalaryInput { in := mainEmployer(clientID); in.FiscalYear = 2025; return in },
			apperr.CodeBusinessRuleViolation, []string{"fiscal_year/parameters_unavailable"}},
		{"20 constancias in the year", func(w *world) { w.fillStatements(20, 2026) }, func() annual.SalaryInput { return mainEmployer(clientID) },
			apperr.CodeBusinessRuleViolation, []string{"body/limit_reached"}},
		{"future year", nil, func() annual.SalaryInput { in := mainEmployer(clientID); in.FiscalYear = 2027; return in },
			apperr.CodeValidationFailed, []string{"fiscal_year/out_of_range"}},
		{"label rules", nil, func() annual.SalaryInput {
			in := mainEmployer(clientID)
			in.EmployerLabel = ptr(strings.Repeat("x", 61))
			return in
		},
			apperr.CodeValidationFailed, []string{"employer_label/invalid"}},
		{"untrimmed label", nil, func() annual.SalaryInput { in := mainEmployer(clientID); in.EmployerLabel = ptr("x "); return in },
			apperr.CodeValidationFailed, []string{"employer_label/invalid"}},
		{"negative figures", nil, func() annual.SalaryInput {
			in := mainEmployer(clientID)
			in.Accumulable, in.Exempt, in.ISRWithheld = -1, -1, -1
			return in
		}, apperr.CodeValidationFailed, []string{"accumulable_cents/invalid", "exempt_cents/invalid", "isr_withheld_cents/invalid"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := newWorld()
			if tt.setup != nil {
				tt.setup(w)
			}
			_, err := w.salary(today).Create(t.Context(), userA, tt.in())
			assert.Equal(t, tt.fields, requireCode(t, err, tt.code))
			assert.NotContains(t, w.db.calls, "insert")
		})
	}
	t.Run("19 in the year and 20 in another are fine", func(t *testing.T) {
		w := newWorld()
		w.fillStatements(19, 2026)
		_, err := w.salary(today).Create(t.Context(), userA, mainEmployer(clientID))
		require.NoError(t, err)
	})
}

func TestSalary_Update(t *testing.T) {
	t.Run("merge-patch; null clears the label", func(t *testing.T) {
		w := newWorld()
		w.seedStatement(clientID, fisicaA, mainFields())

		res, err := w.salary(today).Update(t.Context(), userA, fisicaA, clientID,
			annual.SalaryPatch{LeftBeforeYearEnd: ptr(true), EmployerLabel: annual.Null[string]()})
		require.NoError(t, err)
		want := mainFields()
		want.LeftBeforeYearEnd, want.EmployerLabel = true, ""
		assert.Equal(t, want, res.Statement.SalaryFields)
		assert.Equal(t, []string{"find:write", "taxpayer:write", "update", "find:none"}, w.db.calls)
		assert.Contains(t, w.logs.String(), `"msg":"salary_statement.updated"`)
	})
	t.Run("moving the year applies its bound and its warning", func(t *testing.T) {
		w := newWorld()
		mar2027 := time.Date(2027, time.March, 1, 12, 0, 0, 0, clock.Location())
		w.seedStatement(clientID, fisica2A, mainFields())
		uc := w.salary(mar2027)

		res, err := uc.Update(t.Context(), userA, fisica2A, clientID, annual.SalaryPatch{FiscalYear: ptr(2027)})
		require.NoError(t, err)
		assert.Equal(t, 2027, res.Statement.FiscalYear)
		assert.Contains(t, w.db.calls, "list:2:2027")
		assert.Equal(t, []annual.Warning{{Code: annual.WarningSalaryRegimeNotHeld, Field: annual.FieldFiscalYear}}, res.Warnings)

		w.seedStatement(clientID2, fisicaA, mainFields())
		w.fillStatements(20, 2027)
		_, err = uc.Update(t.Context(), userA, fisicaA, clientID2, annual.SalaryPatch{FiscalYear: ptr(2027)})
		assert.Equal(t, []string{"body/limit_reached"}, requireCode(t, err, apperr.CodeBusinessRuleViolation))

		_, err = uc.Update(t.Context(), userA, fisicaA, clientID2, annual.SalaryPatch{FiscalYear: ptr(2025)})
		assert.Equal(t, []string{"fiscal_year/parameters_unavailable"}, requireCode(t, err, apperr.CodeBusinessRuleViolation))
		_, err = uc.Update(t.Context(), userA, fisicaA, clientID2, annual.SalaryPatch{FiscalYear: ptr(2028)})
		assert.Equal(t, []string{"fiscal_year/out_of_range"}, requireCode(t, err, apperr.CodeValidationFailed))
	})
	t.Run("refusals", func(t *testing.T) {
		w := newWorld()
		w.seedStatement(clientID, fisicaA, mainFields())
		uc := w.salary(today)
		for _, ids := range [][2]string{{fisica2A, clientID}, {fisicaA, clientID2}, {fisicaA, "x"}, {"x", clientID}} {
			_, err := uc.Update(t.Context(), userA, ids[0], ids[1], annual.SalaryPatch{Simultaneous: ptr(false)})
			requireCode(t, err, apperr.CodeNotFound)
		}
		_, err := uc.Update(t.Context(), userA, fisicaA, clientID, annual.SalaryPatch{Exempt: ptr(money.Cents(-5))})
		assert.Equal(t, []string{"exempt_cents/invalid"}, requireCode(t, err, apperr.CodeValidationFailed))

		w.access.readOnly = true
		_, err = uc.Update(t.Context(), userA, fisicaA, clientID, annual.SalaryPatch{Simultaneous: ptr(false)})
		requireCode(t, err, apperr.CodeEntitlementRequired)
		w.access.readOnly = false
		w.archive(fisicaA)
		_, err = uc.Update(t.Context(), userA, fisicaA, clientID, annual.SalaryPatch{Simultaneous: ptr(false)})
		requireCode(t, err, apperr.CodeBusinessRuleViolation)
	})
}

func TestSalary_Delete(t *testing.T) {
	w := newWorld()
	w.access.readOnly = true
	w.seedStatement(clientID, fisicaA, mainFields())
	uc := w.salary(today)

	require.NoError(t, uc.Delete(t.Context(), userA, fisicaA, clientID), "open to a read-only user")
	assert.NotContains(t, w.db.statements, clientID)
	assert.Equal(t, 0, w.access.calls)
	assert.Contains(t, w.logs.String(), `"msg":"salary_statement.deleted"`)

	err := uc.Delete(t.Context(), userA, fisicaA, clientID)
	requireCode(t, err, apperr.CodeNotFound)

	w.seedStatement(clientID, fisicaA, mainFields())
	err = uc.Delete(t.Context(), userA, fisica2A, clientID)
	requireCode(t, err, apperr.CodeNotFound)
	w.archive(fisicaA)
	err = uc.Delete(t.Context(), userA, fisicaA, clientID)
	assert.Equal(t, []string{"taxpayerId/taxpayer_archived"}, requireCode(t, err, apperr.CodeBusinessRuleViolation))
}

func TestSalary_List(t *testing.T) {
	w := newWorld()
	uc := w.salary(today)
	_, err := uc.Create(t.Context(), userA, mainEmployer(clientID2))
	require.NoError(t, err)
	second := mainEmployer(clientID)
	second.EmployerLabel, second.Accumulable, second.ISRWithheld = ptr("Segundo empleo"), 15_000_000, 800_000
	_, err = uc.Create(t.Context(), userA, second)
	require.NoError(t, err)

	list, err := uc.List(t.Context(), userA, fisicaA, 2026)
	require.NoError(t, err)
	require.Len(t, list.Items, 2)
	assert.Equal(t, clientID2, list.Items[0].ID, "by created_at, then id")
	assert.Equal(t, annual.SalaryTotals{Accumulable: 45_000_000, ISRWithheld: 4_800_000}, list.Totals)
	assert.Equal(t, 2026, list.FiscalYear)

	moral, err := uc.List(t.Context(), userA, moralA, 2026)
	require.NoError(t, err)
	assert.NotNil(t, moral.Items)
	assert.Empty(t, moral.Items)

	_, err = uc.List(t.Context(), userA, fisicaA, 2027)
	assert.Equal(t, []string{"fiscal_year/out_of_range"}, requireCode(t, err, apperr.CodeValidationFailed))
	_, err = uc.List(t.Context(), userA, fisicaA, 2025)
	assert.Equal(t, []string{"fiscal_year/parameters_unavailable"}, requireCode(t, err, apperr.CodeBusinessRuleViolation))
	_, err = uc.List(t.Context(), userA, fisicaB, 2026)
	requireCode(t, err, apperr.CodeNotFound)
}
