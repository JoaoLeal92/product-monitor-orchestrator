package crawler

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/JoaoLeal92/product-monitor-orchestrator/data"
	"github.com/JoaoLeal92/product-monitor-orchestrator/entities"
	"github.com/google/uuid"
)

type ChanResult struct {
	Result entities.ProductSearchResult
	Err    error
}

type CrawlerRestultChan struct {
	CrawlerResult      string
	CrawlerErr         string
	ProductID          uuid.UUID
	ProductDescription string
}

type ProductData struct {
	crawlerPath string
	productLink string
	productID   uuid.UUID
}

type Crawler struct {
	parser *ResultParser
	db     *data.Connection
}

func NewCrawler(parser *ResultParser, db *data.Connection) *Crawler {
	return &Crawler{
		parser: parser,
		db:     db,
	}
}

func (c *Crawler) StartCrawler(userProductsRelation map[string][]entities.ProductRelations) ([]entities.ProductSearchResult, error) {
	fmt.Println("Iniciando processamento dos usuários")
	processingResults, err := c.processUsers(userProductsRelation)
	if err != nil {
		return []entities.ProductSearchResult{}, err
	}

	return processingResults, nil
}

func (c *Crawler) processUsers(userProductsRelation map[string][]entities.ProductRelations) ([]entities.ProductSearchResult, error) {
	var processingResults []entities.ProductSearchResult
	var results = []chan ChanResult{}

	chanIndex := 0
	for userID, products := range userProductsRelation {
		results = append(results, make(chan ChanResult))
		fmt.Println("Processing user id: ", userID)

		go c.processProducts(products, userID, results[chanIndex])

		chanIndex++
	}

	for _, result := range results {
		for v := range result {
			if v.Err != nil {
				return []entities.ProductSearchResult{}, v.Err
			}
			processingResults = append(processingResults, v.Result)
		}
	}

	return processingResults, nil
}

func (c *Crawler) processProducts(productsRelations []entities.ProductRelations, userID string, rchan chan ChanResult) {
	fmt.Println("Processando produtos")

	var jobs = make(chan entities.ProductRelations, 5)
	var results = make(chan CrawlerRestultChan, 5)
	var doneChan = make(chan bool)
	noWorkers := 5

	go c.allocateJobs(productsRelations, jobs)
	go c.getResultsFromJobs(userID, results, doneChan, rchan)
	c.createWorkerPool(noWorkers, jobs, results)
	<-doneChan
}

func (c *Crawler) allocateJobs(productsRelations []entities.ProductRelations, jobChan chan entities.ProductRelations) {
	defer close(jobChan)
	for i := 0; i < len(productsRelations); i++ {
		jobChan <- productsRelations[i]
	}
}

func (c *Crawler) getResultsFromJobs(userID string, resultsChan chan CrawlerRestultChan, doneChan chan bool, userResultChan chan ChanResult) {
	defer close(userResultChan)
	for result := range resultsChan {
		fmt.Println("Pegando resultado para o produto ", result.ProductDescription)
		crawlerResult := c.parser.parseCrawlerResult(result.CrawlerResult)

		productSearchResult, err := c.createProductSearchHistory(crawlerResult, userID, result.ProductID)

		if err != nil {
			userResultChan <- ChanResult{Result: entities.ProductSearchResult{}, Err: err}
		}

		fmt.Println("Retornando resultado do produto", result.ProductDescription, " para chanel de usuário")
		userResultChan <- ChanResult{Result: productSearchResult, Err: nil}
		fmt.Println("Resultado do produto ", result.ProductDescription, " retornado")

	}
	fmt.Println("Finalizando processamento dos resultados")
	doneChan <- true
}

func (c *Crawler) createWorkerPool(noWorkers int, jobChan chan entities.ProductRelations, resultsChan chan CrawlerRestultChan) {
	var wg sync.WaitGroup
	for i := 0; i < noWorkers; i++ {
		wg.Add(1)
		fmt.Println("Inicia worker  ", i)
		go c.processProduct(jobChan, resultsChan, &wg)
	}
	wg.Wait()
	close(resultsChan)
}

func (c *Crawler) processProduct(jobChan chan entities.ProductRelations, resultsChan chan CrawlerRestultChan, wg *sync.WaitGroup) {
	for job := range jobChan {
		fmt.Println("Iniciando processamento do produto ", job.Product.Description)

		err := c.setupCrawlerEnv(job.CrawlerPath)
		if err != nil {
			resultsChan <- CrawlerRestultChan{
				CrawlerResult:      "",
				CrawlerErr:         "Error setting up crawler environment",
				ProductID:          job.Product.ID,
				ProductDescription: job.Product.Description,
			}
		}

		fmt.Println("Executando crawler")
		cmd := exec.Command("pipenv", "run", "python", job.CrawlerPath, fmt.Sprintf("-u %s", job.Product.Link))

		var outb, errb bytes.Buffer
		cmd.Stdout = &outb
		cmd.Stderr = &errb

		err = cmd.Run()
		if err != nil {
			resultsChan <- CrawlerRestultChan{}
		}

		fmt.Println("String de retorno do comando de execução: ", outb.String())
		fmt.Println("String de erro do comando de execução: ", errb.String())

		crawlerResult := CrawlerRestultChan{
			CrawlerResult:      outb.String(),
			CrawlerErr:         errb.String(),
			ProductID:          job.Product.ID,
			ProductDescription: job.Product.Description,
		}

		resultsChan <- crawlerResult
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

func (c *Crawler) runCrawler(productData ProductData, crawlerChannel chan CrawlerRestultChan) {
	defer close(crawlerChannel)

	fmt.Println("Executando crawler")
	cmd := exec.Command("pipenv", "run", "python", productData.crawlerPath, fmt.Sprintf("-u %s", productData.productLink))

	var outb, errb bytes.Buffer
	cmd.Stdout = &outb
	cmd.Stderr = &errb

	err := cmd.Run()
	if err != nil {
		crawlerChannel <- CrawlerRestultChan{}
	}

	fmt.Println("String de retorno do comando de execução: ", outb.String())
	fmt.Println("String de erro do comando de execução: ", errb.String())

	crawlerResult := CrawlerRestultChan{
		CrawlerResult: outb.String(),
		CrawlerErr:    errb.String(),
		ProductID:     productData.productID,
	}

	crawlerChannel <- crawlerResult
}

func (c *Crawler) createProductSearchHistory(crawlerResult *entities.CrawlerResult, userID string, productID uuid.UUID) (entities.ProductSearchResult, error) {
	fmt.Println("Inserindo histórico de pesquisa para o produto")
	var priceInt int
	var originalPriceInt int
	var err error

	if crawlerResult.Price != "" {
		priceInt, err = strconv.Atoi(crawlerResult.Price)
		if err != nil {
			return entities.ProductSearchResult{}, err
		}
	}

	if crawlerResult.OriginalPrice != "" {
		originalPriceInt, err = strconv.Atoi(crawlerResult.OriginalPrice)
		if err != nil {
			return entities.ProductSearchResult{}, err
		}
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return entities.ProductSearchResult{}, err
	}

	productSearchResult := entities.ProductSearchResult{
		UserID:        userUUID,
		ProductID:     productID,
		Price:         priceInt,
		OriginalPrice: originalPriceInt,
		Discount:      crawlerResult.Discount,
	}

	c.db.ProductSearchHistory().InsertNewHistory(&productSearchResult)
	fmt.Println("Histórico inserido com sucesso")

	return productSearchResult, nil
}
