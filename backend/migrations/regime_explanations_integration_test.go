//go:build integration

package migrations_test

import (
	"database/sql"
	"regexp"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/platform/pgtest"
	"github.com/romodevx/control-contable/backend/migrations"
)

// Regime explanations (M5-19): the plain-language es-MX text of every regime the
// engine does not fully estimate, rendered for the domain label key
// regime_explanation:<code>/<person_type> (MODULE-tax-engine §2.2, §2.4).

// regimeExplanationsVersion is the goose version of 20261014170000_regime_explanations.sql.
const regimeExplanationsVersion = 20261014170000

// explainedRegimes are the informational (explanation-only) regimes and the
// partial ones whose explanation says what is not estimated.
var explainedRegimes = []string{
	"605/fisica", "607/fisica", "608/fisica", "610/fisica", "610/moral", "611/fisica",
	"614/fisica", "615/fisica", "616/fisica", "620/moral", "623/moral", "624/moral",
	"621/fisica", "622/moral", "603/moral",
}

// untouchedRegimes are full regimes this migration must not write.
var untouchedRegimes = []string{"601/moral", "612/fisica", "626/fisica", "626/moral"}

// maxExplanationRunes keeps the text a short card paragraph, not a tax guide.
const maxExplanationRunes = 480

// legalCitation: every text names its legal basis (a LISR or CFF article).
var legalCitation = regexp.MustCompile(`\b(LISR|CFF) arts?\. [0-9]`)

func explanations(t *testing.T, db *sql.DB) map[string]sql.NullString {
	t.Helper()
	rows, err := db.QueryContext(t.Context(),
		`SELECT code || '/' || person_type, explanation FROM regimes`)
	require.NoError(t, err)
	defer func() { require.NoError(t, rows.Close()) }()
	out := map[string]sql.NullString{}
	for rows.Next() {
		var key string
		var text sql.NullString
		require.NoError(t, rows.Scan(&key, &text))
		out[key] = text
	}
	require.NoError(t, rows.Err())
	return out
}

func TestRegimeExplanations_EveryNotFullyEstimatedRegimeIsExplained(t *testing.T) {
	app := pgtest.NewDatabase(t).Connect(t, pgtest.RoleApp)
	got := explanations(t, app)

	for _, key := range explainedRegimes {
		text := got[key]
		if !assert.True(t, text.Valid, "%s has no explanation", key) {
			continue
		}
		assert.LessOrEqual(t, utf8.RuneCountInString(text.String), maxExplanationRunes, "%s is not short", key)
		assert.Regexp(t, legalCitation, text.String, "%s cites no article", key)
		assert.Contains(t, text.String, "estima", "%s must say what the app does or does not estimate", key)
	}
	for _, key := range untouchedRegimes {
		assert.False(t, got[key].Valid, "%s is fully estimated: no explanation", key)
	}
}

// Down clears exactly the texts it wrote; up writes them again.
func TestRegimeExplanations_DownClearsThemAndUpRestoresThem(t *testing.T) {
	owner := pgtest.NewDatabase(t).Connect(t, pgtest.RoleOwner)
	p, err := migrations.NewProvider(owner)
	require.NoError(t, err)
	written := explanations(t, owner)

	_, err = p.DownTo(t.Context(), regimeExplanationsVersion-1)
	require.NoError(t, err)
	cleared := explanations(t, owner)
	for _, key := range explainedRegimes {
		assert.False(t, cleared[key].Valid, "down left %s", key)
	}

	require.NoError(t, migrations.Up(t.Context(), owner))
	assert.Equal(t, written, explanations(t, owner))
}
