// Code generated by mockery v2.10.0. DO NOT EDIT.

package mocks

import (
	context "context"

	mock "github.com/stretchr/testify/mock"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v1alpha1 "github.com/kyma-project/kyma/components/compass-runtime-agent/pkg/apis/compass/v1alpha1"
)

// CRManager is an autogenerated mock type for the CRManager type
type CRManager struct {
	mock.Mock
}

// Create provides a mock function with given fields: ctx, cc, options
func (_m *CRManager) Create(ctx context.Context, cc *v1alpha1.CompassConnection, options v1.CreateOptions) (*v1alpha1.CompassConnection, error) {
	ret := _m.Called(ctx, cc, options)

	var r0 *v1alpha1.CompassConnection
	if rf, ok := ret.Get(0).(func(context.Context, *v1alpha1.CompassConnection, v1.CreateOptions) *v1alpha1.CompassConnection); ok {
		r0 = rf(ctx, cc, options)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*v1alpha1.CompassConnection)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *v1alpha1.CompassConnection, v1.CreateOptions) error); ok {
		r1 = rf(ctx, cc, options)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Delete provides a mock function with given fields: ctx, name, options
func (_m *CRManager) Delete(ctx context.Context, name string, options v1.DeleteOptions) error {
	ret := _m.Called(ctx, name, options)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, v1.DeleteOptions) error); ok {
		r0 = rf(ctx, name, options)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Get provides a mock function with given fields: ctx, name, options
func (_m *CRManager) Get(ctx context.Context, name string, options v1.GetOptions) (*v1alpha1.CompassConnection, error) {
	ret := _m.Called(ctx, name, options)

	var r0 *v1alpha1.CompassConnection
	if rf, ok := ret.Get(0).(func(context.Context, string, v1.GetOptions) *v1alpha1.CompassConnection); ok {
		r0 = rf(ctx, name, options)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*v1alpha1.CompassConnection)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, v1.GetOptions) error); ok {
		r1 = rf(ctx, name, options)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Update provides a mock function with given fields: ctx, cc, options
func (_m *CRManager) Update(ctx context.Context, cc *v1alpha1.CompassConnection, options v1.UpdateOptions) (*v1alpha1.CompassConnection, error) {
	ret := _m.Called(ctx, cc, options)

	var r0 *v1alpha1.CompassConnection
	if rf, ok := ret.Get(0).(func(context.Context, *v1alpha1.CompassConnection, v1.UpdateOptions) *v1alpha1.CompassConnection); ok {
		r0 = rf(ctx, cc, options)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*v1alpha1.CompassConnection)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *v1alpha1.CompassConnection, v1.UpdateOptions) error); ok {
		r1 = rf(ctx, cc, options)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}