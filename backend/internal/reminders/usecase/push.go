package usecase

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
)

// MaxPushTargets is the most devices one Send may address (FCM SendEach
// takes up to 500 messages; the fair-use cap of devices per user is far
// below it, MODULE-reminders §2.3).
const MaxPushTargets = 500

// Push sender errors. Each one is a whole-call failure: nothing was handed to
// the push provider. Per-device results are PushDelivery values, not errors.
var (
	// ErrPushDisabled reports a worker without push configured (no Firebase
	// project): the reminder job records it and sends nothing.
	ErrPushDisabled = errors.New("reminders: push sender disabled")
	// ErrTooManyPushTargets reports more than MaxPushTargets devices.
	ErrTooManyPushTargets = errors.New("reminders: too many push targets")
	// ErrInvalidPushNotification reports a notification the provider would
	// refuse (missing text, a collapse id over the APNs limit, a blank
	// token); a bug in the message builder, never retried.
	ErrInvalidPushNotification = errors.New("reminders: invalid push notification")
)

// PushSender delivers one notification to a user's devices (M9-07; FCM in
// infra/fcm). It is the port the reminder job (M9-06) sends through.
type PushSender interface {
	// Send delivers n to every target and returns one PushDelivery per
	// target, in the order of targets. A non-nil error means nothing was
	// sent (the sentinel errors above); a provider that answered, even with
	// a failure for every device, reports it per device. No token or
	// notification text is ever part of an error.
	Send(ctx context.Context, n PushNotification, targets []PushTarget) ([]PushDelivery, error)
}

// PushNotification is what the lock screen shows (MODULE-reminders §2.5):
// display text and a deep link only, never an RFC, an id or a regime code.
type PushNotification struct {
	Title    string
	Body     string
	DeepLink string
	// CollapseID makes a resend replace the earlier notification instead of
	// stacking (android collapse_key and tag, apns-collapse-id; ≤ 64 bytes).
	CollapseID string
}

// PushTarget is one device to notify: its row id, so a failed delivery can
// be acted on, and its decrypted provider token.
type PushTarget struct {
	DeviceID uuid.UUID
	Token    PushToken
}

// PushDelivery is the result of a Send for one device.
type PushDelivery struct {
	DeviceID uuid.UUID
	Outcome  PushOutcome
	// ErrorClass names the failure for notification_log.error_class and the
	// reminders_failed_total metric ("" when delivered). It is a fixed
	// vocabulary, never provider text.
	ErrorClass string
}

// PushOutcome classifies a delivery for the reminder job.
type PushOutcome int

const (
	// PushDelivered means the provider accepted the message for the device.
	PushDelivered PushOutcome = iota
	// PushTokenInvalid means the token is no longer valid (UNREGISTERED, or an
	// invalid token); the job deletes the device (MODULE-reminders §2.4).
	PushTokenInvalid
	// PushTransient means the provider was unavailable, overloaded or did not
	// answer in time; the job retries within the day.
	PushTransient
	// PushRejected means a permanent failure that is not the token's (credentials,
	// sender mismatch, a refused payload); not retried, device kept.
	PushRejected
)

// String implements fmt.Stringer for logs and metrics labels.
func (o PushOutcome) String() string {
	switch o {
	case PushDelivered:
		return "delivered"
	case PushTokenInvalid:
		return "token_invalid"
	case PushTransient:
		return "transient"
	case PushRejected:
		return "rejected"
	default:
		return "unknown"
	}
}

// PushToken is a device's provider token. It is a credential for that device:
// every printing path (fmt verbs, JSON, slog) renders redactedToken; only
// Reveal returns it, at the point of use.
type PushToken struct {
	value string
}

const redactedToken = "[REDACTED]"

// NewPushToken wraps value.
func NewPushToken(value string) PushToken { return PushToken{value: value} }

// Reveal returns the token. Call it only where it is handed to the provider.
func (t PushToken) Reveal() string { return t.value }

// IsZero reports whether the token is empty.
func (t PushToken) IsZero() bool { return t.value == "" }

// String implements fmt.Stringer.
func (PushToken) String() string { return redactedToken }

// Format implements fmt.Formatter, so no verb can print the value.
func (PushToken) Format(f fmt.State, _ rune) { _, _ = f.Write([]byte(redactedToken)) }

// LogValue implements slog.LogValuer.
func (PushToken) LogValue() slog.Value { return slog.StringValue(redactedToken) }

// MarshalJSON implements json.Marshaler.
func (PushToken) MarshalJSON() ([]byte, error) { return []byte(`"` + redactedToken + `"`), nil }

// MaxCollapseIDBytes is the apns-collapse-id limit (the collapse id is the
// hex HMAC truncated to it, MODULE-reminders §2.4 step 5).
const MaxCollapseIDBytes = 64

// MaxPushPayloadBytes caps the text a notification carries (title, body,
// deep link and collapse id), in bytes: half of FCM's 4,096-byte message
// limit, so the keys and platform blocks around it always fit and FCM never
// refuses a message for its size (M9 security gate F2). BuildMessage stays
// far below it (three names of at most 80 characters).
const MaxPushPayloadBytes = 2048

// payloadBytes is the size of the text n carries (MaxPushPayloadBytes).
func payloadBytes(n PushNotification) int {
	return len(n.Title) + len(n.Body) + len(n.DeepLink) + len(n.CollapseID)
}

// ValidatePush checks a Send's arguments before anything leaves the process.
// Every PushSender calls it, the fakes included, so a message-builder bug
// fails the job's tests the way it would fail in production. The error
// names the rule broken, never a value.
func ValidatePush(n PushNotification, targets []PushTarget) error {
	switch {
	case len(targets) > MaxPushTargets:
		return fmt.Errorf("%w: %d targets, at most %d", ErrTooManyPushTargets, len(targets), MaxPushTargets)
	case n.Title == "":
		return fmt.Errorf("%w: title is empty", ErrInvalidPushNotification)
	case n.Body == "":
		return fmt.Errorf("%w: body is empty", ErrInvalidPushNotification)
	case len(n.CollapseID) > MaxCollapseIDBytes:
		return fmt.Errorf("%w: collapse id over %d bytes", ErrInvalidPushNotification, MaxCollapseIDBytes)
	case payloadBytes(n) > MaxPushPayloadBytes:
		return fmt.Errorf("%w: payload over %d bytes", ErrInvalidPushNotification, MaxPushPayloadBytes)
	}
	for i, target := range targets {
		if target.Token.IsZero() {
			return fmt.Errorf("%w: target %d has no token", ErrInvalidPushNotification, i)
		}
	}
	return nil
}

// PushSwitch is implemented by a PushSender that can be off for the whole
// process (DisabledPushSender). The reminder job asks it before claiming
// anything, so a worker without push never consumes a reminder; a sender
// that does not implement it is taken to be on.
type PushSwitch interface {
	PushEnabled() bool
}

// pushEnabled reports whether s can deliver at all.
func pushEnabled(s PushSender) bool {
	sw, ok := s.(PushSwitch)
	return !ok || sw.PushEnabled()
}

// DisabledPushSender is the PushSender of a worker without push configured:
// it reports itself off (PushSwitch) and every Send fails with
// ErrPushDisabled.
type DisabledPushSender struct{}

var (
	_ PushSender = DisabledPushSender{}
	_ PushSwitch = DisabledPushSender{}
)

// PushEnabled implements PushSwitch: always false.
func (DisabledPushSender) PushEnabled() bool { return false }

// Send implements PushSender.
func (DisabledPushSender) Send(context.Context, PushNotification, []PushTarget) ([]PushDelivery, error) {
	return nil, ErrPushDisabled
}
