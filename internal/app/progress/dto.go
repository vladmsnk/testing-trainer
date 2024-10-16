package progress

type AddProgressRequest struct {
	HabitName string `json:"habit_name" example:"Drink water"`
}

type GetHabitProgressResponse struct {
	Goal     Goal     `json:"goal"`
	Progress Progress `json:"progress"`
}

type Goal struct {
	FrequencyType     string `json:"frequency_type"`
	TimesPerFrequency int    `json:"times_per_frequency"`
	TotalTrackingDays int    `json:"total_tracking_days"`
}

type Progress struct {
	TotalCompletedPeriods int `json:"total_completed_periods"`
	TotalSkippedPeriods   int `json:"total_skipped_periods"`
	TotalCompletedTimes   int `json:"total_completed_times"`
	MostLongestStreak     int `json:"most_longest_streak"`
	CurrentStreak         int `json:"current_streak"`
}
