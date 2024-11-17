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
			Id:                   progressWithGoal.Habit.Goal.Id,
			FrequencyType:        progressWithGoal.FrequencyType.String(),
			TimesPerFrequency:    progressWithGoal.TimesPerFrequency,
			TotalTrackingPeriods: progressWithGoal.TotalTrackingPeriods,
			IsCompleted:          progressWithGoal.IsCompleted,
		},
		Habit: Habit{
			Id:          progressWithGoal.Habit.Id,
			Description: progressWithGoal.Habit.Description,
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
				Id:                   currentPeriodProgress.Habit.Goal.Id,
				FrequencyType:        currentPeriodProgress.Habit.Goal.FrequencyType.String(),
				TimesPerFrequency:    currentPeriodProgress.Habit.Goal.TimesPerFrequency,
				TotalTrackingPeriods: currentPeriodProgress.Habit.Goal.TotalTrackingPeriods,
			},
			CurrentPeriodCompletedTimes: currentPeriodProgress.CurrentPeriodCompletedTimes,
			RemainingCompletionCount:    currentPeriodProgress.NeedToCompleteTimes,
			CurrentPeriodNumber:         currentPeriodProgress.CurrentPeriod,
		})
	}

	return GetReminderResponse{
		Reminder: reminder,
	}
}
