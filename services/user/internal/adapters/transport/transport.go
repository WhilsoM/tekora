package transport

import (
	"context"
	userpb "tekora/pkg/gen/user"
)

type userService interface {
	CreateUser()
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
	return &userpb.RegisterUserResponse{}, nil
}
