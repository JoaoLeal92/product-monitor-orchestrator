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

	for userID, products := range userProductsRelation {
		fmt.Println("Processing user id: ", userID)

		crawlerResults, err := c.processProducts(products, userID)
		if err != nil {
			return processingResults, err
		}

		processingResults = append(processingResults, crawlerResults...)
	}

	return processingResults, nil
}

func (c *Crawler) processProducts(productsRelations []entities.ProductRelations, userID string) ([]entities.ProductSearchResult, error) {
	fmt.Println("Processando produtos")
	var crawlerResults []entities.ProductSearchResult

	for _, productRelation := range productsRelations {
		fmt.Println("Iniciando processamento do produto ", productRelation.Product.Description)

		err := c.setupCrawlerEnv(productRelation.CrawlerPath)
		if err != nil {
			return crawlerResults, err
		}

		crawlerOutput, crawlerError := c.runCrawler(productRelation.CrawlerPath, productRelation.Product.Link)
		if crawlerError != "" {
			fmt.Println("Crawler error: ", crawlerError)
		}

		crawlerResult := c.parser.parseCrawlerResult(crawlerOutput)

		productSearchResult, err := c.createProductSearchHistory(crawlerResult, userID, productRelation.Product.ID)
		if err != nil {
			return crawlerResults, err
		}

		crawlerResults = append(crawlerResults, productSearchResult)
	}

	return crawlerResults, nil
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

func (c *Crawler) runCrawler(crawlerPath string, productLink string) (string, string) {
	fmt.Println("Executando crawler")
	cmd := exec.Command("pipenv", "run", "python", crawlerPath, fmt.Sprintf("-u %s", productLink))

	var outb, errb bytes.Buffer
	cmd.Stdout = &outb
	cmd.Stderr = &errb

	err := cmd.Run()
	if err != nil {
		return "", ""
	}

	fmt.Println("String de retorno do comando de execução: ", outb.String())
	fmt.Println("String de erro do comando de execução: ", errb.String())

	return outb.String(), errb.String()
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
