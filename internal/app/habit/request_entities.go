package habit

import (
	"fmt"
	"time"
)

type CreateHabitRequest struct {
	Name            string    `json:"name" example:"Drink water"`
	Description     string    `json:"description" example:"Drink 2 liters of water every day"`
	Frequency       int       `json:"frequency" example:"1"`
	DurationInDays  int       `json:"duration_in_days" example:"30"`
	NumOfPeriods    int       `json:"num_of_periods" example:"2"`
	StartTrackingAt time.Time `json:"start_tracking_at" example:"2024-01-01T00:00:00Z"`
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

func (r *CreateHabitRequest) Validate() error {
	if r.Name == "" {
		return fmt.Errorf("name is required")
	}

	if r.Description == "" {
		return fmt.Errorf("description is required")
	}

	if r.Frequency <= 0 {
		return fmt.Errorf("frequency must be greater than 0")
	}

	if r.DurationInDays <= 0 {
		return fmt.Errorf("duration_in_days must be greater than 0")
	}

	if r.NumOfPeriods <= 0 {
		return fmt.Errorf("num_of_periods must be greater than 0")
	}

	return nil
}
