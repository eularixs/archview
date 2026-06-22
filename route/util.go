package route

import (
	"go/ast"
	"go/types"
	"strconv"
	"strings"

	"golang.org/x/tools/go/packages"
)

// imports reports whether pkg imports the given import path.
func imports(pkg *packages.Package, path string) bool {
	if pkg.Types == nil {
		return false
	}
	for _, imp := range pkg.Types.Imports() {
		if imp.Path() == path {
			return true
		}
	}
	return false
}

// deref unwraps a single pointer.
func deref(t types.Type) types.Type {
	if p, ok := t.(*types.Pointer); ok {
		return p.Elem()
	}
	return t
}

// receiverIsType reports whether expr's type is a (pointer to) named type from
// pkgPath whose name is in names.
func receiverIsType(info *types.Info, expr ast.Expr, pkgPath string, names ...string) bool {
	t := info.TypeOf(expr)
	if t == nil {
		return false
	}
	named, ok := deref(t).(*types.Named)
	if !ok || named.Obj() == nil || named.Obj().Pkg() == nil {
		return false
	}
	if named.Obj().Pkg().Path() != pkgPath {
		return false
	}
	for _, n := range names {
		if named.Obj().Name() == n {
			return true
		}
	}
	return false
}

// handlerFunc resolves an argument expression to the *types.Func it refers to,
// for handlers passed as a function or method value (e.g. ctrl.GetUser or
// GetUser). Inline closures and http.Handler values resolve to nil.
func handlerFunc(info *types.Info, expr ast.Expr) *types.Func {
	switch e := expr.(type) {
	case *ast.SelectorExpr:
		if fn, ok := info.ObjectOf(e.Sel).(*types.Func); ok {
			return fn
		}
	case *ast.Ident:
		if fn, ok := info.ObjectOf(e).(*types.Func); ok {
			return fn
		}
	case *ast.ParenExpr:
		return handlerFunc(info, e.X)
	}
	return nil
}

// stringLit returns the value of a string literal expression, or "" if expr is
// not a plain string literal.
func stringLit(expr ast.Expr) string {
	lit, ok := expr.(*ast.BasicLit)
	if !ok || lit.Kind.String() == "" {
		return ""
	}
	s, err := strconv.Unquote(lit.Value)
	if err != nil {
		return ""
	}
	return s
}

// groupPrefixes maps each router-group variable to its accumulated path prefix,
// so a route registered on a group reports the full path (e.g. r.Group("/api")
// then api.GET("/users") -> /api/users). recvMatch reports whether an
// expression is the framework's router/group receiver.
func groupPrefixes(pkg *packages.Package, recvMatch func(*types.Info, ast.Expr) bool) map[*types.Var]string {
	info := pkg.TypesInfo
	type raw struct {
		parent *types.Var
		prefix string
	}
	rawMap := map[*types.Var]raw{}

	for _, file := range pkg.Syntax {
		ast.Inspect(file, func(n ast.Node) bool {
			assign, ok := n.(*ast.AssignStmt)
			if !ok {
				return true
			}
			for i := 0; i < len(assign.Lhs) && i < len(assign.Rhs); i++ {
				call, ok := assign.Rhs[i].(*ast.CallExpr)
				if !ok || len(call.Args) < 1 {
					continue
				}
				sel, ok := call.Fun.(*ast.SelectorExpr)
				if !ok || sel.Sel.Name != "Group" || !recvMatch(info, sel.X) {
					continue
				}
				lhsID, ok := assign.Lhs[i].(*ast.Ident)
				if !ok {
					continue
				}
				lhsVar, ok := info.ObjectOf(lhsID).(*types.Var)
				if !ok {
					continue
				}
				var parent *types.Var
				if pid, ok := sel.X.(*ast.Ident); ok {
					parent, _ = info.ObjectOf(pid).(*types.Var)
				}
				rawMap[lhsVar] = raw{parent: parent, prefix: stringLit(call.Args[0])}
			}
			return true
		})
	}

	resolved := map[*types.Var]string{}
	var resolve func(v *types.Var, seen map[*types.Var]bool) string
	resolve = func(v *types.Var, seen map[*types.Var]bool) string {
		if p, ok := resolved[v]; ok {
			return p
		}
		r, ok := rawMap[v]
		if !ok || seen[v] {
			return ""
		}
		seen[v] = true
		full := joinPath(resolve(r.parent, seen), r.prefix)
		resolved[v] = full
		return full
	}
	for v := range rawMap {
		resolve(v, map[*types.Var]bool{})
	}
	return resolved
}

// groupPrefix returns the prefix for the receiver expression of a route call.
func groupPrefix(info *types.Info, recv ast.Expr, prefixes map[*types.Var]string) string {
	if id, ok := recv.(*ast.Ident); ok {
		if v, ok := info.ObjectOf(id).(*types.Var); ok {
			return prefixes[v]
		}
	}
	return ""
}

// joinPath joins a prefix and a path with exactly one separating slash.
func joinPath(prefix, path string) string {
	if prefix == "" {
		return path
	}
	prefix = strings.TrimRight(prefix, "/")
	if path == "" {
		return prefix
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return prefix + path
}

// httpMethods is the set of recognized HTTP method names.
var httpMethods = map[string]bool{
	"GET": true, "POST": true, "PUT": true, "DELETE": true,
	"PATCH": true, "HEAD": true, "OPTIONS": true, "CONNECT": true, "TRACE": true,
}

// splitMethodPath parses a net/http (Go 1.22+) pattern such as "GET /items/{id}"
// into method and path. Without a leading method it returns ("ANY", pattern).
func splitMethodPath(pattern string) (method, path string) {
	pattern = strings.TrimSpace(pattern)
	if i := strings.IndexByte(pattern, ' '); i > 0 {
		m := pattern[:i]
		if httpMethods[m] {
			return m, strings.TrimSpace(pattern[i+1:])
		}
	}
	return "ANY", pattern
}
