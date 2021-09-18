package main

import (
	"fmt"
	"os"

	"github.com/JoaoLeal92/product-monitor-orchestrator/config"
	"github.com/JoaoLeal92/product-monitor-orchestrator/crawler"
	"github.com/JoaoLeal92/product-monitor-orchestrator/data"
)

func main() {
	fmt.Println("Start orchestrator")
	cfg, err := config.ReadConfig()
	if err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}

	db, _ := data.Instance(cfg.Db)
	agg := NewAggregator(db, &cfg.Crawlers)
	validator := NewValidator(db)
	parser := crawler.NewResultParser()

	userProductsRelations, err := agg.SetupUserProductRelations()
	fmt.Printf("User products relations: %+v", userProductsRelations)

	crawler := crawler.NewCrawler(parser, db)
	crawlerResults, err := crawler.StartCrawler(userProductsRelations)
	if err != nil {
		panic(err)
	}
	fmt.Println("Produtos processados com sucesso")

	err = validator.ValidateCrawlerResults(crawlerResults)

	fmt.Println("Fim")
}
