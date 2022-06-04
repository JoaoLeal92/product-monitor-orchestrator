package main

import (
	"fmt"
	"os"
	"time"

	"github.com/JoaoLeal92/product-monitor-orchestrator/config"
	"github.com/JoaoLeal92/product-monitor-orchestrator/crawler"
	"github.com/JoaoLeal92/product-monitor-orchestrator/data"
	crawlerparser "github.com/JoaoLeal92/product-monitor-orchestrator/infra/crawlerParser"
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
	parser := crawlerparser.NewResultParser()

	queueManager, err := queue.NewQueueManager(&cfg.Queue)
	if err != nil {
		logger.Error(fmt.Sprintf("Erro ao conectar-se com o gerenciador de filas: %v", err))
		os.Exit(1)
	}
	defer queueManager.CloseConnection()
	defer queueManager.CloseChannel()

	products, err := db.Products().GetProductsListForCrawler()
	if err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}

	productNotificationService := services.NewProductNotificationService(db, logger, queueManager)
	crawler := crawler.NewCrawler(&cfg.Crawlers, logger)
	crawlerService := services.NewCrawlerService(parser, &cfg.Crawlers, productNotificationService, logger, crawler)

	startTime := time.Now()
	err = crawlerService.Execute(products)
	elapsedTime := time.Since(startTime)
	if err != nil {
		panic(err)
	}

	logger.ClearField("user_id")
	logger.Info("Produtos processados com sucesso")
	logger.Info(fmt.Sprintf("Tempo de execução do crawler: %v", elapsedTime))

	logger.Info("Fim da operação do crawler")
}
