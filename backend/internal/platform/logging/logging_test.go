package logging_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/platform/logging"
	"github.com/romodevx/control-contable/backend/internal/platform/requestid"
)

const (
	email        = "ana.lopez+cc@example.com.mx"
	rfcPF        = "GODE561231GR8"
	rfcPM        = "ABC680524P76"
	jwt          = "eyJhbGciOiJFZERTQSJ9.eyJzdWIiOiIxMjMifQ.c2lnbmF0dXJl"
	refreshToken = userID + ".Zm9vYmFyYmF6cXV4cXV1eA" // fixture shape <uuid>.<secret>
	userID       = "0199d4a2-8f00-7c44-a1b2-3c4d5e6f7a80"
)

// sensitive lists every value that must never appear in any log line.
var sensitive = []string{email, rfcPF, rfcPM, jwt, refreshToken, "hunter2", "Zm9vYmFy"}

func logLine(t *testing.T, log func(*slog.Logger)) (map[string]any, string) {
	t.Helper()
	var buf bytes.Buffer
	log(logging.New(&buf, slog.LevelDebug))
	raw := buf.String()
	for _, secret := range sensitive {
		assert.NotContains(t, strings.ToLower(raw), strings.ToLower(secret))
	}
	var line map[string]any
	require.NoError(t, json.Unmarshal([]byte(raw), &line), raw)
	return line, raw
}

func TestHandler_redactsSensitiveKeys(t *testing.T) {
	tests := []struct {
		name string
		log  func(*slog.Logger)
		path []string
	}{
		{"rfc", func(l *slog.Logger) { l.Info("m", "rfc", "anything") }, []string{"rfc"}},
		{"email", func(l *slog.Logger) { l.Info("m", "email", "x") }, []string{"email"}},
		{"suffix and case", func(l *slog.Logger) { l.Info("m", "User-Email", "x") }, []string{"User-Email"}},
		{"password", func(l *slog.Logger) { l.Info("m", "password", "hunter2") }, []string{"password"}},
		{"refresh_token", func(l *slog.Logger) { l.Info("m", "refresh_token", "x") }, []string{"refresh_token"}},
		{"authorization", func(l *slog.Logger) { l.Info("m", "Authorization", "x") }, []string{"Authorization"}},
		{"api key", func(l *slog.Logger) { l.Info("m", "api_key", "x") }, []string{"api_key"}},
		{"curp", func(l *slog.Logger) { l.Info("m", "curp", "x") }, []string{"curp"}},
		{"non-string value", func(l *slog.Logger) { l.Info("m", "rfc", 42) }, []string{"rfc"}},
		{
			"inside a group",
			func(l *slog.Logger) { l.Info("m", slog.Group("user", "email", "x")) },
			[]string{"user", "email"},
		},
		{
			"whole sensitive group",
			func(l *slog.Logger) { l.Info("m", slog.Group("credentials", "user", "x")) },
			[]string{"credentials", "user"},
		},
		{"WithAttrs", func(l *slog.Logger) { l.With("email", "x").Info("m") }, []string{"email"}},
		{
			"WithGroup",
			func(l *slog.Logger) { l.WithGroup("req").Info("m", "token", "x") },
			[]string{"req", "token"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			line, _ := logLine(t, tt.log)

			assert.Equal(t, logging.Redacted, valueAt(t, line, tt.path))
		})
	}
}

func TestHandler_scrubsSensitiveValues(t *testing.T) {
	tests := []struct {
		name string
		log  func(*slog.Logger)
		key  string
		want string
	}{
		{"email in message", func(l *slog.Logger) { l.Info("login failed for " + email) }, "msg", "login failed for [REDACTED]"},
		{"rfc PF", func(l *slog.Logger) { l.Info("m", "detail", "rfc="+rfcPF+" dup") }, "detail", "rfc=[REDACTED] dup"},
		{"rfc PM lowercase", func(l *slog.Logger) { l.Info("m", "detail", strings.ToLower(rfcPM)) }, "detail", "[REDACTED]"},
		{"bearer header", func(l *slog.Logger) { l.Info("m", "hdr", "Bearer "+jwt) }, "hdr", "[REDACTED]"},
		{"bare jwt", func(l *slog.Logger) { l.Info("m", "detail", "tok "+jwt) }, "detail", "tok [REDACTED]"},
		{"refresh token", func(l *slog.Logger) { l.Info("m", "detail", refreshToken) }, "detail", "[REDACTED]"},
		{
			"error value",
			func(l *slog.Logger) { l.Error("m", "err", fmt.Errorf("register: %w", errors.New("dup "+email))) },
			"err", "register: dup [REDACTED]",
		},
		{
			"struct value",
			func(l *slog.Logger) { l.Info("m", "payload", struct{ To string }{email}) },
			"payload", "{To:[REDACTED]}",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			line, _ := logLine(t, tt.log)

			assert.Equal(t, tt.want, line[tt.key])
		})
	}
}

// TestHandler_scrubsCredentialsInText covers secrets that reach a log line
// under a harmless key, inside free text: error strings, dumped headers,
// DSNs, JSON bodies and structs printed by fmt.
func TestHandler_scrubsCredentialsInText(t *testing.T) {
	const (
		password   = "hunter2"
		basicCreds = "dXNlcjpodW50ZXIy" //nolint:gosec // G101: fake fixture the redactor must hide
		sessionID  = "s3ss10nV4lu3"
		apiKey     = "ak_live_51Hx" //nolint:gosec // G101: fake fixture the redactor must hide
		curp       = "GODE561231HDFRRN09"
	)
	tests := []struct {
		name   string
		log    func(*slog.Logger)
		secret string
	}{
		{
			"dumped http.Header authorization",
			func(l *slog.Logger) {
				l.Info("m", "headers", http.Header{"Authorization": {"Basic " + basicCreds}})
			},
			basicCreds,
		},
		{
			"dumped http.Header cookie",
			func(l *slog.Logger) { l.Info("m", "headers", http.Header{"Cookie": {"session=" + sessionID}}) },
			sessionID,
		},
		{
			"authorization header line in error",
			func(l *slog.Logger) {
				l.Error("m", "err", errors.New("upstream rejected Authorization: Token "+apiKey))
			},
			apiKey,
		},
		{
			"password in url userinfo",
			func(l *slog.Logger) {
				l.Error("m", "err", errors.New("connect postgres://app:"+password+"@localhost/db: refused"))
			},
			password,
		},
		{
			"password in key=value dsn",
			func(l *slog.Logger) { l.Error("m", "err", errors.New("dsn host=db password="+password+" user=app")) },
			password,
		},
		{
			"password in json text",
			func(l *slog.Logger) {
				l.Error("m", "err", fmt.Errorf("decode: %w", errors.New(`{"password":"`+password+`"}`)))
			},
			password,
		},
		{
			"api key in query string",
			func(l *slog.Logger) { l.Info("m", "url", "https://api.example.com/v1?api_key="+apiKey+"&x=1") },
			apiKey,
		},
		{
			"struct field",
			func(l *slog.Logger) { l.Info("m", "payload", struct{ Password string }{password}) },
			password,
		},
		{
			"map entry",
			func(l *slog.Logger) { l.Info("m", "payload", map[string]string{"password": password}) },
			password,
		},
		{
			"whole curp, not only its rfc-shaped prefix",
			func(l *slog.Logger) { l.Info("m", "detail", "curp "+curp) },
			"RRN09",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, raw := logLine(t, tt.log)

			assert.NotContains(t, raw, tt.secret)
			assert.Contains(t, raw, logging.Redacted)
		})
	}
}

func TestHandler_keepsOrdinaryValues(t *testing.T) {
	line, _ := logLine(t, func(l *slog.Logger) {
		l.Info("income recorded",
			"user_id", userID,
			"amount_cents", 125000,
			"period", "2026-09",
			"regime_code", "612",
			"request_id_hint", "req_01J9Z6Q4X8",
			"ok", true,
		)
	})

	assert.Equal(t, "income recorded", line["msg"])
	assert.Equal(t, "INFO", line["level"])
	assert.Equal(t, userID, line["user_id"])
	assert.InDelta(t, 125000, line["amount_cents"], 0)
	assert.Equal(t, "2026-09", line["period"])
	assert.Equal(t, "612", line["regime_code"])
	assert.Equal(t, "req_01J9Z6Q4X8", line["request_id_hint"])
	assert.Equal(t, true, line["ok"])
	assert.Contains(t, line, "time")
}

func TestHandler_addsRequestIDFromContext(t *testing.T) {
	withID := requestid.With(context.Background(), "req_01J9Z6Q4X8")
	shapes := map[string]func(*slog.Logger) *slog.Logger{
		"plain":       func(l *slog.Logger) *slog.Logger { return l },
		"with attrs":  func(l *slog.Logger) *slog.Logger { return l.With("component", "api") },
		"grouped":     func(l *slog.Logger) *slog.Logger { return l.With("component", "api").WithGroup("g").With("k", "v") },
		"empty group": func(l *slog.Logger) *slog.Logger { return slog.New(l.Handler().WithGroup("")) },
	}
	tests := []struct {
		name   string
		ctx    context.Context
		want   any
		wantOK bool
	}{
		{name: "present", ctx: withID, want: "req_01J9Z6Q4X8", wantOK: true},
		{name: "absent", ctx: context.Background()},
	}
	for _, tt := range tests {
		for shapeName, shape := range shapes {
			t.Run(tt.name+"/"+shapeName, func(t *testing.T) {
				line, _ := logLine(t, func(l *slog.Logger) {
					shape(l).InfoContext(tt.ctx, "m", "n", 1)
				})

				got, ok := line[logging.RequestIDKey]
				assert.Equal(t, tt.wantOK, ok)
				assert.Equal(t, tt.want, got)
			})
		}
	}
}

func TestHandler_groupedAttrsStayGrouped(t *testing.T) {
	ctx := requestid.With(context.Background(), "req_1")

	line, _ := logLine(t, func(l *slog.Logger) {
		l.With("component", "api").WithGroup("g").With("k", "v").InfoContext(ctx, "m", "n", 1)
	})

	assert.Equal(t, "api", line["component"])
	assert.Equal(t, map[string]any{"k": "v", "n": float64(1)}, line["g"])
	assert.Equal(t, "req_1", line[logging.RequestIDKey])
}

func TestHandler_respectsLevel(t *testing.T) {
	var buf bytes.Buffer
	logger := logging.New(&buf, slog.LevelWarn)

	logger.Info("dropped")
	logger.Warn("kept")

	assert.NotContains(t, buf.String(), "dropped")
	assert.Contains(t, buf.String(), "kept")
}

func valueAt(t *testing.T, line map[string]any, path []string) any {
	t.Helper()
	var current any = line
	for _, key := range path {
		obj, ok := current.(map[string]any)
		require.True(t, ok, "no object at %q", key)
		current = obj[key]
	}
	return current
}
