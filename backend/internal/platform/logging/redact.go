package logging

import (
	"fmt"
	"log/slog"
	"regexp"
	"slices"
	"strings"
)

// Redacted replaces every sensitive key's value and every sensitive substring.
const Redacted = "[REDACTED]"

// sensitiveKeyParts match a normalized key (lower case, no '_', '-' or '.')
// by substring, so "user_email", "X-Refresh-Token" and "rfcHash" all match.
var sensitiveKeyParts = []string{
	"rfc", "email", "password", "passwd", "secret", "token", "authorization",
	"cookie", "apikey", "curp", "pepper", "privatekey", "credential",
}

// sensitiveValue matches, case-insensitively and leftmost first: a bearer
// credential, a JWT, a refresh token (<uuid>.<secret>), an email address and
// a Mexican RFC (3 letters for a person moral, 4 for a person físico, a
// YYMMDD date and a 3-character homoclave).
var sensitiveValue = regexp.MustCompile(`(?i)` + strings.Join([]string{
	`bearer\s+\S+`,
	`eyJ[a-z0-9_-]+\.[a-z0-9_-]+\.[a-z0-9_-]*`,
	`[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}\.[a-z0-9_-]{16,}`,
	`[a-z0-9._%+-]+@[a-z0-9-]+(?:\.[a-z0-9-]+)+`,
	`[a-zñ&]{3,4}[0-9]{2}(?:0[1-9]|1[0-2])(?:0[1-9]|[12][0-9]|3[01])[a-z0-9]{3}`,
}, "|"))

// redactAttr is the JSON handler's ReplaceAttr. slog calls it for every
// non-group attribute, including the message, with the enclosing group keys.
func redactAttr(groups []string, a slog.Attr) slog.Attr {
	if a.Equal(slog.Attr{}) || (len(groups) == 0 && isBuiltinMetadata(a.Key)) {
		return a // the handler drops empty attrs; built-in metadata holds no user data
	}
	if isSensitiveKey(a.Key) || slices.ContainsFunc(groups, isSensitiveKey) {
		return slog.String(a.Key, Redacted)
	}
	switch a.Value.Kind() {
	case slog.KindString:
		return slog.String(a.Key, scrub(a.Value.String()))
	case slog.KindAny:
		return slog.String(a.Key, scrub(fmt.Sprint(a.Value.Any())))
	default:
		return a
	}
}

// isBuiltinMetadata reports the handler's own keys that never hold user data.
// The message key is not one of them: it is scrubbed like any string.
func isBuiltinMetadata(key string) bool {
	return key == slog.TimeKey || key == slog.LevelKey || key == slog.SourceKey
}

func isSensitiveKey(key string) bool {
	normalized := strings.NewReplacer("_", "", "-", "", ".", "").Replace(strings.ToLower(key))
	return slices.ContainsFunc(sensitiveKeyParts, func(part string) bool {
		return strings.Contains(normalized, part)
	})
}

func scrub(s string) string {
	return sensitiveValue.ReplaceAllLiteralString(s, Redacted)
}
