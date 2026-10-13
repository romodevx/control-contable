package money

import (
	"math"
	"testing"
)

// sink keeps the compiler from eliding the benchmarked work.
var sink Exact

func BenchmarkExactMulRate(b *testing.B) {
	cases := []struct {
		name string
		e    Exact
		r    Rate
	}{
		{"typical", 7_399_070_000, 213_600},
		{"wide intermediate", math.MaxInt64 / 2, 1_999_999},
		{"negative tie", -5, 100_000},
	}
	for _, c := range cases {
		b.Run(c.name, func(b *testing.B) {
			for b.Loop() {
				sink, _ = c.e.MulRate(c.r)
			}
		})
	}
}

func BenchmarkCentsMulRate(b *testing.B) {
	c, r := Cents(60_000_00), Rate(15_000)
	for b.Loop() {
		sink, _ = c.MulRate(r)
	}
}

func BenchmarkExactMulRatio(b *testing.B) {
	e := Exact(1_600 * ExactPerPeso)
	for b.Loop() {
		sink, _ = e.MulRatio(2, 3)
	}
}
