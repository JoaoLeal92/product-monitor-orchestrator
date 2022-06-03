package contracts

import (
	"github.com/google/uuid"

	"github.com/JoaoLeal92/product-monitor-orchestrator/entities"
)

type RepoManager interface {
	Products() ProductsRepository
	ProductSearchHistory() ProductSearchHistoryRepository
}

type ProductsRepository interface {
	GetProductsListForCrawler() ([]entities.Product, error)
}

type ProductSearchHistoryRepository interface {
	InsertNewHistory(productSearch *entities.ProductSearchResult) error
	GetHistoryByProductID(productID uuid.UUID) ([]entities.ProductSearchResult, error)
}
