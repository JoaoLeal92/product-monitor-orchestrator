package entities

import (
	"time"

	"github.com/google/uuid"
)

// Crawler data
type Crawler struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CrawlerResult struct {
	Price         string `mapstructure:"price"`
	OriginalPrice string `mapstructure:"originalPrice"`
	Discount      string `mapstructure:"discount"`
	Link          string `mapstructure:"link"`
}

type CrawlerOutput struct {
	OutputMessage string
	OutputError   string
}

type CrawlerChanRestult struct {
	CrawlerResult      string
	CrawlerErr         string
	ProductID          uuid.UUID
	ProductDescription string
}
