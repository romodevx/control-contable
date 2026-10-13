package archtest

import "strings"

// importGraph holds the imports of this module's own packages, so a rule can
// follow an import past the package that names it.
type importGraph struct {
	modulePath string
	imports    map[string][]string // package import path -> its imports (all files)
	// reaches memoizes reachesFramework. A package is marked false before
	// its imports are visited, which also ends any (invalid) import cycle.
	reaches map[string]bool
}

func newImportGraph(modulePath string, files []fileImports) *importGraph {
	g := &importGraph{
		modulePath: modulePath,
		imports:    make(map[string][]string),
		reaches:    make(map[string]bool),
	}
	for _, f := range files {
		g.imports[f.importer] = append(g.imports[f.importer], f.imports...)
	}
	return g
}

// check returns the rule importer → imported breaks: a per-import rule
// first, then, for an inner layer (domain, usecase), any framework that the
// imported package of this module reaches through its own imports. Packages
// outside the module are not followed (third-party internals are not ours to
// police here; the usecase-frameworks list names the ones that matter).
func (g *importGraph) check(importer, imported string) Rule {
	if rule := checkImport(g.modulePath, importer, imported); rule != "" {
		return rule
	}
	if isInnerLayer(locate(g.modulePath, importer)) && g.reachesFramework(imported) {
		return RuleInnerTransitiveFramework
	}
	return ""
}

func isInnerLayer(loc location) bool {
	return loc.kind == kindModule && (loc.layer == layerDomain || loc.layer == layerUsecase)
}

// reachesFramework reports whether pkg, a package of this module, imports a
// framework directly or through other packages of this module. O(V+E) over
// the module's packages thanks to the memo.
func (g *importGraph) reachesFramework(pkg string) bool {
	if !g.inModule(pkg) || g.selfPoliced(pkg) {
		return false
	}
	if reached, seen := g.reaches[pkg]; seen {
		return reached
	}
	g.reaches[pkg] = false
	for _, imported := range g.imports[pkg] {
		if isFramework(imported) || g.reachesFramework(imported) {
			g.reaches[pkg] = true
			return true
		}
	}
	return false
}

// selfPoliced reports a package whose own rules already keep frameworks out,
// transitively (domain, usecase, pure platform): a breach there is reported
// at its source, not again at every importer.
func (g *importGraph) selfPoliced(pkg string) bool {
	loc := locate(g.modulePath, pkg)
	return isInnerLayer(loc) || isPurePlatform(loc)
}

func (g *importGraph) inModule(pkg string) bool {
	return strings.HasPrefix(pkg, g.modulePath+"/")
}
