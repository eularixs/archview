// Package graph defines the data model for an archview architecture graph:
// nodes (endpoints and functions, tagged with their layer/module) and edges
// (route bindings and call relationships) plus JSON (de)serialization.
package graph

import (
	"fmt"
	"net/url"
)

// Node kinds.
const (
	KindEndpoint = "endpoint" // an HTTP route entry point
	KindFunc     = "func"     // a Go function/method in a classified layer
	KindPort     = "port"     // an interface that sits as a seam between layers
)

// Edge kinds.
const (
	EdgeRoute      = "route"      // endpoint -> handler function
	EdgeCall       = "call"       // function -> function (caller -> callee)
	EdgeImplements = "implements" // concrete method -> port interface (adapter implements port)
	EdgeDispatch   = "dispatch"   // caller -> handler routed through a command/event bus
)

// Layers. "other" is the fallback for funcs that don't match a known layer.
const (
	LayerEndpoint   = "endpoint"
	LayerController = "controller"
	LayerService    = "service"
	LayerPort       = "port"
	LayerRepository = "repository"
	LayerOther      = "other"
)

// LayerOrder is the left-to-right pipeline order used by the renderer. Ports
// sit between service and repository: a service uses a port, an adapter
// (repository) implements it — both arrows converge on the port.
var LayerOrder = []string{LayerEndpoint, LayerController, LayerService, LayerPort, LayerRepository, LayerOther}

// Node is a single box in the graph.
type Node struct {
	ID        string `json:"id"`
	Kind      string `json:"kind"`
	Label     string `json:"label"`
	Layer     string `json:"layer"`
	Module    string `json:"module"`
	Pkg       string `json:"pkg,omitempty"`
	Func      string `json:"func,omitempty"`
	File      string `json:"file,omitempty"`
	Line      int    `json:"line,omitempty"`
	EditorURL string `json:"editorURL,omitempty"`

	// Hash is a location-independent digest of the function's normalized body,
	// populated only in Raw mode. It lets external consumers (arch-diff) detect
	// a changed body without false positives from moved or reformatted code.
	Hash string `json:"hash,omitempty"`

	// Endpoint-only fields.
	Method string `json:"method,omitempty"`
	Path   string `json:"path,omitempty"`
}

// Edge violation kinds (set by the layer linter).
const (
	ViolationReverse = "reverse"      // calls backward toward the entry (e.g. repository -> service)
	ViolationSkip    = "skip"         // controller -> repository, bypassing the service layer
	ViolationCross   = "cross-module" // calls another module's internals
)

// Edge is a directed arrow between two nodes.
type Edge struct {
	From string `json:"from"`
	To   string `json:"to"`
	Kind string `json:"kind"`
	// Violation, when non-empty, marks an architecture smell on a call edge.
	Violation string `json:"violation,omitempty"`
}

// Graph is the whole picture.
type Graph struct {
	Module string `json:"module"`
	Nodes  []Node `json:"nodes"`
	Edges  []Edge `json:"edges"`
}

// EditorURL builds a deep link that opens file:line:col in the given editor.
// Supported schemes: "vscode", "cursor". Unknown scheme returns "".
func EditorURL(scheme, absFile string, line, col int) string {
	switch scheme {
	case "vscode", "cursor":
		// vscode://file/<abs-path>:<line>:<col> — path must be absolute.
		return fmt.Sprintf("%s://file/%s:%d:%d", scheme, pathEscape(absFile), line, col)
	default:
		return ""
	}
}

// pathEscape escapes a filesystem path for use in a URL while keeping the
// slashes that separate path segments intact.
func pathEscape(p string) string {
	// Escape each segment but preserve "/" separators.
	out := make([]byte, 0, len(p))
	seg := make([]byte, 0, 32)
	flush := func() {
		out = append(out, []byte(url.PathEscape(string(seg)))...)
		seg = seg[:0]
	}
	for i := 0; i < len(p); i++ {
		if p[i] == '/' {
			flush()
			out = append(out, '/')
			continue
		}
		seg = append(seg, p[i])
	}
	flush()
	return string(out)
}
