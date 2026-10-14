package golden

import (
	"cmp"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

// Violation is one way the source tree could rewrite goldens.
type Violation struct {
	File string
	Line int
	What string
}

func (v Violation) String() string { return fmt.Sprintf("%s:%d: %s", v.File, v.Line, v.What) }

// rewriteKnob matches a flag or environment variable name that would let a
// run regenerate expected values ("-update", "UPDATE_GOLDEN", "-accept").
var rewriteKnob = regexp.MustCompile(`(?i)(update|regen|rewrite|overwrite|golden|accept|record)`)

// fileWriters are the calls that create, change or remove files.
var fileWriters = map[string][]string{
	"os": {
		"Chmod", "Create", "CreateTemp", "Link", "Mkdir", "MkdirAll", "MkdirTemp", "OpenFile",
		"Remove", "RemoveAll", "Rename", "Symlink", "Truncate", "WriteFile",
	},
	"io/ioutil": {"TempDir", "TempFile", "WriteFile"},
}

// FindRewriters scans every Go file under fsys (testdata, vendor and
// hidden directories excluded, as the go tool does) for what could
// regenerate a golden: a flag or environment variable named like an update
// knob, anywhere, and any file-writing call in a file under writeScope (the
// tax engine's tree). Goldens are hand-authored from official examples;
// nothing may write them (MODULE-tax-engine US-X3).
func FindRewriters(fsys fs.FS, writeScope string) ([]Violation, error) {
	var out []Violation
	err := fs.WalkDir(fsys, ".", func(name string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			if name != "." && ignoredDir(path.Base(name)) {
				return fs.SkipDir
			}
			return nil
		}
		if path.Ext(name) != ".go" {
			return nil
		}
		vs, err := scanFile(fsys, name, strings.HasPrefix(name, writeScope))
		out = append(out, vs...)
		return err
	})
	if err != nil {
		return nil, fmt.Errorf("golden: scan: %w", err)
	}
	slices.SortFunc(out, func(a, b Violation) int { return cmp.Or(cmp.Compare(a.File, b.File), cmp.Compare(a.Line, b.Line)) })
	return out, nil
}

func ignoredDir(base string) bool {
	return base == "testdata" || base == "vendor" || strings.HasPrefix(base, ".") || strings.HasPrefix(base, "_")
}

func scanFile(fsys fs.FS, name string, inScope bool) ([]Violation, error) {
	data, err := fs.ReadFile(fsys, name)
	if err != nil {
		return nil, err
	}
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, name, data, parser.SkipObjectResolution)
	if err != nil {
		return nil, err
	}
	imports := importNames(file)
	var out []Violation
	ast.Inspect(file, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		pkg, ok := sel.X.(*ast.Ident)
		if !ok {
			return true
		}
		if what := judge(imports[pkg.Name], sel.Sel.Name, call.Args, inScope); what != "" {
			out = append(out, Violation{File: name, Line: fset.Position(call.Pos()).Line, What: what})
		}
		return true
	})
	return out, nil
}

// importNames maps each import's local name to its path.
func importNames(file *ast.File) map[string]string {
	names := map[string]string{}
	for _, imp := range file.Imports {
		p, err := strconv.Unquote(imp.Path.Value)
		if err != nil {
			continue
		}
		local := path.Base(p)
		if imp.Name != nil {
			local = imp.Name.Name
		}
		names[local] = p
	}
	return names
}

// judge describes why pkgPath.fn(args) could rewrite goldens, or "".
func judge(pkgPath, fn string, args []ast.Expr, inScope bool) string {
	switch {
	case pkgPath == "flag":
		if name, ok := literalAt(args, flagNameArg(fn)); ok && rewriteKnob.MatchString(name) {
			return fmt.Sprintf("flag %q can rewrite goldens", name)
		}
	case pkgPath == "os" && (fn == "Getenv" || fn == "LookupEnv"):
		if name, ok := literalAt(args, 0); ok && rewriteKnob.MatchString(name) {
			return fmt.Sprintf("environment variable %q can rewrite goldens", name)
		}
	}
	if inScope && slices.Contains(fileWriters[pkgPath], fn) {
		return fmt.Sprintf("%s.%s writes files under internal/taxengine/", path.Base(pkgPath), fn)
	}
	return ""
}

// flagNameArg is the index of the flag name: flag.BoolVar(&v, name, ...)
// and the other *Var functions take a pointer first.
func flagNameArg(fn string) int {
	if strings.HasSuffix(fn, "Var") {
		return 1
	}
	return 0
}

func literalAt(args []ast.Expr, i int) (string, bool) {
	if i >= len(args) {
		return "", false
	}
	lit, ok := args[i].(*ast.BasicLit)
	if !ok || lit.Kind != token.STRING {
		return "", false
	}
	s, err := strconv.Unquote(lit.Value)
	return s, err == nil
}
