package usecase

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/google/uuid"

	"github.com/romodevx/control-contable/backend/internal/auth/domain"
	"github.com/romodevx/control-contable/backend/internal/platform/tx"
)

// ErrInvalidDeviceID reports a logout device_id that is not a UUID.
var ErrInvalidDeviceID = errors.New("auth: logout device id is not a uuid")

// DeviceRemover removes one of the caller's push devices (the reminders
// module's devices table) inside the unit of work in ctx, so RLS scopes it
// to the caller. Another user's or an unknown id removes nothing.
type DeviceRemover interface {
	DeleteDevice(ctx context.Context, installationID string) (deleted bool, err error)
}

// LogoutDeps are the ports Logout needs.
type LogoutDeps struct {
	// Tx runs the revocation as the bearer token's user.
	Tx     tx.Manager
	Tokens FamilyRevoker
	// Devices removes the push device named by LogoutInput.DeviceID in the
	// revocation's unit of work; nil ignores DeviceID (a build without the
	// reminders module).
	Devices DeviceRemover
	// Logger receives the logout audit event; nil discards it.
	Logger *slog.Logger
}

// LogoutInput is a logout request: the authenticated caller, the refresh
// token whose family ends and, optionally, this installation's push device.
type LogoutInput struct {
	// UserID is the bearer access token's subject.
	UserID       string
	RefreshToken string
	// DeviceID is the push installation id removed with the session; "" for
	// none (M9 security gate F1).
	DeviceID string
}

// Logout revokes a refresh token family (US-A2c). It revokes, never
// deletes: the rotated tokens stay as reuse evidence until the family dies
// (security gate M1-02, condition 4).
type Logout struct {
	deps LogoutDeps
}

// NewLogout builds the use case.
func NewLogout(deps LogoutDeps) *Logout {
	if deps.Logger == nil {
		deps.Logger = slog.New(slog.DiscardHandler)
	}
	return &Logout{deps: deps}
}

// Execute revokes the family of the caller's refresh token and removes the
// caller's push device DeviceID, in one unit of work as the caller. It is
// idempotent: an unknown, already revoked or another user's token is a
// success, and another user's token is not even looked up, so its state is
// never revealed; an unknown or another user's device removes nothing. A
// malformed token is ErrTokenMalformed (the contract's RefreshRequest:
// malformed answers 401), a malformed device id ErrInvalidDeviceID; a
// caller that is not a user id is ErrInvalidAccessToken.
func (uc *Logout) Execute(ctx context.Context, in LogoutInput) error {
	caller, err := domain.ParseUserID(in.UserID)
	if err != nil {
		return fmt.Errorf("%w: logout caller: %w", ErrInvalidAccessToken, err)
	}
	token, err := domain.ParseRefreshToken(in.RefreshToken)
	if err != nil {
		return ErrTokenMalformed
	}
	device, err := uc.deviceToRemove(in.DeviceID)
	if err != nil {
		return err
	}
	ownToken := token.UserID() == caller
	if !ownToken && device == "" {
		return nil
	}
	err = uc.deps.Tx.ReadWrite(ctx, caller.String(), func(ctx context.Context) error {
		if ownToken {
			if err := uc.deps.Tokens.RevokeFamily(ctx, token.Hash(), RevokeLogout); err != nil {
				return err
			}
		}
		if device == "" {
			return nil
		}
		_, err := uc.deps.Devices.DeleteDevice(ctx, device)
		return err
	})
	if err != nil {
		return fmt.Errorf("auth: logout: %w", err)
	}
	if ownToken {
		uc.deps.Logger.InfoContext(ctx, "auth audit", slog.String("event", "logout"), slog.String("user_id", caller.String()))
	}
	return nil
}

// deviceToRemove is raw's canonical form, or "" when there is nothing to
// remove (no id, or no device store wired). A malformed raw is refused
// either way.
func (uc *Logout) deviceToRemove(raw string) (string, error) {
	if raw == "" {
		return "", nil
	}
	id, err := uuid.Parse(raw)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrInvalidDeviceID, err)
	}
	if uc.deps.Devices == nil {
		return "", nil
	}
	return id.String(), nil
}
