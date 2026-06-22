// Package grpc is the order inbound adapter (gRPC driving side). Each RPC method
// delegates to the order service.
package grpc

import (
	"context"

	"archview-example-grpc/api/orderpb"
	"archview-example-grpc/internal/order/service"
)

// OrderServer implements orderpb.OrderServiceServer.
type OrderServer struct {
	orderpb.UnimplementedOrderServiceServer
	svc service.OrderService
}

// New builds the gRPC adapter.
func New(svc service.OrderService) *OrderServer { return &OrderServer{svc: svc} }

// CreateOrder handles the CreateOrder RPC.
func (s *OrderServer) CreateOrder(ctx context.Context, req *orderpb.CreateOrderRequest) (*orderpb.CreateOrderResponse, error) {
	o, err := s.svc.Create(ctx, req.Item, int(req.Qty))
	if err != nil {
		return nil, err
	}
	return &orderpb.CreateOrderResponse{ID: o.ID}, nil
}

// GetOrder handles the GetOrder RPC.
func (s *OrderServer) GetOrder(ctx context.Context, req *orderpb.GetOrderRequest) (*orderpb.GetOrderResponse, error) {
	o, _ := s.svc.Get(ctx, req.ID)
	return &orderpb.GetOrderResponse{ID: o.ID, Item: o.Item}, nil
}

// ListOrders handles the ListOrders RPC.
func (s *OrderServer) ListOrders(ctx context.Context, req *orderpb.ListOrdersRequest) (*orderpb.ListOrdersResponse, error) {
	var ids []string
	for _, o := range s.svc.List(ctx) {
		ids = append(ids, o.ID)
	}
	return &orderpb.ListOrdersResponse{IDs: ids}, nil
}
