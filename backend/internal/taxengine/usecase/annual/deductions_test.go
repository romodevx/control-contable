package annual_test

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/platform/apperr"
	"github.com/romodevx/control-contable/backend/internal/platform/clock"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
	adomain "github.com/romodevx/control-contable/backend/internal/taxengine/domain/annual"
	"github.com/romodevx/control-contable/backend/internal/taxengine/usecase/annual"
)

func (w *world) deductions(now time.Time) *annual.Deductions {
	return annual.NewDeductions(annual.DeductionsDeps{
		Tx: w.tx, Access: w.access, Taxpayers: w.db, Store: w.db, Clock: clock.NewFixed(now),
		NewID: fixedID, Logger: w.logger(),
	})
}

// medical is the AN-01 medical expense: electronic, for oneself.
func medical(id string) annual.DeductionInput {
	return annual.DeductionInput{
		ID: id, TaxpayerID: fisicaA, Category: adomain.CategoryMedical, Amount: 4_000_000,
		PaidOn: mustDate("2026-06-12"), PaymentMethod: adomain.PaymentElectronic, Beneficiary: adomain.BeneficiarySelf,
	}
}

func tuition(id string) annual.DeductionInput {
	in := medical(id)
	in.Category, in.Amount, in.Beneficiary = adomain.CategoryTuition, 3_000_000, adomain.BeneficiaryDescendant
	in.StudentLabel, in.TuitionLevel, in.Scholarship = ptr("Hija menor"), ptr(adomain.LevelPrimaria), ptr(money.Cents(0))
	return in
}

// seedDeduction stores a live deduction of userA's taxpayer directly.
func (w *world) seedDeduction(id, taxpayerID string, f annual.DeductionFields) {
	w.db.deductions[id] = ownedDeduction{owner: userA, row: annual.StoredDeduction{
		ID: id, TaxpayerID: taxpayerID, FiscalYear: f.PaidOn.Year(), DeductionFields: f, CreatedAt: stamp, UpdatedAt: stamp,
	}}
}

func medicalFields(paidOn string, amount money.Cents, method adomain.PaymentMethod) annual.DeductionFields {
	return annual.DeductionFields{
		Category: adomain.CategoryMedical, Amount: amount, PaidOn: mustDate(paidOn), PaymentMethod: method,
		Beneficiary: adomain.BeneficiarySelf,
	}
}

// fill stores n live deductions of fisicaA paid on paidOn.
func (w *world) fill(n int, paidOn string) {
	for i := range n {
		w.seedDeduction(fmt.Sprintf("0199d4b1-ffff-7000-8000-%012d", i), fisicaA, medicalFields(paidOn, 100, adomain.PaymentElectronic))
	}
}

func TestDeductions_CreateStoresAndAudits(t *testing.T) {
	w := newWorld()

	res, err := w.deductions(today).Create(t.Context(), userA, medical(clientID))
	require.NoError(t, err)

	assert.True(t, res.Created)
	assert.Empty(t, res.Warnings)
	assert.Equal(t, annual.Deduction{
		ID: clientID, TaxpayerID: fisicaA, FiscalYear: 2026, DeductionFields: medicalFields("2026-06-12", 4_000_000, adomain.PaymentElectronic),
		CreatedAt: stamp, UpdatedAt: stamp,
	}, res.Deduction)
	assert.Equal(t, []string{"lock-id", "find:none", "taxpayer:write", "list:1:2026", "count", "insert", "find:none"}, w.db.calls)
	assert.Equal(t, []string{"rw"}, w.tx.modes)
	assert.Equal(t, 1, w.access.calls)
	assert.Contains(t, w.logs.String(), `"msg":"personal_deduction.created"`)
	assert.Contains(t, w.logs.String(), clientID)
	assert.NotContains(t, w.logs.String(), "4000000", "amounts are never logged")
}

func TestDeductions_CreateWithoutIDGeneratesOne(t *testing.T) {
	w := newWorld()
	res, err := w.deductions(today).Create(t.Context(), userA, medical(""))
	require.NoError(t, err)
	assert.Equal(t, newRowID, res.Deduction.ID)
	assert.NotContains(t, w.db.calls, "lock-id", "a server id needs no idempotency lock")
}

func TestDeductions_CreateCanonicalizesIDsAndNormalizesText(t *testing.T) {
	w := newWorld()
	in := tuition(strings.ToUpper(clientID))
	in.TaxpayerID = strings.ToUpper(fisicaA)
	in.StudentLabel = ptr("María")
	in.CFDIUUID = ptr("6f3b2c1d-8a4e-4b7f-9c2d-1e5a7b9c3d4f")

	res, err := w.deductions(today).Create(t.Context(), userA, in)
	require.NoError(t, err)
	assert.Equal(t, clientID, res.Deduction.ID)
	assert.Equal(t, fisicaA, res.Deduction.TaxpayerID)
	assert.Equal(t, "María", res.Deduction.StudentLabel, "stored NFC (ADR-0004)")
	assert.Equal(t, "6F3B2C1D-8A4E-4B7F-9C2D-1E5A7B9C3D4F", res.Deduction.CFDIUUID)
	assert.Equal(t, annual.NullCents{Valid: true}, res.Deduction.Scholarship)
	assert.NoError(t, res.Deduction.Domain().Validate(), "every accepted row is a valid calculator input")
}

func TestDeductions_CashIsStoredButExcluded(t *testing.T) {
	cashWarning := []annual.Warning{{Code: annual.WarningCashNotDeductible, Field: annual.FieldPaymentMethod}}
	tests := []struct {
		name     string
		category adomain.Category
		excluded bool
	}{
		{"medical", adomain.CategoryMedical, true},
		{"funeral", adomain.CategoryFuneral, false},
		{"tuition", adomain.CategoryTuition, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := newWorld()
			in := tuition(clientID)
			if tt.category != adomain.CategoryTuition {
				in = medical(clientID)
				in.Category = tt.category
			}
			in.PaymentMethod = adomain.PaymentCash

			res, err := w.deductions(today).Create(t.Context(), userA, in)
			require.NoError(t, err)
			if tt.excluded {
				assert.Equal(t, annual.ExclusionCashNotDeductible, res.Deduction.Excluded)
				assert.Equal(t, cashWarning, res.Warnings)
				return
			}
			assert.Empty(t, res.Deduction.Excluded)
			assert.Empty(t, res.Warnings)
		})
	}
}

func TestDeductions_CreateIdempotency(t *testing.T) {
	t.Run("a retry with the same fields replays without an access check", func(t *testing.T) {
		w := newWorld()
		uc := w.deductions(today)
		first, err := uc.Create(t.Context(), userA, medical(clientID))
		require.NoError(t, err)
		w.access.readOnly = true

		again, err := uc.Create(t.Context(), userA, medical(clientID))
		require.NoError(t, err)
		assert.False(t, again.Created)
		assert.Equal(t, first.Deduction, again.Deduction)
		assert.Equal(t, 1, w.access.calls)
	})
	t.Run("a retry carrying the cash warning replays it", func(t *testing.T) {
		w := newWorld()
		uc := w.deductions(today)
		in := medical(clientID)
		in.PaymentMethod = adomain.PaymentCash
		_, err := uc.Create(t.Context(), userA, in)
		require.NoError(t, err)
		again, err := uc.Create(t.Context(), userA, in)
		require.NoError(t, err)
		assert.Len(t, again.Warnings, 1)
	})
	reused := []struct {
		name  string
		setup func(w *world)
		in    func() annual.DeductionInput
	}{
		{"other fields", func(*world) {}, func() annual.DeductionInput { in := medical(clientID); in.Amount++; return in }},
		{"another taxpayer", func(*world) {}, func() annual.DeductionInput { in := medical(clientID); in.TaxpayerID = fisica2A; return in }},
		{"a deleted row", func(w *world) {
			d := w.db.deductions[clientID]
			d.row.Deleted = true
			w.db.deductions[clientID] = d
		}, func() annual.DeductionInput { return medical(clientID) }},
	}
	for _, tt := range reused {
		t.Run("id reused: "+tt.name, func(t *testing.T) {
			w := newWorld()
			uc := w.deductions(today)
			_, err := uc.Create(t.Context(), userA, medical(clientID))
			require.NoError(t, err)
			tt.setup(w)

			_, err = uc.Create(t.Context(), userA, tt.in())
			assert.Equal(t, []string{"id/id_reused"}, requireCode(t, err, apperr.CodeConflict))
			assert.ErrorIs(t, err, annual.ErrIDReused)
		})
	}
	t.Run("another user's id is a 409 without details", func(t *testing.T) {
		w := newWorld()
		w.db.deductions[clientID] = ownedDeduction{owner: userB, row: annual.StoredDeduction{ID: clientID, TaxpayerID: fisicaB}}

		_, err := w.deductions(today).Create(t.Context(), userA, medical(clientID))
		assert.Empty(t, requireCode(t, err, apperr.CodeConflict))
		assert.ErrorIs(t, err, annual.ErrIDTaken)
	})
}

func TestDeductions_CreateOrderOfChecks(t *testing.T) {
	tests := []struct {
		name   string
		setup  func(w *world)
		in     func() annual.DeductionInput
		code   apperr.Code
		fields []string
		cause  error
	}{
		{"malformed taxpayer id", nil, func() annual.DeductionInput { in := medical(clientID); in.TaxpayerID = "nope"; return in },
			apperr.CodeNotFound, []string{}, annual.ErrTaxpayerNotFound},
		{"malformed client id", nil, func() annual.DeductionInput { return medical("nope") },
			apperr.CodeValidationFailed, []string{"id/invalid"}, annual.ErrInvalidField},
		{"read-only before every rule", func(w *world) { w.access.readOnly = true; w.archive(fisicaA) },
			func() annual.DeductionInput { in := medical(clientID); in.Amount = 0; return in },
			apperr.CodeEntitlementRequired, []string{}, nil},
		{"another user's taxpayer", nil, func() annual.DeductionInput { in := medical(clientID); in.TaxpayerID = fisicaB; return in },
			apperr.CodeNotFound, []string{}, annual.ErrTaxpayerNotFound},
		{"archived before field rules", func(w *world) { w.archive(fisicaA) },
			func() annual.DeductionInput { in := medical(clientID); in.Amount = 0; return in },
			apperr.CodeBusinessRuleViolation, []string{"taxpayerId/taxpayer_archived"}, annual.ErrTaxpayerArchived},
		{"persona moral", nil, func() annual.DeductionInput { in := medical(clientID); in.TaxpayerID = moralA; return in },
			apperr.CodeBusinessRuleViolation, []string{"taxpayerId/person_type_not_allowed"}, annual.ErrPersonTypeNotAllowed},
		{"500 live deductions in the year", func(w *world) { w.fill(500, "2026-01-05") },
			func() annual.DeductionInput { in := medical(clientID); in.Amount = 0; return in },
			apperr.CodeBusinessRuleViolation, []string{"body/limit_reached"}, annual.ErrLimitReached},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := newWorld()
			if tt.setup != nil {
				tt.setup(w)
			}
			_, err := w.deductions(today).Create(t.Context(), userA, tt.in())
			assert.Equal(t, tt.fields, requireCode(t, err, tt.code))
			if tt.cause != nil {
				assert.ErrorIs(t, err, tt.cause)
			}
			assert.NotContains(t, w.db.calls, "insert")
		})
	}
}

func TestDeductions_LimitCountsOnlyThatYearsLiveRows(t *testing.T) {
	w := newWorld()
	w.fill(499, "2026-01-05")
	w.seedDeduction("0199d4b1-eeee-7000-8000-000000000001", fisicaA, medicalFields("2026-02-01", 1, adomain.PaymentCash))
	d := w.db.deductions["0199d4b1-eeee-7000-8000-000000000001"]
	d.row.Deleted = true
	w.db.deductions[d.row.ID] = d

	_, err := w.deductions(today).Create(t.Context(), userA, medical(clientID))
	require.NoError(t, err, "a deleted row does not count")
}

func TestDeductions_FieldRules(t *testing.T) {
	longLabel := strings.Repeat("a", 41)
	tests := []struct {
		name   string
		in     func() annual.DeductionInput
		fields []string
	}{
		{"amount must be positive", func() annual.DeductionInput { in := medical(clientID); in.Amount = 0; return in }, []string{"amount_cents/invalid"}},
		{"unknown category", func() annual.DeductionInput { in := medical(clientID); in.Category = "gym"; return in }, []string{"category/invalid"}},
		{"unknown payment method", func() annual.DeductionInput { in := medical(clientID); in.PaymentMethod = "barter"; return in }, []string{"payment_method/invalid"}},
		{"unknown beneficiary", func() annual.DeductionInput { in := medical(clientID); in.Beneficiary = "friend"; return in }, []string{"beneficiary/invalid"}},
		{"paid before 2026", func() annual.DeductionInput { in := medical(clientID); in.PaidOn = mustDate("2025-12-31"); return in }, []string{"paid_on/out_of_range"}},
		{"paid after today", func() annual.DeductionInput { in := medical(clientID); in.PaidOn = mustDate("2026-10-15"); return in }, []string{"paid_on/out_of_range"}},
		{"donation needs its donee", func() annual.DeductionInput {
			in := medical(clientID)
			in.Category = adomain.CategoryDonation
			return in
		}, []string{"donee_type/required"}},
		{"unknown donee", func() annual.DeductionInput {
			in := medical(clientID)
			in.Category, in.DoneeType = adomain.CategoryDonation, ptr(adomain.DoneeType("church"))
			return in
		}, []string{"donee_type/invalid"}},
		{"donee only for donations", func() annual.DeductionInput {
			in := medical(clientID)
			in.DoneeType = ptr(adomain.DoneeGovernment)
			return in
		}, []string{"donee_type/not_applicable"}},
		{"tuition needs its fields", func() annual.DeductionInput {
			in := medical(clientID)
			in.Category = adomain.CategoryTuition
			return in
		},
			[]string{"student_label/required", "tuition_level/required", "scholarship_cents/required"}},
		{"tuition fields only for tuition", func() annual.DeductionInput {
			in := tuition(clientID)
			in.Category = adomain.CategoryMedical
			return in
		}, []string{"student_label/not_applicable", "tuition_level/not_applicable", "scholarship_cents/not_applicable"}},
		{"unknown level", func() annual.DeductionInput {
			in := tuition(clientID)
			in.TuitionLevel = ptr(adomain.TuitionLevel("posgrado"))
			return in
		}, []string{"tuition_level/invalid"}},
		{"negative scholarship", func() annual.DeductionInput {
			in := tuition(clientID)
			in.Scholarship = ptr(money.Cents(-1))
			return in
		}, []string{"scholarship_cents/invalid"}},
		{"blank label", func() annual.DeductionInput { in := tuition(clientID); in.StudentLabel = ptr("   "); return in }, []string{"student_label/invalid"}},
		{"untrimmed label", func() annual.DeductionInput { in := tuition(clientID); in.StudentLabel = ptr(" Hija"); return in }, []string{"student_label/invalid"}},
		{"control character", func() annual.DeductionInput { in := tuition(clientID); in.StudentLabel = ptr("Hi\u0007ja"); return in }, []string{"student_label/invalid"}},
		{"label over 40 characters", func() annual.DeductionInput { in := tuition(clientID); in.StudentLabel = &longLabel; return in }, []string{"student_label/invalid"}},
		{"label not UTF-8", func() annual.DeductionInput { in := tuition(clientID); in.StudentLabel = ptr("\xff"); return in }, []string{"student_label/invalid"}},
		{"folio format", func() annual.DeductionInput { in := medical(clientID); in.CFDIUUID = ptr("not-a-folio"); return in }, []string{"cfdi_uuid/invalid"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := newWorld()
			_, err := w.deductions(today).Create(t.Context(), userA, tt.in())
			assert.Equal(t, tt.fields, requireCode(t, err, apperr.CodeValidationFailed))
			assert.ErrorIs(t, err, annual.ErrInvalidField)
			assert.NotContains(t, w.db.calls, "insert")
		})
	}
}

func TestDeductions_AcceptsEveryCategoryAndBoundary(t *testing.T) {
	label40 := strings.Repeat("ñ", 40)
	for _, category := range []adomain.Category{
		adomain.CategoryMedical, adomain.CategoryMedicalDisability, adomain.CategoryFuneral, adomain.CategoryDonation,
		adomain.CategoryMortgageRealInterest, adomain.CategoryRetirementContribution, adomain.CategoryHealthInsurance,
		adomain.CategorySchoolTransport, adomain.CategoryLocalPayrollTax, adomain.CategoryTuition,
	} {
		t.Run(string(category), func(t *testing.T) {
			w := newWorld()
			in := medical(clientID)
			in.Category, in.PaidOn, in.Amount = category, mustDate("2026-01-01"), 1
			switch category {
			case adomain.CategoryDonation:
				in.DoneeType = ptr(adomain.DoneeAuthorized)
			case adomain.CategoryTuition:
				in.StudentLabel, in.TuitionLevel, in.Scholarship = &label40, ptr(adomain.LevelBachillerato), ptr(money.Cents(0))
			}
			res, err := w.deductions(today).Create(t.Context(), userA, in)
			require.NoError(t, err)
			assert.NoError(t, res.Deduction.Domain().Validate())
		})
	}
	t.Run("paid today", func(t *testing.T) {
		w := newWorld()
		in := medical(clientID)
		in.PaidOn = mustDate("2026-10-14")
		_, err := w.deductions(today).Create(t.Context(), userA, in)
		require.NoError(t, err)
	})
}

func TestDeductions_CreateInfrastructureFailureIs500(t *testing.T) {
	w := newWorld()
	w.db.failWith = errors.New("boom")
	_, err := w.deductions(today).Create(t.Context(), userA, medical(clientID))
	require.Error(t, err)
	_, isClient := errors.AsType[*apperr.Error](err)
	assert.False(t, isClient)
}

func TestDeductions_Update(t *testing.T) {
	t.Run("merge-patch re-runs the rules and recomputes the exclusion", func(t *testing.T) {
		w := newWorld()
		w.seedDeduction(clientID, fisicaA, medicalFields("2026-09-02", 120_000, adomain.PaymentCash))

		res, err := w.deductions(today).Update(t.Context(), userA, fisicaA, clientID,
			annual.DeductionPatch{PaymentMethod: ptr(adomain.PaymentElectronic), CFDIUUID: annual.Value("6f3b2c1d-8a4e-4b7f-9c2d-1e5a7b9c3d4f")})
		require.NoError(t, err)
		assert.Empty(t, res.Deduction.Excluded)
		assert.Empty(t, res.Warnings)
		assert.Equal(t, "6F3B2C1D-8A4E-4B7F-9C2D-1E5A7B9C3D4F", res.Deduction.CFDIUUID)
		assert.Equal(t, money.Cents(120_000), res.Deduction.Amount, "absent fields are unchanged")
		assert.False(t, res.Created)
		assert.Equal(t, []string{"find:write", "taxpayer:write", "update", "find:none"}, w.db.calls, "a same-year edit takes no list lock")
		assert.Contains(t, w.logs.String(), `"msg":"personal_deduction.updated"`)
	})
	t.Run("null clears a nullable field", func(t *testing.T) {
		w := newWorld()
		f := medicalFields("2026-09-02", 120_000, adomain.PaymentElectronic)
		f.CFDIUUID = "6F3B2C1D-8A4E-4B7F-9C2D-1E5A7B9C3D4F"
		w.seedDeduction(clientID, fisicaA, f)

		res, err := w.deductions(today).Update(t.Context(), userA, fisicaA, clientID, annual.DeductionPatch{CFDIUUID: annual.Null[string]()})
		require.NoError(t, err)
		assert.Empty(t, res.Deduction.CFDIUUID)
	})
	t.Run("changing the category needs its fields", func(t *testing.T) {
		w := newWorld()
		w.seedDeduction(clientID, fisicaA, medicalFields("2026-09-02", 120_000, adomain.PaymentElectronic))
		uc := w.deductions(today)

		_, err := uc.Update(t.Context(), userA, fisicaA, clientID, annual.DeductionPatch{Category: ptr(adomain.CategoryDonation)})
		assert.Equal(t, []string{"donee_type/required"}, requireCode(t, err, apperr.CodeValidationFailed))

		res, err := uc.Update(t.Context(), userA, fisicaA, clientID,
			annual.DeductionPatch{Category: ptr(adomain.CategoryDonation), DoneeType: annual.Value(adomain.DoneeGovernment)})
		require.NoError(t, err)
		assert.Equal(t, adomain.DoneeGovernment, res.Deduction.DoneeType)
	})
	t.Run("moving to another year applies that year's bound", func(t *testing.T) {
		w := newWorld()
		feb2027 := time.Date(2027, time.February, 10, 12, 0, 0, 0, clock.Location())
		w.seedDeduction(clientID, fisicaA, medicalFields("2026-09-02", 120_000, adomain.PaymentElectronic))
		w.fill(500, "2027-01-05")
		uc := w.deductions(feb2027)

		_, err := uc.Update(t.Context(), userA, fisicaA, clientID, annual.DeductionPatch{PaidOn: ptr(mustDate("2027-02-01"))})
		assert.Equal(t, []string{"body/limit_reached"}, requireCode(t, err, apperr.CodeBusinessRuleViolation))
		assert.Contains(t, w.db.calls, "list:1:2027")

		res, err := uc.Update(t.Context(), userA, fisicaA, clientID, annual.DeductionPatch{PaidOn: ptr(mustDate("2026-12-31"))})
		require.NoError(t, err, "a move inside the row's own year is not bounded again")
		assert.Equal(t, 2026, res.Deduction.FiscalYear)
	})
	notFound := []struct {
		name       string
		taxpayerID string
		id         string
		setup      func(w *world)
	}{
		{"deleted", fisicaA, clientID, func(w *world) {
			d := w.db.deductions[clientID]
			d.row.Deleted = true
			w.db.deductions[clientID] = d
		}},
		{"another taxpayer of the user", fisica2A, clientID, nil},
		{"unknown", fisicaA, clientID2, nil},
		{"malformed id", fisicaA, "nope", nil},
		{"malformed taxpayer id", "nope", clientID, nil},
		{"another user's", fisicaB, clientID, func(w *world) {
			d := w.db.deductions[clientID]
			d.owner, d.row.TaxpayerID = userB, fisicaB
			w.db.deductions[clientID] = d
		}},
	}
	for _, tt := range notFound {
		t.Run("404: "+tt.name, func(t *testing.T) {
			w := newWorld()
			w.seedDeduction(clientID, fisicaA, medicalFields("2026-09-02", 120_000, adomain.PaymentElectronic))
			if tt.setup != nil {
				tt.setup(w)
			}
			_, err := w.deductions(today).Update(t.Context(), userA, tt.taxpayerID, tt.id, annual.DeductionPatch{Amount: ptr(money.Cents(5))})
			assert.Empty(t, requireCode(t, err, apperr.CodeNotFound))
		})
	}
	t.Run("read-only is 402, archived is 422", func(t *testing.T) {
		w := newWorld()
		w.seedDeduction(clientID, fisicaA, medicalFields("2026-09-02", 120_000, adomain.PaymentElectronic))
		w.access.readOnly = true
		uc := w.deductions(today)
		_, err := uc.Update(t.Context(), userA, fisicaA, clientID, annual.DeductionPatch{Amount: ptr(money.Cents(5))})
		requireCode(t, err, apperr.CodeEntitlementRequired)

		w.access.readOnly = false
		w.archive(fisicaA)
		_, err = uc.Update(t.Context(), userA, fisicaA, clientID, annual.DeductionPatch{Amount: ptr(money.Cents(5))})
		assert.Equal(t, []string{"taxpayerId/taxpayer_archived"}, requireCode(t, err, apperr.CodeBusinessRuleViolation))
	})
}

func TestDeductions_DeleteAndRestore(t *testing.T) {
	t.Run("delete soft-deletes, needs no write access and is idempotent", func(t *testing.T) {
		w := newWorld()
		w.access.readOnly = true
		w.seedDeduction(clientID, fisicaA, medicalFields("2026-09-02", 120_000, adomain.PaymentCash))
		uc := w.deductions(today)

		require.NoError(t, uc.Delete(t.Context(), userA, fisicaA, clientID))
		assert.True(t, w.db.deductions[clientID].row.Deleted)
		require.NoError(t, uc.Delete(t.Context(), userA, fisicaA, clientID), "deleting a deleted row is a no-op")
		assert.Equal(t, 0, w.access.calls)
		assert.Equal(t, 1, strings.Count(w.logs.String(), `"msg":"personal_deduction.deleted"`))
	})
	t.Run("restore brings it back with its warnings, read-only included", func(t *testing.T) {
		w := newWorld()
		w.access.readOnly = true
		w.seedDeduction(clientID, fisicaA, medicalFields("2026-09-02", 120_000, adomain.PaymentCash))
		uc := w.deductions(today)
		require.NoError(t, uc.Delete(t.Context(), userA, fisicaA, clientID))

		res, err := uc.Restore(t.Context(), userA, fisicaA, clientID)
		require.NoError(t, err)
		assert.False(t, w.db.deductions[clientID].row.Deleted)
		assert.Equal(t, annual.ExclusionCashNotDeductible, res.Deduction.Excluded)
		assert.Len(t, res.Warnings, 1)
		assert.Contains(t, w.db.calls, "list:1:2026")
		assert.Contains(t, w.logs.String(), `"msg":"personal_deduction.restored"`)

		again, err := uc.Restore(t.Context(), userA, fisicaA, clientID)
		require.NoError(t, err, "restoring a live row answers it unchanged")
		assert.Equal(t, res.Deduction, again.Deduction)
		assert.Equal(t, 0, w.access.calls)
	})
	t.Run("restore re-checks the bound; the row stays restorable", func(t *testing.T) {
		w := newWorld()
		w.seedDeduction(clientID, fisicaA, medicalFields("2026-09-02", 120_000, adomain.PaymentElectronic))
		uc := w.deductions(today)
		require.NoError(t, uc.Delete(t.Context(), userA, fisicaA, clientID))
		w.fill(500, "2026-03-01")

		_, err := uc.Restore(t.Context(), userA, fisicaA, clientID)
		assert.Equal(t, []string{"body/limit_reached"}, requireCode(t, err, apperr.CodeBusinessRuleViolation))
		assert.True(t, w.db.deductions[clientID].row.Deleted)
	})
	t.Run("archived taxpayer refuses both", func(t *testing.T) {
		w := newWorld()
		w.seedDeduction(clientID, fisicaA, medicalFields("2026-09-02", 120_000, adomain.PaymentElectronic))
		w.archive(fisicaA)
		uc := w.deductions(today)
		err := uc.Delete(t.Context(), userA, fisicaA, clientID)
		assert.Equal(t, []string{"taxpayerId/taxpayer_archived"}, requireCode(t, err, apperr.CodeBusinessRuleViolation))
		_, err = uc.Restore(t.Context(), userA, fisicaA, clientID)
		requireCode(t, err, apperr.CodeBusinessRuleViolation)
	})
	t.Run("404 parity", func(t *testing.T) {
		w := newWorld()
		w.seedDeduction(clientID, fisicaA, medicalFields("2026-09-02", 120_000, adomain.PaymentElectronic))
		uc := w.deductions(today)
		for _, ids := range [][2]string{{fisica2A, clientID}, {fisicaA, clientID2}, {fisicaA, "nope"}, {"nope", clientID}, {fisicaB, clientID}} {
			err := uc.Delete(t.Context(), userA, ids[0], ids[1])
			requireCode(t, err, apperr.CodeNotFound)
			_, err = uc.Restore(t.Context(), userA, ids[0], ids[1])
			requireCode(t, err, apperr.CodeNotFound)
		}
	})
}

func TestDeductions_List(t *testing.T) {
	w := newWorld()
	w.seedDeduction("0199d4b1-0000-7000-8000-00000000000a", fisicaA, medicalFields("2026-06-12", 4_000_000, adomain.PaymentElectronic))
	w.seedDeduction("0199d4b1-0000-7000-8000-00000000000b", fisicaA, medicalFields("2026-09-02", 120_000, adomain.PaymentCash))
	w.seedDeduction("0199d4b1-0000-7000-8000-00000000000c", fisicaA, medicalFields("2026-09-02", 1, adomain.PaymentElectronic))
	funeral := medicalFields("2026-01-02", 500, adomain.PaymentCash)
	funeral.Category = adomain.CategoryFuneral
	w.seedDeduction("0199d4b1-0000-7000-8000-000000000001", fisicaA, funeral)
	w.seedDeduction("0199d4b1-0000-7000-8000-000000000002", fisica2A, medicalFields("2026-06-12", 9, adomain.PaymentElectronic))
	uc := w.deductions(today)

	list, err := uc.List(t.Context(), userA, annual.DeductionQuery{TaxpayerID: fisicaA, FiscalYear: 2026})
	require.NoError(t, err)
	var ids []string
	for _, d := range list.Items {
		ids = append(ids, d.ID[len(d.ID)-2:])
	}
	assert.Equal(t, []string{"0b", "0c", "0a", "01"}, ids, "category order, paid_on desc, id")
	assert.Equal(t, annual.ExclusionCashNotDeductible, list.Items[0].Excluded)
	assert.Empty(t, list.Items[3].Excluded, "cash is fine for a funeral")
	assert.Equal(t, []annual.CategoryTotal{
		{Category: adomain.CategoryMedical, Registered: 4_120_001, Excluded: 120_000, Count: 3},
		{Category: adomain.CategoryFuneral, Registered: 500, Count: 1},
	}, list.Totals)
	assert.Equal(t, 2026, list.FiscalYear)
	assert.Equal(t, []string{"ro"}, w.tx.modes)

	filtered, err := uc.List(t.Context(), userA, annual.DeductionQuery{TaxpayerID: fisicaA, FiscalYear: 2026, Category: adomain.CategoryFuneral})
	require.NoError(t, err)
	assert.Len(t, filtered.Items, 1)
	assert.Len(t, filtered.Totals, 1)

	moral, err := uc.List(t.Context(), userA, annual.DeductionQuery{TaxpayerID: moralA, FiscalYear: 2026})
	require.NoError(t, err)
	assert.Empty(t, moral.Items)
	assert.NotNil(t, moral.Items)
	assert.NotNil(t, moral.Totals)

	w.access.readOnly = true
	w.archive(fisicaA)
	_, err = uc.List(t.Context(), userA, annual.DeductionQuery{TaxpayerID: fisicaA, FiscalYear: 2026})
	require.NoError(t, err, "reads stay open to read-only users and archived taxpayers")
}

func TestDeductions_ListRefusals(t *testing.T) {
	tests := []struct {
		name   string
		q      annual.DeductionQuery
		code   apperr.Code
		fields []string
	}{
		{"future year", annual.DeductionQuery{TaxpayerID: fisicaA, FiscalYear: 2027}, apperr.CodeValidationFailed, []string{"fiscal_year/out_of_range"}},
		{"before 2026", annual.DeductionQuery{TaxpayerID: fisicaA, FiscalYear: 2025}, apperr.CodeBusinessRuleViolation, []string{"fiscal_year/parameters_unavailable"}},
		{"another user's taxpayer", annual.DeductionQuery{TaxpayerID: fisicaB, FiscalYear: 2026}, apperr.CodeNotFound, []string{}},
		{"malformed taxpayer", annual.DeductionQuery{TaxpayerID: "x", FiscalYear: 2026}, apperr.CodeNotFound, []string{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := newWorld()
			_, err := w.deductions(today).List(t.Context(), userA, tt.q)
			assert.Equal(t, tt.fields, requireCode(t, err, tt.code))
		})
	}
}
