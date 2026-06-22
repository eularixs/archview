package route

import (
	"go/ast"
	"go/types"
	"strings"

	"golang.org/x/tools/go/packages"
)

// echoPkg is matched as a substring so both echo v3 and v4 import paths work.
const echoPkg = "labstack/echo"

// echoExtractor detects routes on *echo.Echo / *echo.Group. In echo the handler
// is the argument right after the path (e.GET(path, handler, middleware...)).
type echoExtractor struct{}

func (echoExtractor) Name() string { return "echo" }

func (echoExtractor) Match(pkg *packages.Package) bool {
	if pkg.Types == nil {
		return false
	}
	for _, imp := range pkg.Types.Imports() {
		if strings.Contains(imp.Path(), echoPkg) {
			return true
		}
	}
	return false
}

var echoVerbs = map[string]bool{
	"GET": true, "POST": true, "PUT": true, "DELETE": true,
	"PATCH": true, "HEAD": true, "OPTIONS": true, "CONNECT": true, "TRACE": true,
}

func echoRecv(info *types.Info, x ast.Expr) bool {
	t := info.TypeOf(x)
	if t == nil {
		return false
	}
	if p, ok := t.(*types.Pointer); ok {
		t = p.Elem()
	}
	n, ok := t.(*types.Named)
	if !ok || n.Obj() == nil || n.Obj().Pkg() == nil {
		return false
	}
	if !strings.Contains(n.Obj().Pkg().Path(), echoPkg) {
		return false
	}
	switch n.Obj().Name() {
	case "Echo", "Group":
		return true
	}
	return false
}

func (echoExtractor) Extract(pkg *packages.Package) []Route {
	info := pkg.TypesInfo
	prefixes := groupPrefixes(pkg, echoRecv)
	var out []Route
	for _, file := range pkg.Syntax {
		ast.Inspect(file, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok || len(call.Args) < 2 {
				return true
			}
			sel, ok := call.Fun.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			name := sel.Sel.Name
			if !echoVerbs[name] && name != "Any" {
				return true
			}
			if !echoRecv(info, sel.X) {
				return true
			}
			method := name
			if name == "Any" {
				method = "ANY"
			}
			out = append(out, Route{
				Method:  method,
				Path:    joinPath(groupPrefix(info, sel.X, prefixes), stringLit(call.Args[0])),
				Handler: handlerFunc(info, call.Args[1]), // handler follows the path
			})
			return true
		})
	}
	return out
}
