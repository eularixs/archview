// Package rest is the order inbound adapter (HTTP), dispatching through the buses.
package rest

import (
	"context"
	"encoding/json"
	"net/http"

	"archview-example-cqrs/internal/order/application/command"
	"archview-example-cqrs/internal/order/application/query"
	"archview-example-cqrs/internal/platform/cqrs"
)

// OrderController turns HTTP requests into commands and queries.
type OrderController struct {
	commands *cqrs.CommandBus
	queries  *cqrs.QueryBus
}

// New builds the controller.
func New(c *cqrs.CommandBus, q *cqrs.QueryBus) *OrderController {
	return &OrderController{commands: c, queries: q}
}

// Create handles POST /orders.
func (ct *OrderController) Create(w http.ResponseWriter, r *http.Request) {
	var body struct {
		ID, Item string
		Qty      int
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if err := ct.commands.Dispatch(context.Background(),
		command.CreateOrder{ID: body.ID, Item: body.Item, Qty: body.Qty}); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

// Cancel handles POST /orders/cancel.
func (ct *OrderController) Cancel(w http.ResponseWriter, r *http.Request) {
	var body struct{ ID string }
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	_ = ct.commands.Dispatch(context.Background(), command.CancelOrder{ID: body.ID})
	w.WriteHeader(http.StatusNoContent)
}

// Get handles GET /orders/{id}.
func (ct *OrderController) Get(w http.ResponseWriter, r *http.Request) {
	res, _ := ct.queries.Dispatch(context.Background(), query.GetOrder{ID: r.PathValue("id")})
	json.NewEncoder(w).Encode(res)
}

// List handles GET /orders.
func (ct *OrderController) List(w http.ResponseWriter, r *http.Request) {
	res, _ := ct.queries.Dispatch(context.Background(), query.ListOrders{})
	json.NewEncoder(w).Encode(res)
}
