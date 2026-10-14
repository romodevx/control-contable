package domain_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/dashboard/domain"
)

func month(t *testing.T, raw string) domain.Month {
	t.Helper()
	m, err := domain.ParseMonth(raw)
	require.NoError(t, err)
	return m
}

func TestParseMonth(t *testing.T) {
	for _, raw := range []string{"2026-01", "2027-12", "0001-01"} {
		m, err := domain.ParseMonth(raw)
		require.NoError(t, err, raw)
		assert.Equal(t, raw, m.String())
	}
	for _, raw := range []string{"", "2026-1", "2026-13", "2026-00", "26-01", "2026/01", "2026-01-01", "abcd-01", "+026-01"} {
		_, err := domain.ParseMonth(raw)
		require.ErrorIs(t, err, domain.ErrInvalidMonth, raw)
	}
}

func TestMonthArithmetic(t *testing.T) {
	m := month(t, "2027-02")
	assert.Equal(t, "2026-03", m.AddMonths(-11).String())
	assert.Equal(t, "2027-03", m.AddMonths(1).String())
	assert.Equal(t, 2027, m.Year())
	assert.Equal(t, time.February, m.Month())
	assert.True(t, month(t, "2026-12").Before(m))
	assert.False(t, m.Before(m))
	assert.Equal(t, m, domain.MonthOf(time.Date(2027, time.February, 28, 23, 0, 0, 0, time.UTC)))
}

// MODULE-dashboard §2.1: [max(period − (months − 1), 2026-01), period], oldest first.
func TestWindow(t *testing.T) {
	for _, tc := range []struct {
		period string
		months int
		first  string
		points int
	}{
		{"2026-10", 12, "2026-01", 10},
		{"2026-01", 12, "2026-01", 1},
		{"2027-06", 12, "2026-07", 12},
		{"2027-06", 6, "2027-01", 6},
		{"2026-12", 12, "2026-01", 12},
		{"2027-01", 12, "2026-02", 12},
	} {
		got := domain.Window(month(t, tc.period), tc.months)
		require.Len(t, got, tc.points, "%s/%d", tc.period, tc.months)
		assert.Equal(t, tc.first, got[0].String())
		assert.Equal(t, tc.period, got[len(got)-1].String())
		for i := 1; i < len(got); i++ {
			assert.Equal(t, got[i-1].AddMonths(1), got[i], "consecutive months")
		}
	}
}

// MODULE-dashboard §2.2: informational (every regime) > incomplete > partial
// > complete; none without a regime.
func TestTaxpayerStatus(t *testing.T) {
	const (
		full = domain.StatusComplete
		inc  = domain.StatusIncomplete
		part = domain.StatusPartial
		info = domain.StatusInformational
	)
	for _, tc := range []struct {
		regimes []domain.Status
		want    domain.Status
	}{
		{nil, domain.StatusNone},
		{[]domain.Status{full}, full},
		{[]domain.Status{info}, info},
		{[]domain.Status{info, info}, info},
		{[]domain.Status{info, full}, full},
		{[]domain.Status{full, part}, part},
		{[]domain.Status{part, inc, full}, inc},
		{[]domain.Status{info, inc}, inc},
		{[]domain.Status{info, part}, part},
	} {
		assert.Equal(t, tc.want, domain.TaxpayerStatus(tc.regimes), "%v", tc.regimes)
	}
}

// OQ-D6: whole days left, rounded up, at least 1 while the trial runs.
func TestTrialDaysLeft(t *testing.T) {
	ends := time.Date(2026, time.October, 20, 16, 0, 0, 0, time.UTC)
	for _, tc := range []struct {
		left time.Duration
		want int
	}{
		{7 * 24 * time.Hour, 7},
		{3*24*time.Hour - time.Second, 3},
		{2*24*time.Hour + time.Second, 3},
		{time.Hour, 1},
		{time.Nanosecond, 1},
		{0, 1},
		{-time.Hour, 1},
	} {
		assert.Equal(t, tc.want, domain.TrialDaysLeft(ends.Add(-tc.left), ends), "%s left", tc.left)
	}
}
