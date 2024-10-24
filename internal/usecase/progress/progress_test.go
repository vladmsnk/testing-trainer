package progress

import (
	"context"
	"github.com/stretchr/testify/require"
	"testing"
	"testing_trainer/internal/entities"
	"testing_trainer/internal/storage"
	"testing_trainer/internal/usecase/user"
)

func TestGetHabitProgress(t *testing.T) {
	initFunc := func(t *testing.T) (*MockStorage, *MockUserUseCase) {
		mockStorage := NewMockStorage(t)
		mockUserUc := NewMockUserUseCase(t)

		return mockStorage, mockUserUc
	}

	var (
		ctx      = context.Background()
		username = "username"
		habitId  = "1"
		goal     = entities.Goal{
			Id:                   1,
			FrequencyType:        entities.Daily,
			TimesPerFrequency:    2,
			TotalTrackingPeriods: 30,
		}

		progress = entities.Progress{
			TotalCompletedPeriods: 10,
			TotalSkippedPeriods:   0,
			TotalCompletedTimes:   20,
			MostLongestStreak:     10,
			CurrentStreak:         10,
		}

		expectedResult = entities.ProgressWithGoal{
			Progress: progress,
			Goal:     goal,
		}
	)

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		mockStorage, mockUserUc := initFunc(t)

		mockUserUc.On("GetUserByUsername", ctx, username).Return(entities.User{}, nil)
		mockStorage.On("GetHabitGoal", ctx, habitId).Return(goal, nil)
		mockStorage.On("GetCurrentProgress", ctx, 1).Return(progress, nil)

		progressUC := New(mockUserUc, mockStorage)

		habitProgress, err := progressUC.GetHabitProgress(ctx, username, habitId)
		require.Nil(t, err)
		require.Equal(t, expectedResult, habitProgress)
	})

	t.Run("user not found", func(t *testing.T) {
		t.Parallel()

		mockStorage, mockUserUc := initFunc(t)

		mockUserUc.On("GetUserByUsername", ctx, username).Return(entities.User{}, user.ErrUserNotFound)

		progressUC := New(mockUserUc, mockStorage)

		habitProgress, err := progressUC.GetHabitProgress(ctx, username, habitId)
		require.ErrorIs(t, err, user.ErrUserNotFound)
		require.Zero(t, habitProgress)
	})

	t.Run("get habit goal not found", func(t *testing.T) {
		t.Parallel()

		mockStorage, mockUserUc := initFunc(t)

		mockUserUc.On("GetUserByUsername", ctx, username).Return(entities.User{}, nil)
		mockStorage.On("GetHabitGoal", ctx, habitId).Return(entities.Goal{}, storage.ErrNotFound)

		progressUC := New(mockUserUc, mockStorage)

		habitProgress, err := progressUC.GetHabitProgress(ctx, username, habitId)
		require.ErrorIs(t, err, ErrHabitGoalNotFound)
		require.Zero(t, habitProgress)
	})
}

func TestAddHabitProgress(t *testing.T) {
	initFunc := func(t *testing.T) (*MockStorage, *MockUserUseCase) {
		mockStorage := NewMockStorage(t)
		mockUserUc := NewMockUserUseCase(t)

		return mockStorage, mockUserUc
	}

	var (
		username = "username"
		ctx      = context.Background()

		goal = entities.Goal{
			Id:                   1,
			FrequencyType:        entities.Daily,
			TimesPerFrequency:    2,
			TotalTrackingPeriods: 30,
		}
		habitId = "habit"
	)

	t.Run("success: increase total completed periods", func(t *testing.T) {
		t.Parallel()

		currentProgress := entities.Progress{
			TotalCompletedPeriods: 10,
			TotalSkippedPeriods:   0,
			TotalCompletedTimes:   21,
			MostLongestStreak:     10,
			CurrentStreak:         10,
		}

		updateGoalStateEntity := entities.Progress{
			TotalCompletedPeriods: 11,
			TotalSkippedPeriods:   0,
			TotalCompletedTimes:   22,
			MostLongestStreak:     11,
			CurrentStreak:         11,
		}

		mockStorage, mockUserUc := initFunc(t)

		mockUserUc.On("GetUserByUsername", ctx, username).Return(entities.User{}, nil)
		mockStorage.On("GetHabitGoal", ctx, habitId).Return(goal, nil)
		mockStorage.On("GetCurrentProgress", ctx, 1).Return(currentProgress, nil)
		mockStorage.On("GetPreviousPeriodExecutionCount", ctx, goal.Id, goal.FrequencyType).Return(2, nil)
		mockStorage.On("GetCurrentPeriodExecutionCount", ctx, goal.Id, goal.FrequencyType).Return(1, nil)
		mockStorage.On("AddHabitProgress", ctx, goal.Id).Return(nil)
		mockStorage.On("UpdateGoalStat", ctx, goal.Id, updateGoalStateEntity).Return(nil)

		progressUC := New(mockUserUc, mockStorage)
		err := progressUC.AddHabitProgress(ctx, username, habitId)
		require.Nil(t, err)
	})

	t.Run("success: increase total completed times", func(t *testing.T) {
		currentProgress := entities.Progress{
			TotalCompletedPeriods: 10,
			TotalSkippedPeriods:   0,
			TotalCompletedTimes:   20,
			MostLongestStreak:     10,
			CurrentStreak:         10,
		}

		updateGoalStateEntity := entities.Progress{
			TotalCompletedPeriods: 10,
			TotalSkippedPeriods:   0,
			TotalCompletedTimes:   21,
			MostLongestStreak:     10,
			CurrentStreak:         10,
		}

		mockStorage, mockUserUc := initFunc(t)

		mockUserUc.On("GetUserByUsername", ctx, username).Return(entities.User{}, nil)
		mockStorage.On("GetHabitGoal", ctx, habitId).Return(goal, nil)
		mockStorage.On("GetCurrentProgress", ctx, 1).Return(currentProgress, nil)
		mockStorage.On("GetPreviousPeriodExecutionCount", ctx, goal.Id, goal.FrequencyType).Return(2, nil)
		mockStorage.On("GetCurrentPeriodExecutionCount", ctx, goal.Id, goal.FrequencyType).Return(0, nil)
		mockStorage.On("AddHabitProgress", ctx, goal.Id).Return(nil)
		mockStorage.On("UpdateGoalStat", ctx, goal.Id, updateGoalStateEntity).Return(nil)

		progressUC := New(mockUserUc, mockStorage)
		err := progressUC.AddHabitProgress(ctx, username, habitId)
		require.Nil(t, err)
	})

	t.Run("success: start from scratch: increase total completed times", func(t *testing.T) {
		currentProgress := entities.Progress{
			TotalCompletedPeriods: 0,
			TotalSkippedPeriods:   0,
			TotalCompletedTimes:   0,
			MostLongestStreak:     0,
			CurrentStreak:         0,
		}

		updateGoalStateEntity := entities.Progress{
			TotalCompletedPeriods: 0,
			TotalSkippedPeriods:   0,
			TotalCompletedTimes:   1,
			MostLongestStreak:     0,
			CurrentStreak:         0,
		}

		mockStorage, mockUserUc := initFunc(t)

		mockUserUc.On("GetUserByUsername", ctx, username).Return(entities.User{}, nil)
		mockStorage.On("GetHabitGoal", ctx, habitId).Return(goal, nil)
		mockStorage.On("GetCurrentProgress", ctx, 1).Return(currentProgress, nil)
		mockStorage.On("GetPreviousPeriodExecutionCount", ctx, goal.Id, goal.FrequencyType).Return(0, nil)
		mockStorage.On("GetCurrentPeriodExecutionCount", ctx, goal.Id, goal.FrequencyType).Return(0, nil)
		mockStorage.On("AddHabitProgress", ctx, goal.Id).Return(nil)
		mockStorage.On("UpdateGoalStat", ctx, goal.Id, updateGoalStateEntity).Return(nil)

		progressUC := New(mockUserUc, mockStorage)
		err := progressUC.AddHabitProgress(ctx, username, habitId)
		require.Nil(t, err)
	})

	t.Run("success: start from scratch: increase total completed periods", func(t *testing.T) {
		currentProgress := entities.Progress{
			TotalCompletedPeriods: 0,
			TotalSkippedPeriods:   0,
			TotalCompletedTimes:   1,
			MostLongestStreak:     0,
			CurrentStreak:         0,
		}

		updateGoalStateEntity := entities.Progress{
			TotalCompletedPeriods: 1,
			TotalSkippedPeriods:   0,
			TotalCompletedTimes:   2,
			MostLongestStreak:     1,
			CurrentStreak:         1,
		}

		mockStorage, mockUserUc := initFunc(t)

		mockUserUc.On("GetUserByUsername", ctx, username).Return(entities.User{}, nil)
		mockStorage.On("GetHabitGoal", ctx, habitId).Return(goal, nil)
		mockStorage.On("GetCurrentProgress", ctx, 1).Return(currentProgress, nil)
		mockStorage.On("GetPreviousPeriodExecutionCount", ctx, goal.Id, goal.FrequencyType).Return(0, nil)
		mockStorage.On("GetCurrentPeriodExecutionCount", ctx, goal.Id, goal.FrequencyType).Return(1, nil)
		mockStorage.On("AddHabitProgress", ctx, goal.Id).Return(nil)
		mockStorage.On("UpdateGoalStat", ctx, goal.Id, updateGoalStateEntity).Return(nil)

		progressUC := New(mockUserUc, mockStorage)
		err := progressUC.AddHabitProgress(ctx, username, habitId)
		require.Nil(t, err)
	})

	t.Run("success: with two days skips: update total completed times", func(t *testing.T) {
		currentProgress := entities.Progress{
			TotalCompletedPeriods: 4,
			TotalSkippedPeriods:   2,
			TotalCompletedTimes:   8,
			MostLongestStreak:     4,
			CurrentStreak:         0,
		}

		updateGoalStateEntity := entities.Progress{
			TotalCompletedPeriods: 4,
			TotalSkippedPeriods:   2,
			TotalCompletedTimes:   9,
			MostLongestStreak:     4,
			CurrentStreak:         0,
		}

		mockStorage, mockUserUc := initFunc(t)

		mockUserUc.On("GetUserByUsername", ctx, username).Return(entities.User{}, nil)
		mockStorage.On("GetHabitGoal", ctx, habitId).Return(goal, nil)
		mockStorage.On("GetCurrentProgress", ctx, 1).Return(currentProgress, nil)
		mockStorage.On("GetPreviousPeriodExecutionCount", ctx, goal.Id, goal.FrequencyType).Return(0, nil)
		mockStorage.On("GetCurrentPeriodExecutionCount", ctx, goal.Id, goal.FrequencyType).Return(0, nil)
		mockStorage.On("AddHabitProgress", ctx, goal.Id).Return(nil)
		mockStorage.On("UpdateGoalStat", ctx, goal.Id, updateGoalStateEntity).Return(nil)

		progressUC := New(mockUserUc, mockStorage)
		err := progressUC.AddHabitProgress(ctx, username, habitId)
		require.Nil(t, err)
	})

	t.Run("success: with two days skips: update total completed periods", func(t *testing.T) {
		currentProgress := entities.Progress{
			TotalCompletedPeriods: 4,
			TotalSkippedPeriods:   2,
			TotalCompletedTimes:   9,
			MostLongestStreak:     4,
			CurrentStreak:         0,
		}

		updateGoalStateEntity := entities.Progress{
			TotalCompletedPeriods: 5,
			TotalSkippedPeriods:   2,
			TotalCompletedTimes:   10,
			MostLongestStreak:     4,
			CurrentStreak:         1,
		}

		mockStorage, mockUserUc := initFunc(t)

		mockUserUc.On("GetUserByUsername", ctx, username).Return(entities.User{}, nil)
		mockStorage.On("GetHabitGoal", ctx, habitId).Return(goal, nil)
		mockStorage.On("GetCurrentProgress", ctx, 1).Return(currentProgress, nil)
		mockStorage.On("GetPreviousPeriodExecutionCount", ctx, goal.Id, goal.FrequencyType).Return(0, nil)
		mockStorage.On("GetCurrentPeriodExecutionCount", ctx, goal.Id, goal.FrequencyType).Return(1, nil)
		mockStorage.On("AddHabitProgress", ctx, goal.Id).Return(nil)
		mockStorage.On("UpdateGoalStat", ctx, goal.Id, updateGoalStateEntity).Return(nil)

		progressUC := New(mockUserUc, mockStorage)
		err := progressUC.AddHabitProgress(ctx, username, habitId)
		require.Nil(t, err)
	})

	t.Run("set goal is completed", func(t *testing.T) {
		currentProgress := entities.Progress{
			TotalCompletedPeriods: 29,
			TotalSkippedPeriods:   0,
			TotalCompletedTimes:   59,
			MostLongestStreak:     29,
			CurrentStreak:         29,
		}

		updateGoalStateEntity := entities.Progress{
			TotalCompletedPeriods: 30,
			TotalSkippedPeriods:   0,
			TotalCompletedTimes:   60,
			MostLongestStreak:     30,
			CurrentStreak:         30,
		}

		mockStorage, mockUserUc := initFunc(t)

		mockUserUc.On("GetUserByUsername", ctx, username).Return(entities.User{}, nil)
		mockStorage.On("GetHabitGoal", ctx, habitId).Return(goal, nil)
		mockStorage.On("GetCurrentProgress", ctx, 1).Return(currentProgress, nil)
		mockStorage.On("GetPreviousPeriodExecutionCount", ctx, goal.Id, goal.FrequencyType).Return(0, nil)
		mockStorage.On("GetCurrentPeriodExecutionCount", ctx, goal.Id, goal.FrequencyType).Return(1, nil)
		mockStorage.On("AddHabitProgress", ctx, goal.Id).Return(nil)
		mockStorage.On("UpdateGoalStat", ctx, goal.Id, updateGoalStateEntity).Return(nil)
		mockStorage.On("SetGoalCompleted", ctx, goal.Id).Return(nil)

		progressUC := New(mockUserUc, mockStorage)
		err := progressUC.AddHabitProgress(ctx, username, habitId)
		require.Nil(t, err)
	})
}
