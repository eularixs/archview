// Package postgres is the order outbound adapter for persistence.
package postgres

import (
	"context"

	"archview-example-graphql/internal/order/domain"
	"archview-example-graphql/internal/order/port"
)

type orderRepository struct{ rows map[string]domain.Order }

// New returns an in-memory OrderRepository.
func New() port.OrderRepository { return &orderRepository{rows: map[string]domain.Order{}} }

func (r *orderRepository) Save(ctx context.Context, o domain.Order) error {
	r.rows[o.ID] = o
	return nil
}

func (r *orderRepository) FindByID(ctx context.Context, id string) (domain.Order, bool) {
	o, ok := r.rows[id]
	return o, ok
}

func (r *orderRepository) FindAll(ctx context.Context) []domain.Order {
	out := make([]domain.Order, 0, len(r.rows))
	for _, o := range r.rows {
		out = append(out, o)
	}
	return out
}

func (r *orderRepository) Delete(ctx context.Context, id string) bool {
	if _, ok := r.rows[id]; !ok {
		return false
	}
	delete(r.rows, id)
	return true
}
