package taxparams_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/taxengine/domain/taxparams"
)

func TestAtResolvesByDate(t *testing.T) {
	p := mustParams(t, validSpec(t))
	tests := []struct {
		name    string
		key     taxparams.Key
		date    string
		want    string // the value's money or rate, as text
		wantErr error
	}{
		{name: "border rate on its last day", key: keyBorderRate, date: "2026-12-31", want: "0.08"},
		{name: "border rate the day it lapses", key: keyBorderRate, date: "2027-01-01", wantErr: taxparams.ErrNotInForce},
		{name: "border rate before it starts", key: keyBorderRate, date: "2025-12-31", wantErr: taxparams.ErrNotInForce},
		{name: "January UMA", key: keyUMA, date: "2026-01-31", want: "100.00"},
		{name: "February UMA", key: keyUMA, date: "2026-02-01", want: "110.00"},
		{name: "UMA into next January", key: keyUMA, date: "2027-01-31", want: "110.00"},
		{name: "UMA past its last range", key: keyUMA, date: "2027-02-01", wantErr: taxparams.ErrNotInForce},
		{name: "UMA before its first range", key: keyUMA, date: "2025-06-30", wantErr: taxparams.ErrNotInForce},
		{name: "unknown key", key: "iva.rate.unknown", date: "2026-03-31", wantErr: taxparams.ErrUnknownKey},
		{name: "a table key is not a value", key: keyMonthly, date: "2026-03-31", wantErr: taxparams.ErrKindMismatch},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, err := p.At(tt.key, mustDate(t, tt.date))
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.key, v.Key)
			got := v.Money.String()
			if v.Kind == taxparams.KindRate {
				got = v.Rate.String()
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestTypedValueLookups(t *testing.T) {
	p := mustParams(t, validSpec(t))
	march := mustDate(t, "2026-03-31")

	rate, err := p.Rate(keyBorderRate, march)
	require.NoError(t, err)
	assert.Equal(t, mustRate(t, "0.08"), rate)

	uma, err := p.Money(keyUMA, march)
	require.NoError(t, err)
	assert.Equal(t, mustCents(t, "110.00"), uma)

	flag, err := p.Flag(keyRounding, march)
	require.NoError(t, err)
	assert.True(t, flag)

	_, err = p.Rate(keyUMA, march)
	require.ErrorIs(t, err, taxparams.ErrKindMismatch)
	_, err = p.Money(keyBorderRate, march)
	require.ErrorIs(t, err, taxparams.ErrKindMismatch)
	_, err = p.Flag(keyUMA, march)
	require.ErrorIs(t, err, taxparams.ErrKindMismatch)
	_, err = p.Rate(keyBorderRate, mustDate(t, "2027-01-01"))
	require.ErrorIs(t, err, taxparams.ErrNotInForce)
	_, err = p.Money("uma.daily", march)
	require.ErrorIs(t, err, taxparams.ErrUnknownKey)
	_, err = p.Flag("rounding.other", march)
	require.ErrorIs(t, err, taxparams.ErrUnknownKey)
}

func TestTableLookups(t *testing.T) {
	p := mustParams(t, validSpec(t))
	march := mustDate(t, "2026-03-31")

	tariff, err := p.Table(keyMonthly, march)
	require.NoError(t, err)
	assert.Equal(t, taxparams.TableMarginal, tariff.Kind)
	require.Len(t, tariff.Rows, 2)
	assert.True(t, tariff.Rows[1].Unbounded)

	brackets, err := p.Table(keyBrackets, march)
	require.NoError(t, err)
	assert.Equal(t, taxparams.TableFlat, brackets.Kind)

	for m := time.January; m <= time.December; m++ {
		tbl, err := p.MonthlyTable(keyAccum, m, march)
		require.NoError(t, err)
		assert.Equal(t, m, tbl.Month)
		assert.Equal(t, accumulatedTable(t, m).Rows, tbl.Rows)
	}

	tests := []struct {
		name    string
		lookup  func() error
		wantErr error
	}{
		{"plain lookup of a monthly key", func() error { _, err := p.Table(keyAccum, march); return err }, taxparams.ErrKindMismatch},
		{"monthly lookup of a plain key", func() error { _, err := p.MonthlyTable(keyMonthly, time.March, march); return err }, taxparams.ErrKindMismatch},
		{"table lookup of a value key", func() error { _, err := p.Table(keyUMA, march); return err }, taxparams.ErrKindMismatch},
		{"monthly lookup of a value key", func() error { _, err := p.MonthlyTable(keyUMA, time.March, march); return err }, taxparams.ErrKindMismatch},
		{"unknown table", func() error { _, err := p.Table("isr.unknown", march); return err }, taxparams.ErrUnknownKey},
		{"unknown monthly table", func() error { _, err := p.MonthlyTable("isr.unknown", time.March, march); return err }, taxparams.ErrUnknownKey},
		{"month zero", func() error { _, err := p.MonthlyTable(keyAccum, 0, march); return err }, taxparams.ErrInvalidMonth},
		{"month 13", func() error { _, err := p.MonthlyTable(keyAccum, 13, march); return err }, taxparams.ErrInvalidMonth},
		{"table before its start", func() error { _, err := p.Table(keyMonthly, mustDate(t, "2025-12-31")); return err }, taxparams.ErrNotInForce},
		{"monthly table before its start", func() error {
			_, err := p.MonthlyTable(keyAccum, time.March, mustDate(t, "2025-12-31"))
			return err
		}, taxparams.ErrNotInForce},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.ErrorIs(t, tt.lookup(), tt.wantErr)
		})
	}
}

func TestParamsAreImmutable(t *testing.T) {
	spec := validSpec(t)
	p := mustParams(t, spec)
	march := mustDate(t, "2026-03-31")

	// Mutating the spec after construction changes nothing.
	spec.Tables[0].Rows[0].Rate = 0
	spec.Values[0].Basis[0].Article = "changed"
	spec.Sources[0].Title = "changed"

	// Mutating what a lookup returned changes nothing either.
	tbl, err := p.Table(keyMonthly, march)
	require.NoError(t, err)
	tbl.Rows[0].Rate = 0
	tbl.Basis[0].Article = "changed"
	v, err := p.At(keyBorderRate, march)
	require.NoError(t, err)
	v.Basis[0].Article = "changed"
	p.Sources()[0].Title = "changed"

	again, err := p.Table(keyMonthly, march)
	require.NoError(t, err)
	assert.Equal(t, mustRate(t, "0.02"), again.Rows[0].Rate)
	assert.Equal(t, "Art. 1", again.Basis[0].Article)
	value, err := p.At(keyBorderRate, march)
	require.NoError(t, err)
	assert.Equal(t, "Art. 1", value.Basis[0].Article)
	assert.Equal(t, "Synthetic decree", p.Sources()[0].Title)
}

func TestRequire(t *testing.T) {
	p := mustParams(t, validSpec(t))
	require.NoError(t, p.Require(
		taxparams.RequireRate(keyBorderRate),
		taxparams.RequireMoney(keyUMA),
		taxparams.RequireFlag(keyRounding),
		taxparams.RequireTable(keyMonthly),
		taxparams.RequireTable(keyBrackets),
		taxparams.RequireMonthlyTables(keyAccum),
	))

	err := p.Require(
		taxparams.RequireRate(keyUMA),
		taxparams.RequireMoney("uma.daily"),
		taxparams.RequireTable(keyAccum),
		taxparams.RequireMonthlyTables(keyMonthly),
		taxparams.RequireFlag(keyBorderRate),
	)
	require.ErrorIs(t, err, taxparams.ErrKindMismatch)
	require.ErrorIs(t, err, taxparams.ErrUnknownKey)
	for _, key := range []string{"uma.monthly", "uma.daily", "isr.art106.accumulated", "isr.art96.monthly", "iva.rate.border_8"} {
		assert.Contains(t, err.Error(), key)
	}
}

func TestKindStrings(t *testing.T) {
	assert.Equal(t, "rate", taxparams.KindRate.String())
	assert.Equal(t, "money", taxparams.KindMoney.String())
	assert.Equal(t, "flag", taxparams.KindFlag.String())
	assert.Equal(t, "unknown", taxparams.ValueKind(0).String())
	assert.Equal(t, "marginal", taxparams.TableMarginal.String())
	assert.Equal(t, "flat", taxparams.TableFlat.String())
	assert.Equal(t, "unknown", taxparams.TableKind(0).String())
}
