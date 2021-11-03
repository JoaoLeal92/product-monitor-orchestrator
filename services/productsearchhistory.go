package services

import (
	"fmt"
	"strconv"

	"github.com/JoaoLeal92/product-monitor-orchestrator/contracts"
	"github.com/JoaoLeal92/product-monitor-orchestrator/entities"
	"github.com/google/uuid"
)

type ProductSearchHistoryService struct {
	db contracts.RepoManager
}

func NewProductSearchHistoryService(db contracts.RepoManager) *ProductSearchHistoryService {
	return &ProductSearchHistoryService{
		db: db,
	}
}

func (s *ProductSearchHistoryService) CreateProductSearchHistory(crawlerResult *entities.CrawlerResult, userID string, productID uuid.UUID) (entities.ProductSearchResult, error) {
	fmt.Println("Inserindo histórico de pesquisa para o produto")

	priceInt, err := s.priceStringToInt(crawlerResult.Price)
	if err != nil {
		return entities.ProductSearchResult{}, err
	}

	originalPriceInt, err := s.priceStringToInt(crawlerResult.OriginalPrice)
	if err != nil {
		return entities.ProductSearchResult{}, err
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return entities.ProductSearchResult{}, err
	}

	productSearchResult := entities.ProductSearchResult{
		UserID:        userUUID,
		ProductID:     productID,
		Price:         priceInt,
		OriginalPrice: originalPriceInt,
		Discount:      crawlerResult.Discount,
	}

	if productSearchResult.Price != 0 {
		s.db.ProductSearchHistory().InsertNewHistory(&productSearchResult)
		fmt.Println("Histórico inserido com sucesso")
	}

	return productSearchResult, nil
}

func (s *ProductSearchHistoryService) priceStringToInt(priceString string) (int, error) {
	var priceInt int
	var err error

	if priceString != "" {
		priceInt, err = strconv.Atoi(priceString)
		if err != nil {
			return 0, err
		}
	}

	return priceInt, nil
}
