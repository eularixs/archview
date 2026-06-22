// Package query holds order read-side use cases (CQRS query handlers).
package query

import (
	"context"

	"archview-example-cqrs/internal/order/port"
	"archview-example-cqrs/internal/platform/cqrs"
)

// GetOrder is a query.
type GetOrder struct{ ID string }

// QueryName satisfies cqrs.Query.
func (GetOrder) QueryName() string { return "order.get" }

// ListOrders is a query.
type ListOrders struct{}

// QueryName satisfies cqrs.Query.
func (ListOrders) QueryName() string { return "order.list" }

// GetOrderHandler reads one order.
type GetOrderHandler struct{ repo port.OrderRepository }

// NewGetOrderHandler wires the handler.
func NewGetOrderHandler(repo port.OrderRepository) *GetOrderHandler {
	return &GetOrderHandler{repo: repo}
}

// Handle satisfies cqrs.QueryHandler.
func (h *GetOrderHandler) Handle(ctx context.Context, q cqrs.Query) (any, error) {
	o, _ := h.repo.FindByID(ctx, q.(GetOrder).ID)
	return o, nil
}

// ListOrdersHandler reads all orders.
type ListOrdersHandler struct{ repo port.OrderRepository }

// NewListOrdersHandler wires the handler.
func NewListOrdersHandler(repo port.OrderRepository) *ListOrdersHandler {
	return &ListOrdersHandler{repo: repo}
}

// Handle satisfies cqrs.QueryHandler.
func (h *ListOrdersHandler) Handle(ctx context.Context, q cqrs.Query) (any, error) {
	return h.repo.FindAll(ctx), nil
}
