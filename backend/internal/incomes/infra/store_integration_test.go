//go:build integration

package infra_test

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/incomes/domain"
	"github.com/romodevx/control-contable/backend/internal/incomes/infra"
	"github.com/romodevx/control-contable/backend/internal/incomes/usecase"
	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/access"
	"github.com/romodevx/control-contable/backend/internal/platform/access/accessgorm"
	"github.com/romodevx/control-contable/backend/internal/platform/clock"
	"github.com/romodevx/control-contable/backend/internal/platform/crypto"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
	"github.com/romodevx/control-contable/backend/internal/platform/pgdb"
	"github.com/romodevx/control-contable/backend/internal/platform/pgtest"
	taxpayersinfra "github.com/romodevx/control-contable/backend/internal/taxpayers/infra"
	taxpayers "github.com/romodevx/control-contable/backend/internal/taxpayers/usecase"
	"github.com/romodevx/control-contable/backend/migrations"
)

func TestMain(m *testing.M) {
	os.Exit(pgtest.Main(m, pgtest.WithMigrator(migrations.Up)))
}

// The store against the real migration, as cc_app under RLS. Taxpayers are
// created through the taxpayers use cases (test-only import). Fixture ids
// are this suite's own.
const (
	userA = "0199d4a2-0000-7000-8000-0000000004a1"
	userB = "0199d4a2-0000-7000-8000-0000000004b1"
)

var registerSQL = []string{
	`WITH v(id) AS (SELECT $1::uuid)
	 INSERT INTO users (id, email_hash, email_encrypted, email_key_id, password_hash, name)
	 SELECT id, sha256(convert_to(id::text, 'UTF8')),
	        '\x0001'::bytea || sha256(convert_to(id::text, 'UTF8')) || sha256(convert_to(id::text, 'UTF8')),
	        1, '$argon2id$v=19$m=65536,t=3,p=4$c2FsdHNhbHQ$aGFzaGhhc2hoYXNo', 'Ana'
	 FROM v`,
	"INSERT INTO user_usage (user_id) VALUES ($1)",
	"INSERT INTO entitlements (user_id, kind, source, starts_at, ends_at) VALUES ($1, 'trial', 'registration', now(), now() + interval '168 hours')",
}

type env struct {
	txm       *pgdb.TxManager
	store     infra.Store
	taxpayers *taxpayers.Taxpayers
}

func newEnv(t *testing.T) *env {
	t.Helper()
	app := pgtest.NewDatabase(t).Connect(t, pgtest.RoleApp)
	for _, u := range []string{userA, userB} {
		require.NoError(t, pgtest.AsUser(t.Context(), app, u, func(tx *sql.Tx) error {
			for _, statement := range registerSQL {
				if _, err := tx.ExecContext(t.Context(), statement, u); err != nil {
					return err
				}
			}
			return nil
		}))
	}
	txm, err := pgdb.NewTxManager(app)
	require.NoError(t, err)
	keyring, err := crypto.NewKeyring(1, map[crypto.KeyID][]byte{1: bytes.Repeat([]byte{0x5a}, crypto.KeySize)})
	require.NoError(t, err)
	index, err := crypto.NewRFCIndex(bytes.Repeat([]byte{0x6b}, crypto.MinPepperSize))
	require.NoError(t, err)
	store, catalog := taxpayersinfra.NewStore(), taxpayersinfra.NewCatalog()
	deps := taxpayers.Deps{
		Tx: txm, Access: access.NewPolicy(txm, accessgorm.NewStore(), accessgorm.NewStore(), clock.NewSystem()),
		Taxpayers: store, Regimes: store, Catalog: catalog, Directory: catalog, Records: taxpayersinfra.NoRecords{},
		RFC: taxpayersinfra.NewRFCVault(keyring, index), Clock: clock.NewSystem(), Logger: slog.New(slog.DiscardHandler),
	}
	return &env{txm: txm, store: infra.NewStore(), taxpayers: taxpayers.NewTaxpayers(deps)}
}

// taxpayer creates a taxpayer of user holding regime since January 2026.
func (e *env) taxpayer(t *testing.T, user, personType, rfc, regime string) string {
	t.Helper()
	created, err := e.taxpayers.Create(t.Context(), user, taxpayers.CreateInput{
		PersonType: personType, RFC: rfc, DisplayName: "Prueba",
		Regimes: []taxpayers.RegimeInput{{Code: regime, Start: "2026-01"}},
	})
	require.NoError(t, err)
	return created.Taxpayer.ID
}

// as runs work in a read-write unit of work of user.
func (e *env) as(t *testing.T, user string, work func(ctx context.Context) error) error {
	t.Helper()
	return e.txm.ReadWrite(t.Context(), user, work)
}

func day(t *testing.T, raw string) fiscal.Date {
	t.Helper()
	d, err := fiscal.ParseDate(raw)
	require.NoError(t, err)
	return d
}

func cents(c int64) money.Cents { return money.Cents(c) }

func id(n int) string { return fmt.Sprintf("0199d4a3-0000-7000-8000-%012d", n) }

func fields(t *testing.T, regime, collected string, amount int64) usecase.Fields {
	return usecase.Fields{
		Regime: fiscal.RegimeCode(regime), Counterparty: domain.CounterpartyPublic, Amount: cents(amount),
		IVARate: fiscal.IVAExempt, Dates: domain.Dates{CollectedOn: day(t, collected)},
	}
}

func TestStore_InsertFindUpdateAndSoftDelete(t *testing.T) {
	e := newEnv(t)
	pm := e.taxpayer(t, userA, "moral", "A&B020304XY2", "601")
	folio, err := fiscal.ParseCFDIUUID("6F3B2C1D-8A4E-4B7F-9C2D-1E5A7B9C3D4F")
	require.NoError(t, err)
	in := usecase.Fields{
		Regime: "601", Counterparty: domain.CounterpartyMoral, Amount: 45_000_00, IVARate: fiscal.IVARate16,
		IVA: 7_200_00, Withheld: fiscal.Withholdings{IVA: 100_00}, Dates: domain.Dates{InvoicedOn: day(t, "2026-09-05")},
		CFDIUUID: folio, Description: "Consultoría",
	}

	require.NoError(t, e.as(t, userA, func(ctx context.Context) error {
		if err := e.store.Insert(ctx, usecase.NewIncome{ID: id(1), UserID: userA, TaxpayerID: pm, Fields: in}); err != nil {
			return err
		}
		got, err := e.store.Find(ctx, id(1), usecase.LockUpdate)
		require.NoError(t, err)
		assert.Equal(t, in, got.Fields, "round trip, nulls included")
		assert.Equal(t, "2026-09-05", got.BookedOn.String(), "accrual: booked on invoiced_on by the trigger")
		assert.False(t, got.Deleted)
		assert.False(t, got.CreatedAt.IsZero())

		in.Dates.CollectedOn, in.Description, in.CFDIUUID = day(t, "2026-10-02"), "", fiscal.CFDIUUID{}
		require.NoError(t, e.store.Update(ctx, id(1), in))
		got, err = e.store.Find(ctx, id(1), usecase.LockNone)
		require.NoError(t, err)
		assert.Equal(t, in, got.Fields, "null clears")

		require.NoError(t, e.store.SetDeleted(ctx, id(1), true))
		got, err = e.store.Find(ctx, id(1), usecase.LockNone)
		require.NoError(t, err)
		assert.True(t, got.Deleted, "a deleted income is still found")
		require.NoError(t, e.store.SetDeleted(ctx, id(1), false))
		got, err = e.store.Find(ctx, id(1), usecase.LockNone)
		require.NoError(t, err)
		assert.False(t, got.Deleted)

		require.ErrorIs(t, e.store.Update(ctx, id(9), in), usecase.ErrIncomeNotFound)
		require.ErrorIs(t, e.store.SetDeleted(ctx, id(9), true), usecase.ErrIncomeNotFound)
		_, err = e.store.Find(ctx, id(9), usecase.LockNone)
		require.ErrorIs(t, err, usecase.ErrIncomeNotFound)
		return nil
	}))
}

// RLS: user B neither sees user A's income nor reuses its id (the primary
// key sees across users: ErrIDTaken).
func TestStore_Isolation(t *testing.T) {
	e := newEnv(t)
	tpA := e.taxpayer(t, userA, "fisica", "PEXJ900215HK8", "612")
	tpB := e.taxpayer(t, userB, "fisica", "PEXJ900215HK8", "612")
	require.NoError(t, e.as(t, userA, func(ctx context.Context) error {
		return e.store.Insert(ctx, usecase.NewIncome{ID: id(1), UserID: userA, TaxpayerID: tpA, Fields: fields(t, "612", "2026-09-10", 100)})
	}))

	err := e.as(t, userB, func(ctx context.Context) error {
		_, err := e.store.Find(ctx, id(1), usecase.LockNone)
		require.ErrorIs(t, err, usecase.ErrIncomeNotFound)
		require.ErrorIs(t, e.store.SetDeleted(ctx, id(1), true), usecase.ErrIncomeNotFound)
		return e.store.Insert(ctx, usecase.NewIncome{ID: id(1), UserID: userB, TaxpayerID: tpB, Fields: fields(t, "612", "2026-09-10", 100)})
	})
	require.ErrorIs(t, err, usecase.ErrIDTaken)

	err = e.as(t, userB, func(ctx context.Context) error {
		return e.store.Insert(ctx, usecase.NewIncome{ID: id(2), UserID: userB, TaxpayerID: tpA, Fields: fields(t, "612", "2026-09-10", 100)})
	})
	require.Error(t, err, "another user's taxpayer is refused (trigger / composite FK)")
	assert.NotErrorIs(t, err, usecase.ErrIDTaken)
}

func TestStore_ListTotalsAndFolios(t *testing.T) {
	e := newEnv(t)
	tp := e.taxpayer(t, userA, "fisica", "PEXJ900215HK8", "612")
	folio, err := fiscal.ParseCFDIUUID("6f3b2c1d-8a4e-4b7f-9c2d-1e5a7b9c3d4f")
	require.NoError(t, err)
	require.NoError(t, e.as(t, userA, func(ctx context.Context) error {
		for i, d := range []string{"2026-09-03", "2026-09-03", "2026-09-17", "2026-08-31", "2026-09-20"} {
			f := fields(t, "612", d, int64(100*(i+1)))
			if i == 0 {
				f.CFDIUUID = folio
			}
			if err := e.store.Insert(ctx, usecase.NewIncome{ID: id(i + 1), UserID: userA, TaxpayerID: tp, Fields: f}); err != nil {
				return err
			}
		}
		return e.store.SetDeleted(ctx, id(5), true)
	}))

	require.NoError(t, e.as(t, userA, func(ctx context.Context) error {
		september := usecase.ListFilter{TaxpayerID: tp, From: day(t, "2026-09-01"), Until: day(t, "2026-10-01"), Regime: "612"}
		first, err := e.store.List(ctx, september, usecase.Keyset{}, 2)
		require.NoError(t, err)
		require.Len(t, first, 2)
		assert.Equal(t, []string{id(3), id(2)}, []string{first[0].ID, first[1].ID})
		rest, err := e.store.List(ctx, september, usecase.Keyset{BookedOn: first[1].BookedOn, ID: first[1].ID}, 2)
		require.NoError(t, err)
		require.Len(t, rest, 1)
		assert.Equal(t, id(1), rest[0].ID)

		totals, err := e.store.Totals(ctx, september)
		require.NoError(t, err)
		assert.Equal(t, usecase.PeriodTotals{Count: 3, Amount: cents(100 + 200 + 300)}, totals)
		none, err := e.store.Totals(ctx, usecase.ListFilter{TaxpayerID: tp, From: day(t, "2026-09-01"), Until: day(t, "2026-10-01"), Regime: "606"})
		require.NoError(t, err)
		assert.Zero(t, none)

		inUse, err := e.store.CFDIUUIDInUse(ctx, tp, folio, id(9))
		require.NoError(t, err)
		assert.True(t, inUse)
		inUse, err = e.store.CFDIUUIDInUse(ctx, tp, folio, id(1))
		require.NoError(t, err)
		assert.False(t, inUse, "the income itself does not count")
		return nil
	}))
}

func TestStore_IncomeBasis(t *testing.T) {
	e := newEnv(t)
	require.NoError(t, e.as(t, userA, func(ctx context.Context) error {
		basis, err := e.store.IncomeBasis(ctx, fiscal.PersonMoral, "601")
		require.NoError(t, err)
		assert.Equal(t, domain.BasisAccrual, basis)
		basis, err = e.store.IncomeBasis(ctx, fiscal.PersonFisica, "612")
		require.NoError(t, err)
		assert.Equal(t, domain.BasisCash, basis)
		_, err = e.store.IncomeBasis(ctx, fiscal.PersonFisica, "601")
		require.Error(t, err, "601 is not a física regime")
		return nil
	}))
}

// Outside a unit of work there is no RLS context: every method fails closed.
func TestStore_OutsideAUnitOfWork(t *testing.T) {
	s, ctx := infra.NewStore(), context.Background()
	_, err := s.Find(ctx, id(1), usecase.LockNone)
	require.ErrorIs(t, err, pgdb.ErrNoTransaction)
	require.ErrorIs(t, s.Insert(ctx, usecase.NewIncome{}), pgdb.ErrNoTransaction)
	require.ErrorIs(t, s.Update(ctx, id(1), usecase.Fields{}), pgdb.ErrNoTransaction)
	_, err = s.CFDIUUIDInUse(ctx, "", fiscal.CFDIUUID{}, "")
	require.ErrorIs(t, err, pgdb.ErrNoTransaction)
	_, err = s.List(ctx, usecase.ListFilter{}, usecase.Keyset{}, 1)
	require.ErrorIs(t, err, pgdb.ErrNoTransaction)
	_, err = s.Totals(ctx, usecase.ListFilter{})
	require.ErrorIs(t, err, pgdb.ErrNoTransaction)
	_, err = s.IncomeBasis(ctx, fiscal.PersonFisica, "612")
	require.ErrorIs(t, err, pgdb.ErrNoTransaction)
}
