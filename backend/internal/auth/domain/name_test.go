package domain_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/auth/domain"
)

func TestNewName(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		want    string
		wantErr error
	}{
		{name: "plain", raw: "Ana Gómez Ruiz", want: "Ana Gómez Ruiz"},
		{name: "trimmed", raw: "  Ana \t", want: "Ana"},
		{name: "NFC normalized", raw: "Gómez", want: "Gómez"},
		{name: "120 characters", raw: strings.Repeat("ñ", 120), want: strings.Repeat("ñ", 120)},
		{name: "blank", raw: " \t\n", wantErr: domain.ErrInvalidName},
		{name: "empty", raw: "", wantErr: domain.ErrInvalidName},
		{name: "121 characters", raw: strings.Repeat("ñ", 121), wantErr: domain.ErrInvalidName},
		{name: "control character", raw: "Ana\x00B", wantErr: domain.ErrNameInvisible},
		{name: "bidi override", raw: "Ana\u202eB", wantErr: domain.ErrNameInvisible},
		{name: "zero-width joiner", raw: "Ana\u200dB", wantErr: domain.ErrNameInvisible},
		{name: "leading BOM", raw: "\ufeffAna", wantErr: domain.ErrNameInvisible},
		{name: "invalid UTF-8", raw: "Ana\xff", wantErr: domain.ErrInvalidName},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := domain.NewName(tt.raw)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got.String())
		})
	}
}
