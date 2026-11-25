package psql

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"tekora/services/product/internal/core"
	"tekora/services/product/internal/testutils"

	"gorm.io/gorm"
)

func setupTestRepository(t *testing.T) (*gorm.DB, *productRepository) {
	t.Helper()

	migrationsDir := filepath.Join(
		"..",
		"..",
		"..",
		"migrations",
	)

	db := testutils.NewPostgresTestContainer(t, migrationsDir)
	repo := NewProductRepository(db)

	return db, repo
}

func TestGetProducts(t *testing.T) {
	db, repo := setupTestRepository(t)

	products := []ProductItem{
		{
			Name:     "Mechanical Keyboard",
			Price:    99,
			Category: "Electronics",
			Stock:    10,
		},
		{
			Name:     "Gaming Mouse",
			Price:    49,
			Category: "Electronics",
			Stock:    20,
		},
	}

	if err := db.Create(&products).Error; err != nil {
		t.Fatalf("failed to create test products: %v", err)
	}

	got, err := repo.GetProducts(context.Background())

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(got) != len(products) {
		t.Fatalf("expected %d products, got %d", len(products), len(got))
	}

	for i, product := range products {
		if got[i].ID != product.ID {
			t.Errorf("expected ID %d, got %d", product.ID, got[i].ID)
		}

		if got[i].Name != product.Name {
			t.Errorf("expected name %q, got %q", product.Name, got[i].Name)
		}

		if got[i].Price != product.Price {
			t.Errorf("expected price %v, got %v", product.Price, got[i].Price)
		}
	}
}

func TestGetProducts_EmptyDatabase(t *testing.T) {
	_, repo := setupTestRepository(t)

	got, err := repo.GetProducts(context.Background())

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(got) != 0 {
		t.Errorf("expected empty products slice, got %d products", len(got))
	}
}

func TestGetProductByID(t *testing.T) {
	db, repo := setupTestRepository(t)

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
		t.Errorf("expected name %q, got %q", product.Name, got.Name)
	}

	if got.Price != product.Price {
		t.Errorf("expected price %v, got %v", product.Price, got.Price)
	}

	if got.Category != product.Category {
		t.Errorf("expected category %q, got %q", product.Category, got.Category)
	}

	if got.Stock != product.Stock {
		t.Errorf("expected stock %d, got %d", product.Stock, got.Stock)
	}
}

func TestGetProductByID_NotFound(t *testing.T) {
	_, repo := setupTestRepository(t)

	_, err := repo.GetProductByID(context.Background(), 999)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Errorf("expected gorm.ErrRecordNotFound, got %v", err)
	}
}

func TestCreateProduct(t *testing.T) {
	db, repo := setupTestRepository(t)

	payload := core.CreateProductPayload{
		Name:     "Mechanical Keyboard",
		Price:    99,
		Category: "Electronics",
		Stock:    10,
	}

	got, err := repo.CreateProduct(context.Background(), payload)

	if err != nil {
		t.Fatalf("failed to create product: %v", err)
	}

	if got.ID == 0 {
		t.Error("expected generated product ID, got 0")
	}

	if got.Name != payload.Name {
		t.Errorf("expected name %q, got %q", payload.Name, got.Name)
	}

	if got.Price != payload.Price {
		t.Errorf("expected price %v, got %v", payload.Price, got.Price)
	}

	if got.Category != payload.Category {
		t.Errorf("expected category %q, got %q", payload.Category, got.Category)
	}

	if got.Stock != payload.Stock {
		t.Errorf("expected stock %d, got %d", payload.Stock, got.Stock)
	}

	var stored ProductItem

	if err := db.First(&stored, got.ID).Error; err != nil {
		t.Fatalf("product was not persisted: %v", err)
	}

	if stored.Name != payload.Name {
		t.Errorf("expected stored name %q, got %q", payload.Name, stored.Name)
	}
}

func TestDeleteProductByID(t *testing.T) {
	db, repo := setupTestRepository(t)

	product := ProductItem{
		Name:     "Mechanical Keyboard",
		Price:    99,
		Category: "Electronics",
		Stock:    10,
	}

	if err := db.Create(&product).Error; err != nil {
		t.Fatalf("failed to create test product: %v", err)
	}

	err := repo.DeleteProductByID(context.Background(), product.ID)

	if err != nil {
		t.Fatalf("failed to delete product: %v", err)
	}

	var stored ProductItem

	err = db.First(&stored, product.ID).Error

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Errorf("expected product to be deleted, got error: %v", err)
	}
}

func TestDeleteProductByID_NotFound(t *testing.T) {
	_, repo := setupTestRepository(t)

	err := repo.DeleteProductByID(context.Background(), 999)

	if !errors.Is(err, core.ErrNotFoundProduct) {
		t.Errorf("expected ErrNotFoundProduct, got %v", err)
	}
}

func TestUpdateProductByID(t *testing.T) {
	db, repo := setupTestRepository(t)

	product := ProductItem{
		Name:     "Keyboard",
		Price:    99,
		Category: "Electronics",
		Stock:    10,
	}

	if err := db.Create(&product).Error; err != nil {
		t.Fatalf("failed to create test product: %v", err)
	}

	updates := map[string]any{
		"name":  "Mechanical Keyboard",
		"price": 150,
	}

	got, err := repo.UpdateProductByID(
		context.Background(),
		product.ID,
		updates,
	)

	if err != nil {
		t.Fatalf("failed to update product: %v", err)
	}

	if got.ID != product.ID {
		t.Errorf("expected ID %d, got %d", product.ID, got.ID)
	}

	if got.Name != "Mechanical Keyboard" {
		t.Errorf("expected updated name, got %q", got.Name)
	}

	if got.Price != 150 {
		t.Errorf("expected updated price 150, got %v", got.Price)
	}

	if got.Category != product.Category {
		t.Errorf("expected category to remain %q, got %q", product.Category, got.Category)
	}

	if got.Stock != product.Stock {
		t.Errorf("expected stock to remain %d, got %d", product.Stock, got.Stock)
	}

	var stored ProductItem

	if err := db.First(&stored, product.ID).Error; err != nil {
		t.Fatalf("failed to retrieve updated product: %v", err)
	}

	if stored.Name != "Mechanical Keyboard" {
		t.Errorf("expected persisted name %q, got %q", "Mechanical Keyboard", stored.Name)
	}
}

func TestUpdateProductByID_NotFound(t *testing.T) {
	_, repo := setupTestRepository(t)

	updates := map[string]any{
		"name": "New Keyboard",
	}

	_, err := repo.UpdateProductByID(
		context.Background(),
		999,
		updates,
	)

	if !errors.Is(err, core.ErrNotFoundProduct) {
		t.Errorf("expected ErrNotFoundProduct, got %v", err)
	}
}

func TestUpdateProductByID_EmptyUpdates(t *testing.T) {
	_, repo := setupTestRepository(t)

	_, err := repo.UpdateProductByID(
		context.Background(),
		1,
		map[string]any{},
	)

	if err == nil {
		t.Fatal("expected error for empty updates, got nil")
	}

	if err.Error() != "no fields to update" {
		t.Errorf("expected 'no fields to update', got %q", err.Error())
	}
}
