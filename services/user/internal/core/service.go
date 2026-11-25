package core

import (
	"context"
	"fmt"
)

type userRepository interface {
	CreateUser(context.Context, CreateUserPayload) (User, error)
	GetUserByID(context.Context, int64) (User, error)
	UpdateUserByID(context.Context, int64, map[string]any) (User, error)
	DeleteUser(context.Context, int64) (bool, error)
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

func (s *userService) GetUserByID(ctx context.Context, id int64) (User, error) {
	user, err := s.repo.GetUserByID(ctx, id)
	if err != nil {
		return User{}, fmt.Errorf("get user by id: %w", err)
	}

	return user, nil
}

func (s *userService) UpdateUserByID(ctx context.Context, id int64, updates map[string]any) (User, error) {
	return s.repo.UpdateUserByID(ctx, id, updates)
}

func (s *userService) DeleteUser(ctx context.Context, id int64) (bool, error) {
	return s.repo.DeleteUser(ctx, id)
}
