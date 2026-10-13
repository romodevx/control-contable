package worker_test

import (
	"testing"
	"time"

	"github.com/romodevx/control-contable/backend/internal/platform/worker"
)

func TestSystemTicker_ticksUntilStopped(t *testing.T) {
	ticker := worker.SystemTicker(time.Millisecond)
	defer ticker.Stop()

	select {
	case <-ticker.C():
	case <-t.Context().Done():
		t.Fatal("no tick")
	}
}
