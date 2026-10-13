package contracttest_test

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/adapter/http/contracttest"
)

func TestScanCoverage_findsOpLiteralsAcrossBuildTagsAndAliases(t *testing.T) {
	got, err := contracttest.ScanCoverage(filepath.Join("testdata", "scan", "ok"))
	require.NoError(t, err)

	assert.ElementsMatch(t, []string{"getHealth", "listEnabledFlags"}, keys(got),
		"another package's Op and anything under testdata/ do not count")
	assert.Len(t, got["getHealth"], 2, "every call site is reported")
	assert.Contains(t, got["listEnabledFlags"][0], filepath.Join("aliased", "aliased_integration_test.go"))
}

func TestScanCoverage_skipsExcludedDirectories(t *testing.T) {
	root := filepath.Join("testdata", "scan", "ok")

	got, err := contracttest.ScanCoverage(root, filepath.Join(root, "plain"))
	require.NoError(t, err)

	assert.ElementsMatch(t, []string{"listEnabledFlags"}, keys(got))
}

// Coverage is tracked statically, so an operationId must be readable from the
// source: a computed one would let an operation look covered (or not) at random.
func TestScanCoverage_rejectsANonLiteralOperationID(t *testing.T) {
	_, err := contracttest.ScanCoverage(filepath.Join("testdata", "scan", "bad"))

	require.Error(t, err)
	assert.Contains(t, err.Error(), "nonliteral_test.go")
	assert.Contains(t, err.Error(), "string literal")
}

func keys(m map[string][]string) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
