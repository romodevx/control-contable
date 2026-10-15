// Package pushtest provides a fake usecase.PushSender for the reminders
// tests (the reminder job, M9-06): it records what was sent, without tokens,
// and answers each device with a scripted outcome.
package pushtest

import (
	"context"
	"sync"

	"github.com/google/uuid"

	"github.com/romodevx/control-contable/backend/internal/reminders/usecase"
)

// Send is one recorded call: the notification and the devices addressed, in
// order. Tokens are not recorded.
type Send struct {
	Notification usecase.PushNotification
	DeviceIDs    []uuid.UUID
}

// scripted is a device's programmed answer.
type scripted struct {
	outcome usecase.PushOutcome
	class   string
}

// Sender is a fake usecase.PushSender, safe for concurrent use. The zero
// value delivers to every device.
type Sender struct {
	// Err, when set, fails every call as a whole (nothing recorded).
	Err error

	mu       sync.Mutex
	outcomes map[uuid.UUID]scripted
	sends    []Send
}

var _ usecase.PushSender = (*Sender)(nil)

// SetOutcome makes every later Send answer device with outcome and class.
func (s *Sender) SetOutcome(device uuid.UUID, outcome usecase.PushOutcome, class string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.outcomes == nil {
		s.outcomes = map[uuid.UUID]scripted{}
	}
	s.outcomes[device] = scripted{outcome: outcome, class: class}
}

// Send implements usecase.PushSender: the same preconditions as the real
// sender (usecase.ValidatePush, the context), then the scripted outcomes.
func (s *Sender) Send(ctx context.Context, n usecase.PushNotification, targets []usecase.PushTarget) ([]usecase.PushDelivery, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if err := usecase.ValidatePush(n, targets); err != nil {
		return nil, err
	}
	if s.Err != nil {
		return nil, s.Err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	record := Send{Notification: n, DeviceIDs: make([]uuid.UUID, len(targets))}
	deliveries := make([]usecase.PushDelivery, len(targets))
	for i, target := range targets {
		record.DeviceIDs[i] = target.DeviceID
		answer := s.outcomes[target.DeviceID]
		deliveries[i] = usecase.PushDelivery{DeviceID: target.DeviceID, Outcome: answer.outcome, ErrorClass: answer.class}
	}
	s.sends = append(s.sends, record)
	return deliveries, nil
}

// Sends returns a copy of the recorded calls, in order.
func (s *Sender) Sends() []Send {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]Send(nil), s.sends...)
}
