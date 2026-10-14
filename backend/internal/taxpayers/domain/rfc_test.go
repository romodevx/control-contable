package domain_test

import (
	"bytes"
	"fmt"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/platform/crypto"
	"github.com/romodevx/control-contable/backend/internal/taxpayers/domain"
)

// Every RFC below is synthetic: built from made-up name letters and dates,
// with its check digit computed by an independent implementation of the SAT
// algorithm (python-stdnum's mx.rfc.calc_check_digit), not by the code under
// test. None belongs to a real person or company.

func TestNewRFCAcceptsAndInfersPersonType(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want string
		typ  domain.PersonType
	}{
		{name: "fisica", raw: "PEXJ900215HK8", want: "PEXJ900215HK8", typ: domain.PersonTypeFisica},
		{name: "fisica with check digit 0 (remainder 0)", raw: "AAAA010101AA0", want: "AAAA010101AA0", typ: domain.PersonTypeFisica},
		{name: "fisica with Ñ in the name letters", raw: "ÑAÑO800101AB3", want: "ÑAÑO800101AB3", typ: domain.PersonTypeFisica},
		{name: "fisica born on 29 Feb 2000", raw: "MUÑA000229Z19", want: "MUÑA000229Z19", typ: domain.PersonTypeFisica},
		{name: "fisica on the last day of a 30-day month", raw: "XOXA850630Q11", want: "XOXA850630Q11", typ: domain.PersonTypeFisica},
		{name: "fisica on 31 Dec", raw: "ZZZZ991231ZZ8", want: "ZZZZ991231ZZ8", typ: domain.PersonTypeFisica},
		{name: "moral with &", raw: "A&B020304XY2", want: "A&B020304XY2", typ: domain.PersonTypeMoral},
		{name: "moral with leading Ñ", raw: "ÑEX991231B22", want: "ÑEX991231B22", typ: domain.PersonTypeMoral},
		{name: "moral with check digit 0", raw: "ABC1001017A0", want: "ABC1001017A0", typ: domain.PersonTypeMoral},
		{name: "moral with check digit A (remainder 1)", raw: "EJE840315K3A", want: "EJE840315K3A", typ: domain.PersonTypeMoral},
		{name: "moral with & in the middle", raw: "J&M120615R79", want: "J&M120615R79", typ: domain.PersonTypeMoral},
		{name: "moral on 3 Feb 2001", raw: "QAZ010203W96", want: "QAZ010203W96", typ: domain.PersonTypeMoral},
		{name: "lower case folded", raw: "pexj900215hk8", want: "PEXJ900215HK8", typ: domain.PersonTypeFisica},
		{name: "lower ñ folded to Ñ", raw: "ñaño800101ab3", want: "ÑAÑO800101AB3", typ: domain.PersonTypeFisica},
		{name: "decomposed N + combining tilde composed to Ñ", raw: "ÑEX991231B22", want: "ÑEX991231B22", typ: domain.PersonTypeMoral},
		{name: "surrounding whitespace trimmed", raw: " \t A&B020304XY2\n", want: "A&B020304XY2", typ: domain.PersonTypeMoral},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rfc, err := domain.NewRFC(tt.raw)
			require.NoError(t, err)
			assert.Equal(t, tt.want, rfc.String())
			assert.Equal(t, tt.typ, rfc.PersonType())
			assert.False(t, rfc.IsZero())
			assert.True(t, rfc.CheckDigitValid())
		})
	}
}

// TestNewRFCAcceptsACheckDigitMismatch pins decision D-RFC: about 1.5% of
// the RFCs the SAT has issued fail the published check-digit algorithm, so
// a mismatch is a warning the RFC carries (CheckDigitValid), never an error.
// Shape, charset and date errors still fail (TestNewRFCRejects).
func TestNewRFCAcceptsACheckDigitMismatch(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want string
		typ  domain.PersonType
	}{
		{name: "fisica check digit off by one", raw: "PEXJ900215HK9", want: "PEXJ900215HK9", typ: domain.PersonTypeFisica},
		{name: "fisica check digit A instead of 0", raw: "AAAA010101AAA", want: "AAAA010101AAA", typ: domain.PersonTypeFisica},
		{name: "moral check digit 0 instead of A", raw: "EJE840315K30", want: "EJE840315K30", typ: domain.PersonTypeMoral},
		{name: "moral check digit letter outside 0-9A", raw: "A&B020304XYB", want: "A&B020304XYB", typ: domain.PersonTypeMoral},
		{name: "transposed homoclave", raw: "PEXJ900215KH8", want: "PEXJ900215KH8", typ: domain.PersonTypeFisica},
		{name: "one name letter changed", raw: "PEXK900215HK8", want: "PEXK900215HK8", typ: domain.PersonTypeFisica},
		{name: "Ñ replaced by N changes the digit", raw: "NEX991231B22", want: "NEX991231B22", typ: domain.PersonTypeMoral},
		{name: "normalized before the digit is checked", raw: " pexj900215hk9 ", want: "PEXJ900215HK9", typ: domain.PersonTypeFisica},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rfc, err := domain.NewRFC(tt.raw)
			require.NoError(t, err)
			assert.Equal(t, tt.want, rfc.String())
			assert.Equal(t, tt.typ, rfc.PersonType())
			assert.False(t, rfc.CheckDigitValid())
		})
	}
}

func TestNewRFCRejects(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		wantErr error
	}{
		{name: "empty", raw: "", wantErr: domain.ErrRFCMalformed},
		{name: "only whitespace", raw: " \t ", wantErr: domain.ErrRFCMalformed},
		{name: "11 characters", raw: "ABC1001017A", wantErr: domain.ErrRFCMalformed},
		{name: "14 characters", raw: "PEXJ900215HK8X", wantErr: domain.ErrRFCMalformed},
		{name: "fisica without homoclave", raw: "PEXJ900215", wantErr: domain.ErrRFCMalformed},
		{name: "inner spaces are not stripped", raw: "PEXJ 900215 HK8", wantErr: domain.ErrRFCMalformed},
		{name: "hyphens are not stripped", raw: "A&B-020304-XY2", wantErr: domain.ErrRFCMalformed},
		{name: "digit in the name letters", raw: "PEX1900215HK8", wantErr: domain.ErrRFCMalformed},
		{name: "fisica with only 3 name letters", raw: "PEX0900215HK8", wantErr: domain.ErrRFCMalformed},
		{name: "moral with 4 name letters", raw: "ABCD001017A0", wantErr: domain.ErrRFCMalformed},
		{name: "letter in the date", raw: "PEXJ9002I5HK8", wantErr: domain.ErrRFCMalformed},
		{name: "& in the homoclave", raw: "PEXJ900215H&8", wantErr: domain.ErrRFCMalformed},
		{name: "Ñ in the homoclave", raw: "PEXJ900215HÑ8", wantErr: domain.ErrRFCMalformed},
		{name: "accented letter", raw: "PÉXJ900215HK8", wantErr: domain.ErrRFCMalformed},
		{name: "dotless i would fold to I", raw: "PEXı900215HK8", wantErr: domain.ErrRFCMalformed},
		{name: "long s would fold to S", raw: "ſEXJ900215HK8", wantErr: domain.ErrRFCMalformed},
		{name: "invalid UTF-8", raw: "PEXJ900215H\xff8", wantErr: domain.ErrRFCMalformed},
		{name: "month 00", raw: "PEXJ900015HK8", wantErr: domain.ErrRFCInvalidDate},
		{name: "month 13", raw: "PEXJ901315HK8", wantErr: domain.ErrRFCInvalidDate},
		{name: "day 00", raw: "PEXJ900200HK8", wantErr: domain.ErrRFCInvalidDate},
		{name: "31 April", raw: "PEXJ900431HK8", wantErr: domain.ErrRFCInvalidDate},
		{name: "32 January", raw: "A&B020132XY2", wantErr: domain.ErrRFCInvalidDate},
		{name: "29 Feb of a non-leap year", raw: "PEXJ990229HK8", wantErr: domain.ErrRFCInvalidDate},
		{name: "30 Feb", raw: "MUÑA000230Z19", wantErr: domain.ErrRFCInvalidDate},
		{name: "bad date with a wrong check digit is still a date error", raw: "PEXJ900431HK9", wantErr: domain.ErrRFCInvalidDate},
		{name: "bad shape with a wrong check digit is still malformed", raw: "PEX1900215HK9", wantErr: domain.ErrRFCMalformed},
		{name: "generic RFC for the general public", raw: "XAXX010101000", wantErr: domain.ErrRFCGeneric},
		{name: "generic RFC for foreign residents", raw: "XEXX010101000", wantErr: domain.ErrRFCGeneric},
		{name: "generic RFC in lower case with spaces", raw: " xaxx010101000 ", wantErr: domain.ErrRFCGeneric},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rfc, err := domain.NewRFC(tt.raw)
			require.ErrorIs(t, err, tt.wantErr)
			assert.True(t, rfc.IsZero())
		})
	}
}

// TestNewRFCErrorsNeverCarryTheInput: an RFC is personal data for a persona
// física, so no error text may echo it.
func TestNewRFCErrorsNeverCarryTheInput(t *testing.T) {
	for _, raw := range []string{"PEXJ9002I5HK8", "PEXJ900431HK8", "XAXX010101000", "PEXJ900215"} {
		_, err := domain.NewRFC(raw)
		require.Error(t, err)
		assert.NotContains(t, err.Error(), "PEXJ")
		assert.NotContains(t, err.Error(), "XAXX")
	}
}

func TestNewRFCForChecksTheDeclaredPersonType(t *testing.T) {
	tests := []struct {
		name     string
		declared domain.PersonType
		raw      string
		wantErr  error
	}{
		{name: "fisica declared fisica", declared: domain.PersonTypeFisica, raw: "PEXJ900215HK8"},
		{name: "moral declared moral", declared: domain.PersonTypeMoral, raw: "A&B020304XY2"},
		{name: "moral RFC declared fisica", declared: domain.PersonTypeFisica, raw: "A&B020304XY2", wantErr: domain.ErrPersonTypeMismatch},
		{name: "fisica RFC declared moral", declared: domain.PersonTypeMoral, raw: "PEXJ900215HK8", wantErr: domain.ErrPersonTypeMismatch},
		{name: "unknown declared type", declared: domain.PersonType("extranjero"), raw: "PEXJ900215HK8", wantErr: domain.ErrInvalidPersonType},
		{name: "zero declared type", declared: domain.PersonType(""), raw: "PEXJ900215HK8", wantErr: domain.ErrInvalidPersonType},
		{name: "invalid RFC reports the RFC error", declared: domain.PersonTypeFisica, raw: "PEXJ900431HK8", wantErr: domain.ErrRFCInvalidDate},
		{name: "check digit mismatch is no error", declared: domain.PersonTypeFisica, raw: "PEXJ900215HK9"},
		{name: "check digit mismatch still checks the type", declared: domain.PersonTypeMoral, raw: "PEXJ900215HK9", wantErr: domain.ErrPersonTypeMismatch},
		{name: "generic RFC rejected whatever the type", declared: domain.PersonTypeFisica, raw: "XAXX010101000", wantErr: domain.ErrRFCGeneric},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rfc, err := domain.NewRFCFor(tt.declared, tt.raw)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				assert.True(t, rfc.IsZero())
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.declared, rfc.PersonType())
		})
	}
}

func TestRFCMasked(t *testing.T) {
	tests := []struct {
		raw  string
		want string
	}{
		{raw: "PEXJ900215HK8", want: "PEXJ******HK8"},
		{raw: "A&B020304XY2", want: "A&B******XY2"},
		{raw: "ÑAÑO800101AB3", want: "ÑAÑO******AB3"},
	}
	for _, tt := range tests {
		t.Run(tt.raw, func(t *testing.T) {
			rfc, err := domain.NewRFC(tt.raw)
			require.NoError(t, err)
			assert.Equal(t, tt.want, rfc.Masked())
		})
	}
}

func TestRFCZeroValue(t *testing.T) {
	var zero domain.RFC
	assert.True(t, zero.IsZero())
	assert.Empty(t, zero.String())
	assert.Empty(t, zero.Masked())
	assert.False(t, zero.PersonType().Valid())
	assert.False(t, zero.CheckDigitValid())
}

func TestRFCNeverPrintsTheValue(t *testing.T) {
	rfc, err := domain.NewRFC("PEXJ900215HK8")
	require.NoError(t, err)
	for _, verb := range []string{"%v", "%+v", "%#v", "%s", "%q"} {
		assert.NotContains(t, fmt.Sprintf(verb, rfc), "PEXJ", verb)
		assert.NotContains(t, fmt.Sprintf(verb, rfc), "900215", verb)
	}
	var buf bytes.Buffer
	slog.New(slog.NewTextHandler(&buf, nil)).Info("x", "rfc", rfc)
	assert.NotContains(t, buf.String(), "PEXJ")
}

// TestRFCMatchesBlindIndexNormalization pins the contract with the rfc_hash
// blind index: for any precomposed input, hashing the raw value and hashing
// RFC.String() give the same index, so a duplicate lookup always finds the
// row written from the user's raw input.
func TestRFCMatchesBlindIndexNormalization(t *testing.T) {
	index, err := crypto.NewRFCIndex(bytes.Repeat([]byte{7}, crypto.MinPepperSize))
	require.NoError(t, err)
	for _, raw := range []string{"PEXJ900215HK8", " pexj900215hk8\t", "ñaño800101ab3", "a&b020304xy2"} {
		rfc, err := domain.NewRFC(raw)
		require.NoError(t, err)
		fromRaw, err := index.Hash(raw)
		require.NoError(t, err)
		fromRFC, err := index.Hash(rfc.String())
		require.NoError(t, err)
		assert.Equal(t, fromRaw, fromRFC, raw)
	}
}
