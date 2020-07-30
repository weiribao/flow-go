// Code generated by mockery v1.0.0. DO NOT EDIT.

package mock

import (
	flow "github.com/dapperlabs/flow-go/model/flow"
	mock "github.com/stretchr/testify/mock"
)

// PingMetrics is an autogenerated mock type for the PingMetrics type
type PingMetrics struct {
	mock.Mock
}

// NodeReachable provides a mock function with given fields: nodeID, reachable
func (_m *PingMetrics) NodeReachable(nodeID flow.Identifier, reachable bool) {
	_m.Called(nodeID, reachable)
}
