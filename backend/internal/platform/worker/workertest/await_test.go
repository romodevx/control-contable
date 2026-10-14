package workertest_test

import (
	"fmt"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/platform/worker/workertest"
)

// shortLimit keeps the hang tests fast; nothing else in them depends on time.
const shortLimit = 10 * time.Millisecond

// recordingTB captures Fatalf and ends the calling goroutine, as testing.T does.
type recordingTB struct {
	testing.TB
	fatal string
}

func (r *recordingTB) Helper() {}

func (r *recordingTB) Fatalf(format string, args ...any) {
	r.fatal = fmt.Sprintf(format, args...)
	runtime.Goexit()
}

// runAsTest runs body on its own goroutine, as testing runs a test, so that
// Fatalf's Goexit ends body and not the caller.
func runAsTest(t *testing.T, body func(tb testing.TB)) *recordingTB {
	t.Helper()
	tb := &recordingTB{TB: t}
	done := make(chan struct{})
	go func() {
		defer close(done)
		body(tb)
	}()
	<-done
	return tb
}

func TestAwaitWithin_returnsTheValueReceived(t *testing.T) {
	c := make(chan int, 1)
	c <- 7

	assert.Equal(t, 7, workertest.AwaitWithin(t, shortLimit, c, "a value"))
}

func TestAwaitWithin_failsAHangWithWhatWasAwaitedAndEveryStack(t *testing.T) {
	tb := runAsTest(t, func(tb testing.TB) {
		workertest.AwaitWithin(tb, shortLimit, make(chan int), "the leader to run the job")
	})

	require.NotEmpty(t, tb.fatal, "a hang must fail the test")
	assert.Contains(t, tb.fatal, "the leader to run the job")
	assert.Contains(t, tb.fatal, "goroutine ", "the failure carries every goroutine's stack")
}

func TestTicks_aTickNobodyReceivesFailsWithinTheLimit(t *testing.T) {
	ticks := workertest.NewTicks(workertest.WithWaitLimit(shortLimit))

	tb := runAsTest(t, func(tb testing.TB) { ticks.Tick(tb, time.Minute) })

	assert.Contains(t, tb.fatal, "a runner to receive the 1m0s tick")
	assert.Contains(t, tb.fatal, "goroutine ")
}
