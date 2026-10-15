package annual

import (
	"fmt"
	"slices"
	"strings"

	"github.com/romodevx/control-contable/backend/internal/platform/money"
)

// Error is an annual domain error. Its values are constants, so callers
// match them with errors.Is and nothing can reassign them.
type Error string

func (e Error) Error() string { return string(e) }

// The annual domain errors.
const (
	// ErrInvalidDeduction reports a personal deduction or a deductions input
	// that breaks MODULE-annual §2.6 (amount ≤ 0, an unknown enum value, a
	// field of another category, a negative income).
	ErrInvalidDeduction Error = "annual: invalid personal deduction"
	// ErrInvalidLimits reports deduction limits that cannot be applied (no
	// UMA, a share outside [0, 1], a tuition level without its cap).
	ErrInvalidLimits Error = "annual: invalid deduction limits"
)

// Category is a personal deduction's kind (MODULE-annual §2.6): LISR art.
// 151 fr. I–VIII and the colegiaturas decree.
type Category string

// The categories, in the order of LISR art. 151 (then the decree): the
// order every result lists them in and the global cap is shared in.
const (
	CategoryMedical                Category = "medical"
	CategoryMedicalDisability      Category = "medical_disability"
	CategoryFuneral                Category = "funeral"
	CategoryDonation               Category = "donation"
	CategoryMortgageRealInterest   Category = "mortgage_real_interest"
	CategoryRetirementContribution Category = "retirement_contribution"
	CategoryHealthInsurance        Category = "health_insurance"
	CategorySchoolTransport        Category = "school_transport"
	CategoryLocalPayrollTax        Category = "local_payroll_tax"
	CategoryTuition                Category = "tuition"
)

// categoryOrder is the article order of the categories.
var categoryOrder = []Category{
	CategoryMedical, CategoryMedicalDisability, CategoryFuneral, CategoryDonation, CategoryMortgageRealInterest,
	CategoryRetirementContribution, CategoryHealthInsurance, CategorySchoolTransport, CategoryLocalPayrollTax,
	CategoryTuition,
}

// Valid reports a spec category.
func (c Category) Valid() bool { return slices.Contains(categoryOrder, c) }

// Capped reports a category inside the global cap (LISR 151 last ¶):
// every category except disability medical expenses, retirement
// contributions and tuition.
func (c Category) Capped() bool {
	return c.Valid() && c != CategoryMedicalDisability && c != CategoryRetirementContribution && c != CategoryTuition
}

// CashRestricted reports a category the law makes deductible only when paid
// by cheque, transfer or card: medical (LISR 151 fr. I ¶2, also disability
// expenses), school transport (fr. VII) and tuition (decree art. 1.8).
func (c Category) CashRestricted() bool {
	return slices.Contains([]Category{
		CategoryMedical, CategoryMedicalDisability, CategorySchoolTransport, CategoryTuition,
	}, c)
}

// PaymentMethod is how a deduction was paid.
type PaymentMethod string

// The payment methods.
const (
	PaymentElectronic PaymentMethod = "electronic"
	PaymentCash       PaymentMethod = "cash"
)

// Valid reports a spec payment method.
func (p PaymentMethod) Valid() bool { return p == PaymentElectronic || p == PaymentCash }

// Beneficiary is whose expense a deduction is. A dependant must have had an
// income below one UMA anual; the user attests it (MODULE-annual §2.6).
type Beneficiary string

// The beneficiaries.
const (
	BeneficiarySelf       Beneficiary = "self"
	BeneficiarySpouse     Beneficiary = "spouse"
	BeneficiaryAscendant  Beneficiary = "ascendant"
	BeneficiaryDescendant Beneficiary = "descendant"
)

// Valid reports a spec beneficiary.
func (b Beneficiary) Valid() bool {
	return slices.Contains([]Beneficiary{BeneficiarySelf, BeneficiarySpouse, BeneficiaryAscendant, BeneficiaryDescendant}, b)
}

// DoneeType is who received a donation (LISR 151 fr. III).
type DoneeType string

// The donee types.
const (
	DoneeAuthorized DoneeType = "authorized"
	DoneeGovernment DoneeType = "government"
)

// Valid reports a spec donee type.
func (d DoneeType) Valid() bool { return d == DoneeAuthorized || d == DoneeGovernment }

// TuitionLevel is a school level of the colegiaturas decree (art. 1.10 table).
type TuitionLevel string

// The tuition levels.
const (
	LevelPreescolar         TuitionLevel = "preescolar"
	LevelPrimaria           TuitionLevel = "primaria"
	LevelSecundaria         TuitionLevel = "secundaria"
	LevelProfesionalTecnico TuitionLevel = "profesional_tecnico"
	LevelBachillerato       TuitionLevel = "bachillerato"
)

// tuitionLevels are every level, each of which needs a cap.
var tuitionLevels = []TuitionLevel{
	LevelPreescolar, LevelPrimaria, LevelSecundaria, LevelProfesionalTecnico, LevelBachillerato,
}

// Valid reports a spec tuition level.
func (l TuitionLevel) Valid() bool { return slices.Contains(tuitionLevels, l) }

// Deduction is one registered personal deduction, as on its CFDI or
// constancia. DoneeType applies to donations only; Student, TuitionLevel
// and Scholarship to tuition only (the cap is per student).
type Deduction struct {
	Category      Category
	Amount        money.Cents
	PaymentMethod PaymentMethod
	Beneficiary   Beneficiary
	DoneeType     DoneeType
	Student       string
	TuitionLevel  TuitionLevel
	Scholarship   money.Cents
}

// Validate reports a deduction that breaks MODULE-annual §2.6 as
// ErrInvalidDeduction.
func (d Deduction) Validate() error {
	var problems []string
	if !d.Category.Valid() {
		problems = append(problems, fmt.Sprintf("category %q", d.Category))
	}
	if d.Amount <= 0 {
		problems = append(problems, "amount must be positive")
	}
	if !d.PaymentMethod.Valid() {
		problems = append(problems, fmt.Sprintf("payment method %q", d.PaymentMethod))
	}
	if !d.Beneficiary.Valid() {
		problems = append(problems, fmt.Sprintf("beneficiary %q", d.Beneficiary))
	}
	problems = append(problems, d.donationProblems()...)
	problems = append(problems, d.tuitionProblems()...)
	if len(problems) > 0 {
		return fmt.Errorf("%s: %s: %w", d.Category, strings.Join(problems, "; "), ErrInvalidDeduction)
	}
	return nil
}

func (d Deduction) donationProblems() []string {
	if d.Category != CategoryDonation {
		if d.DoneeType != "" {
			return []string{"donee type is for donations only"}
		}
		return nil
	}
	if !d.DoneeType.Valid() {
		return []string{fmt.Sprintf("donee type %q", d.DoneeType)}
	}
	return nil
}

func (d Deduction) tuitionProblems() []string {
	if d.Category != CategoryTuition {
		if d.Student != "" || d.TuitionLevel != "" || d.Scholarship != 0 {
			return []string{"student, level and scholarship are for tuition only"}
		}
		return nil
	}
	var problems []string
	if strings.TrimSpace(d.Student) == "" {
		problems = append(problems, "tuition needs its student")
	}
	if !d.TuitionLevel.Valid() {
		problems = append(problems, fmt.Sprintf("tuition level %q", d.TuitionLevel))
	}
	if d.Scholarship < 0 {
		problems = append(problems, "scholarship must not be negative")
	}
	return problems
}
