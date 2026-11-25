package transport

import (
	"context"
	"errors"
	"log"
	productpb "tekora/pkg/gen/product"
	"tekora/services/product/internal/core"

	"google.golang.org/protobuf/types/known/timestamppb"
)

type productService interface {
	GetProducts(ctx context.Context) ([]core.ProductItemShort, error)
	GetProductByID(ctx context.Context, id int64) (core.ProductItem, error)
	DeleteProductByID(ctx context.Context, id int64) error
	CreateProduct(ctx context.Context, payload core.CreateProductPayload) (core.ProductItem, error)
	UpdateProductByID(ctx context.Context, id int64, updates map[string]any) (core.ProductItem, error)
}

type productServer struct {
	productpb.UnimplementedProductServiceServer

	svc productService
}

func NewProductServer(svc productService) *productServer {
	return &productServer{
		svc: svc,
	}
}

func (s *productServer) GetProducts(ctx context.Context, req *productpb.GetProductsRequest) (*productpb.GetProductsResponse, error) {
	_ = &productpb.GetProductsRequest{
		Limit:  req.GetLimit(),
		Offset: req.GetOffset(),
	}

	products, err := s.svc.GetProducts(ctx)
	if err != nil {
		return nil, err
	}

	res := make([]*productpb.ProductItemShort, len(products))

	for i, v := range products {
		res[i] = &productpb.ProductItemShort{
			Id:    v.ID,
			Name:  v.Name,
			Price: v.Price,
		}
	}

	return &productpb.GetProductsResponse{
		Items: res,
	}, nil
}

func (s *productServer) GetProductByID(ctx context.Context, req *productpb.GetProductByIDRequest) (*productpb.GetProductByIDResponse, error) {
	_ = &productpb.GetProductByIDRequest{}

	product, err := s.svc.GetProductByID(ctx, req.GetId())
	if err != nil {
		return nil, err
	}

	return &productpb.GetProductByIDResponse{
		Item: &productpb.ProductItem{
			Id:        product.ID,
			Name:      product.Name,
			Price:     product.Price,
			Category:  product.Category,
			Stock:     product.Stock,
			CreatedAt: timestamppb.New(product.CreatedAt),
			UpdatedAt: timestamppb.New(product.UpdatedAt),
		},
	}, nil
}

func (s *productServer) DeleteProductByID(ctx context.Context, req *productpb.DeleteProductByIDRequest) (*productpb.DeleteProductByIDResponse, error) {
	if err := s.svc.DeleteProductByID(ctx, req.GetId()); err != nil {
		return nil, err
	}

	return &productpb.DeleteProductByIDResponse{
		IsSuccess: true,
		Message:   "successfully deleted",
	}, nil
}

func (s *productServer) CreateProduct(ctx context.Context, req *productpb.CreateProductRequest) (*productpb.CreateProductResponse, error) {
	product, err := s.svc.CreateProduct(ctx, core.CreateProductPayload{
		Name:     req.Name,
		Price:    req.Price,
		Category: req.Category,
		Stock:    req.Stock,
	})
	if err != nil {
		return nil, err
	}

	return &productpb.CreateProductResponse{
		Item: &productpb.ProductItem{
			Id:        product.ID,
			Name:      product.Name,
			Price:     product.Price,
			Category:  product.Category,
			Stock:     product.Stock,
			CreatedAt: timestamppb.New(product.CreatedAt),
			UpdatedAt: timestamppb.New(product.UpdatedAt),
		},
	}, nil
}

func (s *productServer) UpdateProductByID(ctx context.Context, req *productpb.UpdateProductByIDRequest) (*productpb.UpdateProductByIDResponse, error) {
	updates := make(map[string]any)

	if req.Name != nil {
		updates["Name"] = *req.Name
	}

	if req.Price != nil {
		updates["Price"] = *req.Price
	}

	if req.Category != nil {
		updates["Category"] = *req.Category
	}

	if req.Stock != nil {
		updates["Stock"] = *req.Stock
	}

	if len(updates) == 0 {
		return nil, errors.New("no updates")
	}
	log.Printf(
		"fields: name=%v price=%v category=%v stock=%v",
		req.Name != nil,
		req.Price != nil,
		req.Category != nil,
		req.Stock != nil,
	)

	log.Printf("updates: %#v", updates)
	product, err := s.svc.UpdateProductByID(ctx, req.GetId(), updates)
	if err != nil {
		return nil, err
	}

	return &productpb.UpdateProductByIDResponse{
		Item: &productpb.ProductItem{
			Id:        product.ID,
			Name:      product.Name,
			Price:     product.Price,
			Category:  product.Category,
			Stock:     product.Stock,
			CreatedAt: timestamppb.New(product.CreatedAt),
			UpdatedAt: timestamppb.New(product.UpdatedAt),
		},
	}, nil
}
