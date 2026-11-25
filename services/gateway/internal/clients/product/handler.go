package product

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	productpb "tekora/pkg/gen/product"
	"tekora/services/gateway/internal/httpjson"

	"github.com/go-chi/chi/v5"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const productHandlerTimeout = 5 * time.Second

type ProductClient struct {
	client productpb.ProductServiceClient
}

func NewProductClient(client productpb.ProductServiceClient) *ProductClient {
	return &ProductClient{
		client: client,
	}
}

func (h *ProductClient) InitProductHandlers(r chi.Router) {
	r.Route("/products", func(r chi.Router) {
		r.Get("/", h.GetProducts)
		r.Get("/{id}", h.GetProductByID)
		r.Post("/", h.CreateProduct)
		r.Patch("/{id}", h.UpdateProductByID)
		r.Delete("/{id}", h.DeleteProductByID)
	})
}

func (h *ProductClient) handleGRPCError(
	w http.ResponseWriter,
	operation string,
	err error,
) {
	slog.Error("product service request failed",
		"operation", operation,
		"error", err,
	)

	httpStatus := http.StatusBadGateway
	errorCode := "product_service_unavailable"
	message := "product service is unavailable"

	switch status.Code(err) {
	case codes.InvalidArgument:
		httpStatus = http.StatusBadRequest
		errorCode = "invalid_argument"
		message = "invalid request"

	case codes.NotFound:
		httpStatus = http.StatusNotFound
		errorCode = "product_not_found"
		message = "product not found"

	case codes.AlreadyExists:
		httpStatus = http.StatusConflict
		errorCode = "product_already_exists"
		message = "product already exists"

	case codes.Unauthenticated:
		httpStatus = http.StatusUnauthorized
		errorCode = "unauthenticated"
		message = "authentication required"

	case codes.PermissionDenied:
		httpStatus = http.StatusForbidden
		errorCode = "permission_denied"
		message = "permission denied"

	case codes.DeadlineExceeded:
		httpStatus = http.StatusGatewayTimeout
		errorCode = "request_timeout"
		message = "product service request timeout"
	}

	httpjson.WriteJSONError(w, httpStatus, errorCode, message)
}

func parseProductID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)

	if err != nil || id == 0 {
		httpjson.WriteJSONError(
			w,
			http.StatusBadRequest,
			"invalid_product_id",
			"invalid product ID",
		)
		return 0, false
	}

	return id, true
}

func toProductItem(item *productpb.ProductItem) ProductItem {
	var createdAt, updatedAt time.Time

	if item.GetCreatedAt() != nil {
		createdAt = item.GetCreatedAt().AsTime()
	}

	if item.GetUpdatedAt() != nil {
		updatedAt = item.GetUpdatedAt().AsTime()
	}

	return ProductItem{
		ID:        item.GetId(),
		Name:      item.GetName(),
		Price:     item.GetPrice(),
		Category:  item.GetCategory(),
		Stock:     item.GetStock(),
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}
}

func (h *ProductClient) GetProducts(
	w http.ResponseWriter,
	r *http.Request,
) {
	req := GetProductsRequest{
		Limit:  10,
		Offset: 0,
	}

	if value := r.URL.Query().Get("limit"); value != "" {
		parsed, err := strconv.ParseUint(value, 10, 32)
		if err != nil || parsed == 0 || parsed > 100 {
			httpjson.WriteJSONError(
				w,
				http.StatusBadRequest,
				"invalid_limit",
				"limit must be between 1 and 100",
			)
			return
		}
		req.Limit = uint32(parsed)
	}

	if value := r.URL.Query().Get("offset"); value != "" {
		parsed, err := strconv.ParseUint(value, 10, 32)
		if err != nil {
			httpjson.WriteJSONError(
				w,
				http.StatusBadRequest,
				"invalid_offset",
				"invalid offset",
			)
			return
		}
		req.Offset = uint32(parsed)
	}

	ctx, cancel := context.WithTimeout(r.Context(), productHandlerTimeout)
	defer cancel()

	response, err := h.client.GetProducts(ctx, &productpb.GetProductsRequest{
		Limit:  req.Limit,
		Offset: req.Offset,
	})

	if err != nil {
		h.handleGRPCError(w, "get products", err)
		return
	}

	if response == nil {
		httpjson.WriteJSONError(
			w,
			http.StatusBadGateway,
			"invalid_service_response",
			"invalid response from product service",
		)
		return
	}

	items := make([]ProductItemShort, 0, len(response.GetItems()))

	for _, item := range response.GetItems() {
		if item == nil {
			continue
		}

		items = append(items, ProductItemShort{
			ID:    item.GetId(),
			Name:  item.GetName(),
			Price: item.GetPrice(),
		})
	}

	res := GetProductsResponse{
		Items: items,
	}

	if err := httpjson.MyJSONResponse(w, http.StatusOK, res); err != nil {
		slog.Error("failed to encode products response", "error", err)
	}
}

func (h *ProductClient) GetProductByID(
	w http.ResponseWriter,
	r *http.Request,
) {
	id, ok := parseProductID(w, r)
	if !ok {
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), productHandlerTimeout)
	defer cancel()

	response, err := h.client.GetProductByID(ctx, &productpb.GetProductByIDRequest{
		Id: id,
	})

	if err != nil {
		h.handleGRPCError(w, "get product by id", err)
		return
	}

	if response == nil || response.GetItem() == nil {
		httpjson.WriteJSONError(
			w,
			http.StatusBadGateway,
			"invalid_service_response",
			"invalid response from product service",
		)
		return
	}

	res := GetProductByIDResponse{
		Item: toProductItem(response.GetItem()),
	}

	if err := httpjson.MyJSONResponse(w, http.StatusOK, res); err != nil {
		slog.Error("failed to encode product response", "error", err)
	}
}

func (h *ProductClient) DeleteProductByID(
	w http.ResponseWriter,
	r *http.Request,
) {
	id, ok := parseProductID(w, r)
	if !ok {
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), productHandlerTimeout)
	defer cancel()

	response, err := h.client.DeleteProductByID(ctx, &productpb.DeleteProductByIDRequest{
		Id: id,
	})

	if err != nil {
		h.handleGRPCError(w, "delete product by id", err)
		return
	}

	if response == nil {
		httpjson.WriteJSONError(
			w,
			http.StatusBadGateway,
			"invalid_service_response",
			"invalid response from product service",
		)
		return
	}

	res := DeleteProductByIDResponse{
		IsSuccess: response.GetIsSuccess(),
		Message:   response.GetMessage(),
	}

	if err := httpjson.MyJSONResponse(w, http.StatusOK, res); err != nil {
		slog.Error("failed to encode delete response", "error", err)
	}
}

func (h *ProductClient) CreateProduct(
	w http.ResponseWriter,
	r *http.Request,
) {
	var req CreateProductRequest

	if !httpjson.MyJSONDecode(w, r, &req) {
		httpjson.WriteJSONError(
			w,
			http.StatusBadRequest,
			"invalid_request",
			"invalid request body",
		)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), productHandlerTimeout)
	defer cancel()

	response, err := h.client.CreateProduct(ctx, &productpb.CreateProductRequest{
		Name:     req.Name,
		Price:    req.Price,
		Category: req.Category,
		Stock:    req.Stock,
	})

	if err != nil {
		h.handleGRPCError(w, "create product", err)
		return
	}

	if response == nil || response.GetItem() == nil {
		httpjson.WriteJSONError(
			w,
			http.StatusBadGateway,
			"invalid_service_response",
			"invalid response from product service",
		)
		return
	}

	res := CreateProductResponse{
		Item: toProductItem(response.GetItem()),
	}

	if err := httpjson.MyJSONResponse(w, http.StatusCreated, res); err != nil {
		slog.Error("failed to encode create product response", "error", err)
	}
}

func (h *ProductClient) UpdateProductByID(
	w http.ResponseWriter,
	r *http.Request,
) {
	id, ok := parseProductID(w, r)
	if !ok {
		return
	}

	var req UpdateProductByIDRequest

	if !httpjson.MyJSONDecode(w, r, &req) {
		httpjson.WriteJSONError(
			w,
			http.StatusBadRequest,
			"invalid_request",
			"invalid request body",
		)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), productHandlerTimeout)
	defer cancel()

	response, err := h.client.UpdateProductByID(ctx, &productpb.UpdateProductByIDRequest{
		Id:       id,
		Name:     req.Name,
		Price:    req.Price,
		Category: req.Category,
		Stock:    req.Stock,
	})

	if err != nil {
		h.handleGRPCError(w, "update product by id", err)
		return
	}

	if response == nil || response.GetItem() == nil {
		httpjson.WriteJSONError(
			w,
			http.StatusBadGateway,
			"invalid_service_response",
			"invalid response from product service",
		)
		return
	}

	res := UpdateProductByIDResponse{
		Item: toProductItem(response.GetItem()),
	}

	if err := httpjson.MyJSONResponse(w, http.StatusOK, res); err != nil {
		slog.Error("failed to encode update product response", "error", err)
	}
}
