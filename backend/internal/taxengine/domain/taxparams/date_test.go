package taxparams_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/taxengine/domain/taxparams"
)

func TestParseDate(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		want    string
		wantErr bool
	}{
		{name: "calendar date", in: "2026-02-01", want: "2026-02-01"},
		{name: "leap day", in: "2028-02-29", want: "2028-02-29"},
		{name: "last day of the year", in: "2026-12-31", want: "2026-12-31"},
		{name: "not a leap year", in: "2026-02-29", wantErr: true},
		{name: "month 13", in: "2026-13-01", wantErr: true},
		{name: "day zero", in: "2026-01-00", wantErr: true},
		{name: "year zero", in: "0000-01-01", wantErr: true},
		{name: "one-digit month", in: "2026-1-01", wantErr: true},
		{name: "timestamp", in: "2026-01-01T00:00:00Z", wantErr: true},
		{name: "empty", in: "", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := taxparams.ParseDate(tt.in)
			if tt.wantErr {
				require.ErrorIs(t, err, taxparams.ErrInvalidDate)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got.String())
		})
	}
}

func TestNewDate(t *testing.T) {
	tests := []struct {
		name    string
		year    int
		month   time.Month
		day     int
		wantErr bool
	}{
		{name: "valid", year: 2026, month: time.February, day: 1},
		{name: "day overflow is not normalised", year: 2026, month: time.February, day: 30, wantErr: true},
		{name: "month overflow is not normalised", year: 2026, month: 13, day: 1, wantErr: true},
		{name: "year zero", year: 0, month: time.January, day: 1, wantErr: true},
		{name: "year past four digits", year: 10000, month: time.January, day: 1, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := taxparams.NewDate(tt.year, tt.month, tt.day)
			if tt.wantErr {
				require.ErrorIs(t, err, taxparams.ErrInvalidDate)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.year, got.Year())
			assert.False(t, got.IsZero())
		})
	}
}

func TestDateCompare(t *testing.T) {
	jan31, feb1 := mustDate(t, "2026-01-31"), mustDate(t, "2026-02-01")
	assert.Equal(t, -1, jan31.Compare(feb1))
	assert.Equal(t, 1, feb1.Compare(jan31))
	assert.Equal(t, 0, feb1.Compare(mustDate(t, "2026-02-01")))
	assert.True(t, taxparams.Date{}.IsZero())
	assert.Empty(t, taxparams.Date{}.String())
}

func TestValidity(t *testing.T) {
	closed := taxparams.Validity{From: mustDate(t, "2026-01-01"), To: mustDate(t, "2026-12-31")}
	open := taxparams.Validity{From: mustDate(t, "2026-02-01")}
	tests := []struct {
		name     string
		validity taxparams.Validity
		date     string
		want     bool
	}{
		{name: "first day is inclusive", validity: closed, date: "2026-01-01", want: true},
		{name: "last day is inclusive", validity: closed, date: "2026-12-31", want: true},
		{name: "the day after the end", validity: closed, date: "2027-01-01"},
		{name: "the day before the start", validity: closed, date: "2025-12-31"},
		{name: "open-ended reaches far", validity: open, date: "2099-12-31", want: true},
		{name: "open-ended still has a start", validity: open, date: "2026-01-31"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.validity.Contains(mustDate(t, tt.date)))
		})
	}
	assert.True(t, open.IsOpenEnded())
	assert.False(t, closed.IsOpenEnded())
	assert.Equal(t, "[2026-01-01, 2026-12-31]", closed.String())
	assert.Equal(t, "[2026-02-01, open)", open.String())
}
