package progress

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"testing"
	"testing_trainer/internal/entities"
	"testing_trainer/internal/storage"
	"testing_trainer/internal/usecase/user"
	"time"
)

func TestGetHabitProgress(t *testing.T) {
	initFunc := func(t *testing.T) (*MockStorage, *MockUserUseCase, *MockTransactor) {
		mockStorage := NewMockStorage(t)
		mockUserUc := NewMockUserUseCase(t)
		mockTransactor := NewMockTransactor(t)

		return mockStorage, mockUserUc, mockTransactor
	}

	var (
		ctx      = context.Background()
		username = "username"
		habitId  = 1
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

		mockStorage, mockUserUc, mockTx := initFunc(t)

		mockUserUc.On("GetUserByUsername", ctx, username).Return(entities.User{}, nil)
		mockStorage.On("GetHabitGoal", ctx, habitId).Return(goal, nil)
		mockStorage.On("GetCurrentProgress", ctx, 1).Return(progress, nil)

		progressUC := New(mockUserUc, mockStorage, mockTx)

		habitProgress, err := progressUC.GetHabitProgress(ctx, username, habitId)
		require.Nil(t, err)
		require.Equal(t, expectedResult, habitProgress)
	})

	t.Run("user not found", func(t *testing.T) {
		t.Parallel()

		mockStorage, mockUserUc, mockTx := initFunc(t)

		mockUserUc.On("GetUserByUsername", ctx, username).Return(entities.User{}, user.ErrUserNotFound)

		progressUC := New(mockUserUc, mockStorage, mockTx)

		habitProgress, err := progressUC.GetHabitProgress(ctx, username, habitId)
		require.ErrorIs(t, err, user.ErrUserNotFound)
		require.Zero(t, habitProgress)
	})

	t.Run("get habit goal not found", func(t *testing.T) {
		t.Parallel()

		mockStorage, mockUserUc, mockTx := initFunc(t)

		mockUserUc.On("GetUserByUsername", ctx, username).Return(entities.User{}, nil)
		mockStorage.On("GetHabitGoal", ctx, habitId).Return(entities.Goal{}, storage.ErrNotFound)

		progressUC := New(mockUserUc, mockStorage, mockTx)

		habitProgress, err := progressUC.GetHabitProgress(ctx, username, habitId)
		require.ErrorIs(t, err, ErrHabitGoalNotFound)
		require.Zero(t, habitProgress)
	})
}

func TestAddHabitProgress(t *testing.T) {
	initFunc := func(t *testing.T) (*MockStorage, *MockUserUseCase, *MockTransactor) {
		mockStorage := NewMockStorage(t)
		mockUserUc := NewMockUserUseCase(t)
		mockTransactor := NewMockTransactor(t)

		return mockStorage, mockUserUc, mockTransactor
	}

	var (
		username = "username"
		ctx      = context.Background()

		goal = entities.Goal{
			Id:                   1,
			FrequencyType:        entities.Daily,
			TimesPerFrequency:    2,
			TotalTrackingPeriods: 30,
			CreatedAt:            time.Now().AddDate(0, 0, -1),
		}
		habitId       = 1
		currentPeriod = 1
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

		mockStorage, mockUserUc, mockTransactor := initFunc(t)

		mockTransactor.On("RunRepeatableRead", ctx, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
			fx := args.Get(1).(func(ctxTX context.Context) error)
			_ = fx(ctx)
		})

		mockUserUc.On("GetUserByUsername", ctx, username).Return(entities.User{}, nil)
		mockStorage.On("GetHabitGoal", ctx, habitId).Return(goal, nil)
		mockStorage.On("GetCurrentProgress", ctx, 1).Return(currentProgress, nil)
		mockStorage.On("GetPreviousPeriodExecutionCount", ctx, goal.Id, goal.FrequencyType, goal.CreatedAt, currentPeriod).Return(2, nil)
		mockStorage.On("GetCurrentPeriodExecutionCount", ctx, goal.Id, goal.FrequencyType, goal.CreatedAt, currentPeriod).Return(1, nil)
		mockStorage.On("AddHabitProgress", ctx, goal.Id).Return(nil)
		mockStorage.On("UpdateGoalStat", ctx, goal.Id, updateGoalStateEntity).Return(nil)

		progressUC := New(mockUserUc, mockStorage, mockTransactor)
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

		mockStorage, mockUserUc, mockTransactor := initFunc(t)

		mockTransactor.On("RunRepeatableRead", ctx, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
			fx := args.Get(1).(func(ctxTX context.Context) error)
			_ = fx(ctx)
		})

		mockUserUc.On("GetUserByUsername", ctx, username).Return(entities.User{}, nil)
		mockStorage.On("GetHabitGoal", ctx, habitId).Return(goal, nil)
		mockStorage.On("GetCurrentProgress", ctx, 1).Return(currentProgress, nil)
		mockStorage.On("GetPreviousPeriodExecutionCount", ctx, goal.Id, goal.FrequencyType, goal.CreatedAt, currentPeriod).Return(2, nil)
		mockStorage.On("GetCurrentPeriodExecutionCount", ctx, goal.Id, goal.FrequencyType, goal.CreatedAt, currentPeriod).Return(0, nil)
		mockStorage.On("AddHabitProgress", ctx, goal.Id).Return(nil)
		mockStorage.On("UpdateGoalStat", ctx, goal.Id, updateGoalStateEntity).Return(nil)

		progressUC := New(mockUserUc, mockStorage, mockTransactor)
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

		mockStorage, mockUserUc, mockTransactor := initFunc(t)

		mockTransactor.On("RunRepeatableRead", ctx, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
			fx := args.Get(1).(func(ctxTX context.Context) error)
			_ = fx(ctx)
		})

		mockUserUc.On("GetUserByUsername", ctx, username).Return(entities.User{}, nil)
		mockStorage.On("GetHabitGoal", ctx, habitId).Return(goal, nil)
		mockStorage.On("GetCurrentProgress", ctx, 1).Return(currentProgress, nil)
		mockStorage.On("GetPreviousPeriodExecutionCount", ctx, goal.Id, goal.FrequencyType, goal.CreatedAt, currentPeriod).Return(0, nil)
		mockStorage.On("GetCurrentPeriodExecutionCount", ctx, goal.Id, goal.FrequencyType, goal.CreatedAt, currentPeriod).Return(0, nil)
		mockStorage.On("AddHabitProgress", ctx, goal.Id).Return(nil)
		mockStorage.On("UpdateGoalStat", ctx, goal.Id, updateGoalStateEntity).Return(nil)

		progressUC := New(mockUserUc, mockStorage, mockTransactor)
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

		mockStorage, mockUserUc, mockTransactor := initFunc(t)

		mockTransactor.On("RunRepeatableRead", ctx, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
			fx := args.Get(1).(func(ctxTX context.Context) error)
			_ = fx(ctx)
		})

		mockUserUc.On("GetUserByUsername", ctx, username).Return(entities.User{}, nil)
		mockStorage.On("GetHabitGoal", ctx, habitId).Return(goal, nil)
		mockStorage.On("GetCurrentProgress", ctx, 1).Return(currentProgress, nil)
		mockStorage.On("GetPreviousPeriodExecutionCount", ctx, goal.Id, goal.FrequencyType, goal.CreatedAt, currentPeriod).Return(0, nil)
		mockStorage.On("GetCurrentPeriodExecutionCount", ctx, goal.Id, goal.FrequencyType, goal.CreatedAt, currentPeriod).Return(1, nil)
		mockStorage.On("AddHabitProgress", ctx, goal.Id).Return(nil)
		mockStorage.On("UpdateGoalStat", ctx, goal.Id, updateGoalStateEntity).Return(nil)

		progressUC := New(mockUserUc, mockStorage, mockTransactor)
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

		mockStorage, mockUserUc, mockTransactor := initFunc(t)

		mockTransactor.On("RunRepeatableRead", ctx, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
			fx := args.Get(1).(func(ctxTX context.Context) error)
			_ = fx(ctx)
		})

		mockUserUc.On("GetUserByUsername", ctx, username).Return(entities.User{}, nil)
		mockStorage.On("GetHabitGoal", ctx, habitId).Return(goal, nil)
		mockStorage.On("GetCurrentProgress", ctx, 1).Return(currentProgress, nil)
		mockStorage.On("GetPreviousPeriodExecutionCount", ctx, goal.Id, goal.FrequencyType, goal.CreatedAt, currentPeriod).Return(0, nil)
		mockStorage.On("GetCurrentPeriodExecutionCount", ctx, goal.Id, goal.FrequencyType, goal.CreatedAt, currentPeriod).Return(0, nil)
		mockStorage.On("AddHabitProgress", ctx, goal.Id).Return(nil)
		mockStorage.On("UpdateGoalStat", ctx, goal.Id, updateGoalStateEntity).Return(nil)

		progressUC := New(mockUserUc, mockStorage, mockTransactor)
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

		mockStorage, mockUserUc, mockTransactor := initFunc(t)

		mockTransactor.On("RunRepeatableRead", ctx, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
			fx := args.Get(1).(func(ctxTX context.Context) error)
			_ = fx(ctx)
		})

		mockUserUc.On("GetUserByUsername", ctx, username).Return(entities.User{}, nil)
		mockStorage.On("GetHabitGoal", ctx, habitId).Return(goal, nil)
		mockStorage.On("GetCurrentProgress", ctx, 1).Return(currentProgress, nil)
		mockStorage.On("GetPreviousPeriodExecutionCount", ctx, goal.Id, goal.FrequencyType, goal.CreatedAt, currentPeriod).Return(0, nil)
		mockStorage.On("GetCurrentPeriodExecutionCount", ctx, goal.Id, goal.FrequencyType, goal.CreatedAt, currentPeriod).Return(1, nil)
		mockStorage.On("AddHabitProgress", ctx, goal.Id).Return(nil)
		mockStorage.On("UpdateGoalStat", ctx, goal.Id, updateGoalStateEntity).Return(nil)

		progressUC := New(mockUserUc, mockStorage, mockTransactor)
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

		mockStorage, mockUserUc, mockTransactor := initFunc(t)

		mockTransactor.On("RunRepeatableRead", ctx, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
			fx := args.Get(1).(func(ctxTX context.Context) error)
			_ = fx(ctx)
		})

		mockUserUc.On("GetUserByUsername", ctx, username).Return(entities.User{}, nil)
		mockStorage.On("GetHabitGoal", ctx, habitId).Return(goal, nil)
		mockStorage.On("GetCurrentProgress", ctx, 1).Return(currentProgress, nil)
		mockStorage.On("GetPreviousPeriodExecutionCount", ctx, goal.Id, goal.FrequencyType, goal.CreatedAt, currentPeriod).Return(2, nil)
		mockStorage.On("GetCurrentPeriodExecutionCount", ctx, goal.Id, goal.FrequencyType, goal.CreatedAt, currentPeriod).Return(1, nil)
		mockStorage.On("AddHabitProgress", ctx, goal.Id).Return(nil)
		mockStorage.On("UpdateGoalStat", ctx, goal.Id, updateGoalStateEntity).Return(nil)
		mockStorage.On("SetGoalCompleted", ctx, goal.Id).Return(nil)

		progressUC := New(mockUserUc, mockStorage, mockTransactor)
		err := progressUC.AddHabitProgress(ctx, username, habitId)
		require.Nil(t, err)
	})
}

func TestGetCurrentProgressForAllUserHabits(t *testing.T) {
	initFunc := func(t *testing.T) (*MockStorage, *MockUserUseCase) {
		mockStorage := NewMockStorage(t)
		mockUserUc := NewMockUserUseCase(t)

		return mockStorage, mockUserUc
	}

	var (
		ctx    = context.Background()
		habit1 = entities.Habit{
			Id: 1,
			Goal: &entities.Goal{
				Id:                   1,
				FrequencyType:        entities.Daily,
				TimesPerFrequency:    2,
				TotalTrackingPeriods: 30,
				CreatedAt:            time.Now().AddDate(0, 0, -1),
			},
			Description: "habit1",
		}

		// create habit2
		habit2 = entities.Habit{
			Id: 2,
			Goal: &entities.Goal{
				Id:                   2,
				FrequencyType:        entities.Weekly,
				TimesPerFrequency:    2,
				TotalTrackingPeriods: 4,
				CreatedAt:            time.Now().AddDate(0, 0, -10),
			},
			Description: "habit2",
		}

		// create habit3 monthly
		habit3 = entities.Habit{
			Id: 3,
			Goal: &entities.Goal{
				Id:                   3,
				FrequencyType:        entities.Monthly,
				TimesPerFrequency:    4,
				TotalTrackingPeriods: 4,
				CreatedAt:            time.Now().AddDate(0, 0, -10),
			},
			Description: "habit3",
		}

		habit4 = entities.Habit{
			Id: 4,
			Goal: &entities.Goal{
				Id:                   4,
				FrequencyType:        entities.Monthly,
				TimesPerFrequency:    4,
				TotalTrackingPeriods: 4,
				CreatedAt:            time.Now().AddDate(0, 0, -10),
			},
			Description: "habit4",
		}

		testCases = []struct {
			currentExecutionCount int
			habit                 entities.Habit
		}{
			{
				currentExecutionCount: 1,
				habit:                 habit1,
			},
			{
				currentExecutionCount: 1,
				habit:                 habit2,
			},
			{
				currentExecutionCount: 1,
				habit:                 habit3,
			},
			{
				currentExecutionCount: 4,
				habit:                 habit4,
			},
		}

		username       = "username"
		expectedResult = []entities.CurrentPeriodProgress{
			{
				Habit:                       habit1,
				CurrentPeriodCompletedTimes: 1,
				NeedToCompleteTimes:         2,
				CurrentPeriod:               2,
			},
			{
				Habit:                       habit2,
				CurrentPeriodCompletedTimes: 1,
				NeedToCompleteTimes:         2,
				CurrentPeriod:               2,
			},
			{
				Habit:                       habit3,
				CurrentPeriodCompletedTimes: 1,
				NeedToCompleteTimes:         4,
				CurrentPeriod:               1,
			},
		}
	)

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		mockStorage, mockUserUc := initFunc(t)

		mockUserUc.On("GetUserByUsername", ctx, username).Return(entities.User{}, nil)

		mockStorage.On("GetAllUserHabitsWithGoals", ctx, username).Return([]entities.Habit{habit1, habit2, habit3, habit4}, nil)

		for _, testCase := range testCases {
			mockStorage.On("GetCurrentPeriodExecutionCount", ctx, testCase.habit.Goal.Id, testCase.habit.Goal.FrequencyType, testCase.habit.Goal.CreatedAt, testCase.habit.Goal.GetCurrentPeriod()).Return(testCase.currentExecutionCount, nil)
		}

		progressUC := New(mockUserUc, mockStorage, nil)
		currentProgressForAllUserHabits, err := progressUC.GetCurrentProgressForAllUserHabits(ctx, username)
		require.Nil(t, err)
		for i, currentProgress := range currentProgressForAllUserHabits {
			assert.Equal(t, expectedResult[i].CurrentPeriodCompletedTimes, currentProgress.CurrentPeriodCompletedTimes)
			assert.Equal(t, expectedResult[i].NeedToCompleteTimes, currentProgress.NeedToCompleteTimes)
			assert.Equal(t, expectedResult[i].CurrentPeriod, currentProgress.CurrentPeriod)
			assert.Equal(t, *expectedResult[i].Habit.Goal, *currentProgress.Habit.Goal)
			assert.Equal(t, expectedResult[i].Habit.Description, currentProgress.Habit.Description)
		}
	})
}
