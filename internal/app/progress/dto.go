package progress

type AddProgressRequest struct {
	HabitName string `json:"habit_name" example:"Drink water"`
}

type GetHabitProgressResponse struct {
	Goal     Goal     `json:"goal"`
	Progress Progress `json:"progress"`
}

type Goal struct {
	FrequencyType        string `json:"frequency_type"`
	TimesPerFrequency    int    `json:"times_per_frequency"`
	TotalTrackingPeriods int    `json:"total_tracking_periods"`
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
	CurrentPeriodCompletedTimes int
	NeedToCompleteTimes         int
	CurrentPeriod               int
}
