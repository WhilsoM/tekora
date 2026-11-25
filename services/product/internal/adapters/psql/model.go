package psql

import "time"

type GetProducts struct {
	Limit  uint32
	Offset uint32
}

func (GetProducts) TableName() string {
	return "products"
}

type ProductItemShort struct {
	ID    int64
	Name  string
	Price uint32
}

type ProductItem struct {
	ID        int64
	Name      string
	Price     uint32
	Category  string
	Stock     uint32
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (ProductItem) TableName() string {
	return "products"
}
