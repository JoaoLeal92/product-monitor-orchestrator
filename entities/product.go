package entities

import (
	"errors"
	"path/filepath"
	"strings"

	"github.com/JoaoLeal92/product-monitor-orchestrator/config"
	"github.com/google/uuid"
	"github.com/oleiade/reflections"
)

type Product struct {
	ID          uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	UserID      uuid.UUID `gorm:"index"`
	Description string
	MaxPrice    int
	Link        string
	CrawlerName string
}

func (p *Product) IsBelowMaxPrice(price int) bool {
	return price != 0 && price <= p.MaxPrice
}

func (p *Product) GetCrawlerPath(cfg *config.CrawlerConfig) (string, error) {
	crawlerKey := p.getFormatedCrawlerKey()
	crawlerPath, err := reflections.GetField(cfg, crawlerKey)
	if err != nil {
		return "", err
	}

	crawlerPathStr, ok := crawlerPath.(string)
	if !ok {
		return "", errors.New("invalid crawler path")
	}

	crawlerAbsPath, err := filepath.Abs(crawlerPathStr)
	if err != nil {
		return "", err
	}

	return crawlerAbsPath, nil
}

func (p *Product) getFormatedCrawlerKey() string {
	crawlerKey := strings.Title(p.CrawlerName)
	crawlerKey = strings.Replace(crawlerKey, "-", "", 1)

	return crawlerKey
}
