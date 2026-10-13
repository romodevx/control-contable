package archtest

import (
	"errors"
	"fmt"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
)

// Rule names one Dependency Rule. A Violation carries the rule it breaks.
type Rule string

// The rules enforced by Check. An empty Rule means the import is allowed.
const (
	RuleDomainImports          Rule = "domain-imports"
	RuleUsecaseFrameworks      Rule = "usecase-frameworks"
	RuleUsecaseOuterLayer      Rule = "usecase-outer-layer"
	RuleUsecasePlatformAdapter Rule = "usecase-platform-adapter"
	RuleCrossModule            Rule = "cross-module"
	RulePlatformModules        Rule = "platform-modules"
	RulePurePlatformImports    Rule = "pure-platform-imports"
)

// Violation is one forbidden import found in a non-test Go file.
type Violation struct {
	File     string // slash-separated, relative to the module root
	Imported string
	Rule     Rule
}

// Inner layer names, as directories under internal/<module>/. The outer layers
// (adapter, infra) need no constant: the rules only name what is allowed inward.
const (
	layerDomain  = "domain"
	layerUsecase = "usecase"
)

const (
	internalDir = "internal"
	platformDir = "platform"
)

// domainPlatformAllowlist lists the pure platform packages a domain may import.
// They are held to the domain's own import rule, so the domain cannot reach a
// framework through them.
var domainPlatformAllowlist = []string{"money"}

// platformAdapters are the platform packages that wrap a framework (the
// database adapter and its test harness). Use cases reach them only through
// framework-free ports (internal/platform/tx); infra and cmd/ wire them.
var platformAdapters = []string{"pgdb", "pgtest"}

// frameworkPrefixes are delivery and persistence details that domain and
// usecase packages must never see.
var frameworkPrefixes = []string{
	"net/http",
	"database/sql",
	"gorm.io",
	"github.com/jackc",
}

type kind int

const (
	kindExternal kind = iota // outside this module's internal/ tree (stdlib, third-party, cmd/)
	kindPlatform
	kindModule
)

// location is where a package sits in the backend layout.
type location struct {
	kind   kind
	module string // module name for kindModule, sub-package name for kindPlatform
	layer  string // first directory under the module; empty for kindPlatform
}

func locate(modulePath, pkg string) location {
	rel, ok := strings.CutPrefix(pkg, modulePath+"/")
	if !ok {
		return location{kind: kindExternal}
	}
	parts := strings.Split(rel, "/")
	if len(parts) < 2 || parts[0] != internalDir {
		return location{kind: kindExternal}
	}
	if parts[1] == platformDir {
		return location{kind: kindPlatform, module: elementAt(parts, 2)}
	}
	return location{kind: kindModule, module: parts[1], layer: elementAt(parts, 2)}
}

func elementAt(parts []string, i int) string {
	if i < len(parts) {
		return parts[i]
	}
	return ""
}

// checkImport returns the rule that importer → imported breaks, or "" if allowed.
func checkImport(modulePath, importer, imported string) Rule {
	from := locate(modulePath, importer)
	to := locate(modulePath, imported)
	switch from.kind {
	case kindPlatform:
		return checkPlatformImport(from, to, imported)
	case kindModule:
		return checkModuleImport(from, to, imported)
	}
	return ""
}

func checkModuleImport(from, to location, imported string) Rule {
	if from.layer == layerDomain {
		return checkDomainImport(from, to, imported)
	}
	if to.kind == kindModule && to.module != from.module {
		return RuleCrossModule
	}
	if from.layer != layerUsecase {
		return ""
	}
	if isFramework(imported) {
		return RuleUsecaseFrameworks
	}
	if to.kind == kindModule && to.layer != layerDomain && to.layer != layerUsecase {
		return RuleUsecaseOuterLayer
	}
	if to.kind == kindPlatform && slices.Contains(platformAdapters, to.module) {
		return RuleUsecasePlatformAdapter
	}
	return ""
}

// checkPlatformImport keeps platform free of modules, and keeps the pure
// platform packages (domainPlatformAllowlist) as clean as the domain itself.
func checkPlatformImport(from, to location, imported string) Rule {
	if to.kind == kindModule {
		return RulePlatformModules
	}
	if !isPurePlatform(from) {
		return ""
	}
	if to.kind == kindExternal && isPureStdlib(imported) {
		return ""
	}
	if to.kind == kindPlatform && isPurePlatform(to) {
		return ""
	}
	return RulePurePlatformImports
}

// checkDomainImport allows only the standard library (minus frameworks), the
// module's own domain tree and the allow-listed pure platform packages.
func checkDomainImport(from, to location, imported string) Rule {
	switch to.kind {
	case kindExternal:
		if isPureStdlib(imported) {
			return ""
		}
	case kindModule:
		if to.module == from.module && to.layer == layerDomain {
			return ""
		}
	case kindPlatform:
		if isPurePlatform(to) {
			return ""
		}
	}
	return RuleDomainImports
}

func isPurePlatform(loc location) bool {
	return loc.kind == kindPlatform && slices.Contains(domainPlatformAllowlist, loc.module)
}

// isPureStdlib reports a standard-library package that is not a framework.
func isPureStdlib(path string) bool {
	return isStdlib(path) && !isFramework(path)
}

func isStdlib(path string) bool {
	first, _, _ := strings.Cut(path, "/")
	return !strings.Contains(first, ".")
}

func isFramework(path string) bool {
	return slices.ContainsFunc(frameworkPrefixes, func(prefix string) bool {
		return path == prefix || strings.HasPrefix(path, prefix+"/")
	})
}

// Check walks every non-test Go file under root (skipping testdata, vendor and
// hidden directories) and returns its Dependency Rule violations, sorted by
// file and import path.
func Check(root, modulePath string) ([]Violation, error) {
	var violations []Violation
	fset := token.NewFileSet()
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return skipDir(root, path, d.Name())
		}
		if !isProductionGoFile(d.Name()) {
			return nil
		}
		found, err := checkFile(fset, root, path, modulePath)
		violations = append(violations, found...)
		return err
	})
	if err != nil {
		return nil, fmt.Errorf("archtest: walk %s: %w", root, err)
	}
	slices.SortFunc(violations, func(a, b Violation) int {
		if c := strings.Compare(a.File, b.File); c != 0 {
			return c
		}
		return strings.Compare(a.Imported, b.Imported)
	})
	return violations, nil
}

func skipDir(root, path, name string) error {
	if path == root {
		return nil
	}
	if name == "testdata" || name == "vendor" || strings.HasPrefix(name, ".") || strings.HasPrefix(name, "_") {
		return filepath.SkipDir
	}
	return nil
}

func isProductionGoFile(name string) bool {
	return strings.HasSuffix(name, ".go") && !strings.HasSuffix(name, "_test.go")
}

func checkFile(fset *token.FileSet, root, path, modulePath string) ([]Violation, error) {
	file, err := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
	if err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return nil, fmt.Errorf("relative path of %s: %w", path, err)
	}
	rel = filepath.ToSlash(rel)
	importer := modulePath
	if dir := filepath.ToSlash(filepath.Dir(rel)); dir != "." {
		importer = modulePath + "/" + dir
	}
	var violations []Violation
	for _, spec := range file.Imports {
		imported, err := strconv.Unquote(spec.Path.Value)
		if err != nil {
			return nil, fmt.Errorf("import path in %s: %w", path, err)
		}
		if rule := checkImport(modulePath, importer, imported); rule != "" {
			violations = append(violations, Violation{File: rel, Imported: imported, Rule: rule})
		}
	}
	return violations, nil
}

// ErrNoModuleDirective is returned when a go.mod file has no module line.
var ErrNoModuleDirective = errors.New("archtest: go.mod has no module directive")

// ReadModulePath returns the module path declared in the given go.mod file.
func ReadModulePath(goMod string) (string, error) {
	data, err := os.ReadFile(filepath.Clean(goMod))
	if err != nil {
		return "", fmt.Errorf("archtest: read go.mod: %w", err)
	}
	for line := range strings.Lines(string(data)) {
		if rest, ok := strings.CutPrefix(strings.TrimSpace(line), "module "); ok {
			return strings.Trim(strings.TrimSpace(rest), `"`), nil
		}
	}
	return "", ErrNoModuleDirective
}
