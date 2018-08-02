// Code generated by mockery v1.0.0. DO NOT EDIT.
package mocks

import mock "github.com/stretchr/testify/mock"
import shopify "github.com/Shopify/themekit/src/shopify"

// ShopifyClient is an autogenerated mock type for the ShopifyClient type
type ShopifyClient struct {
	mock.Mock
}

// CreateNewTheme provides a mock function with given fields: _a0, _a1
func (_m *ShopifyClient) Themes() ([]shopify.Theme, error) {
	ret := _m.Called()

	var r0 []shopify.Theme
	if rf, ok := ret.Get(0).(func() []shopify.Theme); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]shopify.Theme)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// CreateNewTheme provides a mock function with given fields: _a0, _a1
func (_m *ShopifyClient) CreateNewTheme(_a0 string, _a1 string) (shopify.Theme, error) {
	ret := _m.Called(_a0, _a1)

	var r0 shopify.Theme
	if rf, ok := ret.Get(0).(func(string, string) shopify.Theme); ok {
		r0 = rf(_a0, _a1)
	} else {
		r0 = ret.Get(0).(shopify.Theme)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, string) error); ok {
		r1 = rf(_a0, _a1)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// DeleteAsset provides a mock function with given fields: _a0
func (_m *ShopifyClient) DeleteAsset(_a0 shopify.Asset) error {
	ret := _m.Called(_a0)

	var r0 error
	if rf, ok := ret.Get(0).(func(shopify.Asset) error); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetAllAssets provides a mock function with given fields:
func (_m *ShopifyClient) GetAllAssets() ([]string, error) {
	ret := _m.Called()

	var r0 []string
	if rf, ok := ret.Get(0).(func() []string); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]string)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetAsset provides a mock function with given fields: _a0
func (_m *ShopifyClient) GetAsset(_a0 string) (shopify.Asset, error) {
	ret := _m.Called(_a0)

	var r0 shopify.Asset
	if rf, ok := ret.Get(0).(func(string) shopify.Asset); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Get(0).(shopify.Asset)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetInfo provides a mock function with given fields:
func (_m *ShopifyClient) GetInfo() (shopify.Theme, error) {
	ret := _m.Called()

	var r0 shopify.Theme
	if rf, ok := ret.Get(0).(func() shopify.Theme); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(shopify.Theme)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetShop provides a mock function with given fields:
func (_m *ShopifyClient) GetShop() (shopify.Shop, error) {
	ret := _m.Called()

	var r0 shopify.Shop
	if rf, ok := ret.Get(0).(func() shopify.Shop); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(shopify.Shop)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// UpdateAsset provides a mock function with given fields: _a0
func (_m *ShopifyClient) UpdateAsset(_a0 shopify.Asset) error {
	ret := _m.Called(_a0)

	var r0 error
	if rf, ok := ret.Get(0).(func(shopify.Asset) error); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
