// Package domain holds the order entity and its events (the core).
package domain

// Order is an order aggregate.
type Order struct {
	ID   string `json:"id"`
	Item string `json:"item"`
	Qty  int    `json:"qty"`
}

// OrderCreated is emitted after an order is persisted.
type OrderCreated struct {
	OrderID string
}

// EventName satisfies cqrs.Event.
func (OrderCreated) EventName() string { return "order.created" }

// OrderCancelled is emitted after an order is cancelled.
type OrderCancelled struct {
	OrderID string
}

// EventName satisfies cqrs.Event.
func (OrderCancelled) EventName() string { return "order.cancelled" }
