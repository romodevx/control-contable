package worker_test

import (
	"testing"

	"go.uber.org/goleak"
)

// Every Runner a test starts must be gone when the package's tests end: a
// scheduling or campaign goroutine that outlives Run is a leak.
func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}
