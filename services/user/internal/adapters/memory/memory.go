package memory

import (
	"context"
	"fmt"
	"tekora/services/user/internal/core"
	"time"
)

type memoryRepository struct {
	nextID int
	users  map[int]core.User
}

func NewMemoryRepository() *memoryRepository {
	return &memoryRepository{
		nextID: 0,
		users:  make(map[int]core.User),
	}
}

func (db *memoryRepository) CreateUser(ctx context.Context, payload core.CreateUserPayload) (core.User, error) {
	db.nextID++

	fmt.Println("next ID", db.nextID)

	user := core.User{
		ID:          uint32(db.nextID),
		PhoneNumber: payload.PhoneNumber,
		FullName:    payload.FullName,
		CreatedAt:   time.Now().UTC(),
	}

	db.users[int(user.ID)] = user

	fmt.Println("created a new user", db.users[db.nextID])

	return user, nil
}
