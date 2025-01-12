//go:generate mockery --dir . --name Storage --structname MockStorage --filename storage_mock.go --output . --outpkg=progress
//go:generate mockery --dir . --name UserUseCase --structname MockUserUseCase --filename user_mock.go --output . --outpkg=progress
//go:generate mockery --dir . --name Transactor --structname MockTransactor --filename transactor_mock.go --output . --outpkg=progress
//go:generate mockery --dir . --name TimeManager --structname MockTimeManager --filename time_manager_mock.go --output . --outpkg=progress
package progress_recalculator

import (
	"context"
	"errors"
	"fmt"
	"testing_trainer/internal/storage"
	"time"

	"testing_trainer/internal/entities"
)

type UseCase interface {
	RecalculateFutureProgressesByGoalUpdate(ctx context.Context, username string, prevGoal, newGoal entities.Goal, currentTime time.Time) error
	RecalculateFutureProgresses(ctx context.Context, username string, prevGoal, newGoal entities.Goal, currentTime time.Time) error
	RecalculateCurrentProgress(ctx context.Context, username string, prevGoal, newGoal entities.Goal, currentTime time.Time) error
}

type UserUseCase interface {
	GetUserByUsername(ctx context.Context, username string) (entities.User, error)
}

type Getter interface {
	GetProgressBySnapshot(ctx context.Context, goal entities.Goal, username string, currentTime time.Time) (entities.Progress, error)
}

type Transactor interface {
	RunRepeatableRead(ctx context.Context, fx func(ctxTX context.Context) error) error
}

type Storage interface {
	GetPreviousPeriodExecutionCount(ctx context.Context, goal entities.Goal, currentTime time.Time) (int, error)
	GetCurrentPeriodExecutionCount(ctx context.Context, goal entities.Goal, currentTime time.Time) (int, error)
	UpdateProgressByID(ctx context.Context, progress entities.Progress) error
	GetFutureSnapshots(ctx context.Context, username string, goalID int, currentTime time.Time) ([]entities.ProgressSnapshot, error)
	GetCurrentDayExecutionCount(ctx context.Context, goal entities.Goal, currentDayStartTime, currentDayEndTime time.Time) (int, error)
	GetTimeOfMostRecentSnapshot(ctx context.Context, goalID int) (time.Time, error)
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
	baseProgress, err := i.progressGetter.GetProgressBySnapshot(ctx, prevGoal, username, currentTime)
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

		currentDayExecutionCount, err := i.storage.GetCurrentDayExecutionCount(ctx, prevGoal, snapshot.CreatedAt, time.Now())
		if err != nil {
			return fmt.Errorf("i.storage.GetCurrentDayExecutionCount: %w", err)
		}

		currentPeriodExecCnt, err := i.storage.GetCurrentPeriodExecutionCount(ctx, prevGoal, snapshot.CreatedAt)
		if err != nil {
			return fmt.Errorf("i.storage.GetCurrentPeriodExecutionCount: %w", err)
		}

		if currentPeriodExecCnt < newGoal.TimesPerFrequency {
			basep.TotalCompletedTimes += currentDayExecutionCount
		} else if currentPeriodExecCnt >= newGoal.TimesPerFrequency {
			basep.TotalCompletedTimes += currentDayExecutionCount
			basep.TotalCompletedPeriods += 1
			basep.CurrentStreak += 1
			if basep.CurrentStreak > basep.MostLongestStreak {
				basep.MostLongestStreak = basep.CurrentStreak
			}
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

func (i *Implementation) RecalculateCurrentProgress(ctx context.Context, username string, prevGoal, newGoal entities.Goal, currentTime time.Time) error {
	currentPeriod := prevGoal.GetCurrentPeriod(currentTime)

	var previousPeriodEndTime time.Time

	if currentPeriod == 0 {
		previousPeriodEndTime = currentTime.AddDate(0, 0, -1)
	} else {
		_, previousPeriodEndTime = storage.CalculatePeriodRange(prevGoal.StartTrackingAt, newGoal.FrequencyType, currentPeriod-1)
	}

	baseProgress, err := i.progressGetter.GetProgressBySnapshot(ctx, prevGoal, username, previousPeriodEndTime)
	if err != nil {
		return fmt.Errorf("i.progressGetter.GetProgressBySnapshot: %w", err)
	}

	currentProgress, err := i.progressGetter.GetProgressBySnapshot(ctx, prevGoal, username, currentTime)
	if err != nil {
		return fmt.Errorf("i.progressGetter.GetProgressBySnapshot: %w", err)
	}

	currentPeriodExecCnt, err := i.storage.GetCurrentPeriodExecutionCount(ctx, newGoal, currentTime)
	if err != nil {
		return fmt.Errorf("i.storage.GetCurrentPeriodExecutionCount: %w", err)
	}

	baseProgress.Id = currentProgress.Id
	if currentPeriodExecCnt < newGoal.TimesPerFrequency {
		baseProgress.TotalCompletedTimes += currentPeriodExecCnt

		err = i.storage.UpdateProgressByID(ctx, baseProgress)
		if err != nil {
			return fmt.Errorf("i.storage.UpdateProgressByID: %w", err)
		}
	} else if currentPeriodExecCnt >= newGoal.TimesPerFrequency {
		baseProgress.TotalCompletedTimes += currentPeriodExecCnt
		baseProgress.TotalCompletedPeriods += 1
		baseProgress.CurrentStreak += 1
		if baseProgress.CurrentStreak > baseProgress.MostLongestStreak {
			baseProgress.MostLongestStreak = baseProgress.CurrentStreak
		}

		err = i.storage.UpdateProgressByID(ctx, baseProgress)
		if err != nil {
			return fmt.Errorf("i.storage.UpdateProgressByID: %w", err)
		}
	}

	return nil
}

func (i *Implementation) RecalculateFutureProgresses(ctx context.Context, username string, prevGoal, newGoal entities.Goal, currentTime time.Time) error {
	baseProgress, err := i.progressGetter.GetProgressBySnapshot(ctx, prevGoal, username, currentTime)
	if err != nil {
		return fmt.Errorf("i.progressGetter.GetProgressBySnapshot: %w", err)
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

		currentDayExecutionCount, err := i.storage.GetCurrentDayExecutionCount(ctx, prevGoal, snapshot.StartBound, snapshot.EndBound)
		if err != nil {
			return fmt.Errorf("i.storage.GetCurrentDayExecutionCount: %w", err)
		}

		currentPeriodExecCnt, err := i.storage.GetCurrentPeriodExecutionCount(ctx, prevGoal, snapshot.StartBound)
		if err != nil {
			return fmt.Errorf("i.storage.GetCurrentPeriodExecutionCount: %w", err)
		}

		basep.Id = int(snapshot.ProgressID)

		if currentPeriodExecCnt < newGoal.TimesPerFrequency {
			basep.TotalCompletedTimes += currentDayExecutionCount

			err = i.storage.UpdateProgressByID(ctx, basep)
			if err != nil {
				return fmt.Errorf("i.storage.UpdateProgressByID: %w", err)
			}
		} else if currentPeriodExecCnt >= newGoal.TimesPerFrequency {
			basep.TotalCompletedTimes += currentDayExecutionCount
			basep.TotalCompletedPeriods += 1
			basep.CurrentStreak += 1
			if basep.CurrentStreak > basep.MostLongestStreak {
				basep.MostLongestStreak = basep.CurrentStreak
			}

			err = i.storage.UpdateProgressByID(ctx, basep)
			if err != nil {
				return fmt.Errorf("i.storage.UpdateProgressByID: %w", err)
			}
		}

		baseProgresses = append(baseProgresses, basep)
	}

	return nil
}

func (i *Implementation) RecalculateAllProgressesForGoal(ctx context.Context, username string, goal entities.Goal, startTime time.Time) error {
	currentDayStartTime := startTime
	currentDayEndTime := currentDayStartTime.Add(time.Hour * 23).Add(time.Minute * 59)

	lastSnapshotTime, err := i.storage.GetTimeOfMostRecentSnapshot(ctx, goal.Id)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			lastSnapshotTime = goal.StartTrackingAt
		} else {
			return fmt.Errorf("i.storage.GetTimeOfMostRecentSnapshot: %w", err)
		}
	}

	stopTrackingTime := lastSnapshotTime.AddDate(0, 0, 1)

	for currentDayEndTime.Before(stopTrackingTime) {

		err := i.RecalculateProgressForSpecificTime(ctx, username, goal, currentDayStartTime, currentDayEndTime)
		if err != nil {
			return fmt.Errorf("i.RecalculateProgressForSpecificTime: %w", err)
		}

		currentDayStartTime = currentDayStartTime.AddDate(0, 0, 1)
		currentDayEndTime = currentDayEndTime.AddDate(0, 0, 1)
	}

	return nil
}

func (i *Implementation) RecalculateProgressForSpecificTime(ctx context.Context, username string, goal entities.Goal, currentDayStartTime, currentDayEndTime time.Time) error {
	err := i.transactor.RunRepeatableRead(ctx, func(ctxTX context.Context) error {

		currentPeriod := goal.GetCurrentPeriod(currentDayEndTime)

		var (
			previousPeriodStartTime time.Time
			previousDayEndTime      = currentDayEndTime.AddDate(0, 0, -1)
		)

		if currentPeriod == 0 {
			previousPeriodStartTime = goal.StartTrackingAt.AddDate(0, 0, -1)
		} else {
			previousPeriodStartTime, _ = storage.CalculatePeriodRange(goal.StartTrackingAt, goal.FrequencyType, currentPeriod-1)
		}

		baseProgress, err := i.progressGetter.GetProgressBySnapshot(ctx, goal, username, previousPeriodStartTime)
		if err != nil {
			return fmt.Errorf("i.progressGetter.GetProgressBySnapshot: %w", err)
		}

		previousDayProgress, err := i.progressGetter.GetProgressBySnapshot(ctx, goal, username, previousDayEndTime)
		if err != nil {
			return fmt.Errorf("i.progressGetter.GetProgressBySnapshot: %w", err)
		}

		currentDayProgress, err := i.progressGetter.GetProgressBySnapshot(ctx, goal, username, currentDayEndTime)
		if err != nil {
			return fmt.Errorf("i.progressGetter.GetProgressBySnapshot: %w", err)
		}

		lastPeriodExecutionCount, err := i.storage.GetPreviousPeriodExecutionCount(ctx, goal, currentDayEndTime)
		if err != nil {
			return fmt.Errorf("i.storage.GetPreviousDayExecutionCount: %w", err)
		}

		currentDayExecutionCount, err := i.storage.GetCurrentDayExecutionCount(ctx, goal, currentDayStartTime, currentDayEndTime)
		if err != nil {
			return fmt.Errorf("i.storage.GetCurrentDayExecutionCount: %w", err)
		}

		currentPeriodExecutionCount, err := i.storage.GetCurrentPeriodExecutionCount(ctx, goal, currentDayEndTime)
		if err != nil {
			return fmt.Errorf("i.storage.GetTodayExecutionCount: %w", err)
		}

		baseProgress.TotalCompletedTimes = previousDayProgress.TotalCompletedTimes + currentDayExecutionCount

		if currentPeriodExecutionCount >= goal.TimesPerFrequency {
			baseProgress.TotalCompletedPeriods += 1

			if lastPeriodExecutionCount >= goal.TimesPerFrequency {
				baseProgress.CurrentStreak = baseProgress.CurrentStreak + 1
			} else {
				baseProgress.CurrentStreak = 1
			}

			if baseProgress.CurrentStreak > baseProgress.MostLongestStreak {
				baseProgress.MostLongestStreak = baseProgress.CurrentStreak
			}
		}

		baseProgress.Id = currentDayProgress.Id

		err = i.storage.UpdateProgressByID(ctxTX, baseProgress)
		if err != nil {
			return fmt.Errorf("i.storage.UpdateProgressByID: %w", err)
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}
