package taxengine_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/taxengine"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain"
)

// Every regime with a computation is registered once (ADR-0002 §2); the
// registry refuses a duplicate key at startup.
func TestStrategiesCoverTheLaunchRegimes(t *testing.T) {
	want := []string{
		"601/moral", "603/moral", "606/fisica", "612/fisica", "621/fisica", "622/moral", "626/fisica", "626/moral",
	}
	var keys []string
	for _, s := range taxengine.Strategies() {
		keys = append(keys, s.Key().String())
	}

	assert.ElementsMatch(t, want, keys)
	registry, err := domain.NewRegistry(taxengine.Strategies()...)
	require.NoError(t, err)
	for _, s := range taxengine.Strategies() {
		_, registered := registry.Lookup(s.Key())
		assert.True(t, registered, s.Key().String())
	}
}
