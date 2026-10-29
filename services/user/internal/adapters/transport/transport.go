package transport

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	userpb "tekora/pkg/gen/user"
	"tekora/services/user/internal/core"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type userService interface {
	CreateUser(context.Context, core.CreateUserPayload) (core.User, error)
	GetUserByID(context.Context, uint32) (core.User, error)
	UpdateUserByID(context.Context, uint32, map[string]any) (core.User, error)
	DeleteUser(context.Context) (core.User, error)
}

type userServer struct {
	userpb.UnimplementedUserServiceServer

	svc userService
}

func NewUserServer(svc userService) *userServer {
	return &userServer{
		svc: svc,
	}
}

func (server *userServer) RegisterUser(ctx context.Context, req *userpb.RegisterUserRequest) (*userpb.RegisterUserResponse, error) {
	payload := core.CreateUserPayload{
		PhoneNumber: req.GetPhoneNumber(),
		FullName:    req.GetFullName(),
	}

	user, err := server.svc.CreateUser(ctx, payload)
	if err != nil {
		switch {
		case errors.Is(err, core.ErrPhoneAlreadyExists):
			return nil, status.Error(
				codes.AlreadyExists,
				"phone already exists",
			)

		case errors.Is(err, core.ErrInvalidUserData):
			return nil, status.Error(
				codes.InvalidArgument,
				"invalid user data",
			)

		default:
			return nil, status.Error(
				codes.Internal,
				"internal server error",
			)
		}
	}

	return &userpb.RegisterUserResponse{
		Id:          user.ID,
		PhoneNumber: user.PhoneNumber,
		FullName:    user.FullName,
		CreatedAt:   timestamppb.New(user.CreatedAt),
	}, nil
}

func (server *userServer) LoginUser(ctx context.Context, req *userpb.LoginUserRequest) (*userpb.LoginUserResponse, error) {
	payload := core.User{
		FullName:    req.GetFullName(),
		PhoneNumber: req.GetPhoneNumber(),
	}

	_ = payload

	return &userpb.LoginUserResponse{}, nil
}

func (server *userServer) GetUserByID(ctx context.Context, req *userpb.GetUserByIDRequest) (*userpb.GetUserByIDResponse, error) {
	user, err := server.svc.GetUserByID(ctx, req.GetId())
	if err != nil {
		slog.Warn("failed to get user by id", "error", err)
		return nil, err
	}

	return &userpb.GetUserByIDResponse{
		Id:          user.ID,
		PhoneNumber: user.PhoneNumber,
		FullName:    user.FullName,
		Address:     user.Address,
		CreatedAt:   timestamppb.New(user.CreatedAt),
	}, nil
}

func (server *userServer) UpdateUserByID(ctx context.Context, req *userpb.UpdateUserByIDRequest) (*userpb.UpdateUserByIDResponse, error) {
	fields := make(map[string]any, 3)

	if req.PhoneNumber != nil {
		fields["phone_number"] = *req.PhoneNumber
	}

	if req.FullName != nil {
		fields["full_name"] = *req.FullName
	}

	if req.Address != nil {
		fields["address"] = *req.Address
	}

	if len(fields) == 0 {
		return nil, status.Error(
			codes.InvalidArgument,
			"no fields to update",
		)
	}

	updatedUser, err := server.svc.UpdateUserByID(ctx, req.Id, fields)
	if err != nil {
		return nil, status.Error(
			codes.Internal,
			"internal server error",
		)
	}

	return &userpb.UpdateUserByIDResponse{
		Id:          updatedUser.ID,
		FullName:    updatedUser.FullName,
		PhoneNumber: updatedUser.PhoneNumber,
		Address:     updatedUser.Address,
		CreatedAt:   timestamppb.New(updatedUser.CreatedAt),
	}, nil
}

func (server *userServer) DeleteUser(ctx context.Context, req *userpb.DeleteUserRequest) (*userpb.DeleteUserResponse, error) {
	return &userpb.DeleteUserResponse{}, nil
}
