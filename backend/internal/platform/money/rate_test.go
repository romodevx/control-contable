package money

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseRate(t *testing.T) {
	tests := []struct {
		in      string
		want    Rate
		wantErr error
	}{
		{"0.16", 160_000, nil},
		{"0.08", 80_000, nil},
		{"0.2136", 213_600, nil},
		{"0.0125", 12_500, nil},
		{"0.1234", 123_400, nil},
		{"0.000001", 1, nil},
		{"1", RateScale, nil},
		{"0", 0, nil},
		{"-0.5", -500_000, nil},
		{"9223372036854.775807", math.MaxInt64, nil},
		{"-9223372036854.775808", math.MinInt64, nil},
		{"9223372036854.775808", 0, ErrOverflow},
		{"0.0000001", 0, ErrPrecision},
		{"16%", 0, ErrSyntax},
		{"", 0, ErrSyntax},
		{"0.", 0, ErrSyntax},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			got, err := ParseRate(tt.in)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				assert.Contains(t, err.Error(), "money: parse rate")
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestRateString(t *testing.T) {
	tests := []struct {
		r    Rate
		want string
	}{
		{0, "0"},
		{RateScale, "1"},
		{160_000, "0.16"},
		{213_600, "0.2136"},
		{1, "0.000001"},
		{-12_500, "-0.0125"},
		{2_500_000, "2.5"},
		{math.MaxInt64, "9223372036854.775807"},
		{math.MinInt64, "-9223372036854.775808"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.r.String())
		})
	}
}

// TestRateStringRoundTrip: the canonical string form parses back to the same Rate.
func TestRateStringRoundTrip(t *testing.T) {
	for _, r := range []Rate{0, 1, -1, 160_000, 213_600, RateScale, math.MaxInt64, math.MinInt64} {
		got, err := ParseRate(r.String())
		require.NoError(t, err)
		assert.Equal(t, r, got)
	}
}
