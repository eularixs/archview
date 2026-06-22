// Package service holds order business logic (the application core).
package service

import (
	"context"
	"fmt"

	"archview-example-graphql/internal/order/domain"
	"archview-example-graphql/internal/order/port"
)

// OrderService is the inbound use-case boundary.
type OrderService interface {
	Create(ctx context.Context, item string, qty int) (domain.Order, error)
	Get(ctx context.Context, id string) (domain.Order, bool)
	List(ctx context.Context) []domain.Order
	Cancel(ctx context.Context, id string) bool
}

type orderService struct {
	repo port.OrderRepository
	seq  int
}

// New wires the service over its repository port.
func New(repo port.OrderRepository) OrderService { return &orderService{repo: repo} }

func (s *orderService) Create(ctx context.Context, item string, qty int) (domain.Order, error) {
	s.seq++
	o := domain.Order{ID: fmt.Sprintf("ord-%d", s.seq), Item: item, Qty: qty}
	return o, s.repo.Save(ctx, o)
}

func (s *orderService) Get(ctx context.Context, id string) (domain.Order, bool) {
	return s.repo.FindByID(ctx, id)
}

func (s *orderService) List(ctx context.Context) []domain.Order {
	return s.repo.FindAll(ctx)
}

func (s *orderService) Cancel(ctx context.Context, id string) bool {
	return s.repo.Delete(ctx, id)
}
