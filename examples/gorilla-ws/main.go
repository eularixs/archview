// Command gorilla-ws is a demo backend over gorilla/mux with a WebSocket
// endpoint. archview reads the mux routes (incl. .Methods) and labels the
// upgrade handler "WS".
package main

import (
	"log"
	"net/http"

	"github.com/eularixs/archview"
	"github.com/gorilla/mux"

	"archview-example-gorilla-ws/internal/user/controller"
	"archview-example-gorilla-ws/internal/user/repository"
	"archview-example-gorilla-ws/internal/user/service"
)

func main() {
	ctl := controller.New(service.New(repository.New()))

	r := mux.NewRouter()
	r.HandleFunc("/users", ctl.List).Methods("GET")
	r.HandleFunc("/users", ctl.Create).Methods("POST")
	r.HandleFunc("/ws", ctl.Stream) // WebSocket upgrade

	av, err := archview.New(archview.Options{Root: ".", BasePath: "/graph", ShowPorts: true})
	if err != nil {
		log.Fatal(err)
	}
	r.Handle("/graph", av.Handler())
	r.Handle("/graph/data", av.Handler())

	log.Println("listening on :8100 — open http://localhost:8100/graph")
	log.Fatal(http.ListenAndServe(":8100", r))
}
