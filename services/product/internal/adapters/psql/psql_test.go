package psql

import (
	"context"
	"tekora/services/product/internal/testutils"
	"testing"

	"path/filepath"
)

func TestGetProductByID(t *testing.T) {
	migrationsDir := filepath.Join(
		"..",
		"..",
		"..",
		"migrations",
	)

	db := testutils.NewPostgresTestContainer(t, migrationsDir)
	repo := NewProductRepository(db)

	product := ProductItem{
		Name:     "Mechanical Keyboard",
		Price:    99,
		Category: "Electronics",
		Stock:    10,
	}

	if err := db.Create(&product).Error; err != nil {
		t.Fatalf("failed to create test product: %v", err)
	}

	got, err := repo.GetProductByID(context.Background(), product.ID)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got.ID != product.ID {
		t.Errorf("expected ID %d, got %d", product.ID, got.ID)
	}

	if got.Name != product.Name {
		t.Errorf("expected name %s, got %s", product.Name, got.Name)
	}
}
