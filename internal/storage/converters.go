package storage

import (
	"testing_trainer/internal/entities"
)

func toEntityHabit(daoHabit habit) entities.Habit {
	if !daoHabit.frequencyType.Valid && !daoHabit.timesPerFrequency.Valid && !daoHabit.totalTrackingDays.Valid {
		return entities.Habit{
			Name:        daoHabit.name,
			Description: daoHabit.description,
		}
	}

	return entities.Habit{
		Name:        daoHabit.name,
		Description: daoHabit.description,
		Goal: &entities.Goal{
			TotalTrackingDays: daoHabit.totalTrackingDays.V,
			TimesPerFrequency: daoHabit.timesPerFrequency.V,
			FrequencyType:     entities.FrequencyTypeFromString(daoHabit.frequencyType.V),
		},
	}
}

func toEntityGoal(daoGoal goal) entities.Goal {
	return entities.Goal{
		Id:                daoGoal.id,
		TotalTrackingDays: daoGoal.totalTrackingDays,
		TimesPerFrequency: daoGoal.timesPerFrequency,
		FrequencyType:     entities.FrequencyTypeFromString(daoGoal.frequencyType),
	}
}
