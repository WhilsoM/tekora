package psql

import "time"

type User struct {
	ID          uint32 `gorm:"primaryKey"`
	PhoneNumber string
	FullName    string
	Address     string
	CreatedAt   time.Time
}
