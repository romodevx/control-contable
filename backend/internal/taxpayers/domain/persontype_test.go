package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/taxpayers/domain"
)

func TestParsePersonType(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		want    domain.PersonType
		wantErr error
	}{
		{name: "fisica", raw: "fisica", want: domain.PersonTypeFisica},
		{name: "moral", raw: "moral", want: domain.PersonTypeMoral},
		{name: "empty", raw: "", wantErr: domain.ErrInvalidPersonType},
		{name: "accented spelling is not the contract value", raw: "física", wantErr: domain.ErrInvalidPersonType},
		{name: "upper case is not the contract value", raw: "MORAL", wantErr: domain.ErrInvalidPersonType},
		{name: "surrounding space is not trimmed", raw: " moral", wantErr: domain.ErrInvalidPersonType},
		{name: "unknown", raw: "extranjero", wantErr: domain.ErrInvalidPersonType},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := domain.ParsePersonType(tt.raw)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				assert.False(t, got.Valid())
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
			assert.True(t, got.Valid())
			assert.Equal(t, tt.raw, got.String())
		})
	}
}

func TestPersonTypeZeroValueIsInvalid(t *testing.T) {
	var zero domain.PersonType
	assert.False(t, zero.Valid())
}
