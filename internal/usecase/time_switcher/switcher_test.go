package time_switcher

import (
	"context"
	"github.com/stretchr/testify/require"
	"testing"
	"testing_trainer/internal/entities"
	"testing_trainer/internal/storage"
	"time"
)

func TestSwitchToNextDay(t *testing.T) {
	var (
		ctx         = context.Background()
		currentTime = time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		nextDayTime = time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)
		username    = "username"
	)

	initFunc := func(t *testing.T) (*MockStorage, *MockTimeManager) {
		mockStorage := NewMockStorage(t)
		mockTimeManager := NewMockTimeManager(t)

		return mockStorage, mockTimeManager
	}

	t.Run("Success", func(t *testing.T) {
		currentProgresses := []entities.Progress{
			{
				Id:                    1,
				GoalID:                1,
				TotalCompletedPeriods: 3,
				CurrentStreak:         2,
				TotalSkippedPeriods:   1,
				MostLongestStreak:     2,
				Username:              username,
			},
			{
				Id:                    2,
				GoalID:                2,
				TotalCompletedPeriods: 3,
				CurrentStreak:         3,
				TotalSkippedPeriods:   0,
				MostLongestStreak:     3,
				Username:              username,
			},
		}

		copiedProgresses := []entities.Progress{
			{
				Id:                    1,
				GoalID:                1,
				TotalCompletedPeriods: 3,
				CurrentStreak:         2,
				TotalSkippedPeriods:   1,
				MostLongestStreak:     2,
				Username:              username,
				CreatedAt:             nextDayTime,
				UpdatedAt:             nextDayTime,
			},
			{
				Id:                    2,
				GoalID:                2,
				TotalCompletedPeriods: 3,
				CurrentStreak:         3,
				TotalSkippedPeriods:   0,
				MostLongestStreak:     3,
				Username:              username,
				CreatedAt:             nextDayTime,
				UpdatedAt:             nextDayTime,
			},
		}

		nesProgressIDs := []int64{3, 4}

		createdSnapshot := entities.ProgressSnapshot{
			Username:           username,
			CurrentProgressIDs: []int64{3, 4},
			CreatedAt:          nextDayTime,
		}

		mockStorage, mockTimeManager := initFunc(t)

		mockTimeManager.On("GetCurrentTime", ctx, username).Return(currentTime, nil).Times(1)
		mockTimeManager.On("SetTimeOffset", ctx, username, 1).Return(nil).Times(1)
		mockTimeManager.On("GetCurrentTime", ctx, username).Return(nextDayTime, nil).Times(1)

		mockStorage.On("GetSnapshotForTheTime", ctx, username, nextDayTime).Return(entities.ProgressSnapshot{}, storage.ErrNotFound)
		mockStorage.On("GetSnapshotForTheTime", ctx, username, currentTime).Return(entities.ProgressSnapshot{}, storage.ErrNotFound)
		mockStorage.On("GetProgressesForAllGoals", ctx, username, []int64{}).Return(currentProgresses, nil)

		for i, progress := range copiedProgresses {
			mockStorage.On("CreateProgress", ctx, progress).Return(nesProgressIDs[i], nil)
		}

		mockStorage.On("CreateProgressSnapshot", ctx, createdSnapshot).Return(nil)

		uc := Implementation{
			timeManager: mockTimeManager,
		}

		err := uc.SwitchToNextDay(ctx, username)
		require.Nil(t, err)
	})
}
