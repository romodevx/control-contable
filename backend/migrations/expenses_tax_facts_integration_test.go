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

// Expense tax facts (M4-08): the cash rule and the creditable-IVA candidate the domain
// assessed on the dated parameters (MODULE-expenses §2.4, §2.5) are stored on the row, so
// the tax engine's aggregate sums them instead of re-deriving them from rules and
// parameters that may have changed since.

const expensesTaxFactsVersion = 20261014150000

func TestExpenseTaxFacts_RejectInconsistentRows(t *testing.T) {
	app := pgtest.NewDatabase(t).Connect(t, pgtest.RoleApp)
	seedExpenseOwners(t, app)
	base := validExpense(userA, taxpayerA)
	cash := expenseColumns{"payment_method": "'cash'", "deductible": "false"}
	rejected := []struct {
		name      string
		overrides expenseColumns
		want      string
	}{
		{"cash over the limit on a card payment", expenseColumns{"cash_over_limit": "true"}, sqlstateCheckViolation},
		{"negative creditable iva", expenseColumns{"creditable_iva_cents": "-1"}, sqlstateCheckViolation},
		{"creditable iva above the iva", expenseColumns{"creditable_iva_cents": "8001"}, sqlstateCheckViolation},
		{"creditable iva without a cfdi", expenseColumns{"has_cfdi": "false", "creditable_iva_cents": "8000"}, sqlstateCheckViolation},
		{"creditable iva on a cash payment over the limit", withOverrides(cash, expenseColumns{
			"cash_over_limit": "true", "creditable_iva_cents": "8000",
		}), sqlstateCheckViolation},
		{"no cash verdict", expenseColumns{"cash_over_limit": "NULL"}, sqlstateNotNullViolation},
		{"no creditable iva", expenseColumns{"creditable_iva_cents": "NULL"}, sqlstateNotNullViolation},
	}
	for _, c := range rejected {
		t.Run(c.name, func(t *testing.T) {
			_, err := execAsUser(t, app, userA, insertExpenseSQL(withOverrides(base, c.overrides)))
			requireSQLState(t, err, c.want)
		})
	}
	accepted := []struct {
		name      string
		overrides expenseColumns
	}{
		{"creditable iva in full", expenseColumns{"creditable_iva_cents": "8000"}},
		{"cash over the limit, nothing creditable", withOverrides(cash, expenseColumns{"cash_over_limit": "true"})},
		{"cash under the limit, creditable", withOverrides(cash, expenseColumns{"creditable_iva_cents": "8000"})},
	}
	for _, c := range accepted {
		t.Run("accepted: "+c.name, func(t *testing.T) {
			_, err := execAsUser(t, app, userA, insertExpenseSQL(withOverrides(base, c.overrides)))
			require.NoError(t, err)
		})
	}
	t.Run("a row written before the facts defaults to none", func(t *testing.T) {
		seedExpense(t, app, userB, nil)
		var cashOver bool
		var creditable int64
		require.NoError(t, pgtest.AsUser(t.Context(), app, userB, func(tx *sql.Tx) error {
			return tx.QueryRowContext(t.Context(), "SELECT cash_over_limit, creditable_iva_cents FROM expenses").
				Scan(&cashOver, &creditable)
		}))
		assert.False(t, cashOver)
		assert.Zero(t, creditable)
	})
}

func TestExpenseTaxFacts_AppWritesThem(t *testing.T) {
	app := pgtest.NewDatabase(t).Connect(t, pgtest.RoleApp)
	seedExpenseOwners(t, app)
	seedExpense(t, app, userA, expenseColumns{"cash_over_limit": "false", "creditable_iva_cents": "8000"})
	for _, query := range []string{
		"UPDATE expenses SET payment_method = 'cash', cash_over_limit = true, creditable_iva_cents = 0",
		"UPDATE expenses SET payment_method = 'transfer', cash_over_limit = false, creditable_iva_cents = 8000",
	} {
		affected, err := execAsUser(t, app, userA, query)
		require.NoError(t, err, query)
		assert.EqualValues(t, 1, affected, query)
	}
	affected, err := execAsUser(t, app, userB, "UPDATE expenses SET creditable_iva_cents = 0")
	require.NoError(t, err)
	assert.Zero(t, affected, "RLS: another user's facts match no row")
}

func TestExpenseTaxFactsMigration_DownRemovesTheColumns(t *testing.T) {
	owner := pgtest.NewDatabase(t).Connect(t, pgtest.RoleOwner)
	p, err := migrations.NewProvider(owner)
	require.NoError(t, err)

	_, err = p.DownTo(t.Context(), expensesTaxFactsVersion-1)
	require.NoError(t, err)
	var columns int
	require.NoError(t, owner.QueryRowContext(t.Context(), `
		SELECT count(*) FROM information_schema.columns
		WHERE table_name = 'expenses' AND column_name IN ('cash_over_limit', 'creditable_iva_cents')`).Scan(&columns))
	assert.Zero(t, columns)

	require.NoError(t, migrations.Up(t.Context(), owner))
	require.NoError(t, owner.QueryRowContext(t.Context(), `
		SELECT count(*) FROM information_schema.columns
		WHERE table_name = 'expenses' AND column_name IN ('cash_over_limit', 'creditable_iva_cents')`).Scan(&columns))
	assert.Equal(t, 2, columns)
}
