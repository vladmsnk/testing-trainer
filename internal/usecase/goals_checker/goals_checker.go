package goals_checker

//go:generate mockery --dir . --name Storage --structname MockStorage --filename storage_mock.go --output . --outpkg=goals_checker
//go:generate mockery --dir . --name Transactor --structname MockTransactor --filename transactor_mock.go --output . --outpkg=goals_checker

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
	GetAllGoalsNeedCheck(ctx context.Context) ([]entities.Goal, error)
	SetGoalNextCheckDate(ctx context.Context, goalId int, nextCheckDate time.Time) error
	UpdateGoalStat(ctx context.Context, goalId int, progress entities.Progress) error
	GetCurrentProgress(ctx context.Context, goalId int) (entities.Progress, error)
	GetPreviousPeriodExecutionCount(ctx context.Context, goal entities.Goal) (int, error)
}

type Transactor interface {
	RunRepeatableRead(ctx context.Context, fx func(ctxTX context.Context) error) error
}

func NewChecker(storage Storage, transactor Transactor) Checker {
	return &Implementation{
		storage:    storage,
		transactor: transactor,
	}
}

type Implementation struct {
	storage    Storage
	transactor Transactor
}

func (i *Implementation) CheckGoals(ctx context.Context) error {
	goalsToCheck, err := i.storage.GetAllGoalsNeedCheck(ctx)
	if err != nil {
		return fmt.Errorf("i.storage.GetAllGoalsNeedCheck: %w", err)
	}

	for _, goal := range goalsToCheck {
		err := i.transactor.RunRepeatableRead(ctx, func(ctxTX context.Context) error {
			progress, err := i.storage.GetCurrentProgress(ctx, goal.Id)
			if err != nil {
				return fmt.Errorf("i.storage.GetCurrentProgress: %w", err)
			}

			updatedProgress := progress

			lastPeriodExecutionCount, err := i.storage.GetPreviousPeriodExecutionCount(ctx, goal)
			if err != nil {
				return fmt.Errorf("i.storage.GetPreviousDayExecutionCount: %w", err)
			}

			if lastPeriodExecutionCount < goal.TimesPerFrequency {
				updatedProgress.CurrentStreak = 0
				updatedProgress.TotalSkippedPeriods = updatedProgress.TotalSkippedPeriods + 1

				err = i.storage.UpdateGoalStat(ctx, goal.Id, updatedProgress)
				if err != nil {
					return fmt.Errorf("i.storage.UpdateGoalStat: %w", err)
				}
			}

			var nextCheckDate time.Time

			switch goal.FrequencyType {
			case entities.Daily:
				nextCheckDate = goal.NextCheckDate.Add(24 * time.Hour)
			case entities.Weekly:
				nextCheckDate = goal.NextCheckDate.Add(7 * 24 * time.Hour)
			case entities.Monthly:
				nextCheckDate = goal.NextCheckDate.AddDate(0, 1, 0)
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
