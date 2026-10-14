package main

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
)

// fakeLatest answers a canned latest record date of a record store.
type fakeLatest struct {
	day        fiscal.Date
	err        error
	taxpayerID string
	regime     fiscal.RegimeCode
}

func (f *fakeLatest) latest(taxpayerID string, regime fiscal.RegimeCode) (fiscal.Date, error) {
	f.taxpayerID, f.regime = taxpayerID, regime
	return f.day, f.err
}

func (f *fakeLatest) LatestPaidOn(_ context.Context, taxpayerID string, regime fiscal.RegimeCode) (fiscal.Date, error) {
	return f.latest(taxpayerID, regime)
}

func (f *fakeLatest) LatestRecordDate(_ context.Context, taxpayerID string, regime fiscal.RegimeCode) (fiscal.Date, error) {
	return f.latest(taxpayerID, regime)
}

// M2-14 G-1: the taxpayers module's record lookup reads the expenses and
// the incomes: the latest period holding a live record of the regime, the
// zero Period when neither has one.
func TestRegimeRecords_latestRecordPeriod(t *testing.T) {
	tests := []struct {
		name            string
		expense, income string
		want            string
	}{
		{name: "no records", want: ""},
		{name: "only an expense", expense: "2026-09-20", want: "2026-09"},
		{name: "only an income", income: "2026-10-02", want: "2026-10"},
		{name: "the later income wins", expense: "2026-08-31", income: "2026-10-02", want: "2026-10"},
		{name: "the later expense wins", expense: "2026-11-01", income: "2026-10-02", want: "2026-11"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expenses, incomes := &fakeLatest{}, &fakeLatest{}
			if tt.expense != "" {
				expenses.day = fiscalDate(t, tt.expense)
			}
			if tt.income != "" {
				incomes.day = fiscalDate(t, tt.income)
			}

			got, err := regimeRecords{expenses: expenses, incomes: incomes}.LatestRecordPeriod(t.Context(), "tp", "612")

			require.NoError(t, err)
			if tt.want == "" {
				assert.True(t, got.IsZero())
			} else {
				assert.Equal(t, period(t, tt.want), got)
			}
			for _, store := range []*fakeLatest{expenses, incomes} {
				assert.Equal(t, "tp", store.taxpayerID)
				assert.Equal(t, fiscal.RegimeCode("612"), store.regime)
			}
		})
	}
}

func TestRegimeRecords_passesStoreErrors(t *testing.T) {
	down := errors.New("db down")
	for name, records := range map[string]regimeRecords{
		"expenses": {expenses: &fakeLatest{err: down}, incomes: &fakeLatest{}},
		"incomes":  {expenses: &fakeLatest{}, incomes: &fakeLatest{err: down}},
	} {
		t.Run(name, func(t *testing.T) {
			_, err := records.LatestRecordPeriod(t.Context(), "tp", "612")

			assert.ErrorIs(t, err, down)
		})
	}
}
