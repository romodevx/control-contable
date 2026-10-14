package annual

import (
	"slices"

	"github.com/romodevx/control-contable/backend/internal/platform/money"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain"
)

// NoteKind is what a deductions note reports; it is also the code of the
// informational breakdown line the note becomes (Note.Step).
type NoteKind string

// The note kinds.
const (
	// NoteExcludedCash: a row of a cash-restricted category paid in cash.
	NoteExcludedCash NoteKind = "deductions.excluded_cash"
	// NoteExcludedUnverified: donations without the prior year's
	// acumulables, whose 7 % cap cannot be checked (OQ-A5).
	NoteExcludedUnverified NoteKind = "deductions.excluded_unverified"
	// NoteCapApplied: "Tope aplicado: {concept}, de {x} a {y}".
	NoteCapApplied NoteKind = "deductions.cap_applied"
)

// CapKind names the cap a NoteCapApplied applied.
type CapKind string

// The caps of LISR art. 151 and the colegiaturas decree.
const (
	CapFuneral             CapKind = "funeral"
	CapDonations           CapKind = "donations"
	CapGovernmentDonations CapKind = "government_donations"
	CapRetirement          CapKind = "retirement"
	CapTuition             CapKind = "tuition"
	CapGlobal              CapKind = "global"
)

// noteLabels are the label keys of the lines, rendered by the adapter.
var noteLabels = map[NoteKind]domain.LabelKey{
	NoteExcludedCash:       "annual_deduction_excluded_cash",
	NoteExcludedUnverified: "annual_deduction_excluded_unverified",
	NoteCapApplied:         "annual_deduction_cap_applied",
}

// Note is one exclusion or one cap that bound: From is the amount before,
// To the amount after (0 for an exclusion). Category is empty for the
// global cap; Concept is the student of a tuition note.
type Note struct {
	Kind     NoteKind
	Cap      CapKind
	Category Category
	Concept  string
	From, To money.Cents
	Basis    []domain.Citation
}

// Step is the informational breakdown step of the note, for
// domain.Ledger.Info with To as its amount. Its args are the cap (or the
// category), the concept, From and To.
func (n Note) Step() domain.Step {
	subject := string(n.Cap)
	if subject == "" {
		subject = string(n.Category)
	}
	return domain.Step{
		Tax:    domain.TaxISR,
		Code:   domain.LineCode(n.Kind),
		Label:  noteLabels[n.Kind],
		Args:   []string{subject, n.Concept, n.From.String(), n.To.String()},
		Source: domain.SourceComputed(),
		Basis:  slices.Clone(n.Basis),
	}
}

const (
	lawLISR    = "LISR"
	lawDecree  = "Decreto DOF 26-12-2013"
	articleTop = "151"
)

// The legal basis of each note.
var (
	basisFuneral    = []domain.Citation{{Law: lawLISR, Article: articleTop, Note: "fracción II, hasta 1 UMA anual"}}
	basisDonations  = []domain.Citation{{Law: lawLISR, Article: articleTop, Note: "fracción III, 7 % de los ingresos acumulables del año anterior"}}
	basisGovernment = []domain.Citation{{Law: lawLISR, Article: articleTop, Note: "fracción III, donativos a la Federación, entidades y municipios hasta 4 %"}}
	basisRetirement = []domain.Citation{{Law: lawLISR, Article: articleTop, Note: "fracción V, 10 % de los ingresos acumulables hasta 5 UMA anuales"}}
	basisTuition    = []domain.Citation{{Law: lawDecree, Article: "1.8 a 1.10", Note: "límite anual por alumno y nivel educativo"}}
	basisGlobal     = []domain.Citation{{Law: lawLISR, Article: articleTop, Note: "último párrafo, mínimo entre 5 UMA anuales y 15 % de los ingresos totales"}}
	basisUnverified = basisDonations
)

// cashBasis is the electronic-payment requirement of each cash-restricted
// category.
var cashBasis = map[Category][]domain.Citation{
	CategoryMedical:           {{Law: lawLISR, Article: articleTop, Note: "fracción I, pago con medios electrónicos"}},
	CategoryMedicalDisability: {{Law: lawLISR, Article: articleTop, Note: "fracción I, pago con medios electrónicos"}},
	CategorySchoolTransport:   {{Law: lawLISR, Article: articleTop, Note: "fracción VII, pago con medios electrónicos"}},
	CategoryTuition:           {{Law: lawDecree, Article: "1.8", Note: "pago con medios electrónicos"}},
}
