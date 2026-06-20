// Package build assembles a graph.Graph from analyzer results, extracted routes
// and a layer classifier. It selects the nodes worth showing (endpoints plus
// functions in classified layers), then derives call edges between them —
// collapsing through unclassified helper functions so a controller still links
// to a service even when a thin wrapper sits in between.
package build

import (
	"fmt"
	"sort"

	"github.com/eularix/archview/analyzer"
	"github.com/eularix/archview/classify"
	"github.com/eularix/archview/graph"
	"github.com/eularix/archview/route"
	"golang.org/x/tools/go/ssa"
)

// layeredLayers are the layers whose functions are included as nodes.
var layeredLayers = map[string]bool{
	graph.LayerController: true,
	graph.LayerService:    true,
	graph.LayerRepository: true,
}

type builder struct {
	res    *analyzer.Result
	cl     *classify.Classifier
	editor string

	included map[*ssa.Function]bool
	layerOf  map[*ssa.Function]string
	moduleOf map[*ssa.Function]string
	idOf     map[*ssa.Function]string
}

// Graph builds the architecture graph.
func Graph(res *analyzer.Result, routes []route.Route, cl *classify.Classifier, editor string) graph.Graph {
	b := &builder{
		res:      res,
		cl:       cl,
		editor:   editor,
		included: map[*ssa.Function]bool{},
		layerOf:  map[*ssa.Function]string{},
		moduleOf: map[*ssa.Function]string{},
		idOf:     map[*ssa.Function]string{},
	}
	return b.run(routes)
}

func (b *builder) run(routes []route.Route) graph.Graph {
	// 1. Classify every project func; include those in known layers.
	for fn, f := range b.res.Funcs {
		layer, module := b.cl.Classify(f.Pkg)
		b.layerOf[fn] = layer
		b.moduleOf[fn] = module
		if layeredLayers[layer] {
			b.included[fn] = true
		}
	}
	// 2. Force-include resolved route handlers (the entry into the call chain).
	for _, r := range routes {
		if f := b.res.FuncFor(r.Handler); f != nil {
			b.included[f.SSA] = true
		}
	}

	g := graph.Graph{Module: b.res.Module}

	// 3. Function nodes.
	for fn := range b.included {
		f := b.res.Funcs[fn]
		id := b.id(fn)
		g.Nodes = append(g.Nodes, graph.Node{
			ID:        id,
			Kind:      graph.KindFunc,
			Label:     f.Display(),
			Layer:     b.layerOf[fn],
			Module:    b.moduleOf[fn],
			Pkg:       f.Pkg,
			Func:      f.Name,
			File:      f.File,
			Line:      f.Line,
			EditorURL: graph.EditorURL(b.editor, f.File, f.Line, f.Col),
		})
	}

	// 4. Call edges between included nodes (collapsing through helpers).
	seen := map[string]bool{}
	for fn := range b.included {
		for callee := range b.reachableIncluded(fn) {
			from, to := b.id(fn), b.id(callee)
			key := from + "->" + to
			if from == to || seen[key] {
				continue
			}
			seen[key] = true
			g.Edges = append(g.Edges, graph.Edge{From: from, To: to, Kind: graph.EdgeCall})
		}
	}

	// 5. Endpoint nodes + route edges to their handler.
	for i, r := range routes {
		epID := fmt.Sprintf("ep:%d:%s:%s", i, r.Method, r.Path)
		module := ""
		if f := b.res.FuncFor(r.Handler); f != nil {
			module = b.moduleOf[f.SSA]
		}
		label := r.Path
		if label == "" {
			label = "(dynamic)"
		}
		g.Nodes = append(g.Nodes, graph.Node{
			ID:     epID,
			Kind:   graph.KindEndpoint,
			Label:  label,
			Layer:  graph.LayerEndpoint,
			Module: module,
			Method: r.Method,
			Path:   r.Path,
		})
		if f := b.res.FuncFor(r.Handler); f != nil && b.included[f.SSA] {
			g.Edges = append(g.Edges, graph.Edge{From: epID, To: b.id(f.SSA), Kind: graph.EdgeRoute})
		}
	}

	pruneIsolatedFuncs(&g)
	sortGraph(&g)
	return g
}

// pruneIsolatedFuncs drops function nodes that have no incident edge (e.g.
// constructors wired only from main), keeping the graph focused on the flow.
// Endpoint nodes are always kept.
func pruneIsolatedFuncs(g *graph.Graph) {
	deg := map[string]bool{}
	for _, e := range g.Edges {
		deg[e.From] = true
		deg[e.To] = true
	}
	kept := g.Nodes[:0]
	for _, n := range g.Nodes {
		if n.Kind == graph.KindFunc && !deg[n.ID] {
			continue
		}
		kept = append(kept, n)
	}
	g.Nodes = kept
}

// reachableIncluded returns the set of included functions reachable from fn,
// recursing only through *unincluded project* functions (helpers) and stopping
// at included functions and non-project functions (stdlib/deps).
func (b *builder) reachableIncluded(fn *ssa.Function) map[*ssa.Function]bool {
	out := map[*ssa.Function]bool{}
	node := b.res.CallGraph.Nodes[fn]
	if node == nil {
		return out
	}
	visited := map[*ssa.Function]bool{fn: true}
	var walk func(n *ssa.Function)
	walk = func(n *ssa.Function) {
		cgn := b.res.CallGraph.Nodes[n]
		if cgn == nil {
			return
		}
		for _, e := range cgn.Out {
			callee := e.Callee.Func
			if callee == nil || visited[callee] {
				continue
			}
			visited[callee] = true
			switch {
			case b.included[callee]:
				out[callee] = true // edge target; don't recurse past it
			case b.res.Funcs[callee] != nil:
				walk(callee) // unincluded project helper: collapse through
			default:
				// non-project (stdlib/dep): stop
			}
		}
	}
	walk(fn)
	return out
}

func (b *builder) id(fn *ssa.Function) string {
	if id, ok := b.idOf[fn]; ok {
		return id
	}
	f := b.res.Funcs[fn]
	id := f.Pkg + "." + f.Display()
	b.idOf[fn] = id
	return id
}

func sortGraph(g *graph.Graph) {
	order := map[string]int{}
	for i, l := range graph.LayerOrder {
		order[l] = i
	}
	sort.SliceStable(g.Nodes, func(i, j int) bool {
		a, b := g.Nodes[i], g.Nodes[j]
		if order[a.Layer] != order[b.Layer] {
			return order[a.Layer] < order[b.Layer]
		}
		if a.Module != b.Module {
			return a.Module < b.Module
		}
		return a.Label < b.Label
	})
	sort.SliceStable(g.Edges, func(i, j int) bool {
		a, b := g.Edges[i], g.Edges[j]
		if a.From != b.From {
			return a.From < b.From
		}
		if a.To != b.To {
			return a.To < b.To
		}
		return a.Kind < b.Kind
	})
}
