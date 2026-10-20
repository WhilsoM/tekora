package core

import (
	"context"
	"fmt"
)

type userRepository interface {
	CreateUser(ctx context.Context, payload CreateUserPayload) (User, error)
	GetUserByID(ctx context.Context, id uint32) (User, error)
	UpdateUserByID(ctx context.Context) (User, error)
	DeleteUser(ctx context.Context) (User, error)
}

type userService struct {
	repo userRepository
}

func NewUserService(repo userRepository) *userService {
	return &userService{
		repo: repo,
	}
}

func (s *userService) CreateUser(ctx context.Context, payload CreateUserPayload) (User, error) {
	return s.repo.CreateUser(ctx, payload)
}

func (s *userService) GetUserByID(ctx context.Context, id uint32) (User, error) {
	user, err := s.repo.GetUserByID(ctx, id)
	if err != nil {
		return User{}, fmt.Errorf("get user by id: %w", err)
	}

	return user, nil
}

func (s *userService) UpdateUserByID(ctx context.Context) (User, error) {
	return User{}, nil
}

func (s *userService) DeleteUser(ctx context.Context) (User, error) {
	return User{}, nil
}
