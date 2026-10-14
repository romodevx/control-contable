package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/romodevx/control-contable/backend/adapter/http/middleware"
	"github.com/romodevx/control-contable/backend/internal/platform/clientnet"
)

// ClientNetwork resolves the client like ClientIP (same trusted hops, no
// spoofing via leading X-Forwarded-For entries) but widens IPv4 to its /24:
// the per-(email, network) login budget (ADR-0005) must not be dodged by
// rotating addresses within one network.
func TestClientNetwork(t *testing.T) {
	tests := []struct {
		name       string
		hops       int
		remoteAddr string
		xff        []string
		want       string
	}{
		{name: "IPv4 peer is keyed by its /24", remoteAddr: "198.51.100.77:5000", want: "198.51.100.0/24"},
		{name: "X-Forwarded-For ignored without trusted proxies", remoteAddr: "198.51.100.1:5000", xff: []string{"1.2.3.4"}, want: "198.51.100.0/24"},
		{name: "spoofed leading entries are skipped", hops: 2, remoteAddr: "10.0.0.1:80", xff: []string{"6.6.6.6, 203.0.113.9, 10.1.1.1"}, want: "203.0.113.0/24"},
		{name: "fewer entries than hops falls back to the peer", hops: 2, remoteAddr: "10.0.0.1:80", xff: []string{"203.0.113.9"}, want: "10.0.0.0/24"},
		{name: "IPv6 is keyed by its /64", remoteAddr: "[2001:db8:1:2:aaaa::1]:443", want: "2001:db8:1:2::/64"},
		{name: "IPv4-mapped IPv6 is IPv4", remoteAddr: "[::ffff:198.51.100.1]:443", want: "198.51.100.0/24"},
		{name: "unparsable peer is kept verbatim", remoteAddr: "pipe", want: "pipe"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.RemoteAddr = tt.remoteAddr
			for _, v := range tt.xff {
				req.Header.Add("X-Forwarded-For", v)
			}
			assert.Equal(t, tt.want, middleware.ClientNetwork(tt.hops)(req))
		})
	}
}

func TestWithClientNetwork_putsTheKeyInTheContext(t *testing.T) {
	var got string
	h := middleware.WithClientNetwork(func(*http.Request) string { return "203.0.113.0/24" })(
		http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) { got = clientnet.From(r.Context()) }))

	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))

	assert.Equal(t, "203.0.113.0/24", got)
}
