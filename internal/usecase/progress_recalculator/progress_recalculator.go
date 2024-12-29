//go:generate mockery --dir . --name Storage --structname MockStorage --filename storage_mock.go --output . --outpkg=progress
//go:generate mockery --dir . --name UserUseCase --structname MockUserUseCase --filename user_mock.go --output . --outpkg=progress
//go:generate mockery --dir . --name Transactor --structname MockTransactor --filename transactor_mock.go --output . --outpkg=progress
//go:generate mockery --dir . --name TimeManager --structname MockTimeManager --filename time_manager_mock.go --output . --outpkg=progress
package progress_recalculator

import (
	"context"
	"fmt"
	"time"

	"testing_trainer/internal/entities"
)

type UseCase interface {
	RecalculateFutureProgressesByGoalUpdate(ctx context.Context, username string, prevGoal, newGoal entities.Goal, currentTime time.Time) error
}

type UserUseCase interface {
	GetUserByUsername(ctx context.Context, username string) (entities.User, error)
}

type Getter interface {
	GetProgressBySnapshot(ctx context.Context, goalID int, username string, currentTime time.Time) (entities.Progress, error)
}

type Transactor interface {
	RunRepeatableRead(ctx context.Context, fx func(ctxTX context.Context) error) error
}

type Storage interface {
	GetCurrentPeriodExecutionCount(ctx context.Context, goal entities.Goal, currentTime time.Time) (int, error)
	UpdateProgressByID(ctx context.Context, progress entities.Progress) error
	GetFutureSnapshots(ctx context.Context, username string, goalID int, currentTime time.Time) ([]entities.ProgressSnapshot, error)
	GetCurrentDayExecutionCount(ctx context.Context, goal entities.Goal, currentTime time.Time) (int, error)
}

type TimeManager interface {
	GetCurrentTime(ctx context.Context, username string) (time.Time, error)
}

type Implementation struct {
	userUc         UserUseCase
	storage        Storage
	transactor     Transactor
	timeManager    TimeManager
	progressGetter Getter
}

func NewRecalculator(
	userUc UserUseCase,
	storage Storage,
	progressGetter Getter,
	transactor Transactor,
	timeManager TimeManager,
) *Implementation {
	return &Implementation{
		userUc:         userUc,
		storage:        storage,
		transactor:     transactor,
		timeManager:    timeManager,
		progressGetter: progressGetter,
	}
}

func (i *Implementation) RecalculateFutureProgressesByGoalUpdate(ctx context.Context, username string, prevGoal, newGoal entities.Goal, currentTime time.Time) error {
	baseProgress, err := i.progressGetter.GetProgressBySnapshot(ctx, prevGoal.Id, username, currentTime)
	if err != nil {
		return fmt.Errorf("i.GetProgressBySnapshot: %w", err)
	}

	currentPeriodExecutionCount, err := i.storage.GetCurrentPeriodExecutionCount(ctx, prevGoal, currentTime)
	if err != nil {
		return fmt.Errorf("i.storage.GetCurrentPeriodExecutionCount: %w", err)
	}

	if currentPeriodExecutionCount < newGoal.TimesPerFrequency && currentPeriodExecutionCount >= prevGoal.TimesPerFrequency && baseProgress.TotalCompletedPeriods > 0 {
		baseProgress.TotalCompletedPeriods -= 1
		if baseProgress.CurrentStreak == baseProgress.MostLongestStreak {
			baseProgress.MostLongestStreak -= 1
			baseProgress.CurrentStreak -= 1
		} else {
			baseProgress.CurrentStreak -= 1
		}

		err = i.storage.UpdateProgressByID(ctx, baseProgress)
		if err != nil {
			return fmt.Errorf("i.storage.UpdateProgressByID: %w", err)
		}
	}

	snapshots, err := i.storage.GetFutureSnapshots(ctx, username, prevGoal.Id, currentTime)
	if err != nil {
		return fmt.Errorf("i.storage.GetFutureSnapshots: %w", err)
	}

	var baseProgresses []entities.Progress

	for j, snapshot := range snapshots {
		var basep entities.Progress

		if j == 0 {
			basep = baseProgress.DeepCopy()
		} else {
			basep = baseProgresses[j-1].DeepCopy()
		}

		currentDayExecutionCount, err := i.storage.GetCurrentDayExecutionCount(ctx, prevGoal, snapshot.CreatedAt)
		if err != nil {
			return fmt.Errorf("i.storage.GetCurrentDayExecutionCount: %w", err)
		}

		currentPeriodExecCnt, err := i.storage.GetCurrentPeriodExecutionCount(ctx, prevGoal, snapshot.CreatedAt)
		if err != nil {
			return fmt.Errorf("i.storage.GetCurrentPeriodExecutionCount: %w", err)
		}

		if currentPeriodExecCnt < newGoal.TimesPerFrequency {
			basep.TotalCompletedTimes += currentDayExecutionCount
		} else if currentPeriodExecCnt == newGoal.TimesPerFrequency {
			basep.TotalCompletedTimes += currentDayExecutionCount
			basep.TotalCompletedPeriods += 1
			basep.CurrentStreak += 1
			if basep.CurrentStreak > basep.MostLongestStreak {
				basep.MostLongestStreak = basep.CurrentStreak
			}
		} else {
			continue
		}

		basep.Id = int(snapshot.ProgressID)
		baseProgresses = append(baseProgresses, basep)

		err = i.storage.UpdateProgressByID(ctx, basep)
		if err != nil {
			return fmt.Errorf("i.storage.UpdateProgressByID: %w", err)
		}
	}

	return nil
}
