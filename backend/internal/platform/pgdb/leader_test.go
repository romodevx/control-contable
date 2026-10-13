package pgdb_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/platform/pgdb"
)

func TestNewAdvisoryLeader_requiresAPool(t *testing.T) {
	_, err := pgdb.NewAdvisoryLeader(nil, pgdb.WorkerLeaderLockKey)

	require.Error(t, err)
}
