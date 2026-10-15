package router

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/romodevx/control-contable/backend/adapter/http/httperr"
	"github.com/romodevx/control-contable/backend/internal/platform/apperr"
	"github.com/romodevx/control-contable/backend/internal/platform/principal"
	"github.com/romodevx/control-contable/backend/internal/platform/ratelimit"
)

// Limiter spends one event of a key's budget for an operation
// (ratelimitpg.Bucket). When none is left it answers allowed = false and the
// wait; an error means it could not decide.
type Limiter interface {
	Take(ctx context.Context, key string) (allowed bool, retryAfter time.Duration, err error)
}

// rateLimitedMessage is the 429 message, the same as the edge limiter's.
const rateLimitedMessage = "Too many requests. Try again later."

const limiterFailedMessage = "Something went wrong."

// limit answers 429 with Retry-After when limiter denies the request's
// client key, and the generic 500 (cause logged) when it cannot decide:
// a limit that fails open would be no limit at all.
func limit(limiter Limiter, clientKey func(*http.Request) string, logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		allowed, wait, err := limiter.Take(r.Context(), clientKey(r))
		switch {
		case err != nil:
			logger.ErrorContext(r.Context(), "rate limit unavailable", slog.Any("error", err))
			_ = httperr.Write(w, r, apperr.Mapper{}, apperr.Wrap(err, apperr.CodeInternalError, limiterFailedMessage))
		case !allowed:
			w.Header().Set("Retry-After", strconv.Itoa(ratelimit.RetryAfterSeconds(wait)))
			_ = httperr.Write(w, r, apperr.Mapper{}, apperr.New(apperr.CodeRateLimited, rateLimitedMessage))
		default:
			next.ServeHTTP(w, r)
		}
	})
}

// principalKey keys a per-user limit by the authenticated user. The router
// places it behind authentication, so the principal is always there; an
// empty key would still be one shared bucket, never an unlimited request.
func principalKey(r *http.Request) string {
	userID, _ := principal.UserID(r.Context())
	return userID
}

// checkLimits fails on limits the router could not apply: an operation the
// contract does not declare (a typo would leave it unlimited), client
// limits with no way to key the client, or a user limit on a public
// operation (no user to key it by).
func checkLimits(opts Options, declared, public map[string]bool) error {
	var errs []error
	if len(opts.OperationLimits) > 0 && opts.ClientKey == nil {
		errs = append(errs, errors.New("OperationLimits need a ClientKey"))
	}
	for id := range opts.OperationLimits {
		if !declared[id] {
			errs = append(errs, fmt.Errorf("limited operation %q is not in the contract", id))
		}
	}
	for id := range opts.UserLimits {
		switch {
		case !declared[id]:
			errs = append(errs, fmt.Errorf("user-limited operation %q is not in the contract", id))
		case public[id]:
			errs = append(errs, fmt.Errorf("user-limited operation %q is public: it has no user to key by", id))
		}
	}
	return errors.Join(errs...)
}
