package user

import "time"

type RegisterUserRequest struct {
	PhoneNumber string `json:"phone_number" validate:"required,phone"`
	FullName    string `json:"full_name" validate:"required,min=8"`
}

type RegisterUserResponse struct {
	ID          uint32    `json:"id"`
	PhoneNumber string    `json:"phone_number"`
	FullName    string    `json:"full_name"`
	CreatedAt   time.Time `json:"created_at"`
	AccessToken string    `json:"access_token"`
}

type GetUserByIDResponse struct {
	ID          uint32    `json:"id"`
	PhoneNumber string    `json:"phone_number"`
	FullName    string    `json:"full_name"`
	Address     string    `json:"address"`
	CreatedAt   time.Time `json:"created_at"`
}

type UpdateUserByIDRequest struct {
	PhoneNumber *string `json:"phone_number"`
	FullName    *string `json:"full_name"`
	Address     *string `json:"address"`
}

type UpdateUserByIDResponse struct {
	ID          uint32    `json:"id"`
	PhoneNumber string    `json:"phone_number"`
	FullName    string    `json:"full_name"`
	Address     string    `json:"address"`
	CreatedAt   time.Time `json:"created_at"`
}
