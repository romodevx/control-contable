package main

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseCommand(t *testing.T) {
	tests := []struct {
		args    []string
		want    command
		wantErr bool
	}{
		{args: []string{"up"}, want: cmdUp},
		{args: []string{"down"}, want: cmdDown},
		{args: []string{"status"}, want: cmdStatus},
		{args: nil, wantErr: true},
		{args: []string{"reset"}, wantErr: true},
		{args: []string{"up", "extra"}, wantErr: true},
	}
	for _, tc := range tests {
		t.Run(fmtArgs(tc.args), func(t *testing.T) {
			got, err := parseCommand(tc.args)
			if tc.wantErr {
				require.ErrorIs(t, err, errUsage)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestRun_RequiresTheOwnerDatabaseURL(t *testing.T) {
	noEnv := func(string) (string, bool) { return "", false }

	err := run(t.Context(), []string{"status"}, noEnv, &bytes.Buffer{})

	require.Error(t, err)
	assert.Contains(t, err.Error(), envDatabaseURL)
}

func TestRun_RejectsUsageBeforeTouchingTheDatabase(t *testing.T) {
	env := func(string) (string, bool) { return "postgres://unreachable.invalid/db", true }

	err := run(t.Context(), []string{"sideways"}, env, &bytes.Buffer{})

	require.ErrorIs(t, err, errUsage)
}

func fmtArgs(args []string) string {
	if len(args) == 0 {
		return "no-args"
	}
	var b bytes.Buffer
	for i, a := range args {
		if i > 0 {
			b.WriteByte('_')
		}
		b.WriteString(a)
	}
	return b.String()
}
