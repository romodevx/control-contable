package gen_test

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/romodevx/control-contable/backend/adapter/http/gen"
)

// TestPlatformOperationsAreAllowListed pins which operations skip the flag gate
// (PR #19 follow-up 4). x-module: platform means "never hidden", so labelling a
// new operation platform by mistake would ship it ungated in every preset.
// Adding one here is a security-review decision, not a codegen detail.
func TestPlatformOperationsAreAllowListed(t *testing.T) {
	allowed := []string{"getHealth", "listEnabledFlags"}

	var platform []string
	for operationID, module := range gen.OperationModules() {
		if module == gen.ModulePlatform {
			platform = append(platform, operationID)
		}
	}
	slices.Sort(platform)

	assert.Equal(t, allowed, platform, "operations with x-module: platform")
}
