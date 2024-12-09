//go:generate mockery --dir . --name Storage --structname MockStorage --filename storage_mock.go --output . --outpkg=time_switcher
//go:generate mockery --dir . --name TimeManager --structname MockTimeManager --filename time_manager_mock.go --output . --outpkg=time_switcher
package time_switcher

import (
	"context"
	"errors"
	"fmt"
	"time"

	"testing_trainer/internal/entities"
	"testing_trainer/internal/storage"
)

type UseCase interface {
	SwitchToNextDay(ctx context.Context, username string) error
	ResetToCurrentDay(ctx context.Context, username string) error
	GetCurrentTime(ctx context.Context, username string) (time.Time, error)
}

type TimeManager interface {
	GetCurrentTime(ctx context.Context, username string) (time.Time, error)
	SetTimeOffset(ctx context.Context, username string, offset int) error
	ResetTime(ctx context.Context, username string) error
}

type Storage interface {
	GetProgressesForAllGoals(ctx context.Context, username string, progressIDs []int64) ([]entities.Progress, error)
	GetSnapshotForTheTime(ctx context.Context, username string, time time.Time) (entities.ProgressSnapshot, error)
	CreateProgress(ctx context.Context, progress entities.Progress) (int64, error)
	CreateProgressSnapshot(ctx context.Context, snapshot entities.ProgressSnapshot) error
}

type Implementation struct {
	timeManager TimeManager
	storage     Storage
}

func New(timeManager TimeManager, storage Storage) *Implementation {
	return &Implementation{
		timeManager: timeManager,
		storage:     storage,
	}
}

func (i *Implementation) GetCurrentTime(ctx context.Context, username string) (time.Time, error) {
	currentTime, err := i.timeManager.GetCurrentTime(ctx, username)
	if err != nil {
		return time.Time{}, fmt.Errorf("i.timeManager.GetCurrentTime: %w", err)
	}
	return currentTime, nil
}

func (i *Implementation) ResetToCurrentDay(ctx context.Context, username string) error {
	err := i.timeManager.SetTimeOffset(ctx, username, 0)
	if err != nil {
		return fmt.Errorf("i.timeManager.SetTimeOffset: %w", err)
	}
	return nil
}

func (i *Implementation) SwitchToNextDay(ctx context.Context, username string) error {
	// Get current time_manager for the user
	currentTime, err := i.timeManager.GetCurrentTime(ctx, username)
	if err != nil {
		return fmt.Errorf("i.timeManager.GetCurrentTime: %w", err)
	}

	err = i.timeManager.SetTimeOffset(ctx, username, 1)
	if err != nil {
		return fmt.Errorf("i.timeManager.SetTimeOffset: %w", err)
	}

	nextDayTime, err := i.timeManager.GetCurrentTime(ctx, username)
	if err != nil {
		return fmt.Errorf("i.timeManager.GetCurrentTime: %w", err)
	}

	currentSnapshot, err := i.storage.GetSnapshotForTheTime(ctx, username, nextDayTime)
	if err != nil {
		if !errors.Is(err, storage.ErrNotFound) {
			return fmt.Errorf("i.storage.GetSnapshotForTheCurrentTime: %w", err)
		}
	} else {
		// if snapshot is found, it means that user already switched to the next day
		return nil
	}

	var progresses []entities.Progress

	// get current progress snapshot for the current time_swticher and username
	// if progress snapshot is not found it means that user switched to the next day for the first time_swticher
	// if progress snapshot is found it means that user switched to the next day before
	currentSnapshot, err = i.storage.GetSnapshotForTheTime(ctx, username, currentTime)
	if err != nil {
		if !errors.Is(err, storage.ErrNotFound) {
			return fmt.Errorf("i.storage.GetSnapshotForTheCurrentTime: %w", err)
		}
		// если снимок не найден, берем текущие прогрессы
		currentProgresses, err := i.storage.GetProgressesForAllGoals(ctx, username, []int64{})
		if err != nil {
			return fmt.Errorf("i.storage.GetProgressesForAllGoals: %w", err)
		}

		progresses = make([]entities.Progress, len(currentProgresses))
		copy(progresses, currentProgresses)

	} else {
		currentProgresses, err := i.storage.GetProgressesForAllGoals(ctx, username, currentSnapshot.CurrentProgressIDs)
		if err != nil {
			return fmt.Errorf("i.storage.GetProgressesForAllGoals: %w", err)
		}

		progresses = make([]entities.Progress, len(currentProgresses))
		copy(progresses, currentProgresses)
	}

	var createdProgressIds []int64

	// create new progresses as copies of current progresses but with updated created_at and updated_at fields
	for _, progress := range progresses {
		progress.CreatedAt = nextDayTime
		progress.UpdatedAt = nextDayTime

		createdProgressId, err := i.storage.CreateProgress(ctx, progress)
		if err != nil {
			return fmt.Errorf("i.storage.CreateProgress: %w", err)
		}
		createdProgressIds = append(createdProgressIds, createdProgressId)
	}

	// create new progress snapshot
	snapshot := entities.ProgressSnapshot{
		Username:           username,
		CurrentProgressIDs: createdProgressIds,
		CreatedAt:          nextDayTime,
	}

	err = i.storage.CreateProgressSnapshot(ctx, snapshot)
	if err != nil {
		return fmt.Errorf("i.storage.CreateProgressSnapshot: %w", err)
	}

	return nil
}
