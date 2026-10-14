package archtest

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckRepositoriesReportsTransactionControlAndSessionState(t *testing.T) {
	got, err := CheckRepositories(filepath.Join("testdata", "repository"))
	require.NoError(t, err)

	const file = "internal/ledger/infra/store.go"
	want := []RepositoryViolation{
		{File: file, Line: 5, What: "call Begin"},
		{File: file, Line: 6, What: "call BeginTx"},
		{File: file, Line: 7, What: "call Commit"},
		{File: file, Line: 8, What: "call Rollback"},
		{File: file, Line: 9, What: "call SavePoint"},
		{File: file, Line: 9, What: "call RollbackTo"},
		{File: file, Line: 10, What: "call Transaction"},
		{File: file, Line: 12, What: "session statement in SQL"},
		{File: file, Line: 14, What: "session statement in SQL"},
		{File: file, Line: 16, What: "session statement in SQL"},
		{File: file, Line: 18, What: "session statement in SQL"},
		{File: file, Line: 20, What: "session statement in SQL"},
	}
	assert.Equal(t, want, got)
}

func TestBackendRepositoriesLeaveTheUnitOfWorkToUseCases(t *testing.T) {
	got, err := CheckRepositories(filepath.Join("..", ".."))
	require.NoError(t, err)
	assert.Empty(t, got, "repositories never begin/commit/roll back or set session state (P0-30 D-3)")
}
