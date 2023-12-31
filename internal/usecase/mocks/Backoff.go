// Code generated by mockery v2.32.0. DO NOT EDIT.

package mocks

import (
	http "net/http"
	time "time"

	mock "github.com/stretchr/testify/mock"
)

// Backoff is an autogenerated mock type for the Backoff type
type Backoff struct {
	mock.Mock
}

// Execute provides a mock function with given fields: min, max, attemptNum, resp
func (_m *Backoff) Execute(min time.Duration, max time.Duration, attemptNum int, resp *http.Response) time.Duration {
	ret := _m.Called(min, max, attemptNum, resp)

	var r0 time.Duration
	if rf, ok := ret.Get(0).(func(time.Duration, time.Duration, int, *http.Response) time.Duration); ok {
		r0 = rf(min, max, attemptNum, resp)
	} else {
		r0 = ret.Get(0).(time.Duration)
	}

	return r0
}

// NewBackoff creates a new instance of Backoff. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewBackoff(t interface {
	mock.TestingT
	Cleanup(func())
}) *Backoff {
	mock := &Backoff{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
