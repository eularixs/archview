// Command graphql is a GraphQL backend (gqlgen-style resolvers) in clean
// architecture: resolver fields delegate to a service and repository. archview
// detects the resolver fields as endpoints. It mounts at /graph over HTTP.
package main

import (
	"log"
	"net/http"

	"github.com/eularixs/archview"

	graphqladapter "archview-example-graphql/internal/order/adapter/graphql"
	"archview-example-graphql/internal/order/adapter/postgres"
	"archview-example-graphql/internal/order/service"
)

func main() {
	repo := postgres.New()
	svc := service.New(repo)

	// In a real app this resolver is wired into a gqlgen handler; archview reads
	// the resolver interfaces statically.
	resolver := graphqladapter.NewResolver(svc)
	var _ graphqladapter.ResolverRoot = resolver

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

	log.Println("listening on :8097 — open http://localhost:8097/graph")
	if err := http.ListenAndServe(":8097", mux); err != nil {
		log.Fatal(err)
	}
}
