package other

import (
	"testing"

	contracttest "example.com/not/the/harness"
)

func Op(string) int { return 0 }

func TestOther(t *testing.T) {
	_ = Op("getMe")
	_ = contracttest.Op("deleteMe")
}
