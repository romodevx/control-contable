package middleware

import (
	"net"
	"net/http"
	"net/netip"
	"strconv"
	"strings"
	"time"

	"github.com/romodevx/control-contable/backend/adapter/http/httperr"
	"github.com/romodevx/control-contable/backend/internal/platform/apperr"
	"github.com/romodevx/control-contable/backend/internal/platform/ratelimit"
)

// Limiter decides whether a request keyed by key may proceed and, if not,
// how long the client should wait. platform/ratelimit implements it.
type Limiter interface {
	Allow(key string) (allowed bool, retryAfter time.Duration)
}

// KeyFunc derives the rate-limit key of a request.
type KeyFunc func(*http.Request) string

const rateLimitedMessage = "Too many requests. Try again later."

// ipv6KeyBits is the IPv6 prefix that identifies one client: an end site
// usually gets a whole /64, so per-address keys would be trivial to rotate.
const ipv6KeyBits = 64

// RateLimit answers 429 rate_limited with a Retry-After header (whole
// seconds, at least 1, as the contract's TooManyRequests requires) when
// limiter denies the request's key.
func RateLimit(limiter Limiter, key KeyFunc) Func {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			allowed, wait := limiter.Allow(key(r))
			if allowed {
				next.ServeHTTP(w, r)
				return
			}
			w.Header().Set("Retry-After", strconv.Itoa(ratelimit.RetryAfterSeconds(wait)))
			_ = httperr.Write(w, r, apperr.Mapper{}, apperr.New(apperr.CodeRateLimited, rateLimitedMessage))
		})
	}
}

// ClientIP returns a KeyFunc keyed by the client address. With trustedHops
// = 0 that is the TCP peer. With n > 0 reverse proxies in front of the
// server, each appending to X-Forwarded-For, it is the n-th entry from the
// right: entries further left are client-supplied and spoofable. When the
// header is shorter than n or the entry is not an IP, it falls back to the
// peer. IPv6 clients are keyed by their /64.
func ClientIP(trustedHops int) KeyFunc {
	return func(r *http.Request) string {
		peer := peerAddr(r.RemoteAddr)
		if trustedHops == 0 {
			return key(peer, r.RemoteAddr)
		}
		entries := forwardedFor(r)
		if len(entries) < trustedHops {
			return key(peer, r.RemoteAddr)
		}
		addr, err := netip.ParseAddr(entries[len(entries)-trustedHops])
		if err != nil {
			return key(peer, r.RemoteAddr)
		}
		return key(addr, "")
	}
}

func peerAddr(remoteAddr string) netip.Addr {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		host = remoteAddr
	}
	addr, err := netip.ParseAddr(host)
	if err != nil {
		return netip.Addr{}
	}
	return addr
}

// key renders addr as a limiter key; an invalid addr yields fallback.
func key(addr netip.Addr, fallback string) string {
	if !addr.IsValid() {
		return fallback
	}
	addr = addr.Unmap()
	if addr.Is4() {
		return addr.String()
	}
	prefix, err := addr.Prefix(ipv6KeyBits)
	if err != nil {
		return addr.String()
	}
	return prefix.String()
}

// forwardedFor flattens every X-Forwarded-For header into one list.
func forwardedFor(r *http.Request) []string {
	var entries []string
	for _, header := range r.Header.Values("X-Forwarded-For") {
		for entry := range strings.SplitSeq(header, ",") {
			entries = append(entries, strings.TrimSpace(entry))
		}
	}
	return entries
}
