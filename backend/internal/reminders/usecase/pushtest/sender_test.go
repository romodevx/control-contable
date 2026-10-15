package pushtest_test

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/reminders/usecase"
	"github.com/romodevx/control-contable/backend/internal/reminders/usecase/pushtest"
)

func notification() usecase.PushNotification {
	return usecase.PushNotification{Title: "Hoy vence", Body: "Revisa tu estimado"}
}

func device() usecase.PushTarget {
	return usecase.PushTarget{DeviceID: uuid.New(), Token: usecase.NewPushToken(uuid.NewString())}
}

func TestSender_deliversByDefaultAndRecords(t *testing.T) {
	sender := &pushtest.Sender{}
	a, b := device(), device()

	got, err := sender.Send(context.Background(), notification(), []usecase.PushTarget{a, b})

	require.NoError(t, err)
	assert.Equal(t, []usecase.PushDelivery{
		{DeviceID: a.DeviceID, Outcome: usecase.PushDelivered},
		{DeviceID: b.DeviceID, Outcome: usecase.PushDelivered},
	}, got)
	sends := sender.Sends()
	require.Len(t, sends, 1)
	assert.Equal(t, notification(), sends[0].Notification)
	assert.Equal(t, []uuid.UUID{a.DeviceID, b.DeviceID}, sends[0].DeviceIDs)
}

func TestSender_scriptedOutcomes(t *testing.T) {
	stale, down := device(), device()
	sender := &pushtest.Sender{}
	sender.SetOutcome(stale.DeviceID, usecase.PushTokenInvalid, "unregistered")
	sender.SetOutcome(down.DeviceID, usecase.PushTransient, "unavailable")

	got, err := sender.Send(context.Background(), notification(), []usecase.PushTarget{stale, down})

	require.NoError(t, err)
	assert.Equal(t, []usecase.PushDelivery{
		{DeviceID: stale.DeviceID, Outcome: usecase.PushTokenInvalid, ErrorClass: "unregistered"},
		{DeviceID: down.DeviceID, Outcome: usecase.PushTransient, ErrorClass: "unavailable"},
	}, got)
}

func TestSender_failsWholeCall(t *testing.T) {
	sender := &pushtest.Sender{Err: usecase.ErrPushDisabled}

	got, err := sender.Send(context.Background(), notification(), []usecase.PushTarget{device()})

	require.ErrorIs(t, err, usecase.ErrPushDisabled)
	assert.Nil(t, got)
	assert.Empty(t, sender.Sends(), "a failed call delivered nothing")
}

// The fake enforces the port's preconditions like the real sender, so a
// message-builder bug fails the job's tests.
func TestSender_validates(t *testing.T) {
	sender := &pushtest.Sender{}

	_, err := sender.Send(context.Background(), usecase.PushNotification{}, []usecase.PushTarget{device()})

	require.ErrorIs(t, err, usecase.ErrInvalidPushNotification)
	assert.Empty(t, sender.Sends())
}

func TestSender_honorsContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := (&pushtest.Sender{}).Send(ctx, notification(), []usecase.PushTarget{device()})

	require.True(t, errors.Is(err, context.Canceled))
}

func TestSender_concurrentUse(t *testing.T) {
	sender := &pushtest.Sender{}
	const senders = 8
	var wg sync.WaitGroup
	for range senders {
		wg.Go(func() {
			_, err := sender.Send(context.Background(), notification(), []usecase.PushTarget{device()})
			assert.NoError(t, err)
		})
	}
	wg.Wait()

	assert.Len(t, sender.Sends(), senders)
}
