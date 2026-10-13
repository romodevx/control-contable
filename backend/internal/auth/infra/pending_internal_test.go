package infra

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/auth/domain"
	"github.com/romodevx/control-contable/backend/internal/auth/usecase"
)

func TestPending_roundTrips(t *testing.T) {
	in := usecase.PendingSignUp{
		Documents: []domain.DocumentVersion{
			{Kind: domain.DocumentPrivacyIntegral, Version: "v1"},
			{Kind: domain.DocumentTerms, Version: "v1"},
		},
		AcceptedAt: time.Date(2026, 10, 13, 18, 30, 0, 0, time.UTC),
	}

	raw, err := encodePending(in)
	require.NoError(t, err)
	out, err := decodePending(raw)

	require.NoError(t, err)
	assert.Equal(t, in, out)
	assert.NotContains(t, raw, "email", "the payload never holds the address")
	assert.NotContains(t, raw, "password", "a pending sign-up holds no password (D8)")
	assert.NotContains(t, raw, `"name"`, "nor a name (D8)")
}

// The payload crossed the database boundary: anything this code did not write
// is refused rather than turned into an account. A v1 payload (with a
// password hash, before D8) is refused too: its password was chosen by
// whoever registered, not by whoever verifies.
func TestPending_refusesWhatItDidNotWrite(t *testing.T) {
	docs := `"accepted_documents":[{"kind":"terms","version":"v1"},{"kind":"privacy_notice_integral","version":"v1"}]`
	for name, raw := range map[string]string{
		"not JSON":            `{`,
		"unknown version":     `{"v":3,` + docs + `}`,
		"pre-D8 payload":      `{"v":1,"password_hash":"h",` + docs + `}`,
		"password hash at v2": `{"v":2,"password_hash":"h",` + docs + `}`,
		"name at v2":          `{"v":2,"name":"Ana",` + docs + `}`,
		"unknown field":       `{"v":2,"email":"x",` + docs + `}`,
		"missing terms":       `{"v":2,"accepted_documents":[{"kind":"privacy_notice_integral","version":"v1"}]}`,
		"no documents at all": `{"v":2}`,
	} {
		t.Run(name, func(t *testing.T) {
			_, err := decodePending(raw)
			assert.ErrorIs(t, err, ErrCorruptPending)
		})
	}
}
