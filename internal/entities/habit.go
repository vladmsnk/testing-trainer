package entities

import (
	"fmt"
	"time"
)

type Habit struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Goal        *Goal  `json:"goal,omitempty"`        // Goal can be nil (empty)
	IsArchived  bool   `json:"is_archived,omitempty"` // IsArchived can be false (empty)
}

func (h *Habit) Validate() error {
	if h.Name == "" {
		return fmt.Errorf("habit name must be specified")
	}

	return nil
}

type Goal struct {
	Frequency       int           `json:"frequency"`
	Duration        time.Duration `json:"duration"`
	NumOfPeriods    int           `json:"num_of_periods"`
	StartTrackingAt time.Time     `json:"start_tracking_at"`
	EndTrackingAt   time.Time     `json:"end_tracking_at"`
	IsActive        bool          `json:"is_active,omitempty"`
}

func NewHabit(name, description string, goal *Goal) Habit {
	return Habit{
		Name:        name,
		Description: description,
		Goal:        goal,
		IsArchived:  false,
	}
}

func NewGoal(frequency int, duration time.Duration, numOfPeriods int, startTrackingAt time.Time) Goal {
	return Goal{
		Frequency:       frequency,
		Duration:        duration,
		NumOfPeriods:    numOfPeriods,
		StartTrackingAt: startTrackingAt,
		EndTrackingAt:   startTrackingAt.Add(duration * time.Duration(numOfPeriods)),
	}
}
