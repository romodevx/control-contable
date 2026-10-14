package params_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/platform/money"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain/taxparams"
)

// The published 2026 progressive tariffs are NOT chained to the centavo: a
// row's fixed fee is not always round(fee + (upper - lower) × rate) of the
// row before it. SAT derives each fee from unrounded values, and Anexo 8
// B.VI publishes the January..November accumulated tables as the monthly
// table × the month number (fees included), so the January slip of at most
// one centavo is multiplied by the month. December (and art. 152, its
// copy) is published on its own and drifts by one centavo at most.
//
// Consequence: crossing some row edges the tax computed from the published
// table goes DOWN by a few centavos (November, edge 5->6: 20,425.34 at the
// row 5 upper bound, 20,425.24 at the row 6 lower bound). The values are DOF
// data and the law applies the published table literally, so the engine
// neither "fixes" the fees nor smooths the edge. These tests pin that
// decision and bound the drift, so a transcription slip (which moves a fee
// by pesos, not centavos) still fails.

// marginalTariff2026 is one published progressive tariff of 2026/v1.
type marginalTariff2026 struct {
	name  string
	key   taxparams.Key
	month time.Month // 0: not a monthly table
	// maxDriftCents bounds |published fee - chained fee| on every edge:
	// one centavo, times the month for an accumulated Jan-Nov table.
	maxDriftCents money.Cents
}

// rifBimonthlyMultiplier: B.VII is published identical to the February
// accumulated table (two monthly tables).
const rifBimonthlyMultiplier = 2

func marginalTariffs2026() []marginalTariff2026 {
	tariffs := []marginalTariff2026{
		{name: "art96 monthly", key: "isr.art96.monthly", maxDriftCents: 1},
		{name: "art116 monthly", key: "isr.art116.monthly", maxDriftCents: 1},
		{name: "rif bimonthly", key: "isr.rif.bimonthly", maxDriftCents: rifBimonthlyMultiplier},
		{name: "art152 annual", key: "isr.art152.annual", maxDriftCents: 1},
	}
	for m := time.January; m <= time.December; m++ {
		drift := money.Cents(m)
		if m == time.December {
			drift = 1 // published on its own, not 12 × January
		}
		tariffs = append(tariffs, marginalTariff2026{
			name: "art106 " + m.String(), key: "isr.art106.accumulated", month: m, maxDriftCents: drift,
		})
	}
	return tariffs
}

func (mt marginalTariff2026) load(t *testing.T, p *taxparams.Params) taxparams.Table {
	t.Helper()
	on := date(t, "2026-12-31")
	var (
		table taxparams.Table
		err   error
	)
	if mt.month == 0 {
		table, err = p.Table(mt.key, on)
	} else {
		table, err = p.MonthlyTable(mt.key, mt.month, on)
	}
	require.NoError(t, err)
	require.Equal(t, taxparams.TableMarginal, table.Kind)
	return table
}

// chainedFee is the fee row r's successor would carry if the published fees
// were chained: the tax at r's upper bound, rounded to the centavo.
func chainedFee(t *testing.T, r taxparams.Row) money.Cents {
	t.Helper()
	fee, err := r.FixedFee.Exact()
	require.NoError(t, err)
	excess, err := (r.Upper - r.Lower).MulRate(r.Rate)
	require.NoError(t, err)
	atUpper, err := fee.Add(excess)
	require.NoError(t, err)
	return atUpper.Round()
}

// TestEmbedded2026MarginalFeesDriftByCentavosOnly checks the continuity
// property on every real 2026 progressive tariff (art. 96, art. 116, RIF,
// the twelve art. 106 tables and art. 152): each published fee is within
// the documented drift of the chained fee, so the tariff is continuous up
// to a few centavos at every edge.
func TestEmbedded2026MarginalFeesDriftByCentavosOnly(t *testing.T) {
	p := embedded2026(t)
	for _, mt := range marginalTariffs2026() {
		t.Run(mt.name, func(t *testing.T) {
			rows := mt.load(t, p).Rows
			for i := range len(rows) - 1 {
				drift := rows[i+1].FixedFee - chainedFee(t, rows[i])
				assert.LessOrEqual(t, max(drift, -drift), mt.maxDriftCents,
					"edge %d->%d: published %s, chained %s", i+1, i+2, rows[i+1].FixedFee, chainedFee(t, rows[i]))
			}
		})
	}
}

// TestEmbedded2026AccumulatedFeesAreTheMonthlyFeesTimesTheMonth pins the
// cause of the growing drift: Jan-Nov accumulated fees are January's × n.
func TestEmbedded2026AccumulatedFeesAreTheMonthlyFeesTimesTheMonth(t *testing.T) {
	p := embedded2026(t)
	on := date(t, "2026-12-31")
	january, err := p.MonthlyTable("isr.art106.accumulated", time.January, on)
	require.NoError(t, err)
	for m := time.February; m <= time.November; m++ {
		table, err := p.MonthlyTable("isr.art106.accumulated", m, on)
		require.NoError(t, err)
		require.Len(t, table.Rows, len(january.Rows))
		for i, r := range table.Rows {
			assert.Equal(t, january.Rows[i].FixedFee*money.Cents(m), r.FixedFee, "%s row %d", m, i+1)
		}
	}
}

// TestLookupAppliesThePublishedFeeLiterally: at a row's lower bound the tax
// is exactly the published fee, even where that is below the tax at the
// previous row's upper bound (November, edge 5->6: 10 centavos less).
func TestLookupAppliesThePublishedFeeLiterally(t *testing.T) {
	november, err := embedded2026(t).MonthlyTable("isr.art106.accumulated", time.November, date(t, "2026-12-31"))
	require.NoError(t, err)
	row5, row6 := november.Rows[4], november.Rows[5]
	require.Equal(t, cents(t, "20,425.24"), row6.FixedFee)

	upper, err := row5.Upper.Exact()
	require.NoError(t, err)
	atUpper, err := november.Tax(upper)
	require.NoError(t, err)
	lower, err := row6.Lower.Exact()
	require.NoError(t, err)
	atLower, err := november.Tax(lower)
	require.NoError(t, err)

	assert.Equal(t, cents(t, "20,425.34"), atUpper.Round())
	assert.Equal(t, cents(t, "20,425.24"), atLower.Round(), "the published fee, not the chained one")
}

// TestEmbedded2026RESICOTariffsNeverDecreaseAtAnEdge: RESICO PF tables are
// flat (the rate applies to the whole income), so they jump UP at each edge
// as long as the rates increase; there is no fee to drift.
func TestEmbedded2026RESICOTariffsNeverDecreaseAtAnEdge(t *testing.T) {
	p := embedded2026(t)
	for _, key := range []taxparams.Key{"isr.resico_pf.monthly", "isr.resico_pf.annual"} {
		t.Run(string(key), func(t *testing.T) {
			table, err := p.Table(key, date(t, "2026-12-31"))
			require.NoError(t, err)
			require.Equal(t, taxparams.TableFlat, table.Kind)
			for i := range len(table.Rows) - 1 {
				upper, err := table.Rows[i].Upper.Exact()
				require.NoError(t, err)
				nextLower, err := table.Rows[i+1].Lower.Exact()
				require.NoError(t, err)
				atUpper, err := table.Tax(upper)
				require.NoError(t, err)
				atNextLower, err := table.Tax(nextLower)
				require.NoError(t, err)
				assert.Zero(t, table.Rows[i].FixedFee)
				assert.Less(t, table.Rows[i].Rate, table.Rows[i+1].Rate)
				assert.LessOrEqual(t, atUpper.Round(), atNextLower.Round(), "edge %d->%d", i+1, i+2)
			}
		})
	}
}
