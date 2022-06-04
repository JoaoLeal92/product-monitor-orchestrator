package contracts

import "github.com/JoaoLeal92/product-monitor-orchestrator/entities"

type Crawler interface {
	SetupCrawlerEnv(crawlerPath string, productID string) error
	RunCrawler(crawlerPath string, product entities.Product) (string, error)
}
