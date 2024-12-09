package time_manager

import (
	"context"
	"errors"
	"fmt"
	"testing_trainer/internal/storage"
	"time"
)

type UseCase interface {
	GetCurrentTime(ctx context.Context, username string) (time.Time, error)
	SetTimeOffset(ctx context.Context, username string, offset int) error
	ResetTime(ctx context.Context, username string) error
}

type Implementation struct {
	storage Storage
}

func New(storage Storage) *Implementation {
	return &Implementation{storage: storage}
}

type Storage interface {
	GetUserTimeOffset(ctx context.Context, username string) (int, error)
	UpdateUserTimeOffset(ctx context.Context, username string, offset int) error
}

func (i *Implementation) GetCurrentTime(ctx context.Context, username string) (time.Time, error) {
	userTimeOffset, err := i.storage.GetUserTimeOffset(ctx, username)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return time.Now(), nil
		}
		return time.Time{}, fmt.Errorf("i.storage.GetUserTimeOffset: %w", err)
	}

	currentTime := time.Now().Add(time.Duration(userTimeOffset) * time.Hour * 24)
	return currentTime, nil
}

func (i *Implementation) SetTimeOffset(ctx context.Context, username string, offset int) error {
	err := i.storage.UpdateUserTimeOffset(ctx, username, offset)
	if err != nil {
		return fmt.Errorf("i.storage.UpdateUserTimeOffset: %w", err)
	}
	return nil
}

func (i *Implementation) ResetTime(ctx context.Context, username string) error {
	err := i.storage.UpdateUserTimeOffset(ctx, username, 0)
	if err != nil {
		return fmt.Errorf("i.storage.UpdateUserTimeOffset: %w", err)
	}
	return nil
}
