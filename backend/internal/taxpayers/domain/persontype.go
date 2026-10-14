package domain

// PersonType is the kind of taxpayer: a persona física (an individual) or a
// persona moral (a legal entity). The zero value is not valid.
type PersonType string

// The contract's PersonType enum.
const (
	PersonTypeFisica PersonType = "fisica"
	PersonTypeMoral  PersonType = "moral"
)

// ParsePersonType returns the PersonType spelled exactly raw, else
// ErrInvalidPersonType. It neither trims nor folds case: the contract enum
// is exact.
func ParsePersonType(raw string) (PersonType, error) {
	p := PersonType(raw)
	if !p.Valid() {
		return "", ErrInvalidPersonType
	}
	return p, nil
}

// Valid reports whether p is one of the contract values.
func (p PersonType) Valid() bool { return p == PersonTypeFisica || p == PersonTypeMoral }

func (p PersonType) String() string { return string(p) }
