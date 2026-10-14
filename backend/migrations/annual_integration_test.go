//go:build integration

package migrations_test

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/platform/pgtest"
	"github.com/romodevx/control-contable/backend/migrations"
)

// Annual module migrations (ANN-03): the annual_depreciation tax input and the four
// annual-only input tables, docs/project-specs/MODULE-annual.md §2.6, §5 and §9.

const (
	annualDepreciationVersion = 20261016120000
	annualTablesVersion       = 20261016121000

	// maxCents is docs/api/openapi.yaml NonNegativeCents.maximum.
	maxCents = "99999999999999"
	// overMaxCents is one centavo above it.
	overMaxCents = "100000000000000"

	// annualPlanFirstYear and annualPlanLastYear bound the fiscal years the tables admit.
	annualPlanFirstYear = 2026
	annualPlanLastYear  = 2100
	// planDeductionsPerUser personal deductions per plan user, over five fiscal years.
	planDeductionsPerUser = 1000

	// sqlstateGeneratedAlways: a write names a generated column (refused before privileges).
	sqlstateGeneratedAlways = "428C9"
)

// annualTable is one of the four tables and how a valid row of it is written.
type annualTable struct {
	name  string
	valid func(userID, taxpayerID string) expenseColumns
}

var (
	personalDeductions = annualTable{"personal_deductions", func(userID, taxpayerID string) expenseColumns {
		return expenseColumns{
			"user_id": quote(userID), "taxpayer_id": quote(taxpayerID), "category": "'medical'",
			"amount_cents": "4000000", "paid_on": "'2026-05-14'", "payment_method": "'electronic'",
			"beneficiary": "'self'",
		}
	}}
	salaryStatements = annualTable{"salary_statements", func(userID, taxpayerID string) expenseColumns {
		return expenseColumns{
			"user_id": quote(userID), "taxpayer_id": quote(taxpayerID), "fiscal_year": "2026",
			"accumulable_cents": "36000000", "exempt_cents": "2000000", "isr_withheld_cents": "5423591",
		}
	}}
	filedPayments = annualTable{"annual_filed_payments", func(userID, taxpayerID string) expenseColumns {
		return expenseColumns{
			"user_id": quote(userID), "taxpayer_id": quote(taxpayerID), "regime_code": "'612'",
			"fiscal_year": "2026", "month": "1", "amount_cents": "557386",
		}
	}}
	annualInputs = annualTable{"annual_inputs", func(userID, taxpayerID string) expenseColumns {
		return expenseColumns{
			"user_id": quote(userID), "taxpayer_id": quote(taxpayerID), "fiscal_year": "2026",
			"prior_year_accumulable_income_cents": "50000000",
		}
	}}

	allAnnualTables = []annualTable{personalDeductions, salaryStatements, filedPayments, annualInputs}
)

func (a annualTable) table() pgtest.Table { return pgtest.Table{Name: a.name, OwnerColumn: "user_id"} }

func (a annualTable) insertSQL(cols expenseColumns) string {
	names, values := splitColumns(cols)
	return fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", a.name, names, values)
}

// row is the valid row of userID's own taxpayer with overrides ("" removes a column).
func (a annualTable) row(userID string, overrides expenseColumns) expenseColumns {
	return withOverrides(a.valid(userID, taxpayerOf[userID]), overrides)
}

func (a annualTable) insertOwn(ctx context.Context, tx *sql.Tx, ownerID string) error {
	_, err := tx.ExecContext(ctx, a.insertSQL(a.row(ownerID, nil)))
	return err
}

func (a annualTable) insert(t *testing.T, app *sql.DB, asUser string, cols expenseColumns) error {
	t.Helper()
	_, err := execAsUser(t, app, asUser, a.insertSQL(cols))
	return err
}

func (a annualTable) count(t *testing.T, app *sql.DB, userID string) int {
	t.Helper()
	var n int
	queryAsUser(t, app, userID, "SELECT count(*) FROM "+a.name, &n)
	return n
}

// seedAnnualOwners: users A and B, each with a física taxpayer holding 612 since 2022-01,
// plus a moral taxpayer of A holding 601 (seedTaxInputOwners).
func seedAnnualOwners(t *testing.T, db *pgtest.Database) *sql.DB {
	t.Helper()
	return seedTaxInputOwners(t, db)
}

// §9: RLS ENABLE + FORCE on all four tables, one owner policy per operation the app runs.
func TestAnnualTables_AppRoleIsolation(t *testing.T) {
	for _, tbl := range allAnnualTables {
		t.Run(tbl.name, func(t *testing.T) {
			app := seedAnnualOwners(t, pgtest.NewDatabase(t))
			pgtest.AssertIsolation(t, app, pgtest.IsolationSpec{
				Table: tbl.table(), UserA: userA, UserB: userB, Insert: tbl.insertOwn,
			})
		})
	}
}

// AssertIsolation's generic UPDATE rewrites user_id, which cc_app may not update at all;
// prove the writes the use cases run match no foreign row, and do match an own row.
func TestAnnualTables_CrossUserWritesMatchNoRow(t *testing.T) {
	writes := map[string][]string{
		personalDeductions.name: {
			"UPDATE personal_deductions SET amount_cents = 1 WHERE taxpayer_id = $1",
			"UPDATE personal_deductions SET deleted_at = now() WHERE taxpayer_id = $1",
		},
		salaryStatements.name: {
			"UPDATE salary_statements SET isr_withheld_cents = 1 WHERE taxpayer_id = $1",
			"DELETE FROM salary_statements WHERE taxpayer_id = $1",
		},
		filedPayments.name: {
			"UPDATE annual_filed_payments SET amount_cents = 1 WHERE taxpayer_id = $1",
			"DELETE FROM annual_filed_payments WHERE taxpayer_id = $1 AND regime_code = '612' AND fiscal_year = 2026",
		},
		annualInputs.name: {
			"UPDATE annual_inputs SET told_employer_will_file = true WHERE taxpayer_id = $1",
		},
	}
	for _, tbl := range allAnnualTables {
		t.Run(tbl.name, func(t *testing.T) {
			app := seedAnnualOwners(t, pgtest.NewDatabase(t))
			require.NoError(t, tbl.insert(t, app, userB, tbl.row(userB, nil)))
			for _, query := range writes[tbl.name] {
				affected, err := execAsUser(t, app, userA, query, taxpayerB)
				require.NoError(t, err, query)
				assert.Zero(t, affected, query)
			}
			for _, query := range writes[tbl.name] {
				affected, err := execAsUser(t, app, userB, query, taxpayerB)
				require.NoError(t, err, query)
				assert.EqualValues(t, 1, affected, "own row: "+query)
			}
		})
	}
}

// Composite FKs bypass RLS, so they (not the policies) prove a row belongs to its own
// user's taxpayer: forged ids never land. Deductions and constancias are personas
// físicas only (§5 person_type_not_allowed), filed payments need a regime the taxpayer holds.
func TestAnnualTables_CompositeForeignKeysRefuseForgedReferences(t *testing.T) {
	app := seedAnnualOwners(t, pgtest.NewDatabase(t))
	forge := func(tbl annualTable, userID, taxpayerID string, extra expenseColumns) expenseColumns {
		cols := tbl.valid(userID, taxpayerID)
		return withOverrides(cols, extra)
	}
	type forged struct {
		name string
		cols expenseColumns
		want string
	}
	for _, tbl := range allAnnualTables {
		cases := []forged{
			{"own user id, another user's taxpayer", forge(tbl, userA, taxpayerB, nil), sqlstateForeignKeyViolation},
			{"another user's id and taxpayer", forge(tbl, userB, taxpayerB, nil), sqlstateInsufficientPrivilege},
			{"another user's id, own taxpayer", forge(tbl, userB, taxpayerA, nil), sqlstateInsufficientPrivilege},
			{"unknown taxpayer", forge(tbl, userA, "00000000-0000-7000-8000-0000000000ff", nil), sqlstateForeignKeyViolation},
		}
		switch tbl.name {
		case personalDeductions.name, salaryStatements.name:
			cases = append(cases, forged{"a persona moral", forge(tbl, userA, taxInputsMoralTaxpayer, nil),
				sqlstateForeignKeyViolation})
		case filedPayments.name:
			cases = append(cases,
				forged{"a regime the taxpayer does not hold", forge(tbl, userA, taxpayerA,
					expenseColumns{"regime_code": "'606'"}), sqlstateForeignKeyViolation},
				forged{"an unknown regime", forge(tbl, userA, taxpayerA,
					expenseColumns{"regime_code": "'999'"}), sqlstateForeignKeyViolation})
		}
		for _, c := range cases {
			t.Run(tbl.name+": "+c.name, func(t *testing.T) {
				requireSQLState(t, tbl.insert(t, app, userA, c.cols), c.want)
			})
		}
		t.Run(tbl.name+": the well-formed reference is accepted", func(t *testing.T) {
			require.NoError(t, tbl.insert(t, app, userA, tbl.row(userA, nil)))
		})
	}
	t.Run("annual inputs and filed payments accept a persona moral", func(t *testing.T) {
		require.NoError(t, annualInputs.insert(t, app, userA, annualInputs.valid(userA, taxInputsMoralTaxpayer)))
		require.NoError(t, filedPayments.insert(t, app, userA, withOverrides(
			filedPayments.valid(userA, taxInputsMoralTaxpayer), expenseColumns{"regime_code": "'601'"})))
	})
	t.Run("cc_app cannot write the person type", func(t *testing.T) {
		for _, tbl := range []annualTable{personalDeductions, salaryStatements} {
			err := tbl.insert(t, app, userA, tbl.row(userA, expenseColumns{"person_type": "'moral'"}))
			requireSQLState(t, err, sqlstateInsufficientPrivilege)
		}
	})
}

// invalidCase is one row that a CHECK constraint must refuse.
type invalidCase struct {
	name      string
	overrides expenseColumns
}

func requireChecksRefuse(t *testing.T, app *sql.DB, tbl annualTable, cases []invalidCase) {
	t.Helper()
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			requireSQLState(t, tbl.insert(t, app, userA, tbl.row(userA, c.overrides)), sqlstateCheckViolation)
		})
	}
}

func requireAccepted(t *testing.T, app *sql.DB, tbl annualTable, accepted []expenseColumns) {
	t.Helper()
	for _, overrides := range accepted {
		require.NoError(t, tbl.insert(t, app, userA, tbl.row(userA, overrides)), "%v", overrides)
	}
}

func tuition(level string) expenseColumns {
	return expenseColumns{
		"category": "'tuition'", "student_label": "'Sofía'", "tuition_level": quote(level), "scholarship_cents": "0",
	}
}

// §2.6 personal_deductions fields, §5 validation_failed backstops.
func TestPersonalDeductions_RejectInvalidRows(t *testing.T) {
	app := seedAnnualOwners(t, pgtest.NewDatabase(t))
	donation := expenseColumns{"category": "'donation'", "donee_type": "'authorized'"}
	withTuition := func(extra expenseColumns) expenseColumns {
		return withOverrides(tuition("primaria"), extra)
	}
	requireChecksRefuse(t, app, personalDeductions, []invalidCase{
		{"unknown category", expenseColumns{"category": "'gym'"}},
		{"zero amount", expenseColumns{"amount_cents": "0"}},
		{"negative amount", expenseColumns{"amount_cents": "-1"}},
		{"amount above NonNegativeCents.maximum", expenseColumns{"amount_cents": overMaxCents}},
		{"paid before 2026", expenseColumns{"paid_on": "'2025-12-31'"}},
		{"paid after FiscalYear.maximum", expenseColumns{"paid_on": "'2101-01-01'"}},
		{"unknown payment method", expenseColumns{"payment_method": "'transfer'"}},
		{"unknown beneficiary", expenseColumns{"beneficiary": "'friend'"}},
		{"donation without donee type", expenseColumns{"category": "'donation'"}},
		{"donee type on a medical expense", expenseColumns{"donee_type": "'authorized'"}},
		{"unknown donee type", withOverrides(donation, expenseColumns{"donee_type": "'church'"})},
		{"tuition without level", withTuition(expenseColumns{"tuition_level": ""})},
		{"tuition without student", withTuition(expenseColumns{"student_label": ""})},
		{"tuition without scholarship", withTuition(expenseColumns{"scholarship_cents": ""})},
		{"unknown tuition level", withTuition(expenseColumns{"tuition_level": "'universidad'"})},
		{"negative scholarship", withTuition(expenseColumns{"scholarship_cents": "-1"})},
		{"scholarship above NonNegativeCents.maximum", withTuition(expenseColumns{"scholarship_cents": overMaxCents})},
		{"empty student label", withTuition(expenseColumns{"student_label": "''"})},
		{"untrimmed student label", withTuition(expenseColumns{"student_label": "' Sofía'"})},
		{"student label of 41 characters", withTuition(expenseColumns{"student_label": quote(strings.Repeat("a", 41))})},
		{"student label not NFC", withTuition(expenseColumns{"student_label": "U&'Sofi\\0301a'"})},
		{"tuition level on a medical expense", expenseColumns{"tuition_level": "'primaria'"}},
		{"student label on a medical expense", expenseColumns{"student_label": "'Sofía'"}},
		{"scholarship on a medical expense", expenseColumns{"scholarship_cents": "0"}},
		{"lower-case CFDI UUID", expenseColumns{"cfdi_uuid": "'5b0c2e6a-1f4d-4c7e-9a3b-2d8f6e1c4a90'"}},
		{"malformed CFDI UUID", expenseColumns{"cfdi_uuid": "'5B0C2E6A1F4D4C7E9A3B2D8F6E1C4A90'"}},
	})

	t.Run("the boundary values are accepted", func(t *testing.T) {
		accepted := []expenseColumns{
			{"amount_cents": "1", "paid_on": "'2026-01-01'"},
			{"amount_cents": maxCents, "paid_on": "'2100-12-31'"},
			{"payment_method": "'cash'", "beneficiary": "'descendant'"},
			{"beneficiary": "'spouse'"},
			{"beneficiary": "'ascendant'"},
			{"cfdi_uuid": "'5B0C2E6A-1F4D-4C7E-9A3B-2D8F6E1C4A90'"},
			donation,
			{"category": "'donation'", "donee_type": "'government'"},
			withTuition(expenseColumns{"student_label": quote(strings.Repeat("a", 40)), "scholarship_cents": maxCents}),
		}
		for _, category := range []string{
			"medical", "medical_disability", "funeral", "mortgage_real_interest", "retirement_contribution",
			"health_insurance", "school_transport", "local_payroll_tax",
		} {
			accepted = append(accepted, expenseColumns{"category": quote(category)})
		}
		for _, level := range []string{"preescolar", "primaria", "secundaria", "profesional_tecnico", "bachillerato"} {
			accepted = append(accepted, tuition(level))
		}
		requireAccepted(t, app, personalDeductions, accepted)
	})
}

// The fiscal year of a deduction is the year it was paid (§2.6 paid_on "date in the
// fiscal year"): derived by the database, never written by the app.
func TestPersonalDeductions_FiscalYearIsTheYearPaid(t *testing.T) {
	app := seedAnnualOwners(t, pgtest.NewDatabase(t))
	require.NoError(t, personalDeductions.insert(t, app, userA,
		personalDeductions.row(userA, expenseColumns{"paid_on": "'2027-12-31'"})))
	var year int
	queryAsUser(t, app, userA, "SELECT fiscal_year FROM personal_deductions", &year)
	assert.Equal(t, 2027, year)

	_, err := execAsUser(t, app, userA, "UPDATE personal_deductions SET paid_on = '2026-01-01'")
	require.NoError(t, err)
	queryAsUser(t, app, userA, "SELECT fiscal_year FROM personal_deductions", &year)
	assert.Equal(t, 2026, year, "recomputed on edit")

	for _, stmt := range []string{
		personalDeductions.insertSQL(personalDeductions.row(userA, expenseColumns{"fiscal_year": "2026"})),
		"UPDATE personal_deductions SET fiscal_year = 2026",
	} {
		_, err := execAsUser(t, app, userA, stmt)
		requireSQLState(t, err, sqlstateGeneratedAlways)
	}
}

// §2.6 salary_statements: constancia totals, no employer RFC (§9 data minimization).
func TestSalaryStatements_RejectInvalidRows(t *testing.T) {
	app := seedAnnualOwners(t, pgtest.NewDatabase(t))
	requireChecksRefuse(t, app, salaryStatements, []invalidCase{
		{"fiscal year before 2026", expenseColumns{"fiscal_year": "2025"}},
		{"fiscal year after FiscalYear.maximum", expenseColumns{"fiscal_year": "2101"}},
		{"negative accumulable", expenseColumns{"accumulable_cents": "-1"}},
		{"negative exempt", expenseColumns{"exempt_cents": "-1"}},
		{"negative ISR withheld", expenseColumns{"isr_withheld_cents": "-1"}},
		{"accumulable above maximum", expenseColumns{"accumulable_cents": overMaxCents}},
		{"exempt above maximum", expenseColumns{"exempt_cents": overMaxCents}},
		{"ISR withheld above maximum", expenseColumns{"isr_withheld_cents": overMaxCents}},
		{"empty employer label", expenseColumns{"employer_label": "''"}},
		{"untrimmed employer label", expenseColumns{"employer_label": "'Acme '"}},
		{"employer label of 61 characters", expenseColumns{"employer_label": quote(strings.Repeat("a", 61))}},
		{"employer label not NFC", expenseColumns{"employer_label": "U&'Cafe\\0301'"}},
	})
	t.Run("the boundary values are accepted", func(t *testing.T) {
		requireAccepted(t, app, salaryStatements, []expenseColumns{
			{"accumulable_cents": "0", "exempt_cents": "0", "isr_withheld_cents": "0", "fiscal_year": "2100"},
			{"accumulable_cents": maxCents, "exempt_cents": maxCents, "isr_withheld_cents": maxCents},
			{"employer_label": quote(strings.Repeat("a", 60)), "simultaneous": "true",
				"left_before_year_end": "true", "payer_not_withholding": "true"},
			{},
		})
	})
	t.Run("the flags default to false", func(t *testing.T) {
		var anyFlag bool
		queryAsUser(t, app, userA, `SELECT bool_or(simultaneous OR left_before_year_end OR payer_not_withholding)
			FROM salary_statements WHERE employer_label IS NULL`, &anyFlag)
		assert.False(t, anyFlag)
	})
	t.Run("there is no employer RFC column", func(t *testing.T) {
		owner := pgtest.NewDatabase(t).Connect(t, pgtest.RoleOwner)
		assert.Empty(t, queryStrings(t, owner, `SELECT column_name::text FROM information_schema.columns
			WHERE table_name = 'salary_statements' AND column_name ~ 'rfc'`))
	})
}

// §2.6 annual_filed_payments: (taxpayer, regime, fiscal year, month) unique, amount >= 0.
func TestAnnualFiledPayments_RejectInvalidRows(t *testing.T) {
	app := seedAnnualOwners(t, pgtest.NewDatabase(t))
	requireChecksRefuse(t, app, filedPayments, []invalidCase{
		{"month 0", expenseColumns{"month": "0"}},
		{"month 13", expenseColumns{"month": "13"}},
		{"negative amount", expenseColumns{"amount_cents": "-1"}},
		{"amount above maximum", expenseColumns{"amount_cents": overMaxCents}},
		{"fiscal year before 2026", expenseColumns{"fiscal_year": "2025"}},
		{"fiscal year after FiscalYear.maximum", expenseColumns{"fiscal_year": "2101"}},
	})
	t.Run("the boundary values are accepted", func(t *testing.T) {
		requireAccepted(t, app, filedPayments, []expenseColumns{
			{"month": "12", "amount_cents": "0"},
			{"month": "1", "amount_cents": maxCents, "fiscal_year": "2100"},
		})
	})
	t.Run("one amount per regime, fiscal year and month", func(t *testing.T) {
		requireSQLState(t, filedPayments.insert(t, app, userA, filedPayments.row(userA, expenseColumns{"month": "12"})),
			sqlstateUniqueViolation)
	})
}

// §2.6 annual_inputs: one row per taxpayer and fiscal year.
func TestAnnualInputs_RejectInvalidRows(t *testing.T) {
	app := seedAnnualOwners(t, pgtest.NewDatabase(t))
	requireChecksRefuse(t, app, annualInputs, []invalidCase{
		{"negative prior-year income", expenseColumns{"prior_year_accumulable_income_cents": "-1"}},
		{"prior-year income above maximum", expenseColumns{"prior_year_accumulable_income_cents": overMaxCents}},
		{"negative other exempt income", expenseColumns{"other_exempt_income_cents": "-1"}},
		{"other exempt income above maximum", expenseColumns{"other_exempt_income_cents": overMaxCents}},
		{"fiscal year before 2026", expenseColumns{"fiscal_year": "2025"}},
		{"fiscal year after FiscalYear.maximum", expenseColumns{"fiscal_year": "2101"}},
	})
	t.Run("the defaults and boundary values are accepted", func(t *testing.T) {
		requireAccepted(t, app, annualInputs, []expenseColumns{
			{"prior_year_accumulable_income_cents": ""},
			{"fiscal_year": "2027", "prior_year_accumulable_income_cents": maxCents,
				"other_exempt_income_cents": maxCents, "told_employer_will_file": "true"},
			{"fiscal_year": "2100", "prior_year_accumulable_income_cents": "0"},
		})
		var prior sql.NullInt64
		var exempt int64
		var told bool
		queryAsUser(t, app, userA, `SELECT prior_year_accumulable_income_cents, other_exempt_income_cents,
			told_employer_will_file FROM annual_inputs WHERE fiscal_year = 2026`, &prior, &exempt, &told)
		assert.False(t, prior.Valid, "unknown until typed (OQ-A5)")
		assert.Zero(t, exempt)
		assert.False(t, told)
	})
	t.Run("one row per taxpayer and fiscal year", func(t *testing.T) {
		requireSQLState(t, annualInputs.insert(t, app, userA, annualInputs.row(userA, nil)), sqlstateUniqueViolation)
	})
}

// §2.6 / §6: account and taxpayer deletion cascade to every annual row, soft-deleted
// deductions included; another user's rows stay.
func TestAnnualTables_DeletionCascades(t *testing.T) {
	setup := func(t *testing.T) *sql.DB {
		t.Helper()
		app := seedAnnualOwners(t, pgtest.NewDatabase(t))
		for _, tbl := range allAnnualTables {
			for _, user := range []string{userA, userB} {
				require.NoError(t, tbl.insert(t, app, user, tbl.row(user, nil)))
			}
		}
		_, err := execAsUser(t, app, userA, "UPDATE personal_deductions SET deleted_at = now()")
		require.NoError(t, err)
		return app
	}
	deletions := map[string]string{
		"account deletion":  "DELETE FROM users WHERE id = $1",
		"taxpayer deletion": "DELETE FROM taxpayers WHERE user_id = $1",
	}
	for name, stmt := range deletions {
		t.Run(name, func(t *testing.T) {
			app := setup(t)
			affected, err := execAsUser(t, app, userA, stmt, userA)
			require.NoError(t, err)
			require.Positive(t, affected)
			for _, tbl := range allAnnualTables {
				assert.Zero(t, tbl.count(t, app, userA), tbl.name)
				assert.Equal(t, 1, tbl.count(t, app, userB), "another user's %s stay", tbl.name)
			}
		})
	}
	t.Run("the users-FK guard sees every annual table directly and through its chain", func(t *testing.T) {
		owner := pgtest.NewDatabase(t).Connect(t, pgtest.RoleOwner)
		direct := queryStrings(t, owner,
			"SELECT conrelid::regclass::text FROM pg_constraint WHERE contype = 'f' AND confrelid = 'users'::regclass AND confdeltype = 'c'")
		chained := queryStrings(t, owner, `SELECT conrelid::regclass::text FROM pg_constraint
			WHERE contype = 'f' AND confrelid IN ('taxpayers'::regclass, 'taxpayer_regimes'::regclass) AND confdeltype = 'c'`)
		for _, tbl := range allAnnualTables {
			assert.Contains(t, direct, tbl.name)
			assert.Contains(t, chained, tbl.name)
		}
	})
}

// Least privilege: cc_app writes only the columns the use cases need; the identity and
// audit columns never change; deductions are soft-deleted (no DELETE), annual inputs
// are upserted (no DELETE). cc_webhook and cc_worker get nothing.
func TestAnnualTables_NoAccessBeyondTheGrants(t *testing.T) {
	db := pgtest.NewDatabase(t)
	app := seedAnnualOwners(t, db)
	for _, tbl := range allAnnualTables {
		require.NoError(t, tbl.insert(t, app, userA, tbl.row(userA, nil)))
	}
	var denied []struct {
		role  pgtest.Role
		query string
	}
	deny := func(role pgtest.Role, query string) {
		denied = append(denied, struct {
			role  pgtest.Role
			query string
		}{role, query})
	}
	immutable := map[string][]string{
		personalDeductions.name: {"id", "user_id", "taxpayer_id", "person_type", "created_at", "updated_at"},
		salaryStatements.name: {"id", "user_id", "taxpayer_id", "person_type", "fiscal_year", "created_at",
			"updated_at"},
		filedPayments.name: {"user_id", "taxpayer_id", "regime_code", "fiscal_year", "month", "created_at", "updated_at"},
		annualInputs.name:  {"user_id", "taxpayer_id", "fiscal_year", "created_at", "updated_at"},
	}
	for _, tbl := range allAnnualTables {
		for _, col := range immutable[tbl.name] {
			deny(pgtest.RoleApp, fmt.Sprintf("UPDATE %s SET %s = %s", tbl.name, col, col))
		}
		deny(pgtest.RoleApp, "TRUNCATE "+tbl.name)
		deny(pgtest.RoleWebhook, "SELECT 1 FROM "+tbl.name)
		deny(pgtest.RoleWorker, "SELECT 1 FROM "+tbl.name)
	}
	deny(pgtest.RoleApp, "DELETE FROM personal_deductions")
	deny(pgtest.RoleApp, "DELETE FROM annual_inputs")
	for _, d := range denied {
		t.Run(string(d.role)+": "+d.query, func(t *testing.T) {
			_, err := execAsUser(t, db.Connect(t, d.role), userA, d.query)
			requireSQLState(t, err, sqlstateInsufficientPrivilege)
		})
	}
	t.Run("cc_app runs what the use cases need", func(t *testing.T) {
		for _, query := range []string{
			`UPDATE personal_deductions SET category = 'donation', donee_type = 'government', amount_cents = 2,
			 paid_on = '2026-12-31', payment_method = 'cash', beneficiary = 'spouse', cfdi_uuid = NULL`,
			`UPDATE personal_deductions SET category = 'tuition', donee_type = NULL, student_label = 'Leo',
			 tuition_level = 'secundaria', scholarship_cents = 0`,
			"UPDATE personal_deductions SET deleted_at = now()",
			"UPDATE personal_deductions SET deleted_at = NULL",
			`UPDATE salary_statements SET employer_label = 'Patrón 1', accumulable_cents = 1, exempt_cents = 1,
			 isr_withheld_cents = 1, simultaneous = true, left_before_year_end = true, payer_not_withholding = true`,
			"DELETE FROM salary_statements",
			"DELETE FROM annual_filed_payments WHERE regime_code = '612' AND fiscal_year = 2026",
			`INSERT INTO annual_filed_payments (user_id, taxpayer_id, regime_code, fiscal_year, month, amount_cents)
			 VALUES ('` + userA + `', '` + taxpayerA + `', '612', 2026, 2, 300000)
			 ON CONFLICT (user_id, taxpayer_id, regime_code, fiscal_year, month)
			 DO UPDATE SET amount_cents = EXCLUDED.amount_cents`,
			`INSERT INTO annual_inputs (user_id, taxpayer_id, fiscal_year, prior_year_accumulable_income_cents,
			                           other_exempt_income_cents, told_employer_will_file)
			 VALUES ('` + userA + `', '` + taxpayerA + `', 2026, NULL, 5, true)
			 ON CONFLICT (user_id, taxpayer_id, fiscal_year) DO UPDATE SET
			   prior_year_accumulable_income_cents = EXCLUDED.prior_year_accumulable_income_cents,
			   other_exempt_income_cents = EXCLUDED.other_exempt_income_cents,
			   told_employer_will_file = EXCLUDED.told_employer_will_file`,
		} {
			affected, err := execAsUser(t, app, userA, query)
			require.NoError(t, err, query)
			assert.EqualValues(t, 1, affected, query)
		}
	})
}

// updated_at is now() of the writing transaction (set_updated_at).
func TestAnnualTables_UpdatedAtFollowsWrites(t *testing.T) {
	touch := map[string]string{
		personalDeductions.name: "UPDATE personal_deductions SET amount_cents = 1",
		salaryStatements.name:   "UPDATE salary_statements SET exempt_cents = 1",
		filedPayments.name:      "UPDATE annual_filed_payments SET amount_cents = 1",
		annualInputs.name:       "UPDATE annual_inputs SET told_employer_will_file = true",
	}
	for _, tbl := range allAnnualTables {
		t.Run(tbl.name, func(t *testing.T) {
			app := seedAnnualOwners(t, pgtest.NewDatabase(t))
			require.NoError(t, tbl.insert(t, app, userA, tbl.row(userA, nil)))
			var followed bool
			require.NoError(t, pgtest.AsUser(t.Context(), app, userA, func(tx *sql.Tx) error {
				if _, err := tx.ExecContext(t.Context(), "SELECT pg_sleep(0.01)"); err != nil {
					return err
				}
				if _, err := tx.ExecContext(t.Context(), touch[tbl.name]); err != nil {
					return err
				}
				return tx.QueryRowContext(t.Context(),
					"SELECT bool_and(updated_at = now() AND updated_at > created_at) FROM "+tbl.name).Scan(&followed)
			}))
			assert.True(t, followed)
		})
	}
}

// §2.6 / PLAN ANN-03: annual_depreciation (money, optional) for 612 PF, 601 PM and
// 626 PM, last in each list; no other regime has it.
func TestRegimes_AnnualDepreciationKey(t *testing.T) {
	db := pgtest.NewDatabase(t)
	owner := db.Connect(t, pgtest.RoleOwner)
	want := `{"key": "annual_depreciation", "label": "Depreciación fiscal del ejercicio", "kind": "money",
		"required": false}`
	assert.Equal(t, []string{"601/moral", "612/fisica", "626/moral"}, queryStrings(t, owner, `
		SELECT code || '/' || person_type FROM regimes
		WHERE required_inputs @> '[{"key": "annual_depreciation"}]' ORDER BY 1`))
	for _, regime := range []string{"601/moral", "612/fisica", "626/moral"} {
		keys := jsonKeys(t, owner, "required_inputs")[regime]
		assert.Equal(t, "annual_depreciation", keys[len(keys)-1], regime)
	}
	var def string
	require.NoError(t, owner.QueryRowContext(t.Context(), `
		SELECT d::text FROM regimes, jsonb_array_elements(required_inputs) d
		WHERE code = '612' AND person_type = 'fisica' AND d->>'key' = 'annual_depreciation'`).Scan(&def))
	assert.JSONEq(t, want, def)

	t.Run("a value is stored as money", func(t *testing.T) {
		app := seedTaxInputOwners(t, db)
		require.NoError(t, insertTaxInput(t, app, userA, taxInputRow{
			userID: userA, taxpayerID: taxpayerA, regime: "612", fiscalYear: 2026,
			key: "annual_depreciation", cents: int64(1200000),
		}))
		var kind string
		queryAsUser(t, app, userA, "SELECT kind FROM taxpayer_tax_inputs WHERE input_key = 'annual_depreciation'", &kind)
		assert.Equal(t, "money", kind)
	})
}

// The per-year reads (the annual estimate and each input screen) and the cascades are
// served by the indexes, under RLS as cc_app.
func TestAnnualTables_IndexesServeTheReadPaths(t *testing.T) {
	db := pgtest.NewDatabase(t)
	app := db.Connect(t, pgtest.RoleApp)
	seedSQL := []string{
		fmt.Sprintf(`
		INSERT INTO personal_deductions (user_id, taxpayer_id, category, amount_cents, paid_on, payment_method, beneficiary)
		SELECT $1::uuid, $2::uuid,
		       (ARRAY['medical', 'funeral', 'health_insurance', 'school_transport', 'local_payroll_tax',
		              'mortgage_real_interest', 'retirement_contribution', 'medical_disability'])[1 + g %% 8],
		       1000 + (g * 37) %% 500000, date '2026-01-01' + (g * 11) %% 1826,
		       CASE WHEN g %% 9 = 0 THEN 'cash' ELSE 'electronic' END, 'self'
		FROM generate_series(1, %d) AS g`, planDeductionsPerUser),
		"UPDATE personal_deductions SET deleted_at = now() WHERE amount_cents % 20 = 0",
		fmt.Sprintf(`
		INSERT INTO salary_statements (user_id, taxpayer_id, fiscal_year, accumulable_cents, exempt_cents, isr_withheld_cents)
		SELECT $1::uuid, $2::uuid, y, 30000000, 0, 4000000
		FROM generate_series(%d, %d) AS y, generate_series(1, 2)`, annualPlanFirstYear, annualPlanLastYear),
		fmt.Sprintf(`
		INSERT INTO annual_filed_payments (user_id, taxpayer_id, regime_code, fiscal_year, month, amount_cents)
		SELECT $1::uuid, $2::uuid, '612', y, m, 100000 * m
		FROM generate_series(%d, %d) AS y, generate_series(1, 12) AS m`, annualPlanFirstYear, annualPlanLastYear),
		fmt.Sprintf(`
		INSERT INTO annual_inputs (user_id, taxpayer_id, fiscal_year, prior_year_accumulable_income_cents)
		SELECT $1::uuid, $2::uuid, y, 0 FROM generate_series(%d, %d) AS y`, annualPlanFirstYear, annualPlanLastYear),
	}
	for i := range planUsers {
		user, taxpayer := planUserID(i), planTaxpayerID(i)
		seedUsers(t, app, user)
		seedTaxpayer(t, app, user, taxpayer, "fisica")
		_, err := execAsUser(t, app, user, insertRegimeSQL, user, taxpayer, "612", "fisica", "2022-01", nil)
		require.NoError(t, err)
		for _, stmt := range seedSQL {
			args := []any{user, taxpayer}
			if !strings.Contains(stmt, "$1") {
				args = nil
			}
			_, err = execAsUser(t, app, user, stmt, args...)
			require.NoError(t, err)
		}
	}
	owner := db.Connect(t, pgtest.RoleOwner)
	for _, tbl := range allAnnualTables {
		_, err := owner.ExecContext(t.Context(), "ANALYZE "+tbl.name)
		require.NoError(t, err)
	}

	user, taxpayer := planUserID(7), planTaxpayerID(7)
	queries := []struct {
		name, table, index, query string
	}{
		{"deductions of the year (estimate)", "personal_deductions", "personal_deductions_live_year_idx", fmt.Sprintf(`
			SELECT category, amount_cents, payment_method, beneficiary, donee_type, student_label, tuition_level,
			       scholarship_cents
			FROM personal_deductions
			WHERE taxpayer_id = '%s' AND fiscal_year = 2026 AND deleted_at IS NULL`, taxpayer)},
		{"deductions list by category", "personal_deductions", "personal_deductions_live_year_idx", fmt.Sprintf(`
			SELECT id, amount_cents, paid_on FROM personal_deductions
			WHERE taxpayer_id = '%s' AND fiscal_year = 2026 AND category = 'medical' AND deleted_at IS NULL
			ORDER BY paid_on DESC, id DESC`, taxpayer)},
		{"constancias of the year", "salary_statements", "salary_statements_year_idx", fmt.Sprintf(`
			SELECT id, accumulable_cents, exempt_cents, isr_withheld_cents FROM salary_statements
			WHERE taxpayer_id = '%s' AND fiscal_year = 2026 ORDER BY id`, taxpayer)},
		{"filed payments of a regime and year", "annual_filed_payments", "annual_filed_payments_pkey", fmt.Sprintf(`
			SELECT month, amount_cents FROM annual_filed_payments
			WHERE taxpayer_id = '%s' AND regime_code = '612' AND fiscal_year = 2026 ORDER BY month`, taxpayer)},
		{"annual inputs of the year", "annual_inputs", "annual_inputs_pkey", fmt.Sprintf(`
			SELECT prior_year_accumulable_income_cents, other_exempt_income_cents, told_employer_will_file
			FROM annual_inputs WHERE taxpayer_id = '%s' AND fiscal_year = 2026`, taxpayer)},
	}
	for _, q := range queries {
		t.Run(q.name, func(t *testing.T) {
			plan := explain(t, app, user, q.query)
			assert.Contains(t, plan, q.index)
			assert.NotContains(t, plan, "Seq Scan on "+q.table)
		})
	}
	t.Run("the cascades reach soft-deleted deductions through a full index", func(t *testing.T) {
		assert.Contains(t, queryStrings(t, owner,
			"SELECT regexp_replace(indexdef, '^.* USING ', '') FROM pg_indexes WHERE tablename = 'personal_deductions'"),
			"btree (user_id, taxpayer_id)")
	})
}

func TestAnnualMigrations_DownRemovesEverything(t *testing.T) {
	db := pgtest.NewDatabase(t)
	app := seedTaxInputOwners(t, db)
	require.NoError(t, insertTaxInput(t, app, userA, taxInputRow{
		userID: userA, taxpayerID: taxpayerA, regime: "612", fiscalYear: 2026,
		key: "annual_depreciation", cents: int64(1),
	}))
	require.NoError(t, insertTaxInput(t, app, userA, annualMoney(userA)))
	owner := db.Connect(t, pgtest.RoleOwner)
	p, err := migrations.NewProvider(owner)
	require.NoError(t, err)

	_, err = p.DownTo(t.Context(), annualDepreciationVersion-1)
	require.NoError(t, err)
	for _, tbl := range allAnnualTables {
		var found sql.NullString
		require.NoError(t, owner.QueryRowContext(t.Context(), "SELECT to_regclass($1)::text", "public."+tbl.name).Scan(&found))
		assert.False(t, found.Valid, "down left %s", found.String)
	}
	assert.Empty(t, queryStrings(t, owner,
		`SELECT code FROM regimes WHERE required_inputs @> '[{"key": "annual_depreciation"}]'`))
	var keys []string
	require.NoError(t, pgtest.AsUser(t.Context(), app, userA, func(tx *sql.Tx) error {
		rows, err := tx.QueryContext(t.Context(), "SELECT input_key FROM taxpayer_tax_inputs ORDER BY 1")
		if err != nil {
			return err
		}
		defer func() { _ = rows.Close() }()
		for rows.Next() {
			var k string
			if err := rows.Scan(&k); err != nil {
				return err
			}
			keys = append(keys, k)
		}
		return rows.Err()
	}))
	assert.Equal(t, []string{"ptu_paid"}, keys, "values of the removed key go with it; others stay")
	require.NoError(t, pgtest.CheckRLSForced(t.Context(), owner, taxInputsTable), "FORCE is restored")

	require.NoError(t, migrations.Up(t.Context(), owner))
	require.NoError(t, migrations.Up(t.Context(), owner), "a second up is a no-op")
	assert.Len(t, queryStrings(t, owner, `SELECT d->>'key' FROM regimes, jsonb_array_elements(required_inputs) d
		WHERE code = '612' AND person_type = 'fisica' AND d->>'key' = 'annual_depreciation'`), 1, "added once")
	for _, tbl := range allAnnualTables {
		var exists bool
		require.NoError(t, owner.QueryRowContext(t.Context(), "SELECT to_regclass($1) IS NOT NULL", "public."+tbl.name).Scan(&exists))
		assert.True(t, exists, tbl.name)
	}
}
