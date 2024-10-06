package habit

import "time"

type CreateHabitRequest struct {
	Name            string    `json:"name" example:"Drink water"`
	Description     string    `json:"description" example:"Drink 2 liters of water every day"`
	Frequency       int       `json:"frequency" example:"1"`
	DurationInDays  int       `json:"duration_in_days" example:"30"`
	NumOfPeriods    int       `json:"num_of_periods" example:"2"`
	StartTrackingAt time.Time `json:"start_tracking_at" example:"2024-01-01T00:00:00Z"`
}
