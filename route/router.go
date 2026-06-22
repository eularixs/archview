package route

import (
	"go/ast"
	"go/types"
	"strings"

	"golang.org/x/tools/go/packages"
)

// httpVerbMethods maps a router method name (any case) to an HTTP method.
// Covers the GET/POST/... verbs plus the "all/any" catch-alls.
var httpVerbMethods = map[string]string{
	"GET": "GET", "POST": "POST", "PUT": "PUT", "DELETE": "DELETE",
	"PATCH": "PATCH", "HEAD": "HEAD", "OPTIONS": "OPTIONS", "CONNECT": "CONNECT",
	"TRACE": "TRACE", "ALL": "ANY", "ANY": "ANY",
}

// routerExtractor detects the route shape shared by virtually every Go HTTP
// router — router.GET("/path", handler) — used by gin, echo, fiber, chi,
// httprouter and others. It is framework-agnostic: a route is matched by the
// verb method name, a string path, and a function-typed handler argument,
// not by the router's concrete type. Group("/prefix") prefixes are joined.
type routerExtractor struct{}

func (routerExtractor) Name() string { return "router" }

func (routerExtractor) Match(pkg *packages.Package) bool { return len(routerRoutes(pkg)) > 0 }

func (routerExtractor) Extract(pkg *packages.Package) []Route { return routerRoutes(pkg) }

func routerRoutes(pkg *packages.Package) []Route {
	info := pkg.TypesInfo
	if info == nil {
		return nil
	}
	prefixes := groupPrefixes(pkg, anyReceiver)
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
			method, ok := httpVerbMethods[strings.ToUpper(sel.Sel.Name)]
			if !ok {
				return true
			}
			path := stringLit(call.Args[0])
			if path == "" {
				return true
			}
			handler := call.Args[len(call.Args)-1]
			if !isFuncTyped(info, handler) {
				return true // path + a function handler distinguishes a route from e.g. cache.Get(k)
			}
			out = append(out, Route{
				Method:  method,
				Path:    joinPath(groupPrefix(info, sel.X, prefixes), path),
				Handler: handlerFunc(info, handler),
			})
			return true
		})
	}
	return out
}

// anyReceiver accepts any expression as a router/group receiver, so Group
// prefixes are tracked regardless of framework type.
func anyReceiver(*types.Info, ast.Expr) bool { return true }

// isFuncTyped reports whether expr has a function type.
func isFuncTyped(info *types.Info, expr ast.Expr) bool {
	t := info.TypeOf(expr)
	if t == nil {
		return false
	}
	_, ok := t.Underlying().(*types.Signature)
	return ok
}
