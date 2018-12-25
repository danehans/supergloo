// Code generated by MockGen. DO NOT EDIT.
// Source: pkg/kube/crd.go

// Package mock_kube is a generated GoMock package.
package mock_kube

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	v1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
)

// MockCrdClient is a mock of CrdClient interface
type MockCrdClient struct {
	ctrl     *gomock.Controller
	recorder *MockCrdClientMockRecorder
}

// MockCrdClientMockRecorder is the mock recorder for MockCrdClient
type MockCrdClientMockRecorder struct {
	mock *MockCrdClient
}

// NewMockCrdClient creates a new mock instance
func NewMockCrdClient(ctrl *gomock.Controller) *MockCrdClient {
	mock := &MockCrdClient{ctrl: ctrl}
	mock.recorder = &MockCrdClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockCrdClient) EXPECT() *MockCrdClientMockRecorder {
	return m.recorder
}

// CreateCrds mocks base method
func (m *MockCrdClient) CreateCrds(crds ...*v1beta1.CustomResourceDefinition) error {
	m.ctrl.T.Helper()
	varargs := []interface{}{}
	for _, a := range crds {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "CreateCrds", varargs...)
	ret0, _ := ret[0].(error)
	return ret0
}

// CreateCrds indicates an expected call of CreateCrds
func (mr *MockCrdClientMockRecorder) CreateCrds(crds ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateCrds", reflect.TypeOf((*MockCrdClient)(nil).CreateCrds), crds...)
}

// DeleteCrds mocks base method
func (m *MockCrdClient) DeleteCrds(crdNames ...string) error {
	m.ctrl.T.Helper()
	varargs := []interface{}{}
	for _, a := range crdNames {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "DeleteCrds", varargs...)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteCrds indicates an expected call of DeleteCrds
func (mr *MockCrdClientMockRecorder) DeleteCrds(crdNames ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteCrds", reflect.TypeOf((*MockCrdClient)(nil).DeleteCrds), crdNames...)
}