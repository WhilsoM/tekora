package transport

import (
	"context"
	orderpb "tekora/pkg/gen/order"
)

type orderService interface {
	CreateOrder(ctx context.Context)
	GetOrderByID(ctx context.Context)
	DeleteOrderByID(ctx context.Context)
	UpdateOrderByID(ctx context.Context)
}

type orderServer struct {
	orderpb.UnimplementedOrderServiceServer
	svc orderService
}

func NewOrderServer(svc orderService) *orderServer {
	return &orderServer{
		svc: svc,
	}
}

func (server *orderServer) CreateOrder(ctx context.Context, req *orderpb.CreateOrderRequest) (*orderpb.CreateOrderResponse, error) {
	server.svc.CreateOrder(ctx)
	return &orderpb.CreateOrderResponse{}, nil
}

func (server *orderServer) GetOrderByID(ctx context.Context, req *orderpb.GetOrderByIDRequest) (*orderpb.GetOrderByIDResponse, error) {
	server.svc.GetOrderByID(ctx)

	return &orderpb.GetOrderByIDResponse{}, nil

}

func (server *orderServer) DeleteOrderByID(ctx context.Context, req *orderpb.DeleteOrderByIDRequest) (*orderpb.DeleteOrderByIDRequest, error) {
	server.svc.DeleteOrderByID(ctx)

	return &orderpb.DeleteOrderByIDRequest{}, nil

}

func (server *orderServer) UpdateOrderByID(ctx context.Context, req *orderpb.UpdateOrderByIDRequest) (*orderpb.UpdateOrderByIDResponse, error) {
	server.svc.UpdateOrderByID(ctx)

	return &orderpb.UpdateOrderByIDResponse{}, nil

}
