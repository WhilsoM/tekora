package memory

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"tekora/services/user/internal/core"
	"time"
)

type memoryRepository struct {
	mu     sync.RWMutex
	nextID uint32
	users  map[int]core.User
}

func NewMemoryRepository() *memoryRepository {
	return &memoryRepository{
		nextID: 0,
		users:  make(map[int]core.User),
	}
}

func (db *memoryRepository) CreateUser(ctx context.Context, payload core.CreateUserPayload) (core.User, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	for _, k := range db.users {
		if k.PhoneNumber == payload.PhoneNumber {
			return core.User{}, core.ErrPhoneAlreadyExists
		}
	}

	db.nextID++

	fmt.Println("next ID", db.nextID)

	user := core.User{
		ID:          db.nextID,
		PhoneNumber: payload.PhoneNumber,
		FullName:    payload.FullName,
		CreatedAt:   time.Now().UTC(),
	}

	db.users[int(user.ID)] = user

	fmt.Println("created a new user", db.users[int(db.nextID)])

	return user, nil
}

func (db *memoryRepository) GetUserByID(ctx context.Context, id uint32) (core.User, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	user, ok := db.users[int(id)]
	if !ok {
		return core.User{}, errors.New("user not found")
	}

	return user, nil
}

func (db *memoryRepository) UpdateUserByID(ctx context.Context) (core.User, error) {
	return core.User{}, nil
}

func (db *memoryRepository) DeleteUser(ctx context.Context) (core.User, error) {
	return core.User{}, nil
}
