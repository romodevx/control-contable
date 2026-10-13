package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/auth/domain"
)

func doc(kind domain.DocumentKind, version string) domain.DocumentVersion {
	return domain.DocumentVersion{Kind: kind, Version: version}
}

var currentV1 = map[domain.DocumentKind]string{
	domain.DocumentTerms:             "v1",
	domain.DocumentPrivacyIntegral:   "v1",
	domain.DocumentPrivacySimplified: "v1",
}

func TestNewAcceptance_needsTermsAndOnePrivacyNotice(t *testing.T) {
	tests := []struct {
		name    string
		docs    []domain.DocumentVersion
		wantErr error
	}{
		{"terms + integral", []domain.DocumentVersion{doc(domain.DocumentPrivacyIntegral, "v1"), doc(domain.DocumentTerms, "v1")}, nil},
		{"terms + simplified", []domain.DocumentVersion{doc(domain.DocumentTerms, "v1"), doc(domain.DocumentPrivacySimplified, "v1")}, nil},
		{"terms + both notices", []domain.DocumentVersion{doc(domain.DocumentTerms, "v1"), doc(domain.DocumentPrivacySimplified, "v1"), doc(domain.DocumentPrivacyIntegral, "v1")}, nil},
		{"no terms", []domain.DocumentVersion{doc(domain.DocumentPrivacyIntegral, "v1"), doc(domain.DocumentPrivacySimplified, "v1")}, domain.ErrDocumentMissing},
		{"no privacy notice", []domain.DocumentVersion{doc(domain.DocumentTerms, "v1")}, domain.ErrDocumentMissing},
		{"nothing", nil, domain.ErrDocumentMissing},
		{"a kind twice", []domain.DocumentVersion{doc(domain.DocumentTerms, "v1"), doc(domain.DocumentTerms, "v2"), doc(domain.DocumentPrivacyIntegral, "v1")}, domain.ErrDocumentDuplicate},
		{"unknown kind", []domain.DocumentVersion{doc(domain.DocumentTerms, "v1"), doc("cookies", "v1"), doc(domain.DocumentPrivacyIntegral, "v1")}, domain.ErrDocumentUnknown},
		{"blank version", []domain.DocumentVersion{doc(domain.DocumentTerms, ""), doc(domain.DocumentPrivacyIntegral, "v1")}, domain.ErrDocumentNotCurrent},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := domain.NewAcceptance(tt.docs)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.ElementsMatch(t, tt.docs, got.Documents())
		})
	}
}

func TestAcceptance_CheckCurrent(t *testing.T) {
	accepted, err := domain.NewAcceptance([]domain.DocumentVersion{doc(domain.DocumentTerms, "v1"), doc(domain.DocumentPrivacyIntegral, "v1")})
	require.NoError(t, err)

	assert.NoError(t, accepted.CheckCurrent(currentV1))

	newTerms := map[domain.DocumentKind]string{domain.DocumentTerms: "v2", domain.DocumentPrivacyIntegral: "v1"}
	assert.ErrorIs(t, accepted.CheckCurrent(newTerms), domain.ErrDocumentNotCurrent, "terms moved on")

	noIntegral := map[domain.DocumentKind]string{domain.DocumentTerms: "v1"}
	assert.ErrorIs(t, accepted.CheckCurrent(noIntegral), domain.ErrDocumentNotCurrent, "a kind with no current version")
}

func TestAcceptance_DocumentsIsACopy(t *testing.T) {
	accepted, err := domain.NewAcceptance([]domain.DocumentVersion{doc(domain.DocumentTerms, "v1"), doc(domain.DocumentPrivacyIntegral, "v1")})
	require.NoError(t, err)

	docs := accepted.Documents()
	docs[0].Version = "tampered"

	assert.NoError(t, accepted.CheckCurrent(currentV1))
}
