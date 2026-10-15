package usecase_test

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/reminders/domain"
	"github.com/romodevx/control-contable/backend/internal/reminders/usecase"
)

// A push token is a device credential: no printing path may show it.
func TestPushToken_neverPrinted(t *testing.T) {
	raw := "tok-" + uuid.NewString()
	token := usecase.NewPushToken(raw)

	var logs strings.Builder
	slog.New(slog.NewJSONHandler(&logs, nil)).Info("send", slog.Any("token", token))
	encoded, err := json.Marshal(struct{ T usecase.PushToken }{token})
	require.NoError(t, err)

	printed := []string{
		fmt.Sprint(token), fmt.Sprintf("%v %+v %#v %s %q %x", token, token, token, token, token, token),
		token.String(), logs.String(), string(encoded),
		fmt.Sprintf("%+v", usecase.PushTarget{DeviceID: uuid.New(), Token: token}),
	}
	for _, p := range printed {
		assert.NotContains(t, p, raw)
	}
	assert.Equal(t, raw, token.Reveal())
	assert.False(t, token.IsZero())
	assert.True(t, usecase.PushToken{}.IsZero())
}

func TestPushOutcome_String(t *testing.T) {
	tests := []struct {
		outcome usecase.PushOutcome
		want    string
	}{
		{usecase.PushDelivered, "delivered"},
		{usecase.PushTokenInvalid, "token_invalid"},
		{usecase.PushTransient, "transient"},
		{usecase.PushRejected, "rejected"},
		{usecase.PushOutcome(99), "unknown"},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.want, tt.outcome.String())
	}
}

// The disabled sender is what the worker wires when push is not configured:
// it delivers nothing and says why, so the job records it instead of
// counting a silent success.
func TestDisabledPushSender(t *testing.T) {
	targets := []usecase.PushTarget{{DeviceID: uuid.New(), Token: usecase.NewPushToken("t")}}

	got, err := usecase.DisabledPushSender{}.Send(context.Background(), usecase.PushNotification{}, targets)

	require.ErrorIs(t, err, usecase.ErrPushDisabled)
	assert.Nil(t, got)
}

func TestValidatePush(t *testing.T) {
	valid := usecase.PushNotification{
		Title: "Hoy vence", Body: "Revisa", DeepLink: "controlcontable://estimate?period=2026-04",
		CollapseID: strings.Repeat("a", usecase.MaxCollapseIDBytes),
	}
	one := []usecase.PushTarget{{DeviceID: uuid.New(), Token: usecase.NewPushToken("t")}}
	tooMany := make([]usecase.PushTarget, usecase.MaxPushTargets+1)
	for i := range tooMany {
		tooMany[i] = one[0]
	}
	with := func(edit func(*usecase.PushNotification)) usecase.PushNotification {
		n := valid
		edit(&n)
		return n
	}
	tests := []struct {
		name    string
		n       usecase.PushNotification
		targets []usecase.PushTarget
		want    error
	}{
		{"valid", valid, one, nil},
		{"no targets is valid", valid, nil, nil},
		{"collapse id optional", with(func(n *usecase.PushNotification) { n.CollapseID = "" }), one, nil},
		{"no title", with(func(n *usecase.PushNotification) { n.Title = "" }), one, usecase.ErrInvalidPushNotification},
		{"no body", with(func(n *usecase.PushNotification) { n.Body = "" }), one, usecase.ErrInvalidPushNotification},
		{"collapse id over apns limit", with(func(n *usecase.PushNotification) {
			n.CollapseID = strings.Repeat("a", usecase.MaxCollapseIDBytes+1)
		}), one, usecase.ErrInvalidPushNotification},
		// M9 security gate F2: a message FCM could refuse for its size never
		// leaves the process, so an INVALID_ARGUMENT is never the payload's.
		{"payload at the cap", with(func(n *usecase.PushNotification) { n.Body = fillPayload(*n, 0) }), one, nil},
		{"payload over the cap", with(func(n *usecase.PushNotification) { n.Body = fillPayload(*n, 1) }), one, usecase.ErrInvalidPushNotification},
		{"blank token", valid, []usecase.PushTarget{{DeviceID: uuid.New()}}, usecase.ErrInvalidPushNotification},
		{"too many targets", valid, tooMany, usecase.ErrTooManyPushTargets},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := usecase.ValidatePush(tt.n, tt.targets)

			if tt.want == nil {
				require.NoError(t, err)
				return
			}
			require.ErrorIs(t, err, tt.want)
			assert.NotContains(t, err.Error(), valid.Body)
		})
	}
}

// fillPayload is a body that brings n's payload to MaxPushPayloadBytes+over
// bytes, in multi-byte runes (the cap counts bytes, not characters).
func fillPayload(n usecase.PushNotification, over int) string {
	room := usecase.MaxPushPayloadBytes + over - len(n.Title) - len(n.DeepLink) - len(n.CollapseID)
	return strings.Repeat("ñ", room/2) + strings.Repeat("a", room%2)
}

// The longest reminder BuildMessage can write (longest Spanish date, three
// capped names of 4-byte characters with the largest amounts, "y N más")
// still fits MaxPushPayloadBytes: the cap never drops a real reminder.
func TestValidatePush_theLongestReminderFits(t *testing.T) {
	period, err := domain.NewPeriod(2026, time.September)
	require.NoError(t, err)
	name := strings.Repeat("𝕏", domain.MaxListedNameRunes+10)
	listed := domain.Listed{Name: name, Estimate: domain.EstimateAmount, AmountCents: math.MaxInt64}
	msg, err := domain.BuildMessage(domain.MessageInput{
		Kind: domain.KindEarly, Period: period, DueDate: time.Date(2026, time.September, 30, 0, 0, 0, 0, time.UTC),
		Taxpayers: []domain.Listed{listed, listed, listed, listed},
	})
	require.NoError(t, err)
	n := usecase.PushNotification{
		Title: msg.Title, Body: msg.Body, DeepLink: msg.DeepLink,
		CollapseID: strings.Repeat("c", usecase.MaxCollapseIDBytes),
	}

	require.NoError(t, usecase.ValidatePush(n, []usecase.PushTarget{{DeviceID: uuid.New(), Token: usecase.NewPushToken("t")}}))
}
