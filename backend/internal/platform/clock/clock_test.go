package clock

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLocationIsMexicoCity(t *testing.T) {
	assert.Equal(t, ZoneName, Location().String())
	// Mexico abolished DST in 2022: CST, UTC-6, all year.
	for _, month := range []time.Month{time.January, time.July} {
		_, offset := time.Date(2026, month, 1, 12, 0, 0, 0, Location()).Zone()
		assert.Equal(t, -6*60*60, offset, "offset in %s", month)
	}
}

func TestFixedNowIsInMexicoCity(t *testing.T) {
	tests := []struct {
		name    string
		instant time.Time
		wantDay string
	}{
		// 05:59 UTC on Oct 17 is still Oct 16 in Mexico City: the period
		// and the due date depend on the local calendar day, not on UTC.
		{"UTC instant before local midnight", time.Date(2026, 10, 17, 5, 59, 0, 0, time.UTC), "2026-10-16"},
		{"UTC instant after local midnight", time.Date(2026, 10, 17, 6, 0, 0, 0, time.UTC), "2026-10-17"},
		{"year boundary", time.Date(2027, 1, 1, 3, 0, 0, 0, time.UTC), "2026-12-31"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewFixed(tt.instant).Now()
			assert.Equal(t, ZoneName, got.Location().String())
			assert.True(t, got.Equal(tt.instant), "same instant")
			assert.Equal(t, tt.wantDay, got.Format(time.DateOnly))
		})
	}
}

func TestFixedIsStable(t *testing.T) {
	c := NewFixed(time.Date(2026, 3, 17, 18, 0, 0, 0, time.UTC))
	assert.Equal(t, c.Now(), c.Now())
}

func TestSystemNowIsInMexicoCity(t *testing.T) {
	before := time.Now()
	got := NewSystem().Now()
	after := time.Now()
	assert.Equal(t, ZoneName, got.Location().String())
	assert.False(t, got.Before(before) || got.After(after), "System must read the wall clock")
}

func TestClockImplementations(t *testing.T) {
	var _ Clock = System{}
	var _ Clock = Fixed{}
}

func TestMustLoadLocationPanicsOnUnknownZone(t *testing.T) {
	require.PanicsWithError(t, `clock: load time zone "Mars/Olympus_Mons": unknown time zone Mars/Olympus_Mons`, func() {
		mustLoadLocation("Mars/Olympus_Mons")
	})
}
