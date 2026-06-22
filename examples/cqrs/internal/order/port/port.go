// Package port declares the order outbound port (driven side).
package port

import (
	"context"

	"archview-example-cqrs/internal/order/domain"
)

// OrderRepository is the outbound persistence port. Impl: postgres.orderRepository.
type OrderRepository interface {
	Save(ctx context.Context, o domain.Order) error
	FindByID(ctx context.Context, id string) (domain.Order, bool)
	FindAll(ctx context.Context) []domain.Order
}
