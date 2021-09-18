package contracts

import (
	"github.com/google/uuid"

	"github.com/JoaoLeal92/product-monitor-orchestrator/entities"
)

type RepoManager interface {
	Users()
	Products()
}

type UsersRepository interface {
	GetUsers() ([]entities.User, error)
}

type ProductsRepository interface {
	GetActiveProductsByUser(userID uuid.UUID) ([]entities.Product, error)
	GetProductByID(productId uuid.UUID) (entities.Product, error)
}

type CrawlersRepository interface {
	GetCrawlerName(crawlerID uuid.UUID) (string, error)
}

type ProductSearchHistoryRepository interface {
	InsertNewHistory(productSearch *entities.ProductSearchResult) error
	GetHistoryByProductID(productID uuid.UUID) ([]entities.ProductSearchResult, error)
}
