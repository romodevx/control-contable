package usecase

import (
	"fmt"
	"io"

	"github.com/romodevx/control-contable/backend/internal/auth/domain"
)

// issuedCode is a fresh verification code with its stored form.
type issuedCode struct {
	code domain.VerificationCode
	hash []byte
}

// issueCode draws a code from entropy and computes its code_hash.
func issueCode(entropy io.Reader, codes CodeHasher, emailHash []byte) (issuedCode, error) {
	code, err := domain.NewVerificationCode(entropy)
	if err != nil {
		return issuedCode{}, fmt.Errorf("auth: draw verification code: %w", err)
	}
	return issuedCode{code: code, hash: codes.HashCode(emailHash, code)}, nil
}

// hashEmail computes the blind index of a normalized email.
func hashEmail(index EmailIndex, email domain.Email) ([]byte, error) {
	h, err := index.Hash(email.String())
	if err != nil {
		return nil, fmt.Errorf("auth: email index: %w", err)
	}
	return h, nil
}
