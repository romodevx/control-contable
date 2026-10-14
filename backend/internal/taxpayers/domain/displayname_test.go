package domain_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/taxpayers/domain"
)

func TestNewDisplayName(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		want    string
		wantErr bool
	}{
		{name: "plain", raw: "Ana (personal)", want: "Ana (personal)"},
		{name: "trimmed", raw: "  Ana  ", want: "Ana"},
		{name: "tabs and newlines trimmed", raw: "\tAna\n", want: "Ana"},
		{name: "composed to NFC", raw: "José", want: "José"},
		{name: "80 characters (runes, not bytes)", raw: strings.Repeat("ñ", 80), want: strings.Repeat("ñ", 80)},
		{name: "81 characters", raw: strings.Repeat("a", 81), wantErr: true},
		{name: "empty", raw: "", wantErr: true},
		{name: "blank", raw: "   ", wantErr: true},
		{name: "invalid utf-8", raw: "Ana\xff", wantErr: true},
		{name: "control character inside", raw: "Ana\x00B", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := domain.NewDisplayName(tt.raw)
			if tt.wantErr {
				require.ErrorIs(t, err, domain.ErrInvalidDisplayName)
				assert.True(t, got.IsZero())
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got.String())
		})
	}
}
