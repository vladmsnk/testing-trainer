package habit

import (
	"testing_trainer/internal/entities"
)

func toCreateHabitEntity(habit CreateHabitRequest) entities.Habit {
	if habit.Goal == nil {
		return entities.Habit{
			Description: habit.Description,
		}
	}

	return entities.Habit{
		Description: habit.Description,
		Goal: &entities.Goal{
			FrequencyType:        toEntityFrequencyType(habit.Goal.FrequencyType),
			TimesPerFrequency:    habit.Goal.TimesPerFrequency,
			TotalTrackingPeriods: habit.Goal.TotalTrackingPeriods,
		},
	}
}

func toUpdateHabitEntity(habit UpdateHabitRequest) entities.Habit {
	if habit.Goal == nil {
		return entities.Habit{
			Id:          habit.Id,
			Description: habit.Description,
		}
	}

	return entities.Habit{
		Id:          habit.Id,
		Description: habit.Description,
		Goal: &entities.Goal{
			FrequencyType:        toEntityFrequencyType(habit.Goal.FrequencyType),
			TimesPerFrequency:    habit.Goal.TimesPerFrequency,
			TotalTrackingPeriods: habit.Goal.TotalTrackingPeriods,
		},
	}
}

func toEntityFrequencyType(frequencyType string) entities.FrequencyType {
	switch frequencyType {
	case "daily":
		return entities.Daily
	case "weekly":
		return entities.Weekly
	case "monthly":
		return entities.Monthly
	default:
		return entities.UndefinedFrequencyType
	}

}

func toListUserHabitsResponse(username string, habits []entities.Habit) ListUserHabitsResponse {
	var result []ResponseHabit

	for _, habit := range habits {
		result = append(result, toResponseHabit(habit))
	}

	return ListUserHabitsResponse{Habits: result, Username: username}
}

func toResponseHabit(habit entities.Habit) ResponseHabit {
	r := ResponseHabit{
		Id:          habit.Id,
		Description: habit.Description,
	}
	if habit.Goal != nil {
		r.Goal = &Goal{
			Id:                   habit.Goal.Id,
			FrequencyType:        habit.Goal.FrequencyType.String(),
			TimesPerFrequency:    habit.Goal.TimesPerFrequency,
			TotalTrackingPeriods: habit.Goal.TotalTrackingPeriods,
		}
	}
	return r
}
