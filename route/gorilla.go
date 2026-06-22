package route

import (
	"go/ast"
	"go/types"
	"strings"

	"golang.org/x/tools/go/packages"
)

const gorillaPkg = "gorilla/mux"

// gorillaExtractor detects gorilla/mux routes: r.HandleFunc("/path", handler),
// with the HTTP method taken from a chained .Methods("GET", ...) when present
// (otherwise ANY). gorilla doesn't use the verb-method shape, so it needs its
// own extractor.
type gorillaExtractor struct{}

func (gorillaExtractor) Name() string { return "gorilla" }

func (gorillaExtractor) Match(pkg *packages.Package) bool {
	if pkg.Types == nil {
		return false
	}
	for _, imp := range pkg.Types.Imports() {
		if strings.Contains(imp.Path(), gorillaPkg) {
			return true
		}
	}
	return false
}

func (gorillaExtractor) Extract(pkg *packages.Package) []Route {
	info := pkg.TypesInfo
	if info == nil {
		return nil
	}
	byCall := map[*ast.CallExpr]*Route{}
	var order []*ast.CallExpr

	// Pass 1: HandleFunc registrations (default method ANY).
	for _, file := range pkg.Syntax {
		ast.Inspect(file, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			sel, ok := call.Fun.(*ast.SelectorExpr)
			if !ok || sel.Sel.Name != "HandleFunc" || len(call.Args) != 2 {
				return true
			}
			if !recvPkgContains(info, sel.X, gorillaPkg) {
				return true
			}
			byCall[call] = &Route{
				Method:  "ANY",
				Path:    stringLit(call.Args[0]),
				Handler: handlerFunc(info, call.Args[1]),
				Pos:     call.Pos(),
			}
			order = append(order, call)
			return true
		})
	}

	// Pass 2: a chained .Methods("GET") sets the route's method.
	for _, file := range pkg.Syntax {
		ast.Inspect(file, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			sel, ok := call.Fun.(*ast.SelectorExpr)
			if !ok || sel.Sel.Name != "Methods" || len(call.Args) == 0 {
				return true
			}
			hf, ok := sel.X.(*ast.CallExpr)
			if !ok {
				return true
			}
			if r := byCall[hf]; r != nil {
				if m := strings.ToUpper(stringLit(call.Args[0])); m != "" {
					r.Method = m
				}
			}
			return true
		})
	}

	out := make([]Route, 0, len(order))
	for _, c := range order {
		out = append(out, *byCall[c])
	}
	return out
}

// recvPkgContains reports whether expr's (named) type comes from a package whose
// import path contains sub.
func recvPkgContains(info *types.Info, expr ast.Expr, sub string) bool {
	n := namedType(info.TypeOf(expr))
	if n == nil || n.Obj() == nil || n.Obj().Pkg() == nil {
		return false
	}
	return strings.Contains(n.Obj().Pkg().Path(), sub)
}
