package habit

import (
	"fmt"
	"time"
)

type CreateHabitRequest struct {
	Name        string `json:"name" example:"Drink water"`
	Description string `json:"description" example:"Drink 2 liters of water every day"`
	Goal        *Goal  `json:"goal,omitempty"`
}

type Goal struct {
	FrequencyType        string `json:"frequency_type" example:"daily" enums:"daily,weekly,monthly"` // daily, weekly, monthly
	TimesPerFrequency    int    `json:"times_per_frequency" example:"1"`                             // How many times to complete within each frequency (e.g., per day or per week)
	TotalTrackingPeriods int    `json:"total_tracking_periods" example:"15"`                         // How many periods to track the habit
}

func (r *CreateHabitRequest) Validate() error {
	if r.Name == "" {
		return fmt.Errorf("name is required")
	}

	if r.Description == "" {
		return fmt.Errorf("description is required")
	}

	return nil
}

func (r *UpdateHabitRequest) Validate() error {
	if r.Name == "" {
		return fmt.Errorf("name is required")
	}

	if r.Description == "" {
		return fmt.Errorf("description is required")
	}

	return nil
}

type UpdateHabitRequest struct {
	Name        string `json:"name" example:"Drink water"`
	Description string `json:"description" example:"Drink 2 liters of water every day"`
	Goal        *Goal  `json:"goal,omitempty"`
}

type ListUserHabitsResponse struct {
	Username string          `json:"username" example:"john_doe"`
	Habits   []ResponseHabit `json:"habits"`
}

type ResponseHabit struct {
	Name        string        `json:"name" example:"Drink water"`
	Description string        `json:"description" example:"Drink 2 liters of water every day"`
	Goal        *ResponseGoal `json:"goal,omitempty"`
}

type ResponseGoal struct {
	Frequency       int       `json:"frequency" example:"1"`
	DurationInDays  int       `json:"duration_in_days" example:"30"`
	NumOfPeriods    int       `json:"num_of_periods" example:"2"`
	StartTrackingAt time.Time `json:"start_tracking_at" example:"2024-01-01T00:00:00Z"`
}
