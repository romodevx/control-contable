package domain

import (
	"fmt"
	"time"

	"github.com/romodevx/control-contable/backend/internal/platform/money"
)

// Withheld splits a withholding by who withheld it from the taxpayer.
type Withheld struct {
	ByMoral    money.Cents // a persona moral (LISR 106 ¶5, 116 ¶5, 113-J; LIVA 1-A)
	ByPlatform money.Cents // a digital platform (LISR 113-A, LIVA 18-J)
}

// RateSums are the income collected at one taxable IVA rate: the base
// (amount_cents) and the IVA charged on it (iva_cents), as entered.
type RateSums struct {
	Base money.Cents
	IVA  money.Cents
}

// IVAByRate splits a month's income collected, or expenses paid, by the
// record's IVA treatment (contract IvaRate). The rate is the record's own:
// its module checked it against the rate catalog on the record's IVA date,
// so the border 8 % counts as entered whatever the period
// (MODULE-tax-engine §2.4). A 0 % or exempt record carries no IVA, so only
// its base is kept; on the income side both bases matter to proportional
// crediting (LIVA 5 fr. V).
type IVAByRate struct {
	Rate16     RateSums    // rate_16, LIVA 1
	Rate8      RateSums    // rate_8, border regions while the decree is in force
	Rate0Base  money.Cents // rate_0, LIVA 2-A
	ExemptBase money.Cents // exempt, LIVA 9, 15, 20
}

// IncomeSums are one regime's income aggregates for one month, as entered
// on the records (non-deleted only, MODULE-tax-engine §2.4). Amounts are
// before IVA.
type IncomeSums struct {
	Cash        money.Cents // amount_cents by collected_on (cash-basis ISR)
	Accrual     money.Cents // amount_cents by invoiced_on (accrual-basis ISR, 601)
	ByRate      IVAByRate   // amount_cents and iva_cents by collected_on (IVA is always cash, LIVA 1-B)
	IVAWithheld Withheld    // iva_withheld_cents by collected_on
	ISRWithheld Withheld    // isr_withheld_cents by collected_on
}

// ExpenseSums are one regime's expense aggregates for one month, by
// paid_on (the booking date of every expense, MODULE-expenses §2.1), over
// non-deleted records. Amounts are before IVA.
//
// DeductiblePaid, NonDeductiblePaid and CashOverLimitPaid are the
// deductibility classes and partition the month's amount_cents: deductible
// is the stored effective verdict (MODULE-expenses §2.4, an override
// included); a non-deductible expense paid in cash over the limit (§2.5)
// counts as cash-over-limit, any other one as non-deductible.
type ExpenseSums struct {
	DeductiblePaid    money.Cents // amount_cents of effectively deductible expenses
	NonDeductiblePaid money.Cents // amount_cents of the other non-deductible expenses
	CashOverLimitPaid money.Cents // amount_cents of non-deductible cash payments over the limit (LISR 27 fr. III)
	IVAPaid           IVAByRate   // amount_cents and iva_cents by the record's IVA rate, all classes
	// CreditableIVAPaid is the paid part of the stored creditable IVA (LIVA
	// 5, MODULE-expenses §2.4: CFDI, never a cash payment over the limit, a
	// creditable category, and deductible or RESICO PF): the creditable
	// share of IVA − IVA withheld, credited in its own month.
	CreditableIVAPaid money.Cents
	// CreditableIVAWithheld is the withheld part: the creditable share of
	// the IVA the taxpayer withheld from its suppliers this month. It is
	// credited in the next month's return, the one after the return that
	// remits it (LIVA 5 fr. IV, MODULE-withholdings §4.2), never in its own.
	CreditableIVAWithheld money.Cents
	PredialPaid           money.Cents // amount_cents of deductible predial (606 blind deduction, LISR 115 ¶2)
	ISRWithheldToRemit    money.Cents // ISR the taxpayer withheld from suppliers
	IVAWithheldToRemit    money.Cents // IVA the taxpayer withheld from suppliers
}

// MonthSums are one regime's ledger aggregates for one month: what the
// LedgerReader returns, one indexed GROUP BY per table (ADR-0002 §1).
type MonthSums struct {
	Income   IncomeSums
	Expenses ExpenseSums
}

// summer adds centavos with a sticky overflow error, so a sum of many
// fields checks each one without an if per field.
type summer struct{ err error }

func (s *summer) add(a, b money.Cents) money.Cents {
	if s.err != nil {
		return 0
	}
	sum, err := a.Add(b)
	s.err = err
	return sum
}

func (s *summer) rateSums(a, b RateSums) RateSums {
	return RateSums{Base: s.add(a.Base, b.Base), IVA: s.add(a.IVA, b.IVA)}
}

func (s *summer) byRate(a, b IVAByRate) IVAByRate {
	return IVAByRate{
		Rate16:     s.rateSums(a.Rate16, b.Rate16),
		Rate8:      s.rateSums(a.Rate8, b.Rate8),
		Rate0Base:  s.add(a.Rate0Base, b.Rate0Base),
		ExemptBase: s.add(a.ExemptBase, b.ExemptBase),
	}
}

func (s *summer) withheld(a, b Withheld) Withheld {
	return Withheld{ByMoral: s.add(a.ByMoral, b.ByMoral), ByPlatform: s.add(a.ByPlatform, b.ByPlatform)}
}

// Add returns m + o field by field, or money.ErrOverflow.
func (m MonthSums) Add(o MonthSums) (MonthSums, error) {
	var s summer
	sum := MonthSums{
		Income: IncomeSums{
			Cash:        s.add(m.Income.Cash, o.Income.Cash),
			Accrual:     s.add(m.Income.Accrual, o.Income.Accrual),
			ByRate:      s.byRate(m.Income.ByRate, o.Income.ByRate),
			IVAWithheld: s.withheld(m.Income.IVAWithheld, o.Income.IVAWithheld),
			ISRWithheld: s.withheld(m.Income.ISRWithheld, o.Income.ISRWithheld),
		},
		Expenses: ExpenseSums{
			DeductiblePaid:        s.add(m.Expenses.DeductiblePaid, o.Expenses.DeductiblePaid),
			NonDeductiblePaid:     s.add(m.Expenses.NonDeductiblePaid, o.Expenses.NonDeductiblePaid),
			CashOverLimitPaid:     s.add(m.Expenses.CashOverLimitPaid, o.Expenses.CashOverLimitPaid),
			IVAPaid:               s.byRate(m.Expenses.IVAPaid, o.Expenses.IVAPaid),
			CreditableIVAPaid:     s.add(m.Expenses.CreditableIVAPaid, o.Expenses.CreditableIVAPaid),
			CreditableIVAWithheld: s.add(m.Expenses.CreditableIVAWithheld, o.Expenses.CreditableIVAWithheld),
			PredialPaid:           s.add(m.Expenses.PredialPaid, o.Expenses.PredialPaid),
			ISRWithheldToRemit:    s.add(m.Expenses.ISRWithheldToRemit, o.Expenses.ISRWithheldToRemit),
			IVAWithheldToRemit:    s.add(m.Expenses.IVAWithheldToRemit, o.Expenses.IVAWithheldToRemit),
		},
	}
	if s.err != nil {
		return MonthSums{}, s.err
	}
	return sum, nil
}

// MonthEntry is one month's sums, as NewYearSums takes them.
type MonthEntry struct {
	Month time.Month
	Sums  MonthSums
}

// YearSums are one regime's MonthSums for January..Through of the period's
// fiscal year, plus the withheld creditable IVA of the previous fiscal
// year's December, which January credits (LIVA 5 fr. IV). A month without
// records is all zeros.
type YearSums struct {
	through       time.Month
	months        [monthsPerYear]MonthSums
	priorDecember money.Cents // ExpenseSums.CreditableIVAWithheld of the previous December
}

// monthsPerYear is the length of a fiscal year in tax periods.
const monthsPerYear = 12

// NewYearSums builds the sums for January..through. Each entry's month must
// be in that range and appear once, else ErrInvalidMonth.
func NewYearSums(through time.Month, entries ...MonthEntry) (YearSums, error) {
	if !validMonth(through) {
		return YearSums{}, fmt.Errorf("through month %d: %w", through, ErrInvalidMonth)
	}
	y := YearSums{through: through}
	seen := map[time.Month]bool{}
	for _, e := range entries {
		if !validMonth(e.Month) || e.Month > through || seen[e.Month] {
			return YearSums{}, fmt.Errorf("month %d through %s: %w", e.Month, through, ErrInvalidMonth)
		}
		seen[e.Month] = true
		y.months[e.Month-1] = e.Sums
	}
	return y, nil
}

// Through returns the last month held: the period's month.
func (y YearSums) Through() time.Month { return y.through }

// Month returns month's sums; zero outside January..Through.
func (y YearSums) Month(month time.Month) MonthSums {
	if !validMonth(month) || month > y.through {
		return MonthSums{}
	}
	return y.months[month-1]
}

// WithPriorDecember returns a copy of y holding withheld, the creditable IVA
// withheld from suppliers in the previous fiscal year's December: LIVA is
// monthly, not annual, so January credits it (MODULE-withholdings §4.2).
func (y YearSums) WithPriorDecember(withheld money.Cents) YearSums {
	y.priorDecember = withheld
	return y
}

// CreditableWithheldPriorMonth returns the creditable IVA withheld from
// suppliers in the month before month, which month's return credits (LIVA
// 5 fr. IV): the previous December's for January; zero outside
// January..Through.
func (y YearSums) CreditableWithheldPriorMonth(month time.Month) money.Cents {
	switch {
	case !validMonth(month) || month > y.through:
		return 0
	case month == time.January:
		return y.priorDecember
	}
	return y.months[month-2].Expenses.CreditableIVAWithheld
}

// Sum returns the sums of from..to inclusive: Sum(January, n) is year to
// date, Sum(n-1, n) a bimester. It fails with ErrInvalidMonth for a range
// outside January..Through or ending before it starts, or money.ErrOverflow.
func (y YearSums) Sum(from, to time.Month) (MonthSums, error) {
	if !validMonth(from) || to < from || to > y.through {
		return MonthSums{}, fmt.Errorf("months %d..%d through %s: %w", from, to, y.through, ErrInvalidMonth)
	}
	var total MonthSums
	for m := from; m <= to; m++ {
		var err error
		if total, err = total.Add(y.months[m-1]); err != nil {
			return MonthSums{}, fmt.Errorf("months %s..%s: %w", from, to, err)
		}
	}
	return total, nil
}
