package domain

import "slices"

// DocumentKind is a legal document a person accepts at sign-up (contract
// LegalDocumentKind).
type DocumentKind string

// The legal document kinds: the LFPDPPP privacy notice in its integral or
// simplified form, and the terms of service.
const (
	DocumentPrivacyIntegral   DocumentKind = "privacy_notice_integral"
	DocumentPrivacySimplified DocumentKind = "privacy_notice_simplified"
	DocumentTerms             DocumentKind = "terms"
)

var (
	documentKinds  = []DocumentKind{DocumentPrivacyIntegral, DocumentPrivacySimplified, DocumentTerms}
	privacyNotices = []DocumentKind{DocumentPrivacyIntegral, DocumentPrivacySimplified}
)

// DocumentVersion names one version of one legal document.
type DocumentVersion struct {
	Kind    DocumentKind
	Version string
}

// Acceptance is the set of legal document versions a person accepted: the
// terms and at least one privacy notice, each kind at most once.
type Acceptance struct {
	documents []DocumentVersion
}

// NewAcceptance validates the shape of what was accepted. Whether the
// versions are the current ones is a separate check (CheckCurrent), since it
// needs the catalog.
func NewAcceptance(documents []DocumentVersion) (Acceptance, error) {
	seen := make(map[DocumentKind]bool, len(documents))
	for _, d := range documents {
		switch {
		case !slices.Contains(documentKinds, d.Kind):
			return Acceptance{}, ErrDocumentUnknown
		case seen[d.Kind]:
			return Acceptance{}, ErrDocumentDuplicate
		case d.Version == "":
			return Acceptance{}, ErrDocumentNotCurrent
		}
		seen[d.Kind] = true
	}
	hasNotice := slices.ContainsFunc(privacyNotices, func(k DocumentKind) bool { return seen[k] })
	if !seen[DocumentTerms] || !hasNotice {
		return Acceptance{}, ErrDocumentMissing
	}
	return Acceptance{documents: slices.Clone(documents)}, nil
}

// CheckCurrent reports ErrDocumentNotCurrent unless every accepted version is
// the current version of its kind in current (kind -> version).
func (a Acceptance) CheckCurrent(current map[DocumentKind]string) error {
	for _, d := range a.documents {
		if version, ok := current[d.Kind]; !ok || version != d.Version {
			return ErrDocumentNotCurrent
		}
	}
	return nil
}

// Documents returns a copy of the accepted versions.
func (a Acceptance) Documents() []DocumentVersion { return slices.Clone(a.documents) }
