package storage

import (
	"testing_trainer/internal/entities"
	"time"
)

func toEntityHabit(daoHabit habit) entities.Habit {
	if !daoHabit.frequency.Valid && !daoHabit.duration.Valid && !daoHabit.numOfPeriods.Valid && !daoHabit.startTrackingAt.Valid && !daoHabit.endTrackingAt.Valid {
		return entities.Habit{
			Name:        daoHabit.name,
			Description: daoHabit.description,
		}
	}

	goal := entities.NewGoal(int(daoHabit.frequency.Int64), time.Duration(daoHabit.duration.Int64), int(daoHabit.numOfPeriods.Int64), daoHabit.startTrackingAt.Time)

	return entities.Habit{
		Name:        daoHabit.name,
		Description: daoHabit.description,
		Goal:        &goal,
	}

}
