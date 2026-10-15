package taxengine_test

import (
	"errors"
	"fmt"
	"io/fs"
	"maps"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/taxengine/annualgolden"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain/annual"
	"github.com/romodevx/control-contable/backend/internal/taxengine/golden"
)

// The annual golden gate (PLAN ANN-06, MODULE-annual §3). It runs in
// `go test ./...` (CI `make test`), so a mismatch, an unlisted skip or a
// pending count above the baseline fails CI; `make golden` prints its
// summary. There is no flag that updates an expected value.

const (
	annualGoldenRoot     = "annual_golden" // under testdata/
	annualBaselineFile   = "testdata/annual_golden_baseline.yaml"
	annualResearchFile   = "../../../docs/research/annual-return-2026.md"
	annualResearchSource = "research:"
)

// gateAnnualEstimator is the annual engine the gate runs: the sections the
// strategy PRs (ANN-08..ANN-11) register, each from its own file. ANN-14
// replaces it with the composition root's estimator.
func gateAnnualEstimator(t *testing.T) annualgolden.Estimator {
	sections := annualSections{}
	maps.Copy(sections, pmGateSections(t)) // ANN-10, ANN-11
	return byPersonType{fisica: newPFAnnualSections(t), moral: sections}
}

// byPersonType runs the PF sections (ANN-08, ANN-09, which also decide the
// estimate-level informative-income fact and disclaimers) for a persona
// física and the registered PM sections for a persona moral.
type byPersonType struct {
	fisica pfAnnualSections
	moral  annualSections
}

func (e byPersonType) Registered(kind annual.SectionKind) bool {
	return e.fisica.Registered(kind) || e.moral.Registered(kind)
}

func (e byPersonType) Estimate(in annualgolden.Input) (annualgolden.Result, error) {
	if in.PersonType == fiscal.PersonFisica {
		return e.fisica.Estimate(in)
	}
	return e.moral.Estimate(in)
}

// annualSectionOrder is the fixed order of an estimate's sections
// (MODULE-annual §2.1).
var annualSectionOrder = []annual.SectionKind{
	annual.SectionPFArt152, annual.SectionPFRESICOCheck, annual.SectionPM601, annual.SectionPMRESICO, annual.SectionExplanation,
}

// annualSection computes one section kind for a case's year: none when the
// year holds no regime of its kind.
type annualSection func(annualgolden.Input) ([]annualgolden.SectionResult, error)

// annualSections are the registered sections by kind.
type annualSections map[annual.SectionKind]annualSection

func (s annualSections) Registered(kind annual.SectionKind) bool {
	_, ok := s[kind]
	return ok
}

// errNoAnnualSection is never reached: Evaluate calls Estimate only when
// every section a case expects is registered.
var errNoAnnualSection = errors.New("no annual section is registered")

func (s annualSections) Estimate(in annualgolden.Input) (annualgolden.Result, error) {
	if len(s) == 0 {
		return annualgolden.Result{}, errNoAnnualSection
	}
	var out annualgolden.Result
	for _, kind := range annualSectionOrder {
		compute, ok := s[kind]
		if !ok {
			continue
		}
		sections, err := compute(in)
		if err != nil {
			return annualgolden.Result{}, fmt.Errorf("%s: %w", kind, err)
		}
		out.Sections = append(out.Sections, sections...)
	}
	return out, nil
}

func annualCases(t *testing.T) []annualgolden.Case {
	t.Helper()
	cases, err := annualgolden.Load(os.DirFS("testdata"), annualGoldenRoot)
	require.NoError(t, err)
	return cases
}

// TestGoldenAnnualCases runs every annual case. Skipped (⚠️) and
// pending_strategy cases are listed in the log with their reasons, and may
// only go down (testdata/annual_golden_baseline.yaml).
func TestGoldenAnnualCases(t *testing.T) {
	cases := annualCases(t)

	s := annualgolden.Run(t, cases, gateAnnualEstimator(t))

	assert.Equal(t, len(cases), s.Total())
	data, err := os.ReadFile(annualBaselineFile)
	require.NoError(t, err)
	b, err := golden.ReadBaseline(data)
	require.NoError(t, err)
	assert.Empty(t, s.Ratchet(b),
		"skipped and pending_strategy cases may only go down; edit %s by hand in the PR that changes them", annualBaselineFile)
}

// TestGoldenAnnualSourcesAreTheResearchWorkedExamples ties every research
// case to a worked example of Part D, and every worked example to a case.
func TestGoldenAnnualSourcesAreTheResearchWorkedExamples(t *testing.T) {
	doc, err := os.ReadFile(annualResearchFile)
	require.NoError(t, err)
	examples := map[string]bool{}
	for _, m := range regexp.MustCompile(`(?m)^\| \*\*(AN-[0-9]{2})\*\* \|`).FindAllStringSubmatch(string(doc), -1) {
		examples[m[1]] = false
	}
	require.NotEmpty(t, examples)

	for _, c := range annualCases(t) {
		id, research := strings.CutPrefix(c.Source, annualResearchSource)
		if !research {
			continue
		}
		_, known := examples[id]
		assert.True(t, known, "%s: source %s is not a worked example of %s", c.File, id, annualResearchFile)
		examples[id] = true
	}
	for id, seeded := range examples {
		assert.True(t, seeded, "worked example %s has no annual golden case", id)
	}
}

// TestGoldenAnnualFilesKeepTheirHeader is the annual half of
// TestGoldensAreNeverRewritten (which scans the whole module for rewrite
// paths): every annual case still carries its hand-authored header.
func TestGoldenAnnualFilesKeepTheirHeader(t *testing.T) {
	err := fs.WalkDir(os.DirFS("testdata"), annualGoldenRoot, func(name string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || strings.HasSuffix(name, "README.md") {
			return err
		}
		data, err := fs.ReadFile(os.DirFS("testdata"), name)
		if err != nil {
			return err
		}
		assert.True(t, strings.HasPrefix(string(data), golden.HandAuthoredHeader), "%s lost its header", name)
		return nil
	})
	require.NoError(t, err)
}
