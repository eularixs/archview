package route

import (
	"go/ast"
	"go/types"
	"strings"

	"golang.org/x/tools/go/packages"
)

const ginPkg = "github.com/gin-gonic/gin"

// ginExtractor detects routes registered on *gin.Engine / *gin.RouterGroup.
type ginExtractor struct{}

func (ginExtractor) Name() string { return "gin" }

func (ginExtractor) Match(pkg *packages.Package) bool { return imports(pkg, ginPkg) }

// ginVerbs are the method-named registration calls.
var ginVerbs = map[string]bool{
	"GET": true, "POST": true, "PUT": true, "DELETE": true,
	"PATCH": true, "HEAD": true, "OPTIONS": true,
}

func (ginExtractor) Extract(pkg *packages.Package) []Route {
	info := pkg.TypesInfo
	prefixes := ginGroupPrefixes(pkg)
	var out []Route
	for _, file := range pkg.Syntax {
		ast.Inspect(file, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			sel, ok := call.Fun.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			name := sel.Sel.Name
			if !ginVerbs[name] && name != "Handle" && name != "Any" {
				return true
			}
			if !receiverIsType(info, sel.X, ginPkg, "Engine", "RouterGroup", "IRoutes", "IRouter") {
				return true
			}

			var method, path string
			var handlerExpr ast.Expr
			switch {
			case ginVerbs[name]:
				if len(call.Args) < 2 {
					return true
				}
				method = name
				path = stringLit(call.Args[0])
				handlerExpr = call.Args[len(call.Args)-1]
			case name == "Any":
				if len(call.Args) < 2 {
					return true
				}
				method = "ANY"
				path = stringLit(call.Args[0])
				handlerExpr = call.Args[len(call.Args)-1]
			case name == "Handle":
				if len(call.Args) < 3 {
					return true
				}
				method = strings.ToUpper(stringLit(call.Args[0]))
				path = stringLit(call.Args[1])
				handlerExpr = call.Args[len(call.Args)-1]
			}

			out = append(out, Route{
				Method:  method,
				Path:    joinPath(groupPrefix(info, sel.X, prefixes), path),
				Handler: handlerFunc(info, handlerExpr),
			})
			return true
		})
	}
	return out
}

// ginGroupPrefixes maps each router-group variable to its accumulated path
// prefix, so a route registered on a group reports the full path
// (e.g. r.Group("/api") then api.GET("/users") -> /api/users).
func ginGroupPrefixes(pkg *packages.Package) map[*types.Var]string {
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
				if !ok || sel.Sel.Name != "Group" {
					continue
				}
				if !receiverIsType(info, sel.X, ginPkg, "Engine", "RouterGroup", "IRoutes", "IRouter") {
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
