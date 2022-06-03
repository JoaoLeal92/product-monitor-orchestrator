package data

import (
	"github.com/JoaoLeal92/product-monitor-orchestrator/entities"
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

func (r *ProductRepo) GetProductsListForCrawler() ([]entities.Product, error) {
	var products []entities.Product
	query := `
		SELECT 
			u.id user_id,
			pr.id,
			pr.description,
			pr.max_price,
			pr.link,
			cr.name crawler_name
		FROM users u
		JOIN products pr
				ON u.id = pr.user_id
		JOIN crawlers cr
				ON pr.crawler_id = cr.id
		WHERE u.active = 1
		AND pr.active = true 
	`

	result := r.db.Raw(query).Scan(&products)
	if result.Error != nil {
		return []entities.Product{}, result.Error
	}

	return products, nil
}
