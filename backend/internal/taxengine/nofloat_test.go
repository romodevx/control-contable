package taxengine_test

import (
	"fmt"
	"go/types"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/tools/go/packages"
)

// noFloatPatterns are the packages whose production code computes or
// carries money and rates: the engine, the money type and the shared
// fiscal kernel (ADR-0001 O7, ADR-0002).
var noFloatPatterns = []string{
	"github.com/romodevx/control-contable/backend/internal/taxengine/...",
	"github.com/romodevx/control-contable/backend/internal/platform/money/...",
	"github.com/romodevx/control-contable/backend/internal/kernel/...",
}

// floatProbe is a fixture package with one float of each form forbidigo
// misses (testdata: never built, linted or listed by ./...).
const floatProbe = "./testdata/floatprobe"

// TestNoFloatingPointInMoneyCode (M5-30) type-checks the production code of
// noFloatPatterns and fails on any expression of a floating-point type,
// typed or untyped. forbidigo matches only the identifiers float32 and
// float64, so an inferred `x := 0.5`, strconv.ParseFloat or math.Round
// pass it; the type checker sees every one of them.
func TestNoFloatingPointInMoneyCode(t *testing.T) {
	pkgs := loadTyped(t, noFloatPatterns...)
	require.NotEmpty(t, pkgs)

	assert.Empty(t, floatExpressions(pkgs), "money and rates are int64 centavos / fixed-point Rate, never floats")
}

// TestNoFloatGuardSeesEveryFloatForm proves the guard can fail.
func TestNoFloatGuardSeesEveryFloatForm(t *testing.T) {
	got := floatExpressions(loadTyped(t, floatProbe))

	for _, form := range []string{"0.5", "strconv.ParseFloat(s, 64)", "math.Round(v)", "1.0 / 3"} {
		found := slices.ContainsFunc(got, func(v string) bool { return strings.HasSuffix(v, ": "+form) })
		assert.True(t, found, "the guard misses %q in %v", form, got)
	}
}

// loadTyped loads the production (non-test) files of patterns, type-checked.
func loadTyped(t *testing.T, patterns ...string) []*packages.Package {
	t.Helper()
	const mode = packages.NeedName | packages.NeedFiles | packages.NeedSyntax | packages.NeedTypes | packages.NeedTypesInfo
	pkgs, err := packages.Load(&packages.Config{Mode: mode}, patterns...)
	require.NoError(t, err)
	for _, p := range pkgs {
		for _, e := range p.Errors {
			t.Errorf("%s: %v", p.PkgPath, e)
		}
	}
	return pkgs
}

// floatExpressions lists "file:line: expr" for every expression of a
// floating-point (or complex) type in pkgs.
func floatExpressions(pkgs []*packages.Package) []string {
	var out []string
	for _, p := range pkgs {
		for expr, tv := range p.TypesInfo.Types {
			if !isFloat(tv.Type) {
				continue
			}
			pos := p.Fset.Position(expr.Pos())
			out = append(out, fmt.Sprintf("%s:%d: %s", pos.Filename, pos.Line, types.ExprString(expr)))
		}
	}
	slices.Sort(out)
	return out
}

// isFloat reports a floating-point or complex type, or a tuple (a
// multi-value call such as strconv.ParseFloat) holding one.
func isFloat(t types.Type) bool {
	if tuple, ok := t.(*types.Tuple); ok {
		for v := range tuple.Variables() {
			if isFloat(v.Type()) {
				return true
			}
		}
		return false
	}
	basic, ok := t.Underlying().(*types.Basic)
	return ok && basic.Info()&(types.IsFloat|types.IsComplex) != 0
}
