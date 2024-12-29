package goals_checker

//go:generate mockery --dir . --name Storage --structname MockStorage --filename storage_mock.go --output . --outpkg=goals_checker
//go:generate mockery --dir . --name Transactor --structname MockTransactor --filename transactor_mock.go --output . --outpkg=goals_checker
//go:generate mockery --dir . --name TimeManager --structname MockTimeManager --filename time_manager_mock.go --output . --outpkg=goals_checker

import (
	"context"
	"fmt"
	"time"

	"testing_trainer/internal/entities"
)

type Checker interface {
	CheckGoals(ctx context.Context) error
}

type Storage interface {
	GetAllGoalsNeedCheck(ctx context.Context, currentTime time.Time) ([]entities.Goal, error)
	SetGoalNextCheckDate(ctx context.Context, goalId int, nextCheckDate time.Time) error
	GetPreviousPeriodExecutionCount(ctx context.Context, goal entities.Goal, currentTime time.Time) (int, error)
	UpdateProgressByID(ctx context.Context, progress entities.Progress) error
}

type Transactor interface {
	RunRepeatableRead(ctx context.Context, fx func(ctxTX context.Context) error) error
}

type TimeManager interface {
	GetCurrentTime(ctx context.Context, username string) (time.Time, error)
}

type ProgressGetter interface {
	GetProgressBySnapshot(ctx context.Context, goalID int, username string, currentTime time.Time) (entities.Progress, error)
}

func NewChecker(storage Storage, transactor Transactor, timeManager TimeManager, progressGetter ProgressGetter) Checker {
	return &Implementation{
		storage:        storage,
		transactor:     transactor,
		timeManager:    timeManager,
		progressGetter: progressGetter,
	}
}

type Implementation struct {
	storage        Storage
	transactor     Transactor
	timeManager    TimeManager
	progressGetter ProgressGetter
}

var (
	goalsCheckerUser = "goals_checker"
)

func (i *Implementation) CheckGoals(ctx context.Context) error {
	currentTime, err := i.timeManager.GetCurrentTime(ctx, goalsCheckerUser)
	if err != nil {
		return fmt.Errorf("i.timeManager.GetCurrentTime: %w", err)
	}

	goalsToCheck, err := i.storage.GetAllGoalsNeedCheck(ctx, currentTime)
	if err != nil {
		return fmt.Errorf("i.storage.GetAllGoalsNeedCheck: %w", err)
	}

	for _, goal := range goalsToCheck {
		err := i.transactor.RunRepeatableRead(ctx, func(ctxTX context.Context) error {
			currentProgress, err := i.progressGetter.GetProgressBySnapshot(ctxTX, goal.Id, goalsCheckerUser, currentTime)
			if err != nil {
				return fmt.Errorf("i.progressManager.GetProgressBySnapshot: %w", err)
			}

			updatedProgress := currentProgress

			lastPeriodExecutionCount, err := i.storage.GetPreviousPeriodExecutionCount(ctx, goal, currentTime)
			if err != nil {
				return fmt.Errorf("i.storage.GetPreviousDayExecutionCount: %w", err)
			}

			if lastPeriodExecutionCount < goal.TimesPerFrequency {
				updatedProgress.CurrentStreak = 0
				updatedProgress.TotalSkippedPeriods = updatedProgress.TotalSkippedPeriods + 1
			}

			err = i.storage.UpdateProgressByID(ctxTX, updatedProgress)
			if err != nil {
				return fmt.Errorf("i.storage.UpdateProgressByID: %w", err)
			}

			var nextCheckDate time.Time

			switch goal.FrequencyType {
			case entities.Daily:
				nextCheckDate = goal.NextCheckDate.Add(24 * time.Hour)
			case entities.Weekly:
				nextCheckDate = goal.NextCheckDate.Add(7 * 24 * time.Hour)
			case entities.Monthly:
				nextCheckDate = goal.NextCheckDate.AddDate(0, 1, 0)
			default:

			}

			err = i.storage.SetGoalNextCheckDate(ctx, goal.Id, nextCheckDate)
			if err != nil {
				return fmt.Errorf("i.storage.SetGoalNextCheckDate: %w", err)
			}

			return nil
		})
		if err != nil {
			return fmt.Errorf("i.transactor.RunRepeatableRead goalID=%d: %w", goal.Id, err)
		}
	}
	return nil
}
