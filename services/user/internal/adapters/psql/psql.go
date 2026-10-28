package psql

import (
	"context"
	"tekora/services/user/internal/core"

	"gorm.io/gorm"
)

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *userRepository {
	return &userRepository{
		db: db,
	}
}

func (repo *userRepository) GetUserByPhone(ctx context.Context, payload string) {

}

func (repo *userRepository) CreateUser(ctx context.Context, payload core.CreateUserPayload) (core.User, error) {
	user := User{
		PhoneNumber: payload.PhoneNumber,
		FullName:    payload.FullName,
	}

	if err := repo.db.WithContext(ctx).Create(&user).Error; err != nil {
		return core.User{}, err
	}

	return core.User{
		ID:          user.ID,
		PhoneNumber: user.PhoneNumber,
		FullName:    user.FullName,
		CreatedAt:   user.CreatedAt,
	}, nil
}

func (repo *userRepository) GetUserByID(ctx context.Context, id uint32) (core.User, error) {
	return core.User{}, nil
}

func (repo *userRepository) UpdateUserByID(ctx context.Context) (core.User, error) {
	return core.User{}, nil
}

func (repo *userRepository) DeleteUser(ctx context.Context) (core.User, error) {
	return core.User{}, nil
}
