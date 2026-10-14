package params

// Error is a loading error. Its values are constants, matched with errors.Is.
type Error string

func (e Error) Error() string { return string(e) }

// Loading errors. A file's domain problems wrap the taxparams validation
// errors instead.
const (
	// ErrUnexpectedFile reports a file or directory other than
	// <YYYY>/v<N>.yaml and a top-level README.md.
	ErrUnexpectedFile Error = "params: unexpected file"
	// ErrPathMismatch reports a file whose fiscal_year or version differs
	// from its path.
	ErrPathMismatch Error = "params: fiscal year or version differs from the path"
	// ErrSyntax reports YAML that does not decode strictly into the schema:
	// malformed, empty, more than one document, or an unknown field.
	ErrSyntax Error = "params: invalid yaml"
	// ErrInexactNumber reports a rate or amount that is not a quoted decimal
	// string: an unquoted number would decode as a binary float first.
	ErrInexactNumber Error = "params: decimal must be a quoted string"
	// ErrValueKind reports a value with none, or more than one, of rate,
	// money and flag.
	ErrValueKind Error = "params: value needs exactly one of rate, money, flag"
	// ErrMissingFixedFee reports a marginal tariff row without fixed_fee.
	// Every published marginal row states its cuota fija (0.00 on the first
	// row), so an omitted one is a transcription slip, never a silent zero.
	ErrMissingFixedFee Error = "params: marginal tariff row needs fixed_fee"
	// ErrTableKind reports a table kind other than marginal or flat.
	ErrTableKind Error = "params: table kind must be marginal or flat"
)
