package entities

import "time"

type Habit struct {
	Id          int
	Name        string
	Description string
	Goal        *Goal
	IsArchived  bool
}

type Goal struct {
	Id                   int
	FrequencyType        FrequencyType
	TimesPerFrequency    int
	TotalTrackingPeriods int
	IsActive             bool
	CreatedAt            time.Time
	NextCheckDate        time.Time
}

func IsGoalChanged(old, new *Goal) bool {
	if old == nil && new == nil {
		return false
	}

	if old == nil && new != nil || new == nil && old != nil {
		return true
	}

	if *old != *new {
		return true
	}

	return false
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

func (g Goal) GetCurrentPeriod() int {
	switch g.FrequencyType {
	case Daily:
		// Calculate the number of full days since createdAt
		return int(time.Since(g.CreatedAt).Hours() / 24)

	case Weekly:
		// Calculate the number of full weeks since createdAt
		return int(time.Since(g.CreatedAt).Hours() / (24 * 7))

	case Monthly:
		// Calculate the number of full 31-day months since createdAt
		daysSinceCreated := int(time.Since(g.CreatedAt).Hours() / 24)
		return daysSinceCreated / 31 // Each "month" is treated as 31 days

	default:
		return 0
	}
}
