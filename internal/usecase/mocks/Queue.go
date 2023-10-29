// Code generated by mockery v2.32.0. DO NOT EDIT.

package mocks

import (
	entity "github.com/antonmisa/cliurlfetcher/internal/entity"
	mock "github.com/stretchr/testify/mock"
)

// Queue is an autogenerated mock type for the Queue type
type Queue struct {
	mock.Mock
}

// Close provides a mock function with given fields:
func (_m *Queue) Close() {
	_m.Called()
}

// Pop provides a mock function with given fields:
func (_m *Queue) Pop() (entity.Task, bool) {
	ret := _m.Called()

	var r0 entity.Task
	var r1 bool
	if rf, ok := ret.Get(0).(func() (entity.Task, bool)); ok {
		return rf()
	}
	if rf, ok := ret.Get(0).(func() entity.Task); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(entity.Task)
	}

	if rf, ok := ret.Get(1).(func() bool); ok {
		r1 = rf()
	} else {
		r1 = ret.Get(1).(bool)
	}

	return r0, r1
}

// Push provides a mock function with given fields: _a0
func (_m *Queue) Push(_a0 entity.Task) error {
	ret := _m.Called(_a0)

	var r0 error
	if rf, ok := ret.Get(0).(func(entity.Task) error); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// NewQueue creates a new instance of Queue. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewQueue(t interface {
	mock.TestingT
	Cleanup(func())
}) *Queue {
	mock := &Queue{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}