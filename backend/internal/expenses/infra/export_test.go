package infra

// ExplainExpenseSums exposes the reader's plan to the black-box tests.
var ExplainExpenseSums = explainExpenseSums

// LedgerRange exposes the paid_on range [from, until) a LedgerQuery reads.
func LedgerRange(q LedgerQuery) (from, until string, err error) { return q.dates() }
