package middleware

import (
	"net/http"
	"net/netip"

	"github.com/romodevx/control-contable/backend/internal/platform/clientnet"
)

// ipv4NetworkBits is the IPv4 prefix that identifies one client network for
// per-network budgets: addresses within a /24 are cheap to rotate.
const ipv4NetworkBits = 24

// ClientNetwork returns a KeyFunc keyed by the client's network: the address
// ClientIP resolves (same trusted hops, same fallbacks), widened to its /24
// for IPv4 (IPv6 is already a /64).
func ClientNetwork(trustedHops int) KeyFunc {
	clientIP := ClientIP(trustedHops)
	return func(r *http.Request) string {
		k := clientIP(r)
		addr, err := netip.ParseAddr(k)
		if err != nil || !addr.Is4() {
			return k // an IPv6 /64 or an unparsable peer, already final
		}
		prefix, err := addr.Prefix(ipv4NetworkBits)
		if err != nil {
			return k
		}
		return prefix.String()
	}
}

// WithClientNetwork stores key's value for each request in its context
// (clientnet), for handlers whose use cases keep per-network budgets.
func WithClientNetwork(key KeyFunc) Func {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r.WithContext(clientnet.With(r.Context(), key(r))))
		})
	}
}
