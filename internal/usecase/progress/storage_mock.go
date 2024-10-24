// Code generated by mockery v2.46.3. DO NOT EDIT.

package progress

import (
	context "context"
	entities "testing_trainer/internal/entities"

	mock "github.com/stretchr/testify/mock"
)

// MockStorage is an autogenerated mock type for the Storage type
type MockStorage struct {
	mock.Mock
}

// AddHabitProgress provides a mock function with given fields: ctx, goalId
func (_m *MockStorage) AddHabitProgress(ctx context.Context, goalId int) error {
	ret := _m.Called(ctx, goalId)

	if len(ret) == 0 {
		panic("no return value specified for AddHabitProgress")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, int) error); ok {
		r0 = rf(ctx, goalId)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetCurrentPeriodExecutionCount provides a mock function with given fields: ctx, goalId, frequencyType
func (_m *MockStorage) GetCurrentPeriodExecutionCount(ctx context.Context, goalId int, frequencyType entities.FrequencyType) (int, error) {
	ret := _m.Called(ctx, goalId, frequencyType)

	if len(ret) == 0 {
		panic("no return value specified for GetCurrentPeriodExecutionCount")
	}

	var r0 int
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, int, entities.FrequencyType) (int, error)); ok {
		return rf(ctx, goalId, frequencyType)
	}
	if rf, ok := ret.Get(0).(func(context.Context, int, entities.FrequencyType) int); ok {
		r0 = rf(ctx, goalId, frequencyType)
	} else {
		r0 = ret.Get(0).(int)
	}

	if rf, ok := ret.Get(1).(func(context.Context, int, entities.FrequencyType) error); ok {
		r1 = rf(ctx, goalId, frequencyType)
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

// GetHabitGoal provides a mock function with given fields: ctx, habitName
func (_m *MockStorage) GetHabitGoal(ctx context.Context, habitName string) (entities.Goal, error) {
	ret := _m.Called(ctx, habitName)

	if len(ret) == 0 {
		panic("no return value specified for GetHabitGoal")
	}

	var r0 entities.Goal
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (entities.Goal, error)); ok {
		return rf(ctx, habitName)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) entities.Goal); ok {
		r0 = rf(ctx, habitName)
	} else {
		r0 = ret.Get(0).(entities.Goal)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, habitName)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetPreviousPeriodExecutionCount provides a mock function with given fields: ctx, goalId, frequencyType
func (_m *MockStorage) GetPreviousPeriodExecutionCount(ctx context.Context, goalId int, frequencyType entities.FrequencyType) (int, error) {
	ret := _m.Called(ctx, goalId, frequencyType)

	if len(ret) == 0 {
		panic("no return value specified for GetPreviousPeriodExecutionCount")
	}

	var r0 int
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, int, entities.FrequencyType) (int, error)); ok {
		return rf(ctx, goalId, frequencyType)
	}
	if rf, ok := ret.Get(0).(func(context.Context, int, entities.FrequencyType) int); ok {
		r0 = rf(ctx, goalId, frequencyType)
	} else {
		r0 = ret.Get(0).(int)
	}

	if rf, ok := ret.Get(1).(func(context.Context, int, entities.FrequencyType) error); ok {
		r1 = rf(ctx, goalId, frequencyType)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// UpdateGoalStat provides a mock function with given fields: ctx, goalId, _a2
func (_m *MockStorage) UpdateGoalStat(ctx context.Context, goalId int, _a2 entities.Progress) error {
	ret := _m.Called(ctx, goalId, _a2)

	if len(ret) == 0 {
		panic("no return value specified for UpdateGoalStat")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, int, entities.Progress) error); ok {
		r0 = rf(ctx, goalId, _a2)
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
