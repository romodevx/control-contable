package authhttp

import (
	"time"

	"github.com/romodevx/control-contable/backend/internal/platform/ratelimit"
)

// ClientLimits are the per-client limits of MODULE-auth §5 by operationId,
// keyed by the client address (IPv6 by /64). cmd/api hands them to the
// router, which applies them right after the flag gate. They sit on top of
// the edge limiter and of the per-account limits in the use cases.
var ClientLimits = map[string]ratelimit.Policy{
	"register":           ratelimit.PerWindow("client_register", 10, time.Hour),
	"verifyEmail":        ratelimit.PerWindow("client_verify_email", 20, 15*time.Minute),
	"resendVerification": ratelimit.PerWindow("client_resend_verification", 10, time.Hour),
	"login":              ratelimit.PerWindow("client_login", 20, 15*time.Minute),
	"refreshSession":     ratelimit.PerWindow("client_refresh_session", 60, time.Minute),
	"listLegalDocuments": ratelimit.PerWindow("client_legal_documents", 60, time.Minute),
}
