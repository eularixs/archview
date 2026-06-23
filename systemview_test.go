package archview_test

import (
	"testing"

	archview "github.com/eularixs/archview"
)

// SystemView stitches a gRPC client call in one service to the server handler in
// another, across the wire, and lanes nodes by service.
func TestSystemView_GRPCStitch(t *testing.T) {
	srv, err := archview.New(archview.Options{Root: "examples/microservices", SystemView: true})
	if err != nil {
		t.Fatal(err)
	}
	g := srv.Graph()

	services := map[string]bool{}
	id2 := map[string]string{}
	for _, n := range g.Nodes {
		id2[n.ID] = n.Label
		if n.Service != "" {
			services[n.Service] = true
		}
	}
	if !services["api"] || !services["worker"] {
		t.Fatalf("expected api + worker services, got %v", services)
	}

	var rpc int
	for _, e := range g.Edges {
		if e.Kind == "rpc" {
			rpc++
			if id2[e.To] != "(UserServer).GetUser" {
				t.Fatalf("rpc edge should target the server handler, got %s", id2[e.To])
			}
		}
	}
	if rpc == 0 {
		t.Fatal("expected a cross-service rpc edge (worker -> api)")
	}
}
