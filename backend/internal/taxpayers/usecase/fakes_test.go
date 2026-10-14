package usecase_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"log/slog"
	"maps"
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/platform/access"
	"github.com/romodevx/control-contable/backend/internal/platform/apperr"
	"github.com/romodevx/control-contable/backend/internal/platform/clock"
	"github.com/romodevx/control-contable/backend/internal/platform/tx"
	"github.com/romodevx/control-contable/backend/internal/taxpayers/domain"
	"github.com/romodevx/control-contable/backend/internal/taxpayers/usecase"
)

// Fixtures. Every RFC is synthetic (see domain/rfc_test.go): PEXJ900215HK8
// and A&B020304XY2 carry a valid SAT check digit; PEXJ900215HK9 does not.
const (
	userA = "0199d4a2-0000-7000-8000-00000000000a"
	userB = "0199d4a2-0000-7000-8000-00000000000b"

	rfcFisica         = "PEXJ900215HK8"
	rfcFisicaMasked   = "PEXJ******HK8"
	rfcFisicaMismatch = "PEXJ900215HK9"
	rfcFisicaOther    = "AAAA010101AA0"
	rfcMoral          = "A&B020304XY2"
)

// today is the frozen "now": 13 October 2026, noon in Mexico City.
var today = time.Date(2026, time.October, 13, 12, 0, 0, 0, clock.Location())

// userKey carries the fake unit of work's user, standing in for RLS.
type userKey struct{}

func currentUser(ctx context.Context) string {
	user, _ := ctx.Value(userKey{}).(string)
	return user
}

// fakeTx runs work inline with the user in ctx; it records the modes used.
type fakeTx struct {
	mu    sync.Mutex
	modes []string
}

func (f *fakeTx) ReadWrite(ctx context.Context, userID string, work tx.Work) error {
	return f.run(ctx, "rw", userID, work)
}

func (f *fakeTx) ReadOnly(ctx context.Context, userID string, work tx.Work) error {
	return f.run(ctx, "ro", userID, work)
}

func (f *fakeTx) run(ctx context.Context, mode, userID string, work tx.Work) error {
	if userID == "" {
		return tx.ErrNoUser
	}
	f.mu.Lock()
	f.modes = append(f.modes, mode)
	f.mu.Unlock()
	return work(context.WithValue(ctx, userKey{}, userID))
}

// fakeAccess is the access policy with a switchable read-only state and the
// real cap rule.
type fakeAccess struct {
	readOnly bool
	checks   int
}

func (f *fakeAccess) EnsureWritable(context.Context, string) error {
	f.checks++
	if f.readOnly {
		return apperr.Wrap(access.ErrReadOnly, apperr.CodeEntitlementRequired, "read only")
	}
	return nil
}

func (f *fakeAccess) CheckTaxpayerLimit(ctx context.Context, userID string, existing int) error {
	return new(access.Policy).CheckTaxpayerLimit(ctx, userID, existing)
}

// sealCall is one Seal invocation, in call order with the store's inserts.
type sealCall struct {
	binding usecase.RowBinding
	rfc     string
}

// fakeProtector seals deterministically and records every binding.
type fakeProtector struct {
	calls []sealCall
	log   *[]string
}

func (f *fakeProtector) Seal(rfc domain.RFC, b usecase.RowBinding) (usecase.SealedRFC, error) {
	f.calls = append(f.calls, sealCall{binding: b, rfc: rfc.String()})
	if f.log != nil {
		*f.log = append(*f.log, "seal:"+b.TaxpayerID)
	}
	hash := sha256.Sum256([]byte(b.UserID + "|" + rfc.String()))
	return usecase.SealedRFC{
		Ciphertext:      []byte("ct:" + b.TaxpayerID + ":" + b.UserID),
		KeyID:           1,
		Hash:            hash[:],
		Masked:          rfc.Masked(),
		CheckDigitValid: rfc.CheckDigitValid(),
	}, nil
}

// storedRow is one fake taxpayers row.
type storedRow struct {
	userID string
	t      usecase.StoredTaxpayer
	rfc    usecase.SealedRFC
}

// regimeRow is one fake taxpayer_regimes row.
type regimeRow struct {
	userID string
	held   usecase.HeldRegime
	ack    bool
}

// memStore implements TaxpayerStore and RegimeStore in memory, filtering by
// the fake unit of work's user like RLS does.
type memStore struct {
	catalog  *fakeCatalog
	rows     map[string]*storedRow
	regimes  map[string]*regimeRow // taxpayer id + "/" + code
	created  time.Time
	locks    []string
	log      *[]string
	listRuns int
	regRuns  int
	failNext error
}

func newMemStore(catalog *fakeCatalog) *memStore {
	return &memStore{catalog: catalog, rows: map[string]*storedRow{}, regimes: map[string]*regimeRow{}, created: today}
}

func (s *memStore) own(ctx context.Context, id string) (*storedRow, bool) {
	row, ok := s.rows[id]
	if !ok || row.userID != currentUser(ctx) {
		return nil, false
	}
	return row, true
}

func (s *memStore) CountForCreate(ctx context.Context) (int, error) {
	s.locks = append(s.locks, "user:update")
	n := 0
	for _, row := range s.rows {
		if row.userID == currentUser(ctx) {
			n++
		}
	}
	return n, nil
}

func (s *memStore) Insert(ctx context.Context, t usecase.NewTaxpayer) error {
	if s.log != nil {
		*s.log = append(*s.log, "insert:"+t.ID)
	}
	if t.UserID != currentUser(ctx) {
		return errors.New("rls: user_id is not the context user")
	}
	if _, taken := s.rows[t.ID]; taken {
		return usecase.ErrIDConflict
	}
	for _, row := range s.rows {
		if row.userID == t.UserID && bytes.Equal(row.rfc.Hash, t.RFC.Hash) {
			return usecase.ErrRFCAlreadyRegistered
		}
	}
	s.created = s.created.Add(time.Second)
	s.rows[t.ID] = &storedRow{userID: t.UserID, rfc: t.RFC, t: usecase.StoredTaxpayer{
		ID: t.ID, PersonType: t.PersonType, DisplayName: t.DisplayName.String(),
		RFCMasked: t.RFC.Masked, CreatedAt: s.created,
	}}
	return nil
}

func (s *memStore) Find(ctx context.Context, id string, lock usecase.Lock) (usecase.StoredTaxpayer, error) {
	s.locks = append(s.locks, fmt.Sprintf("taxpayer:%d", lock))
	row, ok := s.own(ctx, id)
	if !ok {
		return usecase.StoredTaxpayer{}, usecase.ErrTaxpayerNotFound
	}
	return row.t, nil
}

func (s *memStore) List(ctx context.Context, includeArchived bool) ([]usecase.StoredTaxpayer, error) {
	s.listRuns++
	var out []usecase.StoredTaxpayer
	for _, row := range s.rows {
		if row.userID == currentUser(ctx) && (includeArchived || !row.t.Archived()) {
			out = append(out, row.t)
		}
	}
	slices.SortFunc(out, func(a, b usecase.StoredTaxpayer) int { return a.CreatedAt.Compare(b.CreatedAt) })
	return out, nil
}

func (s *memStore) Rename(ctx context.Context, id string, name domain.DisplayName) error {
	row, ok := s.own(ctx, id)
	if !ok {
		return usecase.ErrTaxpayerNotFound
	}
	row.t.DisplayName = name.String()
	return nil
}

func (s *memStore) ReplaceRFC(ctx context.Context, id string, rfc usecase.SealedRFC) error {
	row, ok := s.own(ctx, id)
	if !ok {
		return usecase.ErrTaxpayerNotFound
	}
	for otherID, other := range s.rows {
		if otherID != id && other.userID == row.userID && bytes.Equal(other.rfc.Hash, rfc.Hash) {
			return usecase.ErrRFCAlreadyRegistered
		}
	}
	row.rfc = rfc
	row.t.RFCMasked = rfc.Masked
	return nil
}

func (s *memStore) SetArchivedAt(ctx context.Context, id string, at time.Time) error {
	row, ok := s.own(ctx, id)
	if !ok {
		return usecase.ErrTaxpayerNotFound
	}
	row.t.ArchivedAt = at
	return nil
}

func (s *memStore) Delete(ctx context.Context, id string) (bool, error) {
	if _, ok := s.own(ctx, id); !ok {
		return false, nil
	}
	delete(s.rows, id)
	for key, r := range s.regimes {
		if r.held.TaxpayerID == id {
			delete(s.regimes, key)
		}
	}
	return true, nil
}

func (s *memStore) Regimes(ctx context.Context, taxpayerIDs ...string) ([]usecase.HeldRegime, error) {
	s.regRuns++
	var out []usecase.HeldRegime
	for _, r := range s.regimes {
		if r.userID == currentUser(ctx) && slices.Contains(taxpayerIDs, r.held.TaxpayerID) {
			out = append(out, cloneHeld(r.held))
		}
	}
	slices.SortFunc(out, func(a, b usecase.HeldRegime) int {
		if c := a.Window.Start.Compare(b.Window.Start); c != 0 {
			return c
		}
		if a.Code < b.Code {
			return -1
		}
		return 1
	})
	return out, nil
}

func (s *memStore) FindRegime(ctx context.Context, taxpayerID, code string, lock usecase.Lock) (usecase.HeldRegime, error) {
	s.locks = append(s.locks, fmt.Sprintf("regime:%d", lock))
	r, ok := s.regimes[taxpayerID+"/"+code]
	if !ok || r.userID != currentUser(ctx) {
		return usecase.HeldRegime{}, usecase.ErrRegimeNotHeld
	}
	return cloneHeld(r.held), nil
}

func (s *memStore) InsertRegime(ctx context.Context, r usecase.NewRegime) error {
	if s.failNext != nil {
		err := s.failNext
		s.failNext = nil
		return err
	}
	key := r.TaxpayerID + "/" + r.Code
	if _, taken := s.regimes[key]; taken {
		return usecase.ErrRegimeAlreadyActive
	}
	cat, err := s.catalog.Regime(ctx, r.PersonType, r.Code)
	if err != nil {
		return err
	}
	s.regimes[key] = &regimeRow{userID: r.UserID, ack: r.NeedsCPAReviewAck, held: usecase.HeldRegime{
		TaxpayerID: r.TaxpayerID, Code: r.Code, Name: cat.Name, Treatment: cat.Treatment,
		Options: maps.Clone(r.Options), Window: domain.Window{Start: r.Start},
	}}
	return nil
}

func (s *memStore) ReopenRegime(ctx context.Context, taxpayerID, code string, options map[string]domain.OptionValue, ack bool) error {
	r, ok := s.regimes[taxpayerID+"/"+code]
	if !ok || r.userID != currentUser(ctx) || r.held.Window.IsOpen() {
		return usecase.ErrRegimeNotHeld
	}
	r.held.Window.End = domain.Period{}
	r.held.Options = maps.Clone(options)
	r.ack = ack
	return nil
}

func (s *memStore) UpdateOptions(ctx context.Context, taxpayerID, code string, options map[string]domain.OptionValue) error {
	r, ok := s.regimes[taxpayerID+"/"+code]
	if !ok || r.userID != currentUser(ctx) {
		return usecase.ErrRegimeNotHeld
	}
	r.held.Options = maps.Clone(options)
	return nil
}

func (s *memStore) EndRegime(ctx context.Context, taxpayerID, code string, end domain.Period) error {
	r, ok := s.regimes[taxpayerID+"/"+code]
	if !ok || r.userID != currentUser(ctx) {
		return usecase.ErrRegimeNotHeld
	}
	r.held.Window.End = end
	return nil
}

func cloneHeld(h usecase.HeldRegime) usecase.HeldRegime {
	h.Options = maps.Clone(h.Options)
	return h
}

// seedTaxpayer writes a taxpayer of userID directly, with its regimes.
func (s *memStore) seedTaxpayer(userID, id string, personType domain.PersonType, archived bool, regimes ...seedRegime) {
	s.created = s.created.Add(time.Second)
	row := &storedRow{userID: userID, rfc: usecase.SealedRFC{Hash: []byte("hash-" + id)}, t: usecase.StoredTaxpayer{
		ID: id, PersonType: personType, DisplayName: "Seed " + id[len(id)-2:], RFCMasked: rfcFisicaMasked, CreatedAt: s.created,
	}}
	if archived {
		row.t.ArchivedAt = today.Add(-time.Hour)
	}
	s.rows[id] = row
	for _, r := range regimes {
		s.regimes[id+"/"+r.code] = &regimeRow{userID: userID, held: usecase.HeldRegime{
			TaxpayerID: id, Code: r.code, Name: "Regime " + r.code, Treatment: "full",
			Options: maps.Clone(r.options), Window: r.window,
		}}
	}
}

type seedRegime struct {
	code    string
	window  domain.Window
	options map[string]domain.OptionValue
}

// fakeCatalog is a slice of the M2-02 seed.
type fakeCatalog struct {
	regimes map[string]usecase.CatalogRegime // person type + "/" + code
	pairs   map[string]usecase.Compatibility // person type + "/" + a + "/" + b
}

func newFakeCatalog(t *testing.T) *fakeCatalog {
	t.Helper()
	c := &fakeCatalog{regimes: map[string]usecase.CatalogRegime{}, pairs: map[string]usecase.Compatibility{}}
	fisica, moral := domain.PersonTypeFisica, domain.PersonTypeMoral
	c.add(t, fisica, "606", domain.OptionDef{Key: "deduction_mode", Type: domain.OptionTypeEnum, Required: true,
		AllowedValues: []string{"blind_35", "actual"}, Default: domain.StringOption("blind_35")})
	c.add(t, fisica, "612", domain.OptionDef{Key: "iva_fully_taxable", Type: domain.OptionTypeBoolean, Required: true,
		Default: domain.BoolOption(true)})
	c.add(t, fisica, "625",
		domain.OptionDef{Key: "activities", Type: domain.OptionTypeMultiEnum, Required: true,
			AllowedValues: []string{"transport_delivery", "lodging", "goods_services"}},
		domain.OptionDef{Key: "definitive_option", Type: domain.OptionTypeBoolean, Required: true, Default: domain.BoolOption(false)},
		domain.OptionDef{Key: "rfc_given_to_platform", Type: domain.OptionTypeBoolean, Required: true, Default: domain.BoolOption(true)})
	c.add(t, fisica, "621", domain.OptionDef{Key: "rif_year", Type: domain.OptionTypeEnum, Required: true,
		AllowedValues: []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10"}})
	c.add(t, fisica, "626")
	c.add(t, fisica, "605")
	c.add(t, fisica, "607")
	c.add(t, moral, "601", domain.OptionDef{Key: "iva_fully_taxable", Type: domain.OptionTypeBoolean, Required: true,
		Default: domain.BoolOption(true)})
	c.add(t, moral, "626")
	c.pair(fisica, "612", "626", usecase.Compatibility{Reason: "El RESICO sustituye al 612."})
	c.pair(fisica, "606", "626", usecase.Compatibility{Reason: "El RESICO cubre el arrendamiento."})
	c.pair(fisica, "605", "607", usecase.Compatibility{Allowed: true, NeedsCPAReview: true, Reason: "Pendiente de revisión por un contador."})
	c.pair(fisica, "606", "612", usecase.Compatibility{Allowed: true, Reason: "Coexisten."})
	c.pair(moral, "601", "626", usecase.Compatibility{Reason: "Una persona moral tiene un solo régimen."})
	return c
}

func (c *fakeCatalog) add(t *testing.T, pt domain.PersonType, code string, defs ...domain.OptionDef) {
	t.Helper()
	set, err := domain.NewOptionSet(defs...)
	require.NoError(t, err)
	c.regimes[string(pt)+"/"+code] = usecase.CatalogRegime{Code: code, PersonType: pt, Name: "Regime " + code, Treatment: "full", Options: set}
}

func (c *fakeCatalog) pair(pt domain.PersonType, a, b string, verdict usecase.Compatibility) {
	c.pairs[string(pt)+"/"+a+"/"+b] = verdict
}

func (c *fakeCatalog) Regime(_ context.Context, pt domain.PersonType, code string) (usecase.CatalogRegime, error) {
	r, ok := c.regimes[string(pt)+"/"+code]
	if !ok {
		return usecase.CatalogRegime{}, usecase.ErrRegimeNotForPersonType
	}
	return r, nil
}

func (c *fakeCatalog) Compatibility(_ context.Context, pt domain.PersonType, codes []string) (map[usecase.RegimePair]usecase.Compatibility, error) {
	out := map[usecase.RegimePair]usecase.Compatibility{}
	for _, x := range codes {
		for _, y := range codes {
			pair := usecase.PairOf(x, y)
			if verdict, ok := c.pairs[string(pt)+"/"+pair.A+"/"+pair.B]; ok {
				out[pair] = verdict
			}
		}
	}
	return out, nil
}

// fakeRecords reports a latest record period per taxpayer/code.
type fakeRecords map[string]domain.Period

func (f fakeRecords) LatestRecordPeriod(_ context.Context, taxpayerID, code string) (domain.Period, error) {
	return f[taxpayerID+"/"+code], nil
}

// harness wires the use cases over the fakes.
type harness struct {
	tx        *fakeTx
	access    *fakeAccess
	store     *memStore
	catalog   *fakeCatalog
	rfc       *fakeProtector
	records   fakeRecords
	logs      *bytes.Buffer
	order     []string
	ids       []string
	taxpayers *usecase.Taxpayers
	regimes   *usecase.Regimes
	queries   *usecase.Queries
}

func newHarness(t *testing.T) *harness {
	t.Helper()
	h := &harness{tx: &fakeTx{}, access: &fakeAccess{}, records: fakeRecords{}, logs: &bytes.Buffer{}}
	h.catalog = newFakeCatalog(t)
	h.store = newMemStore(h.catalog)
	h.store.log = &h.order
	h.rfc = &fakeProtector{log: &h.order}
	next := 0
	deps := usecase.Deps{
		Tx: h.tx, Access: h.access, Taxpayers: h.store, Regimes: h.store, Catalog: h.catalog,
		Records: h.records, RFC: h.rfc, Clock: clock.NewFixed(today),
		NewID: func() (string, error) {
			next++
			id := fmt.Sprintf("0199d4a2-0000-7000-8000-%012d", next)
			h.ids = append(h.ids, id)
			h.order = append(h.order, "id:"+id)
			return id, nil
		},
		Logger: slog.New(slog.NewJSONHandler(h.logs, &slog.HandlerOptions{Level: slog.LevelDebug})),
	}
	h.taxpayers = usecase.NewTaxpayers(deps)
	h.regimes = usecase.NewRegimes(deps)
	h.queries = usecase.NewQueries(deps)
	return h
}

// requireAppErr asserts err is an *apperr.Error with code and the given
// field/detail pairs (in order), and that it wraps target. It returns the
// client message.
func requireAppErr(t *testing.T, err error, code apperr.Code, target error, fields ...[2]string) string {
	t.Helper()
	require.Error(t, err)
	app, ok := errors.AsType[*apperr.Error](err)
	require.True(t, ok, "want *apperr.Error, got %T: %v", err, err)
	require.Equal(t, code, app.Code, "error: %v", err)
	if target != nil {
		require.ErrorIs(t, err, target)
	}
	got := make([][2]string, 0, len(app.Fields))
	for _, f := range app.Fields {
		got = append(got, [2]string{f.Field, f.Code})
	}
	if len(fields) > 0 {
		require.Equal(t, fields, got)
	}
	return app.Message
}

func period(t *testing.T, raw string) domain.Period {
	t.Helper()
	p, err := domain.ParsePeriod(raw)
	require.NoError(t, err)
	return p
}

func window(t *testing.T, start, end string) domain.Window {
	t.Helper()
	w := domain.Window{Start: period(t, start)}
	if end != "" {
		w.End = period(t, end)
	}
	return w
}
