package domain

import (
	"slices"
	"strings"
)

// cfdiUUIDLength is the length of a folio fiscal: 8-4-4-4-12 hex digits.
const cfdiUUIDLength = 36

// cfdiHyphens are the hyphen positions of a folio fiscal.
var cfdiHyphens = [...]int{8, 13, 18, 23}

// CFDIUUID is a CFDI folio fiscal normalized to upper case (contract
// cfdi_uuid). It is a business document id: never logged, never in an error.
type CFDIUUID struct {
	value string
}

// ParseCFDIUUID accepts 8-4-4-4-12 hex digits in either case, else
// ErrInvalid. It neither trims nor accepts braces.
func ParseCFDIUUID(raw string) (CFDIUUID, error) {
	if len(raw) != cfdiUUIDLength {
		return CFDIUUID{}, ErrInvalid
	}
	for i := range len(raw) {
		if !validUUIDByte(i, raw[i]) {
			return CFDIUUID{}, ErrInvalid
		}
	}
	return CFDIUUID{value: strings.ToUpper(raw)}, nil
}

func validUUIDByte(i int, c byte) bool {
	if slices.Contains(cfdiHyphens[:], i) {
		return c == '-'
	}
	return ('0' <= c && c <= '9') || ('a' <= c && c <= 'f') || ('A' <= c && c <= 'F')
}

// String returns the upper-case folio.
func (u CFDIUUID) String() string { return u.value }
