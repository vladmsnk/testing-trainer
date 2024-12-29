package habit

import (
	"fmt"
)

type CreateHabitRequest struct {
	Description string      `json:"description" example:"Drink 2 liters of water every day"`
	Goal        *UpdateGoal `json:"goal,omitempty"`
}

type UpdateGoal struct {
	FrequencyType        string `json:"frequency_type" example:"daily" enums:"daily,weekly,monthly"` // daily, weekly, monthly
	TimesPerFrequency    int    `json:"times_per_frequency" example:"1"`                             // How many times to complete within each frequency (e.g., per day or per week)
	TotalTrackingPeriods int    `json:"total_tracking_periods" example:"15"`
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
	if len(r.Description) == 0 {
		return fmt.Errorf("description is required")
	}
	if len(r.Description) > 80 {
		return fmt.Errorf("description must not exceed 80 characters")
	}

	// Validate goal fields
	if r.Goal != nil {
		// Validate frequency type
		if _, ok := supportedFrequencyTypes[r.Goal.FrequencyType]; !ok {
			return fmt.Errorf("invalid frequency type")
		}

		// Validate times_per_frequency (1-100)
		if r.Goal.TimesPerFrequency < 1 || r.Goal.TimesPerFrequency > 100 {
			return fmt.Errorf("goal times per frequency must be between 1 and 100")
		}

		// Validate total_tracking_periods (1-1000)
		if r.Goal.TotalTrackingPeriods < 1 || r.Goal.TotalTrackingPeriods > 1000 {
			return fmt.Errorf("goal total tracking periods must be between 1 and 1000")
		}
	}

	return nil
}

func (r *UpdateHabitRequest) Validate() error {
	// Validate ID
	if r.Id <= 0 {
		return fmt.Errorf("id is required")
	}

	// Validate description length
	if len(r.Description) == 0 {
		return fmt.Errorf("description is required")
	}
	if len(r.Description) > 80 {
		return fmt.Errorf("description must not exceed 80 characters")
	}

	// Validate goal fields
	if r.Goal != nil {
		// Validate frequency type
		if _, ok := supportedFrequencyTypes[r.Goal.FrequencyType]; !ok {
			return fmt.Errorf("invalid frequency type")
		}

		// Validate times_per_frequency (1-100)
		if r.Goal.TimesPerFrequency < 1 || r.Goal.TimesPerFrequency > 100 {
			return fmt.Errorf("goal times per frequency must be between 1 and 100")
		}

		// Validate total_tracking_periods (1-1000)
		if r.Goal.TotalTrackingPeriods < 1 || r.Goal.TotalTrackingPeriods > 1000 {
			return fmt.Errorf("goal total tracking periods must be between 1 and 1000")
		}
	}

	return nil
}

type UpdateHabitRequest struct {
	Id          int         `json:"id" example:"1"`
	Description string      `json:"description" example:"Drink 2 liters of water every day"`
	Goal        *UpdateGoal `json:"goal,omitempty"`
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
