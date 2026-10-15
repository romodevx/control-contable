package annual_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/platform/apperr"
	"github.com/romodevx/control-contable/backend/internal/platform/clock"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
	"github.com/romodevx/control-contable/backend/internal/taxengine/usecase/annual"
)

func (w *world) filed(now time.Time) *annual.FiledPayments {
	return annual.NewFiledPayments(annual.FiledDeps{
		Tx: w.tx, Access: w.access, Taxpayers: w.db, Store: w.db, Clock: clock.NewFixed(now), Logger: w.logger(),
	})
}

func filedReq(taxpayerID, code string, year int) annual.FiledRequest {
	return annual.FiledRequest{TaxpayerID: taxpayerID, RegimeCode: code, FiscalYear: year}
}

// summary is each month as "in-window/closed/filed".
func summary(v annual.FiledView) []string {
	out := make([]string, 0, len(v.Months))
	for _, m := range v.Months {
		s := ""
		if m.InWindow {
			s += "w"
		}
		if m.Closed {
			s += "c"
		}
		if m.Filed != nil {
			s += "=" + m.Filed.String()
		}
		out = append(out, s)
	}
	return out
}

func TestFiled_GetTwelveMonths(t *testing.T) {
	w := newWorld()
	t2 := w.db.taxpayers[fisica2A]
	t2.regimes[0].Start = mustPeriod("2026-03")
	w.db.taxpayers[fisica2A] = t2

	got, err := w.filed(today).Get(t.Context(), userA, filedReq(fisica2A, "612", 2026))
	require.NoError(t, err)
	assert.Equal(t, []string{"c", "c", "wc", "wc", "wc", "wc", "wc", "wc", "wc", "w", "w", "w"}, summary(got),
		"612 from March; October is not closed on 2026-10-14")
	assert.Equal(t, "2026-03", got.Months[2].Period.String())
	assert.Equal(t, time.March, got.Months[2].Month)
	assert.Nil(t, got.UpdatedAt)
	assert.Equal(t, []string{"ro"}, w.tx.modes)
}

func TestFiled_PutReplacesAndConfirmAll(t *testing.T) {
	w := newWorld()
	uc := w.filed(today)

	put, err := uc.Put(t.Context(), userA, filedReq(fisicaA, "612", 2026), []annual.FiledEntry{
		{Month: 2, Amount: 500_000}, {Month: 1, Amount: 0},
	})
	require.NoError(t, err)
	assert.Equal(t, []string{"wc=0.00", "wc=5000.00", "wc", "wc", "wc", "wc", "wc", "wc", "wc", "w", "w", "w"}, summary(put),
		"a zero is a confirmed zero")
	require.NotNil(t, put.UpdatedAt)
	assert.Equal(t, []string{"taxpayer:write", "regime:612", "replace"}, w.db.calls)
	assert.Contains(t, w.logs.String(), `"msg":"annual_filed_payments.updated"`)
	assert.NotContains(t, w.logs.String(), "500000")

	// "Confirmar todos": the client sends the nine prefilled closed months.
	var all []annual.FiledEntry
	for m := 1; m <= 9; m++ {
		all = append(all, annual.FiledEntry{Month: m, Amount: money.Cents(m * 100)})
	}
	confirmed, err := uc.Put(t.Context(), userA, filedReq(fisicaA, "612", 2026), all)
	require.NoError(t, err)
	assert.Equal(t, []string{"wc=1.00", "wc=2.00", "wc=3.00", "wc=4.00", "wc=5.00", "wc=6.00", "wc=7.00", "wc=8.00", "wc=9.00", "w", "w", "w"}, summary(confirmed),
		"replace, not merge")

	got, err := uc.Get(t.Context(), userA, filedReq(fisicaA, "612", 2026))
	require.NoError(t, err)
	assert.Equal(t, confirmed, got)

	cleared, err := uc.Put(t.Context(), userA, filedReq(fisicaA, "612", 2026), nil)
	require.NoError(t, err)
	assert.Nil(t, cleared.UpdatedAt, "an empty set clears every month")
}

func TestFiled_PastYearIsClosed(t *testing.T) {
	w := newWorld()
	feb2027 := time.Date(2027, time.February, 10, 12, 0, 0, 0, clock.Location())
	got, err := w.filed(feb2027).Put(t.Context(), userA, filedReq(fisicaA, "612", 2026), []annual.FiledEntry{{Month: 12, Amount: 1}})
	require.NoError(t, err)
	assert.True(t, got.Months[11].Closed)
	assert.NotNil(t, got.Months[11].Filed)
}

func TestFiled_Refusals(t *testing.T) {
	definitive := func(w *world) {
		t2 := w.db.taxpayers[fisica2A]
		t2.regimes = append(t2.regimes, annual.HeldRegime{Code: "625", Start: mustPeriod("2026-01"), DefinitiveOption: true})
		w.db.taxpayers[fisica2A] = t2
	}
	ended := func(w *world) {
		t2 := w.db.taxpayers[fisica2A]
		t2.regimes[0].End = mustPeriod("2026-06")
		w.db.taxpayers[fisica2A] = t2
	}
	tests := []struct {
		name    string
		setup   func(w *world)
		req     annual.FiledRequest
		entries []annual.FiledEntry
		put     bool
		code    apperr.Code
		fields  []string
	}{
		{"malformed taxpayer", nil, filedReq("x", "612", 2026), nil, false, apperr.CodeNotFound, []string{}},
		{"future year", nil, filedReq(fisicaA, "612", 2027), nil, false, apperr.CodeValidationFailed, []string{"fiscalYear/out_of_range"}},
		{"another user's taxpayer", nil, filedReq(fisicaB, "612", 2026), nil, false, apperr.CodeNotFound, []string{}},
		{"regime never held", nil, filedReq(fisicaA, "606", 2026), nil, false, apperr.CodeBusinessRuleViolation, []string{"regimeCode/regime_not_active"}},
		{"regime not held that year", nil, filedReq(moralA, "601", 2023), nil, false, apperr.CodeBusinessRuleViolation, []string{"regimeCode/regime_not_active"}},
		{"before 2026", nil, filedReq(fisicaA, "612", 2025), nil, false, apperr.CodeBusinessRuleViolation, []string{"fiscalYear/parameters_unavailable"}},
		{"605 has no provisional ISR", nil, filedReq(fisicaA, "605", 2026), nil, false, apperr.CodeBusinessRuleViolation, []string{"regimeCode/payments_not_applicable"}},
		{"625 definitive", definitive, filedReq(fisica2A, "625", 2026), nil, false, apperr.CodeBusinessRuleViolation, []string{"regimeCode/payments_not_applicable"}},
		{"put: read-only first", func(w *world) { w.access.readOnly = true; w.archive(fisicaA) }, filedReq(fisicaA, "612", 2026), nil, true,
			apperr.CodeEntitlementRequired, []string{}},
		{"put: archived", func(w *world) { w.archive(fisicaA) }, filedReq(fisicaA, "612", 2026), nil, true,
			apperr.CodeBusinessRuleViolation, []string{"taxpayerId/taxpayer_archived"}},
		{"put: another user's", nil, filedReq(fisicaB, "612", 2026), nil, true, apperr.CodeNotFound, []string{}},
		{"put: never held", nil, filedReq(fisicaA, "606", 2026), nil, true, apperr.CodeBusinessRuleViolation, []string{"regimeCode/regime_not_active"}},
		{"put: 605", nil, filedReq(fisicaA, "605", 2026), nil, true, apperr.CodeBusinessRuleViolation, []string{"regimeCode/payments_not_applicable"}},
		{"put: a month outside the window before month rules", ended, filedReq(fisica2A, "612", 2026),
			[]annual.FiledEntry{{Month: 7, Amount: 1}, {Month: 7, Amount: 1}, {Month: 6, Amount: 1}}, true,
			apperr.CodeBusinessRuleViolation, []string{"months.0.month/regime_not_active", "months.1.month/regime_not_active"}},
		{"put: month rules", nil, filedReq(fisicaA, "612", 2026),
			[]annual.FiledEntry{{Month: 3, Amount: 1}, {Month: 3, Amount: 2}, {Month: 10, Amount: 1}, {Month: 13, Amount: 1}, {Month: 4, Amount: -1}}, true,
			apperr.CodeValidationFailed, []string{"months.1.month/invalid", "months.2.month/out_of_range", "months.3.month/invalid", "months.4.amount_cents/invalid"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := newWorld()
			if tt.setup != nil {
				tt.setup(w)
			}
			var err error
			if tt.put {
				_, err = w.filed(today).Put(t.Context(), userA, tt.req, tt.entries)
			} else {
				_, err = w.filed(today).Get(t.Context(), userA, tt.req)
			}
			assert.Equal(t, tt.fields, requireCode(t, err, tt.code))
			assert.NotContains(t, w.db.calls, "replace")
		})
	}
	t.Run("reads stay open to read-only users and archived taxpayers", func(t *testing.T) {
		w := newWorld()
		w.access.readOnly = true
		w.archive(fisicaA)
		_, err := w.filed(today).Get(t.Context(), userA, filedReq(fisicaA, "612", 2026))
		require.NoError(t, err)
	})
}
