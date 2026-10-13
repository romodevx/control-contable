// Package requestid carries the request correlation id in a context, so the
// error envelope and every log line of a request report the same id.
package requestid

import "context"

type ctxKey struct{}

// With returns a copy of ctx carrying id.
func With(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, ctxKey{}, id)
}

// From returns the request id in ctx, or "" when there is none.
func From(ctx context.Context) string {
	id, _ := ctx.Value(ctxKey{}).(string)
	return id
}
