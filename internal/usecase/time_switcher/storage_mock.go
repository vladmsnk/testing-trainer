// Code generated by mockery v2.46.3. DO NOT EDIT.

package time_switcher

import (
	context "context"
	entities "testing_trainer/internal/entities"

	mock "github.com/stretchr/testify/mock"

	time "time"
)

// MockStorage is an autogenerated mock type for the Storage type
type MockStorage struct {
	mock.Mock
}

// CreateProgress provides a mock function with given fields: ctx, progress
func (_m *MockStorage) CreateProgress(ctx context.Context, progress entities.Progress) (int64, error) {
	ret := _m.Called(ctx, progress)

	if len(ret) == 0 {
		panic("no return value specified for CreateProgress")
	}

	var r0 int64
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, entities.Progress) (int64, error)); ok {
		return rf(ctx, progress)
	}
	if rf, ok := ret.Get(0).(func(context.Context, entities.Progress) int64); ok {
		r0 = rf(ctx, progress)
	} else {
		r0 = ret.Get(0).(int64)
	}

	if rf, ok := ret.Get(1).(func(context.Context, entities.Progress) error); ok {
		r1 = rf(ctx, progress)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// CreateProgressSnapshot provides a mock function with given fields: ctx, snapshot
func (_m *MockStorage) CreateProgressSnapshot(ctx context.Context, snapshot entities.ProgressSnapshot) error {
	ret := _m.Called(ctx, snapshot)

	if len(ret) == 0 {
		panic("no return value specified for CreateProgressSnapshot")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, entities.ProgressSnapshot) error); ok {
		r0 = rf(ctx, snapshot)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetProgressesForAllGoals provides a mock function with given fields: ctx, username, progressIDs
func (_m *MockStorage) GetProgressesForAllGoals(ctx context.Context, username string, progressIDs []int64) ([]entities.Progress, error) {
	ret := _m.Called(ctx, username, progressIDs)

	if len(ret) == 0 {
		panic("no return value specified for GetProgressesForAllGoals")
	}

	var r0 []entities.Progress
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, []int64) ([]entities.Progress, error)); ok {
		return rf(ctx, username, progressIDs)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, []int64) []entities.Progress); ok {
		r0 = rf(ctx, username, progressIDs)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]entities.Progress)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, []int64) error); ok {
		r1 = rf(ctx, username, progressIDs)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetSnapshotForTheTime provides a mock function with given fields: ctx, username, _a2
func (_m *MockStorage) GetSnapshotForTheTime(ctx context.Context, username string, _a2 time.Time) (entities.ProgressSnapshot, error) {
	ret := _m.Called(ctx, username, _a2)

	if len(ret) == 0 {
		panic("no return value specified for GetSnapshotForTheTime")
	}

	var r0 entities.ProgressSnapshot
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, time.Time) (entities.ProgressSnapshot, error)); ok {
		return rf(ctx, username, _a2)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, time.Time) entities.ProgressSnapshot); ok {
		r0 = rf(ctx, username, _a2)
	} else {
		r0 = ret.Get(0).(entities.ProgressSnapshot)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, time.Time) error); ok {
		r1 = rf(ctx, username, _a2)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewMockStorage creates a new instance of MockStorage. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockStorage(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockStorage {
	mock := &MockStorage{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
