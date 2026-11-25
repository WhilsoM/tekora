package psql

import (
	"context"
	"errors"
	"fmt"
	"tekora/services/product/internal/core"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type productRepository struct {
	gorm *gorm.DB
}

func NewProductRepository(gorm *gorm.DB) *productRepository {
	return &productRepository{
		gorm,
	}
}

func (db *productRepository) GetProducts(ctx context.Context) ([]core.ProductItemShort, error) {
	var products []ProductItem

	err := db.gorm.WithContext(ctx).Find(&products).Error
	if err != nil {
		return nil, err
	}

	returnedProducts := make([]core.ProductItemShort, len(products))

	for i, v := range products {
		returnedProducts[i] = core.ProductItemShort{
			ID:    v.ID,
			Name:  v.Name,
			Price: v.Price,
		}
	}

	return returnedProducts, nil
}

func (db *productRepository) GetProductByID(ctx context.Context, id int64) (core.ProductItem, error) {
	var product ProductItem
	err := db.gorm.WithContext(ctx).Where("id = ?", id).First(&product).Error
	if err != nil {
		return core.ProductItem{}, fmt.Errorf("get product by id: %w", err)
	}

	return core.ProductItem{
		ID:        product.ID,
		Name:      product.Name,
		Price:     product.Price,
		Category:  product.Category,
		Stock:     product.Stock,
		CreatedAt: product.CreatedAt,
		UpdatedAt: product.UpdatedAt,
	}, nil
}

func (db *productRepository) DeleteProductByID(ctx context.Context, id int64) error {
	var product ProductItem
	res := db.gorm.WithContext(ctx).Where("id = ?", id).Delete(&product)
	if res.Error != nil {
		return fmt.Errorf("delete product by id: %w", res.Error)
	}

	if res.RowsAffected == 0 {
		return core.ErrNotFoundProduct
	}

	return nil
}

func (db *productRepository) CreateProduct(ctx context.Context, payload core.CreateProductPayload) (core.ProductItem, error) {
	product := ProductItem{
		Name:     payload.Name,
		Price:    payload.Price,
		Category: payload.Category,
		Stock:    payload.Stock,
	}

	err := db.gorm.WithContext(ctx).Create(&product).Error
	if err != nil {
		return core.ProductItem{}, fmt.Errorf("create product: %w", err)
	}

	return core.ProductItem{
		ID:        product.ID,
		Name:      product.Name,
		Price:     product.Price,
		Category:  product.Category,
		Stock:     product.Stock,
		CreatedAt: product.CreatedAt,
		UpdatedAt: product.UpdatedAt,
	}, nil
}

func (db *productRepository) UpdateProductByID(ctx context.Context, id int64, updates map[string]any) (core.ProductItem, error) {
	if len(updates) == 0 {
		return core.ProductItem{}, errors.New("no fields to update")
	}

	var product ProductItem
	result := db.gorm.WithContext(ctx).Model(&product).Clauses(clause.Returning{}).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		return core.ProductItem{}, fmt.Errorf("failed to update product: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return core.ProductItem{}, core.ErrNotFoundProduct
	}

	return core.ProductItem{
		ID:        product.ID,
		Name:      product.Name,
		Price:     product.Price,
		Category:  product.Category,
		Stock:     product.Stock,
		CreatedAt: product.CreatedAt,
		UpdatedAt: product.UpdatedAt,
	}, nil
}
