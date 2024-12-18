//go:generate mockery --dir . --name Storage --structname MockStorage --filename storage_mock.go --output . --outpkg=progress
//go:generate mockery --dir . --name UserUseCase --structname MockUserUseCase --filename user_mock.go --output . --outpkg=progress
//go:generate mockery --dir . --name Transactor --structname MockTransactor --filename transactor_mock.go --output . --outpkg=progress
//go:generate mockery --dir . --name TimeManager --structname MockTimeManager --filename time_manager_mock.go --output . --outpkg=progress
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

var (
	ErrHabitNotFound = fmt.Errorf("habit not found")
	ErrGoalNotFound  = fmt.Errorf("goal not found")
	ErrGoalCompleted = fmt.Errorf("goal is already completed")
)

type UseCase interface {
	GetHabitProgress(ctx context.Context, username string, habitId int) (entities.ProgressWithGoal, error)
	AddHabitProgress(ctx context.Context, username string, habitId int) error
	GetCurrentProgressForAllUserHabits(ctx context.Context, username string) ([]entities.CurrentPeriodProgress, error)
	RecalculateFutureProgressesByGoalUpdate(ctx context.Context, username string, prevGoal, newGoal entities.Goal, currentTime time.Time) error
}

type UserUseCase interface {
	GetUserByUsername(ctx context.Context, username string) (entities.User, error)
}

type Transactor interface {
	RunRepeatableRead(ctx context.Context, fx func(ctxTX context.Context) error) error
}

type Storage interface {
	AddProgressLog(ctx context.Context, goalId int, createdAt time.Time) error
	GetHabitGoal(ctx context.Context, habitId int) (entities.Goal, error)
	GetHabitById(ctx context.Context, username string, habitId int) (entities.Habit, error)
	GetCurrentProgress(ctx context.Context, goalId int) (entities.Progress, error)
	SetGoalCompleted(ctx context.Context, goalId int) error
	GetPreviousPeriodExecutionCount(ctx context.Context, goal entities.Goal, currentTime time.Time) (int, error)
	GetCurrentPeriodExecutionCount(ctx context.Context, goal entities.Goal, currentTime time.Time) (int, error)
	SetGoalNextCheckDate(ctx context.Context, goalId int, nextCheckDate time.Time) error
	GetAllUserHabitsWithGoals(ctx context.Context, username string) ([]entities.Habit, error)

	//V2
	CreateProgress(ctx context.Context, progress entities.Progress) (int64, error)
	UpdateProgressByID(ctx context.Context, progress entities.Progress) error
	GetMostRecentSnapshot(ctx context.Context, username string, goalID int, currentTime time.Time) (entities.ProgressSnapshot, error)
	GetCurrentSnapshot(ctx context.Context, username string, goalID int, currentTime time.Time) (entities.ProgressSnapshot, error)
	GetProgressByID(ctx context.Context, progressID int64) (entities.Progress, error)
	CreateSnapshot(ctx context.Context, snapshot entities.ProgressSnapshot) error

	GetFutureSnapshots(ctx context.Context, username string, goalID int, currentTime time.Time) ([]entities.ProgressSnapshot, error)
	GetCurrentDayExecutionCount(ctx context.Context, goal entities.Goal, currentTime time.Time) (int, error)
}

type TimeManager interface {
	GetCurrentTime(ctx context.Context, username string) (time.Time, error)
}

type Implementation struct {
	userUc      UserUseCase
	storage     Storage
	transactor  Transactor
	timeManager TimeManager
}

func New(userUc UserUseCase, storage Storage, transactor Transactor, timeManager TimeManager) *Implementation {
	return &Implementation{
		userUc:      userUc,
		storage:     storage,
		transactor:  transactor,
		timeManager: timeManager,
	}
}

func (i *Implementation) GetHabitProgress(ctx context.Context, username string, habitId int) (entities.ProgressWithGoal, error) {
	_, err := i.userUc.GetUserByUsername(ctx, username)
	if err != nil {
		return entities.ProgressWithGoal{}, fmt.Errorf("i.userUc.GetUserByUsername: %w", err)
	}

	currentTime, err := i.timeManager.GetCurrentTime(ctx, username)
	if err != nil {
		return entities.ProgressWithGoal{}, fmt.Errorf("i.timeManager.GetCurrentTime: %w", err)
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

	var progress entities.Progress

	progress, err = i.GetProgressBySnapshot(ctx, habitGoal.Id, username, currentTime)
	if err != nil {
		return entities.ProgressWithGoal{}, fmt.Errorf("i.GetProgressBySnapshot: %w", err)
	}

	return entities.ProgressWithGoal{
		Progress: progress,
		Goal:     *habitGoal,
		Habit:    habit,
	}, nil
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

		currentProgress, err := i.GetProgressBySnapshot(ctx, goal.Id, username, currentTime)
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

		err = i.RecalculateFutureProgressesByGoalUpdate(ctx, username, goal, goal, currentTime)
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

func (i *Implementation) GetCurrentProgressForAllUserHabits(ctx context.Context, username string) ([]entities.CurrentPeriodProgress, error) {
	currentTime, err := i.timeManager.GetCurrentTime(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("i.timeManager.GetCurrentTime: %w", err)
	}

	_, err = i.userUc.GetUserByUsername(ctx, username)
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

		currentPeriodExecutionCount, err := i.storage.GetCurrentPeriodExecutionCount(ctx, *habit.Goal, currentTime)
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
		currentPeriodProgress.CurrentPeriod = habit.Goal.GetCurrentPeriod(currentTime) + 1

		result = append(result, currentPeriodProgress)
	}

	slices.SortFunc(result, func(a, b entities.CurrentPeriodProgress) int {
		return a.Habit.Id - b.Habit.Id
	})

	return result, nil
}

// GetProgressBySnapshot Создавать снимок при каждом получении прогресса, если снимка нет
// снимок должен связывать id прогресса и временную метку текущего дня
// по временной метки будет определяться снимок и соответсвующий прогресс
// на каждый день должен быть свой снимок и своя копия прогресса
// снимок создается на основе предыдущего дня
// должен быть базовый прогресс, от которого будут создаваться все снимки
// если нет никакого снимка создается пустой прогресс
func (i *Implementation) GetProgressBySnapshot(ctx context.Context, goalID int, username string, currentTime time.Time) (entities.Progress, error) {
	snapshot, err := i.storage.GetCurrentSnapshot(ctx, username, goalID, currentTime)
	if err != nil {
		if !errors.Is(err, storage.ErrNotFound) {
			return entities.Progress{}, fmt.Errorf("i.storage.GetCurrentSnapshot: %w", err)
		}

		recentSnapshot, err := i.storage.GetMostRecentSnapshot(ctx, username, goalID, currentTime)
		if err != nil {
			if !errors.Is(err, storage.ErrNotFound) {
				return entities.Progress{}, fmt.Errorf("i.storage.GetMostRecentSnapshot: %w", err)
			}

			emptyProgress := entities.Progress{
				Username:  username,
				GoalID:    goalID,
				CreatedAt: currentTime,
				UpdatedAt: currentTime,
			}

			newProgressID, err := i.storage.CreateProgress(ctx, emptyProgress)
			if err != nil {
				return entities.Progress{}, fmt.Errorf("i.storage.CreateProgress: %w", err)
			}

			emptyProgress.Id = int(newProgressID)

			err = i.storage.CreateSnapshot(ctx, entities.ProgressSnapshot{
				Username:   username,
				ProgressID: newProgressID,
				CreatedAt:  currentTime,
				GoalID:     goalID,
			})
			if err != nil {
				return entities.Progress{}, fmt.Errorf("i.storage.CreateSnapshot: %w", err)
			}

			return emptyProgress, nil
		}

		progress, err := i.storage.GetProgressByID(ctx, recentSnapshot.ProgressID)
		if err != nil {
			return entities.Progress{}, fmt.Errorf("i.storage.GetProgressByID: %w", err)
		}

		progress.CreatedAt = currentTime
		progress.UpdatedAt = currentTime

		newProgressID, err := i.storage.CreateProgress(ctx, progress)
		if err != nil {
			return entities.Progress{}, fmt.Errorf("i.storage.CreateProgress: %w", err)
		}

		progress.Id = int(newProgressID)

		err = i.storage.CreateSnapshot(ctx, entities.ProgressSnapshot{
			Username:   username,
			ProgressID: newProgressID,
			CreatedAt:  currentTime,
			GoalID:     goalID,
		})
		if err != nil {
			return entities.Progress{}, fmt.Errorf("i.storage.CreateSnapshot: %w", err)
		}

		return progress, err
	}

	progress, err := i.storage.GetProgressByID(ctx, snapshot.ProgressID)
	if err != nil {
		return entities.Progress{}, fmt.Errorf("i.storage.GetProgressByID: %w", err)
	}

	return progress, nil
}

func (i *Implementation) RecalculateFutureProgressesByGoalUpdate(ctx context.Context, username string, prevGoal, newGoal entities.Goal, currentTime time.Time) error {
	baseProgress, err := i.GetProgressBySnapshot(ctx, prevGoal.Id, username, currentTime)
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
