package crawler

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"

	"github.com/JoaoLeal92/product-monitor-orchestrator/data"
	"github.com/JoaoLeal92/product-monitor-orchestrator/entities"
	"github.com/google/uuid"
)

type ChanResult struct {
	Result entities.ProductSearchResult
	Err    error
}

type CrawlerRestultChan struct {
	CrawlerResult string
	CrawlerErr    string
	ProductID     uuid.UUID
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
	defer close(rchan)
	var crawlerResults = []chan CrawlerRestultChan{}

	fmt.Println("Processando produtos")

	chanIndex := 0
	for _, productRelation := range productsRelations {
		fmt.Println("Iniciando processamento do produto ", productRelation.Product.Description)

		err := c.setupCrawlerEnv(productRelation.CrawlerPath)
		if err != nil {
			rchan <- ChanResult{Result: entities.ProductSearchResult{}, Err: err}
		}

		crawlerResults = append(crawlerResults, make(chan CrawlerRestultChan))
		go c.runCrawler(
			ProductData{
				crawlerPath: productRelation.CrawlerPath,
				productLink: productRelation.Product.Link,
				productID:   productRelation.Product.ID,
			},
			crawlerResults[chanIndex],
		)

		chanIndex++
	}

	for _, channel := range crawlerResults {
		for ch := range channel {

			crawlerResult := c.parser.parseCrawlerResult(ch.CrawlerResult)

			productSearchResult, err := c.createProductSearchHistory(crawlerResult, userID, ch.ProductID)
			if err != nil {
				rchan <- ChanResult{Result: entities.ProductSearchResult{}, Err: err}
			}

			rchan <- ChanResult{Result: productSearchResult, Err: nil}
		}
	}
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
