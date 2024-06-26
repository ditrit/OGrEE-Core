// Code generated by mockery v2.20.0. DO NOT EDIT.

package mocks

import (
	readline "cli/readline"

	mock "github.com/stretchr/testify/mock"
)

// Ogree3DPort is an autogenerated mock type for the Ogree3DPort type
type Ogree3DPort struct {
	mock.Mock
}

// Connect provides a mock function with given fields: url, rl
func (_m *Ogree3DPort) Connect(url string, rl *readline.Instance) error {
	ret := _m.Called(url, rl)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, *readline.Instance) error); ok {
		r0 = rf(url, rl)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Disconnect provides a mock function with given fields:
func (_m *Ogree3DPort) Disconnect() {
	_m.Called()
}

// Inform provides a mock function with given fields: caller, entity, data
func (_m *Ogree3DPort) Inform(caller string, entity int, data map[string]interface{}) error {
	ret := _m.Called(caller, entity, data)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, int, map[string]interface{}) error); ok {
		r0 = rf(caller, entity, data)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// InformOptional provides a mock function with given fields: caller, entity, data
func (_m *Ogree3DPort) InformOptional(caller string, entity int, data map[string]interface{}) error {
	ret := _m.Called(caller, entity, data)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, int, map[string]interface{}) error); ok {
		r0 = rf(caller, entity, data)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// IsConnected provides a mock function with given fields:
func (_m *Ogree3DPort) IsConnected() bool {
	ret := _m.Called()

	var r0 bool
	if rf, ok := ret.Get(0).(func() bool); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// SetDefaultURL provides a mock function with given fields:
func (_m *Ogree3DPort) SetDefaultURL() {
	_m.Called()
}

// SetURL provides a mock function with given fields: url
func (_m *Ogree3DPort) SetURL(url string) error {
	ret := _m.Called(url)

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(url)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// URL provides a mock function with given fields:
func (_m *Ogree3DPort) URL() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

type mockConstructorTestingTNewOgree3DPort interface {
	mock.TestingT
	Cleanup(func())
}

// NewOgree3DPort creates a new instance of Ogree3DPort. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewOgree3DPort(t mockConstructorTestingTNewOgree3DPort) *Ogree3DPort {
	mock := &Ogree3DPort{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
