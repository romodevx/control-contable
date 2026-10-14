package domain_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/taxpayers/domain"
)

func TestParsePeriod(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		wantErr bool
	}{
		{name: "january", raw: "2026-01"},
		{name: "december", raw: "2022-12"},
		{name: "catalog floor", raw: "2022-01"},
		{name: "month 13", raw: "2026-13", wantErr: true},
		{name: "month 00", raw: "2026-00", wantErr: true},
		{name: "single-digit month", raw: "2026-1", wantErr: true},
		{name: "day included", raw: "2026-01-01", wantErr: true},
		{name: "slash separator", raw: "2026/01", wantErr: true},
		{name: "empty", raw: "", wantErr: true},
		{name: "non-digit year", raw: "20x6-01", wantErr: true},
		{name: "sign in year", raw: "+026-01", wantErr: true},
		{name: "surrounding space", raw: " 2026-01", wantErr: true},
		{name: "year zero", raw: "0000-01", wantErr: true},
		{name: "non-ascii digit", raw: "２026-01", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := domain.ParsePeriod(tt.raw)
			if tt.wantErr {
				require.ErrorIs(t, err, domain.ErrInvalidPeriod)
				assert.True(t, got.IsZero())
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.raw, got.String())
			assert.False(t, got.IsZero())
		})
	}
}

func TestPeriodOf_UsesTheInstantsOwnCalendar(t *testing.T) {
	mexico, err := time.LoadLocation("America/Mexico_City")
	require.NoError(t, err)
	// 2026-11-01 05:00 UTC is still 31 October in Mexico City (UTC-6).
	instant := time.Date(2026, time.November, 1, 5, 0, 0, 0, time.UTC).In(mexico)
	assert.Equal(t, "2026-10", domain.PeriodOf(instant).String())
}

func TestPeriodCompare(t *testing.T) {
	jan := mustPeriod(t, "2026-01")
	dec := mustPeriod(t, "2025-12")
	feb := mustPeriod(t, "2026-02")

	assert.True(t, dec.Before(jan))
	assert.True(t, jan.Before(feb))
	assert.False(t, jan.Before(jan))
	assert.True(t, jan.After(dec))
	assert.False(t, jan.After(jan))
	assert.Equal(t, 0, jan.Compare(mustPeriod(t, "2026-01")))
	assert.Equal(t, -1, dec.Compare(jan))
	assert.Equal(t, 1, feb.Compare(jan))
}

func TestPeriodZeroValue(t *testing.T) {
	var zero domain.Period
	assert.True(t, zero.IsZero())
	assert.Empty(t, zero.String())
}

func TestWindowOverlaps(t *testing.T) {
	var open domain.Period
	tests := []struct {
		name string
		a, b domain.Window
		want bool
	}{
		{"both open", domain.Window{Start: mustPeriod(t, "2026-01"), End: open}, domain.Window{Start: mustPeriod(t, "2026-05"), End: open}, true},
		{"adjacent windows do not overlap (113-E exit)", domain.Window{Start: mustPeriod(t, "2026-01"), End: mustPeriod(t, "2026-06")}, domain.Window{Start: mustPeriod(t, "2026-07"), End: open}, false},
		{"same last month overlaps", domain.Window{Start: mustPeriod(t, "2026-01"), End: mustPeriod(t, "2026-07")}, domain.Window{Start: mustPeriod(t, "2026-07"), End: open}, true},
		{"later window ends before", domain.Window{Start: mustPeriod(t, "2026-08"), End: open}, domain.Window{Start: mustPeriod(t, "2026-01"), End: mustPeriod(t, "2026-03")}, false},
		{"nested", domain.Window{Start: mustPeriod(t, "2026-01"), End: mustPeriod(t, "2026-12")}, domain.Window{Start: mustPeriod(t, "2026-03"), End: mustPeriod(t, "2026-04")}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.a.Overlaps(tt.b))
			assert.Equal(t, tt.want, tt.b.Overlaps(tt.a), "overlap is symmetric")
		})
	}
}

func TestWindowContains(t *testing.T) {
	ended := domain.Window{Start: mustPeriod(t, "2026-01"), End: mustPeriod(t, "2026-06")}
	assert.True(t, ended.Contains(mustPeriod(t, "2026-01")))
	assert.True(t, ended.Contains(mustPeriod(t, "2026-06")))
	assert.False(t, ended.Contains(mustPeriod(t, "2025-12")))
	assert.False(t, ended.Contains(mustPeriod(t, "2026-07")))
	assert.False(t, ended.IsOpen())

	open := domain.Window{Start: mustPeriod(t, "2026-01")}
	assert.True(t, open.Contains(mustPeriod(t, "2099-12")))
	assert.True(t, open.IsOpen())
}

func mustPeriod(t *testing.T, raw string) domain.Period {
	t.Helper()
	p, err := domain.ParsePeriod(raw)
	require.NoError(t, err)
	return p
}
