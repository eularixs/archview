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
	"path/filepath"
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

// Options controls how the graph is assembled.
type Options struct {
	Editor      string
	ShowPorts   bool
	DetectBuses bool
	// ShowHelpers keeps trivial helper functions (unexported free functions in a
	// classified layer) as nodes. When false (default) they are collapsed
	// through, so a caller still links to what the helper reaches.
	ShowHelpers bool
	// AutoLayer infers a layer for functions reached from an endpoint whose
	// package name doesn't match a layer keyword: the entry is a controller, a
	// function that calls further into the app is a service, and a sink (one
	// that only reaches external/leaf code) is a repository. This makes archview
	// work on any layout without per-project keyword config. Keyword
	// classification still wins where it applies.
	AutoLayer bool
	// LintLayers flags architecture smells on call edges (reverse dependencies,
	// controller bypassing the service layer, cross-module calls).
	LintLayers bool
}

type builder struct {
	res         *analyzer.Result
	cl          *classify.Classifier
	editor      string
	showPorts   bool
	detectBuses bool
	showHelpers bool
	autoLayer   bool
	lintLayers  bool

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

// Graph builds the architecture graph per opts.
func Graph(res *analyzer.Result, routes []route.Route, cl *classify.Classifier, opts Options) graph.Graph {
	b := &builder{
		res:         res,
		cl:          cl,
		editor:      opts.Editor,
		showPorts:   opts.ShowPorts,
		detectBuses: opts.DetectBuses,
		showHelpers: opts.ShowHelpers,
		autoLayer:   opts.AutoLayer,
		lintLayers:  opts.LintLayers,
		included:    map[*ssa.Function]bool{},
		layerOf:     map[*ssa.Function]string{},
		moduleOf:    map[*ssa.Function]string{},
		idOf:        map[*ssa.Function]string{},
		barrier:     map[*ssa.Function]bool{},
	}
	return b.run(routes)
}

func (b *builder) run(routes []route.Route) graph.Graph {
	// 1. Classify every project func; include those in known layers, skipping
	//    trivial helpers (collapsed through later) unless ShowHelpers is set.
	for fn, f := range b.res.Funcs {
		layer, module := b.classify(f.Pkg)
		b.layerOf[fn] = layer
		b.moduleOf[fn] = module
		if layeredLayers[layer] && (b.showHelpers || !trivialHelper(f)) {
			b.included[fn] = true
		}
	}
	// 2. Force-include resolved route handlers (the entry into the call chain).
	for _, r := range routes {
		if f := b.res.FuncFor(r.Handler); f != nil {
			b.included[f.SSA] = true
		}
	}

	// 2a. Detect mediator buses. Bus methods become call-graph barriers so the
	//     over-approximated edges CHA draws through the bus are dropped; precise
	//     dispatch edges are added in step 8.
	var busInfo *analyzer.BusInfo
	if b.detectBuses {
		busInfo = b.res.Buses()
		b.barrier = busInfo.BusMethods
	}

	// 2b. Auto-layer: include functions reachable from endpoints whose package
	//     isn't keyword-classified, inferring their layer from call-chain role
	//     (entry=controller, calls-onward=service, sink=repository). Runs after
	//     bus detection so it respects the bus barrier.
	if b.autoLayer {
		b.inferLayers(routes)
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

	// 6. Endpoint nodes + route edges to their handler. The same method+path
	//    registered by more than one transport (e.g. grpc + ConnectRPC for one
	//    RPC) collapses to a single endpoint with an edge to each handler. A
	//    handler that performs a WebSocket upgrade is labeled "WS".
	wsUpgraders := b.res.WSUpgraders()
	epExists := map[string]bool{}
	for _, r := range routes {
		method := r.Method
		var handler *analyzer.Func
		if f := b.res.FuncFor(r.Handler); f != nil {
			handler = f
			if wsUpgraders[f.SSA] {
				method = "WS"
			}
		}
		epID := "ep:" + method + ":" + r.Path
		if !epExists[epID] {
			epExists[epID] = true
			label := r.Path
			if label == "" {
				label = "(dynamic)"
			}
			module := ""
			if handler != nil {
				module = b.moduleOf[handler.SSA]
			}
			// Click-to-source for the endpoint: the route registration site.
			var file, editorURL string
			var line int
			if r.Pos.IsValid() {
				pos := b.res.Fset.Position(r.Pos)
				file = pos.Filename
				if a, err := filepath.Abs(file); err == nil {
					file = a
				}
				line = pos.Line
				editorURL = graph.EditorURL(b.editor, file, pos.Line, pos.Column)
			}
			g.Nodes = append(g.Nodes, graph.Node{
				ID:        epID,
				Kind:      graph.KindEndpoint,
				Label:     label,
				Layer:     graph.LayerEndpoint,
				Module:    module,
				Method:    method,
				Path:      r.Path,
				File:      file,
				Line:      line,
				EditorURL: editorURL,
			})
		}
		if handler != nil && b.included[handler.SSA] {
			g.Edges = append(g.Edges, graph.Edge{From: epID, To: b.id(handler.SSA), Kind: graph.EdgeRoute})
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
	pruneDisconnected(&g)
	sortGraph(&g)
	if b.lintLayers {
		lintLayers(&g)
	}
	return g
}

// pruneDisconnected drops nodes not connected (in any edge direction) to an
// endpoint, so wiring/setup clusters that aren't part of any request flow —
// e.g. a grpc RegisterRoutes calling a NewServer constructor — fall away. It is
// a no-op when no endpoints were detected, to avoid emptying the graph.
func pruneDisconnected(g *graph.Graph) {
	hasEndpoint := false
	for _, n := range g.Nodes {
		if n.Kind == graph.KindEndpoint {
			hasEndpoint = true
			break
		}
	}
	if !hasEndpoint {
		return
	}
	adj := map[string][]string{}
	for _, e := range g.Edges {
		adj[e.From] = append(adj[e.From], e.To)
		adj[e.To] = append(adj[e.To], e.From)
	}
	keep := map[string]bool{}
	var queue []string
	for _, n := range g.Nodes {
		if n.Kind == graph.KindEndpoint {
			keep[n.ID] = true
			queue = append(queue, n.ID)
		}
	}
	for len(queue) > 0 {
		n := queue[0]
		queue = queue[1:]
		for _, m := range adj[n] {
			if !keep[m] {
				keep[m] = true
				queue = append(queue, m)
			}
		}
	}
	nodes := g.Nodes[:0]
	for _, n := range g.Nodes {
		if keep[n.ID] {
			nodes = append(nodes, n)
		}
	}
	g.Nodes = nodes
	edges := g.Edges[:0]
	for _, e := range g.Edges {
		if keep[e.From] && keep[e.To] {
			edges = append(edges, e)
		}
	}
	g.Edges = edges
}

// layerRank orders the layers from entry to data; a call to a lower rank is a
// backward (reverse) dependency.
var layerRank = map[string]int{
	graph.LayerEndpoint:   0,
	graph.LayerController: 1,
	graph.LayerService:    2,
	graph.LayerPort:       3,
	graph.LayerRepository: 4,
	graph.LayerOther:      5,
}

// lintLayers marks architecture smells on call edges.
func lintLayers(g *graph.Graph) {
	node := map[string]graph.Node{}
	for _, n := range g.Nodes {
		node[n.ID] = n
	}
	classified := func(l string) bool {
		return l == graph.LayerController || l == graph.LayerService || l == graph.LayerRepository
	}
	for i := range g.Edges {
		e := &g.Edges[i]
		if e.Kind != graph.EdgeCall {
			continue
		}
		a, ok1 := node[e.From]
		b, ok2 := node[e.To]
		if !ok1 || !ok2 || !classified(a.Layer) || !classified(b.Layer) {
			continue
		}
		switch {
		case layerRank[b.Layer] < layerRank[a.Layer]:
			e.Violation = graph.ViolationReverse
		case a.Layer == graph.LayerController && b.Layer == graph.LayerRepository:
			e.Violation = graph.ViolationSkip
		case a.Module != "" && b.Module != "" && a.Module != b.Module:
			e.Violation = graph.ViolationCross
		}
	}
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

// inferLayers walks the call graph from endpoint handlers (and keyword
// controllers), including every project function reached and assigning a layer
// to those the keyword classifier left as "other": the entry is a controller, a
// function that calls onward into the app is a service, and a sink is a
// repository. Keyword-classified layers are left untouched.
func (b *builder) inferLayers(routes []route.Route) {
	entries := map[*ssa.Function]bool{}
	for fn := range b.included {
		if b.layerOf[fn] == graph.LayerController {
			entries[fn] = true
		}
	}
	for _, r := range routes {
		if f := b.res.FuncFor(r.Handler); f != nil {
			entries[f.SSA] = true
		}
	}

	// BFS over project functions reachable from the entries.
	reachable := map[*ssa.Function]bool{}
	var queue []*ssa.Function
	for fn := range entries {
		reachable[fn] = true
		queue = append(queue, fn)
	}
	for len(queue) > 0 {
		n := queue[0]
		queue = queue[1:]
		cgn := b.res.CallGraph.Nodes[n]
		if cgn == nil {
			continue
		}
		for _, e := range cgn.Out {
			c := e.Callee.Func
			if c == nil || reachable[c] || b.res.Funcs[c] == nil || b.barrier[c] {
				continue // skip visited, non-project, and bus-barrier funcs
			}
			reachable[c] = true
			queue = append(queue, c)
		}
	}

	cand := func(fn *ssa.Function) bool {
		return reachable[fn] && (b.showHelpers || !trivialHelper(b.res.Funcs[fn]))
	}

	for fn := range reachable {
		if !cand(fn) {
			continue
		}
		b.included[fn] = true
		if layeredLayers[b.layerOf[fn]] {
			continue // keyword classification wins
		}
		if entries[fn] {
			b.layerOf[fn] = graph.LayerController
			continue
		}
		// A function that calls onward to another included candidate is an
		// intermediate (service); a sink is a repository.
		intermediate := false
		if cgn := b.res.CallGraph.Nodes[fn]; cgn != nil {
			for _, e := range cgn.Out {
				if c := e.Callee.Func; c != nil && c != fn && cand(c) {
					intermediate = true
					break
				}
			}
		}
		if intermediate {
			b.layerOf[fn] = graph.LayerService
		} else {
			b.layerOf[fn] = graph.LayerRepository
		}
	}
}

// trivialHelper reports whether f is an unexported free function — a package
// helper rather than a layer participant (controllers/services/repositories are
// methods). Route/RPC/resolver handlers are methods or force-included, so this
// never hides a real entry point. Such helpers are collapsed through, keeping
// the caller connected to whatever the helper reaches.
func trivialHelper(f *analyzer.Func) bool {
	return f.Recv == "" && f.Name != "" && f.Name[0] >= 'a' && f.Name[0] <= 'z'
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
