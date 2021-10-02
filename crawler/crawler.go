package crawler

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

	"github.com/JoaoLeal92/product-monitor-orchestrator/config"
	"github.com/JoaoLeal92/product-monitor-orchestrator/contracts"
	"github.com/JoaoLeal92/product-monitor-orchestrator/entities"
	"github.com/google/uuid"
)

type Crawler struct {
	parser *ResultParser
	cfg    *config.CrawlerConfig
	svc    contracts.ProductSearchHistoryService
	logger contracts.LoggerContract
}

func NewCrawler(parser *ResultParser, cfg *config.CrawlerConfig, svc contracts.ProductSearchHistoryService, logger contracts.LoggerContract) *Crawler {
	return &Crawler{
		parser: parser,
		cfg:    cfg,
		svc:    svc,
		logger: logger,
	}
}

func (c *Crawler) StartCrawler(userProductsRelation map[string][]entities.ProductRelations) ([]entities.ProductSearchResult, error) {
	c.logger.AddFields(map[string]interface{}{"process_id": uuid.New().String()})
	c.logger.Info("Iniciando processamento dos usuários")

	crawlersResults, err := c.processUsers(userProductsRelation)
	if err != nil {
		return []entities.ProductSearchResult{}, err
	}

	return crawlersResults, nil
}

func (c *Crawler) processUsers(userProductsRelation map[string][]entities.ProductRelations) ([]entities.ProductSearchResult, error) {
	var crawlersResults []entities.ProductSearchResult

	for userID, products := range userProductsRelation {
		c.logger.AddFields(map[string]interface{}{"user_id": userID})
		c.logger.Info(fmt.Sprintf("Processando usuário com id: %s", userID))

		processResult := c.processProducts(products, userID)
		crawlersResults = append(crawlersResults, processResult...)
	}

	return crawlersResults, nil
}

func (c *Crawler) processProducts(productsRelations []entities.ProductRelations, userID string) []entities.ProductSearchResult {
	c.logger.Info("Processando produtos")

	numWorkers := c.cfg.NumCrawlers
	processingChannels := c.setupProcessChannels(numWorkers)

	go c.allocateCrawlerJobs(productsRelations, processingChannels.CrawlerJobsChan)
	go c.processResultsFromJobs(userID, processingChannels)
	c.createWorkerPool(numWorkers, processingChannels)
	processingResults := <-processingChannels.ProcessingResultChan

	return processingResults
}

func (c *Crawler) setupProcessChannels(numWorkers int) entities.ProcessingChannels {
	processingChannels := entities.ProcessingChannels{
		CrawlerJobsChan:      make(chan entities.ProductRelations, numWorkers),
		CrawlerResultsChan:   make(chan entities.CrawlerChanRestult, numWorkers),
		ProcessingResultChan: make(chan []entities.ProductSearchResult),
	}

	return processingChannels
}

func (c *Crawler) allocateCrawlerJobs(productsRelations []entities.ProductRelations, crawlerJobsChan chan entities.ProductRelations) {
	defer close(crawlerJobsChan)
	for i := 0; i < len(productsRelations); i++ {
		crawlerJobsChan <- productsRelations[i]
	}
}

func (c *Crawler) processResultsFromJobs(userID string, processingChannels entities.ProcessingChannels) {
	defer close(processingChannels.ProcessingResultChan)

	var processingResults []entities.ProductSearchResult
	for channelResult := range processingChannels.CrawlerResultsChan {
		c.logger.Info(fmt.Sprintf("%s Pegando resultado para o produto %s", channelResult.ProductID, channelResult.ProductDescription))

		crawlerResult := c.parser.parseCrawlerResult(channelResult.CrawlerResult)
		if channelResult.CrawlerErr != "" {
			c.logger.Error(fmt.Sprintf("Erro no processamento de %s: %s", channelResult.ProductID, channelResult.CrawlerErr))
		}

		productSearchResult, err := c.svc.CreateProductSearchHistory(crawlerResult, userID, channelResult.ProductID)
		if err != nil {
			c.logger.Error(fmt.Sprintf("%s: Erro na criação de histórico", channelResult.ProductID))
		}

		c.logger.Info(fmt.Sprintf("Retornando resultado do produto %s para chanel de usuário", channelResult.ProductDescription))
		processingResults = append(processingResults, productSearchResult)

	}
	c.logger.Info("Finalizando processamento dos resultados")
	processingChannels.ProcessingResultChan <- processingResults
}

func (c *Crawler) createWorkerPool(numWorkers int, processingChannels entities.ProcessingChannels) {
	var wg sync.WaitGroup
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go c.crawlProduct(processingChannels, &wg)
	}
	wg.Wait()
	close(processingChannels.CrawlerResultsChan)
}

func (c *Crawler) crawlProduct(processingChannels entities.ProcessingChannels, wg *sync.WaitGroup) {
	for job := range processingChannels.CrawlerJobsChan {
		c.logger.Info(fmt.Sprintf("%s Iniciando processamento do produto %s", job.Product.ID, job.Product.Description))

		err := c.setupCrawlerEnv(job.CrawlerPath, job.Product.ID.String())
		if err != nil {
			processingChannels.CrawlerResultsChan <- entities.CrawlerChanRestult{
				CrawlerResult:      "",
				CrawlerErr:         "Error setting up crawler environment",
				ProductID:          job.Product.ID,
				ProductDescription: job.Product.Description,
			}
		}

		crawlerOutput, err := c.runCrawler(job.CrawlerPath, job.Product)
		if err != nil {
			processingChannels.CrawlerResultsChan <- entities.CrawlerChanRestult{
				CrawlerResult:      "",
				CrawlerErr:         "Error running crawler for product",
				ProductID:          job.Product.ID,
				ProductDescription: job.Product.Description,
			}
		}

		crawlerResult := entities.CrawlerChanRestult{
			CrawlerResult:      crawlerOutput.OutputMessage,
			CrawlerErr:         crawlerOutput.OutputError,
			ProductID:          job.Product.ID,
			ProductDescription: job.Product.Description,
		}

		processingChannels.CrawlerResultsChan <- crawlerResult
	}
	wg.Done()
}

func (c *Crawler) setupCrawlerEnv(crawlerPath string, productID string) error {
	c.logger.Info(fmt.Sprintf("Preparando ambiente para processamento do produto %s", productID))
	pipfile, err := filepath.Abs(
		filepath.Join(
			crawlerPath,
			"Pipfile",
		),
	)
	if err != nil {
		return err
	}

	os.Setenv("PIPENV_PIPFILE", pipfile)
	fmt.Println("Ambiente pronto para execução")
	return nil
}

func (c *Crawler) runCrawler(crawlerPath string, product entities.Product) (entities.CrawlerOutput, error) {
	c.logger.Info(fmt.Sprintf("%s Executando crawler no link %s", product.ID.String(), product.Link))
	cmd := exec.Command("pipenv", "run", "python", crawlerPath, fmt.Sprintf("-u %s", product.Link))

	var outb, errb bytes.Buffer
	cmd.Stdout = &outb
	cmd.Stderr = &errb

	err := cmd.Run()
	if err != nil {
		return entities.CrawlerOutput{}, err
	}

	crawlerOutput := entities.CrawlerOutput{
		OutputMessage: outb.String(),
		OutputError:   errb.String(),
	}

	return crawlerOutput, nil
}
