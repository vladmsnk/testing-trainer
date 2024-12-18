package goals_checker

import (
	"context"
	"testing"
	"testing_trainer/internal/storage"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"testing_trainer/internal/entities"
)

var (
	initFunc = func(t *testing.T) (*MockStorage, *MockTransactor, *MockTimeManager) {
		mockStorage := NewMockStorage(t)
		mockTransactor := NewMockTransactor(t)
		mockTimeManager := NewMockTimeManager(t)

		return mockStorage, mockTransactor, mockTimeManager
	}

	currentTime = time.Now().UTC()
)

func TestCheckGoals(t *testing.T) {
	username := "username"

	goalsToCheck := []entities.Goal{{
		Id:                   1,
		Username:             username,
		FrequencyType:        entities.Weekly,
		TimesPerFrequency:    3,
		TotalTrackingPeriods: 4,
		IsActive:             true,
		CreatedAt:            currentTime.AddDate(0, 0, -7),
		NextCheckDate:        currentTime.Add(-time.Hour),
		StartTrackingAt:      currentTime.AddDate(0, 0, -7),
	},
	}

	nextCheckDate := goalsToCheck[0].NextCheckDate.Add(7 * 24 * time.Hour)

	ctx := context.Background()

	t.Run("success: increase skipped days", func(t *testing.T) {
		t.Parallel()

		progress := entities.Progress{
			TotalCompletedTimes:   5,
			TotalCompletedPeriods: 1,
			CurrentStreak:         1,
			MostLongestStreak:     1,
			TotalSkippedPeriods:   0,
		}

		progressToUpdate := entities.Progress{
			TotalCompletedTimes:   5,
			TotalCompletedPeriods: 1,
			CurrentStreak:         0,
			MostLongestStreak:     1,
			TotalSkippedPeriods:   1,
		}

		mockStorage, mockTransactor, mockTimeManager := initFunc(t)

		mockTimeManager.On("GetCurrentTime", ctx, goalsCheckerUser).Return(currentTime, nil)

		mockStorage.On("GetAllGoalsNeedCheck", ctx, currentTime).Return(goalsToCheck, nil)

		mockTransactor.On("RunRepeatableRead", ctx, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
			fx := args.Get(1).(func(ctxTX context.Context) error)
			_ = fx(ctx)
		})

		mockStorage.On("GetProgressByTime", ctx, 1, username, currentTime).Return(entities.Progress{}, storage.ErrNotFound)
		mockStorage.On("GetCurrentProgress", ctx, 1).Return(progress, nil)

		mockStorage.On("GetPreviousPeriodExecutionCount", ctx, goalsToCheck[0], currentTime).Return(2, nil)

		mockStorage.On("SetGoalNextCheckDate", ctx, goalsToCheck[0].Id, nextCheckDate).Return(nil)
		mockStorage.On("UpdateProgressByID", ctx, progressToUpdate).Return(nil)

		checker := NewChecker(mockStorage, mockTransactor, mockTimeManager)

		err := checker.CheckGoals(ctx)
		require.Nil(t, err)
	})

	t.Run("success: do not change skipped days", func(t *testing.T) {
		t.Parallel()

		progress := entities.Progress{
			TotalCompletedTimes:   6,
			TotalCompletedPeriods: 2,
			CurrentStreak:         2,
			MostLongestStreak:     2,
			TotalSkippedPeriods:   0,
		}

		mockStorage, mockTransactor, mockTimeManager := initFunc(t)

		mockTimeManager.On("GetCurrentTime", ctx, goalsCheckerUser).Return(currentTime, nil)

		mockStorage.On("GetAllGoalsNeedCheck", ctx, currentTime).Return(goalsToCheck, nil)

		mockTransactor.On("RunRepeatableRead", ctx, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
			fx := args.Get(1).(func(ctxTX context.Context) error)
			_ = fx(ctx)
		})

		mockStorage.On("GetProgressByTime", ctx, 1, username, currentTime).Return(entities.Progress{}, storage.ErrNotFound)
		mockStorage.On("GetCurrentProgress", ctx, 1).Return(progress, nil)

		mockStorage.On("GetPreviousPeriodExecutionCount", ctx, goalsToCheck[0], currentTime).Return(3, nil)

		mockStorage.On("SetGoalNextCheckDate", ctx, goalsToCheck[0].Id, nextCheckDate).Return(nil)

		mockStorage.On("UpdateProgressByID", ctx, progress).Return(nil)

		checker := NewChecker(mockStorage, mockTransactor, mockTimeManager)

		err := checker.CheckGoals(ctx)
		require.Nil(t, err)
	})
}
