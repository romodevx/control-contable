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
		wantErr error
	}{
		{name: "plain", raw: "Ana (personal)", want: "Ana (personal)"},
		{name: "trimmed", raw: "  Ana  ", want: "Ana"},
		{name: "tabs and newlines trimmed", raw: "\tAna\n", want: "Ana"},
		{name: "composed to NFC", raw: "José", want: "José"},
		{name: "80 characters (runes, not bytes)", raw: strings.Repeat("ñ", 80), want: strings.Repeat("ñ", 80)},
		{name: "81 characters", raw: strings.Repeat("a", 81), wantErr: domain.ErrInvalidDisplayName},
		{name: "empty", raw: "", wantErr: domain.ErrInvalidDisplayName},
		{name: "blank", raw: "   ", wantErr: domain.ErrInvalidDisplayName},
		{name: "invalid utf-8", raw: "Ana\xff", wantErr: domain.ErrInvalidDisplayName},
		{name: "control character inside", raw: "Ana\x00B", wantErr: domain.ErrDisplayNameInvisible},
		{name: "bidi override inside", raw: "Ana\u202eB", wantErr: domain.ErrDisplayNameInvisible},
		{name: "zero-width space inside", raw: "Ana\u200bB", wantErr: domain.ErrDisplayNameInvisible},
		{name: "only a BOM", raw: "\ufeff", wantErr: domain.ErrDisplayNameInvisible},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := domain.NewDisplayName(tt.raw)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				assert.True(t, got.IsZero())
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got.String())
		})
	}
}
