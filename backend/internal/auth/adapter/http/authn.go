package authhttp

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/romodevx/control-contable/backend/internal/auth/usecase"
	"github.com/romodevx/control-contable/backend/internal/platform/httperr"
	"github.com/romodevx/control-contable/backend/internal/platform/principal"
)

// bearerScheme is the only accepted authorization scheme (RFC 6750 §2.1).
// Scheme names are case-insensitive (RFC 9110 §11.1).
const bearerScheme = "bearer"

// errNoBearer reports a request without exactly one well-formed
// "Authorization: Bearer <token>" header.
var errNoBearer = fmt.Errorf("%w: no well-formed bearer credentials", usecase.ErrInvalidAccessToken)

// Authenticator is the authentication step of the per-operation pipeline
// (adapter/http/router.Authenticator, P0-17's slot): it admits a request
// carrying a valid bearer access token and puts its principal in the request
// context, where handlers read the user id for tx.Manager. Anything else
// gets the one 401 envelope, so the client never learns why. The router runs
// it after the flag gate and before contract validation, on the operations
// that declare a security requirement.
type Authenticator struct {
	verifier usecase.AccessTokenVerifier
	logger   *slog.Logger
}

// NewAuthenticator returns an Authenticator over verifier; a nil logger
// discards the rejection reasons.
func NewAuthenticator(verifier usecase.AccessTokenVerifier, logger *slog.Logger) *Authenticator {
	if logger == nil {
		logger = slog.New(slog.DiscardHandler)
	}
	return &Authenticator{verifier: verifier, logger: logger}
}

// Authenticate implements router.Authenticator.
func (a *Authenticator) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, err := a.verify(r)
		if err != nil {
			// The reason (never the token) helps debugging; the client
			// gets the generic answer.
			a.logger.DebugContext(r.Context(), "authentication refused", slog.String("reason", err.Error()))
			_ = httperr.Write(w, r, ErrorMapper, err)
			return
		}
		ctx := principal.With(r.Context(), principal.Principal{UserID: claims.UserID.String(), TokenID: claims.TokenID})
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (a *Authenticator) verify(r *http.Request) (usecase.AccessClaims, error) {
	token, ok := bearerToken(r.Header)
	if !ok {
		return usecase.AccessClaims{}, errNoBearer
	}
	claims, err := a.verifier.Verify(token)
	if err != nil {
		if !errors.Is(err, usecase.ErrInvalidAccessToken) {
			err = fmt.Errorf("%w: %w", usecase.ErrInvalidAccessToken, err)
		}
		return usecase.AccessClaims{}, err
	}
	return claims, nil
}

// bearerToken extracts the token of the one Authorization header, which must
// be "Bearer" (any case), one space, and an RFC 6750 b64token.
func bearerToken(h http.Header) (string, bool) {
	values := h.Values("Authorization")
	if len(values) != 1 {
		return "", false
	}
	scheme, token, found := strings.Cut(values[0], " ")
	if !found || !strings.EqualFold(scheme, bearerScheme) || !isB64Token(token) {
		return "", false
	}
	return token, true
}

// isB64Token reports whether s matches RFC 6750's
// b64token = 1*( ALPHA / DIGIT / "-" / "." / "_" / "~" / "+" / "/" ) *"=".
func isB64Token(s string) bool {
	body := strings.TrimRight(s, "=")
	if body == "" {
		return false
	}
	for _, c := range []byte(body) {
		if !isB64TokenChar(c) {
			return false
		}
	}
	return true
}

func isB64TokenChar(c byte) bool {
	switch {
	case c >= 'a' && c <= 'z', c >= 'A' && c <= 'Z', c >= '0' && c <= '9':
		return true
	default:
		return strings.IndexByte("-._~+/", c) >= 0
	}
}
