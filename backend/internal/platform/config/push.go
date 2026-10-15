package config

import (
	"context"
	"errors"
	"fmt"
	"regexp"
)

const (
	// keyFCMProject names the Firebase project of push reminders (M9-07).
	// Unset: push is disabled on this worker.
	keyFCMProject = "FCM_PROJECT_ID"
	// secretFCMCredentials is the optional service account key (JSON) of
	// the push sender. Absent: the runtime's Application Default
	// Credentials (the service account the worker runs as).
	secretFCMCredentials = "FCM_CREDENTIALS_JSON"
)

// projectIDPattern is the shape of a Google Cloud project id: 6–30
// characters, lowercase letters, digits and hyphens, starting with a letter
// and not ending with a hyphen.
var projectIDPattern = regexp.MustCompile(`^[a-z][a-z0-9-]{4,28}[a-z0-9]$`)

// PushConfig configures the push sender of the reminder job.
type PushConfig struct {
	// ProjectID is the Firebase project; empty disables push.
	ProjectID string
	// Credentials is the service account key, or zero for the runtime's
	// credentials.
	Credentials Secret
}

// Enabled reports whether push is configured on this worker.
func (p PushConfig) Enabled() bool { return p.ProjectID != "" }

// loadPush reads the push configuration. The credentials secret is read
// only when a project is set.
func loadPush(ctx context.Context, lookup Lookup, secrets SecretSource) (PushConfig, error) {
	project, ok := lookup(keyFCMProject)
	if !ok || project == "" {
		return PushConfig{}, nil
	}
	if !projectIDPattern.MatchString(project) {
		return PushConfig{}, fmt.Errorf("%s %q is not a Google Cloud project id", keyFCMProject, project)
	}
	creds, err := secrets.Secret(ctx, secretFCMCredentials)
	switch {
	case errors.Is(err, ErrSecretNotFound):
		return PushConfig{ProjectID: project}, nil
	case err != nil:
		return PushConfig{}, fmt.Errorf("secret %s: %w", secretFCMCredentials, err)
	}
	return PushConfig{ProjectID: project, Credentials: creds}, nil
}
