package principal_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/platform/principal"
)

const (
	userID = "0199d4a2-7c1e-7b3a-9f10-2c5e8a1b4d01"
	jti    = "6f1c2a4e-1d1b-4c55-8a0e-5e3f1f9b7d21"
)

func TestFrom_returnsThePrincipalStoredByWith(t *testing.T) {
	want := principal.Principal{UserID: userID, TokenID: jti}

	got, ok := principal.From(principal.With(context.Background(), want))

	require.True(t, ok)
	assert.Equal(t, want, got)
}

func TestFrom_anonymousContextHasNoPrincipal(t *testing.T) {
	_, ok := principal.From(context.Background())
	assert.False(t, ok)
}

func TestUserID(t *testing.T) {
	tests := []struct {
		name    string
		ctx     context.Context
		want    string
		wantErr error
	}{
		{"authenticated", principal.With(context.Background(), principal.Principal{UserID: userID}), userID, nil},
		{"anonymous", context.Background(), "", principal.ErrUnauthenticated},
		{"empty user id", principal.With(context.Background(), principal.Principal{}), "", principal.ErrUnauthenticated},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := principal.UserID(tt.ctx)
			assert.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.want, got)
		})
	}
}
