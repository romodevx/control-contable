package domain

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	// ErrInvalidPeriod reports a monthly period outside years 1-9999 or
	// months 1-12.
	ErrInvalidPeriod = errors.New("reminders: invalid period")
	// ErrInvalidKind reports a reminder kind other than early or due.
	ErrInvalidKind = errors.New("reminders: invalid reminder kind")
	// ErrNothingToRemind reports a message with no taxpayer to list: the
	// job records the reminder as skipped instead (§2.4 step 6).
	ErrNothingToRemind = errors.New("reminders: no taxpayer to remind")
)

const (
	minYear = 1
	maxYear = 9999
)

// Period is a monthly period, the month a reminder is about (never its due
// date, §2.4 step 4).
type Period struct {
	year  int
	month time.Month
}

// NewPeriod validates year and month.
func NewPeriod(year int, month time.Month) (Period, error) {
	if year < minYear || year > maxYear || month < time.January || month > time.December {
		return Period{}, fmt.Errorf("%w: %d-%d", ErrInvalidPeriod, year, month)
	}
	return Period{year: year, month: month}, nil
}

// PeriodOf is the period of the calendar date d.
func PeriodOf(d time.Time) (Period, error) { return NewPeriod(d.Year(), d.Month()) }

// Year is the period's year.
func (p Period) Year() int { return p.year }

// Month is the period's month.
func (p Period) Month() time.Month { return p.month }

// IsZero reports the unset period.
func (p Period) IsZero() bool { return p.year == 0 }

// String is "YYYY-MM".
func (p Period) String() string { return fmt.Sprintf("%04d-%02d", p.year, int(p.month)) }

// Kind is which of a period's two reminders (§2.2): the early one or the
// one on the due date.
type Kind string

// The reminder kinds (notification_log.kind).
const (
	KindEarly Kind = "early"
	KindDue   Kind = "due"
)

// ParseKind parses notification_log.kind.
func ParseKind(raw string) (Kind, error) {
	switch k := Kind(raw); k {
	case KindEarly, KindDue:
		return k, nil
	default:
		return "", fmt.Errorf("%w: %q", ErrInvalidKind, raw)
	}
}

// EstimateShown is what a notification with amounts says about one
// taxpayer (§2.5).
type EstimateShown int

const (
	// EstimateNameOnly lists the taxpayer by name only (amounts off, the
	// engine OFF, no figure to show, or the estimator failed).
	EstimateNameOnly EstimateShown = iota
	// EstimateAmount shows the estimate's rounded total.
	EstimateAmount
	// EstimatePending says "estimado pendiente": the estimate is incomplete.
	EstimatePending
)

// Listed is one taxpayer as a notification lists it: its display name and,
// with amounts, its estimate. Never an RFC, an id or a regime.
type Listed struct {
	Name        string
	Estimate    EstimateShown
	AmountCents int64
}

// MessageInput is what a reminder says.
type MessageInput struct {
	Kind   Kind
	Period Period
	// DueDate is the (moved) due date, a calendar day (the time of day and
	// the location are ignored).
	DueDate time.Time
	// Taxpayers are in the switcher's order.
	Taxpayers []Listed
}

// Message is a reminder's lock-screen text and deep link (§2.5).
type Message struct {
	Title    string
	Body     string
	DeepLink string
}

// maxListed is how many taxpayers a notification names before "y N más".
const maxListed = 3

const (
	titleDue      = "Hoy vence tu pago mensual estimado"
	titleEarlyFmt = "Tu pago mensual estimado vence el %s"
	bodyNamesFmt  = "Revisa tu estimado de %s para: %s."
	amountsSuffix = ". Cifras aproximadas."
	deepLinkFmt   = "controlcontable://estimate?period=%s"
	pendingText   = "estimado pendiente"
	amountFmt     = "%s: %s estimado"
)

var (
	weekdaysES = [...]string{"domingo", "lunes", "martes", "miércoles", "jueves", "viernes", "sábado"}
	monthsES   = [...]string{"enero", "febrero", "marzo", "abril", "mayo", "junio", "julio", "agosto",
		"septiembre", "octubre", "noviembre", "diciembre"}
)

// BuildMessage writes the reminder in Spanish, always "estimado" (§2.5):
// up to three taxpayers, then "y N más"; with at least one amount among the
// named ones, the amounts body ("Cifras aproximadas"), else the names body.
func BuildMessage(in MessageInput) (Message, error) {
	if len(in.Taxpayers) == 0 {
		return Message{}, ErrNothingToRemind
	}
	title, err := title(in.Kind, in.DueDate)
	if err != nil {
		return Message{}, err
	}
	return Message{Title: title, Body: body(in), DeepLink: fmt.Sprintf(deepLinkFmt, in.Period)}, nil
}

func title(kind Kind, due time.Time) (string, error) {
	switch kind {
	case KindDue:
		return titleDue, nil
	case KindEarly:
		return fmt.Sprintf(titleEarlyFmt, spanishDay(due)), nil
	default:
		return "", fmt.Errorf("%w: %q", ErrInvalidKind, kind)
	}
}

func body(in MessageInput) string {
	named := in.Taxpayers[:min(len(in.Taxpayers), maxListed)]
	rest := len(in.Taxpayers) - len(named)
	if !anyAmount(named) {
		names := make([]string, len(named))
		for i, t := range named {
			names[i] = t.Name
		}
		return fmt.Sprintf(bodyNamesFmt, monthsES[in.Period.month-1], withRest(strings.Join(names, ", "), rest))
	}
	items := make([]string, len(named))
	for i, t := range named {
		items[i] = amountItem(t)
	}
	return withRest(strings.Join(items, " · "), rest) + amountsSuffix
}

func anyAmount(ts []Listed) bool {
	for _, t := range ts {
		if t.Estimate == EstimateAmount {
			return true
		}
	}
	return false
}

func amountItem(t Listed) string {
	switch t.Estimate {
	case EstimateAmount:
		return fmt.Sprintf(amountFmt, t.Name, FormatMXN(t.AmountCents))
	case EstimatePending:
		return t.Name + ": " + pendingText
	default:
		return t.Name
	}
}

func withRest(list string, rest int) string {
	if rest == 0 {
		return list
	}
	return fmt.Sprintf("%s y %d más", list, rest)
}

// spanishDay is "lunes 18 de mayo".
func spanishDay(d time.Time) string {
	return fmt.Sprintf("%s %d de %s", weekdaysES[d.Weekday()], d.Day(), monthsES[d.Month()-1])
}

const (
	centsPerPeso   = 100
	thousandsGroup = 3
)

// FormatMXN formats cents es-MX: "$4,520" for whole pesos, "$4,520.50"
// otherwise; a negative amount keeps its sign ("-$10").
func FormatMXN(cents int64) string {
	sign := ""
	if cents < 0 {
		sign, cents = "-", -cents
	}
	pesos, rest := cents/centsPerPeso, cents%centsPerPeso
	digits := fmt.Sprintf("%d", pesos)
	var b strings.Builder
	for i, r := range digits {
		if i > 0 && (len(digits)-i)%thousandsGroup == 0 {
			b.WriteByte(',')
		}
		b.WriteRune(r)
	}
	if rest == 0 {
		return sign + "$" + b.String()
	}
	return fmt.Sprintf("%s$%s.%02d", sign, b.String(), rest)
}

// CollapseID is the notification's collapse id (§2.4 step 5): the hex
// HMAC-SHA256 of (user, period, kind) under key, 64 characters (the
// apns-collapse-id limit). A resend of the same reminder carries the same
// id, so the device replaces the earlier notification instead of stacking.
// The parts are length-free (a uuid, "YYYY-MM", a kind) and joined by a
// byte none of them holds.
func CollapseID(key []byte, userID string, p Period, kind Kind) string {
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(userID + "\x00" + p.String() + "\x00" + string(kind)))
	return hex.EncodeToString(mac.Sum(nil))
}
