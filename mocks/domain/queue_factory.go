// Code generated by mockery v2.42.1. DO NOT EDIT.

package mocks

import (
	domain "github.com/kukwuka/queue/internal/domain"
	mock "github.com/stretchr/testify/mock"
)

// QueueFactory is an autogenerated mock type for the QueueFactory type
type QueueFactory struct {
	mock.Mock
}

type QueueFactory_Expecter struct {
	mock *mock.Mock
}

func (_m *QueueFactory) EXPECT() *QueueFactory_Expecter {
	return &QueueFactory_Expecter{mock: &_m.Mock}
}

// Execute provides a mock function with given fields: maxLen
func (_m *QueueFactory) Execute(maxLen int) domain.Queue {
	ret := _m.Called(maxLen)

	if len(ret) == 0 {
		panic("no return value specified for Execute")
	}

	var r0 domain.Queue
	if rf, ok := ret.Get(0).(func(int) domain.Queue); ok {
		r0 = rf(maxLen)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(domain.Queue)
		}
	}

	return r0
}

// QueueFactory_Execute_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Execute'
type QueueFactory_Execute_Call struct {
	*mock.Call
}

// Execute is a helper method to define mock.On call
//   - maxLen int
func (_e *QueueFactory_Expecter) Execute(maxLen interface{}) *QueueFactory_Execute_Call {
	return &QueueFactory_Execute_Call{Call: _e.mock.On("Execute", maxLen)}
}

func (_c *QueueFactory_Execute_Call) Run(run func(maxLen int)) *QueueFactory_Execute_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(int))
	})
	return _c
}

func (_c *QueueFactory_Execute_Call) Return(_a0 domain.Queue) *QueueFactory_Execute_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *QueueFactory_Execute_Call) RunAndReturn(run func(int) domain.Queue) *QueueFactory_Execute_Call {
	_c.Call.Return(run)
	return _c
}

// NewQueueFactory creates a new instance of QueueFactory. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewQueueFactory(t interface {
	mock.TestingT
	Cleanup(func())
}) *QueueFactory {
	mock := &QueueFactory{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
