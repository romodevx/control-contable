// Package fcmtest builds Firebase service account credentials for tests at
// run time, so nothing key-shaped is ever committed (the repository is
// public and scanned by GitGuardian and gitleaks).
package fcmtest

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"testing"

	"github.com/google/uuid"
)

// keyBits is the size of the throwaway key (the smallest Google accepts).
const keyBits = 2048

// ServiceAccountJSON returns a service account key for projectID with a
// freshly generated RSA key, whose token exchanges go to tokenURL.
func ServiceAccountJSON(tb testing.TB, projectID, tokenURL string) []byte {
	tb.Helper()
	key, err := rsa.GenerateKey(rand.Reader, keyBits)
	if err != nil {
		tb.Fatalf("generate key: %v", err)
	}
	der, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		tb.Fatalf("marshal key: %v", err)
	}
	account := struct {
		Type        string `json:"type"`
		ProjectID   string `json:"project_id"`
		KeyID       string `json:"private_key_id"`
		Key         string `json:"private_key"`
		ClientEmail string `json:"client_email"`
		ClientID    string `json:"client_id"`
		TokenURI    string `json:"token_uri"`
	}{
		Type:        "service_account",
		ProjectID:   projectID,
		KeyID:       uuid.NewString(),
		Key:         string(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der})),
		ClientEmail: "push-sender@" + projectID + ".iam.gserviceaccount.com",
		ClientID:    uuid.NewString(),
		TokenURI:    tokenURL,
	}
	encoded, err := json.Marshal(account) //nolint:gosec // G117: serializing this throwaway test key is the point.
	if err != nil {
		tb.Fatalf("marshal service account: %v", err)
	}
	return encoded
}
