package entities

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsPriceValid(t *testing.T) {
	tests := map[string]struct {
		productSearchResult ProductSearchResult
		expectedResult      bool
	}{
		"valid-price": {
			ProductSearchResult{
				Price: 1000,
			},
			true,
		},
		"invalid-price": {
			ProductSearchResult{
				Price: 0,
			},
			false,
		},
		"invalid-negative-price": {
			ProductSearchResult{
				Price: -100,
			},
			false,
		},
	}

	for testName, testData := range tests {
		t.Run(testName, func(t *testing.T) {
			result := testData.productSearchResult.IsPriceValid()
			assert.Equal(t, testData.expectedResult, result)
		})
	}
}
