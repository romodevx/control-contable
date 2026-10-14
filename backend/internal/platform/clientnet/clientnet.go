// Package clientnet carries the caller's network (IPv4 /24, IPv6 /64), as
// the HTTP edge resolved it behind the trusted proxies, in a context. Use
// cases receive it as plain input; it keys per-network budgets such as the
// login failure limit (ADR-0005), never authorization.
package clientnet

import "context"

type ctxKey struct{}

// With returns a copy of ctx carrying network.
func With(ctx context.Context, network string) context.Context {
	return context.WithValue(ctx, ctxKey{}, network)
}

// From returns the client network in ctx, or "" when there is none.
func From(ctx context.Context) string {
	network, _ := ctx.Value(ctxKey{}).(string)
	return network
}
