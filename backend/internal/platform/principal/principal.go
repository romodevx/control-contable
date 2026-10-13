// Package principal carries the authenticated caller through a request's
// context (M1-07). The HTTP authenticator puts it there after verifying the
// bearer access token; handlers and use cases read the user id from it and
// pass it to tx.Manager, whose unit of work sets app.current_user_id so Row
// Level Security sees that user and nobody else.
//
// Only the authenticator writes a Principal: nothing derived from request
// input (a path or body field) may stand in for it. The package is
// framework-free so every layer may import it.
package principal

import (
	"context"
	"errors"
)

// ErrUnauthenticated reports a context with no authenticated user.
var ErrUnauthenticated = errors.New("principal: no authenticated user")

// Principal is the caller a verified access token names.
type Principal struct {
	// UserID is the canonical (lower-case) UUID of the user: the token's sub.
	UserID string
	// TokenID is the access token's jti, for audit and correlation.
	TokenID string
}

// contextKey is unexported so no other package can read or forge the value.
type contextKey struct{}

// With returns a copy of ctx carrying p.
func With(ctx context.Context, p Principal) context.Context {
	return context.WithValue(ctx, contextKey{}, p)
}

// From returns the principal carried by ctx, if any.
func From(ctx context.Context) (Principal, bool) {
	p, ok := ctx.Value(contextKey{}).(Principal)
	return p, ok
}

// UserID returns the authenticated user's id, or ErrUnauthenticated: a
// missing principal (or one without a user) never becomes an empty user id
// handed to a unit of work.
func UserID(ctx context.Context) (string, error) {
	p, ok := From(ctx)
	if !ok || p.UserID == "" {
		return "", ErrUnauthenticated
	}
	return p.UserID, nil
}
