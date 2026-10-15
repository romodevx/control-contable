//go:build integration

package infra_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/platform/pgtest"
	"github.com/romodevx/control-contable/backend/internal/taxpayers/domain"
	"github.com/romodevx/control-contable/backend/internal/taxpayers/usecase"
)

// The catalog listing (listRegimes) over the seeded M2-02 catalog, as cc_app.
func TestCatalogList_OverTheSeededCatalog(t *testing.T) {
	db := pgtest.NewDatabase(t)
	owner := db.Connect(t, pgtest.RoleOwner)
	// Flag one more pair as under CPA review, as the catalog owner would,
	// to see it listed on both regimes.
	_, err := owner.ExecContext(t.Context(), `UPDATE regime_compatibility
SET needs_cpa_review = true, review_note = 'Pendiente de confirmar con un contador.'
WHERE person_type = 'fisica' AND regime_a = '605' AND regime_b = '625'`)
	require.NoError(t, err)
	e := newEnvOn(t, db, userA)
	uc := usecase.NewCatalog(e.deps())

	fisica, err := uc.List(t.Context(), userA, "fisica")
	require.NoError(t, err)
	moral, err := uc.List(t.Context(), userA, "moral")
	require.NoError(t, err)

	byCode := entriesByCode(t, fisica)
	require.Contains(t, byCode, "626")
	require.Contains(t, byCode, "606")
	assert.Contains(t, links(byCode["626"].IncompatibleWith), "612", "a refused pair, on both sides")
	assert.Contains(t, links(byCode["612"].IncompatibleWith), "626")
	assert.Contains(t, links(byCode["605"].ReviewPendingWith), "625")
	assert.Contains(t, links(byCode["625"].ReviewPendingWith), "605")
	assert.NotContains(t, links(byCode["605"].IncompatibleWith), "625")

	arrendamiento := byCode["606"]
	assert.Equal(t, domain.PersonTypeFisica, arrendamiento.PersonType)
	assert.Equal(t, "full", arrendamiento.Treatment)
	assert.Equal(t, "cash", arrendamiento.IncomeBasis, "regimes.income_basis")
	require.NotEmpty(t, arrendamiento.OptionDefs)
	def := arrendamiento.OptionDefs[0]
	assert.Equal(t, "deduction_mode", def.Key)
	assert.NotEmpty(t, def.Label, "the label travels to the options form")
	assert.Equal(t, domain.OptionTypeEnum, def.Type)
	assert.ElementsMatch(t, []string{"blind_35", "actual"}, def.AllowedValues)
	text, ok := def.Default.Text()
	assert.True(t, ok)
	assert.Equal(t, "blind_35", text)

	informational := byCode["605"]
	assert.Equal(t, "informational", informational.Treatment)

	codes := make([]string, 0, len(fisica))
	for _, r := range fisica {
		codes = append(codes, r.Code)
	}
	assert.IsIncreasing(t, codes, "ordered by code")
	for _, r := range moral {
		assert.Equal(t, domain.PersonTypeMoral, r.PersonType)
		assert.NotNil(t, r.OptionDefs)
		assert.NotNil(t, r.RequiredInputs)
	}
	assert.NotEmpty(t, moral)
	moralByCode := entriesByCode(t, moral)
	assert.Equal(t, "accrual", moralByCode["601"].IncomeBasis, "Título II books on the invoice date")
	assert.Equal(t, "cash", moralByCode["626"].IncomeBasis, "RESICO PM books on the collection date")
}

func entriesByCode(t *testing.T, entries []usecase.CatalogEntry) map[string]usecase.CatalogEntry {
	t.Helper()
	out := make(map[string]usecase.CatalogEntry, len(entries))
	for _, e := range entries {
		out[e.Code] = e
	}
	return out
}

func links(in []usecase.RegimeLink) []string {
	out := make([]string, 0, len(in))
	for _, l := range in {
		out = append(out, l.Code)
	}
	return out
}
