package main

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/tx"
)

var errHolidaysDown = errors.New("fake: holidays down")

// systemKey marks a context inside the fake system unit of work.
type systemKey struct{}

// fakeSystem runs work inline, marking the context, and counts the runs.
type fakeSystem struct{ runs int }

func (f *fakeSystem) System(ctx context.Context, work tx.Work) error {
	f.runs++
	return work(context.WithValue(ctx, systemKey{}, true))
}

// noopTx is a user unit of work that is never run (wiring tests only build).
type noopTx struct{}

func (noopTx) ReadWrite(context.Context, string, tx.Work) error { return nil }
func (noopTx) ReadOnly(context.Context, string, tx.Work) error  { return nil }

// countingHolidays is a holidays reader that counts its reads and refuses
// to run outside a unit of work (as pgdb.Conn does).
type countingHolidays struct {
	set   fiscal.HolidaySet
	err   error
	reads int
}

func (c *countingHolidays) Holidays(ctx context.Context) (fiscal.HolidaySet, error) {
	if ctx.Value(systemKey{}) == nil {
		return fiscal.HolidaySet{}, errors.New("fake: read outside a unit of work")
	}
	c.reads++
	return c.set, c.err
}

func seededSet(t *testing.T) fiscal.HolidaySet {
	t.Helper()
	d, err := fiscal.ParseDate("2026-09-16")
	require.NoError(t, err)
	set, err := fiscal.NewHolidaySet([]int{2026, 2027}, []fiscal.Date{d})
	require.NoError(t, err)
	return set
}

// The holidays are global, immutable at run time (a new year is a
// migration, so a deploy) and tiny: they are read once at boot in a system
// unit of work, and every request is served from that snapshot with no
// statement (MODULE-dashboard §2.4, MODULE-history §2.6 keep their counts).
func TestLoadHolidays_readsOnceAtBootThenServesTheSnapshot(t *testing.T) {
	system := &fakeSystem{}
	reader := &countingHolidays{set: seededSet(t)}

	snapshot, err := loadHolidays(t.Context(), system, reader)
	require.NoError(t, err)
	for range 3 {
		got, err := snapshot.Holidays(t.Context())
		require.NoError(t, err)
		assert.Equal(t, []int{2026, 2027}, got.Years())
		d, _ := fiscal.ParseDate("2026-09-16")
		assert.True(t, got.IsHoliday(d))
	}

	assert.Equal(t, 1, system.runs, "one boot unit of work")
	assert.Equal(t, 1, reader.reads, "one boot read, none per request")
}

// A failed boot read fails the boot: serving due dates weekends-only by
// accident would be a silent wrong answer.
func TestLoadHolidays_failsTheBootWhenTheReadFails(t *testing.T) {
	_, err := loadHolidays(t.Context(), &fakeSystem{}, &countingHolidays{err: errHolidaysDown})

	require.ErrorIs(t, err, errHolidaysDown)
}

func TestLoadHolidays_needsItsPorts(t *testing.T) {
	_, err := loadHolidays(t.Context(), nil, &countingHolidays{})
	require.ErrorIs(t, err, errHolidaysWiring)

	_, err = loadHolidays(t.Context(), &fakeSystem{}, nil)
	require.ErrorIs(t, err, errHolidaysWiring)
}
