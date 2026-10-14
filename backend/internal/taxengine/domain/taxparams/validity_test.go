package taxparams_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/romodevx/control-contable/backend/internal/taxengine/domain/taxparams"
)

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
