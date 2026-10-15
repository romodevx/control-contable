package explain

import (
	"fmt"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain/annual"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain/taxparams"
)

// Error is an explain error; its values are constants matched with
// errors.Is.
type Error string

func (e Error) Error() string { return string(e) }

// The explain errors.
const (
	// ErrNoDeadline reports a deadline kind that has no date.
	ErrNoDeadline Error = "annual explain: no deadline for this kind"
	// ErrNegativeIncome reports a negative total income.
	ErrNegativeIncome Error = "annual explain: negative total income"
	// ErrUnknownCode reports a disclaimer code or CPA topic outside the
	// spec, or cpa_review_pending raised without its topics.
	ErrUnknownCode Error = "annual explain: unknown disclaimer code or CPA topic"
)

// Params is what the package reads from the parameter set;
// *taxparams.Params implements it.
type Params interface {
	Date(key taxparams.Key, on fiscal.Date) (fiscal.Date, error)
	Money(key taxparams.Key, on fiscal.Date) (money.Cents, error)
}

// DeadlineKind names which annual deadline a section shows.
type DeadlineKind string

// The deadline kinds (MODULE-annual §2.9).
const (
	DeadlineNone                 DeadlineKind = ""
	DeadlinePF                   DeadlineKind = "pf"
	DeadlinePM                   DeadlineKind = "pm"
	DeadlineNonprofitInformative DeadlineKind = "nonprofit_informative"
)

// deadlineKeys are the parameter of each kind: PF in April (LISR 150), PM
// within three months (LISR 76 fr. V), 603 informative by 15 February
// (LISR 86 fr. III); the SAT can extend them, so they live in params.
var deadlineKeys = map[DeadlineKind]taxparams.Key{
	DeadlinePF:                   taxparams.KeyDeadlinePF,
	DeadlinePM:                   taxparams.KeyDeadlinePM,
	DeadlineNonprofitInformative: taxparams.KeyDeadlineNonprofitInformative,
}

// sectionDeadlines are the deadline of each computed section kind;
// pf_resico_check shows the PF date "if filed" (research §A.1).
var sectionDeadlines = map[annual.SectionKind]DeadlineKind{
	annual.SectionPFArt152:      DeadlinePF,
	annual.SectionPFRESICOCheck: DeadlinePF,
	annual.SectionPM601:         DeadlinePM,
	annual.SectionPMRESICO:      DeadlinePM,
}

// SectionDeadline is the deadline a computed section shows; an explanation
// section carries its own (Section.Deadline).
func SectionDeadline(kind annual.SectionKind) DeadlineKind { return sectionDeadlines[kind] }

// Deadline is the legal date of kind for fiscalYear's return, read in force
// on the year's last day. The use case moves it to the next business day
// with the M5 calendar (CFF 12 ⚠️, "sujeto a confirmación"); the RFC-digit
// grace does not apply to the annual (research §A.2).
func Deadline(p Params, fiscalYear int, kind DeadlineKind) (fiscal.Date, error) {
	key, ok := deadlineKeys[kind]
	if !ok {
		return fiscal.Date{}, fmt.Errorf("deadline %q: %w", kind, ErrNoDeadline)
	}
	yearEnd, err := yearEnd(fiscalYear)
	if err != nil {
		return fiscal.Date{}, err
	}
	d, err := p.Date(key, yearEnd)
	if err != nil {
		return fiscal.Date{}, fmt.Errorf("annual explain: deadline %s: %w", kind, err)
	}
	return d, nil
}

// InformativeIncome reports whether a persona física's total income of the
// year, exempt and definitively taxed income included, is above the LISR
// 150 ¶3 threshold ($500,000 in 2026): the annual must then also report
// loans, gifts, inheritances and prizes (disclaimer informative_income).
func InformativeIncome(p Params, fiscalYear int, totalIncome money.Cents) (bool, error) {
	if totalIncome < 0 {
		return false, fmt.Errorf("total income %s: %w", totalIncome, ErrNegativeIncome)
	}
	yearEnd, err := yearEnd(fiscalYear)
	if err != nil {
		return false, err
	}
	threshold, err := p.Money(taxparams.KeyInformativeIncomeThreshold, yearEnd)
	if err != nil {
		return false, fmt.Errorf("annual explain: informative income threshold: %w", err)
	}
	return totalIncome > threshold, nil
}

func yearEnd(fiscalYear int) (fiscal.Date, error) {
	d, err := fiscal.ParseDate(fmt.Sprintf("%04d-12-31", fiscalYear))
	if err != nil {
		return fiscal.Date{}, fmt.Errorf("annual explain: fiscal year %d: %w", fiscalYear, err)
	}
	return d, nil
}
