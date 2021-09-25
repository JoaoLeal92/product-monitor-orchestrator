package contracts

import (
	"github.com/JoaoLeal92/product-monitor-orchestrator/entities"
	"github.com/google/uuid"
)

type ProductSearchHistoryService interface {
	CreateProductSearchHistory(crawlerResult *entities.CrawlerResult, userID string, productID uuid.UUID) (entities.ProductSearchResult, error)
}
