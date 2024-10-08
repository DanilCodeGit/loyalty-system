// Code generated by mockery v2.28.2. DO NOT EDIT.

package mocks

import (
	context "context"

	mock "github.com/stretchr/testify/mock"
	entity "loyalty/internal/domain/entity"
)

// UserAuthorizer is an autogenerated mock type for the UserAuthorizer type
type UserAuthorizer struct {
	mock.Mock
}

// Authorize provides a mock function with given fields: ctx, token
func (_m *UserAuthorizer) Authorize(ctx context.Context, token string) (*entity.User, error) {
	ret := _m.Called(ctx, token)

	var r0 *entity.User
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (*entity.User, error)); ok {
		return rf(ctx, token)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) *entity.User); ok {
		r0 = rf(ctx, token)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*entity.User)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, token)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type mockConstructorTestingTNewUserAuthorizer interface {
	mock.TestingT
	Cleanup(func())
}

// NewUserAuthorizer creates a new instance of UserAuthorizer. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewUserAuthorizer(t mockConstructorTestingTNewUserAuthorizer) *UserAuthorizer {
	mock := &UserAuthorizer{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
