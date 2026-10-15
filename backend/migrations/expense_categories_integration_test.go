//go:build integration

package migrations_test

import (
	"context"
	"database/sql"
	"fmt"
	"maps"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/platform/pgtest"
	"github.com/romodevx/control-contable/backend/migrations"
)

// Expense category catalog (M4-02): the global expense_categories and
// category_regime_rules tables, seeded from docs/project-specs/MODULE-expenses.md §2.2
// (DECIDED OQ-E3/OQ-E4) with the legal basis of docs/research/tax-regimes-mx-2026.md.

// expenseCategoriesVersion is the goose version of 20261013210000_expense_categories.sql.
const expenseCategoriesVersion = 20261013210000

// specCategory is one row of the MODULE-expenses §2.2 table. business and lessor are the
// D/P/N letters of the "Business" and "606" columns; the 626 PF column is X for every row.
type specCategory struct {
	code, name       string
	business, lessor string
	defaultBusiness  bool
	defaultLessor    bool
	basis            string
	note             string // "" = no note shown
}

const notePMWithholding = "Si eres persona moral, revisa las retenciones"
const noteDepreciation = "Se deduce por depreciación anual, no completo en el mes"

// specCategories is MODULE-expenses §2.2 in its order (sort_order 1..26). Row 13's
// "Igual que el anterior" is spelled out: the list is searchable and sorted by use, so
// a note may never point at its neighbor.
var specCategories = []specCategory{
	{"mercancia", "Mercancía y materias primas", "D", "N", true, false, "LISR arts. 25, 103, 208", ""},
	{"servicios_profesionales", "Honorarios y servicios profesionales", "D", "D", true, true, "LISR arts. 25, 103, 115", notePMWithholding},
	{"renta_local", "Renta de local, oficina o bodega", "D", "N", true, false, "LISR arts. 25, 103", notePMWithholding},
	{"servicios_basicos", "Luz, agua, teléfono e internet", "D", "P", true, true, "LISR arts. 103, 115", "Arrendamiento: sólo el agua del inmueble rentado"},
	{"software", "Software y suscripciones", "D", "N", true, false, "LISR art. 103", ""},
	{"papeleria", "Papelería y artículos de oficina", "D", "N", true, false, "LISR art. 103", ""},
	{"publicidad", "Publicidad y mercadotecnia", "D", "N", true, false, "LISR art. 103", ""},
	{"combustible", "Combustibles", "D", "N", true, false, "LISR art. 27", "Sólo deducible si lo pagas con tarjeta, transferencia o monedero, sin importar el monto"},
	{"vehiculo_mantenimiento", "Mantenimiento y refacciones de vehículo", "P", "N", true, false, "LISR arts. 28, 36", "Tiene topes de deducción"},
	{"viaticos", "Viáticos y gastos de viaje", "P", "N", true, false, "LISR art. 28", "Tiene topes y requisitos"},
	{"restaurantes", "Consumo en restaurantes", "P", "N", false, false, "LISR art. 28", "Sólo una parte es deducible"},
	{"equipo_computo", "Equipo de cómputo (inversión)", "P", "N", false, false, "LISR arts. 31, 34", noteDepreciation},
	{"mobiliario", "Mobiliario y equipo (inversión)", "P", "N", false, false, "LISR arts. 31, 34", noteDepreciation},
	{"vehiculo_compra", "Compra de automóvil (inversión)", "P", "N", false, false, "LISR arts. 34, 36", "Depreciación con tope"},
	{"construcciones", "Construcciones y mejoras al inmueble", "P", "P", false, false, "LISR arts. 34, 115", "Se deduce por depreciación"},
	{"sueldos", "Sueldos y salarios pagados", "P", "D", true, true, "LISR arts. 28, 115", "Parte de las prestaciones exentas no es deducible"},
	{"cuotas_imss", "Cuotas IMSS e INFONAVIT patronales", "D", "D", true, true, "LISR arts. 25, 103, 115", ""},
	{"intereses", "Intereses pagados", "D", "D", true, true, "LISR arts. 25, 103, 115", "Sólo los intereses reales"},
	{"seguros", "Seguros y fianzas", "D", "D", true, true, "LISR arts. 25, 103, 115", ""},
	{"mantenimiento", "Mantenimiento y reparaciones", "D", "D", true, true, "LISR arts. 103, 115", ""},
	{"predial", "Impuesto predial", "D", "D", true, true, "LISR arts. 103, 115", "En arrendamiento se suma aun con la deducción ciega"},
	{"capacitacion", "Capacitación y cursos", "D", "N", true, false, "LISR art. 103", ""},
	{"comisiones", "Comisiones bancarias y de plataformas", "D", "N", true, false, "LISR art. 103", ""},
	{"fletes", "Fletes y paquetería", "D", "N", true, false, "LISR art. 103", "Si eres persona moral, revisa la retención de IVA (4 %)"},
	{"gastos_personales", "Gastos personales (médicos, colegiaturas, donativos)", "N", "N", false, false, "LISR art. 151", "Se aplican en tu declaración anual, no en el pago mensual"},
	{"otros", "Otros gastos", "P", "N", false, false, "LISR art. 27", "Verifica que cumpla los requisitos"},
}

// Regime classes of MODULE-expenses §2.2: "Business" (one rule row per regime, same
// values), 606 arrendamiento, 626 PF (RESICO PF) and the informational rest.
var (
	businessRegimes = map[string]bool{
		"612/fisica": true, "621/fisica": true, "625/fisica": true,
		"601/moral": true, "603/moral": true, "620/moral": true, "622/moral": true, "626/moral": true,
	}
	// Regimes whose provisional ISR uses the profit coefficient (required input
	// coeficiente_utilidad in the regime catalog).
	coefficientRegimes = map[string]bool{"601/moral": true, "622/moral": true}
)

const (
	lessorRegime   = "606/fisica"
	resicoPFRegime = "626/fisica"

	noteCoefficient   = "En los pagos provisionales del 601 el ISR se calcula con el coeficiente de utilidad; este gasto cuenta para tu IVA y tu declaración anual"
	noteResicoPF      = "En RESICO tus gastos no reducen tu ISR, pero el IVA que pagaste sí puede acreditarse"
	noteInformational = "Este régimen es informativo"
)

var deductibilityOf = map[string]string{"D": "deductible", "P": "partially_deductible", "N": "not_deductible"}

// categoryRule is one category_regime_rules row as read back from the database.
type categoryRule struct {
	deductibility, cashRule, note, legalBasis, sourceURL, reviewNote string
	defaultDeductible, ivaCreditable, needsReview                    bool
}

func categoryRules(t *testing.T, db *sql.DB) map[string]categoryRule {
	t.Helper()
	rows, err := db.QueryContext(t.Context(), `
		SELECT category_code || '@' || regime_code || '/' || person_type, deductibility, default_deductible,
		       iva_creditable_default, cash_payment_rule, coalesce(note, ''), legal_basis, source_url,
		       needs_cpa_review, coalesce(review_note, '')
		FROM category_regime_rules`)
	require.NoError(t, err)
	defer func() { _ = rows.Close() }()
	rules := map[string]categoryRule{}
	for rows.Next() {
		var key string
		var r categoryRule
		require.NoError(t, rows.Scan(&key, &r.deductibility, &r.defaultDeductible, &r.ivaCreditable, &r.cashRule,
			&r.note, &r.legalBasis, &r.sourceURL, &r.needsReview, &r.reviewNote))
		rules[key] = r
	}
	require.NoError(t, rows.Err())
	return rules
}

// expectedRule derives a rule row from the spec table for one regime.
func expectedRule(c specCategory, regime string) categoryRule {
	switch {
	case regime == lessorRegime:
		return liveRule(c.code, c.lessor, c.defaultLessor, c.note)
	case regime == resicoPFRegime:
		return categoryRule{deductibility: "isr_not_applicable", cashRule: "not_applicable", note: noteResicoPF,
			ivaCreditable: c.business != "N"}
	case businessRegimes[regime]:
		r := liveRule(c.code, c.business, c.defaultBusiness, c.note)
		if coefficientRegimes[regime] {
			r.note = joinNotes(c.note, noteCoefficient)
		}
		return r
	default:
		return categoryRule{deductibility: "isr_not_applicable", cashRule: "not_applicable", note: noteInformational}
	}
}

func liveRule(code, letter string, defaultDeductible bool, note string) categoryRule {
	r := categoryRule{deductibility: deductibilityOf[letter], defaultDeductible: defaultDeductible, note: note,
		ivaCreditable: letter != "N", cashRule: "not_applicable"}
	if letter == "N" {
		return r
	}
	r.cashRule = "over_limit"
	if code == "combustible" {
		r.cashRule = "any_cash"
	}
	return r
}

func joinNotes(first, second string) string {
	if first == "" {
		return second
	}
	return first + ". " + second
}

func TestExpenseCategories_SeedEverySpecCategoryInOrder(t *testing.T) {
	owner := pgtest.NewDatabase(t).Connect(t, pgtest.RoleOwner)
	rows, err := owner.QueryContext(t.Context(), `
		SELECT code, name, sort_order, legal_basis, source_url, needs_cpa_review, coalesce(review_note, '')
		FROM expense_categories ORDER BY sort_order`)
	require.NoError(t, err)
	defer func() { _ = rows.Close() }()

	i := 0
	for rows.Next() {
		var code, name, basis, url, reviewNote string
		var sortOrder int
		var flagged bool
		require.NoError(t, rows.Scan(&code, &name, &sortOrder, &basis, &url, &flagged, &reviewNote))
		require.Less(t, i, len(specCategories), "extra category %s", code)
		want := specCategories[i]
		assert.Equal(t, want.code, code)
		assert.Equal(t, want.name, name, code)
		assert.Equal(t, i+1, sortOrder, code)
		assert.Equal(t, want.basis, basis, code)
		assert.Equal(t, "https://www.diputados.gob.mx/LeyesBiblio/pdf/LISR.pdf", url, code)
		assert.True(t, flagged, "%s: every category waits for the CPA (OQ-E3)", code)
		assert.NotEmpty(t, reviewNote, code)
		i++
	}
	require.NoError(t, rows.Err())
	assert.Len(t, specCategories, i)
}

// One rule row per (category, catalog regime, person type), with the values of the
// spec table for the regime's class.
func TestCategoryRegimeRules_CompleteMatrixMatchesTheSpec(t *testing.T) {
	owner := pgtest.NewDatabase(t).Connect(t, pgtest.RoleOwner)
	rules := categoryRules(t, owner)

	regimes := queryStrings(t, owner, "SELECT code || '/' || person_type FROM regimes ORDER BY 1")
	require.Len(t, regimes, len(researchCatalog))
	assert.Len(t, rules, len(specCategories)*len(regimes), "one row per category and regime, no extras")

	for _, c := range specCategories {
		for _, regime := range regimes {
			key := c.code + "@" + regime
			t.Run(key, func(t *testing.T) {
				got, ok := rules[key]
				require.True(t, ok, "missing rule")
				want := expectedRule(c, regime)
				assert.Equal(t, want.deductibility, got.deductibility)
				assert.Equal(t, want.defaultDeductible, got.defaultDeductible, "default_deductible")
				assert.Equal(t, want.ivaCreditable, got.ivaCreditable, "iva_creditable_default")
				assert.Equal(t, want.cashRule, got.cashRule, "cash_payment_rule")
				assert.Equal(t, want.note, got.note)
				assert.NotEmpty(t, got.legalBasis)
				assert.Regexp(t, `^https://`, got.sourceURL)
				assert.True(t, got.needsReview, "every rule waits for the CPA (OQ-E3)")
				assert.NotEmpty(t, got.reviewNote)
			})
		}
	}
}

// The spec's catalog invariant (§2.2), checked on the seeded data.
func TestCategoryRegimeRules_NeverDefaultDeductibleWhenNotDeductible(t *testing.T) {
	owner := pgtest.NewDatabase(t).Connect(t, pgtest.RoleOwner)
	offending := queryStrings(t, owner, `
		SELECT category_code || '@' || regime_code || '/' || person_type FROM category_regime_rules
		WHERE default_deductible AND deductibility IN ('not_deductible', 'isr_not_applicable')`)
	assert.Empty(t, offending)
}

// The coefficient regimes are the catalog's, not a hard-coded list: a regime that
// requires coeficiente_utilidad gets the note.
func TestCategoryRegimeRules_CoefficientRegimesFollowTheRegimeCatalog(t *testing.T) {
	owner := pgtest.NewDatabase(t).Connect(t, pgtest.RoleOwner)
	got := queryStrings(t, owner, `
		SELECT code || '/' || person_type FROM regimes
		WHERE required_inputs @> '[{"key": "coeficiente_utilidad"}]' ORDER BY 1`)
	want := make([]string, 0, len(coefficientRegimes))
	for regime := range coefficientRegimes {
		want = append(want, regime)
	}
	assert.ElementsMatch(t, want, got)
}

func insertCategory(owner *sql.DB, code, name string, sortOrder int, flagged bool, reviewNote any) error {
	_, err := owner.Exec(`
		INSERT INTO expense_categories (code, name, sort_order, legal_basis, source_url, needs_cpa_review, review_note)
		VALUES ($1, $2, $3, 'LISR art. 103', 'https://www.diputados.gob.mx/LeyesBiblio/pdf/LISR.pdf', $4, $5)`,
		code, name, sortOrder, flagged, reviewNote)
	return err
}

func TestExpenseCategories_RejectMalformedRows(t *testing.T) {
	owner := pgtest.NewDatabase(t).Connect(t, pgtest.RoleOwner)
	cases := []struct {
		name, code, catName string
		sortOrder           int
		flagged             bool
		reviewNote          any
		want                string
	}{
		{"code with a hyphen", "viajes-luna", "Viajes", 100, true, "r", sqlstateCheckViolation},
		{"upper-case code", "Viajes", "Viajes", 100, true, "r", sqlstateCheckViolation},
		{"code over 40 characters", "a123456789012345678901234567890123456789x", "Viajes", 100, true, "r", sqlstateCheckViolation},
		{"blank name", "viajes", "  ", 100, true, "r", sqlstateCheckViolation},
		{"untrimmed name", "viajes", " Viajes", 100, true, "r", sqlstateCheckViolation},
		{"zero sort order", "viajes", "Viajes", 0, true, "r", sqlstateCheckViolation},
		{"duplicate sort order", "viajes", "Viajes", 1, true, "r", sqlstateUniqueViolation},
		{"duplicate code", "otros", "Otros", 100, true, "r", sqlstateUniqueViolation},
		{"flagged without a note", "viajes", "Viajes", 100, true, nil, sqlstateCheckViolation},
		{"note without the flag", "viajes", "Viajes", 100, false, "r", sqlstateCheckViolation},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			requireSQLState(t, insertCategory(owner, c.code, c.catName, c.sortOrder, c.flagged, c.reviewNote), c.want)
		})
	}
	t.Run("a well-formed row is accepted", func(t *testing.T) {
		require.NoError(t, insertCategory(owner, "viajes", "Viajes", 100, false, nil))
	})
}

// ruleColumns are the overridable columns of a test rule insert, as SQL literals.
type ruleColumns map[string]string

func insertRule(owner *sql.DB, overrides ruleColumns) error {
	cols := ruleColumns{
		"category_code":          "'software'",
		"regime_code":            "'612'",
		"person_type":            "'fisica'",
		"deductibility":          "'deductible'",
		"default_deductible":     "true",
		"iva_creditable_default": "true",
		"cash_payment_rule":      "'over_limit'",
		"note":                   "NULL",
		"legal_basis":            "'LISR art. 103'",
		"source_url":             "'https://www.diputados.gob.mx/LeyesBiblio/pdf/LISR.pdf'",
		"needs_cpa_review":       "true",
		"review_note":            "'r'",
	}
	for k, v := range overrides {
		cols[k] = v
	}
	names, values := splitColumns(cols)
	_, err := owner.Exec(fmt.Sprintf("INSERT INTO category_regime_rules (%s) VALUES (%s)", names, values))
	return err
}

func TestCategoryRegimeRules_RejectMalformedRows(t *testing.T) {
	owner := pgtest.NewDatabase(t).Connect(t, pgtest.RoleOwner)
	_, err := owner.Exec("DELETE FROM category_regime_rules WHERE category_code = 'software' AND regime_code = '612'")
	require.NoError(t, err)

	cases := []struct {
		name      string
		overrides ruleColumns
		want      string
	}{
		{"unknown category", ruleColumns{"category_code": "'viajes'"}, sqlstateForeignKeyViolation},
		{"unknown regime", ruleColumns{"regime_code": "'699'"}, sqlstateForeignKeyViolation},
		{"regime of the other person type", ruleColumns{
			"person_type": "'moral'", "default_operation_kind": "'business_services'",
		}, sqlstateForeignKeyViolation},
		{"unknown deductibility", ruleColumns{"deductibility": "'maybe'"}, sqlstateCheckViolation},
		{"default true while not deductible", ruleColumns{"deductibility": "'not_deductible'", "cash_payment_rule": "'not_applicable'"}, sqlstateCheckViolation},
		{"default true while ISR does not apply", ruleColumns{"deductibility": "'isr_not_applicable'", "cash_payment_rule": "'not_applicable'"}, sqlstateCheckViolation},
		{"unknown cash rule", ruleColumns{"cash_payment_rule": "'never'"}, sqlstateCheckViolation},
		{"cash rule on a non-deductible row", ruleColumns{"deductibility": "'not_deductible'", "default_deductible": "false"}, sqlstateCheckViolation},
		{"no cash rule on a deductible row", ruleColumns{"cash_payment_rule": "'not_applicable'"}, sqlstateCheckViolation},
		{"blank note", ruleColumns{"note": "' '"}, sqlstateCheckViolation},
		{"blank legal basis", ruleColumns{"legal_basis": "''"}, sqlstateCheckViolation},
		{"source url not http", ruleColumns{"source_url": "'ftp://x'"}, sqlstateCheckViolation},
		{"flagged without a note", ruleColumns{"review_note": "NULL"}, sqlstateCheckViolation},
		{"null default", ruleColumns{"default_deductible": "NULL"}, sqlstateNotNullViolation},
		{"null iva creditable default", ruleColumns{"iva_creditable_default": "NULL"}, sqlstateNotNullViolation},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			requireSQLState(t, insertRule(owner, c.overrides), c.want)
		})
	}
	t.Run("a well-formed row is accepted", func(t *testing.T) {
		require.NoError(t, insertRule(owner, nil))
	})
}

var expenseCatalogTables = []string{"expense_categories", "category_regime_rules"}

func TestExpenseCatalog_ReadOnlyForTheApp(t *testing.T) {
	db := pgtest.NewDatabase(t)
	app := db.Connect(t, pgtest.RoleApp)

	var categories, rules int
	require.NoError(t, app.QueryRowContext(t.Context(), "SELECT count(*) FROM expense_categories").Scan(&categories))
	require.NoError(t, app.QueryRowContext(t.Context(), "SELECT count(*) FROM category_regime_rules").Scan(&rules))
	assert.Len(t, specCategories, categories)
	assert.Equal(t, len(specCategories)*len(researchCatalog), rules)

	for _, stmt := range []string{
		`INSERT INTO expense_categories (code, name, sort_order, legal_basis, source_url)
		 VALUES ('viajes', 'Viajes', 100, 'L', 'https://x.mx')`,
		"UPDATE expense_categories SET name = 'x'",
		"DELETE FROM expense_categories",
		"TRUNCATE expense_categories CASCADE",
		"UPDATE category_regime_rules SET default_deductible = true",
		"DELETE FROM category_regime_rules",
		"TRUNCATE category_regime_rules",
	} {
		t.Run(stmt, func(t *testing.T) {
			_, err := app.ExecContext(t.Context(), stmt)
			requireSQLState(t, err, sqlstateInsufficientPrivilege)
		})
	}
}

func TestExpenseCatalog_NoOtherRuntimeRoleReadsIt(t *testing.T) {
	db := pgtest.NewDatabase(t)
	// cc_worker reads the catalog since WH-08 (20261018100000, the withholdings backfill).
	for _, role := range []pgtest.Role{pgtest.RoleWebhook} {
		for _, table := range expenseCatalogTables {
			t.Run(string(role)+"/"+table, func(t *testing.T) {
				_, err := db.Connect(t, role).ExecContext(t.Context(), "SELECT 1 FROM "+table)
				requireSQLState(t, err, sqlstateInsufficientPrivilege)
			})
		}
	}
}

// Global read-only pattern (P0-10, M2-02): ENABLE (not FORCE), one SELECT policy for
// cc_app, no write policy, so a write granted by mistake is still refused or matches no row.
func TestExpenseCatalog_RowSecurityDeniesUngrantedWrites(t *testing.T) {
	db := pgtest.NewDatabase(t)
	owner := db.Connect(t, pgtest.RoleOwner)
	for _, table := range expenseCatalogTables {
		var enabled, forced bool
		require.NoError(t, owner.QueryRowContext(t.Context(),
			"SELECT relrowsecurity, relforcerowsecurity FROM pg_class WHERE oid = to_regclass($1)", table).
			Scan(&enabled, &forced))
		assert.True(t, enabled, "%s: RLS enabled", table)
		assert.False(t, forced, "%s: not forced, so cc_owner keeps writing through migrations", table)
	}
	_, err := owner.ExecContext(t.Context(),
		"GRANT INSERT, UPDATE, DELETE ON expense_categories, category_regime_rules TO cc_app")
	require.NoError(t, err)
	app := db.Connect(t, pgtest.RoleApp)

	_, err = app.ExecContext(t.Context(), `
		INSERT INTO expense_categories (code, name, sort_order, legal_basis, source_url)
		VALUES ('viajes', 'Viajes', 100, 'L', 'https://x.mx')`)
	requireSQLState(t, err, sqlstateInsufficientPrivilege)

	for _, stmt := range []string{
		"UPDATE expense_categories SET name = 'x'",
		"UPDATE category_regime_rules SET default_deductible = true",
		"DELETE FROM category_regime_rules",
	} {
		res, err := app.ExecContext(t.Context(), stmt)
		require.NoError(t, err, stmt)
		affected, err := res.RowsAffected()
		require.NoError(t, err)
		assert.Zero(t, affected, stmt)
	}
}

// The categories endpoint reads every rule of one (regime, person type): the primary
// key leads with them, so it serves that lookup and the foreign key to regimes.
func TestCategoryRegimeRules_RegimeLookupUsesAnIndex(t *testing.T) {
	owner := pgtest.NewDatabase(t).Connect(t, pgtest.RoleOwner)
	_, err := owner.ExecContext(t.Context(), "ANALYZE category_regime_rules")
	require.NoError(t, err)
	plan := explain(t, owner, "", `
		SELECT category_code, deductibility, default_deductible, note, legal_basis
		FROM category_regime_rules WHERE regime_code = '612' AND person_type = 'fisica'`)
	assert.Contains(t, plan, "category_regime_rules_pkey")
}

func TestExpenseCatalogMigration_DownRemovesItAndUpReseedsIt(t *testing.T) {
	owner := pgtest.NewDatabase(t).Connect(t, pgtest.RoleOwner)
	p, err := migrations.NewProvider(owner)
	require.NoError(t, err)

	_, err = p.DownTo(t.Context(), expenseCategoriesVersion-1)
	require.NoError(t, err)
	for _, obj := range []string{"to_regclass('public.expense_categories')", "to_regclass('public.category_regime_rules')"} {
		var found sql.NullString
		require.NoError(t, owner.QueryRowContext(t.Context(), "SELECT "+obj+"::text").Scan(&found))
		assert.False(t, found.Valid, "down left %s", found.String)
	}

	require.NoError(t, migrations.Up(t.Context(), owner))
	require.NoError(t, migrations.Up(t.Context(), owner), "a second up is a no-op")
	var categories, rules int
	require.NoError(t, owner.QueryRowContext(t.Context(),
		"SELECT (SELECT count(*) FROM expense_categories), (SELECT count(*) FROM category_regime_rules)").
		Scan(&categories, &rules))
	assert.Len(t, specCategories, categories, "seeds applied exactly once")
	assert.Equal(t, len(specCategories)*len(researchCatalog), rules)
}

// splitColumns renders a column map as the column and value lists of an INSERT, in a
// stable order.
func splitColumns[M ~map[string]string](cols M) (names, values string) {
	keys := slices.Sorted(maps.Keys(cols))
	vals := make([]string, len(keys))
	for i, k := range keys {
		vals[i] = cols[k]
	}
	return strings.Join(keys, ", "), strings.Join(vals, ", ")
}

// explain returns the EXPLAIN (ANALYZE, BUFFERS) text of query, run under userID's
// context ("" = directly on db), and logs it for the PR.
func explain(t *testing.T, db *sql.DB, userID, query string) string {
	t.Helper()
	var lines []string
	run := func(q interface {
		QueryContext(context.Context, string, ...any) (*sql.Rows, error)
	}) error {
		rows, err := q.QueryContext(t.Context(), "EXPLAIN (ANALYZE, BUFFERS) "+query)
		if err != nil {
			return err
		}
		defer func() { _ = rows.Close() }()
		for rows.Next() {
			var line string
			if err := rows.Scan(&line); err != nil {
				return err
			}
			lines = append(lines, line)
		}
		return rows.Err()
	}
	if userID == "" {
		require.NoError(t, run(db))
	} else {
		require.NoError(t, pgtest.AsUser(t.Context(), db, userID, func(tx *sql.Tx) error { return run(tx) }))
	}
	plan := strings.Join(lines, "\n")
	t.Logf("EXPLAIN ANALYZE\n%s\n%s", strings.TrimSpace(query), plan)
	return plan
}
