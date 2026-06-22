package route

import (
	"go/types"
	"strings"

	"golang.org/x/tools/go/packages"
)

// gqlResolvers are the gqlgen-generated root resolver interface names; each maps
// to a GraphQL operation kind.
var gqlResolvers = map[string]string{
	"QueryResolver":        "QUERY",
	"MutationResolver":     "MUTATION",
	"SubscriptionResolver": "SUBSCRIPTION",
}

// graphqlExtractor detects gqlgen-style GraphQL resolvers: an interface named
// Query/Mutation/SubscriptionResolver (in this package or an imported one)
// implemented by a project concrete type. Each interface method is a GraphQL
// field, surfaced as an endpoint bound to the implementing resolver method.
type graphqlExtractor struct{}

func (graphqlExtractor) Name() string { return "graphql" }

func (graphqlExtractor) Match(pkg *packages.Package) bool {
	return len(resolveGraphQL(pkg)) > 0
}

func (graphqlExtractor) Extract(pkg *packages.Package) []Route {
	return resolveGraphQL(pkg)
}

// resolveGraphQL finds resolver interfaces and their project implementers in a
// package, returning one route per field.
func resolveGraphQL(pkg *packages.Package) []Route {
	if pkg.Types == nil {
		return nil
	}

	// Candidate resolver interfaces: this package's scope plus imports'.
	type cand struct {
		named *types.Named
		iface *types.Interface
		op    string
	}
	var cands []cand
	scopes := []*types.Scope{pkg.Types.Scope()}
	for _, imp := range pkg.Types.Imports() {
		scopes = append(scopes, imp.Scope())
	}
	for _, sc := range scopes {
		for name, op := range gqlResolvers {
			tn, ok := sc.Lookup(name).(*types.TypeName)
			if !ok {
				continue
			}
			if named, ok := tn.Type().(*types.Named); ok {
				if iface, ok := named.Underlying().(*types.Interface); ok && iface.NumMethods() > 0 {
					cands = append(cands, cand{named, iface, op})
				}
			}
		}
	}
	if len(cands) == 0 {
		return nil
	}

	// Concrete struct types declared in this package.
	var concretes []*types.Named
	scope := pkg.Types.Scope()
	for _, name := range scope.Names() {
		if tn, ok := scope.Lookup(name).(*types.TypeName); ok {
			if named, ok := tn.Type().(*types.Named); ok {
				if _, ok := named.Underlying().(*types.Struct); ok {
					concretes = append(concretes, named)
				}
			}
		}
	}

	var out []Route
	seen := map[string]bool{}
	for _, c := range cands {
		for _, ct := range concretes {
			ptr := types.NewPointer(ct)
			if !types.Implements(ptr, c.iface) {
				continue
			}
			service := strings.TrimSuffix(c.named.Obj().Name(), "Resolver")
			for i := 0; i < c.iface.NumMethods(); i++ {
				m := c.iface.Method(i)
				field := m.Name()
				path := "/" + service + "/" + lcFirst(field)
				if seen[path] {
					continue
				}
				seen[path] = true
				out = append(out, Route{
					Method:  c.op,
					Path:    path,
					Handler: methodByName(ptr, field),
					Pos:     m.Pos(),
				})
			}
			break // one implementer per resolver interface is enough
		}
	}
	return out
}

// lcFirst lowercases the first rune (Go method Order -> GraphQL field order).
func lcFirst(s string) string {
	if s == "" {
		return s
	}
	return strings.ToLower(s[:1]) + s[1:]
}
