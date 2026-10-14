//go:build integration

package infra_test

import (
	"fmt"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/platform/apperr"
	"github.com/romodevx/control-contable/backend/internal/taxpayers/domain"
	"github.com/romodevx/control-contable/backend/internal/taxpayers/usecase"
)

// pairVerdict is the final compatibility of a pair that MODULE-taxpayers §2.5
// and the M2-02 seed disagreed on (M2-13 FU1). refused: 422
// invalid_regime_combination; review: created with the non-blocking
// regime_combination_needs_cpa_review warning.
type pairVerdict int

const (
	refused pairVerdict = iota
	review
)

// reconciledPairs is the decision table of the PR that settled FU1; each
// verdict cites the law (docs/research/tax-regimes-mx-2026.md Part E).
var reconciledPairs = []struct {
	person domain.PersonType
	a, b   string
	want   pairVerdict
}{
	{domain.PersonTypeFisica, "612", "621", refused}, // LISR art. 111 ¶1 and ¶15 (until 2021)
	{domain.PersonTypeFisica, "621", "626", refused}, // LISR art. 113-E ¶1 and ¶6; Transitorio 2º fr. IX
	{domain.PersonTypeMoral, "601", "620", refused},  // LISR art. 194 ¶1
	{domain.PersonTypeMoral, "620", "622", refused},  // LISR art. 194 ¶1
	{domain.PersonTypeMoral, "620", "623", refused},  // LISR art. 194 ¶1
	{domain.PersonTypeMoral, "620", "624", refused},  // LISR art. 194 ¶1
	{domain.PersonTypeMoral, "620", "626", refused},  // LISR art. 206 fr. IV
	{domain.PersonTypeMoral, "601", "622", refused},  // LISR art. 74 ¶1
	{domain.PersonTypeMoral, "622", "623", refused},  // LISR art. 74 ¶2
	{domain.PersonTypeMoral, "622", "624", refused},  // LISR arts. 72 and 74: each needs ≥ 90 % of income
	{domain.PersonTypeMoral, "601", "603", review},   // LISR art. 80: Título III owes Título II tax on some income
	{domain.PersonTypeMoral, "601", "623", review},   // LISR art. 59 ¶4: Cap. VI layers on Título II
	{domain.PersonTypeMoral, "601", "624", review},   // LISR art. 72: optional, not stated exclusive of Título II
	{domain.PersonTypeMoral, "603", "620", review},   // no rule found
	{domain.PersonTypeMoral, "603", "622", review},   // no rule found
	{domain.PersonTypeMoral, "603", "623", review},   // no rule found
	{domain.PersonTypeMoral, "603", "624", review},   // no rule found
	{domain.PersonTypeMoral, "623", "624", review},   // no rule found
}

// Every reconciled pair, through createTaxpayer as cc_app: the use case reads
// the catalog verdict in the same transaction that writes the taxpayer.
func TestCreate_ReconciledRegimeCompatibility(t *testing.T) {
	e := newEnv(t, userA).wire(t)

	for n, p := range reconciledPairs {
		t.Run(fmt.Sprintf("%s %s/%s", p.person, p.a, p.b), func(t *testing.T) {
			in := usecase.CreateInput{
				PersonType: string(p.person), RFC: rfcOf(p.person, n+1), DisplayName: "Contribuyente",
				Regimes: []usecase.RegimeInput{regimeInput(p.a), regimeInput(p.b)},
			}

			created, err := e.taxpayers.Create(t.Context(), userA, in)

			if p.want == refused {
				requireCode(t, err, apperr.CodeInvalidRegimeCombination)
				return
			}
			require.NoError(t, err)
			codes := make([]string, 0, len(created.Taxpayer.Warnings))
			for _, w := range created.Taxpayer.Warnings {
				codes = append(codes, w.Code)
			}
			assert.True(t, slices.Contains(codes, usecase.WarningRegimeNeedsCPAReview), "warnings: %v", codes)
		})
	}
}

// The other write path, addTaxpayerRegime: a refused pair answers 422 and a
// pair still under CPA review is added with the warning on regime_code.
func TestAddRegime_ReconciledRegimeCompatibility(t *testing.T) {
	e := newEnv(t, userA).wire(t)
	rif := createFisica(t, e, regimeInput("621"))

	_, err := e.regimes.Add(t.Context(), userA, rif, regimeInput("612"))
	requireCode(t, err, apperr.CodeInvalidRegimeCombination)
	_, err = e.regimes.Add(t.Context(), userA, rif, regimeInput("626"))
	requireCode(t, err, apperr.CodeInvalidRegimeCombination)

	added, err := e.regimes.Add(t.Context(), userA, rif, regimeInput("606"))
	require.NoError(t, err)
	require.Len(t, added.Warnings, 1)
	assert.Equal(t, usecase.WarningRegimeNeedsCPAReview, added.Warnings[0].Code)
	assert.Equal(t, "regime_code", added.Warnings[0].Field)
}

// rfcOf returns a distinct, well-formed RFC of the person type per n (1..28).
func rfcOf(person domain.PersonType, n int) string {
	if person == domain.PersonTypeMoral {
		return fmt.Sprintf("A&B0203%02dXY2", n)
	}
	return rfcFor(n)
}

// regimeInput sends the options a regime requires with no default (621 rif_year).
func regimeInput(code string) usecase.RegimeInput {
	in := usecase.RegimeInput{Code: code, Start: "2026-01"}
	if code == "621" {
		in.Options = map[string]domain.OptionValue{"rif_year": domain.StringOption("6")}
	}
	return in
}
