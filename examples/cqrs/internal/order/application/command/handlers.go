// Package command holds order write-side use cases (CQRS command handlers).
package command

import (
	"context"

	"archview-example-cqrs/internal/order/domain"
	"archview-example-cqrs/internal/order/port"
	"archview-example-cqrs/internal/platform/cqrs"
)

// CreateOrder is a command.
type CreateOrder struct {
	ID, Item string
	Qty      int
}

// CommandName satisfies cqrs.Command.
func (CreateOrder) CommandName() string { return "order.create" }

// CancelOrder is a command.
type CancelOrder struct{ ID string }

// CommandName satisfies cqrs.Command.
func (CancelOrder) CommandName() string { return "order.cancel" }

// CreateOrderHandler persists a new order and publishes OrderCreated.
type CreateOrderHandler struct {
	repo   port.OrderRepository
	events *cqrs.EventBus
}

// NewCreateOrderHandler wires the handler.
func NewCreateOrderHandler(repo port.OrderRepository, ev *cqrs.EventBus) *CreateOrderHandler {
	return &CreateOrderHandler{repo: repo, events: ev}
}

// Handle satisfies cqrs.CommandHandler.
func (h *CreateOrderHandler) Handle(ctx context.Context, c cqrs.Command) error {
	cmd := c.(CreateOrder)
	o := domain.Order{ID: cmd.ID, Item: cmd.Item, Qty: cmd.Qty}
	if err := h.repo.Save(ctx, o); err != nil {
		return err
	}
	h.events.Publish(ctx, domain.OrderCreated{OrderID: o.ID})
	return nil
}

// CancelOrderHandler removes an order and publishes OrderCancelled.
type CancelOrderHandler struct {
	repo   port.OrderRepository
	events *cqrs.EventBus
}

// NewCancelOrderHandler wires the handler.
func NewCancelOrderHandler(repo port.OrderRepository, ev *cqrs.EventBus) *CancelOrderHandler {
	return &CancelOrderHandler{repo: repo, events: ev}
}

// Handle satisfies cqrs.CommandHandler.
func (h *CancelOrderHandler) Handle(ctx context.Context, c cqrs.Command) error {
	cmd := c.(CancelOrder)
	if _, ok := h.repo.FindByID(ctx, cmd.ID); !ok {
		return nil
	}
	h.events.Publish(ctx, domain.OrderCancelled{OrderID: cmd.ID})
	return nil
}
