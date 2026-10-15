package domain_test

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/reminders/domain"
)

func civil(year int, month time.Month, day int) time.Time {
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}

func april2026(t *testing.T) domain.Period {
	t.Helper()
	p, err := domain.NewPeriod(2026, time.April)
	require.NoError(t, err)
	return p
}

func TestNewPeriod(t *testing.T) {
	p, err := domain.NewPeriod(2026, time.April)
	require.NoError(t, err)
	assert.Equal(t, "2026-04", p.String())
	assert.Equal(t, 2026, p.Year())
	assert.Equal(t, time.April, p.Month())

	for _, bad := range []struct {
		year  int
		month time.Month
	}{{2026, 0}, {2026, 13}, {0, time.April}, {10000, time.April}} {
		_, err := domain.NewPeriod(bad.year, bad.month)
		assert.ErrorIs(t, err, domain.ErrInvalidPeriod, "%d-%d", bad.year, bad.month)
	}
}

func TestPeriodFromDate(t *testing.T) {
	p, err := domain.PeriodOf(civil(2026, time.April, 1))
	require.NoError(t, err)
	assert.Equal(t, "2026-04", p.String())
}

func TestParseKind(t *testing.T) {
	for _, k := range []domain.Kind{domain.KindEarly, domain.KindDue} {
		got, err := domain.ParseKind(string(k))
		require.NoError(t, err)
		assert.Equal(t, k, got)
	}
	_, err := domain.ParseKind("late")
	assert.ErrorIs(t, err, domain.ErrInvalidKind)
}

// MODULE-reminders §2.5, the table's rows, verbatim.
func TestBuildNotification_SpecCopy(t *testing.T) {
	ana := domain.Listed{Name: "Ana"}
	taller := domain.Listed{Name: "Taller Sur"}
	tests := []struct {
		name      string
		kind      domain.Kind
		taxpayers []domain.Listed
		wantTitle string
		wantBody  string
	}{
		{
			name: "early, no amounts", kind: domain.KindEarly, taxpayers: []domain.Listed{ana, taller},
			wantTitle: "Tu pago mensual estimado vence el lunes 18 de mayo",
			wantBody:  "Revisa tu estimado de abril para: Ana, Taller Sur.",
		},
		{
			name: "due date, no amounts", kind: domain.KindDue, taxpayers: []domain.Listed{ana, taller},
			wantTitle: "Hoy vence tu pago mensual estimado",
			wantBody:  "Revisa tu estimado de abril para: Ana, Taller Sur.",
		},
		{
			name: "with amounts", kind: domain.KindEarly,
			taxpayers: []domain.Listed{
				{Name: "Ana", Estimate: domain.EstimateAmount, AmountCents: 4_520_00},
				{Name: "Taller Sur", Estimate: domain.EstimateAmount, AmountCents: 12_300_00},
			},
			wantTitle: "Tu pago mensual estimado vence el lunes 18 de mayo",
			wantBody:  "Ana: $4,520 estimado · Taller Sur: $12,300 estimado. Cifras aproximadas.",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := domain.BuildMessage(domain.MessageInput{
				Kind: tc.kind, Period: april2026(t), DueDate: civil(2026, time.May, 18), Taxpayers: tc.taxpayers,
			})

			require.NoError(t, err)
			assert.Equal(t, tc.wantTitle, got.Title)
			assert.Equal(t, tc.wantBody, got.Body)
			assert.Equal(t, "controlcontable://estimate?period=2026-04", got.DeepLink)
		})
	}
}

func TestBuildNotification_ListsThreeThenCountsTheRest(t *testing.T) {
	names := []domain.Listed{{Name: "A"}, {Name: "B"}, {Name: "C"}, {Name: "D"}, {Name: "E"}}

	got, err := domain.BuildMessage(domain.MessageInput{
		Kind: domain.KindDue, Period: april2026(t), DueDate: civil(2026, time.May, 18), Taxpayers: names,
	})

	require.NoError(t, err)
	assert.Equal(t, "Revisa tu estimado de abril para: A, B, C y 2 más.", got.Body)
}

func TestBuildNotification_OneTaxpayer(t *testing.T) {
	got, err := domain.BuildMessage(domain.MessageInput{
		Kind: domain.KindDue, Period: april2026(t), DueDate: civil(2026, time.May, 18),
		Taxpayers: []domain.Listed{{Name: "Ana"}},
	})

	require.NoError(t, err)
	assert.Equal(t, "Revisa tu estimado de abril para: Ana.", got.Body)
}

// §2.5: with amounts, a taxpayer without a computed estimate is listed by
// name only, or "estimado pendiente" when its estimate is incomplete.
func TestBuildNotification_AmountsMixedWithPendingAndNameOnly(t *testing.T) {
	got, err := domain.BuildMessage(domain.MessageInput{
		Kind: domain.KindDue, Period: april2026(t), DueDate: civil(2026, time.May, 18),
		Taxpayers: []domain.Listed{
			{Name: "Ana", Estimate: domain.EstimateAmount, AmountCents: 1_234_567_50},
			{Name: "Beto", Estimate: domain.EstimatePending},
			{Name: "Caro"},
			{Name: "Dani", Estimate: domain.EstimateAmount, AmountCents: 99},
		},
	})

	require.NoError(t, err)
	assert.Equal(t,
		"Ana: $1,234,567.50 estimado · Beto: estimado pendiente · Caro y 1 más. Cifras aproximadas.", got.Body)
}

// No figure to show: the plain body, never "Cifras aproximadas" alone.
func TestBuildNotification_AmountsModeWithoutAnyAmountFallsBackToNames(t *testing.T) {
	got, err := domain.BuildMessage(domain.MessageInput{
		Kind: domain.KindDue, Period: april2026(t), DueDate: civil(2026, time.May, 18),
		Taxpayers: []domain.Listed{{Name: "Ana", Estimate: domain.EstimatePending}, {Name: "Beto"}},
	})

	require.NoError(t, err)
	assert.Equal(t, "Revisa tu estimado de abril para: Ana, Beto.", got.Body)
}

func TestBuildNotification_SpanishCalendarNames(t *testing.T) {
	days := []struct {
		date time.Time
		want string
	}{
		{civil(2026, time.October, 19), "lunes 19 de octubre"},
		{civil(2026, time.September, 17), "jueves 17 de septiembre"},
		{civil(2026, time.June, 17), "miércoles 17 de junio"},
		{civil(2026, time.January, 17), "sábado 17 de enero"},
		{civil(2026, time.May, 17), "domingo 17 de mayo"},
		{civil(2026, time.March, 17), "martes 17 de marzo"},
		{civil(2026, time.April, 17), "viernes 17 de abril"},
	}
	for _, d := range days {
		got, err := domain.BuildMessage(domain.MessageInput{
			Kind: domain.KindEarly, Period: april2026(t), DueDate: d.date, Taxpayers: []domain.Listed{{Name: "Ana"}},
		})
		require.NoError(t, err)
		assert.Equal(t, "Tu pago mensual estimado vence el "+d.want, got.Title)
	}
	months := []string{"enero", "febrero", "marzo", "abril", "mayo", "junio", "julio", "agosto", "septiembre",
		"octubre", "noviembre", "diciembre"}
	for i, want := range months {
		p, err := domain.NewPeriod(2026, time.Month(i+1))
		require.NoError(t, err)
		got, err := domain.BuildMessage(domain.MessageInput{
			Kind: domain.KindDue, Period: p, DueDate: civil(2026, time.May, 18), Taxpayers: []domain.Listed{{Name: "Ana"}},
		})
		require.NoError(t, err)
		assert.Equal(t, "Revisa tu estimado de "+want+" para: Ana.", got.Body)
	}
}

func TestBuildNotification_Refuses(t *testing.T) {
	_, err := domain.BuildMessage(domain.MessageInput{Kind: domain.KindDue, Period: april2026(t), DueDate: civil(2026, time.May, 18)})
	require.ErrorIs(t, err, domain.ErrNothingToRemind)

	_, err = domain.BuildMessage(domain.MessageInput{Kind: "late", Period: april2026(t), DueDate: civil(2026, time.May, 18),
		Taxpayers: []domain.Listed{{Name: "Ana"}}})
	require.ErrorIs(t, err, domain.ErrInvalidKind)
}

// M9 security gate F2: a listed name is at most MaxListedNameRunes
// characters (the contract's display_name maxLength), so a name that got
// past the taxpayers module's own check cannot grow the payload.
func TestBuildNotification_CapsEachName(t *testing.T) {
	long := strings.Repeat("ñ", domain.MaxListedNameRunes+5)
	exact := strings.Repeat("b", domain.MaxListedNameRunes)
	in := domain.MessageInput{
		Kind: domain.KindDue, Period: april2026(t), DueDate: civil(2026, time.May, 18),
		Taxpayers: []domain.Listed{{Name: long}, {Name: exact, Estimate: domain.EstimatePending}},
	}

	names, err := domain.BuildMessage(in)
	require.NoError(t, err)
	in.Taxpayers[0].Estimate, in.Taxpayers[0].AmountCents = domain.EstimateAmount, 452000
	amounts, err := domain.BuildMessage(in)
	require.NoError(t, err)

	capped := strings.Repeat("ñ", domain.MaxListedNameRunes-1) + "…"
	assert.Equal(t, "Revisa tu estimado de abril para: "+capped+", "+exact+".", names.Body)
	assert.True(t, strings.HasPrefix(amounts.Body, capped+": $"), amounts.Body)
	assert.Contains(t, amounts.Body, exact+": estimado pendiente")
}

// §2.4 step 5: hex HMAC(user, period, kind), 64 characters (the
// apns-collapse-id limit), stable, and distinct per key part.
func TestCollapseID(t *testing.T) {
	key := []byte(strings.Repeat("k", 32))
	user := "0199d4a2-1111-7000-8000-000000000001"
	p := april2026(t)

	id := domain.CollapseID(key, user, p, domain.KindEarly)

	assert.Len(t, id, 64)
	assert.Regexp(t, `^[0-9a-f]{64}$`, id)
	assert.Equal(t, id, domain.CollapseID(key, user, p, domain.KindEarly), "stable across retries")
	assert.NotEqual(t, id, domain.CollapseID(key, user, p, domain.KindDue))
	assert.NotEqual(t, id, domain.CollapseID(key, "0199d4a2-1111-7000-8000-000000000002", p, domain.KindEarly))
	may, err := domain.NewPeriod(2026, time.May)
	require.NoError(t, err)
	assert.NotEqual(t, id, domain.CollapseID(key, user, may, domain.KindEarly))
	assert.NotEqual(t, id, domain.CollapseID([]byte(strings.Repeat("j", 32)), user, p, domain.KindEarly))
	assert.NotContains(t, id, user)
}
