// Package web serves the archview UI: an HTML shell (with an embedded,
// dependency-free SVG renderer) at the base path, and the graph JSON at
// {base}/data.
package web

import (
	_ "embed"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/eularix/archview/graph"
)

//go:embed index.html
var indexHTML string

// Handler serves the archview UI for a pre-built graph.
type Handler struct {
	base string
	page []byte
	data []byte
}

// New builds a Handler for graph g mounted at base (e.g. "/graph").
func New(base string, g graph.Graph) (*Handler, error) {
	base = "/" + strings.Trim(base, "/")
	data, err := json.Marshal(g)
	if err != nil {
		return nil, err
	}
	page := strings.ReplaceAll(indexHTML, "__BASE__", base)
	return &Handler{base: base, page: []byte(page), data: data}, nil
}

// Base returns the normalized mount path.
func (h *Handler) Base() string { return h.base }

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch strings.TrimSuffix(r.URL.Path, "/") {
	case h.base, "":
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(h.page)
	case h.base + "/data":
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Write(h.data)
	default:
		http.NotFound(w, r)
	}
}
