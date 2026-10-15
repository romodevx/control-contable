//go:build integration

package migrations_test

import (
	"database/sql"
	"maps"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/platform/pgtest"
	"github.com/romodevx/control-contable/backend/migrations"
)

// Withholdings expand (WH-03): docs/project-specs/MODULE-withholdings.md §2, §3.3, §5.1,
// ADR-0008 and docs/api/openapi.yaml 0.10.0.

const (
	withholdingsExpandVersion   = 20261016130000
	withholdingsValidateVersion = 20261016131000
)

// derived is a well-formed stored derivation (US-W1: RESICO PF bills a persona moral).
const derivedIncomeSQL = `UPDATE incomes SET withholding_rule_ids = '{ISR-4,IVA-6}',
  withholding_table_version = '2026.1', withholding_cpa_unverified = false`

// newWithholdingConstraints are the CHECKs the expand step adds NOT VALID and the validate
// step validates.
var newWithholdingConstraints = []string{
	"incomes_counterparty_kind_valid",
	"incomes_operation_kind_valid",
	"incomes_residential_rent_is_exempt",
	"incomes_withholding_rule_ids_valid",
	"incomes_withholding_table_version_format",
	"incomes_withholding_derived_together",
	"incomes_withholding_cpa_unverified_needs_derivation",
	"expenses_supplier_kind_valid",
	"expenses_supplier_regime_valid",
	"expenses_supplier_regime_iff_pf_with_activity",
	"expenses_operation_kind_valid",
	"expenses_residential_rent_is_exempt",
	"expenses_creditable_iva_withheld_within_withheld",
	"expenses_creditable_iva_split_within_iva",
	"expenses_creditable_iva_withheld_needs_cfdi_no_cash_over_limit",
	"expenses_withholding_rule_ids_valid",
	"expenses_withholding_table_version_format",
	"expenses_withholding_derived_together",
	"expenses_withholding_cpa_unverified_needs_derivation",
}

func TestWithholdingsIncomes_RejectInvalidFacts(t *testing.T) {
	app := seedIncomeOwners(t, pgtest.NewDatabase(t))
	require.NoError(t, insertIncome(t, app, userA, validIncome(userA, taxpayerA, incomeA)))
	rejected := []struct {
		name, query, want string
	}{
		{"unknown counterparty kind", "UPDATE incomes SET counterparty_kind = 'persona_moral'", sqlstateCheckViolation},
		{"platform is not an operation kind", "UPDATE incomes SET operation_kind = 'platform'", sqlstateCheckViolation},
		{"subcontracting was repealed", "UPDATE incomes SET operation_kind = 'subcontracting'", sqlstateCheckViolation},
		{"residential rent with iva", "UPDATE incomes SET residential_rent = true", sqlstateCheckViolation},
		{"no residential rent verdict", "UPDATE incomes SET residential_rent = NULL", sqlstateNotNullViolation},
		{"no platform verdict", "UPDATE incomes SET via_platform = NULL", sqlstateNotNullViolation},
		{"no rule ids", "UPDATE incomes SET withholding_rule_ids = NULL", sqlstateNotNullViolation},
		{"no cpa verdict", "UPDATE incomes SET withholding_cpa_unverified = NULL", sqlstateNotNullViolation},
	}
	for _, c := range rejected {
		t.Run(c.name, func(t *testing.T) {
			_, err := execAsUser(t, app, userA, c.query)
			requireSQLState(t, err, c.want)
		})
	}
	accepted := []string{
		`UPDATE incomes SET counterparty_kind = 'pm', operation_kind = 'professional_services', via_platform = false`,
		`UPDATE incomes SET counterparty_kind = 'gov_state_municipal', operation_kind = 'scrap'`,
		`UPDATE incomes SET operation_kind = 'rent_real_estate', residential_rent = true, iva_rate = 'exempt',
		   iva_cents = 0, iva_withheld_cents = 0`,
		`UPDATE incomes SET counterparty_kind = NULL, operation_kind = NULL, via_platform = true`,
	}
	for _, query := range accepted {
		t.Run("accepted: "+query, func(t *testing.T) {
			affected, err := execAsUser(t, app, userA, query)
			require.NoError(t, err)
			assert.EqualValues(t, 1, affected)
		})
	}
}

// The stored derivation (§3.3): rule ids are contract WithholdingRuleId values, at most 4;
// table_version follows the contract pattern; ids and version are set together, and the
// CPA flag only on a derived row. Shared by both tables.
func TestWithholdings_StoredDerivationIsWellFormed(t *testing.T) {
	db := pgtest.NewDatabase(t)
	app := seedIncomeOwners(t, db)
	require.NoError(t, insertIncome(t, app, userA, validIncome(userA, taxpayerA, incomeA)))
	seedExpense(t, app, userA, nil)
	rejected := []struct{ name, set string }{
		{"lower-case rule id", "withholding_rule_ids = '{isr-4}', withholding_table_version = '2026.1'"},
		{"unknown rule family", "withholding_rule_ids = '{IEPS-1}', withholding_table_version = '2026.1'"},
		{"null rule id", "withholding_rule_ids = ARRAY['ISR-4', NULL], withholding_table_version = '2026.1'"},
		{"five rule ids", "withholding_rule_ids = '{ISR-1,ISR-2,IVA-1,IVA-2,IVA-3}', withholding_table_version = '2026.1'"},
		{"two-dimensional ids", "withholding_rule_ids = '{{ISR-4},{IVA-6}}', withholding_table_version = '2026.1'"},
		{"malformed table version", "withholding_rule_ids = '{ISR-4,IVA-6}', withholding_table_version = 'v2026.1'"},
		{"rule ids without a table version", "withholding_rule_ids = '{ISR-4,IVA-6}'"},
		{"table version without rule ids", "withholding_table_version = '2026.1'"},
		{"cpa flag on an underived row", "withholding_cpa_unverified = true"},
	}
	for _, table := range []string{"incomes", "expenses"} {
		for _, c := range rejected {
			t.Run(table+": "+c.name, func(t *testing.T) {
				_, err := execAsUser(t, app, userA, "UPDATE "+table+" SET "+c.set)
				requireSQLState(t, err, sqlstateCheckViolation)
			})
		}
		t.Run(table+": accepted derivations", func(t *testing.T) {
			for _, set := range []string{
				"withholding_rule_ids = '{ISR-4,IVA-6}', withholding_table_version = '2026.1'",
				"withholding_rule_ids = '{ISR-0,IVA-NONE}', withholding_table_version = '2026.12', withholding_cpa_unverified = true",
				"withholding_rule_ids = '{ISR-10,IVA-12,ISR-2,IVA-4}', withholding_table_version = '2027.100'",
				"withholding_rule_ids = '{}', withholding_table_version = NULL, withholding_cpa_unverified = false",
			} {
				affected, err := execAsUser(t, app, userA, "UPDATE "+table+" SET "+set)
				require.NoError(t, err, set)
				assert.EqualValues(t, 1, affected, set)
			}
		})
	}
}

func TestWithholdingsExpenses_RejectInvalidFacts(t *testing.T) {
	app := pgtest.NewDatabase(t).Connect(t, pgtest.RoleApp)
	seedExpenseOwners(t, app)
	base := validExpense(userA, taxpayerA)
	withheld := withOverrides(base, expenseColumns{"isr_withheld_cents": "5000", "iva_withheld_cents": "5333"})
	rejected := []struct {
		name string
		cols expenseColumns
		want string
	}{
		{"unknown supplier kind", withOverrides(base, expenseColumns{"supplier_kind": "'pf_no_activity_or_public'"}), sqlstateCheckViolation},
		{"unknown supplier regime", withOverrides(base, expenseColumns{
			"supplier_kind": "'pf_with_activity'", "supplier_regime": "'601'",
		}), sqlstateCheckViolation},
		{"regime of a persona moral supplier", withOverrides(base, expenseColumns{
			"supplier_kind": "'pm'", "supplier_regime": "'612'",
		}), sqlstateCheckViolation},
		{"regime without a supplier kind", withOverrides(base, expenseColumns{"supplier_regime": "'612'"}), sqlstateCheckViolation},
		{"persona fisica supplier without a regime", withOverrides(base, expenseColumns{
			"supplier_kind": "'pf_with_activity'",
		}), sqlstateCheckViolation},
		{"unknown operation kind", withOverrides(base, expenseColumns{"operation_kind": "'platform'"}), sqlstateCheckViolation},
		{"residential rent with iva", withOverrides(base, expenseColumns{"residential_rent": "true"}), sqlstateCheckViolation},
		{"negative withheld creditable iva", withOverrides(withheld, expenseColumns{"creditable_iva_withheld_cents": "-1"}), sqlstateCheckViolation},
		{"withheld creditable iva above the iva withheld", withOverrides(withheld, expenseColumns{
			"creditable_iva_withheld_cents": "5334",
		}), sqlstateCheckViolation},
		{"withheld creditable iva without a withholding", withOverrides(base, expenseColumns{
			"creditable_iva_withheld_cents": "1",
		}), sqlstateCheckViolation},
		{"paid and withheld parts above the iva", withOverrides(withheld, expenseColumns{
			"creditable_iva_cents": "2668", "creditable_iva_withheld_cents": "5333",
		}), sqlstateCheckViolation},
		{"withheld creditable iva without a cfdi", withOverrides(withheld, expenseColumns{
			"has_cfdi": "false", "creditable_iva_withheld_cents": "5333",
		}), sqlstateCheckViolation},
		{"withheld creditable iva on a cash payment over the limit", withOverrides(withheld, expenseColumns{
			"payment_method": "'cash'", "deductible": "false", "cash_over_limit": "true",
			"creditable_iva_withheld_cents": "5333",
		}), sqlstateCheckViolation},
		{"no withheld creditable iva", withOverrides(base, expenseColumns{"creditable_iva_withheld_cents": "NULL"}), sqlstateNotNullViolation},
		{"no residential rent verdict", withOverrides(base, expenseColumns{"residential_rent": "NULL"}), sqlstateNotNullViolation},
	}
	for _, c := range rejected {
		t.Run(c.name, func(t *testing.T) {
			_, err := execAsUser(t, app, userA, insertExpenseSQL(c.cols))
			requireSQLState(t, err, c.want)
		})
	}
	accepted := []struct {
		name string
		cols expenseColumns
	}{
		{"persona fisica taxpayer: no supplier facts", base},
		{"persona moral supplier", withOverrides(base, expenseColumns{
			"supplier_kind": "'pm'", "operation_kind": "'ground_freight'",
		})},
		{"persona fisica supplier, regime unknown", withOverrides(withheld, expenseColumns{
			"supplier_kind": "'pf_with_activity'", "supplier_regime": "'unknown'", "operation_kind": "'professional_services'",
		})},
		{"foreign supplier", withOverrides(base, expenseColumns{"supplier_kind": "'foreign'"})},
		{"paid and withheld parts split the iva", withOverrides(withheld, expenseColumns{
			"supplier_kind": "'pf_with_activity'", "supplier_regime": "'626'",
			"creditable_iva_cents": "2667", "creditable_iva_withheld_cents": "5333",
		})},
		{"housing rent, exempt", withOverrides(base, expenseColumns{
			"operation_kind": "'rent_real_estate'", "residential_rent": "true", "iva_rate": "'exempt'", "iva_cents": "0",
		})},
	}
	for _, c := range accepted {
		t.Run("accepted: "+c.name, func(t *testing.T) {
			_, err := execAsUser(t, app, userA, insertExpenseSQL(c.cols))
			require.NoError(t, err)
		})
	}
}

// A row written the way the app writes today (no withholding facts) takes the neutral
// defaults: no facts, nothing derived, no withheld creditable part.
func TestWithholdings_RowsWithoutTheFactsTakeNeutralDefaults(t *testing.T) {
	db := pgtest.NewDatabase(t)
	app := seedIncomeOwners(t, db)
	require.NoError(t, insertIncome(t, app, userA, validIncome(userA, taxpayerA, incomeA)))
	seedExpense(t, app, userA, nil)

	var kind, operation, version sql.NullString
	var residential, platform, unverified bool
	var ruleIDs int
	queryAsUser(t, app, userA, `SELECT counterparty_kind, operation_kind, residential_rent, via_platform,
	  cardinality(withholding_rule_ids), withholding_table_version, withholding_cpa_unverified FROM incomes`,
		&kind, &operation, &residential, &platform, &ruleIDs, &version, &unverified)
	assert.False(t, kind.Valid || operation.Valid || version.Valid, "income facts start unset")
	assert.False(t, residential || platform || unverified)
	assert.Zero(t, ruleIDs)

	var supplier, regime sql.NullString
	var creditableWithheld int64
	queryAsUser(t, app, userA, `SELECT supplier_kind, supplier_regime, operation_kind, residential_rent,
	  creditable_iva_withheld_cents, cardinality(withholding_rule_ids), withholding_table_version,
	  withholding_cpa_unverified FROM expenses`,
		&supplier, &regime, &operation, &residential, &creditableWithheld, &ruleIDs, &version, &unverified)
	assert.False(t, supplier.Valid || regime.Valid || operation.Valid || version.Valid, "expense facts start unset")
	assert.False(t, residential || unverified)
	assert.Zero(t, creditableWithheld)
	assert.Zero(t, ruleIDs)
}

// The use cases write the facts and the derivation through cc_app (the cash_over_limit
// pattern); RLS keeps another user's rows out of reach, on the new columns too.
func TestWithholdings_AppWritesTheNewColumnsOnlyOnItsOwnRows(t *testing.T) {
	app := seedIncomeOwners(t, pgtest.NewDatabase(t))
	for _, owner := range []string{userA, userB} {
		require.NoError(t, insertIncome(t, app, owner, validIncome(owner, taxpayerOf[owner], incomeOf[owner])))
		seedExpense(t, app, owner, expenseColumns{"isr_withheld_cents": "5000", "iva_withheld_cents": "5333"})
	}
	writes := []string{
		`UPDATE incomes SET counterparty_kind = 'pm', operation_kind = 'professional_services',
		   residential_rent = false, via_platform = false`,
		derivedIncomeSQL,
		`UPDATE expenses SET supplier_kind = 'pf_with_activity', supplier_regime = '612',
		   operation_kind = 'professional_services', residential_rent = false,
		   creditable_iva_cents = 2667, creditable_iva_withheld_cents = 5333,
		   withholding_rule_ids = '{ISR-5,IVA-6}', withholding_table_version = '2026.1',
		   withholding_cpa_unverified = false`,
	}
	for _, query := range writes {
		affected, err := execAsUser(t, app, userA, query+" WHERE user_id = $1", userB)
		require.NoError(t, err, query)
		assert.Zero(t, affected, "RLS: another user's row matches no row: %s", query)

		affected, err = execAsUser(t, app, userA, query)
		require.NoError(t, err, query)
		assert.EqualValues(t, 1, affected, query)
	}

	t.Run("inserts name the new columns", func(t *testing.T) {
		_, err := execAsUser(t, app, userA, insertExpenseSQL(withOverrides(validExpense(userA, taxpayerA), expenseColumns{
			"supplier_kind": "'pm'", "operation_kind": "'scrap'", "residential_rent": "false",
			"creditable_iva_withheld_cents": "0", "withholding_rule_ids": "'{ISR-2,IVA-5}'",
			"withholding_table_version": "'2026.1'", "withholding_cpa_unverified": "false",
		})))
		require.NoError(t, err)
	})

	var unchanged int
	queryAsUser(t, app, userB, `SELECT count(*) FROM incomes WHERE counterparty_kind IS NULL
	  AND withholding_table_version IS NULL`, &unchanged)
	assert.Equal(t, 1, unchanged, "userB's income kept its facts")
}

// MODULE-withholdings §2.2: every persona moral rule row carries the E3 default of its
// category; persona física rows carry none (the section does not exist for them).
func TestWithholdings_CategoryDefaultOperationKind(t *testing.T) {
	app := pgtest.NewDatabase(t).Connect(t, pgtest.RoleApp)

	var moralMissing, fisicaWithDefault int
	require.NoError(t, app.QueryRowContext(t.Context(), `
		SELECT count(*) FILTER (WHERE person_type = 'moral' AND default_operation_kind IS NULL),
		       count(*) FILTER (WHERE person_type = 'fisica' AND default_operation_kind IS NOT NULL)
		FROM category_regime_rules`).Scan(&moralMissing, &fisicaWithDefault))
	assert.Zero(t, moralMissing)
	assert.Zero(t, fisicaWithDefault)

	want := map[string]string{
		"servicios_profesionales": "professional_services",
		"renta_local":             "rent_real_estate",
		"fletes":                  "ground_freight",
		"comisiones":              "commission",
		"mercancia":               "sale_of_goods",
		"software":                "business_services",
		"otros":                   "business_services",
	}
	for category, kind := range want {
		kinds := queryStrings(t, app, `SELECT DISTINCT default_operation_kind FROM category_regime_rules
			WHERE person_type = 'moral' AND category_code = $1`, category)
		assert.Equal(t, []string{kind}, kinds, category)
	}

	t.Run("the catalog stays read-only for cc_app", func(t *testing.T) {
		_, err := app.ExecContext(t.Context(), "UPDATE category_regime_rules SET default_operation_kind = 'other'")
		requireSQLState(t, err, sqlstateInsufficientPrivilege)
	})
}

func TestWithholdings_CategoryDefaultOperationKindRejectsMalformedRows(t *testing.T) {
	owner := pgtest.NewDatabase(t).Connect(t, pgtest.RoleOwner)
	_, err := owner.Exec("DELETE FROM category_regime_rules WHERE category_code = 'software' AND regime_code IN ('612', '601')")
	require.NoError(t, err)
	moral := ruleColumns{"regime_code": "'601'", "person_type": "'moral'"}
	cases := []struct {
		name      string
		overrides ruleColumns
	}{
		{"persona moral row without a default", moral},
		{"persona moral row with an unknown kind", withRule(moral, ruleColumns{"default_operation_kind": "'subcontracting'"})},
		{"persona fisica row with a default", ruleColumns{"default_operation_kind": "'business_services'"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			requireSQLState(t, insertRule(owner, c.overrides), sqlstateCheckViolation)
		})
	}
	t.Run("well-formed rows are accepted", func(t *testing.T) {
		require.NoError(t, insertRule(owner, nil))
		require.NoError(t, insertRule(owner, withRule(moral, ruleColumns{"default_operation_kind": "'business_services'"})))
	})
}

func withRule(base, overrides ruleColumns) ruleColumns {
	out := maps.Clone(base)
	maps.Copy(out, overrides)
	return out
}

// The expand step adds every CHECK NOT VALID and the validate step validates them in its
// own transaction, over rows that existed before the migration.
func TestWithholdingsMigration_ValidatesOverExistingRows(t *testing.T) {
	db := pgtest.NewDatabase(t)
	owner := db.Connect(t, pgtest.RoleOwner)
	p, err := migrations.NewProvider(owner)
	require.NoError(t, err)
	_, err = p.DownTo(t.Context(), withholdingsExpandVersion-1)
	require.NoError(t, err)

	app := seedIncomeOwners(t, db)
	require.NoError(t, insertIncome(t, app, userA, validIncome(userA, taxpayerA, incomeA)))
	seedExpense(t, app, userA, expenseColumns{
		"isr_withheld_cents": "5000", "iva_withheld_cents": "5333", "creditable_iva_cents": "8000",
	})

	_, err = p.UpTo(t.Context(), withholdingsExpandVersion)
	require.NoError(t, err)
	assert.Equal(t, newWithholdingConstraints, constraintsValidated(t, owner, false), "added NOT VALID")

	_, err = p.UpTo(t.Context(), withholdingsValidateVersion)
	require.NoError(t, err)
	assert.Equal(t, newWithholdingConstraints, constraintsValidated(t, owner, true), "validated")

	// FORCE RLS: the owner sees no row, so the rows are read in their user's context.
	var legacyIncomes, legacyExpenses int
	queryAsUser(t, app, userA, `SELECT
	  (SELECT count(*) FROM incomes WHERE counterparty_kind IS NULL AND NOT residential_rent AND NOT via_platform
	     AND withholding_rule_ids = '{}' AND isr_withheld_cents = 100000),
	  (SELECT count(*) FROM expenses WHERE supplier_kind IS NULL AND creditable_iva_withheld_cents = 0
	     AND creditable_iva_cents = 8000 AND iva_withheld_cents = 5333)`, &legacyIncomes, &legacyExpenses)
	assert.Equal(t, 1, legacyIncomes, "the typed withholdings are kept")
	assert.Equal(t, 1, legacyExpenses, "the full creditable candidate is kept until the re-split")
}

// constraintsValidated lists, in newWithholdingConstraints order, the new constraints whose
// convalidated flag equals validated.
func constraintsValidated(t *testing.T, owner *sql.DB, validated bool) []string {
	t.Helper()
	found := map[string]bool{}
	for _, name := range queryStrings(t, owner, `SELECT conname FROM pg_catalog.pg_constraint
		WHERE conrelid IN ('public.incomes'::regclass, 'public.expenses'::regclass)
		  AND contype = 'c' AND convalidated = $1`, validated) {
		found[name] = true
	}
	var out []string
	for _, name := range newWithholdingConstraints {
		if found[name] {
			out = append(out, name)
		}
	}
	return out
}

func TestWithholdingsMigration_DownRemovesEverything(t *testing.T) {
	owner := pgtest.NewDatabase(t).Connect(t, pgtest.RoleOwner)
	p, err := migrations.NewProvider(owner)
	require.NoError(t, err)

	const leftovers = `SELECT
	  (SELECT count(*) FROM information_schema.columns
	   WHERE table_schema = 'public' AND (
	     (table_name = 'incomes' AND column_name IN ('counterparty_kind', 'operation_kind', 'residential_rent',
	       'via_platform', 'withholding_rule_ids', 'withholding_table_version', 'withholding_cpa_unverified'))
	     OR (table_name = 'expenses' AND column_name IN ('supplier_kind', 'supplier_regime', 'operation_kind',
	       'residential_rent', 'creditable_iva_withheld_cents', 'withholding_rule_ids',
	       'withholding_table_version', 'withholding_cpa_unverified'))
	     OR (table_name = 'category_regime_rules' AND column_name = 'default_operation_kind')))
	  + (SELECT count(*) FROM pg_catalog.pg_proc WHERE proname = 'withholding_rule_ids_valid')`
	const created = 7 + 8 + 1 + 1

	_, err = p.DownTo(t.Context(), withholdingsValidateVersion-1)
	require.NoError(t, err)
	assert.Equal(t, newWithholdingConstraints, constraintsValidated(t, owner, true),
		"undoing the validation keeps the (stronger) validated constraints")

	_, err = p.DownTo(t.Context(), withholdingsExpandVersion-1)
	require.NoError(t, err)
	var count int
	require.NoError(t, owner.QueryRowContext(t.Context(), leftovers).Scan(&count))
	assert.Zero(t, count)
	var typed int
	require.NoError(t, owner.QueryRowContext(t.Context(), `SELECT count(*) FROM information_schema.columns
	  WHERE table_name IN ('incomes', 'expenses') AND column_name IN ('isr_withheld_cents', 'iva_withheld_cents')`).Scan(&typed))
	assert.Equal(t, 4, typed, "the typed withholding columns belong to earlier versions")

	require.NoError(t, migrations.Up(t.Context(), owner))
	require.NoError(t, migrations.Up(t.Context(), owner), "a second up is a no-op")
	require.NoError(t, owner.QueryRowContext(t.Context(), leftovers).Scan(&count))
	assert.Equal(t, created, count)
}
