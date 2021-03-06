// Code generated by mockery v1.0.0. DO NOT EDIT.

package mock

import (
	access "github.com/onflow/flow/protobuf/go/flow/access"

	io "io"

	mock "github.com/stretchr/testify/mock"
)

// ConnectionFactory is an autogenerated mock type for the ConnectionFactory type
type ConnectionFactory struct {
	mock.Mock
}

// GetAccessAPIClient provides a mock function with given fields: address
func (_m *ConnectionFactory) GetAccessAPIClient(address string) (access.AccessAPIClient, io.Closer, error) {
	ret := _m.Called(address)

	var r0 access.AccessAPIClient
	if rf, ok := ret.Get(0).(func(string) access.AccessAPIClient); ok {
		r0 = rf(address)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(access.AccessAPIClient)
		}
	}

	var r1 io.Closer
	if rf, ok := ret.Get(1).(func(string) io.Closer); ok {
		r1 = rf(address)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(io.Closer)
		}
	}

	var r2 error
	if rf, ok := ret.Get(2).(func(string) error); ok {
		r2 = rf(address)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}
