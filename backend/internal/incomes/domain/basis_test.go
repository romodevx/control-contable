package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/romodevx/control-contable/backend/internal/incomes/domain"
)

// TestDateFor is MODULE-incomes §2.2: ISR follows the regime's basis
// (accrual → invoiced_on, cash → collected_on); IVA is always cash
// (LIVA 1-B), so it follows collected_on for every regime.
func TestDateFor(t *testing.T) {
	invoiced, collected := "2026-09-05", "2026-10-02"
	tests := []struct {
		name      string
		basis     domain.IncomeBasis
		tax       domain.Tax
		invoiced  string
		collected string
		want      string
		wantOK    bool
	}{
		{"accrual ISR uses the invoice date", domain.BasisAccrual, domain.TaxISR, invoiced, collected, invoiced, true},
		{"accrual IVA uses the collection date", domain.BasisAccrual, domain.TaxIVA, invoiced, collected, collected, true},
		{"accrual pending collection has no IVA date", domain.BasisAccrual, domain.TaxIVA, invoiced, "", "", false},
		{"accrual ISR without invoice date", domain.BasisAccrual, domain.TaxISR, "", collected, "", false},
		{"cash ISR uses the collection date", domain.BasisCash, domain.TaxISR, invoiced, collected, collected, true},
		{"cash IVA uses the collection date", domain.BasisCash, domain.TaxIVA, invoiced, collected, collected, true},
		{"advance: collected before invoiced", domain.BasisCash, domain.TaxISR, collected, invoiced, invoiced, true},
		{"cash without collection date", domain.BasisCash, domain.TaxISR, invoiced, "", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dates := datesOf(t, tt.invoiced, tt.collected)
			got, ok := tt.basis.DateFor(tt.tax, dates)
			assert.Equal(t, tt.wantOK, ok)
			assert.Equal(t, tt.want, got.String())
		})
	}
}

func TestDateForUnknownTax(t *testing.T) {
	_, ok := domain.BasisCash.DateFor(domain.Tax("ieps"), datesOf(t, "2026-09-05", "2026-09-05"))
	assert.False(t, ok)
}

func TestBookingAndRateDates(t *testing.T) {
	tests := []struct {
		name        string
		basis       domain.IncomeBasis
		invoiced    string
		collected   string
		wantBooking string
		wantRate    string
		wantPending bool
	}{
		{"601 invoiced, pending collection", domain.BasisAccrual, "2026-09-05", "", "2026-09-05", "2026-09-05", true},
		{"601 invoiced and collected", domain.BasisAccrual, "2026-09-05", "2026-10-02", "2026-09-05", "2026-10-02", false},
		{"612 collected", domain.BasisCash, "", "2026-09-12", "2026-09-12", "2026-09-12", false},
		{"612 never pending even without collection", domain.BasisCash, "2026-09-05", "", "", "2026-09-05", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dates := datesOf(t, tt.invoiced, tt.collected)
			assert.Equal(t, tt.wantBooking, tt.basis.BookingDate(dates).String())
			assert.Equal(t, tt.wantRate, dates.RateDate().String())
			assert.Equal(t, tt.wantPending, tt.basis.PendingCollection(dates))
		})
	}
}

func TestRequiredDateField(t *testing.T) {
	assert.Equal(t, domain.FieldInvoicedOn, domain.BasisAccrual.RequiredDateField())
	assert.Equal(t, domain.FieldCollectedOn, domain.BasisCash.RequiredDateField())
}

func datesOf(t *testing.T, invoiced, collected string) domain.Dates {
	t.Helper()
	var d domain.Dates
	if invoiced != "" {
		d.InvoicedOn = date(t, invoiced)
	}
	if collected != "" {
		d.CollectedOn = date(t, collected)
	}
	return d
}
