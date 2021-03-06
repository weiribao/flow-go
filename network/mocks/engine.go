// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/onflow/flow-go/network (interfaces: Engine)

// Package mocks is a generated GoMock package.
package mocks

import (
	gomock "github.com/golang/mock/gomock"
	flow "github.com/onflow/flow-go/model/flow"
	reflect "reflect"
)

// MockEngine is a mock of Engine interface
type MockEngine struct {
	ctrl     *gomock.Controller
	recorder *MockEngineMockRecorder
}

// MockEngineMockRecorder is the mock recorder for MockEngine
type MockEngineMockRecorder struct {
	mock *MockEngine
}

// NewMockEngine creates a new mock instance
func NewMockEngine(ctrl *gomock.Controller) *MockEngine {
	mock := &MockEngine{ctrl: ctrl}
	mock.recorder = &MockEngineMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockEngine) EXPECT() *MockEngineMockRecorder {
	return m.recorder
}

// Process mocks base method
func (m *MockEngine) Process(arg0 flow.Identifier, arg1 interface{}) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Process", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// Process indicates an expected call of Process
func (mr *MockEngineMockRecorder) Process(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Process", reflect.TypeOf((*MockEngine)(nil).Process), arg0, arg1)
}

// ProcessLocal mocks base method
func (m *MockEngine) ProcessLocal(arg0 interface{}) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ProcessLocal", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// ProcessLocal indicates an expected call of ProcessLocal
func (mr *MockEngineMockRecorder) ProcessLocal(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ProcessLocal", reflect.TypeOf((*MockEngine)(nil).ProcessLocal), arg0)
}

// Submit mocks base method
func (m *MockEngine) Submit(arg0 flow.Identifier, arg1 interface{}) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Submit", arg0, arg1)
}

// Submit indicates an expected call of Submit
func (mr *MockEngineMockRecorder) Submit(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Submit", reflect.TypeOf((*MockEngine)(nil).Submit), arg0, arg1)
}

// SubmitLocal mocks base method
func (m *MockEngine) SubmitLocal(arg0 interface{}) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SubmitLocal", arg0)
}

// SubmitLocal indicates an expected call of SubmitLocal
func (mr *MockEngineMockRecorder) SubmitLocal(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SubmitLocal", reflect.TypeOf((*MockEngine)(nil).SubmitLocal), arg0)
}
