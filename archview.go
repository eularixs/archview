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

	"github.com/eularix/archview/analyzer"
	"github.com/eularix/archview/build"
	"github.com/eularix/archview/classify"
	"github.com/eularix/archview/graph"
	"github.com/eularix/archview/route"
	"github.com/eularix/archview/web"
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
	cl := classify.New(opts.Classify)
	g := build.Graph(res, routes, cl, opts.Editor)

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

// Mount registers the UI on a *http.ServeMux at the configured base path.
func (s *Server) Mount(mux *http.ServeMux) {
	base := s.handler.Base()
	mux.Handle(base, s.handler)
	mux.Handle(base+"/", s.handler)
}
