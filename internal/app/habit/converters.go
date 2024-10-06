package habit

import (
	"time"

	"testing_trainer/internal/entities"
)

func toEntityHabit(habit CreateHabitRequest) entities.Habit {
	duration := time.Duration(habit.DurationInDays) * 24 * time.Hour
	goal := entities.NewGoal(habit.Frequency, duration, habit.NumOfPeriods, habit.StartTrackingAt)

	return entities.Habit{
		Name:        habit.Name,
		Description: habit.Description,
		Goal:        &goal,
		IsArchived:  false,
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
	if habit.Goal != nil {
		return ResponseHabit{
			Name:        habit.Name,
			Description: habit.Description,
			Goal: &ResponseGoal{
				Frequency:       habit.Goal.Frequency,
				DurationInDays:  int(habit.Goal.Duration.Hours() / 24),
				NumOfPeriods:    habit.Goal.NumOfPeriods,
				StartTrackingAt: habit.Goal.StartTrackingAt,
			},
		}
	}

	return ResponseHabit{
		Name:        habit.Name,
		Description: habit.Description,
	}
}
