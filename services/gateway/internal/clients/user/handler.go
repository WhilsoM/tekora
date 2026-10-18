package user

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	userpb "tekora/pkg/gen/user"
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
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
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

		http.Error(
			w,
			"User service unavailable",
			http.StatusBadGateway,
		)
		return
	}

	if registeredUser == nil || registeredUser.GetCreatedAt() == nil {
		http.Error(w, "Invalid service response", http.StatusBadGateway)
		return
	}

	res := RegisterUserResponse{
		ID:          registeredUser.GetId(),
		PhoneNumber: registeredUser.GetPhoneNumber(),
		FullName:    registeredUser.GetFullName(),
		CreatedAt:   registeredUser.GetCreatedAt().AsTime(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(res); err != nil {
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

		http.Error(
			w,
			"User service unavailable",
			http.StatusBadGateway,
		)
		return
	}

	res := GetUserByIDResponse{
		ID:          userRes.GetId(),
		PhoneNumber: userRes.GetPhoneNumber(),
		FullName:    userRes.GetFullName(),
		Address:     userRes.GetAddress(),
		CreatedAt:   userRes.GetCreatedAt().AsTime(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(res); err != nil {
		slog.Error("failed to encode response", "error", err)
	}
}
