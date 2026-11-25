package core

import "context"

type productRepository interface {
	GetProducts(context.Context) ([]ProductItemShort, error)
	GetProductByID(context.Context, int64) (ProductItem, error)
	DeleteProductByID(context.Context, int64) error
	CreateProduct(context.Context, CreateProductPayload) (ProductItem, error)
	UpdateProductByID(context.Context, int64, map[string]any) (ProductItem, error)
}

type productService struct {
	repo productRepository
}

func NewProductService(repo productRepository) *productService {
	return &productService{
		repo: repo,
	}
}

func (svc *productService) GetProducts(ctx context.Context) ([]ProductItemShort, error) {
	return svc.repo.GetProducts(ctx)
}

func (svc *productService) GetProductByID(ctx context.Context, id int64) (ProductItem, error) {
	return svc.repo.GetProductByID(ctx, id)
}

func (svc *productService) DeleteProductByID(ctx context.Context, id int64) error {
	return svc.repo.DeleteProductByID(ctx, id)
}

func (svc *productService) CreateProduct(ctx context.Context, payload CreateProductPayload) (ProductItem, error) {
	return svc.repo.CreateProduct(ctx, payload)
}

func (svc *productService) UpdateProductByID(ctx context.Context, id int64, updates map[string]any) (ProductItem, error) {
	return svc.repo.UpdateProductByID(ctx, id, updates)
}
