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
)

type Crawler struct {
	parser *ResultParser
	cfg    *config.CrawlerConfig
	svc    contracts.ProductSearchHistoryService
}

func NewCrawler(parser *ResultParser, cfg *config.CrawlerConfig, svc contracts.ProductSearchHistoryService) *Crawler {
	return &Crawler{
		parser: parser,
		cfg:    cfg,
		svc:    svc,
	}
}

func (c *Crawler) StartCrawler(userProductsRelation map[string][]entities.ProductRelations) ([]entities.ProductSearchResult, error) {
	fmt.Println("Iniciando processamento dos usuários")
	crawlersResults, err := c.processUsers(userProductsRelation)
	if err != nil {
		return []entities.ProductSearchResult{}, err
	}

	return crawlersResults, nil
}

func (c *Crawler) processUsers(userProductsRelation map[string][]entities.ProductRelations) ([]entities.ProductSearchResult, error) {
	var crawlersResults []entities.ProductSearchResult

	for userID, products := range userProductsRelation {
		fmt.Println("Processando usuário com id: ", userID)

		processResult := c.processProducts(products, userID)
		crawlersResults = append(crawlersResults, processResult...)
	}

	return crawlersResults, nil
}

func (c *Crawler) processProducts(productsRelations []entities.ProductRelations, userID string) []entities.ProductSearchResult {
	fmt.Println("Processando produtos")

	numWorkers := c.cfg.NumCrawlers
	processingChannels := entities.ProcessingChannels{
		CrawlerJobsChan:      make(chan entities.ProductRelations, numWorkers),
		CrawlerResultsChan:   make(chan entities.CrawlerChanRestult, numWorkers),
		ProcessingResultChan: make(chan []entities.ProductSearchResult),
	}

	go c.allocateCrawlerJobs(productsRelations, processingChannels.CrawlerJobsChan)
	go c.processResultsFromJobs(userID, processingChannels)
	c.createWorkerPool(numWorkers, processingChannels)
	processingResults := <-processingChannels.ProcessingResultChan

	return processingResults
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
		fmt.Println("Pegando resultado para o produto ", channelResult.ProductDescription)
		crawlerResult := c.parser.parseCrawlerResult(channelResult.CrawlerResult)

		productSearchResult, err := c.svc.CreateProductSearchHistory(crawlerResult, userID, channelResult.ProductID)

		if err != nil {
			fmt.Println(err.Error())
		}

		fmt.Println("Retornando resultado do produto ", channelResult.ProductDescription, " para chanel de usuário")
		processingResults = append(processingResults, productSearchResult)

	}
	fmt.Println("Finalizando processamento dos resultados")
	processingChannels.ProcessingResultChan <- processingResults
}

func (c *Crawler) createWorkerPool(numWorkers int, processingChannels entities.ProcessingChannels) {
	var wg sync.WaitGroup
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		fmt.Println("Inicia worker  ", i)
		go c.crawlProduct(processingChannels, &wg)
	}
	wg.Wait()
	close(processingChannels.CrawlerResultsChan)
}

func (c *Crawler) crawlProduct(processingChannels entities.ProcessingChannels, wg *sync.WaitGroup) {
	for job := range processingChannels.CrawlerJobsChan {
		fmt.Println("Iniciando processamento do produto ", job.Product.Description)

		err := c.setupCrawlerEnv(job.CrawlerPath)
		if err != nil {
			processingChannels.CrawlerResultsChan <- entities.CrawlerChanRestult{
				CrawlerResult:      "",
				CrawlerErr:         "Error setting up crawler environment",
				ProductID:          job.Product.ID,
				ProductDescription: job.Product.Description,
			}
		}

		crawlerOutput, err := c.runCrawler(job.CrawlerPath, job.Product.Link)
		if err != nil {
			processingChannels.CrawlerResultsChan <- entities.CrawlerChanRestult{
				CrawlerResult:      "",
				CrawlerErr:         "Error running crawler for product",
				ProductID:          job.Product.ID,
				ProductDescription: job.Product.Description,
			}
		}

		fmt.Println("String de retorno do comando de execução: ", crawlerOutput.OutputMessage)
		fmt.Println("String de erro do comando de execução: ", crawlerOutput.OutputError)

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

func (c *Crawler) setupCrawlerEnv(crawlerPath string) error {
	fmt.Println("Preparando ambiente para processamento")
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

func (c *Crawler) runCrawler(crawlerPath string, productLink string) (entities.CrawlerOutput, error) {
	fmt.Println("Executando crawler")
	cmd := exec.Command("pipenv", "run", "python", crawlerPath, fmt.Sprintf("-u %s", productLink))

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
