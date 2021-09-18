package main

import (
	"fmt"
	"math"

	"github.com/JoaoLeal92/product-monitor-orchestrator/data"
	"github.com/JoaoLeal92/product-monitor-orchestrator/entities"
)

type Validator struct {
	db *data.Connection
}

func NewValidator(db *data.Connection) *Validator {
	return &Validator{
		db: db,
	}
}

func (v *Validator) ValidateCrawlerResults(crawlerResults []entities.ProductSearchResult) error {
	fmt.Println("Validando resultado das buscas")
	for _, result := range crawlerResults {
		product, err := v.db.Products().GetProductByID(result.ProductID)
		if err != nil {
			return err
		}
		// Validade output
		if result.Price <= product.MaxPrice {
			fmt.Println("\nProduto dentro do preço desejado")
			// Calculate average price over the search period
			productHistory, err := v.db.ProductSearchHistory().GetHistoryByProductID(product.ID)
			if err != nil {
				panic(err)
			}

			avgProductPrice := v.getProductAvgPrice(productHistory)
			avgProductDiscount := v.getAvgDiscount(avgProductPrice, result.Price)

			// Informs the user (insert telegram bot logic here)
			fmt.Println("Current price: ", result.Price)
			fmt.Println("Average search price: ", avgProductPrice)
			fmt.Println("Announced discount: ", result.Discount)
			fmt.Println("Discount over average search price: ", avgProductDiscount)
		} else {
			fmt.Println("\nProduto acima do preço desejado: ")
			fmt.Printf("%+v", result)
		}
	}

	return nil
}

func (v *Validator) getProductAvgPrice(productHistory []entities.ProductSearchResult) float64 {
	var sum float64 = 0
	for _, product := range productHistory {
		sum += float64(product.Price) / 100
	}

	avg := sum / float64(len(productHistory))
	roundAvg := math.Round(avg*100) / 100

	return roundAvg
}

func (v *Validator) getAvgDiscount(avgPrice float64, currentPrice int) string {
	currentPriceFloat := float64(currentPrice) / 100

	priceDiff := avgPrice - currentPriceFloat
	discount := priceDiff / avgPrice

	discountString := fmt.Sprintf("%.2f", discount)

	return discountString
}
