package psql

import (
	"context"
	"errors"
	"fmt"
	"tekora/services/user/internal/core"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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
	var user User

	err := repo.db.
		WithContext(ctx).
		Where("id = ?", id).
		Take(&user).
		Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return core.User{}, core.ErrUserNotFound
	}

	if err != nil {
		return core.User{}, fmt.Errorf("get user by id: %w", err)
	}

	return core.User{
		ID:          user.ID,
		PhoneNumber: user.PhoneNumber,
		FullName:    user.FullName,
		CreatedAt:   user.CreatedAt,
	}, nil
}

func (repo *userRepository) UpdateUserByID(ctx context.Context, id uint32, updates map[string]interface{}) (core.User, error) {
	var user User

	result := repo.db.WithContext(ctx).
		Model(&user).
		Clauses(clause.Returning{}).
		Where("id = ?", id).
		Updates(updates)

	if result.Error != nil {
		return core.User{}, result.Error
	}

	if result.RowsAffected == 0 {
		return core.User{}, gorm.ErrRecordNotFound
	}

	return core.User{
		ID:          user.ID,
		FullName:    user.FullName,
		PhoneNumber: user.PhoneNumber,
		Address:     user.Address,
		CreatedAt:   user.CreatedAt,
	}, nil
}

func (repo *userRepository) DeleteUser(ctx context.Context) (core.User, error) {
	return core.User{}, nil
}
