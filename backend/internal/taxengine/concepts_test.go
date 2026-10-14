package taxengine_test

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	taxenginehttp "github.com/romodevx/control-contable/backend/internal/taxengine/adapter/http"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain"
)

// Every line a shipped strategy or the IVA consolidation can append has a
// Spanish concept in the adapter (contract BreakdownLine.concept): the
// breakdown never shows a key or the generic fallback. It runs every golden
// case, authored and generated, through the composition root's registry.
func TestEveryBreakdownLineHasAConcept(t *testing.T) {
	cases := slices.Concat(authoredCases(t), generatedCases(t, rawParams(t)))
	reg, cat := registry(t), catalog(t)
	labels := map[domain.LabelKey]string{}
	for _, c := range cases {
		sel, err := cat.Select(c.Input.Period.Year())
		if err != nil {
			continue // a case about missing parameters
		}
		est, err := reg.Estimate(c.Input, slices.Clone(c.ActiveMonths), sel.Params)
		if err != nil {
			continue // the golden gate reports strategy failures
		}
		for _, l := range est.Lines {
			labels[l.Label] = c.ID
		}
	}
	netted, err := domain.ConsolidateIVA([]domain.IVAContribution{
		{Regime: "612", Components: domain.IVAComponents{ChargedCollected: 3, CreditablePaid: 1, DefinitiveExcluded: 1}},
		{Regime: "606", Components: domain.IVAComponents{ChargedCollected: 1, WithheldFromTaxpayer: 1}},
	}, 5)
	require.NoError(t, err)
	for _, l := range netted.Lines {
		labels[l.Label] = "ConsolidateIVA"
	}

	require.NotEmpty(t, labels)
	for label, seenIn := range labels {
		assert.Truef(t, taxenginehttp.HasConcept(label), "label %q (first seen in %s) has no concept in texts.go", label, seenIn)
	}
	t.Logf("%d distinct labels rendered", len(labels))
}
