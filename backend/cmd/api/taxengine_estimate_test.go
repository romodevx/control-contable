package main

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/platform/ratelimit/ratelimitpg"
	taxdomain "github.com/romodevx/control-contable/backend/internal/taxengine/domain"
	taxusecase "github.com/romodevx/control-contable/backend/internal/taxengine/usecase"
)

func taxPeriod(t *testing.T, raw string) taxdomain.Period {
	t.Helper()
	p, err := taxdomain.ParsePeriod(raw)
	require.NoError(t, err)
	return p
}

// The estimate needs the taxpayers module's unit of work: its profile read
// joins the estimate's transaction only through the same manager.
func TestNewTaxEngineModule_requiresTheSharedUnitOfWork(t *testing.T) {
	_, err := newTaxEngineModule(taxEngineWiring{pool: &sql.DB{}, limits: &ratelimitpg.Store{}})

	assert.ErrorIs(t, err, errNoTaxEngineTx)
}

// The estimate's due date needs the holidays snapshot read at boot
// (loadHolidays): without it the boot fails, never a weekends-only date.
func TestNewTaxEngineModule_requiresTheHolidays(t *testing.T) {
	_, err := newTaxEngineModule(taxEngineWiring{pool: &sql.DB{}, limits: &ratelimitpg.Store{}, tx: noopTx{}, taxpayers: fakeProfiles{}})

	assert.ErrorIs(t, err, taxusecase.ErrMissingDependency)
}
