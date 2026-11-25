package core

import "context"

type orderRepository interface {
	CreateOrder(ctx context.Context)
	GetOrderByID(ctx context.Context)
	DeleteOrderByID(ctx context.Context)
	UpdateOrderByID(ctx context.Context)
}

type orderService struct {
	repo orderRepository
}

func NewOrderService(repo orderRepository) *orderService {
	return &orderService{
		repo,
	}
}

func (svc *orderService) CreateOrder(ctx context.Context) {
	svc.repo.CreateOrder(ctx)
}

func (svc *orderService) GetOrderByID(ctx context.Context) {
	svc.repo.GetOrderByID(ctx)
}

func (svc *orderService) DeleteOrderByID(ctx context.Context) {
	svc.repo.DeleteOrderByID(ctx)
}

func (svc *orderService) UpdateOrderByID(ctx context.Context) {
	svc.repo.UpdateOrderByID(ctx)
}
