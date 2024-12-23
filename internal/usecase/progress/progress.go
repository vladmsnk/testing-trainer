package progress

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"time"

	"testing_trainer/internal/entities"
	"testing_trainer/internal/storage"
)

//go:generate mockery --dir . --name Storage --structname MockStorage --filename storage_mock.go --output . --outpkg=progress
//go:generate mockery --dir . --name UserUseCase --structname MockUserUseCase --filename user_mock.go --output . --outpkg=progress
//go:generate mockery --dir . --name Transactor --structname MockTransactor --filename transactor_mock.go --output . --outpkg=progress

var (
	ErrHabitNotFound = fmt.Errorf("habit not found")
	ErrGoalNotFound  = fmt.Errorf("goal not found")
	ErrGoalCompleted = fmt.Errorf("goal is already completed")
)

type UseCase interface {
	GetHabitProgress(ctx context.Context, username string, habitId int) (entities.ProgressWithGoal, error)
	AddHabitProgress(ctx context.Context, username string, habitId int) error
	GetCurrentProgressForAllUserHabits(ctx context.Context, username string) ([]entities.CurrentPeriodProgress, error)
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
	GetHabitById(ctx context.Context, username string, habitId int) (entities.Habit, error)
	GetCurrentProgress(ctx context.Context, goalId int) (entities.Progress, error)
	UpdateGoalStat(ctx context.Context, goalId int, progress entities.Progress) error
	SetGoalCompleted(ctx context.Context, goalId int) error
	GetPreviousPeriodExecutionCount(ctx context.Context, goal entities.Goal) (int, error)
	GetCurrentPeriodExecutionCount(ctx context.Context, goal entities.Goal) (int, error)
	SetGoalNextCheckDate(ctx context.Context, goalId int, nextCheckDate time.Time) error
	GetAllUserHabitsWithGoals(ctx context.Context, username string) ([]entities.Habit, error)
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

	habit, err := i.storage.GetHabitById(ctx, username, habitId)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return entities.ProgressWithGoal{}, ErrHabitNotFound
		}
		return entities.ProgressWithGoal{}, fmt.Errorf("i.storage.GetHabitById: %w", err)
	}

	habitGoal := habit.Goal
	if habitGoal == nil {
		return entities.ProgressWithGoal{}, ErrGoalNotFound
	}

	progress, err := i.storage.GetCurrentProgress(ctx, habitGoal.Id)
	if err != nil {
		return entities.ProgressWithGoal{}, fmt.Errorf("i.storage.GetCurrentProgress: %w", err)
	}

	return entities.ProgressWithGoal{
		Progress: progress,
		Goal:     *habitGoal,
		Habit:    habit,
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
				return ErrHabitNotFound
			}
			return fmt.Errorf("i.storage.GetHabitGoal: %w", err)
		}

		if goal.IsCompleted {
			return ErrGoalCompleted
		}

		currentProgress, err := i.storage.GetCurrentProgress(ctx, goal.Id)
		if err != nil {
			return fmt.Errorf("i.storage.GetCurrentProgress: %w", err)
		}

		lastPeriodExecutionCount, err := i.storage.GetPreviousPeriodExecutionCount(ctx, goal)
		if err != nil {
			return fmt.Errorf("i.storage.GetPreviousDayExecutionCount: %w", err)
		}

		currentExecutionCount, err := i.storage.GetCurrentPeriodExecutionCount(ctx, goal)
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

func (i *Implementation) GetCurrentProgressForAllUserHabits(ctx context.Context, username string) ([]entities.CurrentPeriodProgress, error) {
	_, err := i.userUc.GetUserByUsername(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("i.userUc.GetUserByUsername: %w", err)
	}

	userHabits, err := i.storage.GetAllUserHabitsWithGoals(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("i.storage.GetAllUserHabitsWithGoals: %w", err)
	}

	var result []entities.CurrentPeriodProgress
	for _, habit := range userHabits {
		var currentPeriodProgress entities.CurrentPeriodProgress

		currentPeriodExecutionCount, err := i.storage.GetCurrentPeriodExecutionCount(ctx, *habit.Goal)
		if err != nil {
			return nil, fmt.Errorf("i.storage.GetCurrentPeriodExecutionCount: %w", err)
		}

		if currentPeriodExecutionCount >= habit.Goal.TimesPerFrequency {
			// Skip habits that are already completed for the current period
			continue
		}

		currentPeriodProgress.Habit = habit
		currentPeriodProgress.CurrentPeriodCompletedTimes = currentPeriodExecutionCount
		currentPeriodProgress.NeedToCompleteTimes = habit.Goal.TimesPerFrequency
		currentPeriodProgress.CurrentPeriod = habit.Goal.GetCurrentPeriod() + 1

		result = append(result, currentPeriodProgress)
	}

	slices.SortFunc(result, func(a, b entities.CurrentPeriodProgress) int {
		return a.Habit.Id - b.Habit.Id
	})

	return result, nil
}
