package money

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCentsAddSub(t *testing.T) {
	tests := []struct {
		name    string
		op      func(Cents, Cents) (Cents, error)
		a, b    Cents
		want    Cents
		wantErr error
	}{
		{"add", Cents.Add, 4_800_00, -1_600_00, 3_200_00, nil},
		{"add overflows", Cents.Add, math.MaxInt64, 1, 0, ErrOverflow},
		{"sub", Cents.Sub, 9_600_00, 2_000_00, 7_600_00, nil},
		{"sub overflows", Cents.Sub, math.MinInt64, 1, 0, ErrOverflow},
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

func TestCentsExact(t *testing.T) {
	tests := []struct {
		name    string
		c       Cents
		want    Exact
		wantErr error
	}{
		{"one centavo", 1, ExactPerCent, nil},
		{"negative", -1_066_67, -1_066_67 * ExactPerCent, nil},
		{"largest convertible", math.MaxInt64 / ExactPerCent, math.MaxInt64 / ExactPerCent * ExactPerCent, nil},
		{"overflow", math.MaxInt64/ExactPerCent + 1, 0, ErrOverflow},
		{"negative overflow", math.MinInt64/ExactPerCent - 1, 0, ErrOverflow},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.c.Exact()
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCentsMulRate(t *testing.T) {
	tests := []struct {
		name    string
		c       Cents
		r       Rate
		want    Exact
		wantErr error
	}{
		// PF-626-ISR-01: 60,000.00 x 1.5% = 900.00; 20,000.00 x 1.25% = 250.00.
		{"RESICO 1.5%", 60_000_00, 15_000, 900_000_000, nil},
		{"RESICO PM withholding 1.25%", 20_000_00, 12_500, 250_000_000, nil},
		// PM-601-ISR-01: 0.15 x 5,000,000.00 = 750,000.00.
		{"coeficiente de utilidad", 5_000_000_00, 150_000, 750_000_000_000, nil},
		// 0.01 x 0.000001 = 0.00000001 peso, which is below Exact's 10^-6 peso.
		{"sub-scale product rounds to zero", 1, 1, 0, nil},
		// 0.50 x 0.000001 = 0.0000005 peso: a tie at Exact scale.
		{"sub-scale tie rounds away from zero", 50, 1, 1, nil},
		{"negative sub-scale tie", -50, 1, -1, nil},
		{"negative amount", -60_000_00, 15_000, -900_000_000, nil},
		{"negative rate", 60_000_00, -15_000, -900_000_000, nil},
		{"zero rate", math.MaxInt64, 0, 0, nil},
		{"max cents at 100% overflows Exact", math.MaxInt64, RateScale, 0, ErrOverflow},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.c.MulRate(tt.r)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestParseCents(t *testing.T) {
	tests := []struct {
		in      string
		want    Cents
		wantErr error
	}{
		{"3439.46", 3_439_46, nil},
		{"0.01", 1, nil},
		{"844.5", 844_50, nil},
		{"25000", 25_000_00, nil},
		{"-0.05", -5, nil},
		{"-0", 0, nil},
		{"007.10", 7_10, nil},
		{"92233720368547758.07", math.MaxInt64, nil},
		{"-92233720368547758.08", math.MinInt64, nil},
		{"92233720368547758.08", 0, ErrOverflow},
		{"-92233720368547758.09", 0, ErrOverflow},
		{"1.005", 0, ErrPrecision},
		{"", 0, ErrSyntax},
		{"-", 0, ErrSyntax},
		{"+1", 0, ErrSyntax},
		{"1.", 0, ErrSyntax},
		{".5", 0, ErrSyntax},
		{"1,000.00", 0, ErrSyntax},
		{"1.2.3", 0, ErrSyntax},
		{"1e3", 0, ErrSyntax},
		{" 1", 0, ErrSyntax},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			got, err := ParseCents(tt.in)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				assert.Contains(t, err.Error(), "money: parse cents")
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCentsString(t *testing.T) {
	tests := []struct {
		c    Cents
		want string
	}{
		{0, "0.00"},
		{1, "0.01"},
		{-5, "-0.05"},
		{2_133_33, "2133.33"},
		{-1_066_67, "-1066.67"},
		{math.MaxInt64, "92233720368547758.07"},
		{math.MinInt64, "-92233720368547758.08"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.c.String())
		})
	}
}

// TestCentsStringRoundTrip: String and ParseCents are inverses over the boundaries.
func TestCentsStringRoundTrip(t *testing.T) {
	for _, c := range []Cents{0, 1, -1, 99, -100, 3_439_46, math.MaxInt64, math.MinInt64} {
		got, err := ParseCents(c.String())
		require.NoError(t, err)
		assert.Equal(t, c, got)
	}
}
