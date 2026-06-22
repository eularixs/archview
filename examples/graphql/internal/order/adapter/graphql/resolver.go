// Package graphql is the order inbound adapter (GraphQL driving side). The
// QueryResolver/MutationResolver interfaces mirror gqlgen-generated code; the
// queryResolver/mutationResolver types implement them and delegate to the
// service. archview surfaces each resolver field as an endpoint.
package graphql

import (
	"context"

	"archview-example-graphql/internal/order/domain"
	"archview-example-graphql/internal/order/service"
)

// --- generated-style resolver interfaces (mirrors gqlgen output) ---

// QueryResolver resolves Query.* fields.
type QueryResolver interface {
	Order(ctx context.Context, id string) (*domain.Order, error)
	Orders(ctx context.Context) ([]*domain.Order, error)
}

// MutationResolver resolves Mutation.* fields.
type MutationResolver interface {
	CreateOrder(ctx context.Context, item string, qty int) (*domain.Order, error)
	CancelOrder(ctx context.Context, id string) (bool, error)
}

// ResolverRoot wires the root resolvers.
type ResolverRoot interface {
	Query() QueryResolver
	Mutation() MutationResolver
}

// --- user resolvers ---

// Resolver is the root resolver holding dependencies.
type Resolver struct{ svc service.OrderService }

// NewResolver builds the root resolver.
func NewResolver(svc service.OrderService) *Resolver { return &Resolver{svc: svc} }

// Query returns the query resolver.
func (r *Resolver) Query() QueryResolver { return &queryResolver{r} }

// Mutation returns the mutation resolver.
func (r *Resolver) Mutation() MutationResolver { return &mutationResolver{r} }

type queryResolver struct{ *Resolver }

func (r *queryResolver) Order(ctx context.Context, id string) (*domain.Order, error) {
	o, _ := r.svc.Get(ctx, id)
	return &o, nil
}

func (r *queryResolver) Orders(ctx context.Context) ([]*domain.Order, error) {
	list := r.svc.List(ctx)
	out := make([]*domain.Order, len(list))
	for i := range list {
		out[i] = &list[i]
	}
	return out, nil
}

type mutationResolver struct{ *Resolver }

func (r *mutationResolver) CreateOrder(ctx context.Context, item string, qty int) (*domain.Order, error) {
	o, err := r.svc.Create(ctx, item, qty)
	return &o, err
}

func (r *mutationResolver) CancelOrder(ctx context.Context, id string) (bool, error) {
	return r.svc.Cancel(ctx, id), nil
}
