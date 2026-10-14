package authhttp

import (
	"context"
	"errors"

	"github.com/romodevx/control-contable/backend/adapter/http/gen"
	"github.com/romodevx/control-contable/backend/internal/auth/usecase"
	"github.com/romodevx/control-contable/backend/internal/platform/ratelimit"
	"github.com/romodevx/control-contable/backend/internal/platform/requestid"
)

// msgTooManyRequests is the 429 message, the same as the router's.
const msgTooManyRequests = "Too many requests. Try again later."

// tooManyRequests reports whether err is a *usecase.RateLimitedError and,
// if so, the contract's TooManyRequests response for it: Retry-After in
// whole seconds, rounded up, at least 1. The strict handlers return it as a
// typed response because an apperr error cannot carry the header.
func tooManyRequests(ctx context.Context, err error) (gen.TooManyRequestsJSONResponse, bool) {
	var limited *usecase.RateLimitedError
	if !errors.As(err, &limited) {
		return gen.TooManyRequestsJSONResponse{}, false
	}
	var body gen.ErrorResponse
	body.Error.Code = gen.ErrorCodeRateLimited
	body.Error.Message = msgTooManyRequests
	body.Error.RequestId = requestid.From(ctx)
	return gen.TooManyRequestsJSONResponse{
		Body:    body,
		Headers: gen.TooManyRequestsResponseHeaders{RetryAfter: ratelimit.RetryAfterSeconds(limited.RetryAfter)},
	}, true
}
