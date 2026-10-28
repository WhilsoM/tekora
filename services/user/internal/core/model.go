package core

import "time"

type User struct {
	ID          uint32
	FullName    string
	PhoneNumber string
	Address     string
	CreatedAt   time.Time
}

type CreateUserPayload struct {
	FullName    string
	PhoneNumber string
}

type LoginUserPayload struct {
	FullName    string
	PhoneNumber string
}
