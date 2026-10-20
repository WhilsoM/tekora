package core

import "errors"

var (
	ErrPhoneAlreadyExists = errors.New("phone already exists")
	ErrInvalidUserData    = errors.New("invalid user data")
	ErrUserNotFound       = errors.New("user not found")
)
