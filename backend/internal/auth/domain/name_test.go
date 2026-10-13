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
		wantErr bool
	}{
		{name: "plain", raw: "Ana Gómez Ruiz", want: "Ana Gómez Ruiz"},
		{name: "trimmed", raw: "  Ana \t", want: "Ana"},
		{name: "NFC normalized", raw: "Gómez", want: "Gómez"},
		{name: "120 characters", raw: strings.Repeat("ñ", 120), want: strings.Repeat("ñ", 120)},
		{name: "blank", raw: " \t\n", wantErr: true},
		{name: "empty", raw: "", wantErr: true},
		{name: "121 characters", raw: strings.Repeat("ñ", 121), wantErr: true},
		{name: "control character", raw: "Ana\x00", wantErr: true},
		{name: "invalid UTF-8", raw: "Ana\xff", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := domain.NewName(tt.raw)
			if tt.wantErr {
				require.ErrorIs(t, err, domain.ErrInvalidName)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got.String())
		})
	}
}
