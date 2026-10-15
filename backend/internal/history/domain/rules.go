package domain

import (
	"cmp"
	"fmt"
	"regexp"
	"slices"
	"time"

	"github.com/romodevx/control-contable/backend/internal/platform/money"
)

// FirstFiscalYear is the first fiscal year with published tax parameters:
// an earlier year is 422 parameters_unavailable (§2.1).
const FirstFiscalYear = 2026

// Status is the taxpayer-level status of one month's estimate (contract
// TaxpayerEstimateStatus); a regime's status uses the same values but
// never StatusNone.
type Status string

// The statuses.
const (
	StatusComplete      Status = "complete"
	StatusIncomplete    Status = "incomplete"
	StatusPartial       Status = "partial"
	StatusInformational Status = "informational"
	StatusNone          Status = "none"
)

// TaxpayerStatus derives the month's status from its regimes' (M5
// precedence, §2.2): none without a regime, informational when every
// regime is, else incomplete if any is, partial if any is, and complete.
func TaxpayerStatus(regimes []Status) Status {
	switch {
	case len(regimes) == 0:
		return StatusNone
	case !slices.ContainsFunc(regimes, func(s Status) bool { return s != StatusInformational }):
		return StatusInformational
	case slices.Contains(regimes, StatusIncomplete):
		return StatusIncomplete
	case slices.Contains(regimes, StatusPartial):
		return StatusPartial
	}
	return StatusComplete
}

// Bounds of what a view may name (contract HistoryViewedRequest and the
// estimate_views CHECKs).
const (
	minParamsYear      = 2022
	maxParamsYear      = 2100
	minParamsVersion   = 1
	maxParamsVersion   = 9999
	maxEngineVersion   = 32
	fieldRecordsFP     = "records_fingerprint"
	fieldFiguresFP     = "figures_fingerprint"
	fieldEngineVersion = "engine_version"
	fieldParameterSet  = "parameter_set"
)

var (
	fingerprintPattern   = regexp.MustCompile(`^[0-9a-f]{64}$`)
	engineVersionPattern = regexp.MustCompile(`^[0-9A-Za-z./-]+$`)
)

// InvalidSeenError reports the field of a Seen or ParameterSet that is
// out of format (400 validation_failed on that field).
type InvalidSeenError struct {
	Field string
}

func (e *InvalidSeenError) Error() string { return "history: invalid " + e.Field }

// ParameterSet names a tax parameter set as the estimate reports it.
type ParameterSet struct {
	FiscalYear int
	Version    int
}

// NewParameterSet checks the bounds of a parameter set a client sent.
func NewParameterSet(fiscalYear, version int) (ParameterSet, error) {
	if fiscalYear < minParamsYear || fiscalYear > maxParamsYear || version < minParamsVersion || version > maxParamsVersion {
		return ParameterSet{}, &InvalidSeenError{Field: fieldParameterSet}
	}
	return ParameterSet{FiscalYear: fiscalYear, Version: version}, nil
}

// String is the storage form, "2026/v1" (§2.3).
func (p ParameterSet) String() string { return fmt.Sprintf("%d/v%d", p.FiscalYear, p.Version) }

// Seen is what an estimate of a month shows the user, as History compares
// it: the two fingerprints, the engine version and the parameter set's
// storage form.
type Seen struct {
	RecordsFingerprint string
	FiguresFingerprint string
	EngineVersion      string
	ParameterSet       string
}

// Validate checks the formats a client-sent Seen must have (§2.5): 64
// lower-case hex fingerprints and a version label of at most 32
// characters of [0-9A-Za-z./-]. The parameter set is checked when built
// (NewParameterSet).
func (s Seen) Validate() error {
	switch {
	case !fingerprintPattern.MatchString(s.RecordsFingerprint):
		return &InvalidSeenError{Field: fieldRecordsFP}
	case !fingerprintPattern.MatchString(s.FiguresFingerprint):
		return &InvalidSeenError{Field: fieldFiguresFP}
	case len(s.EngineVersion) > maxEngineVersion || !engineVersionPattern.MatchString(s.EngineVersion):
		return &InvalidSeenError{Field: fieldEngineVersion}
	}
	return nil
}

// View is the caller's last view of one month.
type View struct {
	Month    Month
	Seen     Seen
	ViewedAt time.Time
}

// ChangeReason is why a viewed month's figures changed (contract
// HistoryChangeReason); "" is unchanged.
type ChangeReason string

// The reasons.
const (
	ChangeRecords ChangeReason = "records"
	ChangeEngine  ChangeReason = "engine"
)

// Change applies §2.4: a month never viewed, or the current month, is never
// changed; a viewed month changed only when its figures fingerprint moved
// (OQ-H4), because of the engine when its version or the parameter set
// differ from the viewed ones (engine wins when both causes apply), else
// because of the records.
func Change(view *View, current Seen, isCurrent bool) ChangeReason {
	switch {
	case view == nil || isCurrent || view.Seen.FiguresFingerprint == current.FiguresFingerprint:
		return ""
	case view.Seen.EngineVersion != current.EngineVersion || view.Seen.ParameterSet != current.ParameterSet:
		return ChangeEngine
	}
	return ChangeRecords
}

// Tax is one month's taxpayer-level figures (contract HistoryTax).
type Tax struct {
	ISRPayable          money.Cents
	IVAPayable          money.Cents
	IVAOutcome          string
	IVAInFavor          money.Cents
	WithholdingsToRemit money.Cents
	TotalPayable        money.Cents
}

// YearTotals is "Suma de estimados mensuales" (contract HistoryYearTotals):
// a sum of provisional estimates, never the annual return.
type YearTotals struct {
	ISRPayable          money.Cents
	IVAPayable          money.Cents
	WithholdingsToRemit money.Cents
	TotalPayable        money.Cents
}

// Add adds a month's payable figures; a saldo a favor never subtracts.
// It fails with money.ErrOverflow and then leaves t unchanged.
func (t *YearTotals) Add(m Tax) error {
	isr, errISR := t.ISRPayable.Add(m.ISRPayable)
	iva, errIVA := t.IVAPayable.Add(m.IVAPayable)
	remit, errRemit := t.WithholdingsToRemit.Add(m.WithholdingsToRemit)
	total, errTotal := t.TotalPayable.Add(m.TotalPayable)
	for _, err := range []error{errISR, errIVA, errRemit, errTotal} {
		if err != nil {
			return fmt.Errorf("history: year totals: %w", err)
		}
	}
	*t = YearTotals{ISRPayable: isr, IVAPayable: iva, WithholdingsToRemit: remit, TotalPayable: total}
	return nil
}

// Severity is a disclaimer's severity (contract DisclaimerSeverity).
type Severity string

// The severities.
const (
	SeverityInfo    Severity = "info"
	SeverityWarning Severity = "warning"
	SeverityStrong  Severity = "strong"
)

// severityRank orders severities strong, warning, info (M5 §2.3).
var severityRank = map[Severity]int{SeverityStrong: 0, SeverityWarning: 1, SeverityInfo: 2}

// codeNotTaxAdvice is the disclaimer every estimate opens with.
const codeNotTaxAdvice = "not_tax_advice"

// Disclaimer is one estimate disclaimer as the engine renders it.
type Disclaimer struct {
	Code     string
	Severity Severity
	Text     string
}

// UnionDisclaimers is the union of the months' disclaimers (§2.2), each
// code once with the text of the first list that has it (months come
// newest first), in the M5 order: not_tax_advice first, then by severity
// (strong, warning, info), then by code.
func UnionDisclaimers(months [][]Disclaimer) []Disclaimer {
	seen := map[string]bool{}
	var out []Disclaimer
	for _, list := range months {
		for _, d := range list {
			if !seen[d.Code] {
				seen[d.Code] = true
				out = append(out, d)
			}
		}
	}
	slices.SortFunc(out, func(a, b Disclaimer) int {
		return cmp.Or(
			cmp.Compare(rankNotAdvice(a), rankNotAdvice(b)),
			cmp.Compare(severityRank[a.Severity], severityRank[b.Severity]),
			cmp.Compare(a.Code, b.Code),
		)
	})
	return out
}

func rankNotAdvice(d Disclaimer) int {
	if d.Code == codeNotTaxAdvice {
		return 0
	}
	return 1
}
