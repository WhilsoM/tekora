package user

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	userpb "tekora/pkg/gen/user"
	"tekora/services/gateway/internal/httpjson"
	"tekora/services/gateway/internal/myjwt"
	"tekora/services/gateway/internal/mymiddlewares"
	"time"

	"github.com/go-chi/chi/v5"
)

type UserClient struct {
	client userpb.UserServiceClient
}

func NewUserClient(client userpb.UserServiceClient) *UserClient {
	return &UserClient{client: client}
}

func (h *UserClient) InitUserHandlers(r chi.Router) {
	r.Route("/users", func(r chi.Router) {
		r.Post("/", h.RegisterUser)

		r.Group(func(r chi.Router) {
			r.Use(mymiddlewares.AuthMiddleware)

			r.Get("/me", h.GetUserByID)
			r.Patch("/me", h.UpdateUserByID)
			r.Delete("/me", h.DeleteUserByID)
		})
	})
}

func (h *UserClient) RegisterUser(
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

	accessToken, err := myjwt.CreateToken(
		strconv.FormatUint(uint64(registeredUser.GetId()), 10),
	)
	if err != nil {
		slog.Error("failed to create access token", "error", err)

		httpjson.WriteJSONError(
			w,
			http.StatusInternalServerError,
			"token_creation_failed",
			"failed to create access token",
		)
		return
	}

	res := RegisterUserResponse{
		ID:          registeredUser.GetId(),
		PhoneNumber: registeredUser.GetPhoneNumber(),
		FullName:    registeredUser.GetFullName(),
		CreatedAt:   registeredUser.GetCreatedAt().AsTime(),
		AccessToken: accessToken,
	}

	if err := httpjson.MyJSONResponse(w, http.StatusCreated, res); err != nil {
		slog.Error("failed to encode response", "error", err)
	}
}

func (h *UserClient) GetUserByID(
	w http.ResponseWriter,
	r *http.Request,
) {
	idStr, ok := mymiddlewares.UserIDFromContext(r.Context())
	if !ok {
		slog.Error("failed to get id from context", "id_str", idStr, "ok", ok)

		httpjson.WriteJSONError(w, http.StatusInternalServerError, "get_user_bad_request", "internal server error")
		return
	}
	slog.Info("id", idStr)

	fmt.Println("HELLO STDOUT")

	fmt.Fprintln(os.Stderr, "HELLO STDERR")

	slog.Info("HELLO SLOG")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		slog.Error("failed to convert id to number", "error", err)

		httpjson.WriteJSONError(w, http.StatusBadRequest, "get_user_bad_request", "invalid request")
		return
	}

	if id <= 0 || uint64(id) > uint64(^uint32(0)) {
		slog.Error("cannot convert id to uint", "error", err)

		httpjson.WriteJSONError(w, http.StatusBadRequest, "get_user_bad_request", "invalid request")
		return
	}

	userID := uint32(id)
	userRes, err := h.client.GetUserByID(r.Context(), &userpb.GetUserByIDRequest{
		Id: userID,
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

// TODO: Id gets from middleware, context.withValue
func (h *UserClient) UpdateUserByID(
	w http.ResponseWriter,
	r *http.Request,
) {
	var req UpdateUserByIDRequest

	if !httpjson.MyJSONDecode(w, r, &req) {
		return
	}

	if req.PhoneNumber == nil &&
		req.FullName == nil &&
		req.Address == nil {
		httpjson.WriteJSONError(
			w,
			http.StatusBadRequest,
			"empty_update",
			"at least one field must be provided",
		)
		return
	}
	// TODO: change id to actual
	userProto := &userpb.UpdateUserByIDRequest{
		Id:          1,
		PhoneNumber: req.PhoneNumber,
		FullName:    req.FullName,
		Address:     req.Address,
	}

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	updatedUser, err := h.client.UpdateUserByID(ctx, userProto)
	if err != nil {
		slog.Error("failed to register user", "error", err)

		httpjson.WriteJSONError(w, http.StatusBadGateway, "register_user_service_unavailable", "service is unavailable")
		return
	}

	res := UpdateUserByIDResponse{
		ID:          updatedUser.GetId(),
		PhoneNumber: updatedUser.GetPhoneNumber(),
		FullName:    updatedUser.GetFullName(),
		Address:     updatedUser.GetAddress(),
		CreatedAt:   updatedUser.GetCreatedAt().AsTime(),
	}

	if err := httpjson.MyJSONResponse(w, http.StatusOK, res); err != nil {
		slog.Error("failed to encode response", "error", err)
	}
}

func (h *UserClient) DeleteUserByID(
	w http.ResponseWriter,
	r *http.Request,
) {

	userProto := &userpb.DeleteUserRequest{
		Id: 1,
	}

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	_, err := h.client.DeleteUser(ctx, userProto)
	if err != nil {
		slog.Error("failed to delete user", "error", err)

		httpjson.WriteJSONError(w, http.StatusBadGateway, "register_user_service_unavailable", "failed to delete user")
		return
	}

	if err := httpjson.MyJSONResponse(w, http.StatusNoContent, nil); err != nil {
		slog.Error("failed to encode response", "error", err)
	}
}
