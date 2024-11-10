package goals_checker

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"testing_trainer/internal/entities"
)

func TestCheckGoals(t *testing.T) {
	goalsToCheck := []entities.Goal{{
		Id:                   1,
		FrequencyType:        entities.Weekly,
		TimesPerFrequency:    3,
		TotalTrackingPeriods: 4,
		IsActive:             true,
		CreatedAt:            time.Now().UTC().AddDate(0, 0, -7),
		NextCheckDate:        time.Now().UTC().Add(-time.Hour),
	},
	}

	nextCheckDate := goalsToCheck[0].NextCheckDate.Add(7 * 24 * time.Hour)

	ctx := context.Background()

	initFunc := func(t *testing.T) (*MockStorage, *MockTransactor) {
		mockStorage := NewMockStorage(t)
		mockTransactor := NewMockTransactor(t)

		return mockStorage, mockTransactor
	}

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

		mockStorage, mockTransactor := initFunc(t)

		mockStorage.On("GetAllGoalsNeedCheck", ctx).Return(goalsToCheck, nil)

		mockTransactor.On("RunRepeatableRead", ctx, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
			fx := args.Get(1).(func(ctxTX context.Context) error)
			_ = fx(ctx)
		})

		mockStorage.On("GetCurrentProgress", ctx, goalsToCheck[0].Id).Return(progress, nil)

		mockStorage.On("GetPreviousPeriodExecutionCount", ctx, goalsToCheck[0].Id, goalsToCheck[0].FrequencyType, goalsToCheck[0].CreatedAt, 1).Return(2, nil)

		mockStorage.On("UpdateGoalStat", ctx, goalsToCheck[0].Id, progressToUpdate).Return(nil)

		mockStorage.On("SetGoalNextCheckDate", ctx, goalsToCheck[0].Id, nextCheckDate).Return(nil)

		checker := NewChecker(mockStorage, mockTransactor)

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

		mockStorage, mockTransactor := initFunc(t)

		mockStorage.On("GetAllGoalsNeedCheck", ctx).Return(goalsToCheck, nil)
		mockTransactor.On("RunRepeatableRead", ctx, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
			fx := args.Get(1).(func(ctxTX context.Context) error)
			_ = fx(ctx)
		})

		mockStorage.On("GetCurrentProgress", ctx, goalsToCheck[0].Id).Return(progress, nil)

		mockStorage.On("GetPreviousPeriodExecutionCount", ctx, goalsToCheck[0].Id, goalsToCheck[0].FrequencyType, goalsToCheck[0].CreatedAt, 1).Return(3, nil)

		mockStorage.On("SetGoalNextCheckDate", ctx, goalsToCheck[0].Id, nextCheckDate).Return(nil)

		checker := NewChecker(mockStorage, mockTransactor)

		err := checker.CheckGoals(ctx)
		require.Nil(t, err)
	})
}
