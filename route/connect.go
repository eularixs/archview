package route

import (
	"go/ast"
	"go/types"
	"regexp"
	"strings"

	"golang.org/x/tools/go/packages"
)

// connectHandler matches the ConnectRPC generated constructor, e.g.
// NewUserServiceHandler(svc) (string, http.Handler).
var connectHandler = regexp.MustCompile(`^New[A-Za-z0-9_]+Handler$`)

// connectExtractor detects ConnectRPC services. The generated
// New<Svc>Handler(impl) returns (string, http.Handler); its first parameter is
// the service handler interface, whose methods are the RPCs. Matching by the
// (string, http.Handler) return shape distinguishes it from ordinary
// NewHandler constructors.
type connectExtractor struct{}

func (connectExtractor) Name() string { return "connect" }

func (connectExtractor) Match(pkg *packages.Package) bool {
	return len(resolveConnect(pkg)) > 0
}

func (connectExtractor) Extract(pkg *packages.Package) []Route {
	return resolveConnect(pkg)
}

func resolveConnect(pkg *packages.Package) []Route {
	info := pkg.TypesInfo
	if info == nil {
		return nil
	}
	var out []Route
	for _, file := range pkg.Syntax {
		ast.Inspect(file, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok || len(call.Args) < 1 {
				return true
			}
			fn := calleeFunc(info, call.Fun)
			if fn == nil || !connectHandler.MatchString(fn.Name()) {
				return true
			}
			sig, ok := fn.Type().(*types.Signature)
			if !ok || sig.Params().Len() < 1 || sig.Results().Len() != 2 {
				return true
			}
			if sig.Results().At(0).Type().String() != "string" || !isHTTPHandler(sig.Results().At(1).Type()) {
				return true
			}
			iface := underlyingIface(sig.Params().At(0).Type())
			named := namedType(sig.Params().At(0).Type())
			if iface == nil || named == nil {
				return true
			}
			service := strings.TrimSuffix(named.Obj().Name(), "Handler")
			implType := info.TypeOf(call.Args[0])
			for i := 0; i < iface.NumMethods(); i++ {
				name := iface.Method(i).Name()
				if strings.HasPrefix(name, "mustEmbed") || strings.HasPrefix(name, "Unimplemented") {
					continue
				}
				out = append(out, Route{
					Method:  "RPC",
					Path:    "/" + service + "/" + name,
					Handler: methodByName(implType, name),
				})
			}
			return true
		})
	}
	return out
}

// isHTTPHandler reports whether t is net/http.Handler.
func isHTTPHandler(t types.Type) bool {
	n := namedType(t)
	if n == nil || n.Obj() == nil || n.Obj().Pkg() == nil {
		return false
	}
	return n.Obj().Pkg().Path() == "net/http" && n.Obj().Name() == "Handler"
}
