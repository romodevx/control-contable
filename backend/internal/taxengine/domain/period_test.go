package domain_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/taxengine/domain"
)

func TestParsePeriod(t *testing.T) {
	tests := []struct {
		raw   string
		year  int
		month time.Month
		end   string
	}{
		{raw: "2026-01", year: 2026, month: time.January, end: "2026-01-31"},
		{raw: "2026-02", year: 2026, month: time.February, end: "2026-02-28"},
		{raw: "2028-02", year: 2028, month: time.February, end: "2028-02-29"},
		{raw: "2026-04", year: 2026, month: time.April, end: "2026-04-30"},
		{raw: "2026-12", year: 2026, month: time.December, end: "2026-12-31"},
		{raw: "0001-01", year: 1, month: time.January, end: "0001-01-31"},
	}
	for _, tt := range tests {
		t.Run(tt.raw, func(t *testing.T) {
			p, err := domain.ParsePeriod(tt.raw)
			require.NoError(t, err)
			assert.Equal(t, tt.year, p.Year())
			assert.Equal(t, tt.month, p.Month())
			assert.Equal(t, tt.end, p.End().String())
			assert.Equal(t, tt.raw, p.String())
			assert.False(t, p.IsZero())
		})
	}
}

func TestParsePeriodRejects(t *testing.T) {
	for _, raw := range []string{"", "2026-1", "2026-001", "2026/01", "2026-00", "2026-13", "0000-01", "20x6-01", "2026-0a", "+026-01", "2026-01-01"} {
		t.Run(raw, func(t *testing.T) {
			_, err := domain.ParsePeriod(raw)
			require.ErrorIs(t, err, domain.ErrInvalidPeriod)
		})
	}
}

func TestNewPeriod(t *testing.T) {
	p, err := domain.NewPeriod(2026, time.March)
	require.NoError(t, err)
	assert.Equal(t, "2026-03", p.String())

	for _, tt := range []struct {
		year  int
		month time.Month
	}{{2026, 0}, {2026, 13}, {0, time.March}, {10_000, time.March}, {-1, time.March}} {
		_, err := domain.NewPeriod(tt.year, tt.month)
		require.ErrorIs(t, err, domain.ErrInvalidPeriod, "%d-%d", tt.year, tt.month)
	}
}

func TestZeroPeriod(t *testing.T) {
	var p domain.Period
	assert.True(t, p.IsZero())
	assert.Empty(t, p.String())
	assert.True(t, p.End().IsZero())
}
