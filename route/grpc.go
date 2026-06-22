package route

import (
	"go/ast"
	"go/types"
	"regexp"
	"strings"

	"golang.org/x/tools/go/packages"
)

// grpcRegister matches the generated registration function name, e.g.
// RegisterOrderServiceServer. This is the stable protoc-gen-go-grpc convention.
var grpcRegister = regexp.MustCompile(`^Register[A-Za-z0-9_]+Server$`)

// grpcExtractor detects gRPC services structurally: a call to a
// Register<Svc>Server(registrar, impl) function. The service interface is the
// function's second parameter; each of its RPC methods becomes an endpoint
// bound to the implementing method. Matching by shape (not by import) means it
// works with google.golang.org/grpc and any code generated in that style.
type grpcExtractor struct{}

func (grpcExtractor) Name() string { return "grpc" }

func (grpcExtractor) Match(pkg *packages.Package) bool {
	if pkg.TypesInfo == nil {
		return false
	}
	for _, file := range pkg.Syntax {
		found := false
		ast.Inspect(file, func(n ast.Node) bool {
			if found {
				return false
			}
			if call, ok := n.(*ast.CallExpr); ok {
				if fn := calleeFunc(pkg.TypesInfo, call.Fun); fn != nil && grpcRegister.MatchString(fn.Name()) {
					found = true
					return false
				}
			}
			return true
		})
		if found {
			return true
		}
	}
	return false
}

func (grpcExtractor) Extract(pkg *packages.Package) []Route {
	info := pkg.TypesInfo
	var out []Route
	for _, file := range pkg.Syntax {
		ast.Inspect(file, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok || len(call.Args) != 2 {
				return true
			}
			fn := calleeFunc(info, call.Fun)
			if fn == nil || !grpcRegister.MatchString(fn.Name()) {
				return true
			}
			sig, ok := fn.Type().(*types.Signature)
			if !ok || sig.Params().Len() != 2 {
				return true
			}
			ifaceType := sig.Params().At(1).Type()
			iface := underlyingIface(ifaceType)
			named := namedType(ifaceType)
			if iface == nil || named == nil {
				return true
			}
			service := strings.TrimSuffix(named.Obj().Name(), "Server")
			implType := info.TypeOf(call.Args[1])

			for i := 0; i < iface.NumMethods(); i++ {
				m := iface.Method(i)
				name := m.Name()
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

// calleeFunc resolves the function object a call expression invokes.
func calleeFunc(info *types.Info, fun ast.Expr) *types.Func {
	switch e := fun.(type) {
	case *ast.Ident:
		if fn, ok := info.ObjectOf(e).(*types.Func); ok {
			return fn
		}
	case *ast.SelectorExpr:
		if fn, ok := info.ObjectOf(e.Sel).(*types.Func); ok {
			return fn
		}
	}
	return nil
}

// namedType unwraps a pointer and returns the named type, or nil.
func namedType(t types.Type) *types.Named {
	if p, ok := t.(*types.Pointer); ok {
		t = p.Elem()
	}
	n, _ := t.(*types.Named)
	return n
}

// underlyingIface returns the interface a (named) type wraps, or nil.
func underlyingIface(t types.Type) *types.Interface {
	n := namedType(t)
	if n == nil {
		return nil
	}
	i, _ := n.Underlying().(*types.Interface)
	return i
}

// methodByName looks up an exported method on a type, or nil.
func methodByName(t types.Type, name string) *types.Func {
	if t == nil {
		return nil
	}
	var pkg *types.Package
	if n := namedType(t); n != nil && n.Obj() != nil {
		pkg = n.Obj().Pkg()
	}
	obj, _, _ := types.LookupFieldOrMethod(t, true, pkg, name)
	if fn, ok := obj.(*types.Func); ok {
		return fn
	}
	return nil
}
