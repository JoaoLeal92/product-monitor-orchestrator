package services

import (
	"errors"
	"fmt"
	"math"

	"github.com/JoaoLeal92/product-monitor-orchestrator/contracts"
	"github.com/JoaoLeal92/product-monitor-orchestrator/entities"
)

type ProductNotificationService struct {
	db           contracts.RepoManager
	logger       contracts.LoggerContract
	queueManager contracts.QueueManager
}

type averageProductData struct {
	avgPrice    float64
	avgDiscount string
}

func NewProductNotificationService(db contracts.RepoManager, logger contracts.LoggerContract, queueManager contracts.QueueManager) *ProductNotificationService {
	return &ProductNotificationService{
		db:           db,
		logger:       logger,
		queueManager: queueManager,
	}
}

func (p *ProductNotificationService) Execute(product *entities.Product, productSearchResult *entities.ProductSearchResult) error {
	if !productSearchResult.IsPriceValid() {
		p.logger.Info("Preço inválido")
		return errors.New("invalid price result (<0)")
	}

	if err := p.db.ProductSearchHistory().InsertNewHistory(productSearchResult); err != nil {
		return err
	}

	if !product.IsBelowMaxPrice(productSearchResult.Price) {
		p.logger.Info("Preço acima do desejado")
		return nil
	}

	productSearchHistory, err := p.db.ProductSearchHistory().GetHistoryByProductID(product.ID)
	if err != nil {
		return err
	}
	avgData := p.getAverageProductData(productSearchHistory, productSearchResult.Price)
	queuePayload := p.formatQueuePayload(*product, *productSearchResult, avgData)
	p.queueManager.SendMessage(queuePayload)

	return nil
}

func (p *ProductNotificationService) getAverageProductData(productHistory []entities.ProductSearchResult, currentPrice int) averageProductData {
	avgPrice := p.getProductAvgPrice(productHistory, currentPrice)
	avgDiscount := p.getAvgDiscount(avgPrice, currentPrice)

	return averageProductData{
		avgPrice:    avgPrice,
		avgDiscount: avgDiscount,
	}
}

func (p *ProductNotificationService) getProductAvgPrice(productHistory []entities.ProductSearchResult, currentPrice int) float64 {
	var sum float64 = 0
	for _, product := range productHistory {
		sum += float64(product.Price) / 100
	}
	sum += float64(currentPrice)

	avg := sum / float64(len(productHistory)+1)
	roundAvg := math.Round(avg*100) / 100

	return roundAvg
}

func (p *ProductNotificationService) getAvgDiscount(avgPrice float64, currentPrice int) string {
	currentPriceFloat := float64(currentPrice) / 100

	priceDiff := avgPrice - currentPriceFloat
	discount := priceDiff / avgPrice

	discountString := fmt.Sprintf("%.2f", discount)

	return discountString
}

func (p *ProductNotificationService) formatQueuePayload(product entities.Product, productSearchResult entities.ProductSearchResult, avgProductData averageProductData) entities.ProductNotification {
	return entities.ProductNotification{
		Description: product.Description,
		Price:       float64(productSearchResult.Price),
		AvgPrice:    avgProductData.avgPrice,
		Discount:    productSearchResult.Discount,
		AvgDiscount: avgProductData.avgDiscount,
		Link:        product.Link,
		UserID:      product.UserID.String(),
	}
}
