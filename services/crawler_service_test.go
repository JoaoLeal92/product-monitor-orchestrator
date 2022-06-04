package services

import (
	"errors"
	"testing"

	"github.com/JoaoLeal92/product-monitor-orchestrator/config"
	mocks "github.com/JoaoLeal92/product-monitor-orchestrator/contracts/mocks"
	"github.com/JoaoLeal92/product-monitor-orchestrator/entities"
	crawlerparser "github.com/JoaoLeal92/product-monitor-orchestrator/infra/crawerParser"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCrawlerService(t *testing.T) {
	mockLogger := mocks.NewLoggerContract(t)
	mockCrawler := mocks.NewCrawler(t)
	mockProductNotificationSvc := mocks.NewProductNotificationService(t)

	parser := crawlerparser.NewResultParser()
	cfg := config.CrawlerConfig{
		Amazon:      "test-amazon-crawler",
		NumCrawlers: 1,
	}
	mockProducts := []entities.Product{
		{
			Description: "test-product-1",
			MaxPrice:    1000,
			CrawlerName: "amazon",
		},
		{
			Description: "test-product-2",
			MaxPrice:    1200,
			CrawlerName: "amazon",
		},
	}

	mockLogger.On("Info", mock.Anything).Return(nil)
	mockLogger.On("AddFields", mock.Anything).Return(nil)
	mockCrawler.On("SetupCrawlerEnv", mock.Anything, mock.Anything).Return(nil)
	mockCrawler.On("RunCrawler", mock.Anything, mockProducts[0]).Return("Product(price=1000, original_price=1500, discount=None, link='http://test-link-1.com')", nil).Once()
	mockCrawler.On("RunCrawler", mock.Anything, mockProducts[1]).Return("Product(price=1500, original_price=1500, discount=None, link='http://test-link-2.com')", nil).Once()
	mockProductNotificationSvc.On("Execute", mock.Anything, mock.Anything).Return(nil)

	crawlerService := NewCrawlerService(parser, &cfg, mockProductNotificationSvc, mockLogger, mockCrawler)
	err := crawlerService.Execute(mockProducts)

	require.NoError(t, err)
	mockProductNotificationSvc.AssertCalled(t, "Execute", mock.Anything, mock.Anything)
	mockCrawler.AssertCalled(t, "SetupCrawlerEnv", mock.Anything, mock.Anything)
	mockCrawler.AssertCalled(t, "RunCrawler", mock.Anything, mockProducts[0])
	mockCrawler.AssertCalled(t, "RunCrawler", mock.Anything, mockProducts[1])
}

func TestCrawlerServiceWithCrawlerEnvError(t *testing.T) {
	mockLogger := mocks.NewLoggerContract(t)
	mockCrawler := mocks.NewCrawler(t)
	mockProductNotificationSvc := mocks.NewProductNotificationService(t)

	parser := crawlerparser.NewResultParser()
	cfg := config.CrawlerConfig{
		Amazon:      "test-amazon-crawler",
		NumCrawlers: 1,
	}
	mockProducts := []entities.Product{
		{
			Description: "test-product-1",
			MaxPrice:    1000,
			CrawlerName: "amazon",
		},
	}

	mockLogger.On("Info", mock.Anything).Return(nil)
	mockLogger.On("Error", mock.Anything).Return(nil)
	mockLogger.On("AddFields", mock.Anything).Return(nil)
	mockCrawler.On("SetupCrawlerEnv", mock.Anything, mock.Anything).Return(errors.New("Env setup error"))

	crawlerService := NewCrawlerService(parser, &cfg, mockProductNotificationSvc, mockLogger, mockCrawler)
	err := crawlerService.Execute(mockProducts)

	require.NoError(t, err)
	mockProductNotificationSvc.AssertNotCalled(t, "Execute", mock.Anything, mock.Anything)
	mockCrawler.AssertCalled(t, "SetupCrawlerEnv", mock.Anything, mock.Anything)
	mockCrawler.AssertNotCalled(t, "RunCrawler", mock.Anything, mockProducts[0])
}

func TestCrawlerServiceWithCrawlerRunError(t *testing.T) {
	mockLogger := mocks.NewLoggerContract(t)
	mockCrawler := mocks.NewCrawler(t)
	mockProductNotificationSvc := mocks.NewProductNotificationService(t)

	parser := crawlerparser.NewResultParser()
	cfg := config.CrawlerConfig{
		Amazon:      "test-amazon-crawler",
		NumCrawlers: 1,
	}
	mockProducts := []entities.Product{
		{
			Description: "test-product-1",
			MaxPrice:    1000,
			CrawlerName: "amazon",
		},
	}

	mockLogger.On("Info", mock.Anything).Return(nil)
	mockLogger.On("Error", mock.Anything).Return(nil)
	mockLogger.On("AddFields", mock.Anything).Return(nil)
	mockCrawler.On("SetupCrawlerEnv", mock.Anything, mock.Anything).Return(nil)
	mockCrawler.On("RunCrawler", mock.Anything, mockProducts[0]).Return("", errors.New("Crawler run error")).Once()

	crawlerService := NewCrawlerService(parser, &cfg, mockProductNotificationSvc, mockLogger, mockCrawler)
	err := crawlerService.Execute(mockProducts)

	require.NoError(t, err)
	mockProductNotificationSvc.AssertNotCalled(t, "Execute", mock.Anything, mock.Anything)
	mockCrawler.AssertCalled(t, "SetupCrawlerEnv", mock.Anything, mock.Anything)
	mockCrawler.AssertCalled(t, "RunCrawler", mock.Anything, mockProducts[0])
}
