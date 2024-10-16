package progress

type HabitProgressResponse struct {
	HabitName        string `json:"habit_name" example:"Drink water"`
	HabitDescription string `json:"habit_description" example:"Drink 2 liters of water every day"`
}

type Stat struct {
	TotalExecutions   int `json:"execution_times" example:"2"`
	MostLongestStreak int `json:"longest_streak" example:"5"`
	CurrentStreak     int `json:"current_streak" example:"3"`
	CompletionRate    int `json:"completion_rate" example:"60"`
	SkippedDays       int `json:"skipped_days" example:"2"`
	RemainingDays     int `json:"remaining_days" example:"3"`
}

type AddProgressRequest struct {
	HabitName string `json:"habit_name" example:"Drink water"`
}
