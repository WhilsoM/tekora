package product

import "time"

type GetProductsRequest struct {
	Limit  uint32
	Offset uint32
}

type GetProductsResponse struct {
	Items []ProductItemShort `json:"items"`
}

type GetProductByIDResponse struct {
	Item ProductItem `json:"item"`
}

type DeleteProductByIDResponse struct {
	IsSuccess bool   `json:"is_success"`
	Message   string `json:"message"`
}

type CreateProductRequest struct {
	Name     string `json:"name"`
	Price    uint32 `json:"price"`
	Category string `json:"category"`
	Stock    uint32 `json:"stock"`
}

type CreateProductResponse struct {
	Item ProductItem `json:"item"`
}

type UpdateProductByIDRequest struct {
	Name     string `json:"name"`
	Price    uint32 `json:"price"`
	Category string `json:"category"`
	Stock    uint32 `json:"stock"`
}

type UpdateProductByIDResponse struct {
	Item ProductItem `json:"item"`
}

type ProductItem struct {
	ID        uint32    `json:"id"`
	Name      string    `json:"name"`
	Price     uint32    `json:"price"`
	Category  string    `json:"category"`
	Stock     uint32    `json:"stock"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ProductItemShort struct {
	ID    uint32 `json:"id"`
	Name  string `json:"name"`
	Price uint32 `json:"price"`
}
