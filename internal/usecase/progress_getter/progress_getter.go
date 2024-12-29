package progress_getter

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
)

type ProgressGetter interface {
	GetProgressBySnapshot(ctx context.Context, goalID int, username string, currentTime time.Time) (entities.Progress, error)
	GetHabitProgress(ctx context.Context, username string, habitId int) (entities.ProgressWithGoal, error)
	GetCurrentProgressForAllUserHabits(ctx context.Context, username string) ([]entities.CurrentPeriodProgress, error)
}

type UserUseCase interface {
	GetUserByUsername(ctx context.Context, username string) (entities.User, error)
}

type Transactor interface {
	RunRepeatableRead(ctx context.Context, fx func(ctxTX context.Context) error) error
}

type Storage interface {
	GetHabitById(ctx context.Context, username string, habitId int) (entities.Habit, error)

	GetCurrentPeriodExecutionCount(ctx context.Context, goal entities.Goal, currentTime time.Time) (int, error)
	GetAllUserHabitsWithGoals(ctx context.Context, username string) ([]entities.Habit, error)

	CreateProgress(ctx context.Context, progress entities.Progress) (int64, error)
	GetMostRecentSnapshot(ctx context.Context, username string, goalID int, currentTime time.Time) (entities.ProgressSnapshot, error)
	GetCurrentSnapshot(ctx context.Context, username string, goalID int, currentTime time.Time) (entities.ProgressSnapshot, error)
	GetProgressByID(ctx context.Context, progressID int64) (entities.Progress, error)
	CreateSnapshot(ctx context.Context, snapshot entities.ProgressSnapshot) error
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

func NewGetter(userUc UserUseCase, storage Storage, transactor Transactor, timeManager TimeManager) *Implementation {
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
