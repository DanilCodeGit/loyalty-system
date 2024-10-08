// Code generated by mockery v2.28.2. DO NOT EDIT.

package mocks

import (
	context "context"

	mock "github.com/stretchr/testify/mock"

	uuid "github.com/google/uuid"
)

// OrderAdder is an autogenerated mock type for the OrderAdder type
type OrderAdder struct {
	mock.Mock
}

// Add provides a mock function with given fields: ctx, orderNumber, userUUID
func (_m *OrderAdder) Add(ctx context.Context, orderNumber int, userUUID uuid.UUID) error {
	ret := _m.Called(ctx, orderNumber, userUUID)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, int, uuid.UUID) error); ok {
		r0 = rf(ctx, orderNumber, userUUID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

type mockConstructorTestingTNewOrderAdder interface {
	mock.TestingT
	Cleanup(func())
}

// NewOrderAdder creates a new instance of OrderAdder. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewOrderAdder(t mockConstructorTestingTNewOrderAdder) *OrderAdder {
	mock := &OrderAdder{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
