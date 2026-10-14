package golden_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/platform/money"
	"github.com/romodevx/control-contable/backend/internal/taxengine/golden"
)

const rawParams = `# a comment
fiscal_year: 2026
version: 1
sources: [{id: x}]
values:
  - key: uma.daily
    money: "113.14"
    valid_from: 2026-01-01
    valid_to: 2026-01-31
    basis: [{source: x, article: "a"}]
  - key: uma.daily
    money: "117.31"
    valid_from: "2026-02-01"
    valid_to: 2027-01-31
  - key: iva.rate.general_16
    rate: "0.16"
    valid_from: 2026-01-01
tables:
  - key: isr.open
    kind: marginal
    month: 3
    valid_from: 2026-01-01
    rows:
      - {lower: "0.01", upper: "100.00", fixed_fee: "0.00", rate: "0.0192"}
      - {lower: "100.01", upper: "200.00", fixed_fee: "1.92", rate: "0.064"}
      - {lower: "200.01", fixed_fee: "8.32", rate: "0.35"}
  - key: isr.capped
    kind: flat
    valid_from: 2026-01-01
    valid_to: 2026-12-31
    rows:
      - {lower: "0.01", upper: "100.00", rate: "0.01"}
      - {lower: "150.00", upper: "200.00", rate: "0.02"}
`

func readRaw(t *testing.T) golden.RawParams {
	t.Helper()
	rp, err := golden.ReadRawParams([]byte(rawParams))
	require.NoError(t, err)
	return rp
}

func TestReadRawParamsKeepsTheTextAsWritten(t *testing.T) {
	rp := readRaw(t)

	require.Len(t, rp.Values, 3)
	assert.Equal(t, golden.RawValue{Key: "uma.daily", Money: "113.14", ValidFrom: "2026-01-01", ValidTo: "2026-01-31"}, rp.Values[0])
	assert.Equal(t, golden.RawDate("2026-02-01"), rp.Values[1].ValidFrom, "quoted and plain dates read alike")
	assert.Equal(t, "0.16", rp.Values[2].Rate)
	require.Len(t, rp.Tables, 2)
	assert.Equal(t, 3, rp.Tables[0].Month)
	assert.Equal(t, golden.RawRow{Lower: "200.01", FixedFee: "8.32", Rate: "0.35"}, rp.Tables[0].Rows[2])
}

func TestReadRawParamsRejectsMalformedYAML(t *testing.T) {
	_, err := golden.ReadRawParams([]byte("values: [\n"))

	require.Error(t, err)
}

func amounts(points []golden.TariffPoint) []money.Cents {
	out := make([]money.Cents, 0, len(points))
	for _, p := range points {
		out = append(out, p.Amount)
	}
	return out
}

func TestTariffEdgesListEveryRowEdgeOnce(t *testing.T) {
	rp := readRaw(t)

	open, err := golden.TariffEdges(rp.Tables[0])
	require.NoError(t, err)
	capped, err := golden.TariffEdges(rp.Tables[1])
	require.NoError(t, err)

	// Contiguous rows: upper + 0.01 is the next lower, listed once, as a lower.
	assert.Equal(t, []money.Cents{1, 10000, 10001, 20000, 20001}, amounts(open))
	assert.Equal(t, []string{"row 1 lower", "row 1 upper", "row 2 lower", "row 2 upper", "row 3 lower"}, labels(open))
	for _, p := range open {
		assert.Equal(t, "isr.open", p.Table.Key)
	}
	// A gap and a bounded last row: upper + 0.01 falls outside every row.
	assert.Equal(t, []money.Cents{1, 10000, 10001, 15000, 20000, 20001}, amounts(capped))
	assert.Equal(t, []string{"row 1 lower", "row 1 upper", "row 1 upper+0.01", "row 2 lower", "row 2 upper", "row 2 upper+0.01"}, labels(capped))
}

func labels(points []golden.TariffPoint) []string {
	out := make([]string, 0, len(points))
	for _, p := range points {
		out = append(out, p.Label)
	}
	return out
}

func TestTariffEdgesRejectAnInexactBound(t *testing.T) {
	_, err := golden.TariffEdges(golden.RawTable{Key: "k", Rows: []golden.RawRow{{Lower: "0.001"}}})

	require.Error(t, err)
}

func TestValidityEdgesListTheDayBeforeTheFirstAndLastDayAndTheDayAfter(t *testing.T) {
	rp := readRaw(t)

	points, err := golden.ValidityEdges(rp)
	require.NoError(t, err)

	type got struct {
		key   string
		month int
		table bool
		date  string
	}
	var all []got
	for _, p := range points {
		all = append(all, got{p.Key, p.Month, p.Table, p.Date.String()})
	}
	assert.Equal(t, []got{
		{"iva.rate.general_16", 0, false, "2025-12-31"},
		{"iva.rate.general_16", 0, false, "2026-01-01"},
		{"uma.daily", 0, false, "2025-12-31"},
		{"uma.daily", 0, false, "2026-01-01"},
		{"uma.daily", 0, false, "2026-01-31"},
		{"uma.daily", 0, false, "2026-02-01"},
		{"uma.daily", 0, false, "2027-01-31"},
		{"uma.daily", 0, false, "2027-02-01"},
		{"isr.capped", 0, true, "2025-12-31"},
		{"isr.capped", 0, true, "2026-01-01"},
		{"isr.capped", 0, true, "2026-12-31"},
		{"isr.capped", 0, true, "2027-01-01"},
		{"isr.open", 3, true, "2025-12-31"},
		{"isr.open", 3, true, "2026-01-01"},
	}, all)
}

func TestValidityEdgesRejectABadDate(t *testing.T) {
	_, err := golden.ValidityEdges(golden.RawParams{Values: []golden.RawValue{{Key: "k", ValidFrom: "2026-02-30"}}})

	require.Error(t, err)
}
