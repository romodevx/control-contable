package domain

import (
	"slices"
	"strings"

	"golang.org/x/text/unicode/norm"
)

// RFC layout: name letters, a YYMMDD date and a 3-character homoclave whose
// last character is the check digit.
const (
	fisicaNameLength = 4
	moralNameLength  = 3
	dateLength       = 6
	homoclaveLength  = 3
	fisicaLength     = fisicaNameLength + dateLength + homoclaveLength // 13
	moralLength      = moralNameLength + dateLength + homoclaveLength  // 12
	maskRune         = '*'
)

// Check digit parameters (SAT Anexo 3): twelve weighted positions, weights
// 13 down to 2, modulus 11.
const (
	checkedLength = fisicaLength - 1 // 12: a moral prefix is left-padded with one blank
	topWeight     = checkedLength + 1
	checkModulus  = 11
)

// checkAlphabet is the SAT Anexo 3 value table: each rune's index is its
// value. "0"-"9" are 0-9, "A"-"N" 10-23, "&" 24, "O"-"Z" 25-36, the blank
// 37 and "Ñ" 38. The first eleven runes double as the check digits 0-10.
var checkAlphabet = []rune("0123456789ABCDEFGHIJKLMN&OPQRSTUVWXYZ Ñ")

// genericRFCs are the SAT's generic RFCs: XAXX010101000 for the general
// public (público en general) and XEXX010101000 for foreign residents
// without an RFC (residentes en el extranjero). They are placeholders for
// CFDI receivers, not taxpayers, so a taxpayer may never hold one (rule:
// reject, whatever the declared person type). XAXX010101000 does not even
// carry a valid check digit, so it is matched before the digit is checked.
var genericRFCs = []string{"XAXX010101000", "XEXX010101000"}

// RFC is a validated, normalized Registro Federal de Contribuyentes. It is
// personal data for a persona física: it renders as REDACTED in fmt and
// slog, so only an explicit String() or Masked() call reveals it. The zero
// value is not a valid RFC (see IsZero).
type RFC struct {
	redactor
	value           string
	personType      PersonType
	checkDigitValid bool
}

// NewRFC validates and normalizes raw, inferring the person type from the
// length: 13 characters (4 name letters) is a persona física, 12 (3 name
// letters) a persona moral.
//
// Normalization trims surrounding whitespace, composes to NFC (so N +
// U+0303 becomes Ñ) and upper-cases only a-z and ñ, the same mapping as the
// rfc_hash blind index: strings.ToUpper would also turn U+0131 (dotless i)
// into "I" and U+017F (long s) into "S". Inner spaces and hyphens are not
// stripped. The use case hashes String(), never the raw input.
//
// Validation, in order:
//  1. Generic RFCs are ErrRFCGeneric (see genericRFCs).
//  2. The shape is 3 or 4 letters from A-Z, Ñ and &, six digits, two
//     homoclave characters from A-Z and 0-9 and a check digit from 0-9 and
//     A (the only values the algorithm yields; MODULE-taxpayers §2.2 and
//     the DB CHECK taxpayers_rfc_masked_format), else ErrRFCMalformed.
//  3. YYMMDD is a calendar date, else ErrRFCInvalidDate. The century is
//     not encoded, so YY is read as 20YY: 29 February is accepted whenever
//     YY is a multiple of 4. That is exact for 1901-2099; the one
//     over-acceptance is "000229" read as 1900, which was not a leap year.
//
// The last character is then compared with the SAT check digit, and the
// result is exposed as CheckDigitValid: a mismatch is NOT an error
// (decision D-RFC, ADR-0006).
//
// The check digit (dígito verificador) follows the SAT's "Algoritmo para
// generar el RFC con homoclave para personas físicas y morales", Anexo 3,
// released by the SAT under transparency request 0610100162005
// (https://www.infomex.org.mx/jspsi/documentos/2005/seguimiento/06101/0610100162005_065.doc):
// left-pad the first 11 (moral) or 12 (física) characters with blanks to
// 12; multiply each character's checkAlphabet value by its weight, 13 for
// the first position down to 2 for the twelfth; sum; take the remainder
// modulo 11. A remainder of 0 gives "0"; otherwise 11 minus the remainder
// is the digit, where 10 is written "A". The same algorithm is implemented
// by python-stdnum (stdnum.mx.rfc.calc_check_digit), which the tests use as
// an independent oracle for their fixtures.
//
// python-stdnum reports that about 1.5% of the RFCs the SAT has issued
// carry a check digit that does not match, and leaves the check off by
// default. Rejecting them would lock real taxpayers out, so the digit is a
// warning the caller may surface, never a validation failure.
func NewRFC(raw string) (RFC, error) {
	value := strings.Map(upperRFCRune, strings.TrimSpace(norm.NFC.String(raw)))
	if slices.Contains(genericRFCs, value) {
		return RFC{}, ErrRFCGeneric
	}
	runes := []rune(value)
	personType, nameLength, ok := layoutFor(len(runes))
	if !ok || !hasRFCShape(runes, nameLength) {
		return RFC{}, ErrRFCMalformed
	}
	if !isCalendarDate(runes[nameLength : nameLength+dateLength]) {
		return RFC{}, ErrRFCInvalidDate
	}
	last := len(runes) - 1
	return RFC{
		value:           value,
		personType:      personType,
		checkDigitValid: runes[last] == checkDigit(runes[:last]),
	}, nil
}

// NewRFCFor validates raw as NewRFC does and checks that its inferred person
// type is declared: ErrInvalidPersonType for an invalid declared type,
// ErrPersonTypeMismatch when the RFC belongs to the other type.
func NewRFCFor(declared PersonType, raw string) (RFC, error) {
	if !declared.Valid() {
		return RFC{}, ErrInvalidPersonType
	}
	rfc, err := NewRFC(raw)
	if err != nil {
		return RFC{}, err
	}
	if rfc.personType != declared {
		return RFC{}, ErrPersonTypeMismatch
	}
	return rfc, nil
}

// String returns the normalized RFC.
func (r RFC) String() string { return r.value }

// PersonType returns the person type implied by the RFC's length.
func (r RFC) PersonType() PersonType { return r.personType }

// CheckDigitValid reports whether the RFC's last character is the SAT check
// digit of the characters before it. False is a warning, not a validation
// failure (see NewRFC); the zero RFC reports false.
func (r RFC) CheckDigitValid() bool { return r.checkDigitValid }

// IsZero reports whether r is the zero RFC.
func (r RFC) IsZero() bool { return r.value == "" }

// Masked returns the RFC with its date hidden, as the contract's rfc_masked
// shows it: "GOMA******AB3". The zero RFC masks to "".
func (r RFC) Masked() string {
	if r.IsZero() {
		return ""
	}
	runes := []rune(r.value)
	nameLength := len(runes) - dateLength - homoclaveLength
	for i := nameLength; i < nameLength+dateLength; i++ {
		runes[i] = maskRune
	}
	return string(runes)
}

// layoutFor returns the person type and name length of an RFC of length
// runes, or false for any other length.
func layoutFor(length int) (PersonType, int, bool) {
	switch length {
	case fisicaLength:
		return PersonTypeFisica, fisicaNameLength, true
	case moralLength:
		return PersonTypeMoral, moralNameLength, true
	}
	return "", 0, false
}

// hasRFCShape reports whether runes are nameLength name letters, dateLength
// digits, a homoclave and a check digit.
func hasRFCShape(runes []rune, nameLength int) bool {
	dateEnd := nameLength + dateLength
	last := len(runes) - 1
	return allRunes(runes[:nameLength], isNameRune) &&
		allRunes(runes[nameLength:dateEnd], isDigit) &&
		allRunes(runes[dateEnd:last], isHomoclaveRune) &&
		isCheckDigitRune(runes[last])
}

func allRunes(runes []rune, valid func(rune) bool) bool {
	return !slices.ContainsFunc(runes, func(r rune) bool { return !valid(r) })
}

func isNameRune(r rune) bool { return isUpperASCII(r) || r == 'Ñ' || r == '&' }

func isHomoclaveRune(r rune) bool { return isUpperASCII(r) || isDigit(r) }

// isCheckDigitRune reports whether r is a value the check digit can take:
// 0-9, or A for 10 (see checkDigit).
func isCheckDigitRune(r rune) bool { return isDigit(r) || r == checkAlphabet[checkModulus-1] }

func isUpperASCII(r rune) bool { return r >= 'A' && r <= 'Z' }

func isDigit(r rune) bool { return r >= '0' && r <= '9' }

// upperRFCRune upper-cases only a-z and ñ (see NewRFC).
func upperRFCRune(r rune) rune {
	switch {
	case r >= 'a' && r <= 'z':
		return r - 'a' + 'A'
	case r == 'ñ':
		return 'Ñ'
	}
	return r
}

// Calendar constants for the YYMMDD part.
const (
	monthsPerYear = 12
	february      = 2
	leapCycle     = 4
)

// daysInMonth is indexed by month (1-12) in a common year.
var daysInMonth = [monthsPerYear + 1]int{0, 31, 28, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31}

// isCalendarDate reports whether the six digits YYMMDD are a date of 20YY
// (see NewRFC for the century rule).
func isCalendarDate(digits []rune) bool {
	year, month, day := twoDigits(digits[0:2]), twoDigits(digits[2:4]), twoDigits(digits[4:6])
	if month < 1 || month > monthsPerYear || day < 1 {
		return false
	}
	last := daysInMonth[month]
	if month == february && year%leapCycle == 0 {
		last++
	}
	return day <= last
}

// twoDigits returns the value of two ASCII digits.
func twoDigits(digits []rune) int {
	const base = 10
	return int(digits[0]-'0')*base + int(digits[1]-'0')
}

// checkDigit returns the SAT check digit of prefix, the RFC without its last
// character (see NewRFC). Every rune of prefix is in checkAlphabet.
func checkDigit(prefix []rune) rune {
	padding := checkedLength - len(prefix)
	blankValue := slices.Index(checkAlphabet, ' ')
	sum := 0
	for i := range checkedLength {
		value := blankValue
		if i >= padding {
			value = slices.Index(checkAlphabet, prefix[i-padding])
		}
		sum += value * (topWeight - i)
	}
	return checkAlphabet[(checkModulus-sum%checkModulus)%checkModulus]
}
