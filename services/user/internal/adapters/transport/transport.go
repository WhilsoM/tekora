package transport

import (
	"context"
	userpb "tekora/pkg/gen/user"
	"tekora/services/user/internal/core"

	"google.golang.org/protobuf/types/known/timestamppb"
)

type userService interface {
	CreateUser(ctx context.Context, payload core.CreateUserPayload) (core.User, error)
}

type userServer struct {
	userpb.UnimplementedUserServiceServer

	service userService
}

func NewServer(service userService) *userServer {
	return &userServer{
		service: service,
	}
}

func (server *userServer) RegisterUser(ctx context.Context, req *userpb.RegisterUserRequest) (*userpb.RegisterUserResponse, error) {
	payload := core.CreateUserPayload{
		PhoneNumber: req.GetPhoneNumber(),
		FullName:    req.GetFullName(),
	}

	user, err := server.service.CreateUser(ctx, payload)
	if err != nil {
		return nil, err
	}

	return &userpb.RegisterUserResponse{
		Id:          user.ID,
		PhoneNumber: user.PhoneNumber,
		FullName:    user.FullName,
		CreatedAt:   timestamppb.New(user.CreatedAt),
	}, nil
}
