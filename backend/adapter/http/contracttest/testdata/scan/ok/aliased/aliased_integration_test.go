//go:build integration

package aliased

import (
	"testing"

	ct "github.com/romodevx/control-contable/backend/adapter/http/contracttest"
)

func TestAliased(t *testing.T) {
	_ = ct.Op("listEnabledFlags")
}
