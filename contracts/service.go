package contracts

import (
	"github.com/JoaoLeal92/product-monitor-orchestrator/entities"
)

type ProductNotificationService interface {
	Execute(product *entities.Product, productSearchResult *entities.ProductSearchResult) error
}
