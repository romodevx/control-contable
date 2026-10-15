package usecase_test

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/romodevx/control-contable/backend/internal/history/domain"
	"github.com/romodevx/control-contable/backend/internal/history/usecase"
	"github.com/romodevx/control-contable/backend/internal/platform/apperr"
	"github.com/romodevx/control-contable/backend/internal/platform/clock"
	"github.com/romodevx/control-contable/backend/internal/platform/tx"
)

const (
	userA     = "0199d4a2-0000-7000-8000-00000000e00a"
	taxpayerA = "0199d4a2-1111-7000-8000-00000000e00a"
)

// today is 14 October 2026, noon in Mexico City.
var today = time.Date(2026, time.October, 14, 12, 0, 0, 0, clock.Location())

var errPortDown = errors.New("fake: port down")

// userKey carries the fake unit of work's user, standing in for RLS.
type userKey struct{}

func userOf(ctx context.Context) string {
	user, _ := ctx.Value(userKey{}).(string)
	return user
}

// fakeTx runs work inline and records the modes used.
type fakeTx struct{ modes []string }

func (f *fakeTx) ReadWrite(ctx context.Context, userID string, work tx.Work) error {
	f.modes = append(f.modes, "rw")
	return work(context.WithValue(ctx, userKey{}, userID))
}

func (f *fakeTx) ReadOnly(ctx context.Context, userID string, work tx.Work) error {
	f.modes = append(f.modes, "ro")
	return work(context.WithValue(ctx, userKey{}, userID))
}

// fakeFlags answers the modules ON in it.
type fakeFlags map[string]bool

func (f fakeFlags) Enabled(_ context.Context, module string) bool { return f[module] }

func hex64(c string) string { return strings.Repeat(c, 64) }

// seenOf is a month's estimate fingerprints: the figures hash names the
// month and a figures tag, so tests move it on purpose.
func seenOf(m domain.Month, figures string) domain.Seen {
	return domain.Seen{
		RecordsFingerprint: hex64("a"), FiguresFingerprint: hex64(figures),
		EngineVersion: "1.0.1", ParameterSet: "2026/v1",
	}
}

// fakeEstimator answers one MonthEstimate per month of [from, to], from
// byMonth or a complete 612 month with 100 of ISR.
type fakeEstimator struct {
	byMonth map[string]usecase.MonthEstimate
	err     error
	short   bool // answers one month fewer (a wiring bug)
	calls   []string
}

func (f *fakeEstimator) EstimateMonths(ctx context.Context, userID, taxpayerID string, from, to domain.Month) (usecase.Estimates, error) {
	f.calls = append(f.calls, userOf(ctx)+"|"+userID+"|"+taxpayerID+"|"+from.String()+".."+to.String())
	if f.err != nil {
		return usecase.Estimates{}, f.err
	}
	var out usecase.Estimates
	for m := from; !to.Before(m); m = domain.NewMonth(m.Year(), m.Month()+1) {
		est, ok := f.byMonth[m.String()]
		if !ok {
			est = usecase.MonthEstimate{
				Regimes: []domain.Status{domain.StatusComplete},
				Tax:     domain.Tax{ISRPayable: 100, IVAOutcome: "zero", TotalPayable: 100},
				DueDate: "2026-01-17", Seen: seenOf(m, "b"), ParameterSet: domain.ParameterSet{FiscalYear: 2026, Version: 1},
				Disclaimers: []domain.Disclaimer{{Code: "not_tax_advice", Severity: domain.SeverityInfo, Text: "n"}},
			}
		}
		est.Month = m
		out.Months = append(out.Months, est)
		if m.Month() == time.December {
			break
		}
	}
	if f.short {
		out.Months = out.Months[1:]
	}
	out.Legend = "legend"
	return out, nil
}

// fakeViews is estimate_views for one user.
type fakeViews struct {
	views   []domain.View
	err     error
	saveErr error
	saved   []string
	reads   []string
}

func (f *fakeViews) Views(ctx context.Context, taxpayerID string, from, to domain.Month) ([]domain.View, error) {
	f.reads = append(f.reads, userOf(ctx)+"|"+taxpayerID+"|"+from.String()+".."+to.String())
	return f.views, f.err
}

func (f *fakeViews) SaveView(ctx context.Context, userID, taxpayerID string, v domain.View) error {
	f.saved = append(f.saved, strings.Join([]string{
		userOf(ctx), userID, taxpayerID, v.Month.String(), v.Seen.RecordsFingerprint[:1], v.Seen.FiguresFingerprint[:1],
		v.Seen.EngineVersion, v.Seen.ParameterSet, v.ViewedAt.Format(time.RFC3339),
	}, "|"))
	return f.saveErr
}

type fixture struct {
	tx        *fakeTx
	flags     fakeFlags
	estimator *fakeEstimator
	views     *fakeViews
}

func newFixture() *fixture {
	return &fixture{
		tx:        &fakeTx{},
		flags:     fakeFlags{usecase.ModuleIncomes: true, usecase.ModuleExpenses: true},
		estimator: &fakeEstimator{byMonth: map[string]usecase.MonthEstimate{}},
		views:     &fakeViews{},
	}
}

func (f *fixture) service() *usecase.Service {
	svc, err := usecase.NewService(usecase.Deps{
		Tx: f.tx, Flags: f.flags, Estimator: f.estimator, Views: f.views, Clock: clock.NewFixed(today),
	})
	if err != nil {
		panic(err)
	}
	return svc
}

func fixedClock(instant time.Time) clock.Clock { return clock.NewFixed(instant) }

// notFoundPort is the taxpayers module's 404 as a port returns it.
var notFoundPort = apperr.New(apperr.CodeNotFound, "Taxpayer not found.")
