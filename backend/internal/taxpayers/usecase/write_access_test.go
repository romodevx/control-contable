package usecase_test

import (
	"context"
	"maps"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/platform/access"
	"github.com/romodevx/control-contable/backend/internal/platform/apperr"
	"github.com/romodevx/control-contable/backend/internal/taxpayers/domain"
	"github.com/romodevx/control-contable/backend/internal/taxpayers/usecase"
)

// The ids of the read-only matrix.
const (
	matrixActive   = "0199d4a2-0000-7000-8000-0000000000c1"
	matrixArchived = "0199d4a2-0000-7000-8000-0000000000c2"
)

// TestReadOnlyUser_EveryWritePath is the read-only matrix of every write
// use case (MODULE-taxpayers OQ-2): the ones that add or change data answer
// 402 entitlement_required and leave the store untouched; the ones that
// only archive or remove data stay allowed. A new write use case belongs in
// this table.
func TestReadOnlyUser_EveryWritePath(t *testing.T) {
	name, rfc := "Renamed", rfcFisicaOther
	tests := []struct {
		name    string
		allowed bool
		call    func(ctx context.Context, h *harness) error
	}{
		{"create", false, func(ctx context.Context, h *harness) error {
			_, err := h.taxpayers.Create(ctx, userA, fisicaInput())
			return err
		}},
		{"rename", false, func(ctx context.Context, h *harness) error {
			_, err := h.taxpayers.Update(ctx, userA, matrixActive, usecase.UpdateInput{DisplayName: &name})
			return err
		}},
		{"correct rfc", false, func(ctx context.Context, h *harness) error {
			_, err := h.taxpayers.Update(ctx, userA, matrixActive, usecase.UpdateInput{RFC: &rfc})
			return err
		}},
		{"unarchive", false, func(ctx context.Context, h *harness) error {
			_, err := h.taxpayers.Unarchive(ctx, userA, matrixArchived)
			return err
		}},
		{"add regime", false, func(ctx context.Context, h *harness) error {
			_, err := h.regimes.Add(ctx, userA, matrixActive, usecase.RegimeInput{Code: "625", Options: map[string]domain.OptionValue{
				"activities": domain.ListOption([]string{"goods_services"}),
			}})
			return err
		}},
		{"reopen regime", false, func(ctx context.Context, h *harness) error {
			_, err := h.regimes.Add(ctx, userA, matrixActive, usecase.RegimeInput{Code: "605"})
			return err
		}},
		{"update regime options", false, func(ctx context.Context, h *harness) error {
			_, err := h.regimes.UpdateOptions(ctx, userA, matrixActive, "606",
				map[string]domain.OptionValue{"deduction_mode": domain.StringOption("actual")})
			return err
		}},
		{"archive", true, func(ctx context.Context, h *harness) error {
			_, err := h.taxpayers.Archive(ctx, userA, matrixActive)
			return err
		}},
		{"end regime", true, func(ctx context.Context, h *harness) error {
			_, err := h.regimes.End(ctx, userA, matrixActive, usecase.EndRegimeInput{Code: "606", End: "2026-09"})
			return err
		}},
		{"delete", true, func(ctx context.Context, h *harness) error {
			return h.taxpayers.Delete(ctx, userA, matrixActive, true)
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newReadOnlyMatrixHarness(t)
			before := storeState(h.store)

			err := tt.call(t.Context(), h)

			if tt.allowed {
				require.NoError(t, err)
				assert.NotEqual(t, before, storeState(h.store), "an allowed write changes the store")
				return
			}
			requireAppErr(t, err, apperr.CodeEntitlementRequired, access.ErrReadOnly)
			assert.Equal(t, before, storeState(h.store), "a refused write leaves the store untouched")
			assert.Empty(t, h.rfc.calls, "nothing is sealed for a refused write")
		})
	}
}

// newReadOnlyMatrixHarness seeds an active taxpayer holding 606 and 612
// (open) and 605 (ended), plus an archived taxpayer, then makes userA
// read-only.
func newReadOnlyMatrixHarness(t *testing.T) *harness {
	t.Helper()
	h := newHarness(t)
	h.store.seedTaxpayer(userA, matrixActive, domain.PersonTypeFisica, false,
		seedRegime{code: "606", window: window(t, "2026-01", ""),
			options: map[string]domain.OptionValue{"deduction_mode": domain.StringOption("blind_35")}},
		seedRegime{code: "612", window: window(t, "2026-01", "")},
		seedRegime{code: "605", window: window(t, "2025-01", "2025-06")})
	h.store.seedTaxpayer(userA, matrixArchived, domain.PersonTypeFisica, true,
		seedRegime{code: "612", window: window(t, "2026-01", "")})
	h.access.readOnly = true
	return h
}

// memState is a deep copy of the fake store's rows.
type memState struct {
	rows    map[string]storedRow
	regimes map[string]regimeRow
}

func storeState(s *memStore) memState {
	state := memState{rows: make(map[string]storedRow, len(s.rows)), regimes: make(map[string]regimeRow, len(s.regimes))}
	for id, row := range s.rows {
		state.rows[id] = *row
	}
	for key, row := range s.regimes {
		copied := *row
		copied.held.Options = maps.Clone(row.held.Options)
		state.regimes[key] = copied
	}
	return state
}
