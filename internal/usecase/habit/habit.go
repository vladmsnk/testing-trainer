package habit

import (
	"context"
	"errors"
	"fmt"

	"testing_trainer/internal/entities"
	"testing_trainer/internal/storage"
)

var (
	ErrHabitAlreadyExists = fmt.Errorf("habit already exists")
)

type UseCase interface {
	CreateHabit(ctx context.Context, username string, habit entities.Habit) (int64, error)
	ListUserHabits(ctx context.Context, username string) ([]entities.Habit, error)
}

type UserUseCase interface {
	GetUserByUsername(ctx context.Context, username string) (entities.User, error)
}

type Storage interface {
	CreateHabit(ctx context.Context, username string, habit entities.Habit) (int64, error)
	GetHabitByName(ctx context.Context, username, habitName string) (entities.Habit, error)
	ListUserHabits(ctx context.Context, username string) ([]entities.Habit, error)
}

type Implementation struct {
	storage Storage
	userUc  UserUseCase
}

func New(storage Storage, userUc UserUseCase) *Implementation {
	return &Implementation{storage: storage, userUc: userUc}
}

func (i *Implementation) CreateHabit(ctx context.Context, username string, habit entities.Habit) (int64, error) {
	_, err := i.userUc.GetUserByUsername(ctx, username)
	if err != nil {
		return 0, fmt.Errorf("i.userUc.GetUserByUsername: %w", err)
	}

	_, err = i.storage.GetHabitByName(ctx, username, habit.Name)
	if err != nil {
		if !errors.Is(err, storage.ErrNotFound) {
			return 0, fmt.Errorf("i.storage.GetHabitByName: %w", err)
		}
	} else {
		return 0, ErrHabitAlreadyExists
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
		return nil, fmt.Errorf("i.userUc.GetUserByUsername: %w", err)
	}

	userHabits, err := i.storage.ListUserHabits(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("i.storage.ListUserHabits: %w", err)
	}

	return userHabits, nil
}
