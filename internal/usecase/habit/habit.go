package habit

import (
	"context"
	"errors"
	"fmt"

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
}

type UserUseCase interface {
	GetUserByUsername(ctx context.Context, username string) (entities.User, error)
}

type Transactor interface {
	RunRepeatableRead(ctx context.Context, fx func(ctxTX context.Context) error) error
}

type Storage interface {
	CreateHabit(ctx context.Context, username string, habit entities.Habit) (int, error)
	GetHabitById(ctx context.Context, username string, habitId int) (entities.Habit, error)
	GetHabitGoal(ctx context.Context, habitId int) (entities.Goal, error)
	ListUserHabits(ctx context.Context, username string) ([]entities.Habit, error)
	DeactivateGoalByID(ctx context.Context, id int) error
	CreateGoal(ctx context.Context, habitID int, goal entities.Goal) (int, error)
	UpdateGoalStat(ctx context.Context, goalId int, progress entities.Progress) error
	GetCurrentProgress(ctx context.Context, goalId int) (entities.Progress, error)
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
		currentGoal := currentHabit.Goal

		newGoal := habit.Goal
		if !entities.IsGoalChanged(currentGoal, newGoal) {
			return nil
		}

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
