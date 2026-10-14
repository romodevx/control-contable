//go:build integration

package migrations_test

import (
	"database/sql"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/platform/pgtest"
	"github.com/romodevx/control-contable/backend/migrations"
)

// Regime catalog (M2-02): the global regimes and regime_compatibility tables, seeded
// from docs/research/tax-regimes-mx-2026.md §A.1 (personas físicas) and §B.1 (personas
// morales). The catalog key is (code, person_type) per ADR-0002: 610 and 626 exist for
// both person types as distinct rows.

// regimesCatalogVersion is the goose version of 20261013180000_regimes_catalog.sql.
const regimesCatalogVersion = 20261013180000

// researchRegime is one row of the research catalog: the treatment of ADR-0002 §2
// (full | partial | informational) and the periodicity of §A.1/§B.1, empty where the
// research does not state it (those rows are flagged for CPA review instead).
type researchRegime struct {
	key         string // "<code>/<person_type>"
	treatment   string
	periodicity string
}

var researchCatalog = []researchRegime{
	{"601/moral", "full", "monthly"},
	{"603/moral", "partial", "annual"},
	{"605/fisica", "informational", "withheld_at_source"},
	{"606/fisica", "full", "monthly"},
	{"607/fisica", "informational", "per_transaction"},
	{"608/fisica", "informational", "annual"},
	{"610/fisica", "informational", "withheld_at_source"},
	{"610/moral", "informational", "withheld_at_source"},
	{"611/fisica", "informational", "annual"},
	{"612/fisica", "full", "monthly"},
	{"614/fisica", "informational", "withheld_at_source"},
	{"615/fisica", "informational", "withheld_at_source"},
	{"616/fisica", "informational", "none"},
	{"620/moral", "informational", ""},
	{"621/fisica", "partial", "bimonthly"},
	{"622/moral", "partial", "monthly"},
	{"623/moral", "informational", ""},
	{"624/moral", "informational", ""},
	{"625/fisica", "full", "monthly"},
	{"626/fisica", "full", "monthly"},
	{"626/moral", "full", "monthly"},
}

// regimesUnderReview are the catalog rows whose research basis is incomplete (⚠️,
// unstated or inferred); each carries a review_note instead of a guess. 603's "annual"
// is inferred: the research only says "generally no provisional ISR".
var regimesUnderReview = []string{"603/moral", "620/moral", "621/fisica", "623/moral", "624/moral", "625/fisica"}

func countRegimes(person string) int {
	n := 0
	for _, r := range researchCatalog {
		if r.key[4:] == person {
			n++
		}
	}
	return n
}

func TestRegimes_SeedsEveryResearchRegimeOnce(t *testing.T) {
	owner := pgtest.NewDatabase(t).Connect(t, pgtest.RoleOwner)

	rows, err := owner.QueryContext(t.Context(), `
		SELECT code || '/' || person_type, treatment, coalesce(periodicity, '')
		FROM regimes ORDER BY code, person_type`)
	require.NoError(t, err)
	defer func() { require.NoError(t, rows.Close()) }()
	var got []researchRegime
	for rows.Next() {
		var r researchRegime
		require.NoError(t, rows.Scan(&r.key, &r.treatment, &r.periodicity))
		got = append(got, r)
	}
	require.NoError(t, rows.Err())

	assert.Equal(t, researchCatalog, got)
}

func TestRegimes_KeyIsCodeAndPersonType(t *testing.T) {
	owner := pgtest.NewDatabase(t).Connect(t, pgtest.RoleOwner)

	var distinct, total int
	require.NoError(t, owner.QueryRowContext(t.Context(),
		"SELECT count(DISTINCT (code, person_type)), count(*) FROM regimes").Scan(&distinct, &total))
	assert.Equal(t, total, distinct, "no duplicate (code, person_type)")

	_, err := owner.ExecContext(t.Context(), `
		INSERT INTO regimes (code, person_type, name, treatment, periodicity, legal_basis, source_url, valid_from,
		                     income_basis)
		VALUES ('612', 'fisica', 'Duplicado', 'full', 'monthly', 'LISR', 'https://x.mx', '2022-01-01', 'cash')`)
	requireSQLState(t, err, sqlstateUniqueViolation)
}

func TestRegimes_EveryRowIsDatedAndCited(t *testing.T) {
	owner := pgtest.NewDatabase(t).Connect(t, pgtest.RoleOwner)

	uncited := queryStrings(t, owner, `
		SELECT code || '/' || person_type FROM regimes
		WHERE btrim(name) = '' OR btrim(legal_basis) = '' OR source_url !~ '^https?://'
		   OR valid_from <> DATE '2022-01-01' OR valid_to IS NOT NULL`)
	assert.Empty(t, uncited, "research §A.1: every key in force since 2022-01-01, no end date")

	flagged := queryStrings(t, owner, `
		SELECT code || '/' || person_type FROM regimes
		WHERE needs_cpa_review AND btrim(review_note) <> '' ORDER BY code, person_type`)
	assert.Equal(t, regimesUnderReview, flagged)
}

// jsonKeys returns "<code>/<person_type>" -> the "key" of each element of column, in order.
func jsonKeys(t *testing.T, db *sql.DB, column string) map[string][]string {
	t.Helper()
	rows, err := db.QueryContext(t.Context(), fmt.Sprintf(`
		SELECT r.code || '/' || r.person_type, e.def->>'key'
		FROM regimes r CROSS JOIN LATERAL jsonb_array_elements(r.%s) WITH ORDINALITY AS e(def, n)
		ORDER BY r.code, r.person_type, e.n`, column))
	require.NoError(t, err)
	defer func() { require.NoError(t, rows.Close()) }()
	out := map[string][]string{}
	for rows.Next() {
		var regime, key string
		require.NoError(t, rows.Scan(&regime, &key))
		out[regime] = append(out[regime], key)
	}
	require.NoError(t, rows.Err())
	return out
}

func TestRegimes_OptionDefsFollowADR0002(t *testing.T) {
	owner := pgtest.NewDatabase(t).Connect(t, pgtest.RoleOwner)

	assert.Equal(t, map[string][]string{
		"601/moral":  {"iva_fully_taxable", "iva_border_region"},
		"603/moral":  {"iva_fully_taxable", "iva_border_region"},
		"606/fisica": {"deduction_mode", "iva_border_region", "pays_quarterly"},
		"612/fisica": {"iva_fully_taxable", "iva_border_region"},
		"621/fisica": {"rif_year", "iva_border_region"},
		"622/moral":  {"iva_fully_taxable", "iva_border_region"},
		"625/fisica": {"activities", "definitive_option", "rfc_given_to_platform"},
		"626/fisica": {"iva_border_region"},
		"626/moral":  {"iva_fully_taxable", "iva_border_region"},
	}, jsonKeys(t, owner, "option_defs"))

	defs := map[string]string{
		"606/fisica deduction_mode": `{"key": "deduction_mode", "label": "Tipo de deducción", "type": "enum",
			"allowed_values": ["blind_35", "actual"], "default": "blind_35", "required": true}`,
		"606/fisica pays_quarterly": `{"key": "pays_quarterly",
			"label": "Hago pagos provisionales trimestrales (art. 116 LISR)", "type": "boolean",
			"default": false, "required": false}`,
		"625/fisica activities": `{"key": "activities", "label": "Actividades a través de plataformas",
			"type": "multi_enum", "allowed_values": ["transport_delivery", "lodging", "goods_services"],
			"default": null, "required": true}`,
		"612/fisica iva_border_region": `{"key": "iva_border_region",
			"label": "Aplico el estímulo de IVA al 8 % en región fronteriza", "type": "boolean",
			"default": false, "required": false}`,
	}
	for name, want := range defs {
		t.Run(name, func(t *testing.T) {
			var code, person, key string
			_, err := fmt.Sscanf(name, "%3s/%s %s", &code, &person, &key)
			require.NoError(t, err)
			var matches bool
			require.NoError(t, owner.QueryRowContext(t.Context(), `
				SELECT EXISTS (SELECT 1 FROM regimes, jsonb_array_elements(option_defs) d
				               WHERE code = $1 AND person_type = $2 AND d = $3::jsonb)`,
				code, person, want).Scan(&matches))
			assert.True(t, matches, "want %s", want)
		})
	}
}

func TestRegimes_RequiredInputsFollowADR0002(t *testing.T) {
	owner := pgtest.NewDatabase(t).Connect(t, pgtest.RoleOwner)

	ytdPM := []string{"prior_provisional_payments", "ptu_paid", "prior_year_losses", "interest_withholding"}
	assert.Equal(t, map[string][]string{
		// M5-20 adds first_fiscal_year to 601 only (MODULE-tax-engine §9 item 6).
		"601/moral": append([]string{"coeficiente_utilidad", "first_fiscal_year"}, ytdPM...),
		"612/fisica": {"prior_provisional_payments", "ytd_before_app_income", "ytd_before_app_deductions",
			"ptu_paid", "prior_year_losses"},
		"622/moral":  append([]string{"coeficiente_utilidad"}, ytdPM...),
		"625/fisica": {"prior_year_income"},
		// Research §B.2.2 step 4 credits the art. 54 interest withholding, as 601 does.
		"626/moral": {"prior_provisional_payments", "ytd_before_app_income", "ytd_before_app_deductions",
			"ptu_paid", "prior_year_losses", "interest_withholding"},
	}, jsonKeys(t, owner, "required_inputs"))

	var coeficiente string
	require.NoError(t, owner.QueryRowContext(t.Context(), `
		SELECT d::text FROM regimes, jsonb_array_elements(required_inputs) d
		WHERE code = '601' AND person_type = 'moral' AND d->>'key' = 'coeficiente_utilidad'`).Scan(&coeficiente))
	assert.JSONEq(t, `{"key": "coeficiente_utilidad", "label": "Coeficiente de utilidad", "kind": "rate",
		"required": true}`, coeficiente)
}

// insertRegime inserts a catalog row as the owner with the given JSON columns.
func insertRegime(t *testing.T, owner *sql.DB, optionDefs, requiredInputs string) error {
	t.Helper()
	_, err := owner.ExecContext(t.Context(), `
		INSERT INTO regimes (code, person_type, name, treatment, periodicity, legal_basis, source_url,
		                     valid_from, option_defs, required_inputs, income_basis)
		VALUES ('699', 'fisica', 'Prueba', 'full', 'monthly', 'LISR', 'https://x.mx', '2022-01-01', $1, $2, 'cash')`,
		optionDefs, requiredInputs)
	return err
}

func TestRegimes_AcceptWellFormedJSON(t *testing.T) {
	owner := pgtest.NewDatabase(t).Connect(t, pgtest.RoleOwner)

	err := insertRegime(t, owner, `[
		{"key": "mode", "label": "Modo", "type": "enum", "allowed_values": ["a", "b"], "default": "a", "required": true},
		{"key": "kinds", "label": "Tipos", "type": "multi_enum", "allowed_values": ["x", "y"], "default": null, "required": false},
		{"key": "factor", "label": "Factor", "type": "rate", "min": "0", "max": "1", "default": "0.1234", "required": false},
		{"key": "flag", "label": "Bandera", "type": "boolean", "required": false}
	]`, `[{"key": "coeficiente_utilidad", "label": "Coeficiente", "kind": "rate", "required": true},
		{"key": "ptu_paid", "label": "PTU pagada", "kind": "money", "required": false}]`)

	require.NoError(t, err)
}

func TestRegimes_RejectMalformedOptionDefs(t *testing.T) {
	owner := pgtest.NewDatabase(t).Connect(t, pgtest.RoleOwner)
	bad := map[string]string{
		"not an array":             `{"key": "x"}`,
		"element not an object":    `["x"]`,
		"missing label":            `[{"key": "x", "type": "boolean", "required": true}]`,
		"blank label":              `[{"key": "x", "label": " ", "type": "boolean", "required": true}]`,
		"key not snake_case":       `[{"key": "Mode", "label": "M", "type": "boolean", "required": true}]`,
		"unknown type":             `[{"key": "x", "label": "X", "type": "integer", "required": true}]`,
		"required not boolean":     `[{"key": "x", "label": "X", "type": "boolean", "required": "yes"}]`,
		"unknown property":         `[{"key": "x", "label": "X", "type": "boolean", "required": true, "min": 1}]`,
		"duplicate keys":           `[{"key": "x", "label": "X", "type": "boolean", "required": true}, {"key": "x", "label": "Y", "type": "boolean", "required": true}]`,
		"enum without values":      `[{"key": "x", "label": "X", "type": "enum", "required": true}]`,
		"enum with empty values":   `[{"key": "x", "label": "X", "type": "enum", "allowed_values": [], "required": true}]`,
		"enum repeated values":     `[{"key": "x", "label": "X", "type": "enum", "allowed_values": ["a", "a"], "required": true}]`,
		"enum non-string value":    `[{"key": "x", "label": "X", "type": "enum", "allowed_values": [1], "required": true}]`,
		"enum default not allowed": `[{"key": "x", "label": "X", "type": "enum", "allowed_values": ["a"], "default": "b", "required": true}]`,
		"multi_enum default not subset": `[{"key": "x", "label": "X", "type": "multi_enum", "allowed_values": ["a"],
			"default": ["a", "b"], "required": true}]`,
		// docs/api/openapi.yaml RegimeOptionDef.default is string | boolean | null.
		"multi_enum default array":  `[{"key": "x", "label": "X", "type": "multi_enum", "allowed_values": ["a"], "default": ["a"], "required": true}]`,
		"multi_enum default scalar": `[{"key": "x", "label": "X", "type": "multi_enum", "allowed_values": ["a"], "default": "a", "required": true}]`,
		"boolean with values":       `[{"key": "x", "label": "X", "type": "boolean", "allowed_values": ["a"], "required": true}]`,
		"boolean default string":    `[{"key": "x", "label": "X", "type": "boolean", "default": "true", "required": true}]`,
		"rate default number":       `[{"key": "x", "label": "X", "type": "rate", "min": "0", "max": "1", "default": 0.35, "required": true}]`,
		"rate default not decimal":  `[{"key": "x", "label": "X", "type": "rate", "min": "0", "max": "1", "default": "35%", "required": true}]`,
		"rate default too precise":  `[{"key": "x", "label": "X", "type": "rate", "min": "0", "max": "1", "default": "0.1234567", "required": true}]`,
		// The contract key pattern is ^[a-z][a-z0-9_]{0,63}$: 64 characters at most.
		"key longer than 64": `[{"key": "a` + strings.Repeat("b", 64) + `", "label": "X", "type": "boolean", "required": true}]`,
		// The domain (OptionDef) refuses an empty allowed value.
		"enum empty-string value": `[{"key": "x", "label": "X", "type": "enum", "allowed_values": ["", "a"], "required": true}]`,
		// A rate carries inclusive DecimalRate bounds, min <= max, and a default inside them.
		"rate without bounds":    `[{"key": "x", "label": "X", "type": "rate", "required": true}]`,
		"rate without max":       `[{"key": "x", "label": "X", "type": "rate", "min": "0", "required": true}]`,
		"rate bound number":      `[{"key": "x", "label": "X", "type": "rate", "min": 0, "max": "1", "required": true}]`,
		"rate bound not decimal": `[{"key": "x", "label": "X", "type": "rate", "min": "0", "max": "1e0", "required": true}]`,
		"rate min above max":     `[{"key": "x", "label": "X", "type": "rate", "min": "0.5", "max": "0.1", "required": true}]`,
		"rate default below min": `[{"key": "x", "label": "X", "type": "rate", "min": "0.1", "max": "0.5", "default": "0.05", "required": true}]`,
		"rate default above max": `[{"key": "x", "label": "X", "type": "rate", "min": "0.1", "max": "0.5", "default": "0.500001", "required": true}]`,
		"enum with bounds":       `[{"key": "x", "label": "X", "type": "enum", "allowed_values": ["a"], "min": "0", "max": "1", "required": true}]`,
	}
	for name, defs := range bad {
		t.Run(name, func(t *testing.T) {
			requireSQLState(t, insertRegime(t, owner, defs, `[]`), sqlstateCheckViolation)
		})
	}
}

func TestRegimes_RejectMalformedRequiredInputs(t *testing.T) {
	owner := pgtest.NewDatabase(t).Connect(t, pgtest.RoleOwner)
	bad := map[string]string{
		"not an array":         `{"key": "x"}`,
		"element not object":   `[1]`,
		"missing required":     `[{"key": "x", "label": "X", "kind": "money"}]`,
		"kind outside the API": `[{"key": "x", "label": "X", "kind": "integer", "required": true}]`,
		"blank label":          `[{"key": "x", "label": "", "kind": "money", "required": true}]`,
		"unknown property":     `[{"key": "x", "label": "X", "kind": "money", "required": true, "default": 0}]`,
		"duplicate keys":       `[{"key": "x", "label": "X", "kind": "money", "required": true}, {"key": "x", "label": "X", "kind": "rate", "required": true}]`,
		"key longer than 64":   `[{"key": "a` + strings.Repeat("b", 64) + `", "label": "X", "kind": "money", "required": true}]`,
	}
	for name, inputs := range bad {
		t.Run(name, func(t *testing.T) {
			requireSQLState(t, insertRegime(t, owner, `[]`, inputs), sqlstateCheckViolation)
		})
	}
}

func TestRegimes_RejectMalformedColumns(t *testing.T) {
	owner := pgtest.NewDatabase(t).Connect(t, pgtest.RoleOwner)
	const insert = `INSERT INTO regimes (code, person_type, name, treatment, periodicity, legal_basis,
		source_url, valid_from, valid_to, needs_cpa_review, review_note, income_basis) VALUES `
	bad := map[string]string{
		"code not 3 digits":     `('61', 'fisica', 'X', 'full', 'monthly', 'L', 'https://x.mx', '2022-01-01', NULL, false, NULL, 'cash')`,
		"unknown person type":   `('699', 'otra', 'X', 'full', 'monthly', 'L', 'https://x.mx', '2022-01-01', NULL, false, NULL, 'cash')`,
		"unknown treatment":     `('699', 'fisica', 'X', 'explanation_only', 'monthly', 'L', 'https://x.mx', '2022-01-01', NULL, false, NULL, 'cash')`,
		"unknown periodicity":   `('699', 'fisica', 'X', 'full', 'weekly', 'L', 'https://x.mx', '2022-01-01', NULL, false, NULL, 'cash')`,
		"unstated periodicity":  `('699', 'fisica', 'X', 'full', NULL, 'L', 'https://x.mx', '2022-01-01', NULL, false, NULL, 'cash')`,
		"blank name":            `('699', 'fisica', ' ', 'full', 'monthly', 'L', 'https://x.mx', '2022-01-01', NULL, false, NULL, 'cash')`,
		"blank legal basis":     `('699', 'fisica', 'X', 'full', 'monthly', '', 'https://x.mx', '2022-01-01', NULL, false, NULL, 'cash')`,
		"source not a URL":      `('699', 'fisica', 'X', 'full', 'monthly', 'L', 'LISR.pdf', '2022-01-01', NULL, false, NULL, 'cash')`,
		"ends before it starts": `('699', 'fisica', 'X', 'full', 'monthly', 'L', 'https://x.mx', '2022-01-01', '2021-12-31', false, NULL, 'cash')`,
		"review without a note": `('699', 'fisica', 'X', 'full', 'monthly', 'L', 'https://x.mx', '2022-01-01', NULL, true, NULL, 'cash')`,
		"note without a review": `('699', 'fisica', 'X', 'full', 'monthly', 'L', 'https://x.mx', '2022-01-01', NULL, false, 'n', 'cash')`,
		"blank review note":     `('699', 'fisica', 'X', 'full', 'monthly', 'L', 'https://x.mx', '2022-01-01', NULL, true, ' ', 'cash')`,
	}
	for name, values := range bad {
		t.Run(name, func(t *testing.T) {
			_, err := owner.ExecContext(t.Context(), insert+values)
			requireSQLState(t, err, sqlstateCheckViolation)
		})
	}
}

// compatibilityRules maps each pair of the matrix, "<person_type> <regime_a>/<regime_b>",
// to "allowed", "denied" or "warn" (allowed, with a non-blocking needs_cpa_review warning).
func compatibilityRules(t *testing.T, db *sql.DB) map[string]string {
	t.Helper()
	rows, err := db.QueryContext(t.Context(), `
		SELECT person_type || ' ' || regime_a || '/' || regime_b,
		       CASE WHEN NOT allowed THEN 'denied' WHEN needs_cpa_review THEN 'warn' ELSE 'allowed' END
		FROM regime_compatibility`)
	require.NoError(t, err)
	defer func() { require.NoError(t, rows.Close()) }()
	out := map[string]string{}
	for rows.Next() {
		var pair, verdict string
		require.NoError(t, rows.Scan(&pair, &verdict))
		out[pair] = verdict
	}
	require.NoError(t, rows.Err())
	return out
}

// The matrix is complete: one row per unordered pair of distinct regimes of the same
// person type, so the M2-06 use case never meets an unknown pair.
func TestRegimeCompatibility_CoversEveryPairOnce(t *testing.T) {
	owner := pgtest.NewDatabase(t).Connect(t, pgtest.RoleOwner)

	for _, person := range []string{"fisica", "moral"} {
		t.Run(person, func(t *testing.T) {
			n := countRegimes(person)
			var pairs int
			require.NoError(t, owner.QueryRowContext(t.Context(),
				"SELECT count(*) FROM regime_compatibility WHERE person_type = $1", person).Scan(&pairs))
			assert.Equal(t, n*(n-1)/2, pairs)
		})
	}
	missing := queryStrings(t, owner, `
		SELECT a.person_type || ' ' || a.code || '/' || b.code
		FROM regimes a JOIN regimes b ON b.person_type = a.person_type AND a.code < b.code
		WHERE NOT EXISTS (SELECT 1 FROM regime_compatibility c
		                  WHERE c.person_type = a.person_type AND c.regime_a = a.code AND c.regime_b = b.code)`)
	assert.Empty(t, missing)
}

func TestRegimeCompatibility_ResearchRules(t *testing.T) {
	rules := compatibilityRules(t, pgtest.NewDatabase(t).Connect(t, pgtest.RoleOwner))
	want := map[string]string{
		"fisica 612/626": "denied",  // 113-E: RESICO replaces 612 for that income; exit goes to 612 or 606
		"fisica 606/626": "denied",  // 113-E: rental income is inside RESICO
		"fisica 611/626": "denied",  // 113-E: shareholders of a PM may not use RESICO
		"fisica 605/626": "allowed", // 113-E: salary allowed with total ≤ $3.5M
		"fisica 614/626": "allowed", // 113-E: interest allowed with total ≤ $3.5M
		"fisica 612/625": "allowed", // 113-B: non-definitive withholding credited under 612
		"fisica 625/626": "allowed", // §A.2.6: "computes under 612 (or RESICO)"
		"fisica 605/612": "allowed", // Part C: salary + 612 combine in the annual return
		"fisica 605/606": "allowed",
		"fisica 606/612": "allowed",
		"moral 601/626":  "denied", // 206, 214: RESICO PM exits to Título II
		"moral 603/626":  "denied", // 206: Título III excluded
		"moral 622/626":  "denied", // 206: Cap. VIII excluded
		"moral 623/626":  "denied", // 206: Cap. VI excluded
		"moral 624/626":  "denied", // 206: Cap. VII excluded
		// M2-13 FU1 (20261014120000_regime_compatibility_reconcile.sql):
		"fisica 612/621": "denied", // 111 (until 2021): RIF only for business income without a degree
		"fisica 621/626": "denied", // 113-E; Transitorio 2º fr. IX: RIF or RESICO, not both
		"moral 601/620":  "denied", // 194: the deferring cooperative applies instead of Título II
		"moral 620/622":  "denied", // 194
		"moral 620/623":  "denied", // 194
		"moral 620/624":  "denied", // 194
		"moral 620/626":  "denied", // 206 fr. IV: Título VII Cap. VII excluded
		"moral 601/622":  "denied", // 74: AGAPES must use Cap. VIII
		"moral 622/623":  "denied", // 74: Cap. VIII not for Cap. VI groups
		"moral 622/624":  "denied", // 72, 74: two 90 % exclusivity tests
	}
	for pair, verdict := range want {
		assert.Equal(t, verdict, rules[pair], pair)
	}
	for pair, verdict := range rules {
		if containsCode(pair, "616") || containsCode(pair, "610") {
			assert.Equal(t, "denied", verdict, "%s: 616 and 610 exclude every other regime", pair)
		}
	}
}

func containsCode(pair, code string) bool {
	return pair[len(pair)-7:len(pair)-4] == code || pair[len(pair)-3:] == code
}

// Where the research is silent the pair is allowed with a needs_cpa_review warning
// (M2-06 surfaces it as a non-blocking warning): an unverified rule never refuses a
// taxpayer's regimes, and it is never silently treated as verified either.
func TestRegimeCompatibility_UndeterminedPairsAllowWithWarning(t *testing.T) {
	owner := pgtest.NewDatabase(t).Connect(t, pgtest.RoleOwner)

	var unexplained int
	require.NoError(t, owner.QueryRowContext(t.Context(), `
		SELECT count(*) FROM regime_compatibility
		WHERE needs_cpa_review AND NOT (allowed AND btrim(review_note) <> '')`).Scan(&unexplained))
	assert.Zero(t, unexplained)

	review := []string{}
	for pair, verdict := range compatibilityRules(t, owner) {
		if verdict == "warn" {
			review = append(review, pair)
		}
	}
	assert.ElementsMatch(t, []string{
		"fisica 605/621", "fisica 606/621", "fisica 607/621", "fisica 608/621", "fisica 611/621",
		"fisica 614/621", "fisica 615/621", "fisica 621/625",
		"fisica 607/626", "fisica 608/626", "fisica 615/626",
		"moral 601/603", "moral 601/623", "moral 601/624",
		"moral 603/620", "moral 603/622", "moral 603/623", "moral 603/624",
		"moral 623/624",
	}, review)
}

func TestRegimeCompatibility_RejectMalformedPairs(t *testing.T) {
	owner := pgtest.NewDatabase(t).Connect(t, pgtest.RoleOwner)
	const insert = `INSERT INTO regime_compatibility (person_type, regime_a, regime_b, allowed, reason,
		legal_basis, source_url, needs_cpa_review, review_note) VALUES `
	cases := []struct {
		name, values, state string
	}{
		{"reversed pair", `('fisica', '626', '612', false, 'r', 'L', 'https://x.mx', false, NULL)`, sqlstateCheckViolation},
		{"self pair", `('fisica', '612', '612', false, 'r', 'L', 'https://x.mx', false, NULL)`, sqlstateCheckViolation},
		{"no verdict", `('fisica', '612', '614', NULL, 'r', 'L', 'https://x.mx', true, 'n')`, sqlstateNotNullViolation},
		{"unverified refusal", `('fisica', '612', '614', false, 'r', 'L', 'https://x.mx', true, 'n')`, sqlstateCheckViolation},
		{"blank reason", `('moral', '601', '610', false, ' ', 'L', 'https://x.mx', false, NULL)`, sqlstateCheckViolation},
		{"regime of the other person type", `('moral', '601', '612', false, 'r', 'L', 'https://x.mx', false, NULL)`, sqlstateForeignKeyViolation},
		{"duplicate pair", `('fisica', '612', '626', true, 'r', 'L', 'https://x.mx', false, NULL)`, sqlstateUniqueViolation},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := owner.ExecContext(t.Context(), insert+c.values)
			requireSQLState(t, err, c.state)
		})
	}
}

const sqlstateNotNullViolation = "23502"

var catalogTables = []string{"regimes", "regime_compatibility"}

func TestRegimeCatalog_ReadOnlyForTheApp(t *testing.T) {
	db := pgtest.NewDatabase(t)
	app := db.Connect(t, pgtest.RoleApp)

	var regimes, pairs int
	require.NoError(t, app.QueryRowContext(t.Context(), "SELECT count(*) FROM regimes").Scan(&regimes))
	require.NoError(t, app.QueryRowContext(t.Context(), "SELECT count(*) FROM regime_compatibility").Scan(&pairs))
	assert.Len(t, researchCatalog, regimes)
	assert.Positive(t, pairs)

	writes := []string{
		`INSERT INTO regimes (code, person_type, name, treatment, periodicity, legal_basis, source_url, valid_from)
		 VALUES ('699', 'fisica', 'X', 'full', 'monthly', 'L', 'https://x.mx', '2022-01-01')`,
		"UPDATE regimes SET name = 'x'",
		"DELETE FROM regimes",
		"TRUNCATE regimes CASCADE",
		`INSERT INTO regime_compatibility (person_type, regime_a, regime_b, allowed, reason, legal_basis, source_url)
		 VALUES ('fisica', '612', '699', true, 'r', 'L', 'https://x.mx')`,
		"UPDATE regime_compatibility SET allowed = true",
		"DELETE FROM regime_compatibility",
		"TRUNCATE regime_compatibility",
	}
	for _, stmt := range writes {
		t.Run(stmt, func(t *testing.T) {
			_, err := app.ExecContext(t.Context(), stmt)
			requireSQLState(t, err, sqlstateInsufficientPrivilege)
		})
	}
}

func TestRegimeCatalog_NoOtherRuntimeRoleReadsIt(t *testing.T) {
	db := pgtest.NewDatabase(t)
	for _, role := range []pgtest.Role{pgtest.RoleWebhook, pgtest.RoleWorker} {
		for _, table := range catalogTables {
			t.Run(string(role)+"/"+table, func(t *testing.T) {
				_, err := db.Connect(t, role).ExecContext(t.Context(), "SELECT 1 FROM "+table)
				requireSQLState(t, err, sqlstateInsufficientPrivilege)
			})
		}
	}
}

// P0-10 pattern: ENABLE (not FORCE) with one SELECT policy for cc_app and no write
// policy, so a write granted by mistake is still refused or matches no row.
func TestRegimeCatalog_RowSecurityDeniesUngrantedWrites(t *testing.T) {
	db := pgtest.NewDatabase(t)
	owner := db.Connect(t, pgtest.RoleOwner)
	for _, table := range catalogTables {
		var enabled, forced bool
		require.NoError(t, owner.QueryRowContext(t.Context(),
			"SELECT relrowsecurity, relforcerowsecurity FROM pg_class WHERE oid = to_regclass($1)", table).
			Scan(&enabled, &forced))
		assert.True(t, enabled, "%s: RLS enabled", table)
		assert.False(t, forced, "%s: not forced, so cc_owner keeps writing through migrations", table)
	}
	_, err := owner.ExecContext(t.Context(),
		"GRANT INSERT, UPDATE, DELETE ON regimes, regime_compatibility TO cc_app")
	require.NoError(t, err)
	app := db.Connect(t, pgtest.RoleApp)

	_, err = app.ExecContext(t.Context(), `
		INSERT INTO regimes (code, person_type, name, treatment, periodicity, legal_basis, source_url, valid_from)
		VALUES ('699', 'fisica', 'X', 'full', 'monthly', 'L', 'https://x.mx', '2022-01-01')`)
	requireSQLState(t, err, sqlstateInsufficientPrivilege)

	for _, stmt := range []string{
		"UPDATE regimes SET name = 'x'", "DELETE FROM regime_compatibility",
		"UPDATE regime_compatibility SET allowed = true",
	} {
		res, err := app.ExecContext(t.Context(), stmt)
		require.NoError(t, err, stmt)
		affected, err := res.RowsAffected()
		require.NoError(t, err)
		assert.Zero(t, affected, stmt)
	}
}

func TestRegimeCatalog_DownRemovesItAndUpReseedsIt(t *testing.T) {
	owner := pgtest.NewDatabase(t).Connect(t, pgtest.RoleOwner)
	p, err := migrations.NewProvider(owner)
	require.NoError(t, err)

	_, err = p.DownTo(t.Context(), regimesCatalogVersion-1)
	require.NoError(t, err)
	for _, obj := range []string{
		"to_regclass('public.regimes')", "to_regclass('public.regime_compatibility')",
		"to_regprocedure('public.regime_option_defs_valid(jsonb)')",
		"to_regprocedure('public.regime_required_inputs_valid(jsonb)')",
		"to_regprocedure('public.regime_is_decimal_rate(jsonb)')",
	} {
		var found sql.NullString
		require.NoError(t, owner.QueryRowContext(t.Context(), "SELECT "+obj+"::text").Scan(&found))
		assert.False(t, found.Valid, "down left %s", found.String)
	}

	require.NoError(t, migrations.Up(t.Context(), owner))
	require.NoError(t, migrations.Up(t.Context(), owner), "a second up is a no-op")
	var regimes int
	require.NoError(t, owner.QueryRowContext(t.Context(), "SELECT count(*) FROM regimes").Scan(&regimes))
	assert.Len(t, researchCatalog, regimes, "seeds applied exactly once")
}

// compatibilityReconcileVersion is the goose version of
// 20261014120000_regime_compatibility_reconcile.sql (M2-13 FU1).
const compatibilityReconcileVersion = 20261014120000

var reconciledRefusals = []string{
	"fisica 612/621", "fisica 621/626",
	"moral 601/620", "moral 620/622", "moral 620/623", "moral 620/624", "moral 620/626",
	"moral 601/622", "moral 622/623", "moral 622/624",
}

// Down puts the M2-02 seed verdict (allowed with a warning) back; up refuses again.
func TestRegimeCompatibilityReconcile_DownRestoresTheSeedAndUpRefusesAgain(t *testing.T) {
	owner := pgtest.NewDatabase(t).Connect(t, pgtest.RoleOwner)
	p, err := migrations.NewProvider(owner)
	require.NoError(t, err)

	_, err = p.DownTo(t.Context(), compatibilityReconcileVersion-1)
	require.NoError(t, err)
	rules := compatibilityRules(t, owner)
	for _, pair := range reconciledRefusals {
		assert.Equal(t, "warn", rules[pair], "down: %s", pair)
	}

	require.NoError(t, migrations.Up(t.Context(), owner))
	rules = compatibilityRules(t, owner)
	for _, pair := range reconciledRefusals {
		assert.Equal(t, "denied", rules[pair], "up: %s", pair)
	}
}

// rent606QuarterlyVersion is the goose version of
// 20261014160000_regime_606_pays_quarterly.sql (M5-12).
const rent606QuarterlyVersion = 20261014160000

// Down removes the 606 quarterly option and leaves the rest of the row; up adds it
// back once, so a stored 606 regime keeps validating against the catalog.
func TestRegime606PaysQuarterly_DownRemovesItAndUpAddsItOnce(t *testing.T) {
	owner := pgtest.NewDatabase(t).Connect(t, pgtest.RoleOwner)
	p, err := migrations.NewProvider(owner)
	require.NoError(t, err)
	keys606 := func() []string { return jsonKeys(t, owner, "option_defs")["606/fisica"] }

	_, err = p.DownTo(t.Context(), rent606QuarterlyVersion-1)
	require.NoError(t, err)
	assert.Equal(t, []string{"deduction_mode", "iva_border_region"}, keys606())

	require.NoError(t, migrations.Up(t.Context(), owner))
	require.NoError(t, migrations.Up(t.Context(), owner), "a second up is a no-op")
	assert.Equal(t, []string{"deduction_mode", "iva_border_region", "pays_quarterly"}, keys606())
}
