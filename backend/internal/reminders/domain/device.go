package domain

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"time"
)

var (
	// ErrInvalidPlatform reports a platform other than ios or android.
	ErrInvalidPlatform = errors.New("reminders: invalid platform")
	// ErrInvalidPushToken reports a push token that is empty, too long or
	// not printable ASCII. Its message never carries the token.
	ErrInvalidPushToken = errors.New("reminders: invalid push token")
	// ErrInvalidAppVersion reports an app version outside the contract's
	// pattern.
	ErrInvalidAppVersion = errors.New("reminders: invalid app version")
)

const (
	// MaxPushTokenLength is the contract's push_token maxLength.
	MaxPushTokenLength = 4096
	// MaxAppVersionLength is the contract's app_version maxLength.
	MaxAppVersionLength = 32

	// firstPrintable and lastPrintable bound printable ASCII without the
	// space ('!' through '~'), the contract's push_token pattern.
	firstPrintable = '!'
	lastPrintable  = '~'
	redacted       = "REDACTED"
)

// Platform is a device's operating system.
type Platform string

// The platforms the app ships on.
const (
	PlatformIOS     Platform = "ios"
	PlatformAndroid Platform = "android"
)

// ParsePlatform parses the contract's DevicePlatform.
func ParsePlatform(raw string) (Platform, error) {
	switch p := Platform(raw); p {
	case PlatformIOS, PlatformAndroid:
		return p, nil
	default:
		return "", fmt.Errorf("%w: %q", ErrInvalidPlatform, raw)
	}
}

// PushToken is an FCM registration token (§2.3). It is a secret of the
// installation: every fmt verb, slog and JSON render it redacted, and only the
// sealer reads it, through Reveal.
type PushToken struct {
	value string
}

// NewPushToken validates raw: 1-MaxPushTokenLength printable ASCII
// characters, no space. The error never quotes raw.
func NewPushToken(raw string) (PushToken, error) {
	if raw == "" || len(raw) > MaxPushTokenLength {
		return PushToken{}, fmt.Errorf("%w: length %d", ErrInvalidPushToken, len(raw))
	}
	for i := range len(raw) {
		if raw[i] < firstPrintable || raw[i] > lastPrintable {
			return PushToken{}, fmt.Errorf("%w: byte %d is not printable ASCII", ErrInvalidPushToken, i)
		}
	}
	return PushToken{value: raw}, nil
}

// Reveal is the token itself, for sealing and hashing only.
func (t PushToken) Reveal() string { return t.value }

// IsZero reports the unset token.
func (t PushToken) IsZero() bool { return t.value == "" }

// Format redacts the token for every fmt verb.
func (PushToken) Format(f fmt.State, _ rune) { _, _ = io.WriteString(f, redacted) }

// LogValue redacts the token in slog.
func (PushToken) LogValue() slog.Value { return slog.StringValue(redacted) }

// MarshalJSON redacts the token in JSON, as usecase.PushToken does.
func (PushToken) MarshalJSON() ([]byte, error) { return []byte(`"` + redacted + `"`), nil }

// ParseAppVersion validates the optional app version: 1-MaxAppVersionLength
// characters of [0-9A-Za-z.+-].
func ParseAppVersion(raw string) (string, error) {
	if raw == "" || len(raw) > MaxAppVersionLength {
		return "", fmt.Errorf("%w: length %d", ErrInvalidAppVersion, len(raw))
	}
	for i := range len(raw) {
		if !isVersionByte(raw[i]) {
			return "", fmt.Errorf("%w: %q", ErrInvalidAppVersion, raw)
		}
	}
	return raw, nil
}

func isVersionByte(b byte) bool {
	switch {
	case b >= '0' && b <= '9', b >= 'A' && b <= 'Z', b >= 'a' && b <= 'z':
		return true
	default:
		return b == '.' || b == '+' || b == '-'
	}
}

// Device is a registered installation as the API shows it: never the token.
type Device struct {
	// ID is the app's installation id (client-generated UUIDv7).
	ID       string
	Platform Platform
	// AppVersion is "" when the app did not send one.
	AppVersion string
	CreatedAt  time.Time
	LastSeenAt time.Time
}
