package money

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// pesos converts whole pesos to Exact for readable test tables.
func pesos(p int64) Exact { return Exact(p * ExactPerPeso) }

func TestExactAddSub(t *testing.T) {
	tests := []struct {
		name    string
		op      func(Exact, Exact) (Exact, error)
		a, b    Exact
		want    Exact
		wantErr error
	}{
		{"add", Exact.Add, pesos(5_570), pesos(1_580), pesos(7_150), nil},
		{"add overflows", Exact.Add, math.MaxInt64, 1, 0, ErrOverflow},
		{"sub", Exact.Sub, pesos(4_800), pesos(1_600), pesos(3_200), nil},
		{"sub overflows", Exact.Sub, math.MinInt64, 1, 0, ErrOverflow},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.op(tt.a, tt.b)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestExactMulRate(t *testing.T) {
	tests := []struct {
		name    string
		e       Exact
		r       Rate
		want    Exact
		wantErr error
	}{
		// PF-612-ISR-01: (60,000 - 52,600.93) x 21.36% = 7,399.07 x 0.2136 = 1,580.441352.
		{"612 marginal excess", 7_399_070_000, 213_600, 1_580_441_352, nil},
		// 30% of a PM taxable result (PM-626-ISR-01: 400,000 x 30% = 120,000).
		{"PM rate", pesos(400_000), 300_000, pesos(120_000), nil},
		{"tie at 10^-6 peso rounds away from zero", 5, 100_000, 1, nil},
		{"negative tie rounds away from zero", -5, 100_000, -1, nil},
		{"below the tie rounds toward zero", 4, 100_000, 0, nil},
		{"identity at 100%", math.MaxInt64, RateScale, math.MaxInt64, nil},
		{"min at 100%", math.MinInt64, RateScale, math.MinInt64, nil},
		{"max at 200% overflows", math.MaxInt64, 2 * RateScale, 0, ErrOverflow},
		{"min at -100% overflows", math.MinInt64, -RateScale, 0, ErrOverflow},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.e.MulRate(tt.r)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestExactMulRatio(t *testing.T) {
	tests := []struct {
		name     string
		e        Exact
		num, den int64
		want     Exact
		wantErr  error
	}{
		// PF-612-IVA-01: IVA withheld = 2/3 x 1,600 = 1,066.666667 (Exact), 1,066.67 shown.
		{"two thirds of IVA", pesos(1_600), 2, 3, 1_066_666_667, nil},
		// PM-601-ISR-01: PTU 120,000 / 8 a month x 2 parts = 30,000.
		{"PTU parts", pesos(120_000), 2, 8, pesos(30_000), nil},
		{"negative two thirds", -pesos(1_600), 2, 3, -1_066_666_667, nil},
		{"zero denominator", pesos(1), 1, 0, 0, ErrInvalidDivisor},
		{"overflow", math.MaxInt64, 3, 2, 0, ErrOverflow},
		// 31 * 1190112520884487201 / 2 = 2^64 - 0.5: rounding must not wrap to 0.
		{"overflow when rounding crosses 2^64", 31, 1_190_112_520_884_487_201, 2, 0, ErrOverflow},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.e.MulRatio(tt.num, tt.den)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestExactRound(t *testing.T) {
	tests := []struct {
		name string
		e    Exact
		want Cents
	}{
		{"zero", 0, 0},
		{"whole centavos", 1_066_670_000, 1_066_67},
		{"below half a centavo", 4_999, 0},
		{"exactly half a centavo rounds away from zero", 5_000, 1},
		{"above half a centavo", 5_001, 1},
		{"negative below half", -4_999, 0},
		{"negative tie rounds away from zero", -5_000, -1},
		{"negative above half", -5_001, -1},
		// PF-612-IVA-01: 1,066.666667 -> 1,066.67.
		{"two thirds of 1,600", 1_066_666_667, 1_066_67},
		// PF-612-ISR-01: 7,150.961352 -> 7,150.96.
		{"612 March ISR", 7_150_961_352, 7_150_96},
		// PF-612-ISR-02: 34,516.542824 -> 34,516.54.
		{"612 October ISR", 34_516_542_824, 34_516_54},
		{"max", math.MaxInt64, math.MaxInt64/ExactPerCent + 1},
		{"min", math.MinInt64, math.MinInt64/ExactPerCent - 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.e.Round())
		})
	}
}

// TestExactRoundProperties checks, over every residue of one centavo on both
// signs, that rounding is symmetric (Round(-e) = -Round(e)), never moves more
// than half a centavo, and that whole centavos are fixed points.
func TestExactRoundProperties(t *testing.T) {
	const base = Exact(123_456 * ExactPerCent)
	for residue := range Exact(ExactPerCent) {
		e := base + residue
		got := e.Round()
		assert.Equal(t, -got, (-e).Round(), "symmetry at %d", e)
		back, err := got.Exact()
		require.NoError(t, err)
		distance := back - e
		if distance < 0 {
			distance = -distance
		}
		assert.LessOrEqual(t, 2*distance, Exact(ExactPerCent), "moved more than half a centavo at %d", e)
		if residue == 0 {
			assert.Equal(t, e, back, "whole centavo must be a fixed point")
		}
	}
}
