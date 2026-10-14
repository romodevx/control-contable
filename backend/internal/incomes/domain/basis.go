package domain

// Tax names a tax whose income date depends on the regime.
type Tax string

// The taxes an income feeds.
const (
	TaxISR Tax = "isr"
	TaxIVA Tax = "iva"
)

// Dates are an income's two business dates; a zero Date is a null field.
type Dates struct {
	InvoicedOn  Date
	CollectedOn Date
}

// DateFor returns the date on which the income counts for tax, and false
// when that date is not set (MODULE-incomes §2.2):
//
//	basis    ISR           IVA
//	accrual  invoiced_on   collected_on (LIVA 1-B: IVA is always cash)
//	cash     collected_on  collected_on
//
// An accrual income with no collected_on counts for ISR in its invoice month
// and for IVA only once collected.
func (b IncomeBasis) DateFor(tax Tax, d Dates) (Date, bool) {
	var on Date
	switch {
	case tax == TaxIVA:
		on = d.CollectedOn
	case tax == TaxISR && b == BasisAccrual:
		on = d.InvoicedOn
	case tax == TaxISR:
		on = d.CollectedOn
	}
	return on, !on.IsZero()
}

// BookingDate is the ISR date: it decides the list period and is stored as
// booked_on. It is the zero Date when the required date is missing.
func (b IncomeBasis) BookingDate(d Dates) Date {
	on, _ := b.DateFor(TaxISR, d)
	return on
}

// PendingCollection reports an accrual income invoiced but not collected
// ("Por cobrar"). A cash income is never pending: it exists once collected.
func (b IncomeBasis) PendingCollection(d Dates) bool {
	return b == BasisAccrual && d.CollectedOn.IsZero()
}

// RequiredDateField is the contract field the basis requires.
func (b IncomeBasis) RequiredDateField() string {
	if b == BasisAccrual {
		return FieldInvoicedOn
	}
	return FieldCollectedOn
}

// RateDate is the day the IVA rate must be valid on (§2.5): collected_on, or
// invoiced_on for an income still pending collection.
func (d Dates) RateDate() Date {
	if d.CollectedOn.IsZero() {
		return d.InvoicedOn
	}
	return d.CollectedOn
}
