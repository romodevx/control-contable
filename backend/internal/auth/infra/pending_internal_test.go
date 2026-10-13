package infra

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/auth/domain"
	"github.com/romodevx/control-contable/backend/internal/auth/usecase"
)

// storedHashSample stands in for an argon2id PHC string (made up, not a credential).
const storedHashSample = "fixture-phc-string"

func TestPending_roundTrips(t *testing.T) {
	name, err := domain.NewName("Ana Gómez")
	require.NoError(t, err)
	in := usecase.PendingSignUp{
		PasswordHash: storedHashSample,
		Name:         name,
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
}

func TestPending_withoutANameRoundTrips(t *testing.T) {
	raw, err := encodePending(usecase.PendingSignUp{
		PasswordHash: "h",
		Documents:    []domain.DocumentVersion{{Kind: domain.DocumentPrivacySimplified, Version: "v1"}, {Kind: domain.DocumentTerms, Version: "v1"}},
	})
	require.NoError(t, err)

	out, err := decodePending(raw)

	require.NoError(t, err)
	assert.True(t, out.Name.IsZero())
	assert.NotContains(t, raw, `"name"`)
}

// The payload crossed the database boundary: anything this code did not write
// is refused rather than turned into an account.
func TestPending_refusesWhatItDidNotWrite(t *testing.T) {
	docs := `"accepted_documents":[{"kind":"terms","version":"v1"},{"kind":"privacy_notice_integral","version":"v1"}]`
	for name, raw := range map[string]string{
		"not JSON":         `{`,
		"unknown version":  `{"v":2,"password_hash":"h",` + docs + `}`,
		"no password hash": `{"v":1,` + docs + `}`,
		"unknown field":    `{"v":1,"password_hash":"h","email":"x",` + docs + `}`,
		"blank name":       `{"v":1,"password_hash":"h","name":" ",` + docs + `}`,
		"missing terms":    `{"v":1,"password_hash":"h","accepted_documents":[{"kind":"privacy_notice_integral","version":"v1"}]}`,
	} {
		t.Run(name, func(t *testing.T) {
			_, err := decodePending(raw)
			assert.ErrorIs(t, err, ErrCorruptPending)
		})
	}
}
