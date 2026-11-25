package transport

import (
	"context"
	"errors"
	"testing"
	"time"

	productpb "tekora/pkg/gen/product"
	"tekora/services/product/internal/core"
)

type productServiceStub struct {
	getProductsFn    func(context.Context) ([]core.ProductItemShort, error)
	getProductByIDFn func(context.Context, uint32) (core.ProductItem, error)
	deleteProductFn  func(context.Context, uint32) error
	createProductFn  func(context.Context, core.CreateProductPayload) (core.ProductItem, error)
	updateProductFn  func(context.Context, uint32, map[string]any) (core.ProductItem, error)
}

func (s *productServiceStub) GetProducts(ctx context.Context) ([]core.ProductItemShort, error) {
	if s.getProductsFn == nil {
		panic("unexpected call: GetProducts")
	}
	return s.getProductsFn(ctx)
}

func (s *productServiceStub) GetProductByID(ctx context.Context, id uint32) (core.ProductItem, error) {
	if s.getProductByIDFn == nil {
		panic("unexpected call: GetProductByID")
	}
	return s.getProductByIDFn(ctx, id)
}

func (s *productServiceStub) DeleteProductByID(ctx context.Context, id uint32) error {
	if s.deleteProductFn == nil {
		panic("unexpected call: DeleteProductByID")
	}
	return s.deleteProductFn(ctx, id)
}

func (s *productServiceStub) CreateProduct(ctx context.Context, payload core.CreateProductPayload) (core.ProductItem, error) {
	if s.createProductFn == nil {
		panic("unexpected call: CreateProduct")
	}
	return s.createProductFn(ctx, payload)
}

func (s *productServiceStub) UpdateProductByID(ctx context.Context, id uint32, updates map[string]any) (core.ProductItem, error) {
	if s.updateProductFn == nil {
		panic("unexpected call: UpdateProductByID")
	}
	return s.updateProductFn(ctx, id, updates)
}

func testProduct() core.ProductItem {
	return core.ProductItem{
		ID:        42,
		Name:      "Mechanical Keyboard",
		Price:     149,
		Category:  "Electronics",
		Stock:     15,
		CreatedAt: time.Date(2026, 11, 23, 10, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2026, 11, 23, 11, 0, 0, 0, time.UTC),
	}
}

func assertProduct(t *testing.T, got *productpb.ProductItem, want core.ProductItem) {
	t.Helper()

	if got == nil {
		t.Fatal("expected product, got nil")
	}

	if got.Id != want.ID {
		t.Errorf("expected ID %d, got %d", want.ID, got.Id)
	}

	if got.Name != want.Name {
		t.Errorf("expected name %q, got %q", want.Name, got.Name)
	}

	if got.Price != want.Price {
		t.Errorf("expected price %v, got %v", want.Price, got.Price)
	}

	if got.Category != want.Category {
		t.Errorf("expected category %q, got %q", want.Category, got.Category)
	}

	if got.Stock != want.Stock {
		t.Errorf("expected stock %v, got %v", want.Stock, got.Stock)
	}

	if got.CreatedAt == nil || !got.CreatedAt.AsTime().Equal(want.CreatedAt) {
		t.Errorf("unexpected CreatedAt: %v", got.CreatedAt)
	}

	if got.UpdatedAt == nil || !got.UpdatedAt.AsTime().Equal(want.UpdatedAt) {
		t.Errorf("unexpected UpdatedAt: %v", got.UpdatedAt)
	}
}

func TestProductServer_GetProducts(t *testing.T) {

	expectedProducts := []core.ProductItemShort{
		{ID: 1, Name: "Keyboard", Price: 99},
		{ID: 2, Name: "Mouse", Price: 49},
	}

	fakeService := &productServiceStub{
		getProductsFn: func(ctx context.Context) ([]core.ProductItemShort, error) {
			return expectedProducts, nil
		},
	}

	server := NewProductServer(fakeService)

	req := &productpb.GetProductsRequest{}

	got, err := server.GetProducts(context.Background(), req)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got == nil {
		t.Fatal("expected response, got nil")
	}

	if len(got.Items) != len(expectedProducts) {
		t.Fatalf("expected %d products, got %d", len(expectedProducts), len(got.Items))
	}

	for i, want := range expectedProducts {
		item := got.Items[i]

		if item.Id != want.ID {
			t.Errorf("item %d: expected ID %d, got %d", i, want.ID, item.Id)
		}

		if item.Name != want.Name {
			t.Errorf("item %d: expected name %q, got %q", i, want.Name, item.Name)
		}

		if item.Price != want.Price {
			t.Errorf("item %d: expected price %v, got %v", i, want.Price, item.Price)
		}
	}
}

func TestProductServer_GetProducts_ServiceError(t *testing.T) {

	serviceErr := errors.New("database unavailable")

	fakeService := &productServiceStub{
		getProductsFn: func(ctx context.Context) ([]core.ProductItemShort, error) {
			return nil, serviceErr
		},
	}

	server := NewProductServer(fakeService)

	got, err := server.GetProducts(context.Background(), &productpb.GetProductsRequest{})

	if !errors.Is(err, serviceErr) {
		t.Fatalf("expected service error, got %v", err)
	}

	if got != nil {
		t.Errorf("expected nil response, got %v", got)
	}
}

func TestProductServer_GetProductByID(t *testing.T) {

	expectedProduct := testProduct()

	var receivedID uint32

	fakeService := &productServiceStub{
		getProductByIDFn: func(ctx context.Context, id uint32) (core.ProductItem, error) {
			receivedID = id
			return expectedProduct, nil
		},
	}

	server := NewProductServer(fakeService)

	req := &productpb.GetProductByIDRequest{
		Id: 42,
	}

	got, err := server.GetProductByID(context.Background(), req)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if receivedID != req.Id {
		t.Errorf("expected service ID %d, got %d", req.Id, receivedID)
	}

	if got == nil {
		t.Fatal("expected response, got nil")
	}

	assertProduct(t, got.Item, expectedProduct)
}

func TestProductServer_GetProductByID_ServiceError(t *testing.T) {

	serviceErr := errors.New("product not found")

	fakeService := &productServiceStub{
		getProductByIDFn: func(ctx context.Context, id uint32) (core.ProductItem, error) {
			return core.ProductItem{}, serviceErr
		},
	}

	server := NewProductServer(fakeService)

	got, err := server.GetProductByID(
		context.Background(),
		&productpb.GetProductByIDRequest{Id: 42},
	)

	if !errors.Is(err, serviceErr) {
		t.Fatalf("expected service error, got %v", err)
	}

	if got != nil {
		t.Errorf("expected nil response, got %v", got)
	}
}

func TestProductServer_DeleteProductByID(t *testing.T) {

	var receivedID uint32

	fakeService := &productServiceStub{
		deleteProductFn: func(ctx context.Context, id uint32) error {
			receivedID = id
			return nil
		},
	}

	server := NewProductServer(fakeService)

	req := &productpb.DeleteProductByIDRequest{
		Id: 42,
	}

	got, err := server.DeleteProductByID(context.Background(), req)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if receivedID != req.Id {
		t.Errorf("expected ID %d, got %d", req.Id, receivedID)
	}

	if got == nil {
		t.Fatal("expected response, got nil")
	}

	if !got.IsSuccess {
		t.Error("expected IsSuccess to be true")
	}

	if got.Message != "successfully deleted" {
		t.Errorf("unexpected message: %q", got.Message)
	}
}

func TestProductServer_DeleteProductByID_ServiceError(t *testing.T) {

	serviceErr := errors.New("delete failed")

	fakeService := &productServiceStub{
		deleteProductFn: func(ctx context.Context, id uint32) error {
			return serviceErr
		},
	}

	server := NewProductServer(fakeService)

	got, err := server.DeleteProductByID(
		context.Background(),
		&productpb.DeleteProductByIDRequest{Id: 42},
	)

	if !errors.Is(err, serviceErr) {
		t.Fatalf("expected service error, got %v", err)
	}

	if got != nil {
		t.Errorf("expected nil response, got %v", got)
	}
}

func TestProductServer_CreateProduct(t *testing.T) {

	expectedProduct := testProduct()

	var receivedPayload core.CreateProductPayload

	fakeService := &productServiceStub{
		createProductFn: func(ctx context.Context, payload core.CreateProductPayload) (core.ProductItem, error) {
			receivedPayload = payload
			return expectedProduct, nil
		},
	}

	server := NewProductServer(fakeService)

	req := &productpb.CreateProductRequest{
		Name:     "Mechanical Keyboard",
		Price:    149,
		Category: "Electronics",
		Stock:    15,
	}

	got, err := server.CreateProduct(context.Background(), req)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if receivedPayload.Name != req.Name {
		t.Errorf("expected name %q, got %q", req.Name, receivedPayload.Name)
	}

	if receivedPayload.Price != req.Price {
		t.Errorf("expected price %v, got %v", req.Price, receivedPayload.Price)
	}

	if receivedPayload.Category != req.Category {
		t.Errorf("expected category %q, got %q", req.Category, receivedPayload.Category)
	}

	if receivedPayload.Stock != req.Stock {
		t.Errorf("expected stock %v, got %v", req.Stock, receivedPayload.Stock)
	}

	if got == nil {
		t.Fatal("expected response, got nil")
	}

	assertProduct(t, got.Item, expectedProduct)
}

func TestProductServer_CreateProduct_ServiceError(t *testing.T) {

	serviceErr := errors.New("create failed")

	fakeService := &productServiceStub{
		createProductFn: func(ctx context.Context, payload core.CreateProductPayload) (core.ProductItem, error) {
			return core.ProductItem{}, serviceErr
		},
	}

	server := NewProductServer(fakeService)

	got, err := server.CreateProduct(
		context.Background(),
		&productpb.CreateProductRequest{
			Name:     "Keyboard",
			Price:    99,
			Category: "Electronics",
			Stock:    10,
		},
	)

	if !errors.Is(err, serviceErr) {
		t.Fatalf("expected service error, got %v", err)
	}

	if got != nil {
		t.Errorf("expected nil response, got %v", got)
	}
}

func TestProductServer_UpdateProductByID(t *testing.T) {

	expectedProduct := testProduct()

	name := "Updated Keyboard"
	price := uint32(199)
	category := "Hardware"
	stock := uint32(20)

	var receivedID uint32
	var receivedUpdates map[string]any

	fakeService := &productServiceStub{
		updateProductFn: func(ctx context.Context, id uint32, updates map[string]any) (core.ProductItem, error) {
			receivedID = id
			receivedUpdates = updates
			return expectedProduct, nil
		},
	}

	server := NewProductServer(fakeService)

	req := &productpb.UpdateProductByIDRequest{
		Id:       42,
		Name:     &name,
		Price:    &price,
		Category: &category,
		Stock:    &stock,
	}

	got, err := server.UpdateProductByID(context.Background(), req)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if receivedID != req.Id {
		t.Errorf("expected ID %d, got %d", req.Id, receivedID)
	}

	if len(receivedUpdates) != 4 {
		t.Fatalf("expected 4 updates, got %d", len(receivedUpdates))
	}

	if receivedUpdates["Name"] != name {
		t.Errorf("unexpected Name update: %v", receivedUpdates["Name"])
	}

	if receivedUpdates["Price"] != price {
		t.Errorf("unexpected Price update: %v", receivedUpdates["Price"])
	}

	if receivedUpdates["Category"] != category {
		t.Errorf("unexpected Category update: %v", receivedUpdates["Category"])
	}

	if receivedUpdates["Stock"] != stock {
		t.Errorf("unexpected Stock update: %v", receivedUpdates["Stock"])
	}

	if got == nil {
		t.Fatal("expected response, got nil")
	}

	assertProduct(t, got.Item, expectedProduct)
}

func TestProductServer_UpdateProductByID_NoUpdates(t *testing.T) {

	server := NewProductServer(&productServiceStub{})

	req := &productpb.UpdateProductByIDRequest{
		Id: 42,
	}

	got, err := server.UpdateProductByID(context.Background(), req)

	if err == nil {
		t.Fatal("expected error when no updates are provided")
	}

	if err.Error() != "no updates" {
		t.Errorf("expected 'no updates' error, got %q", err.Error())
	}

	if got != nil {
		t.Errorf("expected nil response, got %v", got)
	}
}

func TestProductServer_UpdateProductByID_ServiceError(t *testing.T) {

	serviceErr := errors.New("update failed")

	name := "Updated Keyboard"

	fakeService := &productServiceStub{
		updateProductFn: func(ctx context.Context, id uint32, updates map[string]any) (core.ProductItem, error) {
			return core.ProductItem{}, serviceErr
		},
	}

	server := NewProductServer(fakeService)

	req := &productpb.UpdateProductByIDRequest{
		Id:   42,
		Name: &name,
	}

	got, err := server.UpdateProductByID(context.Background(), req)

	if !errors.Is(err, serviceErr) {
		t.Fatalf("expected service error, got %v", err)
	}

	if got != nil {
		t.Errorf("expected nil response, got %v", got)
	}
}
