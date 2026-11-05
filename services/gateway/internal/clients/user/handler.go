package user

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"
	userpb "tekora/pkg/gen/user"
	"tekora/services/gateway/internal/httpjson"
	"tekora/services/gateway/internal/myjwt"
	"tekora/services/gateway/internal/mymiddlewares"
	"time"

	"github.com/go-chi/chi/v5"
)

const userHandlerTimeout = 5 * time.Second

type UserClient struct {
	client userpb.UserServiceClient
}

func NewUserClient(client userpb.UserServiceClient) *UserClient {
	return &UserClient{client: client}
}

func (h *UserClient) InitUserHandlers(r chi.Router) {
	r.Route("/users", func(r chi.Router) {
		r.Post("/", h.RegisterUser)
		r.Post("/refresh", h.RefreshToken)

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
	if !httpjson.MyJSONDecode(w, r, &req) {
		httpjson.WriteJSONError(w, http.StatusBadRequest, "register_user_bad_request", "invalid request")
		return
	}

	userProto := &userpb.RegisterUserRequest{
		PhoneNumber: req.PhoneNumber,
		FullName:    req.FullName,
	}

	ctx, cancel := context.WithTimeout(r.Context(), userHandlerTimeout)
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

	accessToken, refreshToken, err := myjwt.CreateTokens(
		strconv.FormatUint(uint64(registeredUser.GetId()), 10),
	)
	if err != nil {
		slog.Error("failed to create access/refresh token", "error", err)

		httpjson.WriteJSONError(
			w,
			http.StatusInternalServerError,
			"token_creation_failed",
			"failed to create token",
		)
		return
	}

	res := RegisterUserResponse{
		ID:           registeredUser.GetId(),
		PhoneNumber:  registeredUser.GetPhoneNumber(),
		FullName:     registeredUser.GetFullName(),
		CreatedAt:    registeredUser.GetCreatedAt().AsTime(),
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
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

	userID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		slog.Error("failed to parse id to int", "id_str", idStr, "error", err)

		httpjson.WriteJSONError(w, http.StatusBadRequest, "get_user_bad_request", "invalid request")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), userHandlerTimeout)
	defer cancel()

	userRes, err := h.client.GetUserByID(ctx, &userpb.GetUserByIDRequest{
		Id: uint32(userID),
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

func (h *UserClient) UpdateUserByID(
	w http.ResponseWriter,
	r *http.Request,
) {
	idStr, ok := mymiddlewares.UserIDFromContext(r.Context())
	if !ok {
		slog.Error("failed to get id from context", "id_str", idStr, "ok", ok)

		httpjson.WriteJSONError(w, http.StatusInternalServerError, "get_user_bad_request", "internal server error")
		return
	}

	userID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		slog.Error("failed to parse id to int", "id_str", idStr, "error", err)

		httpjson.WriteJSONError(w, http.StatusBadRequest, "get_user_bad_request", "invalid request")
		return
	}

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

	userProto := &userpb.UpdateUserByIDRequest{
		Id:          uint32(userID),
		PhoneNumber: req.PhoneNumber,
		FullName:    req.FullName,
		Address:     req.Address,
	}

	ctx, cancel := context.WithTimeout(r.Context(), userHandlerTimeout)
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
	idStr, ok := mymiddlewares.UserIDFromContext(r.Context())
	if !ok {
		slog.Error("failed to get id from context", "id_str", idStr, "ok", ok)

		httpjson.WriteJSONError(w, http.StatusInternalServerError, "get_user_bad_request", "internal server error")
		return
	}

	userID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		slog.Error("failed to parse id to int", "id_str", idStr, "error", err)

		httpjson.WriteJSONError(w, http.StatusBadRequest, "get_user_bad_request", "invalid request")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), userHandlerTimeout)
	defer cancel()

	_, err = h.client.DeleteUser(ctx, &userpb.DeleteUserRequest{
		Id: uint32(userID),
	})
	if err != nil {
		slog.Error("failed to delete user", "error", err)

		httpjson.WriteJSONError(w, http.StatusBadGateway, "register_user_service_unavailable", "failed to delete user")
		return
	}

	if err := httpjson.MyJSONResponse(w, http.StatusNoContent, nil); err != nil {
		slog.Error("failed to encode response", "error", err)
	}
}

func (h *UserClient) RefreshToken(
	w http.ResponseWriter,
	r *http.Request,
) {
	var req RefreshTokenRequest

	if !httpjson.MyJSONDecode(w, r, &req) {
		return
	}

	claims, err := myjwt.VerifyRefreshToken(req.RefreshToken)
	if err != nil {
		slog.Warn("failed to verify refresh token", "error", err)

		httpjson.WriteJSONError(
			w,
			http.StatusUnauthorized,
			"invalid_refresh_token",
			"invalid or expired refresh token",
		)
		return
	}

	accessToken, err := myjwt.CreateAccessToken(claims.Subject)
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

	res := RefreshTokenResponse{
		AccessToken: accessToken,
	}

	if err := httpjson.MyJSONResponse(
		w,
		http.StatusOK,
		res,
	); err != nil {
		slog.Error("failed to encode response", "error", err)
	}
}
