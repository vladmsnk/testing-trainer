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
