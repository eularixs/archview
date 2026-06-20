package route

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/packages"
)

const httpPkg = "net/http"

// netHTTPExtractor detects routes registered via net/http:
//   - mux.HandleFunc(pattern, handler) on *http.ServeMux
//   - http.HandleFunc(pattern, handler) on the DefaultServeMux
//
// The Go 1.22+ method-prefixed pattern ("GET /path") is parsed for the method.
type netHTTPExtractor struct{}

func (netHTTPExtractor) Name() string { return "net/http" }

func (netHTTPExtractor) Match(pkg *packages.Package) bool { return imports(pkg, httpPkg) }

func (netHTTPExtractor) Extract(pkg *packages.Package) []Route {
	info := pkg.TypesInfo
	var out []Route
	for _, file := range pkg.Syntax {
		ast.Inspect(file, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			sel, ok := call.Fun.(*ast.SelectorExpr)
			if !ok || sel.Sel.Name != "HandleFunc" || len(call.Args) < 2 {
				return true
			}

			// Two shapes: receiver is *http.ServeMux, or the call is the
			// package-level http.HandleFunc (selector X is the "http" package).
			isMux := receiverIsType(info, sel.X, httpPkg, "ServeMux")
			isPkgFn := identIsPackage(info, sel.X, httpPkg)
			if !isMux && !isPkgFn {
				return true
			}

			method, path := splitMethodPath(stringLit(call.Args[0]))
			out = append(out, Route{
				Method:  method,
				Path:    path,
				Handler: handlerFunc(info, call.Args[1]),
			})
			return true
		})
	}
	return out
}

// identIsPackage reports whether expr is an identifier referring to an imported
// package with the given import path (e.g. the "http" in http.HandleFunc).
func identIsPackage(info *types.Info, expr ast.Expr, path string) bool {
	id, ok := expr.(*ast.Ident)
	if !ok {
		return false
	}
	pn, ok := info.ObjectOf(id).(*types.PkgName)
	return ok && pn.Imported() != nil && pn.Imported().Path() == path
}
