package services

import (
	"fmt"
	"sync"

	"github.com/JoaoLeal92/product-monitor-orchestrator/config"
	"github.com/JoaoLeal92/product-monitor-orchestrator/contracts"
	"github.com/JoaoLeal92/product-monitor-orchestrator/entities"
	crawlerparser "github.com/JoaoLeal92/product-monitor-orchestrator/infra/crawerParser"
	"github.com/google/uuid"
)

type CrawlerService struct {
	parser          *crawlerparser.ResultParser
	cfg             *config.CrawlerConfig
	notificationSvc contracts.ProductNotificationService
	logger          contracts.LoggerContract
	crawler         contracts.Crawler
}

type processingChannels struct {
	CrawlerJobsChan      chan entities.Product
	CrawlerResultsChan   chan crawlerChanResult
	EndProcessingChannel chan bool
}

type crawlerChanResult struct {
	CrawlerResult string
	Product       entities.Product
}

func NewCrawlerService(parser *crawlerparser.ResultParser, cfg *config.CrawlerConfig, notificationSvc contracts.ProductNotificationService, logger contracts.LoggerContract, crawler contracts.Crawler) *CrawlerService {
	return &CrawlerService{
		parser:          parser,
		cfg:             cfg,
		notificationSvc: notificationSvc,
		logger:          logger,
		crawler:         crawler,
	}
}

func (c *CrawlerService) Execute(products []entities.Product) error {
	c.logger.AddFields(map[string]interface{}{"process_id": uuid.New().String()})
	c.logger.Info("Iniciando processamento dos produtos")

	err := c.processProducts(products)
	if err != nil {
		return err
	}

	return nil
}

func (c *CrawlerService) processProducts(productsRelations []entities.Product) error {
	c.logger.Info("Processando produtos")

	numWorkers := c.cfg.NumCrawlers
	processingChannels := c.setupProcessChannels(numWorkers)

	go c.allocateCrawlerJobs(productsRelations, processingChannels.CrawlerJobsChan)
	go c.processResultsFromJobs(processingChannels)
	c.createWorkerPool(numWorkers, processingChannels)
	<-processingChannels.EndProcessingChannel

	return nil
}

func (c *CrawlerService) setupProcessChannels(numWorkers int) processingChannels {
	processingChannels := processingChannels{
		CrawlerJobsChan:      make(chan entities.Product, numWorkers),
		CrawlerResultsChan:   make(chan crawlerChanResult, numWorkers),
		EndProcessingChannel: make(chan bool),
	}

	return processingChannels
}

func (c *CrawlerService) allocateCrawlerJobs(products []entities.Product, crawlerJobsChan chan entities.Product) {
	defer close(crawlerJobsChan)
	for i := 0; i < len(products); i++ {
		crawlerJobsChan <- products[i]
	}
}

func (c *CrawlerService) createWorkerPool(numWorkers int, processingChannels processingChannels) {
	var wg sync.WaitGroup
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go c.crawlProduct(processingChannels, &wg)
	}
	wg.Wait()
	close(processingChannels.CrawlerResultsChan)
}

func (c *CrawlerService) crawlProduct(processingChannels processingChannels, wg *sync.WaitGroup) {
	for job := range processingChannels.CrawlerJobsChan {
		c.logger.Info(fmt.Sprintf("%s Iniciando processamento do produto %s", job.ID, job.Description))

		crawlerPath, err := job.GetCrawlerPath(c.cfg)
		if err != nil {
			c.logger.Error(fmt.Sprintf("erro na busca de produto de id %s para usuário %s", job.ID.String(), job.UserID.String()))
			c.logger.Error(err.Error())
			continue
		}

		err = c.crawler.SetupCrawlerEnv(crawlerPath, job.ID.String())
		if err != nil {
			c.logger.Error(fmt.Sprintf("erro na busca de produto de id %s para usuário %s", job.ID.String(), job.UserID.String()))
			c.logger.Error(err.Error())
			continue
		}

		crawlerOutput, err := c.crawler.RunCrawler(crawlerPath, job)
		if err != nil {
			c.logger.Error(fmt.Sprintf("erro na busca de produto de id %s para usuário %s", job.ID.String(), job.UserID.String()))
			c.logger.Error(err.Error())
			continue
		}

		crawlerResult := crawlerChanResult{
			CrawlerResult: crawlerOutput,
			Product:       job,
		}

		processingChannels.CrawlerResultsChan <- crawlerResult
	}
	wg.Done()
}

func (c *CrawlerService) processResultsFromJobs(processingChannels processingChannels) {
	defer close(processingChannels.EndProcessingChannel)

	for channelResult := range processingChannels.CrawlerResultsChan {
		c.logger.Info(fmt.Sprintf("%s Pegando resultado para o produto %s", channelResult.Product.ID, channelResult.Product.Description))

		crawlerResult := c.parser.ParseCrawlerResult(channelResult.CrawlerResult)
		productSearchResult := entities.ProductSearchResult{
			ProductID:     channelResult.Product.ID,
			UserID:        channelResult.Product.UserID,
			Price:         crawlerResult.Price,
			OriginalPrice: crawlerResult.OriginalPrice,
			Discount:      crawlerResult.Discount,
		}

		err := c.notificationSvc.Execute(&channelResult.Product, &productSearchResult)
		if err != nil {
			c.logger.Error(fmt.Sprintf("%s: Erro na criação de histórico", channelResult.Product.ID))
			c.logger.Error(err.Error())
		}
	}
	c.logger.Info("Finalizando processamento dos resultados")
	processingChannels.EndProcessingChannel <- true
}
