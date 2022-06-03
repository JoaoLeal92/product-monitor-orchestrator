package entities

import (
	"path/filepath"
	"testing"

	"github.com/JoaoLeal92/product-monitor-orchestrator/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPriceValidation(t *testing.T) {
	type testScenarios struct {
		product        Product
		testPrice      int
		expectedResult bool
	}

	tests := map[string]testScenarios{
		"valid-price": {
			Product{MaxPrice: 1000},
			999,
			true,
		},
		"valid-equal-price": {
			Product{MaxPrice: 1000},
			1000,
			true,
		},
		"invalid-higher-price": {
			Product{MaxPrice: 999},
			1000,
			false,
		},
		"invalid-zero-price": {
			Product{MaxPrice: 1000},
			0,
			false,
		},
	}

	for testName, testData := range tests {
		t.Run(testName, func(t *testing.T) {
			got := testData.product.IsBelowMaxPrice(testData.testPrice)
			assert.Equal(t, testData.expectedResult, got)
		})
	}
}

func TestCrawlerPath(t *testing.T) {
	type testScenarios struct {
		product        Product
		cfg            config.CrawlerConfig
		expectedResult string
		hasError       bool
	}

	validFilePath, _ := filepath.Abs("./test-amazon-crawler-path")
	tests := map[string]testScenarios{
		"valid-crawler": {
			Product{CrawlerName: "amazon"},
			config.CrawlerConfig{Amazon: "./test-amazon-crawler-path"},
			validFilePath,
			false,
		},
		"invalid-crawler": {
			Product{CrawlerName: "invalid-crawler"},
			config.CrawlerConfig{},
			"",
			true,
		},
	}

	for testName, testData := range tests {
		t.Run(testName, func(t *testing.T) {
			got, err := testData.product.GetCrawlerPath(&testData.cfg)
			if testData.hasError {
				require.Error(t, err)
			}
			assert.Equal(t, testData.expectedResult, got)
		})
	}
}
