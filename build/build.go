// Package build assembles a graph.Graph from analyzer results, extracted routes
// and a layer classifier. It selects the nodes worth showing (endpoints plus
// functions in classified layers), then derives call edges between them —
// collapsing through unclassified helper functions so a controller still links
// to a service even when a thin wrapper sits in between.
//
// When ports are enabled, outbound interface ports (interfaces implemented by a
// repository-layer type) become first-class nodes: the direct service ->
// repository call is replaced by service -> port (uses) and repository -> port
// (implements), surfacing the hexagonal seam.
package build

import (
	"fmt"
	"sort"
	"strings"

	"github.com/eularixs/archview/analyzer"
	"github.com/eularixs/archview/classify"
	"github.com/eularixs/archview/graph"
	"github.com/eularixs/archview/route"
	"golang.org/x/tools/go/ssa"
)

// layeredLayers are the layers whose functions are included as nodes.
var layeredLayers = map[string]bool{
	graph.LayerController: true,
	graph.LayerService:    true,
	graph.LayerRepository: true,
}

type builder struct {
	res         *analyzer.Result
	cl          *classify.Classifier
	editor      string
	showPorts   bool
	detectBuses bool

	included map[*ssa.Function]bool
	layerOf  map[*ssa.Function]string
	moduleOf map[*ssa.Function]string
	idOf     map[*ssa.Function]string
	barrier  map[*ssa.Function]bool // funcs CHA must not collapse through (bus methods)
}

// outboundPort is a port retained for rendering plus its included endpoints.
type outboundPort struct {
	port    *analyzer.Port
	id      string
	module  string
	callers []*ssa.Function
	impls   []*ssa.Function
}

// Graph builds the architecture graph. When showPorts is true, outbound
// interface ports are surfaced as nodes. When detectBuses is true, mediator
// dispatch (command/query/event buses) is recovered as precise edges.
func Graph(res *analyzer.Result, routes []route.Route, cl *classify.Classifier, editor string, showPorts, detectBuses bool) graph.Graph {
	b := &builder{
		res:         res,
		cl:          cl,
		editor:      editor,
		showPorts:   showPorts,
		detectBuses: detectBuses,
		included:    map[*ssa.Function]bool{},
		layerOf:     map[*ssa.Function]string{},
		moduleOf:    map[*ssa.Function]string{},
		idOf:        map[*ssa.Function]string{},
		barrier:     map[*ssa.Function]bool{},
	}
	return b.run(routes)
}

func (b *builder) run(routes []route.Route) graph.Graph {
	// 1. Classify every project func; include those in known layers.
	for fn, f := range b.res.Funcs {
		layer, module := b.classify(f.Pkg)
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

	// 2b. Detect mediator buses. Bus methods become call-graph barriers so the
	//     over-approximated edges CHA draws through the bus are dropped; precise
	//     dispatch edges are added in step 8.
	var busInfo *analyzer.BusInfo
	if b.detectBuses {
		busInfo = b.res.Buses()
		b.barrier = busInfo.BusMethods
	}

	// 3. Resolve outbound ports and the direct call edges they replace.
	var ports []outboundPort
	suppress := map[string]bool{} // "callerID->implID" call edges mediated by a port
	if b.showPorts {
		for _, p := range b.res.Ports() {
			var impls []*ssa.Function
			outbound := false
			for fn := range p.ImplMethods {
				if !b.included[fn] {
					continue
				}
				impls = append(impls, fn)
				if b.layerOf[fn] == graph.LayerRepository {
					outbound = true
				}
			}
			if !outbound || len(impls) == 0 {
				continue
			}
			var callers []*ssa.Function
			for fn := range p.Callers {
				if b.included[fn] {
					callers = append(callers, fn)
				}
			}
			_, module := b.classify(p.Pkg)
			op := outboundPort{
				port:    p,
				id:      "port:" + p.Pkg + "." + p.Name,
				module:  module,
				callers: callers,
				impls:   impls,
			}
			ports = append(ports, op)
			for _, c := range callers {
				for _, m := range impls {
					suppress[b.id(c)+"->"+b.id(m)] = true
				}
			}
		}
	}

	g := graph.Graph{Module: b.res.Module}

	// 4. Function nodes.
	for fn := range b.included {
		f := b.res.Funcs[fn]
		g.Nodes = append(g.Nodes, graph.Node{
			ID:        b.id(fn),
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

	// 5. Call edges between included nodes (collapsing through helpers, skipping
	//    any edge a port now mediates).
	seen := map[string]bool{}
	for fn := range b.included {
		for callee := range b.reachableIncluded(fn) {
			from, to := b.id(fn), b.id(callee)
			key := from + "->" + to
			if from == to || seen[key] || suppress[key] {
				continue
			}
			seen[key] = true
			g.Edges = append(g.Edges, graph.Edge{From: from, To: to, Kind: graph.EdgeCall})
		}
	}

	// 6. Endpoint nodes + route edges to their handler.
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

	// 7. Port nodes + converging edges (caller uses port, adapter implements it).
	for _, op := range ports {
		g.Nodes = append(g.Nodes, graph.Node{
			ID:        op.id,
			Kind:      graph.KindPort,
			Label:     op.port.Name,
			Layer:     graph.LayerPort,
			Module:    op.module,
			Pkg:       op.port.Pkg,
			File:      op.port.File,
			Line:      op.port.Line,
			EditorURL: graph.EditorURL(b.editor, op.port.File, op.port.Line, op.port.Col),
		})
		for _, c := range op.callers {
			g.Edges = append(g.Edges, graph.Edge{From: b.id(c), To: op.id, Kind: graph.EdgeCall})
		}
		for _, m := range op.impls {
			g.Edges = append(g.Edges, graph.Edge{From: b.id(m), To: op.id, Kind: graph.EdgeImplements})
		}
	}

	// 8. Precise dispatch edges recovered from bus registrations (caller ->
	//    concrete handler), replacing the over-approximation CHA would draw.
	if busInfo != nil {
		seenD := map[string]bool{}
		for _, d := range busInfo.Dispatches {
			if !b.included[d.Caller] {
				continue
			}
			from := b.id(d.Caller)
			for _, h := range d.Handlers {
				if !b.included[h] {
					continue
				}
				to := b.id(h)
				key := from + "|" + to
				if from == to || seenD[key] {
					continue
				}
				seenD[key] = true
				g.Edges = append(g.Edges, graph.Edge{From: from, To: to, Kind: graph.EdgeDispatch})
			}
		}
	}

	pruneIsolatedFuncs(&g)
	sortGraph(&g)
	return g
}

// pruneIsolatedFuncs drops function nodes that have no incident edge (e.g.
// constructors wired only from main), keeping the graph focused on the flow.
// Endpoint and port nodes are always kept.
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
			case b.barrier[callee]:
				// bus method: stop so CHA's over-approx fan-out through the bus
				// is dropped (precise dispatch edges are added separately).
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

// classify strips the module path prefix before classifying, so a package is
// classified by its position within the module — never by the module name
// itself (which may, e.g., end in "grpc" and falsely match a layer keyword).
func (b *builder) classify(pkgPath string) (layer, module string) {
	rel := pkgPath
	if b.res.Module != "" {
		rel = strings.TrimPrefix(rel, b.res.Module)
		rel = strings.TrimPrefix(rel, "/")
	}
	return b.cl.Classify(rel)
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
