package core

import "log/slog"

type userService struct {
}

func NewUserService() *userService {
	return &userService{}
}

func (s *userService) CreateUser() {
	slog.Info("create user")
}
