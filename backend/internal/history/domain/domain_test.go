package domain_test

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/history/domain"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
)

func month(t *testing.T, raw string) domain.Month {
	t.Helper()
	m, err := domain.ParseMonth(raw)
	require.NoError(t, err)
	return m
}

func strings64(c string) string { return strings.Repeat(c, 64) }

func TestParseMonth(t *testing.T) {
	m := month(t, "2026-03")
	assert.Equal(t, 2026, m.Year())
	assert.Equal(t, time.March, m.Month())
	assert.Equal(t, "2026-03", m.String())
	for _, bad := range []string{"", "2026-3", "2026-13", "26-03", "+026-03", "2026-03-01", "abcd-ef"} {
		_, err := domain.ParseMonth(bad)
		assert.ErrorIs(t, err, domain.ErrInvalidMonth, bad)
	}
}

func TestMonthOrder(t *testing.T) {
	assert.True(t, month(t, "2026-12").Before(month(t, "2027-01")))
	assert.False(t, month(t, "2026-03").Before(month(t, "2026-03")))
	assert.Equal(t, month(t, "2026-10"), domain.MonthOf(time.Date(2026, time.October, 31, 23, 0, 0, 0, time.UTC)))
}

// §2.1: January .. December for a past year, January .. the current month
// for the current year; newest first.
func TestYearMonths(t *testing.T) {
	current := month(t, "2027-03")

	past := domain.YearMonths(2026, current)
	require.Len(t, past, 12)
	assert.Equal(t, "2026-12", past[0].String(), "newest first")
	assert.Equal(t, "2026-01", past[11].String())

	this := domain.YearMonths(2027, current)
	assert.Equal(t, []domain.Month{month(t, "2027-03"), month(t, "2027-02"), month(t, "2027-01")}, this)

	january := domain.YearMonths(2027, month(t, "2027-01"))
	assert.Equal(t, []domain.Month{month(t, "2027-01")}, january)
}

func TestTaxpayerStatus(t *testing.T) {
	tests := []struct {
		name    string
		regimes []domain.Status
		want    domain.Status
	}{
		{"no regime", nil, domain.StatusNone},
		{"all informational", []domain.Status{domain.StatusInformational, domain.StatusInformational}, domain.StatusInformational},
		{"incomplete wins", []domain.Status{domain.StatusPartial, domain.StatusIncomplete, domain.StatusInformational}, domain.StatusIncomplete},
		{"partial", []domain.Status{domain.StatusComplete, domain.StatusPartial}, domain.StatusPartial},
		{"complete", []domain.Status{domain.StatusComplete, domain.StatusInformational}, domain.StatusComplete},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, domain.TaxpayerStatus(tt.regimes))
		})
	}
}

func TestParameterSetLabel(t *testing.T) {
	assert.Equal(t, "2026/v1", domain.ParameterSet{FiscalYear: 2026, Version: 1}.String())
}

// §2.4: the marker is driven only by the figures fingerprint; the reason
// is engine when the engine version or parameter set moved, else records.
func TestChange(t *testing.T) {
	seen := domain.Seen{
		RecordsFingerprint: strings64("a"), FiguresFingerprint: strings64("b"),
		EngineVersion: "1.0.1", ParameterSet: "2026/v1",
	}
	view := &domain.View{Seen: seen, ViewedAt: time.Date(2026, time.October, 3, 12, 0, 0, 0, time.UTC)}
	with := func(mutate func(*domain.Seen)) domain.Seen {
		s := seen
		mutate(&s)
		return s
	}
	tests := []struct {
		name      string
		view      *domain.View
		current   domain.Seen
		isCurrent bool
		want      domain.ChangeReason
	}{
		{"never viewed (OQ-H1)", nil, with(func(s *domain.Seen) { s.FiguresFingerprint = strings64("c") }), false, ""},
		{"same figures", view, seen, false, ""},
		{"same figures, other inputs and engine (OQ-H4)", view, with(func(s *domain.Seen) {
			s.RecordsFingerprint, s.EngineVersion, s.ParameterSet = strings64("d"), "1.0.2", "2026/v2"
		}), false, ""},
		{"figures moved with the records", view, with(func(s *domain.Seen) {
			s.RecordsFingerprint, s.FiguresFingerprint = strings64("d"), strings64("c")
		}), false, domain.ChangeRecords},
		{"figures moved with the engine", view, with(func(s *domain.Seen) {
			s.FiguresFingerprint, s.EngineVersion = strings64("c"), "1.0.2"
		}), false, domain.ChangeEngine},
		{"figures moved with the parameters", view, with(func(s *domain.Seen) {
			s.FiguresFingerprint, s.ParameterSet = strings64("c"), "2026/v2"
		}), false, domain.ChangeEngine},
		{"both causes: engine wins", view, with(func(s *domain.Seen) {
			s.RecordsFingerprint, s.FiguresFingerprint, s.EngineVersion = strings64("d"), strings64("c"), "1.0.2"
		}), false, domain.ChangeEngine},
		{"the current month is never marked (OQ-H3)", view, with(func(s *domain.Seen) { s.FiguresFingerprint = strings64("c") }), true, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, domain.Change(tt.view, tt.current, tt.isCurrent))
		})
	}
}

func TestSeenValidate(t *testing.T) {
	valid := domain.Seen{RecordsFingerprint: strings64("a"), FiguresFingerprint: strings64("0"), EngineVersion: "1.0.1", ParameterSet: "2026/v1"}
	require.NoError(t, valid.Validate())
	tests := []struct {
		name   string
		mutate func(*domain.Seen)
		field  string
	}{
		{"upper-case records fingerprint", func(s *domain.Seen) { s.RecordsFingerprint = strings64("A") }, "records_fingerprint"},
		{"short figures fingerprint", func(s *domain.Seen) { s.FiguresFingerprint = "abc" }, "figures_fingerprint"},
		{"empty engine version", func(s *domain.Seen) { s.EngineVersion = "" }, "engine_version"},
		{"long engine version", func(s *domain.Seen) { s.EngineVersion = strings.Repeat("1", 33) }, "engine_version"},
		{"engine version with a space", func(s *domain.Seen) { s.EngineVersion = "1.0 1" }, "engine_version"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := valid
			tt.mutate(&s)
			err := s.Validate()
			var invalid *domain.InvalidSeenError
			require.ErrorAs(t, err, &invalid)
			assert.Equal(t, tt.field, invalid.Field)
		})
	}
}

func TestNewParameterSet(t *testing.T) {
	_, err := domain.NewParameterSet(2026, 1)
	require.NoError(t, err)
	for _, bad := range [][2]int{{2021, 1}, {2101, 1}, {2026, 0}, {2026, 10000}} {
		_, err := domain.NewParameterSet(bad[0], bad[1])
		var invalid *domain.InvalidSeenError
		require.ErrorAs(t, err, &invalid, "%v", bad)
		assert.Equal(t, "parameter_set", invalid.Field)
	}
}

// §2.2: the sum of the months' figures, a null month adding nothing;
// overflow fails rather than wraps.
func TestYearTotals(t *testing.T) {
	var totals domain.YearTotals
	require.NoError(t, totals.Add(domain.Tax{ISRPayable: 100, IVAPayable: 200, WithholdingsToRemit: 5, TotalPayable: 305, IVAInFavor: 0}))
	require.NoError(t, totals.Add(domain.Tax{ISRPayable: 1, IVAPayable: 0, IVAInFavor: 900, TotalPayable: 1}))
	assert.Equal(t, domain.YearTotals{ISRPayable: 101, IVAPayable: 200, WithholdingsToRemit: 5, TotalPayable: 306}, totals,
		"a saldo a favor never subtracts")

	huge := domain.YearTotals{TotalPayable: money.Cents(1<<63 - 1)}
	assert.ErrorIs(t, huge.Add(domain.Tax{TotalPayable: 1}), money.ErrOverflow)
}

// §2.2: the union of the months' disclaimers, each code once (the newest
// month's text), not_tax_advice first, then strong, warning, info, then code.
func TestUnionDisclaimers(t *testing.T) {
	notAdvice := domain.Disclaimer{Code: "not_tax_advice", Severity: domain.SeverityInfo, Text: "n"}
	cpaNew := domain.Disclaimer{Code: "cpa_review_pending", Severity: domain.SeverityStrong, Text: "october"}
	cpaOld := domain.Disclaimer{Code: "cpa_review_pending", Severity: domain.SeverityStrong, Text: "march"}
	partial := domain.Disclaimer{Code: "partial_regime", Severity: domain.SeverityWarning, Text: "p"}
	deductible := domain.Disclaimer{Code: "deductibility_claimed", Severity: domain.SeverityInfo, Text: "d"}
	informational := domain.Disclaimer{Code: "informational_regime", Severity: domain.SeverityInfo, Text: "i"}

	got := domain.UnionDisclaimers([][]domain.Disclaimer{
		{notAdvice, cpaNew, informational},
		{notAdvice, cpaOld, partial, deductible},
	})

	assert.Equal(t, []domain.Disclaimer{notAdvice, cpaNew, partial, deductible, informational}, got)
}
