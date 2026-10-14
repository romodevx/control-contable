package golden_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.yaml.in/yaml/v3"

	"github.com/romodevx/control-contable/backend/internal/platform/money"
	"github.com/romodevx/control-contable/backend/internal/taxengine/golden"
)

// The helpers the annual harness shares with this one (PLAN ANN-06).

func scalar(t *testing.T, doc string) *yaml.Node {
	t.Helper()
	var n yaml.Node
	require.NoError(t, yaml.Unmarshal([]byte(doc), &n))
	return n.Content[0]
}

func TestDecodeCentsTakesOnlyQuotedCentavoStrings(t *testing.T) {
	tests := []struct {
		doc     string
		want    money.Cents
		wantErr bool
	}{
		{doc: `"150.96"`, want: 15096},
		{doc: `"-1320.00"`, want: -132000},
		{doc: `150.96`, wantErr: true},
		{doc: `"150.9"`, wantErr: true},
		{doc: `"150"`, wantErr: true},
		{doc: `[1]`, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.doc, func(t *testing.T) {
			got, err := golden.DecodeCents(scalar(t, tt.doc))

			if tt.wantErr {
				assert.ErrorIs(t, err, golden.ErrInexactAmount)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestDecodeRateTakesOnlyQuotedDecimals(t *testing.T) {
	got, err := golden.DecodeRate(scalar(t, `"0.1380"`))
	require.NoError(t, err)
	assert.Equal(t, money.Rate(138000), got)

	_, err = golden.DecodeRate(scalar(t, `0.1380`))
	assert.ErrorIs(t, err, golden.ErrInexactAmount)
}

func TestBaselineFindingsTakePlainCounts(t *testing.T) {
	b := golden.Baseline{Skipped: []string{"AN-04"}, Pending: map[string]int{"pf_art152": 2}}

	assert.Empty(t, b.Findings(map[string]int{"pf_art152": 2}, []string{"AN-04"}))
	assert.Equal(t, []string{
		"pending_strategy pf_art152: 3 cases, baseline 2: register the strategy instead of adding pending cases",
		"skipped AN-04 no longer skips: remove it from the baseline",
	}, b.Findings(map[string]int{"pf_art152": 3}, nil))
}
