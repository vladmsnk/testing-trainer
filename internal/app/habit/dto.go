package habit

import (
	"fmt"
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
	if r.Id == "" {
		return fmt.Errorf("id is required")
	}

	if r.Description == "" {
		return fmt.Errorf("description is required")
	}

	return nil
}

type UpdateHabitRequest struct {
	Id          string `json:"id" example:"1"`
	Description string `json:"description" example:"Drink 2 liters of water every day"`
	Goal        *Goal  `json:"goal,omitempty"`
}

type ListUserHabitsResponse struct {
	Username string          `json:"username" example:"john_doe"`
	Habits   []ResponseHabit `json:"habits"`
}

type ResponseHabit struct {
	Id          string `json:"id" example:"1"`
	Description string `json:"description" example:"Drink 2 liters of water every day"`
	Goal        *Goal  `json:"goal,omitempty"`
}
