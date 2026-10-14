package archtest

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

// RepositoryViolation is one statement of a module repository that takes
// over the unit of work (P0-30 D-3, MODULE-taxpayers S-8).
type RepositoryViolation struct {
	File string // slash-separated, relative to the root
	Line int
	What string
}

// transactionCalls are the method names that begin, settle or nest a
// transaction (database/sql, GORM). A repository runs inside the use case's
// unit of work and must never call them: a commit there would end the
// transaction the use case still relies on, and RLS context with it.
var transactionCalls = []string{"Begin", "BeginTx", "Commit", "Rollback", "SavePoint", "RollbackTo", "Transaction"}

// sessionStatement matches SQL that changes session state: set_config
// anywhere, or a statement (at the start or after ";") that is SET
// [LOCAL|SESSION] name = / TO …, SET ROLE, SET SESSION AUTHORIZATION or
// RESET. "UPDATE … SET col =" does not match: that SET does not start a
// statement; nor does a label like "set archived_at" (no = or TO).
var sessionStatement = regexp.MustCompile(`(?i)\bset_config\s*\(|` +
	`(^|;)\s*(set\s+((local|session)\s+)?[a-z_][a-z0-9_.]*\s*(=|to\s)|set\s+(role|session\s+authorization)\s|reset\s+[a-z_])`)

// infraLayer is the repository layer of a module: internal/<module>/infra.
const infraLayer = "infra"

// CheckRepositories scans the non-test Go files of every module's infra
// layer under root (internal/<module>/infra/...) for calls that control a
// transaction and for SQL string literals that set session state. Platform
// stores that own their own pool (flagsgorm) are out of scope.
func CheckRepositories(root string) ([]RepositoryViolation, error) {
	var violations []RepositoryViolation
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return skipDir(root, path, d.Name())
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		if !isRepositoryFile(rel) {
			return nil
		}
		found, err := scanRepositoryFile(path, rel)
		violations = append(violations, found...)
		return err
	})
	if err != nil {
		return nil, fmt.Errorf("archtest: scan repositories: %w", err)
	}
	slices.SortStableFunc(violations, func(a, b RepositoryViolation) int {
		if c := strings.Compare(a.File, b.File); c != 0 {
			return c
		}
		return a.Line - b.Line
	})
	return violations, nil
}

// isRepositoryFile reports whether rel is a non-test Go file under
// internal/<module>/infra.
func isRepositoryFile(rel string) bool {
	if !strings.HasSuffix(rel, ".go") || strings.HasSuffix(rel, "_test.go") {
		return false
	}
	parts := strings.Split(rel, "/")
	return len(parts) >= 4 && parts[0] == internalDir && parts[1] != platformDir && parts[2] == infraLayer
}

func scanRepositoryFile(path, rel string) ([]RepositoryViolation, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, nil, parser.SkipObjectResolution)
	if err != nil {
		return nil, fmt.Errorf("archtest: parse %s: %w", rel, err)
	}
	var found []RepositoryViolation
	ast.Inspect(file, func(n ast.Node) bool {
		if what := repositoryBreach(n); what != "" {
			found = append(found, RepositoryViolation{File: rel, Line: fset.Position(n.Pos()).Line, What: what})
		}
		return true
	})
	return found, nil
}

// repositoryBreach names what n does wrong, or "" when it is fine.
func repositoryBreach(n ast.Node) string {
	switch node := n.(type) {
	case *ast.CallExpr:
		if selector, ok := node.Fun.(*ast.SelectorExpr); ok && slices.Contains(transactionCalls, selector.Sel.Name) {
			return "call " + selector.Sel.Name
		}
	case *ast.BasicLit:
		if node.Kind != token.STRING {
			return ""
		}
		if text, err := strconv.Unquote(node.Value); err == nil && sessionStatement.MatchString(text) {
			return "session statement in SQL"
		}
	}
	return ""
}
