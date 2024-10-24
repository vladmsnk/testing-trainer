package progress

import (
	"context"
	"errors"
	"fmt"
	"testing_trainer/internal/storage"

	"testing_trainer/internal/entities"
)

//go:generate mockery --dir . --name Storage --structname MockStorage --filename storage_mock.go --output . --outpkg=progress
//go:generate mockery --dir . --name UserUseCase --structname MockUserUseCase --filename user_mock.go --output . --outpkg=progress

var (
	ErrHabitGoalNotFound = fmt.Errorf("habit goal not found")
)

type UseCase interface {
	GetHabitProgress(ctx context.Context, username, habitName string) (entities.Progress, error)
	AddHabitProgress(ctx context.Context, username, habitName string) error
}

type UserUseCase interface {
	GetUserByUsername(ctx context.Context, username string) (entities.User, error)
}

type Storage interface {
	AddHabitProgress(ctx context.Context, goalId int) error
	GetHabitGoal(ctx context.Context, habitName string) (entities.Goal, error)
	GetCurrentPeriodExecutionCount(ctx context.Context, goalId int, frequencyType entities.FrequencyType) (int, error)
	GetCurrentProgress(ctx context.Context, goalId int) (entities.Progress, error)
	UpdateGoalStat(ctx context.Context, goalId int, progress entities.Progress) error
	GetPreviousPeriodExecutionCount(ctx context.Context, goalId int, frequencyType entities.FrequencyType) (int, error)
}

type Implementation struct {
	userUc  UserUseCase
	storage Storage
}

func New(userUc UserUseCase, storage Storage) *Implementation {
	return &Implementation{
		userUc:  userUc,
		storage: storage,
	}
}

func (i *Implementation) GetHabitProgress(ctx context.Context, username, habitName string) (entities.ProgressWithGoal, error) {
	_, err := i.userUc.GetUserByUsername(ctx, username)
	if err != nil {
		return entities.ProgressWithGoal{}, fmt.Errorf("i.userUc.GetUserByUsername: %w", err)
	}

	goal, err := i.storage.GetHabitGoal(ctx, habitName)
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

func (i *Implementation) AddHabitProgress(ctx context.Context, username, habitName string) error {
	_, err := i.userUc.GetUserByUsername(ctx, username)
	if err != nil {
		return fmt.Errorf("i.userUc.GetUserByUsername: %w", err)
	}

	goal, err := i.storage.GetHabitGoal(ctx, habitName)
	if err != nil {
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

	updatedProgress.TotalCompletedTimes = currentProgress.TotalCompletedTimes + 1
	if currentExecutionCount == goal.TimesPerFrequency { // If the goal is completed for the current period
		updatedProgress.TotalCompletedPeriods = currentProgress.TotalCompletedPeriods + 1
		updatedProgress.CurrentStreak = currentProgress.CurrentStreak + 1

		// Check if the current streak is the longest streak
		if (lastPeriodExecutionCount >= goal.TimesPerFrequency || lastPeriodExecutionCount == 0) && currentProgress.CurrentStreak+1 > currentProgress.MostLongestStreak {
			updatedProgress.MostLongestStreak = currentProgress.CurrentStreak + 1
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
	return nil
}
