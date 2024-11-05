package progress

import (
	"context"
	"errors"
	"fmt"

	"testing_trainer/internal/entities"
	"testing_trainer/internal/storage"
)

//go:generate mockery --dir . --name Storage --structname MockStorage --filename storage_mock.go --output . --outpkg=progress
//go:generate mockery --dir . --name UserUseCase --structname MockUserUseCase --filename user_mock.go --output . --outpkg=progress
//go:generate mockery --dir . --name Transactor --structname MockTransactor --filename transactor_mock.go --output . --outpkg=progress

var (
	ErrHabitGoalNotFound = fmt.Errorf("habit goal not found")
)

type UseCase interface {
	GetHabitProgress(ctx context.Context, username string, habitId int) (entities.ProgressWithGoal, error)
	AddHabitProgress(ctx context.Context, username string, habitId int) error
}

type UserUseCase interface {
	GetUserByUsername(ctx context.Context, username string) (entities.User, error)
}

type Transactor interface {
	RunRepeatableRead(ctx context.Context, fx func(ctxTX context.Context) error) error
}

type Storage interface {
	AddHabitProgress(ctx context.Context, goalId int) error
	GetHabitGoal(ctx context.Context, habitId int) (entities.Goal, error)
	GetCurrentPeriodExecutionCount(ctx context.Context, goalId int, frequencyType entities.FrequencyType) (int, error)
	GetCurrentProgress(ctx context.Context, goalId int) (entities.Progress, error)
	UpdateGoalStat(ctx context.Context, goalId int, progress entities.Progress) error
	GetPreviousPeriodExecutionCount(ctx context.Context, goalId int, frequencyType entities.FrequencyType) (int, error)
	SetGoalCompleted(ctx context.Context, goalId int) error
}

type Implementation struct {
	userUc     UserUseCase
	storage    Storage
	transactor Transactor
}

func New(userUc UserUseCase, storage Storage, transactor Transactor) *Implementation {
	return &Implementation{
		userUc:     userUc,
		storage:    storage,
		transactor: transactor,
	}
}

func (i *Implementation) GetHabitProgress(ctx context.Context, username string, habitId int) (entities.ProgressWithGoal, error) {
	_, err := i.userUc.GetUserByUsername(ctx, username)
	if err != nil {
		return entities.ProgressWithGoal{}, fmt.Errorf("i.userUc.GetUserByUsername: %w", err)
	}

	goal, err := i.storage.GetHabitGoal(ctx, habitId)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return entities.ProgressWithGoal{}, ErrHabitGoalNotFound
		}
		return entities.ProgressWithGoal{}, fmt.Errorf("i.storage.GetHabitGoal: %w", err)
	}

	progress, err := i.storage.GetCurrentProgress(ctx, goal.Id)
	if err != nil {
		return entities.ProgressWithGoal{}, fmt.Errorf("i.storage.GetCurrentProgress: %w", err)
	}

	return entities.ProgressWithGoal{
		Progress: progress,
		Goal:     goal,
	}, nil
}

func (i *Implementation) AddHabitProgress(ctx context.Context, username string, habitId int) error {
	err := i.transactor.RunRepeatableRead(ctx, func(ctxTX context.Context) error {
		_, err := i.userUc.GetUserByUsername(ctx, username)
		if err != nil {
			return fmt.Errorf("i.userUc.GetUserByUsername: %w", err)
		}

		goal, err := i.storage.GetHabitGoal(ctx, habitId)
		if err != nil {
			if errors.Is(err, storage.ErrNotFound) {
				return ErrHabitGoalNotFound
			}
			return fmt.Errorf("i.storage.GetHabitGoal: %w", err)
		}

		currentProgress, err := i.storage.GetCurrentProgress(ctx, goal.Id)
		if err != nil {
			return fmt.Errorf("i.storage.GetCurrentProgress: %w", err)
		}

		lastPeriodExecutionCount, err := i.storage.GetPreviousPeriodExecutionCount(ctx, goal.Id, goal.FrequencyType)
		if err != nil {
			return fmt.Errorf("i.storage.GetPreviousDayExecutionCount: %w", err)
		}

		currentExecutionCount, err := i.storage.GetCurrentPeriodExecutionCount(ctx, goal.Id, goal.FrequencyType)
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

		err = i.storage.AddHabitProgress(ctx, goal.Id)
		if err != nil {
			return fmt.Errorf("i.storage.AddHabitProgress: %w", err)
		}

		err = i.storage.UpdateGoalStat(ctx, goal.Id, updatedProgress)
		if err != nil {
			return fmt.Errorf("i.storage.UpdateGoalStat: %w", err)
		}

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
func (i *Implementation) AsyncUpdateGoalStat(ctx context.Context) {

}
