package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/platform/apperr"
	"github.com/romodevx/control-contable/backend/internal/taxpayers/domain"
	"github.com/romodevx/control-contable/backend/internal/taxpayers/usecase"
)

// fakeDirectory is the regime catalog listing, per person type.
type fakeDirectory struct {
	regimes  map[domain.PersonType][]usecase.ListedRegime
	pairs    map[domain.PersonType][]usecase.PairVerdict
	err      error
	sawUsers []string
}

func (f *fakeDirectory) ListRegimes(ctx context.Context, pt domain.PersonType) ([]usecase.ListedRegime, error) {
	f.sawUsers = append(f.sawUsers, currentUser(ctx))
	return f.regimes[pt], f.err
}

func (f *fakeDirectory) ListCompatibility(_ context.Context, pt domain.PersonType) ([]usecase.PairVerdict, error) {
	return f.pairs[pt], f.err
}

func listed(code string, pt domain.PersonType) usecase.ListedRegime {
	return usecase.ListedRegime{Code: code, PersonType: pt, Name: "Régimen " + code, Treatment: "full"}
}

func verdictOf(a, b string, allowed, review bool, reason string) usecase.PairVerdict {
	return usecase.PairVerdict{
		Pair:          usecase.PairOf(a, b),
		Compatibility: usecase.Compatibility{Allowed: allowed, NeedsCPAReview: review, Reason: reason},
	}
}

func TestCatalogList_groupsPairsOnBothRegimes(t *testing.T) {
	dir := &fakeDirectory{
		regimes: map[domain.PersonType][]usecase.ListedRegime{
			domain.PersonTypeFisica: {
				listed("605", domain.PersonTypeFisica), listed("606", domain.PersonTypeFisica),
				listed("612", domain.PersonTypeFisica), listed("625", domain.PersonTypeFisica),
				listed("626", domain.PersonTypeFisica),
			},
		},
		pairs: map[domain.PersonType][]usecase.PairVerdict{
			domain.PersonTypeFisica: {
				verdictOf("612", "626", false, false, "RESICO sustituye 612."),
				verdictOf("606", "626", false, false, "RESICO cubre arrendamiento."),
				verdictOf("625", "626", true, true, "Pendiente de revisión."),
				verdictOf("605", "612", true, false, "Compatible."),
			},
		},
	}
	tx := &fakeTx{}
	uc := usecase.NewCatalog(usecase.Deps{Tx: tx, Directory: dir})

	got, err := uc.List(t.Context(), userA, "fisica")

	require.NoError(t, err)
	byCode := map[string]usecase.CatalogEntry{}
	var order []string
	for _, e := range got {
		byCode[e.Code] = e
		order = append(order, e.Code)
	}
	assert.Equal(t, []string{"605", "606", "612", "625", "626"}, order, "the directory's code order is kept")
	assert.Equal(t, []usecase.RegimeLink{
		{Code: "606", Reason: "RESICO cubre arrendamiento."},
		{Code: "612", Reason: "RESICO sustituye 612."},
	}, byCode["626"].IncompatibleWith, "ordered by code")
	assert.Equal(t, []usecase.RegimeLink{{Code: "626", Reason: "RESICO sustituye 612."}}, byCode["612"].IncompatibleWith, "symmetric")
	assert.Equal(t, []usecase.RegimeLink{{Code: "625", Reason: "Pendiente de revisión."}}, byCode["626"].ReviewPendingWith)
	assert.Equal(t, []usecase.RegimeLink{{Code: "626", Reason: "Pendiente de revisión."}}, byCode["625"].ReviewPendingWith)
	assert.Empty(t, byCode["605"].IncompatibleWith, "an allowed, verified pair is listed nowhere")
	assert.Empty(t, byCode["605"].ReviewPendingWith)
	assert.NotNil(t, byCode["605"].IncompatibleWith, "empty lists, never nil (the contract requires the array)")
	assert.NotNil(t, byCode["605"].ReviewPendingWith)
	assert.Equal(t, []string{"ro"}, tx.modes, "one read-only unit of work")
	assert.Equal(t, []string{userA}, dir.sawUsers)
}

func TestCatalogList_personTypeIsValidated(t *testing.T) {
	uc := usecase.NewCatalog(usecase.Deps{Tx: &fakeTx{}, Directory: &fakeDirectory{}})

	_, err := uc.List(t.Context(), userA, "juridica")

	requireAppErr(t, err, apperr.CodeValidationFailed, domain.ErrInvalidPersonType, [2]string{"person_type", usecase.DetailInvalid})
}

func TestCatalogList_infrastructureFailureIsAServerFault(t *testing.T) {
	boom := errors.New("connection reset")
	uc := usecase.NewCatalog(usecase.Deps{Tx: &fakeTx{}, Directory: &fakeDirectory{err: boom}})

	_, err := uc.List(t.Context(), userA, "moral")

	require.ErrorIs(t, err, boom)
	_, isAppErr := errors.AsType[*apperr.Error](err)
	assert.False(t, isAppErr, "the router answers the generic 500")
}
