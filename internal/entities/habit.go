package entities

import (
	"time"
)

type Habit struct {
	Id          int
	Name        string
	Description string
	Goal        *Goal
	IsArchived  bool
}

type Goal struct {
	Id                   int
	Username             string
	FrequencyType        FrequencyType
	TimesPerFrequency    int
	TotalTrackingPeriods int
	IsActive             bool
	CreatedAt            time.Time
	NextCheckDate        time.Time
	StartTrackingAt      time.Time
	IsCompleted          bool
	PreviousGoalId       int
	PreviousGoalIDs      []int
}

func IsHabitChanged(old, new Habit) bool {
	if old.Description != new.Description {
		return true
	}
	return false
}

func IsGoalChanged(old, new *Goal) bool {
	if old == nil && new == nil {
		return false
	}

	if old == nil && new != nil || new == nil && old != nil {
		return true
	}

	if !GoalsEqual(*old, *new) {
		return true
	}

	return false
}

func GoalsEqual(old, new Goal) bool {
	if old.FrequencyType != new.FrequencyType {
		return false
	}
	if old.TimesPerFrequency != new.TimesPerFrequency {
		return false
	}
	if old.TotalTrackingPeriods != new.TotalTrackingPeriods {
		return false
	}
	return true
}

type FrequencyType int64

const (
	UndefinedFrequencyType FrequencyType = iota
	Daily
	Weekly
	Monthly
)

func (f FrequencyType) String() string {
	switch f {
	case Daily:
		return "daily"
	case Weekly:
		return "weekly"
	case Monthly:
		return "monthly"
	default:
		return "undefined"
	}
}

func FrequencyTypeFromString(s string) FrequencyType {
	switch s {
	case "daily":
		return Daily
	case "weekly":
		return Weekly
	case "monthly":
		return Monthly
	default:
		return UndefinedFrequencyType
	}

}

func (g Goal) GetCurrentPeriod(currentTime time.Time) int {
	var (
		utcTime   = currentTime.UTC()
		dayOffset = utcTime.Sub(g.StartTrackingAt).Hours() / 24
	)

	switch g.FrequencyType {
	case Daily:
		// Calculate the number of full days since createdAt
		return int(time.Since(g.StartTrackingAt).Hours()/24) + int(dayOffset)

	case Weekly:
		// Calculate the number of full weeks since createdAt
		return int(time.Since(g.StartTrackingAt).Hours()/(24*7)) + int(dayOffset/7)

	case Monthly:
		// Calculate the number of full 31-day months since createdAt
		daysSinceCreated := int(time.Since(g.StartTrackingAt).Hours()/24) + int(dayOffset)
		return daysSinceCreated / 31 // Each "month" is treated as 31 days

	default:
		return 0
	}
}
