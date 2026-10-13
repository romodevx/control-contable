package money

import (
	"math"
	"math/big"
	"math/rand/v2"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMulDiv(t *testing.T) {
	tests := []struct {
		name    string
		a, b, d int64
		want    int64
		wantErr error
	}{
		{"exact product", 6, 7, 1, 42, nil},
		{"zero factor", 0, math.MaxInt64, 3, 0, nil},
		{"below half rounds toward zero", 14, 1, 10, 1, nil},
		{"tie rounds away from zero", 15, 1, 10, 2, nil},
		{"above half rounds away from zero", 16, 1, 10, 2, nil},
		{"negative below half", -14, 1, 10, -1, nil},
		{"negative tie rounds away from zero", -15, 1, 10, -2, nil},
		{"negative divisor-side tie", 15, -1, 10, -2, nil},
		{"two negatives make a positive tie", -15, -1, 10, 2, nil},
		{"odd divisor never ties", 1, 1, 3, 0, nil},
		{"two thirds of one", 2, 1, 3, 1, nil},
		{"max int64 identity", math.MaxInt64, 1, 1, math.MaxInt64, nil},
		{"min int64 identity", math.MinInt64, 1, 1, math.MinInt64, nil},
		{"min int64 times minus one overflows", math.MinInt64, -1, 1, 0, ErrOverflow},
		{"max int64 squared then divided back", math.MaxInt64, math.MaxInt64, math.MaxInt64, math.MaxInt64, nil},
		{"min int64 squared is 2^126, over 2^62 is 2^64", math.MinInt64, math.MinInt64, 1 << 62, 0, ErrOverflow},
		{"128-bit intermediate survives", math.MaxInt64, 1_000_000, 1_000_000, math.MaxInt64, nil},
		{"quotient above int64 overflows", math.MaxInt64, 2, 1, 0, ErrOverflow},
		{"quotient above 64 bits overflows", math.MaxInt64, math.MaxInt64, 1, 0, ErrOverflow},
		{"max int64 doubled then halved", math.MaxInt64, 2, 2, math.MaxInt64, nil},
		// (2^32-1)(2^32+1) = 2^64-1, so /2 = MaxInt64 remainder 1: the tie rounds to 2^63.
		{"positive tie rounding past max overflows", 4_294_967_295, 4_294_967_297, 2, 0, ErrOverflow},
		{"negative tie rounding lands exactly on min", -4_294_967_295, 4_294_967_297, 2, math.MinInt64, nil},
		// 274177 * 67280421310721 = 2^64+1, so /2 = 2^63 remainder 1: the tie rounds to 2^63+1.
		{"negative tie rounding past min overflows", -274_177, 67_280_421_310_721, 2, 0, ErrOverflow},
		// 31 * 1190112520884487201 = 2^65-1, so /2 = 2^64-1 remainder 1: the round-up
		// must report overflow, not wrap the 64-bit quotient to zero.
		{"round-up of a 2^64-1 quotient overflows", 31, 1_190_112_520_884_487_201, 2, 0, ErrOverflow},
		{"negative round-up of a 2^64-1 quotient overflows", -31, 1_190_112_520_884_487_201, 2, 0, ErrOverflow},
		{"exactly min int64 from a negative product", math.MinInt64 / 2, 2, 1, math.MinInt64, nil},
		{"exact near min", math.MinInt64 + 1, 2, 2, math.MinInt64 + 1, nil},
		{"zero divisor", 1, 1, 0, 0, ErrInvalidDivisor},
		{"negative divisor", 1, 1, -1, 0, ErrInvalidDivisor},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := mulDiv(tt.a, tt.b, tt.d)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestMulDivAgainstBigInt is a property test with a fixed seed: mulDiv agrees
// with an arbitrary-precision reference on random operands, both in value and
// in when it reports overflow.
func TestMulDivAgainstBigInt(t *testing.T) {
	const iterations = 20_000
	rng := rand.New(rand.NewPCG(2026, 10)) //nolint:gosec // a fixed, reproducible seed is the point; nothing here is secret
	operand := func() int64 {
		switch rng.IntN(4) {
		case 0:
			return rng.Int64N(2_000_001) - 1_000_000
		case 1:
			return int64(rng.Uint64()) //nolint:gosec // deliberate full-range reinterpretation for the property
		case 2:
			return rng.Int64N(1 << 32)
		default:
			return -rng.Int64N(1 << 40)
		}
	}
	for range iterations {
		a, b := operand(), operand()
		d := rng.Int64N(math.MaxInt64) + 1
		if rng.IntN(2) == 0 {
			d = rng.Int64N(1_000_000) + 1
		}
		want, inRange := referenceMulDiv(a, b, d)
		got, err := mulDiv(a, b, d)
		if !inRange {
			require.ErrorIsf(t, err, ErrOverflow, "mulDiv(%d, %d, %d)", a, b, d)
			continue
		}
		require.NoErrorf(t, err, "mulDiv(%d, %d, %d)", a, b, d)
		require.Equalf(t, want, got, "mulDiv(%d, %d, %d)", a, b, d)
	}
}

// referenceMulDiv computes round-half-away-from-zero(a*b/d) with big.Int.
func referenceMulDiv(a, b, d int64) (int64, bool) {
	product := new(big.Int).Mul(big.NewInt(a), big.NewInt(b))
	divisor := big.NewInt(d)
	quotient, remainder := new(big.Int).QuoRem(new(big.Int).Abs(product), divisor, new(big.Int))
	if new(big.Int).Lsh(remainder, 1).Cmp(divisor) >= 0 {
		quotient.Add(quotient, big.NewInt(1))
	}
	if product.Sign() < 0 {
		quotient.Neg(quotient)
	}
	return quotient.Int64(), quotient.IsInt64()
}

func TestCheckedAdd(t *testing.T) {
	tests := []struct {
		name    string
		a, b    int64
		want    int64
		wantErr bool
	}{
		{"small", 2, 3, 5, false},
		{"mixed signs never overflow", math.MaxInt64, math.MinInt64, -1, false},
		{"up to max", math.MaxInt64 - 1, 1, math.MaxInt64, false},
		{"past max", math.MaxInt64, 1, 0, true},
		{"down to min", math.MinInt64 + 1, -1, math.MinInt64, false},
		{"past min", math.MinInt64, -1, 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := checkedAdd(tt.a, tt.b)
			if tt.wantErr {
				require.ErrorIs(t, err, ErrOverflow)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCheckedSub(t *testing.T) {
	tests := []struct {
		name    string
		a, b    int64
		want    int64
		wantErr bool
	}{
		{"small", 2, 3, -1, false},
		{"zero minus min overflows", 0, math.MinInt64, 0, true},
		{"minus one minus min", -1, math.MinInt64, math.MaxInt64, false},
		{"past min", math.MinInt64, 1, 0, true},
		{"past max", math.MaxInt64, -1, 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := checkedSub(tt.a, tt.b)
			if tt.wantErr {
				require.ErrorIs(t, err, ErrOverflow)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
