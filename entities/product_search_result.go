package entities

import (
	"time"

	"github.com/google/uuid"
)

type ProductSearchResult struct {
	ID            uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	UserID        uuid.UUID
	ProductID     uuid.UUID
	Price         int
	OriginalPrice int
	Discount      string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func (p *ProductSearchResult) IsPriceValid() bool {
	return p.Price > 0
}

type Tabler interface {
	TableName() string
}

func (ProductSearchResult) TableName() string {
	return "product_search_history"
}
