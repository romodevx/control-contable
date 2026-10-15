package infra

// ExplainExpenseSums exposes the reader's plan to the black-box tests.
var ExplainExpenseSums = explainExpenseSums

// The write allowlist (PR #136 note 1), for the black-box tests.
var (
	InsertStatement = insertStatement
	UpdateStatement = updateStatement
	EditableValues  = editable
)

// EditableColumnNames are the columns every write names.
func EditableColumnNames() []string {
	out := make([]string, len(editableColumns))
	for i, c := range editableColumns {
		out[i] = c.name
	}
	return out
}
