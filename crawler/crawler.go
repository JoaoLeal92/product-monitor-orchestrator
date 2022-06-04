package crawler

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/JoaoLeal92/product-monitor-orchestrator/config"
	"github.com/JoaoLeal92/product-monitor-orchestrator/contracts"
	"github.com/JoaoLeal92/product-monitor-orchestrator/entities"
)

type Crawler struct {
	cfg    *config.CrawlerConfig
	logger contracts.LoggerContract
}

func NewCrawler(cfg *config.CrawlerConfig, logger contracts.LoggerContract) *Crawler {
	return &Crawler{
		cfg:    cfg,
		logger: logger,
	}
}

func (c *Crawler) SetupCrawlerEnv(crawlerPath string, productID string) error {
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

func (c *Crawler) RunCrawler(crawlerPath string, product entities.Product) (string, error) {
	c.logger.Info(fmt.Sprintf("%s Executando crawler no link %s", product.ID.String(), product.Link))
	cmd := exec.Command("pipenv", "run", "python", crawlerPath, fmt.Sprintf("-u %s", product.Link))

	var outb, errb bytes.Buffer
	cmd.Stdout = &outb
	cmd.Stderr = &errb

	err := cmd.Run()
	if err != nil {
		return "", err
	}

	if !strings.Contains(errb.String(), "Loading .env environment variables") {
		c.logger.Error(fmt.Sprintf("Erro no crawler do produto %s", product.Description))
		c.logger.Error(errb.String())
		return "", errors.New(errb.String())
	}

	return outb.String(), nil
}
