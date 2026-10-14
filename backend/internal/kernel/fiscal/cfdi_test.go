package fiscal_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
)

func TestParseCFDIUUID(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		want    string
		wantErr bool
	}{
		{name: "upper case", raw: "6F1C2A3B-4D5E-4F60-8A9B-0C1D2E3F4A5B", want: "6F1C2A3B-4D5E-4F60-8A9B-0C1D2E3F4A5B"},
		{name: "lower case is normalized", raw: "6f1c2a3b-4d5e-4f60-8a9b-0c1d2e3f4a5b", want: "6F1C2A3B-4D5E-4F60-8A9B-0C1D2E3F4A5B"},
		{name: "missing a group", raw: "6F1C2A3B-4D5E-4F60-0C1D2E3F4A5B", wantErr: true},
		{name: "hyphen misplaced", raw: "6F1C2A3B4-D5E-4F60-8A9B-0C1D2E3F4A5B", wantErr: true},
		{name: "non-hex", raw: "6F1C2A3B-4D5E-4F60-8A9B-0C1D2E3F4A5G", wantErr: true},
		{name: "braces", raw: "{6F1C2A3B-4D5E-4F60-8A9B-0C1D2E3F4A5B}", wantErr: true},
		{name: "empty", raw: "", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := fiscal.ParseCFDIUUID(tt.raw)
			if tt.wantErr {
				require.ErrorIs(t, err, fiscal.ErrInvalid)
				assert.True(t, got.IsZero())
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got.String())
			assert.False(t, got.IsZero())
		})
	}
}

func TestCFDIUUIDZero(t *testing.T) {
	assert.True(t, fiscal.CFDIUUID{}.IsZero())
}
