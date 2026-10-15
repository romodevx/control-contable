package infra_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/expenses/infra"
)

// rangeTaxpayer is any well-formed taxpayer id: the range does not read it.
const rangeTaxpayer = "0199d4a2-1111-7000-8000-0000000000aa"

// The paid_on range of a ledger query, [from, until): January of the first
// year through the end of Through, or from the previous December when the
// tax engine needs January's withheld IVA credit (LIVA 5 fr. IV).
func TestLedgerQueryRange(t *testing.T) {
	tests := []struct {
		name        string
		q           infra.LedgerQuery
		from, until string
	}{
		{"one fiscal year", infra.LedgerQuery{FiscalYear: 2026, Through: time.September}, "2026-01-01", "2026-10-01"},
		{"December reaches the year end", infra.LedgerQuery{FiscalYear: 2026, Through: time.December}, "2026-01-01", "2027-01-01"},
		{"two fiscal years", infra.LedgerQuery{FirstYear: 2026, FiscalYear: 2027, Through: time.February}, "2026-01-01", "2027-03-01"},
		{
			"the previous December", infra.LedgerQuery{FiscalYear: 2027, Through: time.January, PriorDecember: true},
			"2026-12-01", "2027-02-01",
		},
		{
			"the December before the first year",
			infra.LedgerQuery{FirstYear: 2026, FiscalYear: 2027, Through: time.March, PriorDecember: true},
			"2025-12-01", "2027-04-01",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := tt.q
			q.TaxpayerID = rangeTaxpayer

			from, until, err := infra.LedgerRange(q)

			require.NoError(t, err)
			assert.Equal(t, tt.from, from)
			assert.Equal(t, tt.until, until)
		})
	}
}

func TestLedgerQueryRangeRefusesAnInvalidQuery(t *testing.T) {
	for name, q := range map[string]infra.LedgerQuery{
		"month zero":       {TaxpayerID: rangeTaxpayer, FiscalYear: 2026, Through: 0},
		"before the floor": {TaxpayerID: rangeTaxpayer, FiscalYear: 2021, Through: time.January, PriorDecember: true},
		"malformed payer":  {TaxpayerID: "not-a-uuid", FiscalYear: 2026, Through: time.January},
		"first year after": {TaxpayerID: rangeTaxpayer, FirstYear: 2027, FiscalYear: 2026, Through: time.January},
	} {
		t.Run(name, func(t *testing.T) {
			_, _, err := infra.LedgerRange(q)
			require.ErrorIs(t, err, infra.ErrInvalidLedgerQuery)
		})
	}
}
