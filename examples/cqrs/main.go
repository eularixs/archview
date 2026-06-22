// Command cqrs is a CQRS + event-driven demo in clean architecture: an HTTP
// controller dispatches commands and queries through runtime buses (a mediator),
// and handlers publish domain events to subscribers. It mounts archview at
// /graph to show what static analysis can and cannot follow across the buses.
package main

import (
	"log"
	"net/http"

	"github.com/eularixs/archview"

	"archview-example-cqrs/internal/order/adapter/postgres"
	"archview-example-cqrs/internal/order/adapter/rest"
	"archview-example-cqrs/internal/order/application/command"
	"archview-example-cqrs/internal/order/application/event"
	"archview-example-cqrs/internal/order/application/query"
	"archview-example-cqrs/internal/order/domain"
	"archview-example-cqrs/internal/platform/cqrs"
)

func main() {
	repo := postgres.New()
	events := cqrs.NewEventBus()
	commands := cqrs.NewCommandBus()
	queries := cqrs.NewQueryBus()

	// Runtime wiring: command/query names -> handlers, event names -> subscribers.
	commands.Register(command.CreateOrder{}.CommandName(), command.NewCreateOrderHandler(repo, events))
	commands.Register(command.CancelOrder{}.CommandName(), command.NewCancelOrderHandler(repo, events))
	queries.Register(query.GetOrder{}.QueryName(), query.NewGetOrderHandler(repo))
	queries.Register(query.ListOrders{}.QueryName(), query.NewListOrdersHandler(repo))
	events.Subscribe(domain.OrderCreated{}.EventName(), event.EmailOnOrderCreated{})
	events.Subscribe(domain.OrderCreated{}.EventName(), event.AdjustInventory{})
	events.Subscribe(domain.OrderCancelled{}.EventName(), event.AdjustInventory{})

	ctl := rest.New(commands, queries)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /orders", ctl.Create)
	mux.HandleFunc("POST /orders/cancel", ctl.Cancel)
	mux.HandleFunc("GET /orders/{id}", ctl.Get)
	mux.HandleFunc("GET /orders", ctl.List)

	av, err := archview.New(archview.Options{
		Root:        ".",
		BasePath:    "/graph",
		Editor:      "vscode",
		ShowPorts:   true,
		DetectBuses: true, // recover command/query/event routing
	})
	if err != nil {
		log.Fatal(err)
	}
	av.Mount(mux)

	log.Println("listening on :8095 — open http://localhost:8095/graph")
	if err := http.ListenAndServe(":8095", mux); err != nil {
		log.Fatal(err)
	}
}
