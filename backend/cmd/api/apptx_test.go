package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// txManagerCall is one pgdb.NewTxManager call in this package's sources.
type txManagerCall struct {
	file, function string
}

// txManagerCalls finds every pgdb.NewTxManager call in the non-test sources
// of this package, with the function that makes it.
func txManagerCalls(t *testing.T) []txManagerCall {
	t.Helper()
	fset := token.NewFileSet()
	entries, err := os.ReadDir(".")
	require.NoError(t, err)
	var calls []txManagerCall
	for _, entry := range entries {
		name := entry.Name()
		if !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		file, err := parser.ParseFile(fset, name, nil, 0)
		require.NoError(t, err)
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Body == nil {
				continue
			}
			ast.Inspect(fn.Body, func(n ast.Node) bool {
				if isTxManagerCall(n) {
					calls = append(calls, txManagerCall{file: name, function: fn.Name.Name})
				}
				return true
			})
		}
	}
	return calls
}

func isTxManagerCall(n ast.Node) bool {
	call, ok := n.(*ast.CallExpr)
	if !ok {
		return false
	}
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok || sel.Sel.Name != "NewTxManager" {
		return false
	}
	pkg, ok := sel.X.(*ast.Ident)
	return ok && pkg.Name == "pgdb"
}

// The composition root opens ONE cc_app unit of work and hands it to every
// module: a port that joins its caller's transaction (CheckRegime, the
// access policy, the estimate's profile read) works only through the same
// manager (tx.ErrForeignTransaction otherwise), so a module building its own
// would fail at the first nested call instead of at wiring.
func TestCompositionRoot_opensOneAppUnitOfWork(t *testing.T) {
	assert.Equal(t, []txManagerCall{{file: "main.go", function: "run"}}, txManagerCalls(t))
}

// A module wired without the shared unit of work fails at boot.
func TestModules_requireTheSharedUnitOfWork(t *testing.T) {
	_, authErr := newAuthModule(authWiring{})
	_, taxpayersErr := newTaxpayersDeps(taxpayersWiring{})

	assert.ErrorIs(t, authErr, errNoAppTx)
	assert.ErrorIs(t, taxpayersErr, errNoAppTx)
}
