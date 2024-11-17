package progress

type AddProgressRequest struct {
	HabitName string `json:"habit_name" example:"Drink water"`
}

type GetHabitProgressResponse struct {
	Goal     Goal     `json:"goal"`
	Progress Progress `json:"progress"`
	Habit    Habit    `json:"habit"`
}

type Goal struct {
	Id                   int    `json:"id"`
	FrequencyType        string `json:"frequency_type"`
	TimesPerFrequency    int    `json:"times_per_frequency"`
	TotalTrackingPeriods int    `json:"total_tracking_periods"`
	IsCompleted          bool   `json:"is_completed"`
}

type Habit struct {
	Id          int    `json:"id"`
	Description string `json:"description"`
}

type Progress struct {
	TotalCompletedPeriods int `json:"total_completed_periods"`
	TotalSkippedPeriods   int `json:"total_skipped_periods"`
	TotalCompletedTimes   int `json:"total_completed_times"`
	MostLongestStreak     int `json:"most_longest_streak"`
	CurrentStreak         int `json:"current_streak"`
}

type GetReminderResponse struct {
	Reminder []CurrentPeriodProgress `json:"reminder"`
}

type CurrentPeriodProgress struct {
	Habit                       Habit `json:"habit"`
	Goal                        Goal  `json:"goal"`
	CurrentPeriodCompletedTimes int   `json:"current_period_completed_times"`
	RemainingCompletionCount    int   `json:"remaining_completion_count"`
	CurrentPeriodNumber         int   `json:"current_period_number"`
}
