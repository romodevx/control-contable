package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/romodevx/control-contable/backend/internal/platform/apperr"
	"github.com/romodevx/control-contable/backend/internal/taxpayers/domain"
)

// Audit events (S-9).
const (
	eventCreated    = "taxpayer.created"
	eventArchived   = "taxpayer.archived"
	eventUnarchived = "taxpayer.unarchived"
	eventDeleted    = "taxpayer.deleted"
)

// Taxpayers manages a user's taxpayers (US-T1..T3c).
type Taxpayers struct {
	service
}

// NewTaxpayers returns the use case over deps.
func NewTaxpayers(deps Deps) *Taxpayers { return &Taxpayers{service: newService(deps)} }

// Create onboards a taxpayer with its regimes and options in one
// transaction (createTaxpayer). Order: idempotent retry (a client id the
// caller already owns answers it unchanged, Created false, whatever the
// access level or the cap) → access policy (402) → cap under the user lock
// (422) → validation (400) → regime compatibility (422) → writes (409 for a
// duplicate RFC or a foreign client id).
func (uc *Taxpayers) Create(ctx context.Context, userID string, in CreateInput) (CreateResult, error) {
	id, clientID, err := uc.rowID(in.ID)
	if err != nil {
		return CreateResult{}, present("create taxpayer", err)
	}
	var result CreateResult
	err = uc.deps.Tx.ReadWrite(ctx, userID, func(ctx context.Context) error {
		var err error
		result, err = uc.create(ctx, userID, id, clientID, in)
		return err
	})
	if err != nil {
		return CreateResult{}, present("create taxpayer", err)
	}
	if result.Created {
		uc.audit(ctx, eventCreated, userID, id)
	}
	return result, nil
}

// rowID returns the canonical client id, or a generated uuidv7 (S-1: the
// id exists before the RFC is sealed).
func (uc *Taxpayers) rowID(raw string) (id string, clientID bool, err error) {
	if raw == "" {
		id, err = uc.deps.NewID()
		return id, false, err
	}
	id, ok := canonicalID(raw)
	if !ok {
		fields := &fieldErrors{}
		fields.add(fieldID, DetailInvalid, msgInvalidField, ErrInvalidID)
		return "", false, fields.err()
	}
	return id, true, nil
}

func (uc *Taxpayers) create(ctx context.Context, userID, id string, clientID bool, in CreateInput) (CreateResult, error) {
	if existing, found, err := uc.existing(ctx, id, clientID); err != nil || found {
		return existing, err
	}
	if err := uc.deps.Access.EnsureWritable(ctx, userID); err != nil {
		return CreateResult{}, err
	}
	count, err := uc.deps.Taxpayers.CountForCreate(ctx)
	if err != nil {
		return CreateResult{}, err
	}
	// The lock serializes this user's creates: a concurrent retry of the
	// same client id has committed by now, so look again.
	if existing, found, err := uc.existing(ctx, id, clientID); err != nil || found {
		return existing, err
	}
	if err := uc.deps.Access.CheckTaxpayerLimit(ctx, userID, count); err != nil {
		return CreateResult{}, err
	}
	draft, err := uc.validateCreate(ctx, in)
	if err != nil {
		return CreateResult{}, err
	}
	if err := uc.insert(ctx, userID, id, draft); err != nil {
		return CreateResult{}, err
	}
	created, err := uc.load(ctx, id)
	if err != nil {
		return CreateResult{}, err
	}
	created.Warnings = append(rfcWarnings(draft.rfc), draft.warnings...)
	return CreateResult{Taxpayer: created, Created: true}, nil
}

// existing returns the caller's taxpayer with a client id, if any.
func (uc *Taxpayers) existing(ctx context.Context, id string, clientID bool) (CreateResult, bool, error) {
	if !clientID {
		return CreateResult{}, false, nil
	}
	t, err := uc.load(ctx, id)
	if errors.Is(err, ErrTaxpayerNotFound) {
		return CreateResult{}, false, nil
	}
	if err != nil {
		return CreateResult{}, false, err
	}
	return CreateResult{Taxpayer: t}, true, nil
}

// createDraft is a validated create.
type createDraft struct {
	personType domain.PersonType
	name       domain.DisplayName
	rfc        domain.RFC
	regimes    []NewRegime
	warnings   []Warning
}

func (uc *Taxpayers) validateCreate(ctx context.Context, in CreateInput) (createDraft, error) {
	fields := &fieldErrors{}
	var draft createDraft
	var err error
	if draft.name, err = domain.NewDisplayName(in.DisplayName); err != nil {
		fields.add(fieldDisplayName, DetailInvalid, msgInvalidField, err)
	}
	if draft.personType, err = domain.ParsePersonType(in.PersonType); err != nil {
		fields.add(fieldPersonType, DetailInvalid, msgInvalidField, err)
		return createDraft{}, fields.err() // RFC length and catalog depend on it
	}
	if draft.rfc, err = parseRFC(draft.personType, in.RFC); err != nil {
		fields.addRFC(err)
	}
	if draft.regimes, err = uc.validateRegimes(ctx, draft.personType, in.Regimes, fields); err != nil {
		return createDraft{}, err
	}
	if err := fields.err(); err != nil {
		return createDraft{}, err
	}
	draft.warnings, err = uc.checkBodyCompatibility(ctx, draft.personType, draft.regimes)
	return draft, err
}

// validateRegimes checks every regime of a create, recording field errors;
// it returns an error only for an infrastructure failure.
func (uc *Taxpayers) validateRegimes(ctx context.Context, personType domain.PersonType, inputs []RegimeInput, fields *fieldErrors) ([]NewRegime, error) {
	if len(inputs) == 0 {
		fields.add(fieldRegimesPrefix, DetailRequired, msgRequired, ErrNoRegimes)
		return nil, nil
	}
	regimes := make([]NewRegime, 0, len(inputs))
	seen := make(map[string]bool, len(inputs))
	for i, in := range inputs {
		prefix := fmt.Sprintf("%s.%d.", fieldRegimesPrefix, i)
		if seen[in.Code] {
			fields.add(prefix+fieldRegimeCode, DetailDuplicate, msgDuplicate, ErrDuplicateRegime)
			continue
		}
		seen[in.Code] = true
		regime, ok, err := uc.validateRegime(ctx, personType, in, prefix, fields)
		if err != nil {
			return nil, err
		}
		if ok {
			regimes = append(regimes, regime)
		}
	}
	return regimes, nil
}

// validateRegime checks one regime input (code for the person type,
// options, start); field names start with prefix.
func (s service) validateRegime(ctx context.Context, personType domain.PersonType, in RegimeInput, prefix string, fields *fieldErrors) (NewRegime, bool, error) {
	before := len(fields.fields)
	regime := NewRegime{PersonType: personType, Code: in.Code}
	catalog, err := s.deps.Catalog.Regime(ctx, personType, in.Code)
	switch {
	case errors.Is(err, ErrRegimeNotForPersonType):
		fields.add(prefix+fieldRegimeCode, DetailNotForPersonType, msgNotForPersonType, err)
	case err != nil:
		return NewRegime{}, false, err
	default:
		if regime.Options, err = catalog.Options.Validate(in.Options); err != nil {
			if err := fields.addOptions(prefix+fieldOptions, err); err != nil {
				return NewRegime{}, false, err
			}
		}
	}
	start, detail, err := s.startPeriod(in.Start)
	if err != nil {
		fields.addPeriod(prefix+fieldStartPeriod, detail, err)
	}
	regime.Start = start
	return regime, len(fields.fields) == before, nil
}

// checkBodyCompatibility checks every pair of the create's regimes: an
// overlapping pair the catalog refuses is 422 naming the later one; one
// under CPA review is allowed with a warning on the later one, which also
// records the acknowledgement.
func (uc *Taxpayers) checkBodyCompatibility(ctx context.Context, personType domain.PersonType, regimes []NewRegime) ([]Warning, error) {
	codes := make([]string, 0, len(regimes))
	for _, r := range regimes {
		codes = append(codes, r.Code)
	}
	matrix, err := uc.deps.Catalog.Compatibility(ctx, personType, codes)
	if err != nil {
		return nil, err
	}
	var warnings []Warning
	for j := range regimes {
		field := fmt.Sprintf("%s.%d.%s", fieldRegimesPrefix, j, fieldRegimeCode)
		for i := range j {
			if !openFrom(regimes[i].Start).Overlaps(openFrom(regimes[j].Start)) {
				continue
			}
			warning, err := verdict(matrix, field, regimes[j].Code, regimes[i].Code)
			if err != nil {
				return nil, err
			}
			if warning != nil {
				warnings = append(warnings, *warning)
				regimes[j].NeedsCPAReviewAck = true
			}
		}
	}
	return warnings, nil
}

// verdict applies the catalog verdict of an overlapping pair: an error when
// refused, a warning when under CPA review, nothing otherwise.
func verdict(matrix map[RegimePair]Compatibility, field, code, other string) (*Warning, error) {
	v, ok := matrix[PairOf(code, other)]
	switch {
	case !ok || (v.Allowed && !v.NeedsCPAReview):
		return nil, nil
	case !v.Allowed:
		return nil, incompatibleError(field, code, other, v.Reason)
	default:
		return &Warning{Code: WarningRegimeNeedsCPAReview, Field: field, Message: v.Reason}, nil
	}
}

func openFrom(start domain.Period) domain.Window { return domain.Window{Start: start} }

// insert seals the RFC for the row id, then writes the taxpayer and its
// regimes.
func (uc *Taxpayers) insert(ctx context.Context, userID, id string, draft createDraft) error {
	sealed, err := uc.deps.RFC.Seal(draft.rfc, RowBinding{TaxpayerID: id, UserID: userID})
	if err != nil {
		return err
	}
	err = uc.deps.Taxpayers.Insert(ctx, NewTaxpayer{
		ID: id, UserID: userID, PersonType: draft.personType, DisplayName: draft.name, RFC: sealed,
	})
	if err != nil {
		return err
	}
	for _, regime := range draft.regimes {
		regime.TaxpayerID, regime.UserID = id, userID
		if err := uc.deps.Regimes.InsertRegime(ctx, regime); err != nil {
			return err
		}
	}
	return nil
}

// List returns the user's taxpayers by creation, archived ones only when
// includeArchived (listTaxpayers). Regimes load in one query.
func (uc *Taxpayers) List(ctx context.Context, userID string, includeArchived bool) ([]Taxpayer, error) {
	var out []Taxpayer
	err := uc.deps.Tx.ReadOnly(ctx, userID, func(ctx context.Context) error {
		stored, err := uc.deps.Taxpayers.List(ctx, includeArchived)
		if err != nil || len(stored) == 0 {
			return err
		}
		ids := make([]string, 0, len(stored))
		for _, t := range stored {
			ids = append(ids, t.ID)
		}
		held, err := uc.deps.Regimes.Regimes(ctx, ids...)
		if err != nil {
			return err
		}
		out = groupRegimes(stored, held)
		return nil
	})
	if err != nil {
		return nil, present("list taxpayers", err)
	}
	return out, nil
}

func groupRegimes(stored []StoredTaxpayer, held []HeldRegime) []Taxpayer {
	byTaxpayer := make(map[string][]HeldRegime, len(stored))
	for _, h := range held {
		byTaxpayer[h.TaxpayerID] = append(byTaxpayer[h.TaxpayerID], h)
	}
	out := make([]Taxpayer, 0, len(stored))
	for _, t := range stored {
		out = append(out, toTaxpayer(t, byTaxpayer[t.ID]))
	}
	return out
}

// Get returns one taxpayer, archived or not (getTaxpayer).
func (uc *Taxpayers) Get(ctx context.Context, userID, rawID string) (Taxpayer, error) {
	id, ok := canonicalID(rawID)
	if !ok {
		return Taxpayer{}, notFound()
	}
	var out Taxpayer
	err := uc.deps.Tx.ReadOnly(ctx, userID, func(ctx context.Context) error {
		var err error
		out, err = uc.load(ctx, id)
		return err
	})
	if err != nil {
		return Taxpayer{}, present("get taxpayer", err)
	}
	return out, nil
}

// Update renames a taxpayer and/or corrects its RFC (updateTaxpayer, OQ-1):
// the corrected RFC is re-validated for the immutable person type, sealed
// again for the same row id and re-hashed.
func (uc *Taxpayers) Update(ctx context.Context, userID, rawID string, in UpdateInput) (Taxpayer, error) {
	id, ok := canonicalID(rawID)
	if !ok {
		return Taxpayer{}, notFound()
	}
	var out Taxpayer
	err := uc.deps.Tx.ReadWrite(ctx, userID, func(ctx context.Context) error {
		var err error
		out, err = uc.update(ctx, userID, id, in)
		return err
	})
	if err != nil {
		return Taxpayer{}, present("update taxpayer", err)
	}
	return out, nil
}

func (uc *Taxpayers) update(ctx context.Context, userID, id string, in UpdateInput) (Taxpayer, error) {
	if err := uc.deps.Access.EnsureWritable(ctx, userID); err != nil {
		return Taxpayer{}, err
	}
	t, err := uc.findWritable(ctx, id)
	if err != nil {
		return Taxpayer{}, err
	}
	fields := &fieldErrors{}
	var name domain.DisplayName
	if in.DisplayName != nil {
		if name, err = domain.NewDisplayName(*in.DisplayName); err != nil {
			fields.add(fieldDisplayName, DetailInvalid, msgInvalidField, err)
		}
	}
	var rfc domain.RFC
	if in.RFC != nil {
		if rfc, err = parseRFC(t.PersonType, *in.RFC); err != nil {
			fields.addRFC(err)
		}
	}
	if err := fields.err(); err != nil {
		return Taxpayer{}, err
	}
	if err := uc.write(ctx, userID, id, name, rfc); err != nil {
		return Taxpayer{}, err
	}
	updated, err := uc.load(ctx, id)
	if err != nil {
		return Taxpayer{}, err
	}
	updated.Warnings = rfcWarnings(rfc)
	return updated, nil
}

// write applies a rename and an RFC correction; zero values are skipped.
func (uc *Taxpayers) write(ctx context.Context, userID, id string, name domain.DisplayName, rfc domain.RFC) error {
	if !name.IsZero() {
		if err := uc.deps.Taxpayers.Rename(ctx, id, name); err != nil {
			return err
		}
	}
	if rfc.IsZero() {
		return nil
	}
	sealed, err := uc.deps.RFC.Seal(rfc, RowBinding{TaxpayerID: id, UserID: userID})
	if err != nil {
		return err
	}
	return uc.deps.Taxpayers.ReplaceRFC(ctx, id, sealed)
}

// Archive hides a taxpayer, keeping its data (archiveTaxpayer). Allowed for
// a read-only user; archiving an archived taxpayer answers it unchanged.
func (uc *Taxpayers) Archive(ctx context.Context, userID, rawID string) (Taxpayer, error) {
	return uc.setArchived(ctx, userID, rawID, true)
}

// Unarchive returns a taxpayer to the list (unarchiveTaxpayer, OQ-2). A
// read-only user gets 402. There is no cap check: the cap counts archived
// taxpayers too (OQ-9), so unarchiving never adds one.
func (uc *Taxpayers) Unarchive(ctx context.Context, userID, rawID string) (Taxpayer, error) {
	return uc.setArchived(ctx, userID, rawID, false)
}

func (uc *Taxpayers) setArchived(ctx context.Context, userID, rawID string, archive bool) (Taxpayer, error) {
	op, event := "unarchive taxpayer", eventUnarchived
	if archive {
		op, event = "archive taxpayer", eventArchived
	}
	id, ok := canonicalID(rawID)
	if !ok {
		return Taxpayer{}, notFound()
	}
	var out Taxpayer
	changed := false
	err := uc.deps.Tx.ReadWrite(ctx, userID, func(ctx context.Context) error {
		if !archive {
			if err := uc.deps.Access.EnsureWritable(ctx, userID); err != nil {
				return err
			}
		}
		t, err := uc.deps.Taxpayers.Find(ctx, id, LockUpdate)
		if err != nil {
			return err
		}
		if t.Archived() != archive {
			changed = true
			var at time.Time
			if archive {
				at = uc.now()
			}
			if err := uc.deps.Taxpayers.SetArchivedAt(ctx, id, at); err != nil {
				return err
			}
		}
		out, err = uc.load(ctx, id)
		return err
	})
	if err != nil {
		return Taxpayer{}, present(op, err)
	}
	if changed {
		uc.audit(ctx, event, userID, id)
	}
	return out, nil
}

// Delete permanently deletes a taxpayer (deleteTaxpayer): its regimes and,
// from M3 on, its records cascade in the database. confirm must be true.
// Allowed for a read-only user; the trial counters are not given back.
func (uc *Taxpayers) Delete(ctx context.Context, userID, rawID string, confirm bool) error {
	if !confirm {
		return apperr.Wrap(ErrConfirmRequired, apperr.CodeValidationFailed, msgInvalidInput).
			WithFields(apperr.FieldError{Field: fieldConfirm, Code: DetailConfirmRequired, Message: msgConfirm})
	}
	id, ok := canonicalID(rawID)
	if !ok {
		return notFound()
	}
	err := uc.deps.Tx.ReadWrite(ctx, userID, func(ctx context.Context) error {
		deleted, err := uc.deps.Taxpayers.Delete(ctx, id)
		if err == nil && !deleted {
			err = ErrTaxpayerNotFound
		}
		return err
	})
	if err != nil {
		return present("delete taxpayer", err)
	}
	uc.audit(ctx, eventDeleted, userID, id)
	return nil
}
