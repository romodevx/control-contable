package archtest

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckWeekendRuleReportsWeekendTestsOutsideTheKernel(t *testing.T) {
	got, err := CheckWeekendRule(filepath.Join("testdata", "weekend"))
	require.NoError(t, err)

	want := []WeekendViolation{
		{File: "internal/billing/usecase/due.go", Line: 6, What: "time.Saturday"},
		{File: "internal/billing/usecase/due.go", Line: 6, What: "time.Sunday"},
		{File: "internal/reports/domain/aliased.go", Line: 6, What: "time.Saturday"},
	}
	assert.Equal(t, want, got)
}

// ADR-0007 Amendment 1 / MODULE-reminders §2.1: the due-date rule has one
// copy, fiscal.NextBusinessDay. A weekend test anywhere else is a second
// copy in the making.
func TestBackendWeekendRuleLivesOnlyInTheKernel(t *testing.T) {
	got, err := CheckWeekendRule(filepath.Join("..", ".."))
	require.NoError(t, err)
	assert.Empty(t, got, "time.Saturday/time.Sunday only in internal/kernel (fiscal.NextBusinessDay)")
}
