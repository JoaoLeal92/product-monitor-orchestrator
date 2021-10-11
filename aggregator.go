package main

import (
	"errors"
	"path/filepath"
	"strings"

	"github.com/JoaoLeal92/product-monitor-orchestrator/config"
	"github.com/JoaoLeal92/product-monitor-orchestrator/data"
	"github.com/JoaoLeal92/product-monitor-orchestrator/entities"

	"github.com/oleiade/reflections"
)

type Aggregator struct {
	db  *data.Connection
	cfg *config.CrawlerConfig
}

func NewAggregator(db *data.Connection, cfg *config.CrawlerConfig) *Aggregator {
	return &Aggregator{
		db:  db,
		cfg: cfg,
	}
}

func (a *Aggregator) SetupUserProductRelations() (map[string][]entities.ProductRelations, error) {
	usersProductsRelation := make(map[string][]entities.ProductRelations)

	users, err := a.db.Users().GetActiveUsers()
	if err != nil {
		return usersProductsRelation, err
	}

	for _, user := range users {
		var productsRelations []entities.ProductRelations

		products, err := a.db.Products().GetActiveProductsByUser(user.ID)
		if err != nil {
			return usersProductsRelation, err
		}

		for _, product := range products {
			crawlerName, err := a.db.Crawlers().GetCrawlerName(product.CrawlerID)
			if err != nil {
				return usersProductsRelation, err
			}

			crawlerPath, err := a.getCrawlerPath(crawlerName)

			relation := entities.ProductRelations{
				Product:     product,
				CrawlerPath: crawlerPath,
			}
			productsRelations = append(productsRelations, relation)
		}

		usersProductsRelation[user.ID.String()] = productsRelations
	}

	return usersProductsRelation, nil
}

func (a *Aggregator) getCrawlerPath(crawlerName string) (string, error) {
	crawlerPath, err := reflections.GetField(a.cfg, strings.Title(crawlerName))

	crawlerPathStr, ok := crawlerPath.(string)
	if !ok {
		return "", errors.New("Invalid crawler path")
	}

	crawlerAbsPath, err := filepath.Abs(crawlerPathStr)
	if err != nil {
		return "", err
	}

	return crawlerAbsPath, nil
}
