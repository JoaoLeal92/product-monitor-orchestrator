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
