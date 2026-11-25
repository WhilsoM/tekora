package psql

import "context"

type orderRepository struct {
}

func NewOrderRepository() *orderRepository {
	return &orderRepository{}
}

func (svc *orderRepository) CreateOrder(ctx context.Context) {

}

func (svc *orderRepository) GetOrderByID(ctx context.Context) {

}

func (svc *orderRepository) DeleteOrderByID(ctx context.Context) {

}

func (svc *orderRepository) UpdateOrderByID(ctx context.Context) {

}
