package usecase

import (
	"time"

	"github.com/romodevx/control-contable/backend/internal/platform/ratelimit"
)

// Per-account limits (MODULE-auth §5, ADR-0005). Each is a shared bucket per
// key, refilled evenly over its window: once spent, one more event is
// admitted every window/limit (the backoff). The per-client limits of each
// operation live in the HTTP adapter (authhttp.ClientLimits).
var (
	// LoginFailuresPolicy: 5 failures / 15 min per email blind index
	// (known or unknown address alike); a success refills it.
	LoginFailuresPolicy = ratelimit.PerWindow("auth_login_email", 5, 15*time.Minute)
	// SignUpEmailPolicy: 5 / h per email blind index, shared by register
	// and resendVerification (D7).
	SignUpEmailPolicy = ratelimit.PerWindow("auth_signup_email", 5, time.Hour)
	// RefreshFamilyPolicy: 20 rotations / h per token family. A client
	// refreshes about once per access-token lifetime (15 min); this only
	// stops a runaway or replaying client.
	RefreshFamilyPolicy = ratelimit.PerWindow("auth_refresh_family", 20, time.Hour)
	// DeleteAttemptsPolicy: 5 password attempts, then 1 / 3 min per account (D10).
	DeleteAttemptsPolicy = ratelimit.PerWindow("auth_delete_password", 5, 15*time.Minute)
)
