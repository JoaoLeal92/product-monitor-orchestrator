package data

import (
	"github.com/JoaoLeal92/product-monitor-orchestrator/entities"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CrawlerRepo repository struct
type CrawlerRepo struct {
	db *gorm.DB
}

// NewCrawlerRepository instantiates a new user repository
func NewCrawlerRepository(conn *gorm.DB) *CrawlerRepo {
	return &CrawlerRepo{
		db: conn,
	}
}

func (r *CrawlerRepo) GetCrawlerName(crawlerID uuid.UUID) (string, error) {
	var crawler entities.Crawler

	result := r.db.Where("id = ?", crawlerID).First(&crawler)
	if result.Error != nil {
		return "", result.Error
	}

	return crawler.Name, nil
}
