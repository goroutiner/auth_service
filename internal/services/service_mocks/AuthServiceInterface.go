// Code generated by mockery v2.53.3. DO NOT EDIT.

package service_mocks

import (
	entities "auth_service/internal/entities"

	mock "github.com/stretchr/testify/mock"
)

// AuthServiceInterface is an autogenerated mock type for the AuthServiceInterface type
type AuthServiceInterface struct {
	mock.Mock
}

// GenerateTokens provides a mock function with given fields: userId, ip
func (_m *AuthServiceInterface) GenerateTokens(userId string, ip string) (*entities.TokensPair, error) {
	ret := _m.Called(userId, ip)

	if len(ret) == 0 {
		panic("no return value specified for GenerateTokens")
	}

	var r0 *entities.TokensPair
	var r1 error
	if rf, ok := ret.Get(0).(func(string, string) (*entities.TokensPair, error)); ok {
		return rf(userId, ip)
	}
	if rf, ok := ret.Get(0).(func(string, string) *entities.TokensPair); ok {
		r0 = rf(userId, ip)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*entities.TokensPair)
		}
	}

	if rf, ok := ret.Get(1).(func(string, string) error); ok {
		r1 = rf(userId, ip)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// RefreshTokens provides a mock function with given fields: ip, tokensPair
func (_m *AuthServiceInterface) RefreshTokens(ip string, tokensPair *entities.TokensPair) (*entities.TokensPair, error) {
	ret := _m.Called(ip, tokensPair)

	if len(ret) == 0 {
		panic("no return value specified for RefreshTokens")
	}

	var r0 *entities.TokensPair
	var r1 error
	if rf, ok := ret.Get(0).(func(string, *entities.TokensPair) (*entities.TokensPair, error)); ok {
		return rf(ip, tokensPair)
	}
	if rf, ok := ret.Get(0).(func(string, *entities.TokensPair) *entities.TokensPair); ok {
		r0 = rf(ip, tokensPair)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*entities.TokensPair)
		}
	}

	if rf, ok := ret.Get(1).(func(string, *entities.TokensPair) error); ok {
		r1 = rf(ip, tokensPair)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewAuthServiceInterface creates a new instance of AuthServiceInterface. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewAuthServiceInterface(t interface {
	mock.TestingT
	Cleanup(func())
}) *AuthServiceInterface {
	mock := &AuthServiceInterface{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
