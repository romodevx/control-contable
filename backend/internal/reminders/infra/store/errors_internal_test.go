package store

import (
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/reminders/usecase"
)

// claim_push_token documents two races (migration 20261021101000, step
// 4): both are a retry, anything else is an infrastructure failure.
func TestTranslateClaim(t *testing.T) {
	for code, retry := range map[string]bool{
		sqlstateSerializationFailure: true, sqlstateUniqueViolation: true,
		"42501": false, "23514": false,
	} {
		t.Run(code, func(t *testing.T) {
			pgErr := &pgconn.PgError{Code: code}

			err := translateClaim(pgErr)

			assert.Equal(t, retry, errors.Is(err, usecase.ErrClaimContention))
			require.ErrorIs(t, err, pgErr, "the driver error is kept for the logs")
		})
	}
	errOther := errors.New("connection reset")
	assert.NotErrorIs(t, translateClaim(errOther), usecase.ErrClaimContention)
}

func TestNullableAndBytea(t *testing.T) {
	assert.Nil(t, nullable(""))
	assert.Equal(t, "1.2.0", *nullable("1.2.0"))

	v, err := bytea([]byte{1, 2}).Value()
	require.NoError(t, err)
	assert.Equal(t, []byte{1, 2}, v)
}
