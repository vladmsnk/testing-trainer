package habit

import (
	"context"
	"fmt"

	"testing_trainer/internal/entities"
)

type UseCase interface {
	CreateHabit(ctx context.Context, username string, habit entities.Habit) (int64, error)
}

type UserUseCase interface {
	GetUserByUsername(ctx context.Context, username string) (entities.User, error)
}

type Storage interface {
	CreateHabit(ctx context.Context, username string, habit entities.Habit) (int64, error)
}

type Implementation struct {
	storage Storage
	userUc  UserUseCase
}

func New(storage Storage) *Implementation {
	return &Implementation{storage: storage}
}

func (i *Implementation) CreateHabit(ctx context.Context, username string, habit entities.Habit) (int64, error) {
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
