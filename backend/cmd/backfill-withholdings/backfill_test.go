package main

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	expenses "github.com/romodevx/control-contable/backend/internal/expenses/usecase"
	incomes "github.com/romodevx/control-contable/backend/internal/incomes/usecase"
	"github.com/romodevx/control-contable/backend/internal/platform/tx"
)

const (
	user1 = "0199d4a2-0000-7000-8000-0000000000b1"
	user2 = "0199d4a2-0000-7000-8000-0000000000b2"
)

// fakeUnits records the per-user units of work and their modes.
type fakeUnits struct{ calls []string }

func (f *fakeUnits) ReadWrite(ctx context.Context, userID string, work tx.Work) error {
	f.calls = append(f.calls, "rw:"+userID)
	return work(ctx)
}

func (f *fakeUnits) ReadOnly(ctx context.Context, userID string, work tx.Work) error {
	f.calls = append(f.calls, "ro:"+userID)
	return work(ctx)
}

type fakeUsers struct {
	ids []string
	err error
}

func (f fakeUsers) UserIDs(context.Context) ([]string, error) { return f.ids, f.err }

type fakeIncomes struct {
	byUser map[string]incomes.BackfillResult
	err    error
	calls  []string
}

func (f *fakeIncomes) Run(_ context.Context, userID string, apply bool) (incomes.BackfillResult, error) {
	f.calls = append(f.calls, userID)
	return f.byUser[userID], f.err
}

type fakeExpenses struct {
	byUser map[string]expenses.BackfillResult
	err    error
}

func (f *fakeExpenses) Run(_ context.Context, userID string, _ bool) (expenses.BackfillResult, error) {
	return f.byUser[userID], f.err
}

func newFakeBackfiller() (*backfiller, *fakeUnits, *fakeIncomes, *fakeExpenses) {
	units := &fakeUnits{}
	inc := &fakeIncomes{byUser: map[string]incomes.BackfillResult{
		user1: {Rederived: []incomes.Rederived{{ID: "i1", AmountsChanged: true, FactsAssumed: true}}, NeedsInput: []string{"i9"}},
		user2: {Rederived: []incomes.Rederived{{ID: "i2"}}},
	}}
	exp := &fakeExpenses{byUser: map[string]expenses.BackfillResult{
		user2: {Rederived: []expenses.Rederived{{ID: "e2", FactsAssumed: true}}},
	}}
	return &backfiller{units: units, users: fakeUsers{ids: []string{user1, user2}}, incomes: inc, expenses: exp}, units, inc, exp
}

func TestBackfiller_OneUnitOfWorkPerUser(t *testing.T) {
	for _, apply := range []bool{false, true} {
		b, units, _, _ := newFakeBackfiller()

		got, err := b.run(t.Context(), apply)

		require.NoError(t, err)
		mode := "ro:"
		if apply {
			mode = "rw:"
		}
		assert.Equal(t, []string{mode + user1, mode + user2}, units.calls)
		assert.Equal(t, report{
			Apply: apply, Users: 2,
			Incomes: tableReport{
				Rederived: []string{"i1", "i2"}, AmountsChanged: []string{"i1"}, FactsAssumed: []string{"i1"},
				NeedsInput: []string{"i9"},
			},
			Expenses: tableReport{Rederived: []string{"e2"}, FactsAssumed: []string{"e2"}},
		}, got)
	}
}

func TestBackfiller_AbortsOnTheFirstError(t *testing.T) {
	boom := errors.New("boom")
	t.Run("users", func(t *testing.T) {
		b, _, _, _ := newFakeBackfiller()
		b.users = fakeUsers{err: boom}
		_, err := b.run(t.Context(), true)
		require.ErrorIs(t, err, boom)
	})
	t.Run("incomes", func(t *testing.T) {
		b, _, inc, _ := newFakeBackfiller()
		inc.err = boom
		_, err := b.run(t.Context(), true)
		require.ErrorIs(t, err, boom)
		assert.Equal(t, []string{user1}, inc.calls, "no later user is visited")
	})
	t.Run("expenses", func(t *testing.T) {
		b, _, _, exp := newFakeBackfiller()
		exp.err = boom
		_, err := b.run(t.Context(), false)
		require.ErrorIs(t, err, boom)
	})
}

func TestReport_IDsAndCountsOnly(t *testing.T) {
	r := report{Users: 3, Incomes: tableReport{
		Rederived: []string{"b", "a"}, AmountsChanged: []string{"b", "a"}, NeedsInput: []string{"c"},
	}}
	var out bytes.Buffer
	require.NoError(t, r.write(&out))

	text := out.String()
	assert.Contains(t, text, "DRY RUN")
	assert.Contains(t, text, "table_version: 2026.1")
	assert.Contains(t, text, "users visited: 3")
	assert.Contains(t, text, "incomes: 2 to re-derive, 2 with different amounts, 0 with assumed facts, 1 need user input")
	assert.Contains(t, text, "expenses: 0 to re-derive")
	assert.Less(t, bytes.Index(out.Bytes(), []byte("    a\n")), bytes.Index(out.Bytes(), []byte("    b\n")), "ids sorted")
	assert.NotContains(t, text, "assumed facts (ask")

	out.Reset()
	r.Apply = true
	require.NoError(t, r.write(&out))
	assert.Contains(t, out.String(), "APPLIED")
}
