package services

import (
	"errors"
	"testing"

	mocks "github.com/JoaoLeal92/product-monitor-orchestrator/contracts/mocks"
	"github.com/JoaoLeal92/product-monitor-orchestrator/entities"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestProductNotificationService(t *testing.T) {
	mockRepoManager := mocks.NewRepoManager(t)
	mockProductSearcHistoryRepo := mocks.NewProductSearchHistoryRepository(t)
	mockQueueManager := mocks.NewQueueManager(t)
	mockLogger := mocks.NewLoggerContract(t)

	mockRepoManager.On("ProductSearchHistory").Return(mockProductSearcHistoryRepo).Twice()
	mockProductSearcHistoryRepo.On("InsertNewHistory", mock.Anything).Return(nil)
	mockQueueManager.On("SendMessage", mock.Anything).Return(nil)
	mockProductSearcHistoryRepo.On("GetHistoryByProductID", mock.Anything).Return([]entities.ProductSearchResult{
		{
			Price: 100000,
		},
		{
			Price: 100100,
		},
	}, nil)

	productNotificationService := NewProductNotificationService(mockRepoManager, mockLogger, mockQueueManager)
	productSearchResultStub := entities.ProductSearchResult{
		Price: 999,
	}
	product := entities.Product{
		Description: "test-product",
		MaxPrice:    1000,
	}
	expectedProductNotification := entities.ProductNotification{
		Description: "test-product",
		Price:       999,
		AvgPrice:    1000,
		AvgDiscount: "0.99",
		UserID:      product.UserID.String(),
	}
	err := productNotificationService.Execute(&product, &productSearchResultStub)

	require.NoError(t, err)
	mockRepoManager.AssertNumberOfCalls(t, "ProductSearchHistory", 2)
	mockProductSearcHistoryRepo.AssertCalled(t, "InsertNewHistory", mock.Anything)
	mockProductSearcHistoryRepo.AssertCalled(t, "GetHistoryByProductID", mock.Anything)
	mockQueueManager.AssertCalled(t, "SendMessage", expectedProductNotification)
}

func TestInvalidProductSearchResult(t *testing.T) {
	mockRepoManager := mocks.NewRepoManager(t)
	mockProductSearcHistoryRepo := mocks.NewProductSearchHistoryRepository(t)
	mockQueueManager := mocks.NewQueueManager(t)
	mockLogger := mocks.NewLoggerContract(t)
	mockLogger.On("Info", mock.Anything).Return(nil)

	productNotificationService := NewProductNotificationService(mockRepoManager, mockLogger, mockQueueManager)
	productSearchResultStub := entities.ProductSearchResult{}
	product := entities.Product{}
	err := productNotificationService.Execute(&product, &productSearchResultStub)

	require.Error(t, err)
	assert.Equal(t, err.Error(), "invalid price result (<0)")
	mockRepoManager.AssertNotCalled(t, "ProductSearchHistory")
	mockProductSearcHistoryRepo.AssertNotCalled(t, "InsertNewHistory", mock.Anything)
	mockQueueManager.AssertNotCalled(t, "SendMessage", mock.Anything)
}

func TestErrorOnDbInsert(t *testing.T) {
	mockRepoManager := mocks.NewRepoManager(t)
	mockProductSearcHistoryRepo := mocks.NewProductSearchHistoryRepository(t)
	mockQueueManager := mocks.NewQueueManager(t)
	mockLogger := mocks.NewLoggerContract(t)

	mockRepoManager.On("ProductSearchHistory").Return(mockProductSearcHistoryRepo)
	mockProductSearcHistoryRepo.On("InsertNewHistory", mock.Anything).Return(errors.New("db error"))

	productNotificationService := NewProductNotificationService(mockRepoManager, mockLogger, mockQueueManager)
	productSearchResultStub := entities.ProductSearchResult{
		Price: 999,
	}
	product := entities.Product{}
	err := productNotificationService.Execute(&product, &productSearchResultStub)

	require.Error(t, err)
	assert.Equal(t, err.Error(), "db error")
	mockRepoManager.AssertCalled(t, "ProductSearchHistory")
	mockProductSearcHistoryRepo.AssertCalled(t, "InsertNewHistory", &productSearchResultStub)
	mockQueueManager.AssertNotCalled(t, "SendMessage", mock.Anything)
}

func TestProductWithPriceAboveMaxPrice(t *testing.T) {
	mockRepoManager := mocks.NewRepoManager(t)
	mockProductSearcHistoryRepo := mocks.NewProductSearchHistoryRepository(t)
	mockQueueManager := mocks.NewQueueManager(t)
	mockLogger := mocks.NewLoggerContract(t)

	mockRepoManager.On("ProductSearchHistory").Return(mockProductSearcHistoryRepo)
	mockProductSearcHistoryRepo.On("InsertNewHistory", mock.Anything).Return(nil)
	mockLogger.On("Info", mock.Anything).Return(nil)

	productNotificationService := NewProductNotificationService(mockRepoManager, mockLogger, mockQueueManager)
	productSearchResultStub := entities.ProductSearchResult{
		Price: 1001,
	}
	product := entities.Product{
		Description: "test-product",
		MaxPrice:    1000,
	}
	err := productNotificationService.Execute(&product, &productSearchResultStub)

	require.NoError(t, err)
	mockRepoManager.AssertCalled(t, "ProductSearchHistory")
	mockProductSearcHistoryRepo.AssertCalled(t, "InsertNewHistory", &productSearchResultStub)
	mockQueueManager.AssertNotCalled(t, "SendMessage", mock.Anything)
}
