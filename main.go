package main

import (
	"fmt"
	"os"
	"time"

	"github.com/JoaoLeal92/product-monitor-orchestrator/config"
	"github.com/JoaoLeal92/product-monitor-orchestrator/crawler"
	"github.com/JoaoLeal92/product-monitor-orchestrator/data"
	"github.com/JoaoLeal92/product-monitor-orchestrator/infra/logs"
	"github.com/JoaoLeal92/product-monitor-orchestrator/infra/queue"
	"github.com/JoaoLeal92/product-monitor-orchestrator/services"
)

func main() {
	fmt.Println("Start orchestrator")
	cfg, err := config.ReadConfig()
	if err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}

	logger := logs.NewLogger(&cfg.Log)
	db, _ := data.Instance(cfg.Db)
	agg := NewAggregator(db, &cfg.Crawlers)
	parser := crawler.NewResultParser()

	queueManager, err := queue.NewQueueManager(&cfg.Queue)
	if err != nil {
		logger.Error(fmt.Sprintf("Erro ao conectar-se com o gerenciador de filas: %v", err))
	}
	defer queueManager.CloseConnection()
	defer queueManager.CloseChannel()

	validator := NewValidator(db, queueManager)

	productSearchHistoryService := services.NewProductSearchHistoryService(db)

	userProductsRelations, err := agg.SetupUserProductRelations()
	if err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}

	crawler := crawler.NewCrawler(parser, &cfg.Crawlers, productSearchHistoryService, logger)

	startTime := time.Now()
	crawlerResults, err := crawler.StartCrawler(userProductsRelations)
	elapsedTime := time.Since(startTime)
	if err != nil {
		panic(err)
	}

	logger.ClearField("user_id")
	logger.Info("Produtos processados com sucesso")
	logger.Info(fmt.Sprintf("Tempo de execução do crawler: %v", elapsedTime))

	err = validator.ValidateCrawlerResults(crawlerResults)
	if err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}

	logger.Info("Fim da operação do crawler")
}
