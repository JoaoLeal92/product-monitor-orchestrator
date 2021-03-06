// Code generated by mockery v2.12.3. DO NOT EDIT.

package mocks

import (
	entities "github.com/JoaoLeal92/product-monitor-orchestrator/entities"
	mock "github.com/stretchr/testify/mock"
)

// Crawler is an autogenerated mock type for the Crawler type
type Crawler struct {
	mock.Mock
}

// RunCrawler provides a mock function with given fields: crawlerPath, product
func (_m *Crawler) RunCrawler(crawlerPath string, product entities.Product) (string, error) {
	ret := _m.Called(crawlerPath, product)

	var r0 string
	if rf, ok := ret.Get(0).(func(string, entities.Product) string); ok {
		r0 = rf(crawlerPath, product)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, entities.Product) error); ok {
		r1 = rf(crawlerPath, product)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// SetupCrawlerEnv provides a mock function with given fields: crawlerPath, productID
func (_m *Crawler) SetupCrawlerEnv(crawlerPath string, productID string) error {
	ret := _m.Called(crawlerPath, productID)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string) error); ok {
		r0 = rf(crawlerPath, productID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

type NewCrawlerT interface {
	mock.TestingT
	Cleanup(func())
}

// NewCrawler creates a new instance of Crawler. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewCrawler(t NewCrawlerT) *Crawler {
	mock := &Crawler{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
