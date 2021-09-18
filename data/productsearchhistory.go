package data

import (
	"errors"

	"github.com/JoaoLeal92/product-monitor-orchestrator/entities"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ProductSearchHisotry repository struct
type ProductSearchHisotry struct {
	db *gorm.DB
}

// NewProductSearchHistoryRepository instantiates a new user repository
func NewProductSearchHistoryRepository(conn *gorm.DB) *ProductSearchHisotry {
	return &ProductSearchHisotry{
		db: conn,
	}
}

func (r *ProductSearchHisotry) InsertNewHistory(productSearch *entities.ProductSearchResult) error {
	result := r.db.Create(&productSearch)

	if result.Error != nil {
		return errors.New(result.Error.Error())
	}
	return nil
}

func (r *ProductSearchHisotry) GetHistoryByProductID(productID uuid.UUID) ([]entities.ProductSearchResult, error) {
	var searchHistory []entities.ProductSearchResult

	result := r.db.Where("product_id = ?", productID).Find(&searchHistory)
	if result.Error != nil {
		return []entities.ProductSearchResult{}, result.Error
	}

	return searchHistory, nil
}
