// Package orderpb is a hand-written stand-in for protoc-gen-go-grpc output:
// the service interface, message types and the Register function archview keys
// off. Real generated code has the same shape.
package orderpb

import (
	"context"

	"archview-example-grpc/internal/platform/grpcrt"
)

// Messages.
type CreateOrderRequest struct {
	Item string
	Qty  int32
}
type CreateOrderResponse struct{ ID string }
type GetOrderRequest struct{ ID string }
type GetOrderResponse struct {
	ID   string
	Item string
}
type ListOrdersRequest struct{}
type ListOrdersResponse struct{ IDs []string }

// OrderServiceServer is the server API for the OrderService service.
type OrderServiceServer interface {
	CreateOrder(context.Context, *CreateOrderRequest) (*CreateOrderResponse, error)
	GetOrder(context.Context, *GetOrderRequest) (*GetOrderResponse, error)
	ListOrders(context.Context, *ListOrdersRequest) (*ListOrdersResponse, error)
	mustEmbedUnimplementedOrderServiceServer()
}

// UnimplementedOrderServiceServer must be embedded for forward compatibility.
type UnimplementedOrderServiceServer struct{}

func (UnimplementedOrderServiceServer) CreateOrder(context.Context, *CreateOrderRequest) (*CreateOrderResponse, error) {
	return nil, nil
}
func (UnimplementedOrderServiceServer) GetOrder(context.Context, *GetOrderRequest) (*GetOrderResponse, error) {
	return nil, nil
}
func (UnimplementedOrderServiceServer) ListOrders(context.Context, *ListOrdersRequest) (*ListOrdersResponse, error) {
	return nil, nil
}
func (UnimplementedOrderServiceServer) mustEmbedUnimplementedOrderServiceServer() {}

// ServiceDesc is the grpc.ServiceDesc for OrderService.
var ServiceDesc = grpcrt.ServiceDesc{
	ServiceName: "order.OrderService",
	HandlerType: (*OrderServiceServer)(nil),
}

// RegisterOrderServiceServer registers the service implementation.
func RegisterOrderServiceServer(s grpcrt.ServiceRegistrar, srv OrderServiceServer) {
	s.RegisterService(&ServiceDesc, srv)
}
