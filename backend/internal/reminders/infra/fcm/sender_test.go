package fcm_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/api/option"

	"github.com/romodevx/control-contable/backend/internal/reminders/infra/fcm"
	"github.com/romodevx/control-contable/backend/internal/reminders/usecase"
)

const testProject = "cc-test-project"

// fcmReply is what the fake FCM v1 server answers for one token.
type fcmReply struct {
	status     int
	errStatus  string // google.rpc status, e.g. NOT_FOUND
	fcmCode    string // FcmError.errorCode, e.g. UNREGISTERED; "" for none
	bare       bool   // no JSON body (a proxy's 5xx)
	block      bool   // never answer until the test ends (timeout)
	retryAfter string // Retry-After header
}

var (
	replyOK          = fcmReply{status: http.StatusOK}
	replyUnreg       = fcmReply{status: http.StatusNotFound, errStatus: "NOT_FOUND", fcmCode: "UNREGISTERED"}
	replyInvalidArg  = fcmReply{status: http.StatusBadRequest, errStatus: "INVALID_ARGUMENT", fcmCode: "INVALID_ARGUMENT"}
	replyInternal    = fcmReply{status: http.StatusInternalServerError, errStatus: "INTERNAL", fcmCode: "INTERNAL"}
	replyUnavailable = fcmReply{status: http.StatusServiceUnavailable, errStatus: "UNAVAILABLE", fcmCode: "UNAVAILABLE"}
	replyQuota       = fcmReply{status: http.StatusTooManyRequests, errStatus: "RESOURCE_EXHAUSTED", fcmCode: "QUOTA_EXCEEDED"}
	replyMismatch    = fcmReply{status: http.StatusForbidden, errStatus: "PERMISSION_DENIED", fcmCode: "SENDER_ID_MISMATCH"}
	replyAPNsAuth    = fcmReply{status: http.StatusUnauthorized, errStatus: "UNAUTHENTICATED", fcmCode: "THIRD_PARTY_AUTH_ERROR"}
	replyDenied      = fcmReply{status: http.StatusForbidden, errStatus: "PERMISSION_DENIED"}
	replyBadGateway  = fcmReply{status: http.StatusBadGateway, bare: true}
	replyBlock       = fcmReply{block: true}
	// A Retry-After past the SDK's 2-minute cap: the SDK does not retry.
	replyUnavailableLong = fcmReply{status: http.StatusServiceUnavailable, errStatus: "UNAVAILABLE", fcmCode: "UNAVAILABLE", retryAfter: "600"}
	replyThrottled       = fcmReply{status: http.StatusTooManyRequests, errStatus: "RESOURCE_EXHAUSTED"}
	replyBareInternal    = fcmReply{status: http.StatusInternalServerError, bare: true}
)

// sentMessage is the part of an FCM v1 request the tests read.
type sentMessage struct {
	Message struct {
		Token        string            `json:"token"`
		Data         map[string]string `json:"data"`
		Notification struct {
			Title string `json:"title"`
			Body  string `json:"body"`
		} `json:"notification"`
		Android struct {
			CollapseKey  string `json:"collapse_key"`
			Notification struct {
				Tag string `json:"tag"`
			} `json:"notification"`
		} `json:"android"`
		APNS struct {
			Headers map[string]string `json:"headers"`
		} `json:"apns"`
	} `json:"message"`
}

// fakeFCM emulates projects.messages:send of FCM HTTP v1: each token gets
// the replies queued for it, in order; the last one repeats.
type fakeFCM struct {
	t       *testing.T
	release chan struct{}

	mu      sync.Mutex
	replies map[string][]fcmReply
	sent    []sentMessage
	calls   map[string]int
}

func newFakeFCM(t *testing.T, replies map[string][]fcmReply) (*fakeFCM, *httptest.Server) {
	t.Helper()
	f := &fakeFCM{t: t, release: make(chan struct{}), replies: replies, calls: map[string]int{}}
	srv := httptest.NewServer(f)
	// Unblock held handlers before Close waits for them.
	t.Cleanup(srv.Close)
	t.Cleanup(func() { close(f.release) })
	return f, srv
}

func (f *fakeFCM) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost || r.URL.Path != "/projects/"+testProject+"/messages:send" {
		f.t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
		http.NotFound(w, r)
		return
	}
	var msg sentMessage
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		f.t.Errorf("decode request: %v", err)
		return
	}
	reply := f.record(msg)
	if reply.block {
		select {
		case <-f.release:
		case <-r.Context().Done():
		}
		return
	}
	writeReply(w, reply)
}

func (f *fakeFCM) record(msg sentMessage) fcmReply {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.sent = append(f.sent, msg)
	token := msg.Message.Token
	queue := f.replies[token]
	i := min(f.calls[token], len(queue)-1)
	f.calls[token]++
	if i < 0 {
		return replyOK
	}
	return queue[i]
}

func (f *fakeFCM) messages() []sentMessage {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]sentMessage(nil), f.sent...)
}

func (f *fakeFCM) callsFor(token string) int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.calls[token]
}

func writeReply(w http.ResponseWriter, reply fcmReply) {
	if reply.retryAfter != "" {
		w.Header().Set("Retry-After", reply.retryAfter)
	}
	if reply.bare {
		w.WriteHeader(reply.status)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(reply.status)
	if reply.status == http.StatusOK {
		_, _ = fmt.Fprintf(w, `{"name":"projects/%s/messages/%s"}`, testProject, uuid.NewString())
		return
	}
	details := "[]"
	if reply.fcmCode != "" {
		details = fmt.Sprintf(`[{"@type":"type.googleapis.com/google.firebase.fcm.v1.FcmError","errorCode":%q}]`, reply.fcmCode)
	}
	_, _ = fmt.Fprintf(w, `{"error":{"code":%d,"message":"fake failure","status":%q,"details":%s}}`,
		reply.status, reply.errStatus, details)
}

func newSender(t *testing.T, srv *httptest.Server, timeout time.Duration) *fcm.Sender {
	t.Helper()
	s, err := fcm.New(t.Context(), fcm.Config{ProjectID: testProject, SendTimeout: timeout},
		option.WithEndpoint(srv.URL), option.WithHTTPClient(srv.Client()))
	require.NoError(t, err)
	return s
}

func notification() usecase.PushNotification {
	return usecase.PushNotification{
		Title:      "Hoy vence tu pago mensual estimado",
		Body:       "Revisa tu estimado de abril para: Ana, Taller Sur.",
		DeepLink:   "controlcontable://estimate?period=2026-04",
		CollapseID: strings.Repeat("c0", usecase.MaxCollapseIDBytes/2),
	}
}

// target returns a device with a token unique to the test run.
func target(name string) usecase.PushTarget {
	return usecase.PushTarget{DeviceID: uuid.New(), Token: usecase.NewPushToken(name + "-" + uuid.NewString())}
}

func TestSender_mapsTheNotification(t *testing.T) {
	fake, srv := newFakeFCM(t, nil)
	sender := newSender(t, srv, time.Second)
	dev := target("ok")
	n := notification()

	got, err := sender.Send(t.Context(), n, []usecase.PushTarget{dev})

	require.NoError(t, err)
	assert.Equal(t, []usecase.PushDelivery{{DeviceID: dev.DeviceID, Outcome: usecase.PushDelivered}}, got)
	sent := fake.messages()
	require.Len(t, sent, 1)
	m := sent[0].Message
	assert.Equal(t, dev.Token.Reveal(), m.Token)
	assert.Equal(t, n.Title, m.Notification.Title)
	assert.Equal(t, n.Body, m.Notification.Body)
	assert.Equal(t, map[string]string{"link": n.DeepLink}, m.Data)
	assert.Equal(t, n.CollapseID, m.Android.CollapseKey)
	assert.Equal(t, n.CollapseID, m.Android.Notification.Tag)
	assert.Equal(t, n.CollapseID, m.APNS.Headers["apns-collapse-id"])
}

func TestSender_withoutCollapseIDOrLink(t *testing.T) {
	fake, srv := newFakeFCM(t, nil)
	sender := newSender(t, srv, time.Second)
	n := notification()
	n.CollapseID, n.DeepLink = "", ""

	_, err := sender.Send(t.Context(), n, []usecase.PushTarget{target("ok")})

	require.NoError(t, err)
	m := fake.messages()[0].Message
	assert.Empty(t, m.Data)
	assert.Empty(t, m.Android.CollapseKey)
	assert.Empty(t, m.APNS.Headers)
}

// Every FCM v1 answer the reminder job must tell apart (MODULE-reminders
// §2.4): delete the device, retry later, or give up keeping the device.
func TestSender_classifiesEachDevice(t *testing.T) {
	tests := []struct {
		name    string
		replies []fcmReply
		outcome usecase.PushOutcome
		class   string
	}{
		{"delivered", []fcmReply{replyOK}, usecase.PushDelivered, ""},
		{"unregistered token", []fcmReply{replyUnreg}, usecase.PushTokenInvalid, "unregistered"},
		{"invalid token", []fcmReply{replyInvalidArg}, usecase.PushTokenInvalid, "invalid_argument"},
		{"fcm internal error", []fcmReply{replyInternal}, usecase.PushTransient, "internal"},
		{"proxy 5xx without body", []fcmReply{replyBadGateway}, usecase.PushTransient, "server_error"},
		{"500 without fcm detail", []fcmReply{replyBareInternal}, usecase.PushTransient, "server_error"},
		{"unavailable, retry-after beyond the sdk cap", []fcmReply{replyUnavailableLong}, usecase.PushTransient, "unavailable"},
		{"throttled without fcm detail", []fcmReply{replyThrottled}, usecase.PushTransient, "quota_exceeded"},
		{"quota exceeded", []fcmReply{replyQuota}, usecase.PushTransient, "quota_exceeded"},
		{"unavailable then delivered on the sdk retry", []fcmReply{replyUnavailable, replyOK}, usecase.PushDelivered, ""},
		{"sender id mismatch", []fcmReply{replyMismatch}, usecase.PushRejected, "sender_id_mismatch"},
		{"apns credentials", []fcmReply{replyAPNsAuth}, usecase.PushRejected, "third_party_auth"},
		{"permission denied", []fcmReply{replyDenied}, usecase.PushRejected, "rejected"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dev := target("dev")
			_, srv := newFakeFCM(t, map[string][]fcmReply{dev.Token.Reveal(): tt.replies})
			sender := newSender(t, srv, 5*time.Second)

			got, err := sender.Send(t.Context(), notification(), []usecase.PushTarget{dev})

			require.NoError(t, err)
			assert.Equal(t, []usecase.PushDelivery{{DeviceID: dev.DeviceID, Outcome: tt.outcome, ErrorClass: tt.class}}, got)
		})
	}
}

// One stale phone must not cost the user's other devices their reminder,
// and the results come back in the order of the targets.
func TestSender_partialFailureKeepsOrder(t *testing.T) {
	stale, ok, down := target("stale"), target("ok"), target("down")
	_, srv := newFakeFCM(t, map[string][]fcmReply{
		stale.Token.Reveal(): {replyUnreg},
		down.Token.Reveal():  {replyInternal},
	})
	sender := newSender(t, srv, 5*time.Second)

	got, err := sender.Send(t.Context(), notification(), []usecase.PushTarget{stale, ok, down})

	require.NoError(t, err)
	assert.Equal(t, []usecase.PushDelivery{
		{DeviceID: stale.DeviceID, Outcome: usecase.PushTokenInvalid, ErrorClass: "unregistered"},
		{DeviceID: ok.DeviceID, Outcome: usecase.PushDelivered},
		{DeviceID: down.DeviceID, Outcome: usecase.PushTransient, ErrorClass: "internal"},
	}, got)
}

// INVALID_ARGUMENT names the token only when it is not every device's
// answer: when all of several devices get it, the message is what FCM
// refused, and deleting every device of the user would be the wrong cleanup.
func TestSender_invalidArgumentForEveryDeviceIsThePayload(t *testing.T) {
	a, b := target("a"), target("b")
	_, srv := newFakeFCM(t, map[string][]fcmReply{
		a.Token.Reveal(): {replyInvalidArg},
		b.Token.Reveal(): {replyInvalidArg},
	})
	sender := newSender(t, srv, 5*time.Second)

	got, err := sender.Send(t.Context(), notification(), []usecase.PushTarget{a, b})

	require.NoError(t, err)
	assert.Equal(t, []usecase.PushDelivery{
		{DeviceID: a.DeviceID, Outcome: usecase.PushRejected, ErrorClass: "invalid_argument"},
		{DeviceID: b.DeviceID, Outcome: usecase.PushRejected, ErrorClass: "invalid_argument"},
	}, got)
}

// A 503 that outlasts the send deadline is transient: the SDK's retries stop
// at the deadline and the job retries the reminder later.
func TestSender_unavailableBeyondTheDeadline(t *testing.T) {
	dev := target("dev")
	fake, srv := newFakeFCM(t, map[string][]fcmReply{dev.Token.Reveal(): {replyUnavailable}})
	sender := newSender(t, srv, 300*time.Millisecond)

	got, err := sender.Send(t.Context(), notification(), []usecase.PushTarget{dev})

	require.NoError(t, err)
	assert.Equal(t, []usecase.PushDelivery{{DeviceID: dev.DeviceID, Outcome: usecase.PushTransient, ErrorClass: "deadline"}}, got)
	assert.GreaterOrEqual(t, fake.callsFor(dev.Token.Reveal()), 2, "the SDK retries a 503")
}

func TestSender_timeout(t *testing.T) {
	slow, ok := target("slow"), target("ok")
	_, srv := newFakeFCM(t, map[string][]fcmReply{slow.Token.Reveal(): {replyBlock}})
	sender := newSender(t, srv, 100*time.Millisecond)

	start := time.Now()
	got, err := sender.Send(t.Context(), notification(), []usecase.PushTarget{slow, ok})

	require.NoError(t, err)
	assert.Less(t, time.Since(start), 5*time.Second, "the send deadline bounds the call")
	assert.Equal(t, []usecase.PushDelivery{
		{DeviceID: slow.DeviceID, Outcome: usecase.PushTransient, ErrorClass: "deadline"},
		{DeviceID: ok.DeviceID, Outcome: usecase.PushDelivered},
	}, got)
}

// The caller's own deadline wins over a longer SendTimeout.
func TestSender_callerDeadline(t *testing.T) {
	dev := target("slow")
	_, srv := newFakeFCM(t, map[string][]fcmReply{dev.Token.Reveal(): {replyBlock}})
	sender := newSender(t, srv, time.Minute)
	ctx, cancel := context.WithTimeout(t.Context(), 100*time.Millisecond)
	defer cancel()

	got, err := sender.Send(ctx, notification(), []usecase.PushTarget{dev})

	require.NoError(t, err)
	assert.Equal(t, usecase.PushTransient, got[0].Outcome)
	assert.Equal(t, "deadline", got[0].ErrorClass)
}

func TestSender_canceled(t *testing.T) {
	dev := target("dev")
	_, srv := newFakeFCM(t, nil)
	sender := newSender(t, srv, time.Second)
	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	got, err := sender.Send(ctx, notification(), []usecase.PushTarget{dev})

	require.NoError(t, err)
	assert.Equal(t, []usecase.PushDelivery{{DeviceID: dev.DeviceID, Outcome: usecase.PushTransient, ErrorClass: "canceled"}}, got)
}

func TestSender_unreachable(t *testing.T) {
	dev := target("dev")
	_, srv := newFakeFCM(t, nil)
	sender := newSender(t, srv, 300*time.Millisecond)
	srv.Close()

	got, err := sender.Send(t.Context(), notification(), []usecase.PushTarget{dev})

	require.NoError(t, err)
	assert.Equal(t, usecase.PushTransient, got[0].Outcome)
	assert.Contains(t, []string{"network", "deadline"}, got[0].ErrorClass, "the SDK retries a refused connection until the deadline")
}

// Nothing leaves the process when the arguments are wrong, and the error
// carries neither the token nor the text.
func TestSender_refusesBeforeSending(t *testing.T) {
	fake, srv := newFakeFCM(t, nil)
	sender := newSender(t, srv, time.Second)
	n := notification()
	n.CollapseID = strings.Repeat("x", usecase.MaxCollapseIDBytes+1)
	dev := target("dev")

	got, err := sender.Send(t.Context(), n, []usecase.PushTarget{dev})

	require.ErrorIs(t, err, usecase.ErrInvalidPushNotification)
	assert.Nil(t, got)
	assert.Empty(t, fake.messages())
	assert.NotContains(t, err.Error(), dev.Token.Reveal())
	assert.NotContains(t, err.Error(), n.Body)
}

func TestSender_noTargets(t *testing.T) {
	fake, srv := newFakeFCM(t, nil)
	sender := newSender(t, srv, time.Second)

	got, err := sender.Send(t.Context(), notification(), nil)

	require.NoError(t, err)
	assert.Empty(t, got)
	assert.Empty(t, fake.messages())
}

func TestNew_config(t *testing.T) {
	_, err := fcm.New(t.Context(), fcm.Config{})
	require.ErrorIs(t, err, fcm.ErrConfig)

	_, err = fcm.New(t.Context(), fcm.Config{ProjectID: testProject, SendTimeout: -time.Second})
	require.ErrorIs(t, err, fcm.ErrConfig)
}

// Malformed credentials fail at construction (the worker's startup), and the
// error shows none of what was passed in.
func TestNew_malformedCredentials(t *testing.T) {
	marker := uuid.NewString()
	creds := []byte(`{"type":"service_account","project_id":"` + marker + `","private_key":"` + marker + `"}`)

	_, err := fcm.New(t.Context(), fcm.Config{ProjectID: testProject, CredentialsJSON: creds})

	require.ErrorIs(t, err, fcm.ErrConfig)
	assert.NotContains(t, err.Error(), marker)
}
