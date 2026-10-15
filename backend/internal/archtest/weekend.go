package archtest

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
)

// WeekendViolation is one weekend test outside the fiscal kernel: a second
// copy of the due-date rule (ADR-0007 Amendment 1, MODULE-reminders §2.1).
type WeekendViolation struct {
	File string // slash-separated, relative to the root
	Line int
	What string
}

// weekendDays are the time constants a weekend test names.
var weekendDays = []string{"Saturday", "Sunday"}

// kernelPrefix is the one place the weekend test may live.
const kernelPrefix = internalDir + "/kernel/"

// CheckWeekendRule scans the non-test Go files under root, outside
// internal/kernel, for time.Saturday and time.Sunday (whatever the import
// is named): the CFF 12 move is fiscal.NextBusinessDay and nothing else.
func CheckWeekendRule(root string) ([]WeekendViolation, error) {
	var violations []WeekendViolation
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
		if !isProductionGoFile(d.Name()) || strings.HasPrefix(rel, kernelPrefix) {
			return nil
		}
		found, err := scanWeekendFile(path, rel)
		violations = append(violations, found...)
		return err
	})
	if err != nil {
		return nil, fmt.Errorf("archtest: scan weekend rule: %w", err)
	}
	slices.SortStableFunc(violations, func(a, b WeekendViolation) int {
		if c := strings.Compare(a.File, b.File); c != 0 {
			return c
		}
		return a.Line - b.Line
	})
	return violations, nil
}

func scanWeekendFile(path, rel string) ([]WeekendViolation, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, nil, parser.SkipObjectResolution)
	if err != nil {
		return nil, fmt.Errorf("archtest: parse %s: %w", rel, err)
	}
	timeName := importName(file, "time")
	if timeName == "" {
		return nil, nil
	}
	var found []WeekendViolation
	ast.Inspect(file, func(n ast.Node) bool {
		sel, ok := n.(*ast.SelectorExpr)
		if !ok || !slices.Contains(weekendDays, sel.Sel.Name) {
			return true
		}
		if pkg, ok := sel.X.(*ast.Ident); ok && pkg.Name == timeName {
			found = append(found, WeekendViolation{File: rel, Line: fset.Position(sel.Pos()).Line, What: "time." + sel.Sel.Name})
		}
		return true
	})
	return found, nil
}

// importName is the name file uses for the import path, or "" when it does
// not import it (a blank or dot import cannot name a selector).
func importName(file *ast.File, path string) string {
	for _, spec := range file.Imports {
		if p, err := strconv.Unquote(spec.Path.Value); err != nil || p != path {
			continue
		}
		if spec.Name == nil {
			return path[strings.LastIndex(path, "/")+1:]
		}
		return spec.Name.Name
	}
	return ""
}
