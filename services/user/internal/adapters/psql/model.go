package psql

import "time"

type User struct {
	ID          int64 `gorm:"primaryKey"`
	PhoneNumber string
	FullName    string
	Address     string
	CreatedAt   time.Time
}
