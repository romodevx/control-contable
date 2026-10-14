package taxengine_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/taxengine"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain"
)

// launchStrategyKeys are the (regime, person type) pairs with a computation
// at EngineVersion 1.0.0 and 1.0.1 (1.0.1: 626 PM remits the ISR it
// withheld; same strategy list, different figures).
var launchStrategyKeys = []string{
	"601/moral", "603/moral", "606/fisica", "612/fisica", "621/fisica", "622/moral", "625/fisica", "626/fisica", "626/moral",
}

func strategyKeys() []string {
	var keys []string
	for _, s := range taxengine.Strategies() {
		keys = append(keys, s.Key().String())
	}
	return keys
}

// Every regime with a computation is registered once (ADR-0002 §2); the
// registry refuses a duplicate key at startup.
func TestStrategiesCoverTheLaunchRegimes(t *testing.T) {
	assert.ElementsMatch(t, launchStrategyKeys, strategyKeys())
	registry, err := domain.NewRegistry(taxengine.Strategies()...)
	require.NoError(t, err)
	for _, s := range taxengine.Strategies() {
		_, registered := registry.Lookup(s.Key())
		assert.True(t, registered, s.Key().String())
	}
}

// 625 PF (M5-14, PR #94) resolves to its own strategy, not the
// informational fallback.
func TestRegistryResolves625PF(t *testing.T) {
	registry, err := domain.NewRegistry(taxengine.Strategies()...)
	require.NoError(t, err)

	s, registered := registry.Lookup(domain.StrategyKey{Regime: fiscal.RegimePlatforms, PersonType: fiscal.PersonFisica})

	require.True(t, registered)
	assert.IsType(t, domain.NewStrategy625PF(), s)
}

// A strategy added or removed changes figures for the same records, so the
// engine version, which the records fingerprint hashes, must move with it:
// M8 marks a month "changed since viewed" only when the fingerprint moves.
// Change this pin and EngineVersion in the same commit.
func TestEngineVersionPinsTheStrategyList(t *testing.T) {
	pinned := map[string][]string{"1.0.1": launchStrategyKeys}

	want, ok := pinned[domain.EngineVersion]

	require.True(t, ok, "EngineVersion %s has no pinned strategy list", domain.EngineVersion)
	assert.ElementsMatch(t, want, strategyKeys(), "the strategy list changed: bump domain.EngineVersion")
}
