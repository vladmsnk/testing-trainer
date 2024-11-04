package habit

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"testing_trainer/internal/entities"
	"testing_trainer/internal/storage"
	"testing_trainer/internal/usecase/user"
)

func TestCreateHabit(t *testing.T) {

	initFunc := func(t *testing.T) (*MockStorage, *MockUserUseCase) {
		mockStorage := NewMockStorage(t)
		mockUserUc := NewMockUserUseCase(t)

		return mockStorage, mockUserUc
	}

	var (
		ctx             = context.Background()
		username        = "username"
		expectedHabitID = int64(1)
		habitId         = "1"
	)

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		mockStorage, mockUserUc := initFunc(t)

		mockUserUc.On("GetUserByUsername", ctx, username).Return(entities.User{Name: username}, nil)
		mockStorage.On("CreateHabit", ctx, username, entities.Habit{Id: habitId}).Return(expectedHabitID, nil)
		habituc := New(mockStorage, mockUserUc)

		habitID, err := habituc.CreateHabit(ctx, username, entities.Habit{Id: habitId})
		require.Nil(t, err, "unexpected error")
		require.Equal(t, expectedHabitID, habitID)
	})

	t.Run("user not found", func(t *testing.T) {
		t.Parallel()

		mockStorage, mockUserUc := initFunc(t)

		mockUserUc.On("GetUserByUsername", ctx, username).Return(entities.User{}, user.ErrUserNotFound)
		habituc := New(mockStorage, mockUserUc)

		habitID, err := habituc.CreateHabit(ctx, username, entities.Habit{Id: habitId})
		require.ErrorIs(t, err, user.ErrUserNotFound, "unexpected error")
		require.Zero(t, habitID)
	})
}

func TestListUserHabits(t *testing.T) {
	var (
		ctx            = context.Background()
		username       = "username"
		expectedHabits = []entities.Habit{{Id: "1"}, {Id: "2"}}
	)

	initFunc := func(t *testing.T) (*MockStorage, *MockUserUseCase) {
		mockStorage := NewMockStorage(t)
		mockUserUc := NewMockUserUseCase(t)

		return mockStorage, mockUserUc
	}

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		mockStorage, mockUserUc := initFunc(t)

		mockUserUc.On("GetUserByUsername", ctx, username).Return(entities.User{Name: username}, nil)
		mockStorage.On("ListUserHabits", ctx, username).Return(expectedHabits, nil)
		habituc := New(mockStorage, mockUserUc)

		habits, err := habituc.ListUserHabits(ctx, username)
		require.Nil(t, err, "unexpected error")
		require.Equal(t, expectedHabits, habits)
	})

	t.Run("user not found", func(t *testing.T) {
		t.Parallel()

		mockStorage, mockUserUc := initFunc(t)

		mockUserUc.On("GetUserByUsername", ctx, username).Return(entities.User{}, storage.ErrNotFound)
		habituc := New(mockStorage, mockUserUc)

		habits, err := habituc.ListUserHabits(ctx, username)
		require.ErrorIs(t, err, ErrUserNotFound, "unexpected error")
		require.Nil(t, habits)
	})
}

func TestUpdateHabit(t *testing.T) {

	initFunc := func(t *testing.T) (*MockStorage, *MockUserUseCase) {
		mockStorage := NewMockStorage(t)
		mockUserUc := NewMockUserUseCase(t)

		return mockStorage, mockUserUc
	}

	var (
		ctx      = context.Background()
		username = "username"
		habitId  = "1"
	)

	var (
		goal            = entities.Goal{Id: 1, FrequencyType: entities.Daily, TimesPerFrequency: 2, TotalTrackingPeriods: 30}
		habit           = entities.Habit{Id: habitId, Description: "Drink water", Goal: &goal}
		newGoal         = entities.Goal{Id: 2, FrequencyType: entities.Daily, TimesPerFrequency: 3, TotalTrackingPeriods: 30}
		currentProgress = entities.Progress{
			TotalCompletedPeriods: 1,
			TotalCompletedTimes:   2,
			CurrentStreak:         1,
			MostLongestStreak:     1,
		}
		newHabit = entities.Habit{
			Id:          habitId,
			Description: "Drink water",
			Goal:        &newGoal,
		}
	)

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		mockStorage, mockUserUc := initFunc(t)
		mockUserUc.On("GetUserByUsername", ctx, username).Return(entities.User{Name: username}, nil)

		mockStorage.On("GetHabitById", ctx, username, habitId).Return(habit, nil)
		mockStorage.On("DeactivateGoalByID", ctx, goal.Id).Return(nil)
		mockStorage.On("CreateGoal", ctx, habitId, newGoal).Return(newGoal.Id, nil)
		mockStorage.On("GetCurrentProgress", ctx, goal.Id).Return(currentProgress, nil)
		mockStorage.On("UpdateGoalStat", ctx, newGoal.Id, currentProgress).Return(nil)

		habituc := New(mockStorage, mockUserUc)
		err := habituc.UpdateHabit(ctx, username, newHabit)
		require.Nil(t, err)
	})
}
