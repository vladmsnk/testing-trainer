package progress

import "testing_trainer/internal/entities"

func toHabitProgressResponse(progressWithGoal entities.ProgressWithGoal) GetHabitProgressResponse {
	return GetHabitProgressResponse{
		Progress: Progress{
			TotalCompletedPeriods: progressWithGoal.TotalCompletedPeriods,
			TotalSkippedPeriods:   progressWithGoal.TotalSkippedPeriods,
			TotalCompletedTimes:   progressWithGoal.TotalCompletedTimes,
			MostLongestStreak:     progressWithGoal.MostLongestStreak,
			CurrentStreak:         progressWithGoal.CurrentStreak,
		},
		Goal: Goal{
			FrequencyType:        progressWithGoal.FrequencyType.String(),
			TimesPerFrequency:    progressWithGoal.TimesPerFrequency,
			TotalTrackingPeriods: progressWithGoal.TotalTrackingPeriods,
		},
	}
}

func toReminderResponse(currentProgressForAllUserHabits []entities.CurrentPeriodProgress) GetReminderResponse {
	var reminder []CurrentPeriodProgress

	for _, currentPeriodProgress := range currentProgressForAllUserHabits {
		reminder = append(reminder, CurrentPeriodProgress{
			Habit: Habit{
				Id:          currentPeriodProgress.Habit.Id,
				Description: currentPeriodProgress.Habit.Description,
			},
			Goal: Goal{
				FrequencyType:        currentPeriodProgress.Habit.Goal.FrequencyType.String(),
				TimesPerFrequency:    currentPeriodProgress.Habit.Goal.TimesPerFrequency,
				TotalTrackingPeriods: currentPeriodProgress.Habit.Goal.TotalTrackingPeriods,
			},
			CurrentPeriodCompletedTimes: currentPeriodProgress.CurrentPeriodCompletedTimes,
			NeedToCompleteTimes:         currentPeriodProgress.NeedToCompleteTimes,
			CurrentPeriod:               currentPeriodProgress.CurrentPeriod,
		})
	}

	return GetReminderResponse{
		Reminder: reminder,
	}
}
