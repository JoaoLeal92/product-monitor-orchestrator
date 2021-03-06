// Code generated by mockery v2.12.3. DO NOT EDIT.

package mocks

import (
	contracts "github.com/JoaoLeal92/product-monitor-orchestrator/contracts"
	mock "github.com/stretchr/testify/mock"
)

// RepoManager is an autogenerated mock type for the RepoManager type
type RepoManager struct {
	mock.Mock
}

// ProductSearchHistory provides a mock function with given fields:
func (_m *RepoManager) ProductSearchHistory() contracts.ProductSearchHistoryRepository {
	ret := _m.Called()

	var r0 contracts.ProductSearchHistoryRepository
	if rf, ok := ret.Get(0).(func() contracts.ProductSearchHistoryRepository); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(contracts.ProductSearchHistoryRepository)
		}
	}

	return r0
}

// Products provides a mock function with given fields:
func (_m *RepoManager) Products() contracts.ProductsRepository {
	ret := _m.Called()

	var r0 contracts.ProductsRepository
	if rf, ok := ret.Get(0).(func() contracts.ProductsRepository); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(contracts.ProductsRepository)
		}
	}

	return r0
}

type NewRepoManagerT interface {
	mock.TestingT
	Cleanup(func())
}

// NewRepoManager creates a new instance of RepoManager. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewRepoManager(t NewRepoManagerT) *RepoManager {
	mock := &RepoManager{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
