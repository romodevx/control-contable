package plain

import (
	"testing"

	"github.com/romodevx/control-contable/backend/adapter/http/contracttest"
)

func TestPlain(t *testing.T) {
	_ = contracttest.Op("getHealth")
	_ = contracttest.Op("getHealth")
}
