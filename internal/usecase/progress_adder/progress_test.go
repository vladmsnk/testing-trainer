package progress_adder

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"testing_trainer/internal/entities"
	"testing_trainer/internal/storage"
	"testing_trainer/internal/usecase/user"
)

var (
	initFunc = func(t *testing.T) (*MockStorage, *MockUserUseCase, *MockTransactor, *MockTimeManager) {
		mockStorage := NewMockStorage(t)
		mockUserUc := NewMockUserUseCase(t)
		mockTransactor := NewMockTransactor(t)
		mockTimeManager := NewMockTimeManager(t)

		return mockStorage, mockUserUc, mockTransactor, mockTimeManager
	}

	currentTime = time.Now()
)

func TestGetHabitProgress(t *testing.T) {
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

		habit = entities.Habit{
			Id:          1,
			Goal:        &goal,
			Name:        "habit1",
			Description: "habit1",
		}

		progress = entities.Progress{
			Id:                    1,
			TotalCompletedPeriods: 10,
			TotalSkippedPeriods:   0,
			TotalCompletedTimes:   20,
			MostLongestStreak:     10,
			CurrentStreak:         10,
			Username:              username,
		}

		snapshot = entities.ProgressSnapshot{
			ProgressID: int64(progress.Id),
			Username:   username,
			GoalID:     goal.Id,
		}

		expectedResult = entities.ProgressWithGoal{
			Progress: progress,
			Goal:     goal,
			Habit:    habit,
		}
	)

	t.Run("success get progress by time", func(t *testing.T) {
		t.Parallel()

		mockStorage, mockUserUc, mockTx, mockTime := initFunc(t)

		mockUserUc.On("GetUserByUsername", ctx, username).Return(entities.User{}, nil)

		mockTime.On("GetCurrentTime", ctx, username).Return(currentTime, nil)

		mockStorage.On("GetHabitById", ctx, username, habitId).Return(habit, nil)

		mockStorage.On("GetCurrentSnapshot", ctx, username, goal.Id, currentTime).Return(snapshot, nil)

		mockStorage.On("GetProgressByID", ctx, snapshot.ProgressID).Return(progress, nil)

		progressUC := New(mockUserUc, mockStorage, mockTx, mockTime)

		habitProgress, err := progressUC.GetHabitProgress(ctx, username, habitId)
		require.Nil(t, err)
		require.Equal(t, expectedResult, habitProgress)
	})

	t.Run("user not found", func(t *testing.T) {
		t.Parallel()

		mockStorage, mockUserUc, mockTx, mockTime := initFunc(t)

		mockUserUc.On("GetUserByUsername", ctx, username).Return(entities.User{}, user.ErrUserNotFound)

		progressUC := New(mockUserUc, mockStorage, mockTx, mockTime)

		habitProgress, err := progressUC.GetHabitProgress(ctx, username, habitId)
		require.ErrorIs(t, err, user.ErrUserNotFound)
		require.Zero(t, habitProgress)
	})

	t.Run("get habit goal not found", func(t *testing.T) {
		t.Parallel()

		mockStorage, mockUserUc, mockTx, mockTime := initFunc(t)

		mockUserUc.On("GetUserByUsername", ctx, username).Return(entities.User{}, nil)
		mockTime.On("GetCurrentTime", ctx, username).Return(currentTime, nil)
		mockStorage.On("GetHabitById", ctx, username, habitId).Return(entities.Habit{}, storage.ErrNotFound)

		progressUC := New(mockUserUc, mockStorage, mockTx, mockTime)

		habitProgress, err := progressUC.GetHabitProgress(ctx, username, habitId)
		require.ErrorIs(t, err, ErrHabitNotFound)
		require.Zero(t, habitProgress)
	})
}

func TestAddHabitProgress(t *testing.T) {
	var (
		username = "username"
		ctx      = context.Background()

		startTracking = time.Now().AddDate(0, 0, -1)

		goal = entities.Goal{
			Id:                   1,
			FrequencyType:        entities.Daily,
			TimesPerFrequency:    2,
			TotalTrackingPeriods: 30,
			CreatedAt:            time.Now().AddDate(0, 0, -1),
			StartTrackingAt:      startTracking,
		}
		habitId = 1
	)

	t.Run("success: increase total completed periods", func(t *testing.T) {
		t.Parallel()

		currentProgress := entities.Progress{
			Id:                    1,
			TotalCompletedPeriods: 10,
			TotalSkippedPeriods:   0,
			TotalCompletedTimes:   21,
			MostLongestStreak:     10,
			CurrentStreak:         10,
		}

		updatedProgress := entities.Progress{
			Id:                    1,
			TotalCompletedPeriods: 11,
			TotalSkippedPeriods:   0,
			TotalCompletedTimes:   22,
			MostLongestStreak:     11,
			CurrentStreak:         11,
		}

		mockStorage, mockUserUc, mockTransactor, mockTime := initFunc(t)

		mockTransactor.On("RunRepeatableRead", ctx, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
			fx := args.Get(1).(func(ctxTX context.Context) error)
			_ = fx(ctx)
		})

		mockTime.On("GetCurrentTime", ctx, username).Return(currentTime, nil)
		mockUserUc.On("GetUserByUsername", ctx, username).Return(entities.User{}, nil)
		mockStorage.On("GetHabitGoal", ctx, habitId).Return(goal, nil)

		mockStorage.On("GetProgressByTime", ctx, 1, username, currentTime).Return(currentProgress, nil)

		mockStorage.On("GetPreviousPeriodExecutionCount", ctx, goal, currentTime).Return(2, nil)
		mockStorage.On("GetCurrentPeriodExecutionCount", ctx, goal, currentTime).Return(1, nil)

		mockStorage.On("AddHabitProgress", ctx, goal.Id, currentTime).Return(nil)
		mockStorage.On("UpdateProgressByID", ctx, updatedProgress).Return(nil)

		progressUC := New(mockUserUc, mockStorage, mockTransactor, mockTime)
		err := progressUC.AddHabitProgress(ctx, username, habitId)
		require.Nil(t, err)
	})

	t.Run("success: increase total completed times", func(t *testing.T) {
		currentProgress := entities.Progress{
			Id:                    1,
			TotalCompletedPeriods: 10,
			TotalSkippedPeriods:   0,
			TotalCompletedTimes:   20,
			MostLongestStreak:     10,
			CurrentStreak:         10,
		}

		updatedProgress := entities.Progress{
			Id:                    1,
			TotalCompletedPeriods: 10,
			TotalSkippedPeriods:   0,
			TotalCompletedTimes:   21,
			MostLongestStreak:     10,
			CurrentStreak:         10,
		}

		snapshot := entities.ProgressSnapshot{
			ProgressID: int64(currentProgress.Id),
			GoalID:     goal.Id,
			Username:   username,
		}

		mockStorage, mockUserUc, mockTransactor, mockTime := initFunc(t)

		mockTransactor.On("RunRepeatableRead", ctx, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
			fx := args.Get(1).(func(ctxTX context.Context) error)
			_ = fx(ctx)
		})

		mockTime.On("GetCurrentTime", ctx, username).Return(currentTime, nil)

		mockUserUc.On("GetUserByUsername", ctx, username).Return(entities.User{}, nil)
		mockStorage.On("GetHabitGoal", ctx, habitId).Return(goal, nil)

		mockStorage.On("GetCurrentSnapshot", ctx, username, goal.Id, currentTime).Return(snapshot, nil)
		mockStorage.On("GetProgressByID", ctx, snapshot.ProgressID).Return(currentProgress, nil)

		mockStorage.On("GetPreviousPeriodExecutionCount", ctx, goal, currentTime).Return(2, nil)
		mockStorage.On("GetCurrentPeriodExecutionCount", ctx, goal, currentTime).Return(0, nil)

		mockStorage.On("AddProgressLog", ctx, goal.Id, currentTime).Return(nil)
		mockStorage.On("UpdateProgressByID", ctx, updatedProgress).Return(nil)

		progressUC := New(mockUserUc, mockStorage, mockTransactor, mockTime)
		err := progressUC.AddHabitProgress(ctx, username, habitId)
		require.Nil(t, err)
	})

	t.Run("success: start from scratch: increase total completed times", func(t *testing.T) {
		progressToCreate := entities.Progress{
			Username:              username,
			TotalCompletedPeriods: 0,
			TotalSkippedPeriods:   0,
			TotalCompletedTimes:   1,
			MostLongestStreak:     0,
			CurrentStreak:         0,
		}

		mockStorage, mockUserUc, mockTransactor, mockTime := initFunc(t)

		mockTransactor.On("RunRepeatableRead", ctx, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
			fx := args.Get(1).(func(ctxTX context.Context) error)
			_ = fx(ctx)
		})

		mockTime.On("GetCurrentTime", ctx, username).Return(currentTime, nil)
		mockUserUc.On("GetUserByUsername", ctx, username).Return(entities.User{}, nil)

		mockStorage.On("GetHabitGoal", ctx, habitId).Return(goal, nil)

		mockStorage.On("GetProgressByTime", ctx, 1, username, currentTime).Return(entities.Progress{}, storage.ErrNotFound)
		mockStorage.On("GetCurrentProgress", ctx, 1).Return(entities.Progress{}, storage.ErrNotFound)

		mockStorage.On("GetPreviousPeriodExecutionCount", ctx, goal, currentTime).Return(0, nil)
		mockStorage.On("GetCurrentPeriodExecutionCount", ctx, goal, currentTime).Return(0, nil)
		mockStorage.On("AddHabitProgress", ctx, goal.Id, currentTime).Return(nil)

		mockStorage.On("CreateProgress", ctx, progressToCreate).Return(int64(1), nil)

		progressUC := New(mockUserUc, mockStorage, mockTransactor, mockTime)
		err := progressUC.AddHabitProgress(ctx, username, habitId)
		require.Nil(t, err)
	})

	t.Run("success: start from scratch: increase total completed periods", func(t *testing.T) {
		currentProgress := entities.Progress{
			Id:                    1,
			TotalCompletedPeriods: 0,
			TotalSkippedPeriods:   0,
			TotalCompletedTimes:   1,
			MostLongestStreak:     0,
			CurrentStreak:         0,
		}

		updatedProgress := entities.Progress{
			Id:                    1,
			TotalCompletedPeriods: 1,
			TotalSkippedPeriods:   0,
			TotalCompletedTimes:   2,
			MostLongestStreak:     1,
			CurrentStreak:         1,
		}

		mockStorage, mockUserUc, mockTransactor, mockTime := initFunc(t)

		mockTransactor.On("RunRepeatableRead", ctx, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
			fx := args.Get(1).(func(ctxTX context.Context) error)
			_ = fx(ctx)
		})
		mockTime.On("GetCurrentTime", ctx, username).Return(currentTime, nil)

		mockUserUc.On("GetUserByUsername", ctx, username).Return(entities.User{}, nil)
		mockStorage.On("GetHabitGoal", ctx, habitId).Return(goal, nil)

		mockStorage.On("GetProgressByTime", ctx, 1, username, currentTime).Return(entities.Progress{}, storage.ErrNotFound)
		mockStorage.On("GetCurrentProgress", ctx, 1).Return(currentProgress, nil)

		mockStorage.On("GetPreviousPeriodExecutionCount", ctx, goal, currentTime).Return(0, nil)
		mockStorage.On("GetCurrentPeriodExecutionCount", ctx, goal, currentTime).Return(1, nil)
		mockStorage.On("AddHabitProgress", ctx, goal.Id, currentTime).Return(nil)

		mockStorage.On("UpdateProgressByID", ctx, updatedProgress).Return(nil)

		progressUC := New(mockUserUc, mockStorage, mockTransactor, mockTime)
		err := progressUC.AddHabitProgress(ctx, username, habitId)
		require.Nil(t, err)
	})

	t.Run("success: with two days skips: update total completed times", func(t *testing.T) {
		currentProgress := entities.Progress{
			Id:                    1,
			TotalCompletedPeriods: 4,
			TotalSkippedPeriods:   2,
			TotalCompletedTimes:   8,
			MostLongestStreak:     4,
			CurrentStreak:         0,
		}

		updatedProgress := entities.Progress{
			Id:                    1,
			TotalCompletedPeriods: 4,
			TotalSkippedPeriods:   2,
			TotalCompletedTimes:   9,
			MostLongestStreak:     4,
			CurrentStreak:         0,
		}

		mockStorage, mockUserUc, mockTransactor, mockTime := initFunc(t)

		mockTransactor.On("RunRepeatableRead", ctx, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
			fx := args.Get(1).(func(ctxTX context.Context) error)
			_ = fx(ctx)
		})

		mockTime.On("GetCurrentTime", ctx, username).Return(currentTime, nil)
		mockUserUc.On("GetUserByUsername", ctx, username).Return(entities.User{}, nil)

		mockStorage.On("GetHabitGoal", ctx, habitId).Return(goal, nil)

		mockStorage.On("GetProgressByTime", ctx, 1, username, currentTime).Return(entities.Progress{}, storage.ErrNotFound)
		mockStorage.On("GetCurrentProgress", ctx, 1).Return(currentProgress, nil)

		mockStorage.On("GetPreviousPeriodExecutionCount", ctx, goal, currentTime).Return(0, nil)
		mockStorage.On("GetCurrentPeriodExecutionCount", ctx, goal, currentTime).Return(0, nil)
		mockStorage.On("AddHabitProgress", ctx, goal.Id, currentTime).Return(nil)

		mockStorage.On("UpdateProgressByID", ctx, updatedProgress).Return(nil)

		progressUC := New(mockUserUc, mockStorage, mockTransactor, mockTime)
		err := progressUC.AddHabitProgress(ctx, username, habitId)
		require.Nil(t, err)
	})

	t.Run("success: with two days skips: update total completed periods", func(t *testing.T) {
		currentProgress := entities.Progress{
			Id:                    1,
			TotalCompletedPeriods: 4,
			TotalSkippedPeriods:   2,
			TotalCompletedTimes:   9,
			MostLongestStreak:     4,
			CurrentStreak:         0,
		}

		updatedProgress := entities.Progress{
			Id:                    1,
			TotalCompletedPeriods: 5,
			TotalSkippedPeriods:   2,
			TotalCompletedTimes:   10,
			MostLongestStreak:     4,
			CurrentStreak:         1,
		}

		mockStorage, mockUserUc, mockTransactor, mockTime := initFunc(t)

		mockTransactor.On("RunRepeatableRead", ctx, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
			fx := args.Get(1).(func(ctxTX context.Context) error)
			_ = fx(ctx)
		})
		mockTime.On("GetCurrentTime", ctx, username).Return(currentTime, nil)

		mockUserUc.On("GetUserByUsername", ctx, username).Return(entities.User{}, nil)
		mockStorage.On("GetHabitGoal", ctx, habitId).Return(goal, nil)

		mockStorage.On("GetProgressByTime", ctx, 1, username, currentTime).Return(entities.Progress{}, storage.ErrNotFound)
		mockStorage.On("GetCurrentProgress", ctx, 1).Return(currentProgress, nil)

		mockStorage.On("GetPreviousPeriodExecutionCount", ctx, goal, currentTime).Return(0, nil)
		mockStorage.On("GetCurrentPeriodExecutionCount", ctx, goal, currentTime).Return(1, nil)

		mockStorage.On("AddHabitProgress", ctx, goal.Id, currentTime).Return(nil)

		mockStorage.On("UpdateProgressByID", ctx, updatedProgress).Return(nil)

		progressUC := New(mockUserUc, mockStorage, mockTransactor, mockTime)
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

		updatedProgress := entities.Progress{
			TotalCompletedPeriods: 30,
			TotalSkippedPeriods:   0,
			TotalCompletedTimes:   60,
			MostLongestStreak:     30,
			CurrentStreak:         30,
		}

		mockStorage, mockUserUc, mockTransactor, mockTime := initFunc(t)

		mockTransactor.On("RunRepeatableRead", ctx, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
			fx := args.Get(1).(func(ctxTX context.Context) error)
			_ = fx(ctx)
		})
		mockTime.On("GetCurrentTime", ctx, username).Return(currentTime, nil)

		mockUserUc.On("GetUserByUsername", ctx, username).Return(entities.User{}, nil)

		mockStorage.On("GetHabitGoal", ctx, habitId).Return(goal, nil)

		mockStorage.On("GetProgressByTime", ctx, 1, username, currentTime).Return(entities.Progress{}, storage.ErrNotFound)
		mockStorage.On("GetCurrentProgress", ctx, 1).Return(currentProgress, nil)

		mockStorage.On("GetPreviousPeriodExecutionCount", ctx, goal, currentTime).Return(2, nil)
		mockStorage.On("GetCurrentPeriodExecutionCount", ctx, goal, currentTime).Return(1, nil)

		mockStorage.On("AddHabitProgress", ctx, goal.Id, currentTime).Return(nil)

		mockStorage.On("UpdateProgressByID", ctx, updatedProgress).Return(nil)
		mockStorage.On("SetGoalCompleted", ctx, goal.Id).Return(nil)

		progressUC := New(mockUserUc, mockStorage, mockTransactor, mockTime)
		err := progressUC.AddHabitProgress(ctx, username, habitId)
		require.Nil(t, err)
	})

	t.Run("err goal is completed", func(t *testing.T) {
		completedGoal := entities.Goal{
			Id:                   1,
			FrequencyType:        entities.Daily,
			TimesPerFrequency:    2,
			TotalTrackingPeriods: 30,
			IsCompleted:          true,
		}

		mockStorage, mockUserUc, mockTransactor, mockTime := initFunc(t)

		mockTransactor.On("RunRepeatableRead", ctx, mock.Anything).Return(ErrGoalCompleted).Run(func(args mock.Arguments) {
			fx := args.Get(1).(func(ctxTX context.Context) error)
			_ = fx(ctx)
		})

		mockTime.On("GetCurrentTime", ctx, username).Return(currentTime, nil)

		mockUserUc.On("GetUserByUsername", ctx, username).Return(entities.User{}, nil)
		mockStorage.On("GetHabitGoal", ctx, habitId).Return(completedGoal, nil)
		progressUC := New(mockUserUc, mockStorage, mockTransactor, mockTime)
		err := progressUC.AddHabitProgress(ctx, username, habitId)
		require.ErrorIs(t, err, ErrGoalCompleted)
	})
}

func TestGetCurrentProgressForAllUserHabits(t *testing.T) {
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
				StartTrackingAt:      time.Now().AddDate(0, 0, -1),
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
				StartTrackingAt:      time.Now().AddDate(0, 0, -10),
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
				StartTrackingAt:      time.Now().AddDate(0, 0, -10),
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
				StartTrackingAt:      time.Now().AddDate(0, 0, -10),
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

		mockStorage, mockUserUc, _, mockTimeManager := initFunc(t)

		mockTimeManager.On("GetCurrentTime", ctx, username).Return(currentTime, nil)

		mockUserUc.On("GetUserByUsername", ctx, username).Return(entities.User{}, nil)

		mockStorage.On("GetAllUserHabitsWithGoals", ctx, username).Return([]entities.Habit{habit1, habit2, habit3, habit4}, nil)

		for _, testCase := range testCases {
			goal := testCase.habit.Goal
			mockStorage.On("GetCurrentPeriodExecutionCount", ctx, *goal, currentTime).Return(testCase.currentExecutionCount, nil)
		}

		progressUC := New(mockUserUc, mockStorage, nil, mockTimeManager)
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

func TestRecalculateFutureProgresses(t *testing.T) {
	t.Run("success daily", func(t *testing.T) {
		var (
			currentTime = time.Now()
			ctx         = context.Background()
			day2Time    = currentTime.AddDate(0, 0, 1)
			day3Time    = currentTime.AddDate(0, 0, 2)
			day4Time    = currentTime.AddDate(0, 0, 3)
			username    = "username"
		)

		prevGoal := entities.Goal{
			Id:                   1,
			FrequencyType:        entities.Daily,
			TimesPerFrequency:    2,
			TotalTrackingPeriods: 10,
			CreatedAt:            time.Now().AddDate(0, 0, -5),
		}

		newGoal := entities.Goal{
			Id:                   1,
			FrequencyType:        entities.Daily,
			TimesPerFrequency:    3,
			TotalTrackingPeriods: 10,
			CreatedAt:            time.Now().AddDate(0, 0, -5),
		}

		baseProgress := entities.Progress{
			Id:                    1,
			TotalCompletedPeriods: 5,
			TotalSkippedPeriods:   1,
			TotalCompletedTimes:   10,
			MostLongestStreak:     3,
			CurrentStreak:         3,
		}

		day2Progress := entities.Progress{
			Id:                    2,
			TotalCompletedPeriods: 6,
			TotalSkippedPeriods:   1,
			TotalCompletedTimes:   12,
			MostLongestStreak:     4,
			CurrentStreak:         4,
		}

		day3Progress := entities.Progress{
			Id:                    3,
			TotalCompletedPeriods: 7,
			TotalSkippedPeriods:   1,
			TotalCompletedTimes:   15,
			MostLongestStreak:     5,
			CurrentStreak:         5,
		}

		day4Progress := entities.Progress{
			Id:                    4,
			TotalCompletedPeriods: 7,
			TotalSkippedPeriods:   2,
			TotalCompletedTimes:   16,
			MostLongestStreak:     5,
			CurrentStreak:         1,
		}

		day1Snapshot := entities.ProgressSnapshot{
			Username:   username,
			ProgressID: int64(baseProgress.Id),
			CreatedAt:  currentTime,
			GoalID:     prevGoal.Id,
		}

		day2Snapshot := entities.ProgressSnapshot{
			Username:   username,
			ProgressID: int64(day2Progress.Id),
			CreatedAt:  day2Time,
			GoalID:     prevGoal.Id,
		}

		day3Snapshot := entities.ProgressSnapshot{
			Username:   username,
			ProgressID: int64(day3Progress.Id),
			CreatedAt:  day3Time,
			GoalID:     prevGoal.Id,
		}

		day4Snapshot := entities.ProgressSnapshot{
			Username:   username,
			ProgressID: int64(day4Progress.Id),
			CreatedAt:  day4Time,
			GoalID:     prevGoal.Id,
		}

		futureSnapshots := []entities.ProgressSnapshot{day2Snapshot, day3Snapshot, day4Snapshot}
		futureExecutions := []int{2, 3, 1}
		progressUpdates := []entities.Progress{
			{
				Id:                    day2Progress.Id,
				TotalCompletedPeriods: 4,
				TotalSkippedPeriods:   1,
				TotalCompletedTimes:   12,
				CurrentStreak:         2,
				MostLongestStreak:     2,
			},
			{
				Id:                    day3Progress.Id,
				TotalCompletedPeriods: 5,
				TotalSkippedPeriods:   1,
				TotalCompletedTimes:   15,
				CurrentStreak:         3,
				MostLongestStreak:     3,
			},
			{
				Id:                    day4Progress.Id,
				TotalCompletedPeriods: 5,
				TotalSkippedPeriods:   1,
				TotalCompletedTimes:   16,
				CurrentStreak:         3,
				MostLongestStreak:     3,
			},
		}

		mockStorage, _, _, _ := initFunc(t)

		mockStorage.On("GetCurrentSnapshot", ctx, username, prevGoal.Id, currentTime).Return(day1Snapshot, nil)
		mockStorage.On("GetProgressByID", ctx, day1Snapshot.ProgressID).Return(baseProgress, nil)
		mockStorage.On("GetCurrentPeriodExecutionCount", ctx, prevGoal, currentTime).Return(2, nil)
		mockStorage.On("GetFutureSnapshots", ctx, username, prevGoal.Id, currentTime).Return(futureSnapshots, nil)

		for j, futureSnapshot := range futureSnapshots {
			mockStorage.On("GetCurrentDayExecutionCount", ctx, prevGoal, futureSnapshot.CreatedAt).Return(futureExecutions[j], nil)
			mockStorage.On("GetCurrentPeriodExecutionCount", ctx, prevGoal, futureSnapshot.CreatedAt).Return(futureExecutions[j], nil)
			mockStorage.On("UpdateProgressByID", ctx, progressUpdates[j]).Return(nil)
		}

		i := Implementation{
			storage: mockStorage,
		}

		err := i.RecalculateFutureProgressesByGoalUpdate(ctx, username, prevGoal, newGoal, currentTime)
		require.Nil(t, err)
	})

	t.Run("success weekly", func(t *testing.T) {
		//var (
		//	currentTime = time.Now()
		//	ctx         = context.Background()
		//	day2Time    = currentTime.AddDate(0, 0, 1)
		//	day3Time    = currentTime.AddDate(0, 0, 2)
		//	day4Time    = currentTime.AddDate(0, 0, 3)
		//	day5Time    = currentTime.AddDate(0, 0, 4)
		//	day6Time    = currentTime.AddDate(0, 0, 5)
		//	day7Time    = currentTime.AddDate(0, 0, 6)
		//	username    = "username"
		//)
		//
		//prevGoal := entities.Goal{
		//	Id:                   1,
		//	FrequencyType:        entities.Weekly,
		//	TimesPerFrequency:    10,
		//	TotalTrackingPeriods: 4,
		//	CreatedAt:            time.Now().AddDate(0, 0, -14),
		//}
		//
		//newGoal := entities.Goal{
		//	Id:                   1,
		//	FrequencyType:        entities.Weekly,
		//	TimesPerFrequency:    14,
		//	TotalTrackingPeriods: 4,
		//	CreatedAt:            time.Now().AddDate(0, 0, -14),
		//}
		//
		//baseProgress := entities.Progress{
		//	Id:                    1,
		//	TotalCompletedPeriods: 2, // Represents 2 weeks completed
		//	TotalSkippedPeriods:   0,
		//	TotalCompletedTimes:   20,
		//	MostLongestStreak:     2,
		//	CurrentStreak:         2,
		//}

		//day2Progress := entities.Progress{Id: 2, TotalCompletedTimes: 22, TotalCompletedPeriods: 2, MostLongestStreak: 2, CurrentStreak: 2}
		//day3Progress := entities.Progress{Id: 3, TotalCompletedTimes: 24, TotalCompletedPeriods: 2, MostLongestStreak: 2, CurrentStreak: 2}
		//day4Progress := entities.Progress{Id: 4, TotalCompletedTimes: 26, TotalCompletedPeriods: 2, MostLongestStreak: 2, CurrentStreak: 2}
		//day5Progress := entities.Progress{Id: 5, TotalCompletedTimes: 28, TotalCompletedPeriods: 2, MostLongestStreak: 2, CurrentStreak: 2}
		//day6Progress := entities.Progress{Id: 6, TotalCompletedTimes: 30, TotalCompletedPeriods: 3, MostLongestStreak: 3, CurrentStreak: 3}
		//day7Progress := entities.Progress{Id: 7, TotalCompletedTimes: 30, TotalCompletedPeriods: 3, MostLongestStreak: 3, CurrentStreak: 3}

		//daySnapshots := []entities.ProgressSnapshot{
		//	{Username: username, ProgressID: int64(day2Progress.Id), CreatedAt: day2Time, GoalID: prevGoal.Id},
		//	{Username: username, ProgressID: int64(day3Progress.Id), CreatedAt: day3Time, GoalID: prevGoal.Id},
		//	{Username: username, ProgressID: int64(day4Progress.Id), CreatedAt: day4Time, GoalID: prevGoal.Id},
		//	{Username: username, ProgressID: int64(day5Progress.Id), CreatedAt: day5Time, GoalID: prevGoal.Id},
		//	{Username: username, ProgressID: int64(day6Progress.Id), CreatedAt: day6Time, GoalID: prevGoal.Id},
		//	{Username: username, ProgressID: int64(day7Progress.Id), CreatedAt: day7Time, GoalID: prevGoal.Id},
		//}
		//
		//dailyExecutions := []int{2, 2, 2, 2, 2, 0}
		//periodExecutions := []int{2, 4, 6, 8, 10, 10}

	})
}

func generateDays(n int) []time.Time {
	var days []time.Time
	for i := 0; i < n; i++ {
		days = append(days, time.Now().AddDate(0, 0, i))
	}
	return days

}
