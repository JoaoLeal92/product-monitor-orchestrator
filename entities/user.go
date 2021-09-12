package entities

import (
	"time"

	"github.com/google/uuid"
)

// User data
type User struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;default:uuid_generate_v4()"`
	Name        string    `json:"name"`
	Email       string    `json:"email" gorm:"unique"`
	Password    string    `json:"password"`
	PhoneNumber int64     `json:"phone_number"`
	Products    []Product
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
