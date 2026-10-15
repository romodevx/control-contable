package taxengine_test

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/taxengine/domain"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain/taxparams"
	"github.com/romodevx/control-contable/backend/internal/taxengine/golden"
)

// goldenFiguresDigest pins, per EngineVersion, the SHA-256 of what the
// engine computes for every golden input (authored and generated, skipped
// ones included: the engine still runs them for users). The records
// fingerprint hashes the inputs and EngineVersion, not the figures, so a
// strategy that changes a figure for the same records without a version
// bump leaves M8's "changed since viewed" silent; the strategy-list pin
// (TestEngineVersionPinsTheStrategyList) misses that case.
//
// When this fails, the change moved a figure, a line or a status: bump
// domain.EngineVersion and pin the new digest in the same commit. Label,
// argument and citation wording is not hashed.
var goldenFiguresDigest = map[string]string{
	"1.0.1": "2825ad18d43022db573a9bd194ae78c2a5add51ecd8dda032b7f09ef5e0a1f14",
	"1.1.0": "25e7e526f5dba7dab3f2456c2c3fc17f9ce8aa28ba88461bd3e9c1d3fafe4128",
}

// TestEngineVersionPinsTheGoldenFigures (M5-30).
func TestEngineVersionPinsTheGoldenFigures(t *testing.T) {
	cases := slices.Concat(authoredCases(t), generatedCases(t, rawParams(t)))
	reg, cat := registry(t), catalog(t)

	var text strings.Builder
	ran := 0
	for _, c := range cases {
		_, registered := reg.Lookup(c.Input.Key)
		if !registered && c.Want.Status != domain.StatusInformational {
			continue // pending_strategy: nothing computes it yet
		}
		ran++
		writeFigures(&text, c, reg, cat)
	}
	require.Positive(t, ran)
	sum := sha256.Sum256([]byte(text.String()))
	got := hex.EncodeToString(sum[:])

	want, pinned := goldenFiguresDigest[domain.EngineVersion]
	require.True(t, pinned, "EngineVersion %s has no pinned golden figures digest", domain.EngineVersion)
	assert.Equal(t, want, got, "the engine's figures changed for the same inputs: bump domain.EngineVersion and pin the new digest")
}

// writeFigures appends c's computed figures to w: what a user sees change,
// not how it is worded.
func writeFigures(w *strings.Builder, c golden.Case, reg domain.Registry, cat *taxparams.Catalog) {
	fmt.Fprintf(w, "case %s %s %s\n", c.ID, c.Input.Key, c.Input.Period)
	selection, err := cat.Select(c.Input.Period.Year())
	if err != nil {
		fmt.Fprintf(w, "params error %v\n", err)
		return
	}
	est, err := reg.Estimate(c.Input, slices.Clone(c.ActiveMonths), selection.Params)
	if err != nil {
		fmt.Fprintf(w, "error %v\n", err)
		return
	}
	fmt.Fprintf(w, "status %s coverage %s explanation %s\n", est.Status(), est.Coverage, est.Explanation)
	if est.ISR != nil {
		fmt.Fprintf(w, "isr %s\n", est.ISR.Amount)
	}
	if est.IVA != nil {
		fmt.Fprintf(w, "iva %+v\n", *est.IVA)
	}
	fmt.Fprintf(w, "to_remit %s missing %v topics %v disclaimers %v\n",
		est.WithheldToRemit, est.MissingInputs, est.CPATopics, est.Disclaimers)
	for _, l := range est.Lines {
		fmt.Fprintf(w, "line %d %s %s %s %s rate %s row %d %s cpa %t\n",
			l.Seq, l.Tax, l.Code, l.Operation, l.Amount, l.Rate, l.Row, l.Source, l.NeedsCPAReview)
	}
}
