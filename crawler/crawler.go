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
	parser          *ResultParser
	cfg             *config.CrawlerConfig
	notificationSvc contracts.ProductNotificationService
	logger          contracts.LoggerContract
}

type processingChannels struct {
	CrawlerJobsChan      chan entities.Product
	CrawlerResultsChan   chan crawlerChanResult
	EndProcessingChannel chan bool
}

type crawlerChanResult struct {
	CrawlerResult string
	CrawlerErr    string
	Product       entities.Product
}

type crawlerOutput struct {
	OutputMessage string
	OutputError   string
}

func NewCrawler(parser *ResultParser, cfg *config.CrawlerConfig, notificationSvc contracts.ProductNotificationService, logger contracts.LoggerContract) *Crawler {
	return &Crawler{
		parser:          parser,
		cfg:             cfg,
		notificationSvc: notificationSvc,
		logger:          logger,
	}
}

func (c *Crawler) StartCrawler(products []entities.Product) error {
	c.logger.AddFields(map[string]interface{}{"process_id": uuid.New().String()})
	c.logger.Info("Iniciando processamento dos produtos")

	err := c.processProducts(products)
	if err != nil {
		return err
	}

	return nil
}

func (c *Crawler) processProducts(productsRelations []entities.Product) error {
	c.logger.Info("Processando produtos")

	numWorkers := c.cfg.NumCrawlers
	processingChannels := c.setupProcessChannels(numWorkers)

	go c.allocateCrawlerJobs(productsRelations, processingChannels.CrawlerJobsChan)
	go c.processResultsFromJobs(processingChannels)
	c.createWorkerPool(numWorkers, processingChannels)
	<-processingChannels.EndProcessingChannel

	return nil
}

func (c *Crawler) setupProcessChannels(numWorkers int) processingChannels {
	processingChannels := processingChannels{
		CrawlerJobsChan:      make(chan entities.Product, numWorkers),
		CrawlerResultsChan:   make(chan crawlerChanResult, numWorkers),
		EndProcessingChannel: make(chan bool),
	}

	return processingChannels
}

func (c *Crawler) allocateCrawlerJobs(products []entities.Product, crawlerJobsChan chan entities.Product) {
	defer close(crawlerJobsChan)
	for i := 0; i < len(products); i++ {
		crawlerJobsChan <- products[i]
	}
}

func (c *Crawler) processResultsFromJobs(processingChannels processingChannels) {
	defer close(processingChannels.EndProcessingChannel)

	for channelResult := range processingChannels.CrawlerResultsChan {
		c.logger.Info(fmt.Sprintf("%s Pegando resultado para o produto %s", channelResult.Product.ID, channelResult.Product.Description))

		crawlerResult := c.parser.parseCrawlerResult(channelResult.CrawlerResult)
		if channelResult.CrawlerErr != "" {
			c.logger.Error(fmt.Sprintf("Erro no processamento de %s: %s", channelResult.Product.ID, channelResult.CrawlerErr))
		}
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

func (c *Crawler) createWorkerPool(numWorkers int, processingChannels processingChannels) {
	var wg sync.WaitGroup
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go c.crawlProduct(processingChannels, &wg)
	}
	wg.Wait()
	close(processingChannels.CrawlerResultsChan)
}

func (c *Crawler) crawlProduct(processingChannels processingChannels, wg *sync.WaitGroup) {
	for job := range processingChannels.CrawlerJobsChan {
		c.logger.Info(fmt.Sprintf("%s Iniciando processamento do produto %s", job.ID, job.Description))

		crawlerPath, err := job.GetCrawlerPath(c.cfg)
		if err != nil {
			c.logger.Error(fmt.Sprintf("erro na busca de produto de id %s para usuário %s", job.ID.String(), job.UserID.String()))
			c.logger.Error(err.Error())
			continue
		}

		err = c.setupCrawlerEnv(crawlerPath, job.ID.String())
		if err != nil {
			c.logger.Error(fmt.Sprintf("erro na busca de produto de id %s para usuário %s", job.ID.String(), job.UserID.String()))
			c.logger.Error(err.Error())
			continue
		}

		crawlerOutput, err := c.runCrawler(crawlerPath, job)
		if err != nil {
			c.logger.Error(fmt.Sprintf("erro na busca de produto de id %s para usuário %s", job.ID.String(), job.UserID.String()))
			c.logger.Error(err.Error())
			continue
		}

		crawlerResult := crawlerChanResult{
			CrawlerResult: crawlerOutput.OutputMessage,
			CrawlerErr:    crawlerOutput.OutputError,
			Product:       job,
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

func (c *Crawler) runCrawler(crawlerPath string, product entities.Product) (crawlerOutput, error) {
	c.logger.Info(fmt.Sprintf("%s Executando crawler no link %s", product.ID.String(), product.Link))
	cmd := exec.Command("pipenv", "run", "python", crawlerPath, fmt.Sprintf("-u %s", product.Link))

	var outb, errb bytes.Buffer
	cmd.Stdout = &outb
	cmd.Stderr = &errb

	err := cmd.Run()
	if err != nil {
		return crawlerOutput{}, err
	}

	crawlerOutput := crawlerOutput{
		OutputMessage: outb.String(),
		OutputError:   errb.String(),
	}

	return crawlerOutput, nil
}
