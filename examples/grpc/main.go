// Command grpc is a gRPC backend in clean architecture: RPC methods on the
// order server delegate to a service and repository. archview detects the RPC
// methods as endpoints from the RegisterOrderServiceServer call. It mounts at
// /graph over HTTP (the gRPC server itself need not run for the demo).
package main

import (
	"log"
	"net/http"

	"github.com/eularixs/archview"

	"archview-example-grpc/api/orderpb"
	grpcadapter "archview-example-grpc/internal/order/adapter/grpc"
	"archview-example-grpc/internal/order/adapter/postgres"
	"archview-example-grpc/internal/order/service"
	"archview-example-grpc/internal/platform/grpcrt"
)

func main() {
	repo := postgres.New()
	svc := service.New(repo)

	// gRPC wiring: the RegisterOrderServiceServer call is archview's entry point.
	server := grpcrt.NewServer()
	orderpb.RegisterOrderServiceServer(server, grpcadapter.New(svc))

	av, err := archview.New(archview.Options{
		Root:      ".",
		BasePath:  "/graph",
		Editor:    "vscode",
		ShowPorts: true,
	})
	if err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()
	av.Mount(mux)

	log.Println("listening on :8096 — open http://localhost:8096/graph")
	if err := http.ListenAndServe(":8096", mux); err != nil {
		log.Fatal(err)
	}
}
