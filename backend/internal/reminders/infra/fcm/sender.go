// Package fcm is the reminders module's push sender over Firebase Cloud
// Messaging HTTP v1 (Firebase Admin Go SDK, ADR-0001 Push row; M9-07). It
// implements usecase.PushSender: one SendEach per notification, one
// classified PushDelivery per device, so the reminder job (M9-06) can delete
// a device whose token is gone, retry a transient failure, or give up.
//
// It never logs, and no error it returns carries a token, notification text
// or FCM's own message: per-device failures become a fixed ErrorClass.
package fcm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/errorutils"
	"firebase.google.com/go/v4/messaging"
	"google.golang.org/api/option"

	"github.com/romodevx/control-contable/backend/internal/reminders/usecase"
)

// DefaultSendTimeout bounds one Send, retries included: the spec's 30 s send
// deadline, far below the job's 10-minute claim lease (MODULE-reminders §2.4).
const DefaultSendTimeout = 30 * time.Second

// ErrConfig reports a sender that cannot be built: no project, a negative
// timeout or credentials the SDK refuses. It never carries the credentials.
var ErrConfig = errors.New("fcm: invalid configuration")

// Payload keys and headers of the message.
const (
	dataKeyLink        = "link"
	headerAPNsCollapse = "apns-collapse-id"
)

// ErrorClass vocabulary (notification_log.error_class, reminders_failed_total).
const (
	classUnregistered     = "unregistered"
	classInvalidArgument  = "invalid_argument"
	classInternal         = "internal"
	classUnavailable      = "unavailable"
	classServerError      = "server_error"
	classQuotaExceeded    = "quota_exceeded"
	classDeadline         = "deadline"
	classCanceled         = "canceled"
	classNetwork          = "network"
	classSenderIDMismatch = "sender_id_mismatch"
	classThirdPartyAuth   = "third_party_auth"
	classRejected         = "rejected"
)

// Config configures a Sender.
type Config struct {
	// ProjectID is the Firebase project (required by FCM v1).
	ProjectID string
	// CredentialsJSON is a service account key from the secret manager.
	// Empty: Application Default Credentials (the runtime's service account).
	CredentialsJSON []byte
	// SendTimeout bounds one Send; zero means DefaultSendTimeout.
	SendTimeout time.Duration
}

// Sender implements usecase.PushSender over FCM. It is safe for concurrent
// use (the job sends for several users at once).
type Sender struct {
	client  *messaging.Client
	timeout time.Duration
}

var _ usecase.PushSender = (*Sender)(nil)

// New builds a Sender. The credentials are parsed here, so a malformed key or
// a runtime without credentials fails the worker's startup, not the first
// reminder. opts are extra client options (tests point the endpoint and HTTP
// client at a fake FCM server).
func New(ctx context.Context, cfg Config, opts ...option.ClientOption) (*Sender, error) {
	if cfg.ProjectID == "" {
		return nil, fmt.Errorf("%w: project id is required", ErrConfig)
	}
	if cfg.SendTimeout < 0 {
		return nil, fmt.Errorf("%w: send timeout is negative", ErrConfig)
	}
	timeout := cfg.SendTimeout
	if timeout == 0 {
		timeout = DefaultSendTimeout
	}
	if len(cfg.CredentialsJSON) > 0 {
		// Typed: only a service account key is accepted, never an external
		// account configuration that could point the SDK elsewhere.
		opts = append([]option.ClientOption{option.WithAuthCredentialsJSON(option.ServiceAccount, cfg.CredentialsJSON)}, opts...)
	}
	app, err := firebase.NewApp(ctx, &firebase.Config{ProjectID: cfg.ProjectID}, opts...)
	if err != nil {
		return nil, fmt.Errorf("%w: firebase app: %s", ErrConfig, safeCause(err))
	}
	client, err := app.Messaging(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w: messaging client: %s", ErrConfig, safeCause(err))
	}
	return &Sender{client: client, timeout: timeout}, nil
}

// safeCause names a construction failure without its text: credential
// parsers may quote the input they choke on.
func safeCause(err error) string {
	return fmt.Sprintf("%T", err)
}

// Send implements usecase.PushSender.
func (s *Sender) Send(ctx context.Context, n usecase.PushNotification, targets []usecase.PushTarget) ([]usecase.PushDelivery, error) {
	if err := usecase.ValidatePush(n, targets); err != nil {
		return nil, err
	}
	if len(targets) == 0 {
		return nil, nil
	}
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()
	batch, err := s.client.SendEach(ctx, toMessages(n, targets))
	if err != nil {
		// SendEach fails as a whole only on its own validation, which
		// ValidatePush mirrors; its text names the rule, not a value.
		return nil, fmt.Errorf("%w: %w", usecase.ErrInvalidPushNotification, err)
	}
	return classifyBatch(ctx, targets, batch.Responses), nil
}

// toMessages builds one message per target, in order.
func toMessages(n usecase.PushNotification, targets []usecase.PushTarget) []*messaging.Message {
	messages := make([]*messaging.Message, len(targets))
	for i, target := range targets {
		messages[i] = toMessage(n, target.Token)
	}
	return messages
}

func toMessage(n usecase.PushNotification, token usecase.PushToken) *messaging.Message {
	m := &messaging.Message{
		// Token is deprecated in favour of Fid (FlutterFire's register());
		// the app on firebase_messaging 16.7.0 registers getToken() tokens
		// (ADR-0001 Mobile row), which FCM still addresses by Token.
		Token:        token.Reveal(), //nolint:staticcheck // SA1019: registration tokens until the app moves to FIDs (ADR-0001).
		Notification: &messaging.Notification{Title: n.Title, Body: n.Body},
	}
	if n.DeepLink != "" {
		m.Data = map[string]string{dataKeyLink: n.DeepLink}
	}
	if n.CollapseID != "" {
		m.Android = &messaging.AndroidConfig{
			CollapseKey:  n.CollapseID,
			Notification: &messaging.AndroidNotification{Tag: n.CollapseID},
		}
		m.APNS = &messaging.APNSConfig{Headers: map[string]string{headerAPNsCollapse: n.CollapseID}}
	}
	return m
}

// classifyBatch turns FCM's per-message responses into deliveries, in the
// order of the targets.
func classifyBatch(ctx context.Context, targets []usecase.PushTarget, responses []*messaging.SendResponse) []usecase.PushDelivery {
	deliveries := make([]usecase.PushDelivery, len(targets))
	for i, target := range targets {
		outcome, class := classify(ctx, responses[i])
		deliveries[i] = usecase.PushDelivery{DeviceID: target.DeviceID, Outcome: outcome, ErrorClass: class}
	}
	return deliveries
}

// classify maps one FCM response. FCM's own error code (the FcmError detail)
// decides first; then the platform status, the HTTP status and the context.
func classify(ctx context.Context, r *messaging.SendResponse) (usecase.PushOutcome, string) {
	if r == nil {
		return usecase.PushTransient, classNetwork
	}
	if r.Success {
		return usecase.PushDelivered, ""
	}
	err := r.Error
	switch {
	case messaging.IsUnregistered(err):
		return usecase.PushTokenInvalid, classUnregistered
	case messaging.IsInvalidArgument(err):
		return invalidArgument(err), classInvalidArgument
	case messaging.IsSenderIDMismatch(err):
		return usecase.PushRejected, classSenderIDMismatch
	case messaging.IsThirdPartyAuthError(err):
		return usecase.PushRejected, classThirdPartyAuth
	case messaging.IsQuotaExceeded(err) || errorutils.IsResourceExhausted(err):
		return usecase.PushTransient, classQuotaExceeded
	case messaging.IsUnavailable(err):
		return usecase.PushTransient, classUnavailable
	case messaging.IsInternal(err):
		return usecase.PushTransient, classInternal
	}
	return classifyTransport(ctx, err)
}

// classifyTransport maps what FCM did not classify: a server error without
// an FCM body, a deadline, a cancellation or a network failure (transient),
// or any other answer (rejected).
func classifyTransport(ctx context.Context, err error) (usecase.PushOutcome, string) {
	if status, answered := httpStatus(err); answered {
		if status >= http.StatusInternalServerError {
			return usecase.PushTransient, classServerError
		}
		return usecase.PushRejected, classRejected
	}
	switch {
	case errors.Is(ctx.Err(), context.Canceled) || errors.Is(err, context.Canceled):
		return usecase.PushTransient, classCanceled
	case ctx.Err() != nil || errorutils.IsDeadlineExceeded(err) || errors.Is(err, context.DeadlineExceeded):
		return usecase.PushTransient, classDeadline
	default:
		return usecase.PushTransient, classNetwork
	}
}

// httpStatus returns the HTTP status FCM answered with, if it answered.
// The response is a buffered copy; its body is closed all the same.
func httpStatus(err error) (int, bool) {
	resp := errorutils.HTTPResponse(err)
	if resp == nil {
		return 0, false
	}
	if resp.Body != nil {
		_ = resp.Body.Close() // a buffered copy: closing cannot fail
	}
	return resp.StatusCode, true
}

// fcmTokenField is the field FCM names in its google.rpc.BadRequest detail
// when the registration token itself is what it refused.
const fcmTokenField = "message.token"

// invalidArgument attributes an INVALID_ARGUMENT (M9 security gate F2): the
// device's token is invalid only when FCM's BadRequest detail names the
// token field; any other field, or none, is the message, so the device is
// kept (rejected) and an UNREGISTERED later cleans a token that is gone.
func invalidArgument(err error) usecase.PushOutcome {
	if refusedField(err, fcmTokenField) {
		return usecase.PushTokenInvalid
	}
	return usecase.PushRejected
}

// fcmErrorBody is the part of an FCM v1 error body that names the refused
// fields. Descriptions are never read: they may quote the input.
type fcmErrorBody struct {
	Error struct {
		Details []struct {
			FieldViolations []struct {
				Field string `json:"field"`
			} `json:"fieldViolations"`
		} `json:"details"`
	} `json:"error"`
}

// maxErrorBodyBytes bounds how much of an error body is decoded.
const maxErrorBodyBytes = 64 << 10

// refusedField reports whether FCM's answer to err names field among its
// BadRequest field violations.
func refusedField(err error, field string) bool {
	resp := errorutils.HTTPResponse(err)
	if resp == nil || resp.Body == nil {
		return false
	}
	defer func() { _ = resp.Body.Close() }() // a buffered copy: closing cannot fail
	var body fcmErrorBody
	if json.NewDecoder(io.LimitReader(resp.Body, maxErrorBodyBytes)).Decode(&body) != nil {
		return false
	}
	for _, detail := range body.Error.Details {
		for _, violation := range detail.FieldViolations {
			if violation.Field == field {
				return true
			}
		}
	}
	return false
}
