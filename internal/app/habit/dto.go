package habit

import (
	"fmt"
)

type CreateHabitRequest struct {
	Description string `json:"description" example:"Drink 2 liters of water every day"`
	Goal        *Goal  `json:"goal,omitempty"`
}

type Goal struct {
	Id                   int    `json:"id,omitempty" example:"1"`
	FrequencyType        string `json:"frequency_type" example:"daily" enums:"daily,weekly,monthly"` // daily, weekly, monthly
	TimesPerFrequency    int    `json:"times_per_frequency" example:"1"`                             // How many times to complete within each frequency (e.g., per day or per week)
	TotalTrackingPeriods int    `json:"total_tracking_periods" example:"15"`                         // How many periods to track the habit
}

var (
	supportedFrequencyTypes = map[string]bool{
		"daily":   true,
		"weekly":  true,
		"monthly": true,
	}
)

func (r *CreateHabitRequest) Validate() error {
	if r.Description == "" {
		return fmt.Errorf("description is required")
	}

	if r.Goal != nil {
		if _, ok := supportedFrequencyTypes[r.Goal.FrequencyType]; !ok {
			return fmt.Errorf("invalid frequency type")
		}

		if r.Goal.TimesPerFrequency <= 0 {
			return fmt.Errorf("goal times per frequency is required and should be greater then zero")
		}

		if r.Goal.TotalTrackingPeriods <= 0 {
			return fmt.Errorf("goal total tracking periods is required and should be greater then zero")
		}
	}
	return nil
}

func (r *UpdateHabitRequest) Validate() error {
	if r.Id == 0 {
		return fmt.Errorf("id is required")
	}

	if r.Description == "" {
		return fmt.Errorf("description is required")
	}

	if r.Goal != nil {
		if _, ok := supportedFrequencyTypes[r.Goal.FrequencyType]; !ok {
			return fmt.Errorf("invalid frequency type")
		}

		if r.Goal.TimesPerFrequency <= 0 {
			return fmt.Errorf("goal times per frequency is required and should be greater then zero")
		}

		if r.Goal.TotalTrackingPeriods <= 0 {
			return fmt.Errorf("goal total tracking periods is required and should be greater then zero")
		}
	}

	return nil
}

type UpdateHabitRequest struct {
	Id          int    `json:"id" example:"1"`
	Description string `json:"description" example:"Drink 2 liters of water every day"`
	Goal        *Goal  `json:"goal,omitempty"`
}

type ListUserHabitsResponse struct {
	Username string          `json:"username" example:"john_doe"`
	Habits   []ResponseHabit `json:"habits"`
}

type ResponseHabit struct {
	Id          int    `json:"id" example:"1"`
	Description string `json:"description" example:"Drink 2 liters of water every day"`
	Goal        *Goal  `json:"goal,omitempty"`
}
