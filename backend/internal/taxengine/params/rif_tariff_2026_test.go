package params_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
)

// The 621 strategy (M5-17) looks up the bimonthly RIF tariff, Anexo 8 B.VII,
// which the research (A.2.7) and MODULE-tax-engine §2.4 state is identical to
// the February accumulated table of art. 106 (B.VI). Pin that, row by row, so
// a transcription slip in either table shows up as a 621 / 612 divergence.
func TestRIFBimonthlyTariffIsTheFebruaryAccumulatedTable(t *testing.T) {
	p := embedded2026(t)
	for _, raw := range []string{"2026-01-01", "2026-06-30", "2026-12-31"} {
		d, err := fiscal.ParseDate(raw)
		require.NoError(t, err)

		rif, err := p.Table("isr.rif.bimonthly", d)
		require.NoError(t, err, raw)
		february, err := p.MonthlyTable("isr.art106.accumulated", time.February, d)
		require.NoError(t, err, raw)

		assert.Equal(t, february.Kind, rif.Kind, raw)
		assert.Equal(t, february.Rows, rif.Rows, raw)
	}
}
