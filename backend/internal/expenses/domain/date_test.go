package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/expenses/domain"
)

func TestParseDate(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		wantErr bool
	}{
		{name: "ordinary day", raw: "2026-09-05"},
		{name: "leap day", raw: "2024-02-29"},
		{name: "record floor", raw: "2022-01-01"},
		{name: "not a leap year", raw: "2026-02-29", wantErr: true},
		{name: "day 32", raw: "2026-01-32", wantErr: true},
		{name: "day zero", raw: "2026-01-00", wantErr: true},
		{name: "month 13", raw: "2026-13-01", wantErr: true},
		{name: "month zero", raw: "2026-00-10", wantErr: true},
		{name: "year zero", raw: "0000-01-01", wantErr: true},
		{name: "period only", raw: "2026-09", wantErr: true},
		{name: "slash separators", raw: "2026/09/05", wantErr: true},
		{name: "day first", raw: "05-09-2026", wantErr: true},
		{name: "time attached", raw: "2026-09-05T00:00", wantErr: true},
		{name: "non-digit", raw: "2026-0x-05", wantErr: true},
		{name: "empty", raw: "", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := domain.ParseDate(tt.raw)
			if tt.wantErr {
				require.ErrorIs(t, err, domain.ErrInvalid)
				assert.True(t, got.IsZero())
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.raw, got.String())
			assert.False(t, got.IsZero())
		})
	}
}

func TestDateCompare(t *testing.T) {
	tests := []struct {
		name string
		a, b string
		want int
	}{
		{name: "same day", a: "2026-09-05", b: "2026-09-05", want: 0},
		{name: "earlier year", a: "2025-12-31", b: "2026-01-01", want: -1},
		{name: "later month", a: "2026-10-01", b: "2026-09-30", want: 1},
		{name: "earlier day", a: "2026-09-04", b: "2026-09-05", want: -1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a, b := date(t, tt.a), date(t, tt.b)
			assert.Equal(t, tt.want, a.Compare(b))
			assert.Equal(t, tt.want < 0, a.Before(b))
			assert.Equal(t, tt.want > 0, a.After(b))
		})
	}
}

func TestDateZero(t *testing.T) {
	var zero domain.Date
	assert.True(t, zero.IsZero())
	assert.Empty(t, zero.String())
	assert.Equal(t, "2022-01-01", domain.MinRecordDate.String())
}

func TestDatePeriod(t *testing.T) {
	assert.Equal(t, "2026-09", date(t, "2026-09-05").Period())
	assert.Empty(t, domain.Date{}.Period())
}
