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

func ginGroupRecv(info *types.Info, x ast.Expr) bool {
	return receiverIsType(info, x, ginPkg, "Engine", "RouterGroup", "IRoutes", "IRouter")
}

func (ginExtractor) Extract(pkg *packages.Package) []Route {
	info := pkg.TypesInfo
	prefixes := groupPrefixes(pkg, ginGroupRecv)
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
