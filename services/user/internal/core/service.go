package core

import "context"

type userRepository interface {
	CreateUser(ctx context.Context, payload CreateUserPayload) (User, error)
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
