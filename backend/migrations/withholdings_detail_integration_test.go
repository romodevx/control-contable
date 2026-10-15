//go:build integration

package migrations_test

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/platform/pgtest"
	"github.com/romodevx/control-contable/backend/migrations"
)

// Withholdings stored detail (WH-06 / WH-07): isr_rate, iva_basis and withholder of a
// derived row, so a read serves the contract Withholding without deriving again (ADR-0008).

const (
	withholdingsDetailVersion         = 20261017100000
	withholdingsDetailValidateVersion = 20261017101000
)

// derivedRuleSQL is a stored WH-03 derivation (US-W1: RESICO PF bills a persona moral);
// detailSQL completes it with the detail columns.
const derivedRuleSQL = "withholding_rule_ids = '{ISR-4,IVA-6}', withholding_table_version = '2026.1'"

func detailSQL(rate, basis, withholder string) string {
	return "withholding_isr_rate = " + rate + ", withholding_iva_basis = " + basis + ", withholding_withholder = " + withholder
}

// derivedDetailSQL is the complete US-W1 stored derivation.
var derivedDetailSQL = derivedRuleSQL + ", " + detailSQL("0.0125", "'two_thirds_of_iva'", "'receptor'")

var withholdingDetailConstraints = []string{
	"incomes_withholding_isr_rate_range",
	"incomes_withholding_iva_basis_valid",
	"incomes_withholding_withholder_valid",
	"incomes_withholding_detail_together",
	"incomes_withholding_detail_needs_derivation",
	"expenses_withholding_isr_rate_range",
	"expenses_withholding_iva_basis_valid",
	"expenses_withholding_withholder_valid",
	"expenses_withholding_detail_together",
	"expenses_withholding_detail_needs_derivation",
}

func TestWithholdingsDetail_IsWellFormed(t *testing.T) {
	app := seedIncomeOwners(t, pgtest.NewDatabase(t))
	require.NoError(t, insertIncome(t, app, userA, validIncome(userA, taxpayerA, incomeA)))
	seedExpense(t, app, userA, nil)
	derived := func(rate, basis, withholder string) string {
		return derivedRuleSQL + ", " + detailSQL(rate, basis, withholder)
	}
	rejected := []struct{ name, set string }{
		{"rate above 100 %", derived("1.5", "'none'", "'none'")},
		{"negative rate", derived("-0.01", "'none'", "'none'")},
		{"unknown iva basis", derived("0.1", "'six_percent'", "'receptor'")},
		{"wire withholder spelling", derived("0.1", "'two_thirds_of_iva'", "'counterparty'")},
		{"rate without the rest", derived("0.1", "NULL", "'receptor'")},
		{"withholder without the rest", derived("NULL", "NULL", "'receptor'")},
		{"detail on an underived row", detailSQL("0", "'none'", "'none'")},
	}
	for _, table := range []string{"incomes", "expenses"} {
		for _, c := range rejected {
			t.Run(table+": "+c.name, func(t *testing.T) {
				_, err := execAsUser(t, app, userA, "UPDATE "+table+" SET "+c.set)
				requireSQLState(t, err, sqlstateCheckViolation)
			})
		}
		t.Run(table+": accepted", func(t *testing.T) {
			for _, set := range []string{
				derivedDetailSQL,
				derived("0", "'none'", "'none'"),
				derived("0.2", "'all_of_iva'", "'platform'"),
				"withholding_rule_ids = '{}', withholding_table_version = NULL, " + detailSQL("NULL", "NULL", "NULL"),
			} {
				affected, err := execAsUser(t, app, userA, "UPDATE "+table+" SET "+set)
				require.NoError(t, err, set)
				assert.EqualValues(t, 1, affected, set)
			}
		})
	}
}

// RLS still keeps another user's rows out of reach on the new columns.
func TestWithholdingsDetail_AppWritesOnlyItsOwnRows(t *testing.T) {
	app := seedIncomeOwners(t, pgtest.NewDatabase(t))
	for _, owner := range []string{userA, userB} {
		require.NoError(t, insertIncome(t, app, owner, validIncome(owner, taxpayerOf[owner], incomeOf[owner])))
		seedExpense(t, app, owner, nil)
	}
	for _, table := range []string{"incomes", "expenses"} {
		affected, err := execAsUser(t, app, userA, "UPDATE "+table+" SET "+derivedDetailSQL+" WHERE user_id = $1", userB)
		require.NoError(t, err)
		assert.Zero(t, affected, "RLS: %s of another user", table)

		affected, err = execAsUser(t, app, userA, "UPDATE "+table+" SET "+derivedDetailSQL)
		require.NoError(t, err)
		assert.EqualValues(t, 1, affected, table)
	}
}

func TestWithholdingsDetailMigration_ValidatesThenDownRemovesTheColumns(t *testing.T) {
	db := pgtest.NewDatabase(t)
	owner := db.Connect(t, pgtest.RoleOwner)
	p, err := migrations.NewProvider(owner)
	require.NoError(t, err)
	_, err = p.DownTo(t.Context(), withholdingsDetailVersion-1)
	require.NoError(t, err)
	app := seedIncomeOwners(t, db)
	require.NoError(t, insertIncome(t, app, userA, validIncome(userA, taxpayerA, incomeA)))

	_, err = p.UpTo(t.Context(), withholdingsDetailVersion)
	require.NoError(t, err)
	assert.Equal(t, withholdingDetailConstraints, detailConstraintsValidated(t, owner, false), "added NOT VALID")
	_, err = p.UpTo(t.Context(), withholdingsDetailValidateVersion)
	require.NoError(t, err)
	assert.Equal(t, withholdingDetailConstraints, detailConstraintsValidated(t, owner, true), "validated")

	_, err = p.DownTo(t.Context(), withholdingsDetailVersion-1)
	require.NoError(t, err)
	var left int
	require.NoError(t, owner.QueryRowContext(t.Context(), `SELECT count(*) FROM information_schema.columns
	  WHERE table_name IN ('incomes', 'expenses')
	    AND column_name IN ('withholding_isr_rate', 'withholding_iva_basis', 'withholding_withholder')`).Scan(&left))
	assert.Zero(t, left)

	require.NoError(t, migrations.Up(t.Context(), owner))
	require.NoError(t, migrations.Up(t.Context(), owner), "a second up is a no-op")
}

// detailConstraintsValidated lists, in withholdingDetailConstraints order, the detail
// constraints whose convalidated flag equals validated.
func detailConstraintsValidated(t *testing.T, owner *sql.DB, validated bool) []string {
	t.Helper()
	found := map[string]bool{}
	for _, name := range queryStrings(t, owner, `SELECT conname FROM pg_catalog.pg_constraint
		WHERE conrelid IN ('public.incomes'::regclass, 'public.expenses'::regclass)
		  AND contype = 'c' AND convalidated = $1`, validated) {
		found[name] = true
	}
	var out []string
	for _, name := range withholdingDetailConstraints {
		if found[name] {
			out = append(out, name)
		}
	}
	return out
}
