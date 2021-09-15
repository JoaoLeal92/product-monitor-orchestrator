package main

import (
	"bytes"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"

	"github.com/dlclark/regexp2"
	"github.com/mitchellh/mapstructure"

	"github.com/JoaoLeal92/product-monitor-orchestrator/config"
	"github.com/JoaoLeal92/product-monitor-orchestrator/data"
	"github.com/JoaoLeal92/product-monitor-orchestrator/entities"
)

func main() {

	crawlerPathMap := map[string]string{
		"amazon": "../crawlers/amazon",
	}

	regexMap := map[string]string{
		"price":         `(?<=price=).*(?=,\s?original_price)`,
		"originalPrice": `(?<=original_price=).*(?=,\s?discount)`,
		"discount":      `(?<=discount=).*(?=,\s?link)`,
		"link":          `(?<=link=').*(?='\))`,
	}

	fmt.Println("Start orchestrator")
	cfg, err := config.ReadConfig()
	if err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}

	db, err := data.Instance(cfg.Db)

	users, err := db.Users().GetUsers()
	if err != nil {
		panic(err)
	}

	for _, user := range users {
		products, err := db.Products().GetActiveProductsByUser(user.ID)
		if err != nil {
			panic(err)
		}

		for _, product := range products {
			crawlerName, _ := db.Crawlers().GetCrawlerName(product.CrawlerID)
			if err != nil {
				panic(err)
			}

			crawlerPath, err := filepath.Abs(crawlerPathMap[crawlerName])
			if err != nil {
				panic(err)
			}

			pipfile, err := filepath.Abs(
				filepath.Join(
					crawlerPathMap[crawlerName],
					"Pipfile",
				),
			)
			if err != nil {
				panic(err)
			}

			os.Setenv("PIPENV_PIPFILE", pipfile)

			cmd := exec.Command("pipenv", "run", "python", crawlerPath, fmt.Sprintf("-u %s", product.Link))

			var outb, errb bytes.Buffer
			cmd.Stdout = &outb
			cmd.Stderr = &errb

			cmd.Run()
			fmt.Println("output: ", outb.String())

			crawlerResult, err := parseCrawlerOutput(regexMap, outb.String())
			if err != nil {
				panic(err)
			}

			var priceInt int
			var originalPriceInt int
			if crawlerResult.Price != "" {
				priceInt, err = strconv.Atoi(crawlerResult.Price)
				if err != nil {
					panic(err)
				}
			}

			if crawlerResult.OriginalPrice != "" {
				originalPriceInt, err = strconv.Atoi(crawlerResult.OriginalPrice)
				if err != nil {
					panic(err)
				}
			}

			productSearchResult := entities.ProductSearchHistory{
				UserID:        user.ID,
				ProductID:     product.ID,
				Price:         priceInt,
				OriginalPrice: originalPriceInt,
				Discount:      crawlerResult.Discount,
			}

			db.ProductSearchHistory().InsertNewHistory(&productSearchResult)

			// Validade output
			if productSearchResult.Price <= product.MaxPrice {
				// Calculate average price over the search period
				productHistory, err := db.ProductSearchHistory().GetHistoryByProductID(product.ID)
				if err != nil {
					panic(err)
				}

				avgProductPrice := getProductAvgPrice(productHistory)
				avgProductDiscount := getAvgDiscount(avgProductPrice, productSearchResult.Price)

				// Informs the user (insert telegram bot logic here)
				fmt.Println("Current price: ", productSearchResult.Price)
				fmt.Println("Average search price: ", avgProductPrice)
				fmt.Println("Announced discount: ", productSearchResult.Discount)
				fmt.Println("Discount over average search price: ", avgProductDiscount)
			}

			fmt.Printf("%+v\n", crawlerResult)
		}
	}

	fmt.Println("Fim")
}

func parseCrawlerOutput(reMap map[string]string, out string) (*entities.CrawlerResult, error) {
	crawlerResult := entities.CrawlerResult{}
	parsedData := make(map[string]interface{})

	for k, v := range reMap {
		re := regexp2.MustCompile(v, 0)
		m, err := re.FindStringMatch(out)
		if err != nil {
			fmt.Println(err.Error())
		}

		if m.String() == "None" {
			parsedData[k] = nil
		} else {
			parsedData[k] = m.String()
		}
	}

	err := mapstructure.Decode(parsedData, &crawlerResult)
	if err != nil {
		return &entities.CrawlerResult{}, err
	}

	return &crawlerResult, nil
}

func getProductAvgPrice(productHistory []entities.ProductSearchHistory) float64 {
	var sum float64 = 0
	for _, product := range productHistory {
		sum += float64(product.Price) / 100
	}

	avg := sum / float64(len(productHistory))
	roundAvg := math.Round(avg*100) / 100

	return roundAvg
}

func getAvgDiscount(avgPrice float64, currentPrice int) string {
	currentPriceFloat := float64(currentPrice) / 100

	priceDiff := avgPrice - currentPriceFloat
	discount := priceDiff / avgPrice

	discountString := fmt.Sprintf("%.2f", discount)

	return discountString
}
