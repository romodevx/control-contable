package infra

// ExplainExpenseSums exposes the reader's plan to the black-box tests.
var ExplainExpenseSums = explainExpenseSums

// The write allowlist (PR #136 note 1), for the black-box tests.
var (
	InsertStatement = insertStatement
	UpdateStatement = updateStatement
	EditableValues  = editable
)

// The backfill's write (WH-08), for the black-box tests.
var (
	DerivationStatement = derivationStatement
	DerivationArgs      = derivationArgs
)

// EditableColumnNames are the columns every write names.
func EditableColumnNames() []string {
	out := make([]string, len(editableColumns))
	for i, c := range editableColumns {
		out[i] = c.name
	}
	return out
}

// LedgerRange exposes the paid_on range [from, until) a LedgerQuery reads.
func LedgerRange(q LedgerQuery) (from, until string, err error) { return q.dates() }

// ExpenseRow and RowWithholding expose the read of a stored derivation.
type ExpenseRow = expenseRow

// RowWithholding is expenseRow.withholding.
var RowWithholding = expenseRow.withholding
