package contracttest

import (
	"bytes"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// importPath is this package's import path, as test files import it.
const importPath = "github.com/romodevx/control-contable/backend/adapter/http/contracttest"

// opFunc is the constructor whose literal argument marks a covered operation.
const opFunc = "Op"

// ScanCoverage parses every Go file under root (any build tag) and returns,
// per operationId, the call sites ("file:line") of contracttest.Op with that
// literal. Directories named testdata or starting with "." or "_" are skipped
// (as the go tool does), and so is every directory in exclude. A call whose
// argument is not a string literal is an error: coverage must be readable from
// the source.
func ScanCoverage(root string, exclude ...string) (map[string][]string, error) {
	skip := make(map[string]bool, len(exclude))
	for _, dir := range exclude {
		abs, err := filepath.Abs(dir)
		if err != nil {
			return nil, fmt.Errorf("exclude %s: %w", dir, err)
		}
		skip[abs] = true
	}
	covered := map[string][]string{}
	var problems []error
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return walkDir(path, path != root && ignoredDir(d.Name()), skip)
		}
		if !strings.HasSuffix(path, ".go") {
			return nil
		}
		problems = append(problems, scanFile(path, covered)...)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("scan %s: %w", root, err)
	}
	if err := errors.Join(problems...); err != nil {
		return nil, err
	}
	return covered, nil
}

// walkDir decides whether WalkDir descends into dir.
func walkDir(dir string, ignored bool, skip map[string]bool) error {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return err
	}
	if ignored || skip[abs] {
		return filepath.SkipDir
	}
	return nil
}

func ignoredDir(name string) bool {
	return name == "testdata" || name == "vendor" || strings.HasPrefix(name, ".") || strings.HasPrefix(name, "_")
}

// scanFile adds the Op literals of one file to covered.
func scanFile(path string, covered map[string][]string) []error {
	src, err := os.ReadFile(path) // #nosec G304 -- walking the module's own source tree
	if err != nil {
		return []error{err}
	}
	if !bytes.Contains(src, []byte(importPath)) {
		return nil
	}
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, src, parser.SkipObjectResolution)
	if err != nil {
		return []error{fmt.Errorf("parse %s: %w", path, err)}
	}
	alias, ok := importName(file)
	if !ok {
		return nil
	}
	var problems []error
	ast.Inspect(file, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok || !isOpCall(call, alias) {
			return true
		}
		pos := fset.Position(call.Pos())
		site := fmt.Sprintf("%s:%d", pos.Filename, pos.Line)
		id, ok := literalArg(call)
		if !ok {
			problems = append(problems, fmt.Errorf("%s: %s.%s needs a string literal operationId", site, alias, opFunc))
			return true
		}
		covered[id] = append(covered[id], site)
		return true
	})
	return problems
}

// importName returns the name the file uses for this package, if it imports it.
func importName(file *ast.File) (string, bool) {
	for _, spec := range file.Imports {
		path, err := strconv.Unquote(spec.Path.Value)
		if err != nil || path != importPath {
			continue
		}
		if spec.Name == nil {
			return filepath.Base(importPath), true
		}
		if spec.Name.Name == "_" {
			return "", false
		}
		return spec.Name.Name, true
	}
	return "", false
}

func isOpCall(call *ast.CallExpr, alias string) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok || sel.Sel.Name != opFunc {
		return false
	}
	pkg, ok := sel.X.(*ast.Ident)
	return ok && pkg.Name == alias
}

func literalArg(call *ast.CallExpr) (string, bool) {
	if len(call.Args) != 1 {
		return "", false
	}
	lit, ok := call.Args[0].(*ast.BasicLit)
	if !ok || lit.Kind != token.STRING {
		return "", false
	}
	id, err := strconv.Unquote(lit.Value)
	return id, err == nil
}
