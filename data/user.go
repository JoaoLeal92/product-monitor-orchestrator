package data

import (
	"github.com/JoaoLeal92/product-monitor-orchestrator/entities"
	"gorm.io/gorm"
)

// UserRepo repository struct
type UserRepo struct {
	db *gorm.DB
}

// NewUsersRepository instantiates a new user repository
func NewUsersRepository(conn *gorm.DB) *UserRepo {
	return &UserRepo{
		db: conn,
	}
}

func (r *UserRepo) GetActiveUsers() ([]entities.User, error) {
	var users []entities.User

	result := r.db.Where("active = ?", 1).Find(&users)
	if result.Error != nil {
		return []entities.User{}, result.Error
	}

	return users, nil
}
