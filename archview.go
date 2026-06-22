// Package archview renders a live architecture flow graph of a Go backend.
//
// At startup it statically analyzes the module's source (dev-live mode): it
// builds a call graph, detects HTTP endpoints per framework (net/http, gin),
// classifies functions into layers (controller/service/repository), and serves
// an interactive graph at a mount path (default "/graph"). Clicking a node
// opens its source in the configured editor (vscode/cursor).
//
//	mux := http.NewServeMux()
//	av, err := archview.New(archview.Options{Root: "."})
//	if err != nil { log.Fatal(err) }
//	av.Mount(mux)
//
// dev-live requires the source tree and Go toolchain to be present at runtime.
package archview

import (
	"net/http"
	"strings"

	"github.com/eularixs/archview/analyzer"
	"github.com/eularixs/archview/build"
	"github.com/eularixs/archview/classify"
	"github.com/eularixs/archview/graph"
	"github.com/eularixs/archview/route"
	"github.com/eularixs/archview/web"
)

// Extractor is a per-framework route detector. Implement it to support a
// framework beyond the built-ins.
type Extractor = route.Extractor

// Options configures a Server.
type Options struct {
	// Root is the module directory to analyze. Defaults to ".".
	Root string
	// BasePath is the mount path for the UI. Defaults to "/graph".
	BasePath string
	// Editor is the deep-link scheme for click-to-source: "vscode" or "cursor".
	// Defaults to "vscode".
	Editor string
	// Extractors overrides the route extractors. Defaults to the built-ins
	// (gin + net/http).
	Extractors []Extractor
	// Classify optionally extends layer-classification keywords.
	Classify *classify.Config
	// ShowPorts surfaces outbound interface ports (a hexagonal seam) as nodes:
	// service -> port (uses) and repository -> port (implements). Off by default
	// so MVC graphs stay unchanged.
	ShowPorts bool
	// DetectBuses recovers command/query/event mediator routing: it reads the
	// bus registration sites and draws precise caller -> handler dispatch edges
	// instead of the over-approximation a static call graph produces. Off by
	// default.
	DetectBuses bool
	// ShowHelpers keeps trivial helper functions (unexported free functions in a
	// classified layer) as graph nodes. By default they are hidden and collapsed
	// through, so the flow stays connected without the clutter.
	ShowHelpers bool
	// DisableAutoLayer turns off chain-based layer inference. By default archview
	// infers layers for endpoint-reachable functions whose package name doesn't
	// match a layer keyword (entry=controller, calls-onward=service,
	// sink=repository), so it works on any layout — including microservices with
	// their own conventions — without keyword config. Keyword classification
	// still takes precedence where it applies. Set this to true for the curated
	// keyword-only view.
	DisableAutoLayer bool
}

// Server holds the analyzed graph and serves the UI.
type Server struct {
	handler *web.Handler
	graph   graph.Graph
}

// New analyzes the module at opts.Root and prepares the UI handler.
func New(opts Options) (*Server, error) {
	if opts.Root == "" {
		opts.Root = "."
	}
	if opts.BasePath == "" {
		opts.BasePath = "/graph"
	}
	if opts.Editor == "" {
		opts.Editor = "vscode"
	}
	extractors := opts.Extractors
	if extractors == nil {
		extractors = route.Default()
	}

	res, err := analyzer.Load(opts.Root)
	if err != nil {
		return nil, err
	}
	routes := route.Extract(res.Pkgs, extractors)
	routes = dropBasePathRoutes(routes, opts.BasePath)
	cl := classify.New(opts.Classify)
	g := build.Graph(res, routes, cl, build.Options{
		Editor:      opts.Editor,
		ShowPorts:   opts.ShowPorts,
		DetectBuses: opts.DetectBuses,
		ShowHelpers: opts.ShowHelpers,
		AutoLayer:   !opts.DisableAutoLayer,
	})

	h, err := web.New(opts.BasePath, g)
	if err != nil {
		return nil, err
	}
	return &Server{handler: h, graph: g}, nil
}

// Handler returns the http.Handler serving the UI (and {base}/data JSON).
func (s *Server) Handler() http.Handler { return s.handler }

// Graph returns the analyzed graph (useful for tests or custom rendering).
func (s *Server) Graph() graph.Graph { return s.graph }

// Base returns the normalized mount path (e.g. "/graph").
func (s *Server) Base() string { return s.handler.Base() }

// dropBasePathRoutes removes archview's own UI routes (the mount path and
// anything under it) so the graph doesn't show itself as an endpoint.
func dropBasePathRoutes(routes []route.Route, basePath string) []route.Route {
	base := "/" + strings.Trim(basePath, "/")
	if base == "/" {
		return routes
	}
	out := routes[:0]
	for _, r := range routes {
		if r.Path == base || strings.HasPrefix(r.Path, base+"/") {
			continue
		}
		out = append(out, r)
	}
	return out
}

// Mount registers the UI on a *http.ServeMux at the configured base path.
func (s *Server) Mount(mux *http.ServeMux) {
	base := s.handler.Base()
	mux.Handle(base, s.handler)
	mux.Handle(base+"/", s.handler)
}
