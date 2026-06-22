// Package route extracts HTTP endpoints from a Go project's source and maps
// each to the handler function that serves it. Each web framework has its own
// registration API, so detection is delegated to per-framework Extractors. The
// net/http extractor is the universal base.
package route

import (
	"go/types"

	"golang.org/x/tools/go/packages"
)

// Route is one detected HTTP endpoint.
type Route struct {
	Method  string      // GET/POST/... or "ANY"
	Path    string      // URL path (best-effort; "" when dynamic/unknown)
	Handler *types.Func // resolved handler func, or nil if unresolvable (e.g. inline closure)
}

// Extractor detects routes for a single framework.
type Extractor interface {
	Name() string
	// Match reports whether the package uses this framework (by imports).
	Match(pkg *packages.Package) bool
	// Extract walks the package syntax and returns its routes.
	Extract(pkg *packages.Package) []Route
}

// Default returns the built-in extractors (auto-detected per package).
func Default() []Extractor {
	return []Extractor{routerExtractor{}, netHTTPExtractor{}, gorillaExtractor{}, grpcExtractor{}, graphqlExtractor{}, connectExtractor{}}
}

// Extract runs the given extractors over all packages, returning every route
// found. An extractor only runs on a package it Match-es.
func Extract(pkgs []*packages.Package, extractors []Extractor) []Route {
	var out []Route
	for _, pkg := range pkgs {
		if pkg.TypesInfo == nil {
			continue
		}
		for _, ex := range extractors {
			if ex.Match(pkg) {
				out = append(out, ex.Extract(pkg)...)
			}
		}
	}
	return out
}
