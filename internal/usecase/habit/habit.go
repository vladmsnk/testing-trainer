package habit

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"time"

	"testing_trainer/internal/entities"
	"testing_trainer/internal/storage"
)

//go:generate mockery --dir . --name Storage --structname MockStorage --filename storage_mock.go --output . --outpkg=habit
//go:generate mockery --dir . --name UserUseCase --structname MockUserUseCase --filename user_mock.go --output . --outpkg=habit
//go:generate mockery --dir . --name Transactor --structname MockTransactor --filename transactor_mock.go --output . --outpkg=habit

var (
	ErrUserNotFound  = fmt.Errorf("user not found")
	ErrHabitNotFound = fmt.Errorf("habit not found")
)

type UseCase interface {
	CreateHabit(ctx context.Context, username string, habit entities.Habit) (int, error)
	ListUserHabits(ctx context.Context, username string) ([]entities.Habit, error)
	UpdateHabit(ctx context.Context, username string, habit entities.Habit) error
	ListUserCompletedHabits(ctx context.Context, username string) ([]entities.Habit, error)
	DeleteHabit(ctx context.Context, username string, habitId int) error
}

type UserUseCase interface {
	GetUserByUsername(ctx context.Context, username string) (entities.User, error)
}

type Transactor interface {
	RunRepeatableRead(ctx context.Context, fx func(ctxTX context.Context) error) error
}

type Storage interface {
	ArchiveHabitById(ctx context.Context, habitId int) error
	CreateHabit(ctx context.Context, username string, habit entities.Habit) (int, error)
	GetHabitById(ctx context.Context, username string, habitId int) (entities.Habit, error)
	GetHabitGoal(ctx context.Context, habitId int) (entities.Goal, error)
	ListUserHabits(ctx context.Context, username string) ([]entities.Habit, error)
	ListUserCompletedHabits(ctx context.Context, username string) ([]entities.Habit, error)
	DeactivateGoalByID(ctx context.Context, id int) error
	CreateGoal(ctx context.Context, habitID int, goal entities.Goal) (int, error)
	UpdateGoalStat(ctx context.Context, goalId int, progress entities.Progress) error
	GetCurrentProgress(ctx context.Context, goalId int) (entities.Progress, error)
	UpdateHabit(ctx context.Context, habit entities.Habit) error
}

type Implementation struct {
	storage Storage
	userUc  UserUseCase
	tx      Transactor
}

func New(storage Storage, userUc UserUseCase, tx Transactor) *Implementation {
	return &Implementation{storage: storage, userUc: userUc, tx: tx}
}

func (i *Implementation) CreateHabit(ctx context.Context, username string, habit entities.Habit) (int, error) {
	_, err := i.userUc.GetUserByUsername(ctx, username)
	if err != nil {
		return 0, fmt.Errorf("i.userUc.GetUserByUsername: %w", err)
	}

	if habit.Goal != nil {
		var nextCheckDate time.Time

		switch habit.Goal.FrequencyType {
		case entities.Daily:
			nextCheckDate = time.Now().Add(24 * time.Hour).Add(5 * time.Minute)
		case entities.Weekly:
			nextCheckDate = time.Now().AddDate(0, 0, 7).Add(5 * time.Minute)
		case entities.Monthly:
			nextCheckDate = time.Now().AddDate(0, 1, 0).Add(5 * time.Minute)
		}

		habit.Goal.NextCheckDate = nextCheckDate.UTC()
	}

	createdHabitId, err := i.storage.CreateHabit(ctx, username, habit)
	if err != nil {
		return 0, fmt.Errorf("i.storage.CreateHabit: %w", err)
	}

	return createdHabitId, nil
}

func (i *Implementation) ListUserHabits(ctx context.Context, username string) ([]entities.Habit, error) {
	_, err := i.userUc.GetUserByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("i.userUc.GetUserByUsername: %w", err)
	}

	userHabits, err := i.storage.ListUserHabits(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("i.storage.ListUserHabits: %w", err)
	}

	slices.SortFunc(userHabits, func(h1, h2 entities.Habit) int {
		return h1.Id - h2.Id
	})

	return userHabits, nil
}

func (i *Implementation) UpdateHabit(ctx context.Context, username string, habit entities.Habit) error {
	err := i.tx.RunRepeatableRead(ctx, func(ctxTX context.Context) error {
		_, err := i.userUc.GetUserByUsername(ctx, username)
		if err != nil {
			if errors.Is(err, storage.ErrNotFound) {
				return ErrUserNotFound
			}
			return fmt.Errorf("i.userUc.GetUserByUsername: %w", err)
		}

		currentHabit, err := i.storage.GetHabitById(ctx, username, habit.Id)
		if err != nil {
			if errors.Is(err, storage.ErrNotFound) {
				return ErrHabitNotFound
			}
			return fmt.Errorf("storage.GetHabitById: %w", err)
		}

		if entities.IsHabitChanged(currentHabit, habit) {
			err := i.storage.UpdateHabit(ctx, habit)
			if err != nil {
				return fmt.Errorf("storage.UpdateHabit: %w", err)
			}
		}
		currentGoal := currentHabit.Goal

		newGoal := habit.Goal
		newGoal.PreviousGoalId = currentGoal.PreviousGoalId
		if !entities.IsGoalChanged(currentGoal, newGoal) {
			return nil
		}
		newGoal.NextCheckDate = currentGoal.NextCheckDate

		err = i.storage.DeactivateGoalByID(ctx, currentGoal.Id)
		if err != nil {
			return fmt.Errorf("storage.DeactivateGoalByID: %w", err)
		}

		newGoalId, err := i.storage.CreateGoal(ctx, habit.Id, *newGoal)
		if err != nil {
			return fmt.Errorf("storage.CreateGoal: %w", err)
		}

		currentProgress, err := i.storage.GetCurrentProgress(ctx, currentGoal.Id)
		if err != nil {
			return fmt.Errorf("storage.GetCurrentProgress: %w", err)
		}

		err = i.storage.UpdateGoalStat(ctx, newGoalId, currentProgress)
		if err != nil {
			return fmt.Errorf("storage.UpdateGoalStat: %w", err)
		}
		// Текущая цель перестает быть активной
		// Создается новая запись прогресса, куда переносится вся текущая статистика
		// Теперь привычка отслеживается по новым правилам
		return nil
	})

	return err
}

func (i *Implementation) ListUserCompletedHabits(ctx context.Context, username string) ([]entities.Habit, error) {
	_, err := i.userUc.GetUserByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("i.userUc.GetUserByUsername: %w", err)
	}

	habits, err := i.storage.ListUserCompletedHabits(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("i.storage.ListUserCompletedHabits: %w", err)
	}

	slices.SortFunc(habits, func(h1, h2 entities.Habit) int {
		return h1.Id - h2.Id
	})

	return habits, nil
}

func (i *Implementation) DeleteHabit(ctx context.Context, username string, habitId int) error {
	err := i.tx.RunRepeatableRead(ctx, func(ctxTX context.Context) error {
		_, err := i.userUc.GetUserByUsername(ctx, username)
		if err != nil {
			if errors.Is(err, storage.ErrNotFound) {
				return ErrUserNotFound
			}
			return fmt.Errorf("i.userUc.GetUserByUsername: %w", err)
		}

		habit, err := i.storage.GetHabitById(ctx, username, habitId)
		if err != nil {
			if errors.Is(err, storage.ErrNotFound) {
				return ErrHabitNotFound
			}
			return fmt.Errorf("i.storage.GetHabitById: %w", err)
		}

		err = i.storage.ArchiveHabitById(ctx, habitId)
		if err != nil {
			return fmt.Errorf("i.storage.ArchiveHabitById: %w", err)
		}

		if habit.Goal != nil {
			err := i.storage.DeactivateGoalByID(ctx, habit.Goal.Id)
			if err != nil {
				return fmt.Errorf("i.storage.DeactivateGoalByID: %w", err)
			}
		}
		return nil
	})

	return err
}
