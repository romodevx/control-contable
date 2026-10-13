package logging_test

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"testing"
	"testing/slogtest"

	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/platform/logging"
)

// TestHandler_conformance runs the standard slog.Handler contract suite
// against the wrapped, redacting handler.
func TestHandler_conformance(t *testing.T) {
	var buf bytes.Buffer
	newHandler := func(*testing.T) slog.Handler {
		buf.Reset()
		return logging.NewHandler(&buf, slog.LevelDebug)
	}
	result := func(t *testing.T) map[string]any {
		line := map[string]any{}
		if buf.Len() > 0 {
			require.NoError(t, json.Unmarshal(buf.Bytes(), &line))
		}
		return line
	}

	slogtest.Run(t, newHandler, result)
}
