// Code generated by mockery v2.46.3. DO NOT EDIT.

package habit

import (
	context "context"
	entities "testing_trainer/internal/entities"

	mock "github.com/stretchr/testify/mock"
)

// MockStorage is an autogenerated mock type for the Storage type
type MockStorage struct {
	mock.Mock
}

// CreateHabit provides a mock function with given fields: ctx, username, _a2
func (_m *MockStorage) CreateHabit(ctx context.Context, username string, _a2 entities.Habit) (int64, error) {
	ret := _m.Called(ctx, username, _a2)

	if len(ret) == 0 {
		panic("no return value specified for CreateHabit")
	}

	var r0 int64
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, entities.Habit) (int64, error)); ok {
		return rf(ctx, username, _a2)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, entities.Habit) int64); ok {
		r0 = rf(ctx, username, _a2)
	} else {
		r0 = ret.Get(0).(int64)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, entities.Habit) error); ok {
		r1 = rf(ctx, username, _a2)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetHabitByName provides a mock function with given fields: ctx, username, habitName
func (_m *MockStorage) GetHabitByName(ctx context.Context, username string, habitName string) (entities.Habit, error) {
	ret := _m.Called(ctx, username, habitName)

	if len(ret) == 0 {
		panic("no return value specified for GetHabitByName")
	}

	var r0 entities.Habit
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string) (entities.Habit, error)); ok {
		return rf(ctx, username, habitName)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, string) entities.Habit); ok {
		r0 = rf(ctx, username, habitName)
	} else {
		r0 = ret.Get(0).(entities.Habit)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, string) error); ok {
		r1 = rf(ctx, username, habitName)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListUserHabits provides a mock function with given fields: ctx, username
func (_m *MockStorage) ListUserHabits(ctx context.Context, username string) ([]entities.Habit, error) {
	ret := _m.Called(ctx, username)

	if len(ret) == 0 {
		panic("no return value specified for ListUserHabits")
	}

	var r0 []entities.Habit
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) ([]entities.Habit, error)); ok {
		return rf(ctx, username)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) []entities.Habit); ok {
		r0 = rf(ctx, username)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]entities.Habit)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, username)
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