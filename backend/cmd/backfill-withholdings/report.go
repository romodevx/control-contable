package main

import (
	"fmt"
	"io"
	"slices"
	"strings"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
)

// tableReport is one table's outcome: counts and record ids only, never an
// amount, a description or a user id (MODULE-withholdings §5.2).
type tableReport struct {
	// Rederived are the records the run derives (or would, in a dry run).
	Rederived []string
	// AmountsChanged are those whose derived ISR or IVA withheld differs
	// from the stored one.
	AmountsChanged []string
	// FactsAssumed are those whose facts were assumed per §5.2: the user
	// should review them.
	FactsAssumed []string
	// NeedsInput are the records the rules refuse: left as stored.
	NeedsInput []string
}

// record adds one derived record.
func (t *tableReport) record(id string, amountsChanged, factsAssumed bool) {
	t.Rederived = append(t.Rederived, id)
	if amountsChanged {
		t.AmountsChanged = append(t.AmountsChanged, id)
	}
	if factsAssumed {
		t.FactsAssumed = append(t.FactsAssumed, id)
	}
}

// report is the whole run.
type report struct {
	Apply    bool
	Users    int
	Incomes  tableReport
	Expenses tableReport
}

// write prints r for the operator, ids sorted.
func (r report) write(w io.Writer) error {
	mode := "DRY RUN (nothing written; re-run with -apply to write)"
	if r.Apply {
		mode = "APPLIED"
	}
	var b strings.Builder
	fmt.Fprintf(&b, "withholdings backfill: %s\n", mode)
	fmt.Fprintf(&b, "table_version: %s\nusers visited: %d\n", fiscal.WithholdingTableVersion, r.Users)
	writeTable(&b, "incomes", r.Incomes)
	writeTable(&b, "expenses", r.Expenses)
	_, err := io.WriteString(w, b.String())
	return err
}

func writeTable(b *strings.Builder, name string, t tableReport) {
	fmt.Fprintf(b, "\n%s: %d to re-derive, %d with different amounts, %d with assumed facts, %d need user input\n",
		name, len(t.Rederived), len(t.AmountsChanged), len(t.FactsAssumed), len(t.NeedsInput))
	writeIDs(b, "different amounts", t.AmountsChanged)
	writeIDs(b, "assumed facts (ask the user to review)", t.FactsAssumed)
	writeIDs(b, "need user input (left as stored)", t.NeedsInput)
}

func writeIDs(b *strings.Builder, label string, ids []string) {
	if len(ids) == 0 {
		return
	}
	fmt.Fprintf(b, "  %s:\n", label)
	for _, id := range slices.Sorted(slices.Values(ids)) {
		fmt.Fprintf(b, "    %s\n", id)
	}
}
