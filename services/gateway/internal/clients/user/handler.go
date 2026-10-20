package user

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	userpb "tekora/pkg/gen/user"
	"tekora/services/gateway/internal/httpjson"
	"time"

	"github.com/go-chi/chi/v5"
)

type UserHandler struct {
	client userpb.UserServiceClient
}

func NewUserHandler(client userpb.UserServiceClient) *UserHandler {
	return &UserHandler{client: client}
}

func (h *UserHandler) InitUserHandlers(r chi.Router) {
	r.Route("/users", func(r chi.Router) {
		r.Post("/", h.RegisterUser)
		r.Get("/me", h.GetUserByID)

	})
}

func (h *UserHandler) RegisterUser(
	w http.ResponseWriter,
	r *http.Request,
) {
	var req RegisterUserRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		httpjson.WriteJSONError(w, http.StatusBadRequest, "register_user_bad_request", "invalid request")
		return
	}

	userProto := &userpb.RegisterUserRequest{
		PhoneNumber: req.PhoneNumber,
		FullName:    req.FullName,
	}

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	registeredUser, err := h.client.RegisterUser(ctx, userProto)
	if err != nil {
		slog.Error("failed to register user", "error", err)

		httpjson.WriteJSONError(w, http.StatusBadGateway, "register_user_service_unavailable", "service is unavailable")
		return
	}

	if registeredUser == nil || registeredUser.GetCreatedAt() == nil {
		httpjson.WriteJSONError(w, http.StatusBadGateway, "register_bad_gateway", "Invalid service response")
		return
	}

	res := RegisterUserResponse{
		ID:          registeredUser.GetId(),
		PhoneNumber: registeredUser.GetPhoneNumber(),
		FullName:    registeredUser.GetFullName(),
		CreatedAt:   registeredUser.GetCreatedAt().AsTime(),
	}

	if err := httpjson.MyJSONResponse(w, http.StatusCreated, res); err != nil {
		slog.Error("failed to encode response", "error", err)
	}
}

func (h *UserHandler) GetUserByID(
	w http.ResponseWriter,
	r *http.Request,
) {
	userRes, err := h.client.GetUserByID(r.Context(), &userpb.GetUserByIDRequest{
		Id: 1,
	})
	if err != nil {
		slog.Error("failed to get user", "error", err)

		httpjson.WriteJSONError(w, http.StatusBadRequest, "get_user_bad_request", "invalid request")
		return
	}

	res := GetUserByIDResponse{
		ID:          userRes.GetId(),
		PhoneNumber: userRes.GetPhoneNumber(),
		FullName:    userRes.GetFullName(),
		Address:     userRes.GetAddress(),
		CreatedAt:   userRes.GetCreatedAt().AsTime(),
	}

	if err := httpjson.MyJSONResponse(w, http.StatusOK, res); err != nil {
		slog.Error("failed to encode response", "error", err)
	}
}
