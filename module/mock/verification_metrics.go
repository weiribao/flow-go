// Code generated by mockery v1.0.0. DO NOT EDIT.

package mock

import mock "github.com/stretchr/testify/mock"

// VerificationMetrics is an autogenerated mock type for the VerificationMetrics type
type VerificationMetrics struct {
	mock.Mock
}

// LogVerifiableChunkSize provides a mock function with given fields: size
func (_m *VerificationMetrics) LogVerifiableChunkSize(size float64) {
	_m.Called(size)
}

// OnChunkDataPackReceived provides a mock function with given fields:
func (_m *VerificationMetrics) OnChunkDataPackReceived() {
	_m.Called()
}

// OnChunkDataPackRequested provides a mock function with given fields:
func (_m *VerificationMetrics) OnChunkDataPackRequested() {
	_m.Called()
}

// OnExecutionReceiptReceived provides a mock function with given fields:
func (_m *VerificationMetrics) OnExecutionReceiptReceived() {
	_m.Called()
}

// OnExecutionResultReceived provides a mock function with given fields:
func (_m *VerificationMetrics) OnExecutionResultReceived() {
	_m.Called()
}

// OnExecutionResultSent provides a mock function with given fields:
func (_m *VerificationMetrics) OnExecutionResultSent() {
	_m.Called()
}

// OnResultApproval provides a mock function with given fields:
func (_m *VerificationMetrics) OnResultApproval() {
	_m.Called()
}

// OnVerifiableChunkReceived provides a mock function with given fields:
func (_m *VerificationMetrics) OnVerifiableChunkReceived() {
	_m.Called()
}

// OnVerifiableChunkSent provides a mock function with given fields:
func (_m *VerificationMetrics) OnVerifiableChunkSent() {
	_m.Called()
}
