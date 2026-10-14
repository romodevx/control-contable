package workertest

import (
	"runtime"
	"testing"
	"time"
)

// WaitLimit bounds every wait of a test on a runner: a hang fails the test
// within it, with every goroutine's stack, instead of running into go test's
// -timeout, which kills the whole binary. It is a failure bound only; no
// passing test waits anywhere near it.
const WaitLimit = 30 * time.Second

// stackBufSize is the first buffer allStacks tries; it doubles until the dump fits.
const stackBufSize = 1 << 16

// Await waits for a value on c for up to WaitLimit; what names the awaited
// event in the failure.
func Await[T any](tb testing.TB, c <-chan T, what string) T {
	tb.Helper()
	return AwaitWithin(tb, WaitLimit, c, what)
}

// AwaitWithin waits for a value on c for up to limit and fails tb with what
// and every goroutine's stack otherwise. It does not watch tb.Context, which
// is already done in Cleanup, where stopping a runner is awaited too.
func AwaitWithin[T any](tb testing.TB, limit time.Duration, c <-chan T, what string) T {
	tb.Helper()
	timer := time.NewTimer(limit)
	defer timer.Stop()
	select {
	case v := <-c:
		return v
	case <-timer.C:
		failHung(tb, limit, what)
		var zero T
		return zero
	}
}

func failHung(tb testing.TB, limit time.Duration, what string) {
	tb.Helper()
	tb.Fatalf("workertest: waited %s for %s\n\n%s", limit, what, allStacks())
}

func allStacks() []byte {
	buf := make([]byte, stackBufSize)
	for {
		n := runtime.Stack(buf, true)
		if n < len(buf) {
			return buf[:n]
		}
		buf = make([]byte, 2*len(buf))
	}
}
