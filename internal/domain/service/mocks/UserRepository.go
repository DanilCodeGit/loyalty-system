// Code generated by mockery v2.28.2. DO NOT EDIT.

package mocks

import (
	context "context"
	entity "loyalty/internal/domain/entity"

	mock "github.com/stretchr/testify/mock"

	uuid "github.com/google/uuid"
)

// UserRepository is an autogenerated mock type for the UserRepository type
type UserRepository struct {
	mock.Mock
}

// BeginTx provides a mock function with given fields: ctx
func (_m *UserRepository) BeginTx(ctx context.Context) (context.Context, error) {
	ret := _m.Called(ctx)

	var r0 context.Context
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context) (context.Context, error)); ok {
		return rf(ctx)
	}
	if rf, ok := ret.Get(0).(func(context.Context) context.Context); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(context.Context)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// CommitTx provides a mock function with given fields: ctx
func (_m *UserRepository) CommitTx(ctx context.Context) error {
	ret := _m.Called(ctx)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context) error); ok {
		r0 = rf(ctx)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// CreateUser provides a mock function with given fields: ctx, user
func (_m *UserRepository) CreateUser(ctx context.Context, user *entity.User) (*entity.User, error) {
	ret := _m.Called(ctx, user)

	var r0 *entity.User
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *entity.User) (*entity.User, error)); ok {
		return rf(ctx, user)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *entity.User) *entity.User); ok {
		r0 = rf(ctx, user)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*entity.User)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *entity.User) error); ok {
		r1 = rf(ctx, user)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetUserByLogin provides a mock function with given fields: ctx, login
func (_m *UserRepository) GetUserByLogin(ctx context.Context, login string) (*entity.User, error) {
	ret := _m.Called(ctx, login)

	var r0 *entity.User
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (*entity.User, error)); ok {
		return rf(ctx, login)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) *entity.User); ok {
		r0 = rf(ctx, login)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*entity.User)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, login)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetUserByUUID provides a mock function with given fields: ctx, userUUID
func (_m *UserRepository) GetUserByUUID(ctx context.Context, userUUID uuid.UUID) (*entity.User, error) {
	ret := _m.Called(ctx, userUUID)

	var r0 *entity.User
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) (*entity.User, error)); ok {
		return rf(ctx, userUUID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) *entity.User); ok {
		r0 = rf(ctx, userUUID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*entity.User)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, uuid.UUID) error); ok {
		r1 = rf(ctx, userUUID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// RollbackTx provides a mock function with given fields: ctx
func (_m *UserRepository) RollbackTx(ctx context.Context) error {
	ret := _m.Called(ctx)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context) error); ok {
		r0 = rf(ctx)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

type mockConstructorTestingTNewUserRepository interface {
	mock.TestingT
	Cleanup(func())
}

// NewUserRepository creates a new instance of UserRepository. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewUserRepository(t mockConstructorTestingTNewUserRepository) *UserRepository {
	mock := &UserRepository{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
