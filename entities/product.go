package entities

import (
	"time"

	"github.com/google/uuid"
)

// Product product entity
type Product struct {
	ID          uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	UserID      uuid.UUID `gorm:"index"`
	Description string
	MaxPrice    int
	Active      bool `gorm:"default:true"`
	Link        string
	CrawlerID   uuid.UUID
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type CrawlerResult struct {
	Price         string `mapstructure:"price"`
	OriginalPrice string `mapstructure:"originalPrice"`
	Discount      string `mapstructure:"discount"`
	Link          string `mapstructure:"link"`
}

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

type Tabler interface {
	TableName() string
}

func (ProductSearchResult) TableName() string {
	return "product_search_history"
}
