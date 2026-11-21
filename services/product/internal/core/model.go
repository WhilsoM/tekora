package core

import "time"

type CreateProductPayload struct {
	Name     string
	Price    uint32
	Category string
	Stock    uint32
}

type ProductItem struct {
	ID        uint32
	Name      string
	Price     uint32
	Category  string
	Stock     uint32
	CreatedAt time.Time
	UpdatedAt time.Time
}

type ProductItemShort struct {
	ID    uint32
	Name  string
	Price uint32
}
