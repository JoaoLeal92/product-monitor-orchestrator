package data

import (
	"github.com/JoaoLeal92/product-monitor-orchestrator/entities"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ProductRepo repository struct
type ProductRepo struct {
	db *gorm.DB
}

// NewProductsRepository instantiates a new user repository
func NewProductsRepository(conn *gorm.DB) *ProductRepo {
	return &ProductRepo{
		db: conn,
	}
}

func (r *ProductRepo) GetActiveProductsByUser(userID uuid.UUID) ([]entities.Product, error) {
	var products []entities.Product

	result := r.db.Where("user_id = ? AND active = true", userID).Find(&products)
	if result.Error != nil {
		return []entities.Product{}, result.Error
	}

	return products, nil
}
