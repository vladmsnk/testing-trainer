//go:generate mockery --dir . --name Storage --structname MockStorage --filename storage_mock.go --output . --outpkg=progress
//go:generate mockery --dir . --name UserUseCase --structname MockUserUseCase --filename user_mock.go --output . --outpkg=progress
//go:generate mockery --dir . --name Transactor --structname MockTransactor --filename transactor_mock.go --output . --outpkg=progress
//go:generate mockery --dir . --name TimeManager --structname MockTimeManager --filename time_manager_mock.go --output . --outpkg=progress
package progress_adder

import (
	"context"
	"errors"
	"fmt"
	"time"

	"testing_trainer/internal/entities"
	"testing_trainer/internal/storage"
)

var (
	ErrHabitNotFound = fmt.Errorf("habit not found")
	ErrGoalCompleted = fmt.Errorf("goal is already completed")
)

type UseCase interface {
	AddHabitProgress(ctx context.Context, username string, habitId int) error
}

type UserUseCase interface {
	GetUserByUsername(ctx context.Context, username string) (entities.User, error)
}

type Getter interface {
	GetProgressBySnapshot(ctx context.Context, goalID int, username string, currentTime time.Time) (entities.Progress, error)
}

type ProgressRecalculator interface {
	RecalculateFutureProgressesByGoalUpdate(ctx context.Context, username string, prevGoal, newGoal entities.Goal, currentTime time.Time) error
}

type Transactor interface {
	RunRepeatableRead(ctx context.Context, fx func(ctxTX context.Context) error) error
}

type Storage interface {
	AddProgressLog(ctx context.Context, goalId int, createdAt time.Time) error
	GetHabitGoal(ctx context.Context, habitId int) (entities.Goal, error)
	GetHabitById(ctx context.Context, username string, habitId int) (entities.Habit, error)
	SetGoalCompleted(ctx context.Context, goalId int) error
	GetPreviousPeriodExecutionCount(ctx context.Context, goal entities.Goal, currentTime time.Time) (int, error)
	GetCurrentPeriodExecutionCount(ctx context.Context, goal entities.Goal, currentTime time.Time) (int, error)
	UpdateProgressByID(ctx context.Context, progress entities.Progress) error
}

type TimeManager interface {
	GetCurrentTime(ctx context.Context, username string) (time.Time, error)
}

type Implementation struct {
	userUc               UserUseCase
	storage              Storage
	transactor           Transactor
	timeManager          TimeManager
	progressGetter       Getter
	progressRecalculator ProgressRecalculator
}

func New(
	userUc UserUseCase,
	storage Storage,
	progressGetter Getter,
	transactor Transactor,
	timeManager TimeManager,
	progressRecalculator ProgressRecalculator,
) *Implementation {
	return &Implementation{
		userUc:               userUc,
		storage:              storage,
		transactor:           transactor,
		timeManager:          timeManager,
		progressGetter:       progressGetter,
		progressRecalculator: progressRecalculator,
	}
}

func (i *Implementation) AddHabitProgress(ctx context.Context, username string, habitId int) error {
	err := i.transactor.RunRepeatableRead(ctx, func(ctxTX context.Context) error {
		currentTime, err := i.timeManager.GetCurrentTime(ctx, username)
		if err != nil {
			return fmt.Errorf("i.timeManager.GetCurrentTime: %w", err)
		}

		_, err = i.userUc.GetUserByUsername(ctx, username)
		if err != nil {
			return fmt.Errorf("i.userUc.GetUserByUsername: %w", err)
		}

		goal, err := i.storage.GetHabitGoal(ctx, habitId)
		if err != nil {
			if errors.Is(err, storage.ErrNotFound) {
				return ErrHabitNotFound
			}
			return fmt.Errorf("i.storage.GetHabitGoal: %w", err)
		}

		if goal.IsCompleted {
			return ErrGoalCompleted
		}

		currentProgress, err := i.progressGetter.GetProgressBySnapshot(ctx, goal.Id, username, currentTime)
		if err != nil {
			return fmt.Errorf("i.GetProgressBySnapshot: %w", err)
		}

		lastPeriodExecutionCount, err := i.storage.GetPreviousPeriodExecutionCount(ctx, goal, currentTime)
		if err != nil {
			return fmt.Errorf("i.storage.GetPreviousDayExecutionCount: %w", err)
		}

		currentExecutionCount, err := i.storage.GetCurrentPeriodExecutionCount(ctx, goal, currentTime)
		if err != nil {
			return fmt.Errorf("i.storage.GetTodayExecutionCount: %w", err)
		}

		currentExecutionCount += 1
		updatedProgress := currentProgress
		goalIsCompleted := false

		updatedProgress.TotalCompletedTimes = currentProgress.TotalCompletedTimes + 1

		// Check if the goal is completed for the current period
		if currentExecutionCount == goal.TimesPerFrequency {
			updatedProgress.TotalCompletedPeriods = currentProgress.TotalCompletedPeriods + 1

			// Streak logic: reset or increment the streak
			if lastPeriodExecutionCount >= goal.TimesPerFrequency {
				updatedProgress.CurrentStreak = currentProgress.CurrentStreak + 1
			} else {
				updatedProgress.CurrentStreak = 1
			}

			if updatedProgress.CurrentStreak > currentProgress.MostLongestStreak {
				updatedProgress.MostLongestStreak = updatedProgress.CurrentStreak
			}

			if updatedProgress.TotalCompletedPeriods == goal.TotalTrackingPeriods {
				goalIsCompleted = true
			}
		}

		err = i.storage.AddProgressLog(ctx, goal.Id, currentTime)
		if err != nil {
			return fmt.Errorf("i.storage.AddHabitProgress: %w", err)
		}

		err = i.storage.UpdateProgressByID(ctx, updatedProgress)
		if err != nil {
			return fmt.Errorf("i.storage.UpdateProgressByID: %w", err)
		}

		err = i.progressRecalculator.RecalculateFutureProgressesByGoalUpdate(ctx, username, goal, goal, currentTime)
		if err != nil {
			return fmt.Errorf("i.RecalculateFutureProgressesByGoalUpdate: %w", err)
		}
		// add record to table execution_times_per_period

		if goalIsCompleted {
			err := i.storage.SetGoalCompleted(ctx, goal.Id)
			if err != nil {
				return fmt.Errorf("i.storage.SetGoalCompleted: %w", err)
			}
		}
		return nil
	})

	return err
}
