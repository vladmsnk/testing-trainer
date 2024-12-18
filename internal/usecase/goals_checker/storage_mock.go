// Code generated by mockery v2.46.3. DO NOT EDIT.

package goals_checker

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

// GetAllGoalsNeedCheck provides a mock function with given fields: ctx, currentTime
func (_m *MockStorage) GetAllGoalsNeedCheck(ctx context.Context, currentTime time.Time) ([]entities.Goal, error) {
	ret := _m.Called(ctx, currentTime)

	if len(ret) == 0 {
		panic("no return value specified for GetAllGoalsNeedCheck")
	}

	var r0 []entities.Goal
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, time.Time) ([]entities.Goal, error)); ok {
		return rf(ctx, currentTime)
	}
	if rf, ok := ret.Get(0).(func(context.Context, time.Time) []entities.Goal); ok {
		r0 = rf(ctx, currentTime)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]entities.Goal)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, time.Time) error); ok {
		r1 = rf(ctx, currentTime)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetCurrentProgress provides a mock function with given fields: ctx, goalId
func (_m *MockStorage) GetCurrentProgress(ctx context.Context, goalId int) (entities.Progress, error) {
	ret := _m.Called(ctx, goalId)

	if len(ret) == 0 {
		panic("no return value specified for GetCurrentProgress")
	}

	var r0 entities.Progress
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, int) (entities.Progress, error)); ok {
		return rf(ctx, goalId)
	}
	if rf, ok := ret.Get(0).(func(context.Context, int) entities.Progress); ok {
		r0 = rf(ctx, goalId)
	} else {
		r0 = ret.Get(0).(entities.Progress)
	}

	if rf, ok := ret.Get(1).(func(context.Context, int) error); ok {
		r1 = rf(ctx, goalId)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetPreviousPeriodExecutionCount provides a mock function with given fields: ctx, goal, currentTime
func (_m *MockStorage) GetPreviousPeriodExecutionCount(ctx context.Context, goal entities.Goal, currentTime time.Time) (int, error) {
	ret := _m.Called(ctx, goal, currentTime)

	if len(ret) == 0 {
		panic("no return value specified for GetPreviousPeriodExecutionCount")
	}

	var r0 int
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, entities.Goal, time.Time) (int, error)); ok {
		return rf(ctx, goal, currentTime)
	}
	if rf, ok := ret.Get(0).(func(context.Context, entities.Goal, time.Time) int); ok {
		r0 = rf(ctx, goal, currentTime)
	} else {
		r0 = ret.Get(0).(int)
	}

	if rf, ok := ret.Get(1).(func(context.Context, entities.Goal, time.Time) error); ok {
		r1 = rf(ctx, goal, currentTime)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// SetGoalNextCheckDate provides a mock function with given fields: ctx, goalId, nextCheckDate
func (_m *MockStorage) SetGoalNextCheckDate(ctx context.Context, goalId int, nextCheckDate time.Time) error {
	ret := _m.Called(ctx, goalId, nextCheckDate)

	if len(ret) == 0 {
		panic("no return value specified for SetGoalNextCheckDate")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, int, time.Time) error); ok {
		r0 = rf(ctx, goalId, nextCheckDate)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// UpdateProgressByID provides a mock function with given fields: ctx, progress
func (_m *MockStorage) UpdateProgressByID(ctx context.Context, progress entities.Progress) error {
	ret := _m.Called(ctx, progress)

	if len(ret) == 0 {
		panic("no return value specified for UpdateProgressByID")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, entities.Progress) error); ok {
		r0 = rf(ctx, progress)
	} else {
		r0 = ret.Error(0)
	}

	return r0
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
