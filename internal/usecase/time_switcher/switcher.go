//go:generate mockery --dir . --name Storage --structname MockStorage --filename storage_mock.go --output . --outpkg=time_switcher
//go:generate mockery --dir . --name TimeManager --structname MockTimeManager --filename time_manager_mock.go --output . --outpkg=time_switcher
package time_switcher

import (
	"context"
	"fmt"
	"time"
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

type Implementation struct {
	timeManager TimeManager
}

func New(timeManager TimeManager) *Implementation {
	return &Implementation{
		timeManager: timeManager,
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
	err := i.timeManager.SetTimeOffset(ctx, username, 1)
	if err != nil {
		return fmt.Errorf("i.timeManager.SetTimeOffset: %w", err)
	}

	return nil
}
